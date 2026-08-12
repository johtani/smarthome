#!/usr/bin/env python3
"""Local-only UI for reviewing resolver dataset candidates."""

from __future__ import annotations

import argparse
import csv
import json
import os
import tempfile
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Any
from urllib.parse import urlsplit

REVIEW_SCHEMA_VERSION = "resolver-reviewed/v1"
STATUSES = {"unreviewed", "accepted", "corrected", "excluded", "pending"}


def text(value: Any) -> str:
    return "" if value is None else str(value)


def read_rows(path: Path) -> list[dict[str, Any]]:
    if path.suffix.lower() == ".csv":
        with path.open(encoding="utf-8-sig", newline="") as handle:
            rows = list(csv.DictReader(handle))
    else:
        rows = []
        with path.open(encoding="utf-8") as handle:
            for line_number, line in enumerate(handle, 1):
                if not line.strip():
                    continue
                try:
                    value = json.loads(line)
                except json.JSONDecodeError as exc:
                    raise ValueError(f"{path}:{line_number}: invalid JSON: {exc.msg}") from exc
                if not isinstance(value, dict):
                    raise ValueError(f"{path}:{line_number}: each row must be a JSON object")
                rows.append(value)
    seen: set[str] = set()
    for number, row in enumerate(rows, 1):
        case_id = text(row.get("case_id") or row.get("request_id"))
        if not case_id:
            raise ValueError(f"row {number}: case_id is required")
        if case_id in seen:
            raise ValueError(f"row {number}: duplicate case_id: {case_id}")
        seen.add(case_id)
        row["case_id"] = case_id
        row.setdefault("group_id", case_id)
        row.setdefault("source", "fixture" if not row.get("trace_id") else "resolver-log")
    return rows


def read_catalog(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8") as handle:
        value = json.load(handle)
    if not isinstance(value, list):
        raise ValueError("command catalog must be a JSON array")
    catalog = []
    for number, item in enumerate(value, 1):
        if not isinstance(item, dict) or not text(item.get("name")).strip():
            raise ValueError(f"catalog entry {number}: name is required")
        catalog.append({
            "name": text(item["name"]).strip(),
            "description": text(item.get("description")),
            "args": text(item.get("args")),
        })
    return catalog


def initial_review(row: dict[str, Any]) -> dict[str, Any]:
    status = text(row.get("review_status")) or "unreviewed"
    if status not in STATUSES:
        status = "unreviewed"
    return {
        "review_status": status,
        "reviewed": bool(row.get("reviewed", status != "unreviewed")),
        "expected_command": text(row.get("expected_command")),
        "expected_args": text(row.get("expected_args")),
        "reviewed_at": text(row.get("reviewed_at")),
        "review_note": text(row.get("review_note")),
    }


def load_reviews(path: Path) -> dict[str, dict[str, Any]]:
    if not path.exists():
        return {}
    return {row["case_id"]: initial_review(row) for row in read_rows(path)}


def catalog_violation(command: str, catalog_names: set[str]) -> bool:
    return bool(command) and command not in catalog_names


class ReviewStore:
    def __init__(self, candidates: list[dict[str, Any]], catalog: list[dict[str, str]], output: Path):
        self.candidates = candidates
        self.catalog = catalog
        self.output = output
        saved = load_reviews(output)
        self.reviews = {row["case_id"]: saved.get(row["case_id"], initial_review(row)) for row in candidates}
        self.catalog_names = {entry["name"] for entry in catalog}

    def update(self, case_id: str, payload: dict[str, Any]) -> None:
        if case_id not in self.reviews:
            raise KeyError(case_id)
        status = text(payload.get("review_status"))
        if status not in STATUSES:
            raise ValueError(f"invalid review_status: {status}")
        command = text(payload.get("expected_command")).strip()
        args = text(payload.get("expected_args")).strip()
        if status in {"accepted", "corrected"}:
            if not command:
                raise ValueError("expected_command is required when accepting or correcting")
            if catalog_violation(command, self.catalog_names):
                raise ValueError(f"expected_command is not in catalog: {command}")
        else:
            command, args = "", ""
        reviewed = status != "unreviewed"
        self.reviews[case_id] = {
            "review_status": status,
            "reviewed": reviewed,
            "expected_command": command,
            "expected_args": args,
            "reviewed_at": datetime.now(timezone.utc).isoformat() if reviewed else "",
            "review_note": text(payload.get("review_note")).strip(),
        }
        self.save()

    def rows(self) -> list[dict[str, Any]]:
        return [
            {**row, "schema_version": REVIEW_SCHEMA_VERSION, **self.reviews[row["case_id"]]}
            for row in self.candidates
        ]

    def save(self) -> None:
        self.output.parent.mkdir(parents=True, exist_ok=True)
        fd, temporary = tempfile.mkstemp(prefix=self.output.name + ".", suffix=".tmp", dir=self.output.parent)
        try:
            with os.fdopen(fd, "w", encoding="utf-8", newline="\n") as handle:
                for row in self.rows():
                    handle.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")
                handle.flush()
                os.fsync(handle.fileno())
            os.replace(temporary, self.output)
        except BaseException:
            try:
                os.unlink(temporary)
            except FileNotFoundError:
                pass
            raise

    def payload(self) -> dict[str, Any]:
        rows = self.rows()
        input_groups: dict[str, list[dict[str, Any]]] = {}
        for row in rows:
            input_groups.setdefault(text(row.get("input_text")), []).append(row)
        duplicate_ids: set[str] = set()
        conflict_ids: set[str] = set()
        for group in input_groups.values():
            if len(group) > 1:
                duplicate_ids.update(row["case_id"] for row in group)
            targets = {(row["expected_command"], row["expected_args"]) for row in group if row["review_status"] in {"accepted", "corrected"}}
            if len(targets) > 1:
                conflict_ids.update(row["case_id"] for row in group)
        for row in rows:
            row["warnings"] = {
                "duplicate": row["case_id"] in duplicate_ids,
                "conflict": row["case_id"] in conflict_ids,
                "candidate_catalog_violation": catalog_violation(text(row.get("candidate_command")), self.catalog_names),
                "expected_catalog_violation": catalog_violation(row["expected_command"], self.catalog_names),
                "incorrect_without_correction": row.get("feedback_label") == "incorrect" and not text(row.get("feedback_correction")),
            }
        return {"rows": rows, "catalog": self.catalog, "output": str(self.output)}


def handler_for(store: ReviewStore, html: bytes) -> type[BaseHTTPRequestHandler]:
    class Handler(BaseHTTPRequestHandler):
        def send_bytes(self, status: int, body: bytes, content_type: str) -> None:
            self.send_response(status)
            self.send_header("Content-Type", content_type)
            self.send_header("Content-Length", str(len(body)))
            self.send_header("Cache-Control", "no-store")
            self.end_headers()
            self.wfile.write(body)

        def send_json(self, status: int, value: Any) -> None:
            self.send_bytes(status, json.dumps(value, ensure_ascii=False).encode(), "application/json; charset=utf-8")

        def do_GET(self) -> None:
            path = urlsplit(self.path).path
            if path == "/":
                self.send_bytes(200, html, "text/html; charset=utf-8")
            elif path == "/api/data":
                self.send_json(200, store.payload())
            else:
                self.send_json(404, {"error": "not found"})

        def do_POST(self) -> None:
            if urlsplit(self.path).path != "/api/review":
                self.send_json(404, {"error": "not found"})
                return
            try:
                length = int(self.headers.get("Content-Length", "0"))
                if not 0 < length <= 1_000_000:
                    raise ValueError("request is too large")
                payload = json.loads(self.rfile.read(length))
                if not isinstance(payload, dict):
                    raise ValueError("request body must be a JSON object")
                store.update(text(payload.get("case_id")), payload)
                self.send_json(200, store.payload())
            except (ValueError, KeyError, json.JSONDecodeError) as exc:
                self.send_json(400, {"error": str(exc)})

        def log_message(self, fmt: str, *args: Any) -> None:
            print(f"{self.address_string()} - {fmt % args}")

    return Handler


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True, help="candidate JSONL or CSV")
    parser.add_argument("--catalog", type=Path, required=True, help="command catalog JSON")
    parser.add_argument("--output", type=Path, required=True, help="reviewed JSONL (also used to resume)")
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8765)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    if args.host not in {"127.0.0.1", "localhost", "::1"}:
        raise SystemExit("--host must be a loopback address; this UI is local-only")
    if not 1 <= args.port <= 65535:
        raise SystemExit("--port must be between 1 and 65535")
    store = ReviewStore(read_rows(args.input), read_catalog(args.catalog), args.output)
    html = Path(__file__).with_name("index.html").read_bytes()
    server = HTTPServer((args.host, args.port), handler_for(store, html))
    print(f"review UI: http://{args.host}:{server.server_port}")
    print(f"reviewed JSONL: {args.output}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nstopped")
    finally:
        server.server_close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
