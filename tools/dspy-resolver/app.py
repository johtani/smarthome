#!/usr/bin/env python3
"""Minimal external DSPy resolver HTTP server."""

from __future__ import annotations

import os
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List

import dspy
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

TOOLS_DIR = Path(__file__).resolve().parents[1]
if str(TOOLS_DIR) not in sys.path:
    sys.path.insert(0, str(TOOLS_DIR))

from dspy_common.artifacts import load_artifact
from dspy_common.resolver_program import format_command_catalog

from lm_config import build_lm_config
from otel_config import (
    add_event,
    clean_attributes,
    instrument_fastapi_app,
    set_error,
    setup_otel,
    start_span,
    text_trace_attrs,
)

class ResolveRequest(BaseModel):
    text: str = Field(..., min_length=1)
    command_list: str = Field(..., min_length=1)
    prompt_version: str = ""


class ResolveResponse(BaseModel):
    command: str
    args: str
    thought: str
    model: str = ""
    prompt_version: str = ""
    artifact_version: str = ""
    dataset_version: str = ""


class ResolveMusicIntentRequest(BaseModel):
    text: str = Field(..., min_length=1)


class ResolveMusicIntentResponse(BaseModel):
    artist_candidates: List[str] = Field(default_factory=list)
    track_candidates: List[str] = Field(default_factory=list)
    genre_candidates: List[str] = Field(default_factory=list)
    must_terms: List[str] = Field(default_factory=list)
    confidence: float = 0.0
    ambiguous: bool = False
    reason: str = ""
    model: str = ""


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


class ResolveMusicIntentSignature(dspy.Signature):
    utterance = dspy.InputField(desc="User input text for music search/play request")
    artist_candidates = dspy.OutputField(desc="Comma separated artist candidates")
    track_candidates = dspy.OutputField(desc="Comma separated track candidates")
    genre_candidates = dspy.OutputField(desc="Comma separated genre candidates")
    must_terms = dspy.OutputField(desc="Comma separated required terms")
    confidence = dspy.OutputField(desc="Confidence score between 0.0 and 1.0")
    ambiguous = dspy.OutputField(desc="true if multiple top candidates remain and autoplay should be avoided")
    reason = dspy.OutputField(desc="Short explanation")


class MusicIntentResolverModule(dspy.Module):
    def __init__(self) -> None:
        super().__init__()
        self.predict = dspy.Predict(ResolveMusicIntentSignature)

    def forward(self, utterance: str) -> dspy.Prediction:
        return self.predict(utterance=utterance)


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


def split_candidates(value: str) -> List[str]:
    if not value:
        return []
    parts = re.split(r"[,\n/、|]+", value)
    seen = set()
    items: List[str] = []
    for part in parts:
        normalized = part.strip()
        if not normalized or normalized in seen:
            continue
        seen.add(normalized)
        items.append(normalized)
        if len(items) >= 8:
            break
    return items


def parse_confidence(value: str) -> float:
    try:
        score = float(value.strip())
    except Exception:
        return 0.0
    if score < 0:
        return 0.0
    if score > 1:
        return 1.0
    return score


def parse_bool(value: str) -> bool:
    v = (value or "").strip().lower()
    return v in {"true", "1", "yes", "y"}


MODEL = os.getenv("MODEL", "openai/gpt-4o-mini")
ARTIFACT_DIR = os.getenv("DSPY_ARTIFACT_DIR", "").strip()
EXPECTED_PROMPT_VERSION = os.getenv("DSPY_PROMPT_VERSION", "").strip()
ARTIFACT_VERSION = ""
DATASET_VERSION = ""
ARTIFACT_LOADED = False
ARTIFACT_LOAD_ERROR = ""
LM_HEALTH: Dict[str, Any] = {
    "model": MODEL,
    "api_base": "",
    "model_type": "chat",
    "temperature": None,
    "max_tokens": None,
    "api_key_source": "none",
}
LM_INIT_ERROR = ""
LM_CONFIGURED = False
try:
    lm_conf = build_lm_config()
    MODEL = lm_conf["model"]
    LM_HEALTH = lm_conf["health"]
    dspy.configure(lm=dspy.LM(MODEL, **lm_conf["kwargs"]))
    LM_CONFIGURED = True
except Exception as err:
    LM_INIT_ERROR = str(err)
    # Keep process alive; request handler returns 503 and smarthome can fallback to legacy.
    pass

artifact_result = load_artifact(
    Path(ARTIFACT_DIR) if ARTIFACT_DIR else None,
    expected_model=MODEL,
    expected_prompt_version=EXPECTED_PROMPT_VERSION,
)
resolver = artifact_result.program
ARTIFACT_LOADED = artifact_result.loaded
ARTIFACT_VERSION = artifact_result.artifact_version
DATASET_VERSION = artifact_result.dataset_version
ARTIFACT_LOAD_ERROR = artifact_result.error
music_intent_resolver = MusicIntentResolverModule()
setup_otel()
app = FastAPI(title="smarthome-dspy-resolver", version="0.2.0")
instrument_fastapi_app(app)


def lm_trace_attrs() -> Dict[str, Any]:
    return clean_attributes(
        {
            "dspy.lm.model": MODEL,
            "dspy.lm.api_base": LM_HEALTH.get("api_base", ""),
            "dspy.lm.model_type": LM_HEALTH.get("model_type", ""),
            "dspy.lm.temperature": LM_HEALTH.get("temperature"),
            "dspy.lm.max_tokens": LM_HEALTH.get("max_tokens"),
            "dspy.lm.api_key_source": LM_HEALTH.get("api_key_source", "none"),
            "resolver.artifact_version": ARTIFACT_VERSION,
            "resolver.dataset_version": DATASET_VERSION,
            "resolver.artifact_loaded": ARTIFACT_LOADED,
        }
    )


@app.get("/healthz")
def healthz() -> dict:
    if not LM_CONFIGURED:
        raise HTTPException(
            status_code=503,
            detail={
                "status": "not_ready",
                "model": MODEL,
                "api_base": LM_HEALTH["api_base"],
                "model_type": LM_HEALTH["model_type"],
                "temperature": LM_HEALTH["temperature"],
                "max_tokens": LM_HEALTH["max_tokens"],
                "api_key_source": LM_HEALTH["api_key_source"],
                "artifact_version": ARTIFACT_VERSION,
                "dataset_version": DATASET_VERSION,
                "artifact_loaded": ARTIFACT_LOADED,
                "artifact_load_error": ARTIFACT_LOAD_ERROR,
                "error": LM_INIT_ERROR,
            },
        )
    return {
        "status": "ok",
        "model": MODEL,
        "api_base": LM_HEALTH["api_base"],
        "model_type": LM_HEALTH["model_type"],
        "temperature": LM_HEALTH["temperature"],
        "max_tokens": LM_HEALTH["max_tokens"],
        "api_key_source": LM_HEALTH["api_key_source"],
        "artifact_version": ARTIFACT_VERSION,
        "dataset_version": DATASET_VERSION,
        "artifact_loaded": ARTIFACT_LOADED,
        "artifact_load_error": ARTIFACT_LOAD_ERROR,
    }


@app.post("/resolve", response_model=ResolveResponse)
def resolve(req: ResolveRequest) -> ResolveResponse:
    prompt_version = req.prompt_version.strip()
    text = req.text.strip()
    with start_span(
        "dspy.resolve.command",
        {
            "resolver.path": "dspy",
            "resolver.prompt_version": prompt_version,
            **lm_trace_attrs(),
            **text_trace_attrs("resolver.input", req.text),
            **text_trace_attrs("resolver.command_list", req.command_list),
        },
    ) as span:
        entries = parse_command_list(req.command_list)
        span.set_attributes(clean_attributes({"command_catalog.count": len(entries)}))
        if not entries:
            set_error(span, "command_list does not contain command entries")
            raise HTTPException(status_code=400, detail="command_list does not contain command entries")

        if not LM_CONFIGURED or not hasattr(dspy.settings, "lm") or dspy.settings.lm is None:
            set_error(span, "dspy lm is not configured")
            raise HTTPException(status_code=503, detail="dspy lm is not configured")

        command_catalog = build_catalog_text(entries)
        add_event(
            span,
            "dspy.request",
            {
                "dspy.signature": "ResolveSignature",
                "resolver.prompt_version": prompt_version,
                "command_catalog.count": len(entries),
                **lm_trace_attrs(),
                **text_trace_attrs("dspy.request.utterance", text),
                **text_trace_attrs("dspy.request.command_catalog", command_catalog),
            },
        )

        try:
            pred = resolver(
                utterance=text,
                command_catalog=command_catalog,
                prompt_version=prompt_version,
            )
        except Exception as err:
            set_error(span, err)
            raise HTTPException(status_code=503, detail=f"dspy resolve failed: {err}") from err

        command = (getattr(pred, "selected_command", "") or "").strip()
        args = (getattr(pred, "selected_args", "") or "").strip()
        thought = (getattr(pred, "rationale", "") or "").strip()

        # Safety: only allow commands that exist in the provided catalog.
        allowed = {e.name for e in entries}
        if command not in allowed:
            command = ""
            args = ""
            if not thought:
                thought = "no compatible command in catalog"

        span.set_attributes(
            clean_attributes(
                {
                    "resolver.selected_command": command,
                    "resolver.args.length": len(args),
                    "resolver.thought.length": len(thought),
                    "resolver.resolved": command != "",
                }
            )
        )
        add_event(
            span,
            "dspy.response",
            {
                "resolver.selected_command": command,
                "resolver.args.length": len(args),
                "resolver.thought.length": len(thought),
                "resolver.resolved": command != "",
                "llm.model": MODEL,
                "resolver.prompt_version": prompt_version,
                "resolver.artifact_version": ARTIFACT_VERSION,
                "resolver.dataset_version": DATASET_VERSION,
            },
        )

        return ResolveResponse(
            command=command,
            args=args,
            thought=thought,
            model=MODEL,
            prompt_version=prompt_version,
            artifact_version=ARTIFACT_VERSION,
            dataset_version=DATASET_VERSION,
        )


@app.post("/resolve-music-intent", response_model=ResolveMusicIntentResponse)
def resolve_music_intent(req: ResolveMusicIntentRequest) -> ResolveMusicIntentResponse:
    text = req.text.strip()
    with start_span(
        "dspy.resolve.music_intent",
        {
            "resolver.path": "dspy_music_intent",
            **lm_trace_attrs(),
            **text_trace_attrs("music_intent.input", req.text),
        },
    ) as span:
        if not LM_CONFIGURED or not hasattr(dspy.settings, "lm") or dspy.settings.lm is None:
            set_error(span, "dspy lm is not configured")
            raise HTTPException(status_code=503, detail="dspy lm is not configured")

        add_event(
            span,
            "dspy.request",
            {
                "dspy.signature": "ResolveMusicIntentSignature",
                **lm_trace_attrs(),
                **text_trace_attrs("dspy.request.utterance", text),
            },
        )

        try:
            pred = music_intent_resolver(utterance=text)
        except Exception as err:
            set_error(span, err)
            raise HTTPException(status_code=503, detail=f"dspy resolve music intent failed: {err}") from err

        artists = split_candidates((getattr(pred, "artist_candidates", "") or "").strip())
        tracks = split_candidates((getattr(pred, "track_candidates", "") or "").strip())
        genres = split_candidates((getattr(pred, "genre_candidates", "") or "").strip())
        must_terms = split_candidates((getattr(pred, "must_terms", "") or "").strip())
        confidence = parse_confidence((getattr(pred, "confidence", "") or "").strip())
        ambiguous = parse_bool((getattr(pred, "ambiguous", "") or "").strip())
        reason = (getattr(pred, "reason", "") or "").strip()

        response_attrs = {
            "music_intent.artist_candidates.count": len(artists),
            "music_intent.track_candidates.count": len(tracks),
            "music_intent.genre_candidates.count": len(genres),
            "music_intent.must_terms.count": len(must_terms),
            "music_intent.confidence": confidence,
            "music_intent.ambiguous": ambiguous,
            "music_intent.reason.length": len(reason),
        }
        span.set_attributes(clean_attributes(response_attrs))
        add_event(span, "dspy.response", response_attrs)

        return ResolveMusicIntentResponse(
            artist_candidates=artists,
            track_candidates=tracks,
            genre_candidates=genres,
            must_terms=must_terms,
            confidence=confidence,
            ambiguous=ambiguous,
            reason=reason,
            model=MODEL,
        )
