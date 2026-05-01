"""LM configuration helpers for dspy-resolver."""

from __future__ import annotations

import os
from typing import Any, Dict, Mapping, Optional


def _get_str(env: Mapping[str, str], key: str, default: str = "") -> str:
    return (env.get(key, default) or "").strip()


def _parse_optional_float_strict(value: str, key: str) -> Optional[float]:
    if not value:
        return None
    try:
        return float(value)
    except Exception as err:
        raise ValueError(f"{key} must be a number: {value}") from err


def _parse_optional_int_strict(value: str, key: str) -> Optional[int]:
    if not value:
        return None
    try:
        return int(value)
    except Exception as err:
        raise ValueError(f"{key} must be an integer: {value}") from err


def build_lm_config(env: Optional[Mapping[str, str]] = None) -> Dict[str, Any]:
    src = env if env is not None else os.environ

    model = _get_str(src, "MODEL", "openai/gpt-4o-mini")
    api_base = _get_str(src, "LM_API_BASE")
    lm_api_key = _get_str(src, "LM_API_KEY")
    openai_api_key = _get_str(src, "OPENAI_API_KEY")
    model_type = _get_str(src, "LM_MODEL_TYPE", "chat")
    temperature_raw = _get_str(src, "LM_TEMPERATURE")
    max_tokens_raw = _get_str(src, "LM_MAX_TOKENS")

    temperature = _parse_optional_float_strict(temperature_raw, "LM_TEMPERATURE")
    max_tokens = _parse_optional_int_strict(max_tokens_raw, "LM_MAX_TOKENS")

    api_key_source = "none"
    api_key = ""
    if lm_api_key:
        api_key = lm_api_key
        api_key_source = "LM_API_KEY"
    elif openai_api_key:
        api_key = openai_api_key
        api_key_source = "OPENAI_API_KEY"

    kwargs: Dict[str, Any] = {}
    if api_base:
        kwargs["api_base"] = api_base
    if api_key:
        kwargs["api_key"] = api_key
    if model_type:
        kwargs["model_type"] = model_type
    if temperature is not None:
        kwargs["temperature"] = temperature
    if max_tokens is not None:
        kwargs["max_tokens"] = max_tokens

    return {
        "model": model,
        "kwargs": kwargs,
        "health": {
            "model": model,
            "api_base": api_base,
            "model_type": model_type,
            "temperature": temperature,
            "max_tokens": max_tokens,
            "api_key_source": api_key_source,
        },
    }
