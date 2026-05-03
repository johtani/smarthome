"""LM configuration helpers shared by DSPy tools."""

from __future__ import annotations

import os
from typing import Any, Dict, Mapping, Optional


def _get_str(env: Mapping[str, str], key: str, default: str = "") -> str:
    return (env.get(key, default) or "").strip()


def _override_str(overrides: Mapping[str, Any], key: str) -> str:
    value = overrides.get(key)
    if value is None:
        return ""
    return str(value).strip()


def _parse_optional_float_strict(value: Any, key: str) -> Optional[float]:
    if value is None or value == "":
        return None
    try:
        return float(value)
    except Exception as err:
        raise ValueError(f"{key} must be a number: {value}") from err


def _parse_optional_int_strict(value: Any, key: str) -> Optional[int]:
    if value is None or value == "":
        return None
    try:
        return int(value)
    except Exception as err:
        raise ValueError(f"{key} must be an integer: {value}") from err


def _override_or_env(
    overrides: Mapping[str, Any],
    override_key: str,
    env: Mapping[str, str],
    env_key: str,
    default: str = "",
) -> str:
    override_value = _override_str(overrides, override_key)
    if override_value:
        return override_value
    return _get_str(env, env_key, default)


def build_lm_config(
    env: Optional[Mapping[str, str]] = None,
    overrides: Optional[Mapping[str, Any]] = None,
) -> Dict[str, Any]:
    """Build dspy.LM configuration from environment values and optional overrides."""

    src = env if env is not None else os.environ
    override_values = overrides or {}

    model = _override_or_env(override_values, "model", src, "MODEL", "openai/gpt-4o-mini")
    api_base = _override_or_env(override_values, "api_base", src, "LM_API_BASE")
    model_type = _override_or_env(override_values, "model_type", src, "LM_MODEL_TYPE", "chat")

    temperature_raw: Any = override_values.get("temperature")
    if temperature_raw is None or temperature_raw == "":
        temperature_raw = _get_str(src, "LM_TEMPERATURE")
    max_tokens_raw: Any = override_values.get("max_tokens")
    if max_tokens_raw is None or max_tokens_raw == "":
        max_tokens_raw = _get_str(src, "LM_MAX_TOKENS")

    temperature = _parse_optional_float_strict(temperature_raw, "LM_TEMPERATURE")
    max_tokens = _parse_optional_int_strict(max_tokens_raw, "LM_MAX_TOKENS")

    api_key_source = "none"
    api_key = _override_str(override_values, "api_key")
    if api_key:
        api_key_source = "cli"
    else:
        lm_api_key = _get_str(src, "LM_API_KEY")
        openai_api_key = _get_str(src, "OPENAI_API_KEY")
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
