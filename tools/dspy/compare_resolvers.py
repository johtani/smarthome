"""Compare the production Go and DSPy resolvers on finalized splits without executing devices."""

import argparse
import hashlib
import json
import os
import statistics
import subprocess
import time
import unicodedata
from pathlib import Path

from optimize_and_evaluate import (
    ResolverProgram, build_lm_config, compile_resolver, dspy, format_command_catalog,
    read_jsonl, to_examples,
)
from dspy_common.catalog import build_catalog_text, parse_command_list

ROOT = Path(__file__).resolve().parents[2]
ROUTES = ("legacy", "baseline", "optimized")


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_splits(paths):
    owners = {}
    splits = {}
    for name, path in paths.items():
        rows = read_jsonl(path)
        if not rows:
            raise ValueError(f"{name} is empty")
        for row in rows:
            for field in ("input_text", "group_id", "expected_command", "expected_args"):
                if not isinstance(row.get(field), str):
                    raise ValueError(f"{name}: missing or invalid {field}")
            text = " ".join(unicodedata.normalize("NFKC", row["input_text"]).casefold().split())
            group = row["group_id"].strip()
            if not group or not text:
                raise ValueError(f"{name}: empty input or group_id")
            for key in (("group", group), ("input", text)):
                if owners.setdefault(key, name) != name:
                    raise ValueError(f"{key[0]} shared across splits")
        splits[name] = rows
    return splits


def summarize(rows, route, stage):
    exact = false_positives = unsupported = unresolved = invalid_output = 0
    for row in rows:
        pred = row["decisions"][route][stage]
        cmd, args = pred["command"], pred["args"]
        unsupported += not row["expected_command"]
        if pred.get("error"):
            invalid_output += 1
            continue
        exact += (cmd, args) == (row["expected_command"], row["expected_args"])
        false_positives += not row["expected_command"] and bool(cmd)
        unresolved += not cmd
    return dict(total=len(rows), exact=exact, accuracy=exact / len(rows),
                false_positives=false_positives, unsupported=unsupported, unresolved=unresolved,
                invalid_output=invalid_output)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("train", "dev", "test", "command-catalog", "report-out"):
        parser.add_argument("--" + name, required=True, type=Path)
    for name in ("model", "api-base", "prompt-version", "artifact-version", "server-cache"):
        parser.add_argument("--" + name, required=True)
    parser.add_argument("--config-dir", type=Path, default=Path("config"))
    opts = parser.parse_args()
    paths = {name: getattr(opts, name) for name in ("train", "dev", "test")}
    splits = load_splits(paths)
    helper = ["go", "run", "./tools/resolver-compare", "-config-dir", str(opts.config_dir.resolve())]
    help_text = subprocess.check_output(helper + ["-catalog"], cwd=ROOT).decode("utf-8")
    catalog = format_command_catalog(json.loads(opts.command_catalog.read_text(encoding="utf-8")))
    if build_catalog_text(parse_command_list(help_text)) != catalog:
        parser.error("catalog snapshot differs from config-dir command definitions")
    if not opts.model.startswith("openai/"):
        parser.error("--model must be openai/<served model ID>")
    conf = build_lm_config(overrides={"model": opts.model, "api_base": opts.api_base})
    if conf["kwargs"]["model_type"] != "chat":
        parser.error("the legacy resolver requires a chat model")
    lm = dspy.LM(conf["model"], cache=False, **conf["kwargs"])
    dspy.configure(lm=lm)
    trainset = to_examples(splits["train"], catalog, opts.prompt_version)
    programs = {"baseline": ResolverProgram(), "optimized": compile_resolver(trainset)}
    samples, rows = [], []
    for item in splits["test"]:
        sample = {"Text": item["input_text"], "Predictions": {}}
        timings = {}
        for name, program in programs.items():
            start = time.perf_counter()
            pred = program(utterance=item["input_text"], command_catalog=catalog,
                           prompt_version=opts.prompt_version)
            timings[name] = (time.perf_counter() - start) * 1000
            sample["Predictions"][name] = {"command": pred.selected_command.strip(),
                                           "args": pred.selected_args.strip()}
        samples.append(sample)
        rows.append({key: item[key] for key in ("input_text", "expected_command", "expected_args")})
        rows[-1]["latency_ms"] = timings

    request = dict(Help=help_text, Model=opts.model.removeprefix("openai/"),
                   Endpoint=opts.api_base.rstrip("/") + "/chat/completions",
                   PromptVersion=opts.prompt_version, Samples=samples)
    env = dict(os.environ, LM_API_KEY=conf["kwargs"].get("api_key", ""))
    results = json.loads(subprocess.check_output(helper, cwd=ROOT, env=env,
                                                 input=json.dumps(request).encode("utf-8")))
    for row, result in zip(rows, results, strict=True):
        row["latency_ms"]["legacy"] = result["LegacyMS"]
        row["decisions"] = {}
        expected = (row["expected_command"], row["expected_args"])
        for route, stages in result["Decisions"].items():
            row["decisions"][route] = {
                stage: {"command": pred["command"], "args": pred["args"],
                        "correct": (pred["command"], pred["args"]) == expected}
                for stage, pred in stages.items()
            }
            if route == "legacy" and result.get("LegacyError"):
                for pred in row["decisions"][route].values():
                    pred.update(error=result["LegacyError"], correct=False)

    summary = {route: {stage: summarize(rows, route, stage) for stage in ("Raw", "Final")}
               for route in ROUTES}
    for route in ROUTES:
        summary[route]["median_ms"] = statistics.median(row["latency_ms"][route] for row in rows)
    report = dict(model=opts.model, api_base=opts.api_base, prompt_version=opts.prompt_version,
                  artifact_version=opts.artifact_version, split_sha256={k: sha256(v) for k,v in paths.items()},
                  catalog_sha256=sha256(opts.command_catalog), dspy_cache=False,
                  server_cache=opts.server_cache, dspy_settings={k: lm.kwargs.get(k) for k in
                  ("temperature", "max_tokens")}, legacy_settings="server defaults, JSON response format",
                  summary=summary, rows=rows)
    opts.report_out.parent.mkdir(parents=True, exist_ok=True)
    opts.report_out.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")


if __name__ == "__main__":
    main()
