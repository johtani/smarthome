"""Pure reporting helpers for DSPy evaluation output."""

from __future__ import annotations

from typing import Any, Dict, Iterable, Mapping


def build_metadata_breakdown(rows: Iterable[Mapping[str, Any]]) -> Dict[str, Dict[str, Dict[str, Any]]]:
    materialized = list(rows)
    return {
        "by_llm_model": _group_accuracy(materialized, "llm_model"),
        "by_prompt_version": _group_accuracy(materialized, "prompt_version"),
    }


def _group_accuracy(rows: list[Mapping[str, Any]], field: str) -> Dict[str, Dict[str, Any]]:
    grouped: Dict[str, Dict[str, int]] = {}
    for row in rows:
        key = str(row.get(field, "") or "").strip() or "(unknown)"
        stats = grouped.setdefault(key, {"count": 0, "command_ok": 0, "args_ok": 0})
        stats["count"] += 1
        if row.get("expected_command", "") == row.get("pred_command", ""):
            stats["command_ok"] += 1
        if row.get("expected_args", "") == row.get("pred_args", ""):
            stats["args_ok"] += 1

    result: Dict[str, Dict[str, Any]] = {}
    for key, stats in sorted(grouped.items()):
        count = stats["count"]
        result[key] = {
            "count": count,
            "command_accuracy": stats["command_ok"] / count,
            "arg_accuracy": stats["args_ok"] / count,
        }
    return result
