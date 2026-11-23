"""Convert Go command help into the DSPy catalog."""

import re
from dataclasses import dataclass
from typing import List

from .resolver_program import format_command_catalog


_ARGS_PREFIX = "args:"


@dataclass
class CommandEntry:
    name: str
    description: str
    args_text: str = ""


def parse_command_list(command_list: str) -> List[CommandEntry]:
    rows: List[CommandEntry] = []
    last_entry: CommandEntry | None = None
    for line in command_list.splitlines():
        trimmed = line.strip()
        if not trimmed:
            continue
        if trimmed.startswith(_ARGS_PREFIX):
            if last_entry is not None:
                last_entry.args_text = trimmed[len(_ARGS_PREFIX):].strip()
            continue
        if not line.startswith("  "):
            continue

        # examples:
        #   light on : turn on the light
        #   light on [lo]: turn on the light
        m = re.match(r"^\s*([^\[:]+?)(?:\s*\[[^\]]+\])?\s*:\s*(.+?)\s*$", line)
        if not m:
            continue
        entry = CommandEntry(name=m.group(1).strip(), description=m.group(2).strip())
        rows.append(entry)
        last_entry = entry
    return rows


def build_catalog_text(entries: List[CommandEntry]) -> str:
    catalog_entries = []
    for e in entries:
        args_hint = e.args_text
        if e.args_text:
            args_hint = re.sub(
                r"\(([^)]*)\)",
                lambda m: "(" + m.group(1) + (",no-prefix)" if "prefix=" not in m.group(1) else ")"),
                e.args_text,
            )
        catalog_entries.append({"name": e.name, "description": e.description, "args": args_hint})
    return format_command_catalog(catalog_entries)
