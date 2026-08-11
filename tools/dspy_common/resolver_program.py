"""Shared DSPy command resolver used by production and offline optimization."""

from __future__ import annotations

from typing import Any, Iterable, Mapping

import dspy


COMMAND_SELECTION_POLICY = (
    "Select only a command name present in command_catalog. Prioritize targets explicitly named "
    "in the utterance; for example, an explicit PS5 request selects 'start ps5'. Do not infer an "
    "unstated activity or device. Select 'start music' only when the utterance explicitly requests "
    "music playback without a specified artist, song, album, playlist, genre, or other search "
    "target, or explicitly requests random playback. A music request with a specified target "
    "selects 'search and play'. Return an empty string when no catalog command is supported."
)

ARGUMENT_SELECTION_POLICY = (
    "Return arguments exactly as allowed by the selected command's args hint. Never invent an "
    "argument name, prefix, or enum value. For a no-prefix argument, return only its value (for "
    "example 'Meja', not 'keyword:Meja'). For a prefixed argument, use the declared prefix (for "
    "example 'type:artist'). Separate multiple arguments with spaces. Return an empty string when "
    "the command needs no arguments or selected_command is empty."
)


def format_command_catalog(entries: Iterable[Mapping[str, Any]]) -> str:
    """Render command metadata in the canonical production prompt format."""

    lines = []
    for entry in entries:
        name = str(entry.get("name", "")).strip()
        description = str(entry.get("description", "")).strip()
        args_hint = str(entry.get("args", "")).strip()
        if not name:
            continue
        suffix = f" [args: {args_hint}]" if args_hint else ""
        lines.append(f"- {name}: {description}{suffix}")
    return "\n".join(lines)


class ResolveSignature(dspy.Signature):
    """Resolve an utterance to one command from the supplied catalog."""

    utterance = dspy.InputField(desc="User input text")
    command_catalog = dspy.InputField(
        desc="Command catalog with names, descriptions, and optional args format hints"
    )
    prompt_version = dspy.InputField(desc="Prompt version for traceability")
    selected_command = dspy.OutputField(desc=COMMAND_SELECTION_POLICY)
    selected_args = dspy.OutputField(desc=ARGUMENT_SELECTION_POLICY)
    rationale = dspy.OutputField(desc="Short reason based only on explicit evidence in the utterance")


class ResolverProgram(dspy.Module):
    """Single-call resolver shared by the HTTP service and optimization batch.

    A single stage is intentional: it preserves the production API and latency,
    while ensuring the exact program optimized offline is the one executed online.
    """

    def __init__(self) -> None:
        super().__init__()
        self.predict = dspy.Predict(ResolveSignature)

    def forward(self, utterance: str, command_catalog: str, prompt_version: str) -> dspy.Prediction:
        return self.predict(
            utterance=utterance,
            command_catalog=command_catalog,
            prompt_version=prompt_version,
        )
