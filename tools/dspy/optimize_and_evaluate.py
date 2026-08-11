#!/usr/bin/env python3
"""Run DSPy optimization and offline evaluation with a gate."""

from __future__ import annotations

import argparse
import json
import random
import sys
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Dict, List

TOOLS_DIR = Path(__file__).resolve().parents[1]
if str(TOOLS_DIR) not in sys.path:
    sys.path.insert(0, str(TOOLS_DIR))

import dspy

from dspy_common.lm_config import build_lm_config
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


def metric(example: dspy.Example, pred: dspy.Prediction, trace: Any = None) -> float:
    exp_cmd = (example.selected_command or "").strip().lower()
    exp_args = (example.selected_args or "").strip()
    got_cmd = (pred.selected_command or "").strip().lower()
    got_args = (pred.selected_args or "").strip()

    cmd_score = 1.0 if exp_cmd == got_cmd else 0.0
    arg_score = 1.0 if exp_args == got_args else 0.0
    if exp_cmd == "":
        return 1.0 if got_cmd == "" else 0.0
    return (cmd_score * 0.8) + (arg_score * 0.2)


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
            "breakdown": {"by_llm_model": {}, "by_prompt_version": {}},
            "rows": [],
        }

    eval_rows: List[EvalRow] = []
    cmd_ok = 0
    arg_ok = 0
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
        "breakdown": build_metadata_breakdown(serialized_rows),
        "rows": serialized_rows[:20],
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset-jsonl", required=True)
    parser.add_argument("--command-catalog", required=True)
    parser.add_argument("--model", required=True, help="e.g. openai/gpt-4o-mini")
    parser.add_argument("--api-base", default=None, help="OpenAI-compatible API base URL")
    parser.add_argument("--api-key", default=None, help="API key for the configured LM")
    parser.add_argument("--model-type", default=None, help="DSPy LM model_type, e.g. chat")
    parser.add_argument("--temperature", default=None, type=float)
    parser.add_argument("--max-tokens", default=None, type=int)
    parser.add_argument("--prompt-version", default="offline-eval-v1")
    parser.add_argument("--report-out", required=True)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument("--train-ratio", type=float, default=0.8)
    parser.add_argument("--min-command-accuracy", type=float, default=0.80)
    parser.add_argument("--min-arg-accuracy", type=float, default=0.60)
    args = parser.parse_args()

    random.seed(args.seed)
    data = read_jsonl(Path(args.dataset_jsonl))
    if len(data) < 10:
        raise SystemExit("dataset is too small: need at least 10 rows")

    with Path(args.command_catalog).open("r", encoding="utf-8") as f:
        catalog_obj = json.load(f)
    catalog_text = format_command_catalog(catalog_obj)

    random.shuffle(data)
    split = max(1, int(len(data) * args.train_ratio))
    train_rows = data[:split]
    dev_rows = data[split:]
    if not dev_rows:
        dev_rows = data[-1:]

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
    baseline_eval = evaluate(baseline, dev_rows, catalog_text, args.prompt_version)

    trainset = to_examples(train_rows, catalog_text, args.prompt_version)
    optimizer = dspy.BootstrapFewShot(metric=metric, max_bootstrapped_demos=4, max_labeled_demos=4)
    optimized = optimizer.compile(ResolverProgram(), trainset=trainset)
    optimized_eval = evaluate(optimized, dev_rows, catalog_text, args.prompt_version)

    gate_passed = (
        optimized_eval["command_accuracy"] >= args.min_command_accuracy
        and optimized_eval["arg_accuracy"] >= args.min_arg_accuracy
    )

    report = {
        "dataset_size": len(data),
        "train_size": len(train_rows),
        "dev_size": len(dev_rows),
        "prompt_version": args.prompt_version,
        "thresholds": {
            "min_command_accuracy": args.min_command_accuracy,
            "min_arg_accuracy": args.min_arg_accuracy,
        },
        "baseline": baseline_eval,
        "optimized": optimized_eval,
        "gate_passed": gate_passed,
    }

    out = Path(args.report_out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"wrote report: {out}")
    print(f"gate_passed={gate_passed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
