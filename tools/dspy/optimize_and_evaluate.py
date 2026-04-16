#!/usr/bin/env python3
"""Run DSPy optimization and offline evaluation with a gate."""

from __future__ import annotations

import argparse
import json
import random
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Dict, List

import dspy


@dataclass
class EvalRow:
    request_id: str
    input_text: str
    expected_command: str
    expected_args: str
    pred_command: str
    pred_args: str


class IntentParse(dspy.Signature):
    utterance = dspy.InputField(desc="User input text")
    normalized_intent = dspy.OutputField(desc="Normalized concise intent")


class CommandSelect(dspy.Signature):
    utterance = dspy.InputField()
    normalized_intent = dspy.InputField()
    command_catalog = dspy.InputField()
    selected_command = dspy.OutputField()
    selection_reason = dspy.OutputField()


class ArgFill(dspy.Signature):
    utterance = dspy.InputField()
    selected_command = dspy.InputField()
    args_hint = dspy.InputField()
    filled_args = dspy.OutputField()


class SafetyCheck(dspy.Signature):
    selected_command = dspy.InputField()
    filled_args = dspy.InputField()
    safety_decision = dspy.OutputField(desc="allow or reject")


class ResolverProgram(dspy.Module):
    def __init__(self) -> None:
        super().__init__()
        self.intent = dspy.Predict(IntentParse)
        self.command = dspy.Predict(CommandSelect)
        self.args = dspy.Predict(ArgFill)
        self.safety = dspy.Predict(SafetyCheck)

    def forward(self, utterance: str, command_catalog: str) -> dspy.Prediction:
        i = self.intent(utterance=utterance)
        c = self.command(
            utterance=utterance,
            normalized_intent=i.normalized_intent,
            command_catalog=command_catalog,
        )
        a = self.args(
            utterance=utterance,
            selected_command=c.selected_command,
            args_hint="Return empty string when no args are required.",
        )
        s = self.safety(selected_command=c.selected_command, filled_args=a.filled_args)
        if (s.safety_decision or "").strip().lower() == "reject":
            return dspy.Prediction(selected_command="", filled_args="", safety_decision="reject")
        return dspy.Prediction(
            selected_command=(c.selected_command or "").strip(),
            filled_args=(a.filled_args or "").strip(),
            safety_decision="allow",
        )


def read_jsonl(path: Path) -> List[Dict[str, Any]]:
    rows: List[Dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            rows.append(json.loads(line))
    return rows


def to_examples(rows: List[Dict[str, Any]], command_catalog: str) -> List[dspy.Example]:
    ex = []
    for r in rows:
        ex.append(
            dspy.Example(
                request_id=r.get("request_id", ""),
                utterance=r["input_text"],
                command_catalog=command_catalog,
                expected_command=r.get("expected_command", ""),
                expected_args=r.get("expected_args", ""),
            ).with_inputs("utterance", "command_catalog")
        )
    return ex


def metric(example: dspy.Example, pred: dspy.Prediction, trace: Any = None) -> float:
    exp_cmd = (example.expected_command or "").strip().lower()
    exp_args = (example.expected_args or "").strip()
    got_cmd = (pred.selected_command or "").strip().lower()
    got_args = (pred.filled_args or "").strip()

    cmd_score = 1.0 if exp_cmd == got_cmd else 0.0
    arg_score = 1.0 if exp_args == got_args else 0.0
    if exp_cmd == "":
        return 1.0 if got_cmd == "" else 0.0
    return (cmd_score * 0.8) + (arg_score * 0.2)


def evaluate(program: ResolverProgram, rows: List[Dict[str, Any]], command_catalog: str) -> Dict[str, Any]:
    if not rows:
        return {"count": 0, "command_accuracy": 0.0, "arg_accuracy": 0.0, "rows": []}

    eval_rows: List[EvalRow] = []
    cmd_ok = 0
    arg_ok = 0
    for r in rows:
        pred = program(utterance=r["input_text"], command_catalog=command_catalog)
        exp_cmd = (r.get("expected_command") or "").strip()
        exp_args = (r.get("expected_args") or "").strip()
        got_cmd = (pred.selected_command or "").strip()
        got_args = (pred.filled_args or "").strip()

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
            )
        )

    n = len(rows)
    return {
        "count": n,
        "command_accuracy": cmd_ok / n,
        "arg_accuracy": arg_ok / n,
        "rows": [asdict(x) for x in eval_rows[:20]],
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dataset-jsonl", required=True)
    parser.add_argument("--command-catalog", required=True)
    parser.add_argument("--model", required=True, help="e.g. openai/gpt-4o-mini")
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
    catalog_text = json.dumps(catalog_obj, ensure_ascii=False)

    random.shuffle(data)
    split = max(1, int(len(data) * args.train_ratio))
    train_rows = data[:split]
    dev_rows = data[split:]
    if not dev_rows:
        dev_rows = data[-1:]

    dspy.configure(lm=dspy.LM(args.model))

    baseline = ResolverProgram()
    baseline_eval = evaluate(baseline, dev_rows, catalog_text)

    trainset = to_examples(train_rows, catalog_text)
    optimizer = dspy.BootstrapFewShot(metric=metric, max_bootstrapped_demos=4, max_labeled_demos=4)
    optimized = optimizer.compile(ResolverProgram(), trainset=trainset)
    optimized_eval = evaluate(optimized, dev_rows, catalog_text)

    gate_passed = (
        optimized_eval["command_accuracy"] >= args.min_command_accuracy
        and optimized_eval["arg_accuracy"] >= args.min_arg_accuracy
    )

    report = {
        "dataset_size": len(data),
        "train_size": len(train_rows),
        "dev_size": len(dev_rows),
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
