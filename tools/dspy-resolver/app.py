#!/usr/bin/env python3
"""Minimal external DSPy resolver HTTP server."""

from __future__ import annotations

import os
import re
from dataclasses import dataclass
from typing import List

import dspy
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field


class ResolveRequest(BaseModel):
    text: str = Field(..., min_length=1)
    command_list: str = Field(..., min_length=1)
    prompt_version: str = ""


class ResolveResponse(BaseModel):
    command: str
    args: str
    thought: str


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


@dataclass
class CommandEntry:
    name: str
    description: str


def parse_command_list(command_list: str) -> List[CommandEntry]:
    rows: List[CommandEntry] = []
    for line in command_list.splitlines():
        trimmed = line.strip()
        if not trimmed:
            continue
        if trimmed.startswith("args:"):
            continue
        if not line.startswith("  "):
            continue

        # examples:
        #   light on : turn on the light
        #   light on [lo]: turn on the light
        m = re.match(r"^\s*([^\[:]+?)(?:\s*\[[^\]]+\])?\s*:\s*(.+?)\s*$", line)
        if not m:
            continue
        rows.append(CommandEntry(name=m.group(1).strip(), description=m.group(2).strip()))
    return rows


class ResolveSignature(dspy.Signature):
    utterance = dspy.InputField(desc="User input text")
    command_catalog = dspy.InputField(desc="Command catalog list with names and descriptions")
    prompt_version = dspy.InputField(desc="Prompt version for traceability")
    selected_command = dspy.OutputField(desc="Best matching command name. Use empty string if none.")
    selected_args = dspy.OutputField(desc="Args text for selected command. Empty string when no args.")
    rationale = dspy.OutputField(desc="Short reason for selection")


class ResolveMusicIntentSignature(dspy.Signature):
    utterance = dspy.InputField(desc="User input text for music search/play request")
    artist_candidates = dspy.OutputField(desc="Comma separated artist candidates")
    track_candidates = dspy.OutputField(desc="Comma separated track candidates")
    genre_candidates = dspy.OutputField(desc="Comma separated genre candidates")
    must_terms = dspy.OutputField(desc="Comma separated required terms")
    confidence = dspy.OutputField(desc="Confidence score between 0.0 and 1.0")
    ambiguous = dspy.OutputField(desc="true if multiple top candidates remain and autoplay should be avoided")
    reason = dspy.OutputField(desc="Short explanation")


class ResolverModule(dspy.Module):
    def __init__(self) -> None:
        super().__init__()
        self.predict = dspy.Predict(ResolveSignature)

    def forward(self, utterance: str, command_catalog: str, prompt_version: str) -> dspy.Prediction:
        return self.predict(
            utterance=utterance,
            command_catalog=command_catalog,
            prompt_version=prompt_version,
        )


class MusicIntentResolverModule(dspy.Module):
    def __init__(self) -> None:
        super().__init__()
        self.predict = dspy.Predict(ResolveMusicIntentSignature)

    def forward(self, utterance: str) -> dspy.Prediction:
        return self.predict(utterance=utterance)


def build_catalog_text(entries: List[CommandEntry]) -> str:
    return "\n".join(f"- {e.name}: {e.description}" for e in entries)


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
LM_CONFIGURED = False
try:
    dspy.configure(lm=dspy.LM(MODEL))
    LM_CONFIGURED = True
except Exception:
    # Keep process alive; request handler returns 503 and smarthome can fallback to legacy.
    pass

resolver = ResolverModule()
music_intent_resolver = MusicIntentResolverModule()
app = FastAPI(title="smarthome-dspy-resolver", version="0.2.0")


@app.get("/healthz")
def healthz() -> dict:
    if not LM_CONFIGURED:
        raise HTTPException(status_code=503, detail={"status": "not_ready", "model": MODEL})
    return {"status": "ok", "model": MODEL}


@app.post("/resolve", response_model=ResolveResponse)
def resolve(req: ResolveRequest) -> ResolveResponse:
    entries = parse_command_list(req.command_list)
    if not entries:
        raise HTTPException(status_code=400, detail="command_list does not contain command entries")

    if not LM_CONFIGURED or not hasattr(dspy.settings, "lm") or dspy.settings.lm is None:
        raise HTTPException(status_code=503, detail="dspy lm is not configured")

    try:
        pred = resolver(
            utterance=req.text.strip(),
            command_catalog=build_catalog_text(entries),
            prompt_version=req.prompt_version.strip(),
        )
    except Exception as err:
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

    return ResolveResponse(command=command, args=args, thought=thought)


@app.post("/resolve-music-intent", response_model=ResolveMusicIntentResponse)
def resolve_music_intent(req: ResolveMusicIntentRequest) -> ResolveMusicIntentResponse:
    if not LM_CONFIGURED or not hasattr(dspy.settings, "lm") or dspy.settings.lm is None:
        raise HTTPException(status_code=503, detail="dspy lm is not configured")

    try:
        pred = music_intent_resolver(utterance=req.text.strip())
    except Exception as err:
        raise HTTPException(status_code=503, detail=f"dspy resolve music intent failed: {err}") from err

    artists = split_candidates((getattr(pred, "artist_candidates", "") or "").strip())
    tracks = split_candidates((getattr(pred, "track_candidates", "") or "").strip())
    genres = split_candidates((getattr(pred, "genre_candidates", "") or "").strip())
    must_terms = split_candidates((getattr(pred, "must_terms", "") or "").strip())
    confidence = parse_confidence((getattr(pred, "confidence", "") or "").strip())
    ambiguous = parse_bool((getattr(pred, "ambiguous", "") or "").strip())
    reason = (getattr(pred, "reason", "") or "").strip()

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
