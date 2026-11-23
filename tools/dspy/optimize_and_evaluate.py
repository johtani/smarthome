#!/usr/bin/env python3
"""Run DSPy optimization and offline evaluation with a gate."""

from __future__ import annotations

import argparse
import hashlib
import json
import random
import re
import sys
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Dict, List

TOOLS_DIR = Path(__file__).resolve().parents[1]
if str(TOOLS_DIR) not in sys.path:
    sys.path.insert(0, str(TOOLS_DIR))

import dspy

from dspy_common.lm_config import build_lm_config
from dspy_common.artifacts import PROGRAM_SCHEMA_VERSION, save_artifact
from dspy_common.reporting import build_metadata_breakdown
from dspy_common.resolver_program import ResolverProgram, format_command_catalog


@dataclass
class EvalRow:
    request_id: str
    input_text: str
    expected_command: str
    expected_args: str
    pred_command: str
    pred_args: str
    llm_model: str
    prompt_version: str
    artifact_version: str
    dataset_version: str


def read_jsonl(path: Path) -> List[Dict[str, Any]]:
    rows: List[Dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            rows.append(json.loads(line))
    return rows


def to_examples(
    rows: List[Dict[str, Any]], command_catalog: str, prompt_version: str
) -> List[dspy.Example]:
    ex = []
    for r in rows:
        ex.append(
            dspy.Example(
                request_id=r.get("request_id", ""),
                utterance=r["input_text"],
                command_catalog=command_catalog,
                prompt_version=prompt_version,
                selected_command=r.get("expected_command", ""),
                selected_args=r.get("expected_args", ""),
                rationale="reviewed expected resolution",
            ).with_inputs("utterance", "command_catalog", "prompt_version")
        )
    return ex


def metric(example: dspy.Example, pred: dspy.Prediction, trace: Any = None) -> bool:
    expected = ((example.selected_command or "").strip(), (example.selected_args or "").strip())
    predicted = ((pred.selected_command or "").strip(), (pred.selected_args or "").strip())
    return predicted == expected and (bool(expected[0]) or not expected[1])


def compile_resolver(trainset):
    optimizer = dspy.BootstrapFewShot(metric=metric, max_bootstrapped_demos=4, max_labeled_demos=4)
    return optimizer.compile(ResolverProgram(), trainset=trainset)


def violates_catalog(command: str, args: str, command_catalog: str) -> bool:
    """Check command presence and the catalog's no-args, prefix, and enum constraints."""

    matching_line = next(
        (line for line in command_catalog.splitlines() if line.startswith(f"- {command}:")),
        "",
    )
    if not command:
        return bool(args)
    if not matching_line:
        return True

    hint_match = re.search(r" \[args: (.+)\]$", matching_line)
    if not hint_match:
        return bool(args)
    if not args:
        return "required" in hint_match.group(1)

    hint = hint_match.group(1)
    allowed_prefixes = set(re.findall(r"prefix=([^,)]+)", hint))
    supplied_prefixed = re.findall(r"(?:^|\s)([^\s:]+:)([^\s]+)", args)
    for prefix, value in supplied_prefixed:
        if prefix not in allowed_prefixes:
            return True
        spec_match = re.search(
            rf"[^,]*prefix={re.escape(prefix)}[^)]*enum=([^)]*)", hint
        )
        if spec_match and value not in spec_match.group(1).split("|"):
            return True

    no_prefix_enum = re.search(r"\([^)]*enum=([^)]*)[^)]*no-prefix[^)]*\)", hint)
    if no_prefix_enum and not supplied_prefixed and args not in no_prefix_enum.group(1).split("|"):
        return True
    return False


def evaluate(
    program: ResolverProgram,
    rows: List[Dict[str, Any]],
    command_catalog: str,
    prompt_version: str,
) -> Dict[str, Any]:
    if not rows:
        return {
            "count": 0,
            "command_accuracy": 0.0,
            "arg_accuracy": 0.0,
            "unresolved_rate": 0.0,
            "catalog_violation_rate": 0.0,
            "breakdown": {"by_llm_model": {}, "by_prompt_version": {}},
            "rows": [],
        }

    eval_rows: List[EvalRow] = []
    cmd_ok = 0
    arg_ok = 0
    unresolved = 0
    catalog_violations = 0
    for r in rows:
        pred = program(
            utterance=r["input_text"],
            command_catalog=command_catalog,
            prompt_version=prompt_version,
        )
        exp_cmd = (r.get("expected_command") or "").strip()
        exp_args = (r.get("expected_args") or "").strip()
        got_cmd = (pred.selected_command or "").strip()
        got_args = (pred.selected_args or "").strip()

        if not got_cmd:
            unresolved += 1
        if violates_catalog(got_cmd, got_args, command_catalog):
            catalog_violations += 1

        if exp_cmd == got_cmd:
            cmd_ok += 1
        if exp_args == got_args:
            arg_ok += 1

        eval_rows.append(
            EvalRow(
                request_id=r.get("request_id", ""),
                input_text=r["input_text"],
                expected_command=exp_cmd,
                expected_args=exp_args,
                pred_command=got_cmd,
                pred_args=got_args,
                llm_model=(r.get("llm_model") or "").strip(),
                prompt_version=(r.get("prompt_version") or "").strip(),
                artifact_version=(r.get("artifact_version") or "").strip(),
                dataset_version=(r.get("dataset_version") or "").strip(),
            )
        )

    n = len(rows)
    serialized_rows = [asdict(x) for x in eval_rows]
    return {
        "count": n,
        "command_accuracy": cmd_ok / n,
        "arg_accuracy": arg_ok / n,
        "unresolved_rate": unresolved / n,
        "catalog_violation_rate": catalog_violations / n,
        "breakdown": build_metadata_breakdown(serialized_rows),
        "rows": serialized_rows[:20],
    }


def read_manifest(path: Path, model: str):
    manifest = json.loads(path.read_bytes())
    if manifest["schema_version"] != "resolver-evaluation/v1" or model not in manifest["model_ids"]:
        raise ValueError("unsupported evaluation manifest or model")
    files = {}
    for name in ("dataset.jsonl", "train.jsonl", "dev.jsonl", "test.jsonl", "regression.jsonl", "catalog.json"):
        data = (path.parent / name).read_bytes()
        if hashlib.sha256(data).hexdigest() != manifest["sha256"][name]:
            raise ValueError(f"evaluation manifest hash mismatch for {name}")
        files[name] = data
    for field, name in (("dataset_version", "dataset.jsonl"), ("fixture_version", "regression.jsonl")):
        if manifest[field] != "sha256:" + manifest["sha256"][name]:
            raise ValueError(f"evaluation manifest version mismatch for {name}")
    split_names = ("train.jsonl", "dev.jsonl", "test.jsonl")
    if b"".join(files[name] for name in split_names) != files["dataset.jsonl"]:
        raise ValueError("dataset does not match its splits")
    splits = [[json.loads(line) for line in files[name].splitlines()] for name in split_names]
    return manifest, splits, json.loads(files["catalog.json"])


def main() -> int:
    parser = argparse.ArgumentParser()
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--dataset-jsonl")
    source.add_argument("--evaluation-manifest", type=Path)
    parser.add_argument("--command-catalog")
    parser.add_argument("--model", required=True, help="e.g. openai/gpt-4o-mini")
    parser.add_argument("--api-base", default=None, help="OpenAI-compatible API base URL")
    parser.add_argument("--api-key", default=None, help="API key for the configured LM")
    parser.add_argument("--model-type", default=None, help="DSPy LM model_type, e.g. chat")
    parser.add_argument("--temperature", default=None, type=float)
    parser.add_argument("--max-tokens", default=None, type=int)
    parser.add_argument("--prompt-version")
    parser.add_argument("--dataset-version")
    parser.add_argument("--artifact-version", default="")
    parser.add_argument("--artifact-out", default="")
    parser.add_argument("--report-out", required=True)
    parser.add_argument("--seed", type=int)
    parser.add_argument("--train-ratio", type=float)
    parser.add_argument("--dev-ratio", type=float)
    parser.add_argument("--min-command-accuracy", type=float)
    parser.add_argument("--min-arg-accuracy", type=float)
    args = parser.parse_args()

    manifest_hash = None
    if args.evaluation_manifest:
        manifest, (train_rows, dev_rows, test_rows), catalog_obj = read_manifest(args.evaluation_manifest, args.model)
        conditions = dict(manifest["thresholds"], seed=manifest["seed"], prompt_version=manifest["prompt_version"],
                          dataset_version=manifest["dataset_version"], train_ratio=manifest["target_group_ratios"]["train"],
                          dev_ratio=manifest["target_group_ratios"]["dev"])
        if args.command_catalog or any(getattr(args, key) is not None for key in conditions):
            parser.error("the evaluation manifest supplies the catalog and all dataset conditions")
        for key, value in conditions.items():
            setattr(args, key, value)
        data = train_rows + dev_rows + test_rows
        manifest_hash = hashlib.sha256(args.evaluation_manifest.read_bytes()).hexdigest()
        random.seed(args.seed)
    else:
        if not args.command_catalog:
            parser.error("--command-catalog is required with --dataset-jsonl")
        if args.prompt_version is None:
            args.prompt_version = 'offline-eval-v1'
        if args.dataset_version is None:
            args.dataset_version = ''
        if args.seed is None:
            args.seed = 42
        if args.train_ratio is None:
            args.train_ratio = 0.6
        if args.dev_ratio is None:
            args.dev_ratio = 0.2
        if args.min_command_accuracy is None:
            args.min_command_accuracy = 0.8
        if args.min_arg_accuracy is None:
            args.min_arg_accuracy = 0.6
        random.seed(args.seed)
        data = read_jsonl(Path(args.dataset_jsonl))
        if len(data) < 10:
            raise SystemExit("dataset is too small: need at least 10 rows")

        with Path(args.command_catalog).open("r", encoding="utf-8") as f:
            catalog_obj = json.load(f)

        if args.train_ratio <= 0 or args.dev_ratio <= 0 or args.train_ratio + args.dev_ratio >= 1:
            raise SystemExit("train-ratio and dev-ratio must be positive and sum to less than 1")
        random.shuffle(data)
        train_end = max(1, int(len(data) * args.train_ratio))
        dev_end = max(train_end + 1, train_end + int(len(data) * args.dev_ratio))
        train_rows = data[:train_end]
        dev_rows = data[train_end:dev_end]
        test_rows = data[dev_end:]
        if not dev_rows or not test_rows:
            raise SystemExit("dataset split must leave at least one dev and one test row")

    catalog_text = format_command_catalog(catalog_obj)

    dataset_version = args.dataset_version.strip()
    if not dataset_version:
        dataset_version = "sha256:" + hashlib.sha256(
            Path(args.dataset_jsonl).read_bytes()
        ).hexdigest()[:16]
    artifact_version = args.artifact_version.strip() or (
        f"{args.prompt_version}-{dataset_version.replace(':', '-')}-{args.seed}"
    )

    lm_conf = build_lm_config(
        overrides={
            "model": args.model,
            "api_base": args.api_base,
            "api_key": args.api_key,
            "model_type": args.model_type,
            "temperature": args.temperature,
            "max_tokens": args.max_tokens,
        }
    )
    dspy.configure(lm=dspy.LM(lm_conf["model"], **lm_conf["kwargs"]))

    baseline = ResolverProgram()

    trainset = to_examples(train_rows, catalog_text, args.prompt_version)
    optimized = compile_resolver(trainset)
    optimized_dev_eval = evaluate(optimized, dev_rows, catalog_text, args.prompt_version)
    baseline_eval = evaluate(baseline, test_rows, catalog_text, args.prompt_version)
    optimized_eval = evaluate(optimized, test_rows, catalog_text, args.prompt_version)

    gate_passed = (
        optimized_eval["command_accuracy"] >= args.min_command_accuracy
        and optimized_eval["arg_accuracy"] >= args.min_arg_accuracy
        and optimized_eval["catalog_violation_rate"] == 0.0
    )

    report = {
        "dataset_size": len(data),
        "evaluation_manifest_sha256": manifest_hash,
        "train_size": len(train_rows),
        "dev_size": len(dev_rows),
        "test_size": len(test_rows),
        "model": lm_conf["model"],
        "prompt_version": args.prompt_version,
        "program_schema_version": PROGRAM_SCHEMA_VERSION,
        "dataset_version": dataset_version,
        "artifact_version": artifact_version,
        "optimizer": {
            "name": "BootstrapFewShot",
            "max_bootstrapped_demos": 4,
            "max_labeled_demos": 4,
            "seed": args.seed,
            "train_ratio": args.train_ratio,
            "dev_ratio": args.dev_ratio,
        },
        "thresholds": {
            "min_command_accuracy": args.min_command_accuracy,
            "min_arg_accuracy": args.min_arg_accuracy,
        },
        "baseline": baseline_eval,
        "optimized_dev": optimized_dev_eval,
        "optimized": optimized_eval,
        "gate_passed": gate_passed,
    }

    out = Path(args.report_out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    if gate_passed and args.artifact_out:
        save_artifact(Path(args.artifact_out), optimized, report)
        print(f"wrote artifact: {args.artifact_out}")
    print(f"wrote report: {out}")
    print(f"gate_passed={gate_passed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
