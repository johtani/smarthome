"""Save and safely load versioned DSPy resolver artifacts."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Mapping

from dspy_common.resolver_program import ResolverProgram


PROGRAM_SCHEMA_VERSION = "resolver-program-v1"
MANIFEST_FILE = "manifest.json"
PROGRAM_FILE = "program.json"


@dataclass(frozen=True)
class ArtifactLoadResult:
    program: ResolverProgram
    loaded: bool = False
    artifact_version: str = ""
    dataset_version: str = ""
    error: str = ""


def save_artifact(directory: Path, program: ResolverProgram, manifest: Mapping[str, Any]) -> None:
    """Persist an accepted program and its manifest in one artifact directory."""

    directory.mkdir(parents=True, exist_ok=True)
    program.save(str(directory / PROGRAM_FILE))
    (directory / MANIFEST_FILE).write_text(
        json.dumps(dict(manifest), ensure_ascii=False, indent=2), encoding="utf-8"
    )


def load_artifact(
    directory: Path | None,
    *,
    expected_model: str,
    expected_prompt_version: str,
) -> ArtifactLoadResult:
    """Load only a compatible, gate-approved artifact; otherwise return a fresh program."""

    fallback = ResolverProgram()
    if directory is None:
        return ArtifactLoadResult(program=fallback, error="artifact directory is not configured")

    try:
        manifest = json.loads((directory / MANIFEST_FILE).read_text(encoding="utf-8"))
        if manifest.get("program_schema_version") != PROGRAM_SCHEMA_VERSION:
            raise ValueError("program schema version mismatch")
        if manifest.get("model") != expected_model:
            raise ValueError("model mismatch")
        if not expected_prompt_version:
            raise ValueError("expected prompt version is not configured")
        if manifest.get("prompt_version") != expected_prompt_version:
            raise ValueError("prompt version mismatch")
        if manifest.get("gate_passed") is not True:
            raise ValueError("artifact did not pass the evaluation gate")

        program = ResolverProgram()
        program.load(str(directory / PROGRAM_FILE))
        return ArtifactLoadResult(
            program=program,
            loaded=True,
            artifact_version=str(manifest.get("artifact_version", "")),
            dataset_version=str(manifest.get("dataset_version", "")),
        )
    except Exception as err:
        return ArtifactLoadResult(program=fallback, error=str(err))
