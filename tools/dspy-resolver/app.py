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


def build_catalog_text(entries: List[CommandEntry]) -> str:
    return "\n".join(f"- {e.name}: {e.description}" for e in entries)


MODEL = os.getenv("MODEL", "openai/gpt-4o-mini")
LM_CONFIGURED = False
try:
    dspy.configure(lm=dspy.LM(MODEL))
    LM_CONFIGURED = True
except Exception:
    # Keep process alive; request handler returns 503 and smarthome can fallback to legacy.
    pass

resolver = ResolverModule()
app = FastAPI(title="smarthome-dspy-resolver", version="0.1.0")


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
