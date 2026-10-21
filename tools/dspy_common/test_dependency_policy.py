"""Guard the LiteLLM security policy for both installation entrypoints."""

from importlib.metadata import version
from pathlib import Path

import pytest
from packaging.requirements import Requirement


TOOLS_DIR = Path(__file__).resolve().parents[1]


@pytest.mark.parametrize("tool", ["dspy", "dspy-resolver"])
def test_litellm_dependency_policy(tool):
    requirements = [
        Requirement(line)
        for line in (TOOLS_DIR / tool / "requirements.txt").read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]
    litellm = next(requirement for requirement in requirements if requirement.name == "litellm")

    for allowed in ("1.106.0", "1.106.2"):
        assert allowed in litellm.specifier
    for blocked in ("1.82.7", "1.82.8", "1.106.3", "1.107.0", "2.0.0"):
        assert blocked not in litellm.specifier

    # Retain explicit exclusions even when the upper bound is raised in future.
    for compromised in ("1.82.7", "1.82.8", "1.107.0"):
        assert f"!={compromised}" in {str(specifier) for specifier in litellm.specifier}

    assert version("litellm") in litellm.specifier
