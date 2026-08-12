#!/usr/bin/env python3
"""Generate deterministic resolver review candidates from templates."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import tempfile
from pathlib import Path
from typing import Any

SCHEMA_VERSION = "resolver-review-candidate/v1"
SOURCE = "synthetic_template"
GENERATOR = "resolver-synthetic/v1"


def load_json(path: Path) -> Any:
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def stable_id(prefix: str, value: Any) -> str:
    encoded = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return f"{prefix}-{hashlib.sha256(encoded.encode()).hexdigest()[:16]}"


def read_catalog(path: Path) -> set[str]:
    value = load_json(path)
    if not isinstance(value, list):
        raise ValueError("command catalog must be a JSON array")
    names = set()
    for number, item in enumerate(value, 1):
        if not isinstance(item, dict) or not str(item.get("name", "")).strip():
            raise ValueError(f"catalog entry {number}: name is required")
        names.add(str(item["name"]).strip())
    return names


def nonempty_strings(value: Any, field: str) -> list[str]:
    if not isinstance(value, list) or not value or any(not isinstance(item, str) or not item for item in value):
        raise ValueError(f"{field} must be a non-empty array of non-empty strings")
    return value


def generate(config: Any, catalog_names: set[str]) -> list[dict[str, Any]]:
    if not isinstance(config, dict) or not isinstance(config.get("groups"), list):
        raise ValueError("config.groups must be an array")
    rows = []
    seen_text_targets: set[tuple[str, str, str]] = set()
    for group_number, group in enumerate(config["groups"], 1):
        if not isinstance(group, dict):
            raise ValueError(f"group {group_number} must be an object")
        command = str(group.get("expected_command", "")).strip()
        args = str(group.get("expected_args", "")).strip()
        if command and command not in catalog_names:
            raise ValueError(f"group {group_number}: expected_command is not in catalog: {command}")
        targets = nonempty_strings(group.get("targets", [""]), f"group {group_number}.targets")
        variants = group.get("variants")
        if not isinstance(variants, list) or not variants:
            raise ValueError(f"group {group_number}.variants must be a non-empty array")
        group_seed = {"command": command, "args": args, "targets": targets, "variants": variants}
        group_id = stable_id("synthetic-group", group_seed)
        for variant_number, variant in enumerate(variants, 1):
            if not isinstance(variant, dict):
                raise ValueError(f"group {group_number} variant {variant_number} must be an object")
            scenario = str(variant.get("scenario", "")).strip()
            variant_command = str(variant.get("expected_command", command)).strip()
            variant_args = str(variant.get("expected_args", args)).strip()
            if variant_command and variant_command not in catalog_names:
                raise ValueError(
                    f"group {group_number} variant {variant_number}: "
                    f"expected_command is not in catalog: {variant_command}"
                )
            templates = nonempty_strings(variant.get("templates"), f"group {group_number} variant {variant_number}.templates")
            if not scenario:
                raise ValueError(f"group {group_number} variant {variant_number}: scenario is required")
            for target in targets:
                for template in templates:
                    try:
                        input_text = template.format(target=target)
                    except (KeyError, ValueError) as exc:
                        raise ValueError(f"invalid template {template!r}: only {{target}} is supported") from exc
                    key = (input_text, variant_command, variant_args)
                    if key in seen_text_targets:
                        continue
                    seen_text_targets.add(key)
                    identity = {
                        "group_id": group_id,
                        "scenario": scenario,
                        "input_text": input_text,
                        "command": variant_command,
                        "args": variant_args,
                    }
                    rows.append({
                        "schema_version": SCHEMA_VERSION,
                        "case_id": stable_id("synthetic-case", identity),
                        "group_id": group_id,
                        "source": SOURCE,
                        "generator": GENERATOR,
                        "scenario": scenario,
                        "input_text": input_text,
                        "candidate_command": variant_command,
                        "candidate_args": variant_args,
                        "expected_command": variant_command,
                        "expected_args": variant_args,
                        "review_status": "unreviewed",
                        "reviewed": False,
                    })
    return rows


def write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8", newline="\n") as handle:
            for row in rows:
                handle.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")
        os.replace(temporary, path)
    except BaseException:
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass
        raise


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--catalog", type=Path, required=True, help="command catalog JSON")
    parser.add_argument("--config", type=Path, required=True, help="synthetic template configuration JSON")
    parser.add_argument("--output", type=Path, required=True, help="candidate JSONL output")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    rows = generate(load_json(args.config), read_catalog(args.catalog))
    write_jsonl(args.output, rows)
    print(f"generated {len(rows)} unreviewed candidates: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
