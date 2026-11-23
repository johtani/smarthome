"""Freeze human-reviewed resolver labels into reproducible, group-disjoint splits."""

import argparse
from collections import Counter, defaultdict
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys
from tempfile import TemporaryDirectory
import unicodedata

ROOT = Path(__file__).resolve().parents[2]
FIXTURES = Path(__file__).with_name("regression_cases.jsonl")
SOURCES = {"production_log", "handwritten", "synthetic", "synthetic_template", "synthetic_llm"}
SENSITIVE = re.compile(r"(?i)bearer\s+[\w.\-]{12,}|(?:api[_-]?key|access[_-]?token|password|secret)\s*[:=]\s*\S+")
REQUIRED_FIXTURES = {"regression-ps5-explicit", "regression-generic-music", "regression-out-of-scope"}


def encode(value):
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"), allow_nan=False)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def normalized(text):
    return " ".join(unicodedata.normalize("NFKC", text).casefold().split())


def read_rows(path):
    rows = []
    for line, raw in enumerate(path.read_text(encoding="utf-8-sig").splitlines(), 1):
        if raw.strip():
            try:
                row = json.loads(raw)
            except ValueError as exc:
                raise ValueError(f"{path.name}:{line}: invalid JSON") from exc
            if not isinstance(row, dict):
                raise ValueError(f"{path.name}:{line}: expected an object")
            rows.append(row)
    return rows


def check_fields(row, names, allow_empty=()):
    for name in names:
        value = row.get(name)
        if not isinstance(value, str) or (not value.strip() and not (name in allow_empty and value == "")):
            raise ValueError(f"missing or invalid {name}")


def load_reviewed(path, counts):
    rows = []
    ids = set()
    for n, row in enumerate(read_rows(path), 1):
        try:
            check_fields(row, ["case_id", "review_status"])
            if row["case_id"] in ids:
                raise ValueError("duplicate case_id")
            ids.add(row["case_id"])
            if row.get("schema_version") != "resolver-reviewed/v1" or row.get("reviewed") is not True:
                raise ValueError("reviewed resolver-reviewed/v1 row required")
            if row["review_status"] == "excluded":
                counts["excluded"] += 1
                continue
            if row["review_status"] not in {"accepted", "corrected"}:
                raise ValueError("unreviewed or pending decision")
            check_fields(row, ["input_text", "expected_command", "expected_args", "group_id", "source",
                               "reviewed_at"], ("expected_command", "expected_args"))
            correction = row.get("feedback_correction")
            if row.get("feedback_label") == "incorrect" and not (isinstance(correction, str) and correction.strip()):
                candidate = (row.get("candidate_command"), row.get("candidate_args"))
                expected = (row["expected_command"], row["expected_args"])
                if (row["review_status"] != "corrected" or
                        not all(isinstance(value, str) for value in candidate) or candidate == expected):
                    raise ValueError("incorrect feedback without a correction")
            if any(SENSITIVE.search(row[k]) for k in ("input_text", "expected_command", "expected_args")):
                raise ValueError("possible secret in dataset fields")
            source = row["source"]
            if source == "handwritten-fixture":
                source = "handwritten"
            elif row.get("trace_id") and source not in SOURCES:
                # The ES extractor records the contributing event names as source.
                if all(event.startswith(("resolver.", "llm.", "dspy.")) for event in source.split(",")) or source == "resolver-log":
                    source = "production_log"
            if source not in SOURCES:
                raise ValueError("unknown source")
            rows.append({k:row[k] for k in ("case_id", "input_text", "expected_command", "expected_args", "group_id")} |
                        {"source":source, "request_id":row.get("resolver_request_id") or row["case_id"],
                         "llm_model":row.get("llm_model", ""), "prompt_version":row.get("prompt_version", "")})
        except ValueError as exc:
            raise ValueError(f"{path.name}:{n}: {exc}") from exc
    return rows


def load_fixtures():
    fixtures = read_rows(FIXTURES)
    for row in fixtures:
        check_fields(row, ["request_id","input_text","expected_command","expected_args"], ("expected_command","expected_args"))
        row.update(case_id=row["request_id"], group_id=row["request_id"], source="handwritten")
    ids = {r["case_id"] for r in fixtures}
    if len(ids) != len(fixtures) or not REQUIRED_FIXTURES <= ids:
        raise ValueError("duplicate or missing required regression fixture")
    return fixtures


def validate_catalog(config_dir, rows, fixtures):
    if not config_dir.is_dir():
        raise ValueError("config-dir does not exist")
    targets = sorted({(r["expected_command"], r["expected_args"]) for r in rows + fixtures})
    invalid = [(p["command"],p["args"]) for r in fixtures for p in r.get("invalid_predictions", [])]
    labels = [{"command":c,"args":a} for c,a in targets + invalid]
    result = subprocess.run(["go","run","./tools/resolver-catalog","-config-dir",str(config_dir.resolve())],
                            input=encode(labels), text=True, encoding="utf-8", stdout=subprocess.PIPE, cwd=ROOT, check=True)
    payload = json.loads(result.stdout)
    reasons = payload["errors"]
    failures = {target:reason for target,reason in zip(targets,reasons) if reason}
    if failures:
        affected = [r["case_id"] + ": " + failures[(r["expected_command"],r["expected_args"])]
                    for r in rows + fixtures if (r["expected_command"],r["expected_args"]) in failures]
        raise ValueError("invalid expected label: " + ", ".join(affected))
    expected_reasons = [p["reason"] for r in fixtures for p in r.get("invalid_predictions", [])]
    if set(expected_reasons) != {"invalid_prefix", "invalid_enum_value"} or reasons[len(targets):] != expected_reasons:
        raise ValueError("negative argument fixture missing or accepted by runtime")
    return payload["catalog"]


def split_rows(rows, fixtures, aliases, seed):
    # Connect explicit groups, normalized duplicates, and fixed examples before deduplication.
    parents = {}
    def root(k):
        while k != parents.setdefault(k,k):
            parents[k] = parents[parents[k]]
            k = parents[k]
        return k

    groups = {r["group_id"] for r in rows + fixtures}
    if not isinstance(aliases, dict) or any(k not in groups or not isinstance(v,str) or not v.strip() for k,v in aliases.items()):
        raise ValueError("group map requires known group IDs and non-empty string family names")
    # Resolve maps that reference another reviewed group as well as an external family name.
    for group, alias in aliases.items():
        a, b = root(group), root(alias)
        parents[max(a, b)] = min(a, b)
    text_groups = {}
    targets = {}
    for row in fixtures + rows:
        group = aliases.get(row["group_id"], row["group_id"])
        key = normalized(row["input_text"])
        target = (row["expected_command"],row["expected_args"])
        if targets.setdefault(key,target) != target:
            raise ValueError(f"conflicting labels for normalized input at {row['case_id']}")
        a, b = root(group), root(text_groups.setdefault(key,group))
        parents[max(a,b)] = min(a,b)

    def family(row):
        return root(text_groups[normalized(row["input_text"])])

    reserved = {family(r) for r in fixtures}
    unique = {}
    split_groups = defaultdict(list)
    overlaps = []
    duplicates = []
    # Prefer real traffic when several sources provide the same input/target.
    for row in sorted(rows, key=lambda r:(r["source"] != "production_log",r["case_id"])):
        key = normalized(row["input_text"])
        group = family(row)
        if key in unique:
            duplicates.append([row["case_id"], unique[key]])
        unique.setdefault(key,row["case_id"])
        if group in reserved:
            overlaps.append(row["case_id"])
        elif unique[key] == row["case_id"]:
            split_groups[group].append(dict(row,group_id=group))
    ordered = sorted(split_groups,key=lambda g:sha(encode([seed,g]).encode()))
    if len(ordered) < 3:
        raise ValueError("at least three independent non-regression groups are required")
    # Prefer real traffic for the held-out test. Keep at least one group for train and dev.
    size_test = min(max(1,round(len(ordered)*0.2)),len(ordered)-2)
    candidates = sorted(ordered,key=lambda g:not all(r["source"]=="production_log" for r in split_groups[g]))
    negative = next((g for g in candidates if any(not r["expected_command"] and r["source"] == "production_log"
                                                   for r in split_groups[g])), None)
    if negative is not None:
        candidates.remove(negative)
        candidates.insert(0, negative)
    test = candidates[:size_test]
    remaining = [g for g in ordered if g not in test]
    size_dev = min(max(1,round(len(ordered)*0.2)),len(remaining)-1)
    membership = {"train":remaining[size_dev:], "dev":remaining[:size_dev], "test":test}
    splits = {name:[row for group in member for row in split_groups[group]] for name,member in membership.items()}
    return splits, duplicates, overlaps


def write_json(path, obj):
    path.write_bytes((json.dumps(obj, ensure_ascii=False, indent=2, sort_keys=True, allow_nan=False)+"\n").encode())


def jsonl(rows):
    return "".join(encode(r)+"\n" for r in rows).encode()


def generate(opts, folder, report):
    rows = load_reviewed(opts.reviewed, report["counts"])
    fixtures = load_fixtures()
    aliases = json.loads(opts.group_map.read_text(encoding="utf-8-sig")) if opts.group_map else {}
    catalog = validate_catalog(opts.config_dir, rows, fixtures)
    splits, duplicates, overlaps = split_rows(rows, fixtures, aliases, opts.seed)
    dataset = [r for split in splits.values() for r in split]
    report.update(duplicate_case_pairs=duplicates, regression_overlap_ids=overlaps)
    report["counts"].update(reviewed=len(rows), duplicate_rows=len(duplicates), regression_overlap_rows=len(overlaps),
                            dataset=len(dataset), regression=len(fixtures))
    report["splits"] = {k:{"rows":len(v),"groups":len({r["group_id"] for r in v}),
                           "sources":dict(Counter(r["source"] for r in v)),
                           "unresolved":sum(not r["expected_command"] for r in v)} for k,v in splits.items()}
    if not report["splits"]["test"]["sources"].get("production_log"):
        report["warnings"].append("No production-log examples in test")
    if not report["splits"]["test"]["unresolved"]:
        report["warnings"].append("No out-of-scope examples in test")
    files = {name+".jsonl":jsonl(values) for name,values in dict(splits,dataset=dataset,regression=fixtures).items()}
    files["catalog.json"] = (encode(catalog)+"\n").encode()
    hashes = {name:sha(data) for name,data in files.items()}
    manifest = {"schema_version":"resolver-evaluation/v1", "dataset_version":"sha256:"+hashes["dataset.jsonl"],
                "fixture_version":"sha256:"+hashes["regression.jsonl"], "sha256":hashes,
                "reviewed_sha256":sha(opts.reviewed.read_bytes()), "group_map_sha256":sha(encode(aliases).encode()),
                "split_algorithm":"normalized-group-sha256/v1", "target_group_ratios":{"train":0.6,"dev":0.2,"test":0.2},
                "seed":opts.seed,"model_ids":opts.model,"prompt_version":opts.prompt_version,
                "thresholds":{"min_command_accuracy":opts.min_command_accuracy,"min_arg_accuracy":opts.min_arg_accuracy},
                "period":{"from":opts.period_from,"to_exclusive":opts.period_to},"splits":report["splits"]}
    for name,data in files.items():
        (folder/name).write_bytes(data)
    write_json(folder/"evaluation-manifest.json",manifest)


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--reviewed",required=True,type=Path)
    parser.add_argument("--group-map",type=Path)
    parser.add_argument("--config-dir",required=True,type=Path)
    parser.add_argument("--output-dir",required=True,type=Path)
    parser.add_argument("--model",required=True,action="append")
    parser.add_argument("--prompt-version",required=True)
    parser.add_argument("--from",dest="period_from",required=True)
    parser.add_argument("--to",dest="period_to",required=True)
    parser.add_argument("--seed",type=int,default=42)
    parser.add_argument("--min-command-accuracy",type=float,default=0.8)
    parser.add_argument("--min-arg-accuracy",type=float,default=0.6)
    opts = parser.parse_args(argv)
    dest = opts.output_dir
    if dest.exists():
        parser.error("output-dir must be a new directory")
    dest.parent.mkdir(parents=True,exist_ok=True)
    with TemporaryDirectory(dir=dest.parent) as temp:
        stage = Path(temp)/"result"
        stage.mkdir()
        report = {"errors":[], "warnings":[], "counts":{"excluded":0}}
        try:
            start,end = [datetime.fromisoformat(t.replace("Z","+00:00")) for t in (opts.period_from,opts.period_to)]
            if not start.tzinfo or not end.tzinfo or start>=end:
                raise ValueError("expected a timezone-aware non-empty source period")
            if not all(m.strip() for m in opts.model+[opts.prompt_version]) or not all(0<=x<=1 for x in (opts.min_command_accuracy,opts.min_arg_accuracy)):
                raise ValueError("invalid model, prompt version or accuracy threshold")
            generate(opts,stage,report)
        except (ValueError,OSError,subprocess.CalledProcessError) as exc:
            report["errors"].append(str(exc))
            stage = Path(temp)/"invalid"
            stage.mkdir()
        write_json(stage/"validation.json",report)
        lines = ["# Dataset validation","", "## Counts",""] + [f"- {k}: {v}" for k,v in report["counts"].items()]
        for name,summary in report.get("splits",{}).items():
            lines.append(f"- {name}: {summary['rows']} rows, {summary['groups']} groups, sources {summary['sources']}")
        lines.extend(["", "## Errors and warnings","",*(report["errors"]+report["warnings"] or ["None"])])
        (stage/"validation.md").write_text("\n".join(lines)+"\n",encoding="utf-8")
        stage.rename(dest)
    print(dest/"validation.md")
    return int(bool(report["errors"]))


if __name__ == "__main__":
    sys.exit(main())
