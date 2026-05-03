"""Compatibility wrapper for the shared DSPy LM configuration helper."""

from __future__ import annotations

import sys
from pathlib import Path

TOOLS_DIR = Path(__file__).resolve().parents[1]
if str(TOOLS_DIR) not in sys.path:
    sys.path.insert(0, str(TOOLS_DIR))

from dspy_common.lm_config import build_lm_config  # noqa: E402,F401
