#!/usr/bin/env python3
"""Build DSPy training/evaluation dataset from resolver event CSV."""

from __future__ import annotations

import argparse
import csv
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional


@dataclass
class RequestAggregate:
    request_id: str
    input_text: str = ""
    resolved_command: str = ""
    resolved_args: str = ""
    feedback_label: str = ""
    feedback_correction: str = ""
    event_count: int = 0


def normalize(s: Optional[str]) -> str:
    return (s or "").strip()


def split_correction_target(correction: str, known_commands: set[str]) -> tuple[str, str]:
    text = normalize(correction)
    if not text:
        return "", ""

    matches = [cmd for cmd in known_commands if text == cmd or text.startswith(cmd + " ")]
    if matches:
        cmd = max(matches, key=len)
        return cmd, normalize(text[len(cmd):])

    parts = text.split(maxsplit=1)
    return parts[0], parts[1] if len(parts) > 1 else ""


def build_dataset_rows(csv_path: Path) -> List[dict]:
    grouped: Dict[str, RequestAggregate] = {}
    known_commands: set[str] = set()
    with csv_path.open("r", encoding="utf-8-sig", newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            request_id = normalize(row.get("resolver_request_id"))
            if not request_id:
                continue

            agg = grouped.setdefault(request_id, RequestAggregate(request_id=request_id))
            agg.event_count += 1

            input_text = normalize(row.get("input_text"))
            if input_text:
                agg.input_text = input_text

            event_name = normalize(row.get("event_name"))
            command = normalize(row.get("resolver_resolved_command"))
            args = normalize(row.get("resolver_resolved_args"))
            if command:
                known_commands.add(command)
            if event_name in ("resolver.decision", "resolver.execution"):
                if command:
                    agg.resolved_command = command
                if args:
                    agg.resolved_args = args

            feedback_label = normalize(row.get("feedback_label"))
            feedback_correction = normalize(row.get("feedback_correction"))
            if event_name == "resolver.feedback":
                if feedback_label:
                    agg.feedback_label = feedback_label
                if feedback_correction:
                    agg.feedback_correction = feedback_correction

    dataset = []
    for agg in grouped.values():
        if not agg.input_text:
            # Without natural-language text, DSPy optimization data is not useful.
            continue

        expected_command = agg.resolved_command
        expected_args = agg.resolved_args

        # If explicit negative feedback has correction, treat it as preferred target.
        if agg.feedback_label == "incorrect" and agg.feedback_correction:
            expected_command, expected_args = split_correction_target(agg.feedback_correction, known_commands)

        dataset.append(
            {
                "request_id": agg.request_id,
                "input_text": agg.input_text,
                "expected_command": expected_command,
                "expected_args": expected_args,
                "feedback_label": agg.feedback_label or "skip",
            }
        )
    return dataset


def write_jsonl(rows: List[dict], out_path: Path) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with out_path.open("w", encoding="utf-8") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input-csv", required=True)
    parser.add_argument("--output-jsonl", required=True)
    parser.add_argument("--min-row-per-request", type=int, default=1)
    args = parser.parse_args()

    rows = build_dataset_rows(Path(args.input_csv))
    if args.min_row_per_request > 1:
        # Reserved for future filtering; currently dataset rows are already per request.
        pass
    write_jsonl(rows, Path(args.output_jsonl))
    print(f"wrote {len(rows)} rows -> {args.output_jsonl}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
