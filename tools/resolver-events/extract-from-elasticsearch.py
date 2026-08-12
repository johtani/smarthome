#!/usr/bin/env python3
"""Extract review candidates from OTLP traces stored in Elasticsearch."""

from __future__ import annotations

import argparse
import base64
import csv
import hashlib
import json
import os
import re
import ssl
import sys
import urllib.error
import urllib.parse
import urllib.request
from collections import Counter
from dataclasses import asdict, dataclass, fields
from pathlib import Path
from typing import Any, Iterable, Iterator

SCHEMA_VERSION = "resolver-review-candidate/v1"
EVENT_NAMES = {"resolver.decision", "resolver.execution", "resolver.feedback", "dspy.request", "llm.request"}
INPUT_SOURCES = (
    "dspy.request.text",
    "resolver.input.preview",
    "dspy.request.utterance.preview",
    "llm.request_body.preview",
    "llm.http.request_body.preview",
)
SENSITIVE_RE = re.compile(
    r"(?i)(authorization\s*[:=]|bearer\s+[a-z0-9._~+/=-]{12,}|api[_-]?key\s*[:=]|"
    r"access[_-]?token\s*[:=]|secret\s*[:=])"
)


@dataclass
class Candidate:
    schema_version: str = SCHEMA_VERSION
    case_id: str = ""
    group_id: str = ""
    trace_id: str = ""
    resolver_request_id: str = ""
    timestamp: str = ""
    service_name: str = ""
    input_text: str = ""
    input_text_source: str = ""
    input_text_truncated: bool = False
    candidate_command: str = ""
    candidate_args: str = ""
    source: str = ""
    resolver_path: str = ""
    resolver_mode: str = ""
    execution_status: str = ""
    feedback_label: str = ""
    feedback_correction: str = ""
    llm_model: str = ""
    prompt_version: str = ""
    artifact_version: str = ""
    dataset_version: str = ""
    resolver_schema_version: str = ""


def scalar(value: Any) -> Any:
    if isinstance(value, dict):
        for key in ("stringValue", "intValue", "doubleValue", "boolValue", "string_value", "int_value", "double_value", "bool_value"):
            if key in value:
                return value[key]
    return value


def attrs_map(attrs: Any) -> dict[str, Any]:
    if isinstance(attrs, dict):
        return {str(k): scalar(v) for k, v in attrs.items()}
    result: dict[str, Any] = {}
    for item in attrs or []:
        if isinstance(item, dict) and "key" in item:
            result[str(item["key"])] = scalar(item.get("value"))
    return result


def first(attrs: dict[str, Any], *keys: str) -> str:
    for key in keys:
        value = attrs.get(key)
        if value is not None and str(value).strip():
            return str(value)
    return ""


def bool_value(value: Any) -> bool:
    return value is True or str(value).strip().lower() in {"true", "1", "yes"}


def iter_spans(source: dict[str, Any]) -> Iterator[tuple[str, str, dict[str, Any]]]:
    roots = source.get("resourceSpans") or source.get("resource_spans")
    if roots:
        for resource_span in roots:
            resource = resource_span.get("resource") or {}
            service = first(attrs_map(resource.get("attributes")), "service.name")
            scopes = resource_span.get("scopeSpans") or resource_span.get("scope_spans") or []
            for scope in scopes:
                for span in scope.get("spans") or []:
                    yield str(span.get("traceId") or span.get("trace_id") or ""), service, span
        return

    # Common OTel Elasticsearch exporter shape: one span per document.
    span = source.get("span") if isinstance(source.get("span"), dict) else source
    resource = source.get("resource") or {}
    service = first(attrs_map(resource.get("attributes")), "service.name") or str(
        source.get("service", {}).get("name", "") if isinstance(source.get("service"), dict) else source.get("service.name", "")
    )
    trace = source.get("trace") or {}
    trace_id = str(trace.get("id", "") if isinstance(trace, dict) else "") or str(span.get("traceId") or span.get("trace_id") or source.get("trace.id") or "")
    yield trace_id, service, span


def events_from_source(source: dict[str, Any]) -> Iterator[dict[str, Any]]:
    for trace_id, service, span in iter_spans(source):
        span_attrs = attrs_map(span.get("attributes"))
        for event in span.get("events") or []:
            name = str(event.get("name") or "")
            if name not in EVENT_NAMES:
                continue
            event_attrs = {**span_attrs, **attrs_map(event.get("attributes"))}
            yield {
                "trace_id": trace_id,
                "service_name": service,
                "timestamp": str(event.get("timeUnixNano") or event.get("time_unix_nano") or source.get("@timestamp") or ""),
                "name": name,
                "attrs": event_attrs,
            }


def extract_body_input(value: str) -> str:
    try:
        payload = json.loads(value)
    except (json.JSONDecodeError, TypeError):
        return ""
    candidates = ("text", "utterance", "input", "prompt")
    stack = [payload]
    while stack:
        item = stack.pop(0)
        if isinstance(item, dict):
            for key in candidates:
                val = item.get(key)
                if isinstance(val, str) and val.strip():
                    return val
            stack.extend(item.values())
        elif isinstance(item, list):
            stack.extend(item)
    return ""


def normalize(events: Iterable[dict[str, Any]]) -> list[Candidate]:
    groups: dict[str, list[dict[str, Any]]] = {}
    for event in events:
        request_id = first(event["attrs"], "resolver.request_id")
        # A trace is the primary join key. Some child spans (notably the DSPy
        # service span) do not repeat resolver.request_id, so splitting on the
        # pair would lose their input/model attributes.
        key = event["trace_id"] or f"request:{request_id}"
        groups.setdefault(key, []).append(event)

    candidates: list[Candidate] = []
    for group_key, group in sorted(groups.items()):
        merged: dict[str, Any] = {}
        sources: list[str] = []
        for event in sorted(group, key=lambda item: item["timestamp"]):
            merged.update({k: v for k, v in event["attrs"].items() if v not in (None, "")})
            sources.append(event["name"])

        trace_id = next((e["trace_id"] for e in group if e["trace_id"]), "")
        request_id = first(merged, "resolver.request_id")

        input_text = ""
        input_source = ""
        truncated = False
        for key in INPUT_SOURCES:
            raw = first(merged, key)
            if not raw:
                continue
            parsed = extract_body_input(raw) if "body" in key else raw
            if parsed:
                input_text, input_source = parsed, key
                truncated = bool_value(merged.get(key.removesuffix(".preview") + ".truncated"))
                break

        group_id = request_id or trace_id or group_key
        digest = hashlib.sha256(f"{trace_id}\0{request_id}\0{input_text}".encode("utf-8")).hexdigest()[:20]
        candidates.append(Candidate(
            case_id=f"trace-{digest}", group_id=group_id, trace_id=trace_id,
            resolver_request_id=request_id, timestamp=min((e["timestamp"] for e in group if e["timestamp"]), default=""),
            service_name=next((e["service_name"] for e in group if e["service_name"]), ""),
            input_text=input_text, input_text_source=input_source, input_text_truncated=truncated,
            candidate_command=first(merged, "resolver.resolved_command", "resolver.selected_command"),
            candidate_args=first(merged, "resolver.resolved_args", "resolver.selected_args"),
            source=",".join(dict.fromkeys(sources)), resolver_path=first(merged, "resolver.path", "resolver.path_hint"),
            resolver_mode=first(merged, "resolver.mode"), execution_status=first(merged, "resolver.execution_status"),
            feedback_label=first(merged, "feedback.label"), feedback_correction=first(merged, "feedback.correction"),
            llm_model=first(merged, "llm.model", "dspy.lm.model"),
            prompt_version=first(merged, "resolver.prompt_version", "dspy.request.prompt_version"),
            artifact_version=first(merged, "resolver.artifact_version"), dataset_version=first(merged, "resolver.dataset_version"),
            resolver_schema_version=first(merged, "resolver.schema_version"),
        ))
    return candidates


class ElasticsearchClient:
    def __init__(self, url: str, index: str, api_key: str, username: str, password: str, insecure: bool):
        self.base = url.rstrip("/")
        self.index = index
        self.context = ssl._create_unverified_context() if insecure else ssl.create_default_context()
        self.headers = {"Content-Type": "application/json"}
        if api_key:
            self.headers["Authorization"] = f"ApiKey {api_key}"
        elif username:
            token = base64.b64encode(f"{username}:{password}".encode()).decode()
            self.headers["Authorization"] = f"Basic {token}"

    def request(self, path: str, body: dict[str, Any] | None = None, method: str = "POST") -> dict[str, Any]:
        data = json.dumps(body).encode() if body is not None else None
        req = urllib.request.Request(self.base + path, data, self.headers, method=method)
        try:
            with urllib.request.urlopen(req, context=self.context, timeout=60) as response:
                return json.load(response)
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"Elasticsearch returned HTTP {exc.code}: {detail[:1000]}") from exc

    def hits(self, start: str, end: str, service: str, page_size: int) -> Iterator[dict[str, Any]]:
        filters: list[dict[str, Any]] = [{"range": {"@timestamp": {"gte": start, "lt": end}}}]
        pit = self.request(f"/{self.index}/_pit?keep_alive=2m")
        pit_id = str(pit.get("id") or "")
        if not pit_id:
            raise RuntimeError("Elasticsearch did not return a point-in-time ID")
        search_after = None
        try:
            while True:
                body: dict[str, Any] = {
                    "size": page_size, "track_total_hits": False,
                    "query": {"bool": {"filter": filters}},
                    "pit": {"id": pit_id, "keep_alive": "2m"},
                    "sort": [{"@timestamp": "asc"}, "_shard_doc"],
                }
                if search_after is not None:
                    body["search_after"] = search_after
                result = self.request("/_search", body)
                pit_id = str(result.get("pit_id") or pit_id)
                hits = result.get("hits", {}).get("hits", [])
                if not hits:
                    break
                for hit in hits:
                    yield hit
                search_after = hits[-1].get("sort")
                if not search_after or len(hits) < page_size:
                    break
        finally:
            self.request("/_pit", {"id": pit_id}, method="DELETE")


def audit(candidates: list[Candidate], meta: dict[str, Any]) -> dict[str, Any]:
    missing_fields = ("input_text", "llm_model", "prompt_version", "artifact_version", "dataset_version")
    sensitive = [c.case_id for c in candidates if SENSITIVE_RE.search(c.input_text)]
    return {
        "schema_version": SCHEMA_VERSION,
        "query": meta,
        "counts": {
            "documents": meta["documents"], "events": meta["events"], "candidates": len(candidates),
            "unique_traces": len({c.trace_id for c in candidates if c.trace_id}),
            "unique_requests": len({c.resolver_request_id for c in candidates if c.resolver_request_id}),
            "truncated_inputs": sum(c.input_text_truncated for c in candidates),
        },
        "missing": {name: sum(not getattr(c, name) for c in candidates) for name in missing_fields},
        "models": dict(sorted(Counter(c.llm_model or "(missing)" for c in candidates).items())),
        "sensitive_input_warning_count": len(sensitive),
        "sensitive_input_case_ids": sensitive,
    }


def write_outputs(output_dir: Path, candidates: list[Candidate], report: dict[str, Any]) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    rows = [asdict(c) for c in candidates]
    with (output_dir / "review-candidates.jsonl").open("w", encoding="utf-8", newline="\n") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")
    names = [field.name for field in fields(Candidate)]
    with (output_dir / "review-candidates.csv").open("w", encoding="utf-8-sig", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=names)
        writer.writeheader(); writer.writerows(rows)
    with (output_dir / "audit.json").open("w", encoding="utf-8", newline="\n") as f:
        json.dump(report, f, ensure_ascii=False, indent=2, sort_keys=True); f.write("\n")
    q, counts, missing = report["query"], report["counts"], report["missing"]
    lines = ["# Resolver trace audit", "", f"- Period: `{q['from']}` to `{q['to']}`", f"- Index: `{q['index']}`", f"- Service: `{q['service'] or '(all)'}`", "", "## Counts", ""]
    lines.extend(f"- {key}: {value}" for key, value in counts.items())
    lines += ["", "## Missing metadata", ""]
    lines.extend(f"- {key}: {value}" for key, value in missing.items())
    lines += ["", "## Sensitive input candidates", "", f"- warnings: {report['sensitive_input_warning_count']}", ""]
    (output_dir / "audit.md").write_text("\n".join(lines), encoding="utf-8")


def parse_args(argv: list[str]) -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--url", required=True, help="Elasticsearch base URL (credentials must not be included)")
    p.add_argument("--index", required=True)
    p.add_argument("--from", dest="start", required=True, help="Inclusive ISO-8601 timestamp")
    p.add_argument("--to", dest="end", required=True, help="Exclusive ISO-8601 timestamp")
    p.add_argument("--service", default="")
    p.add_argument("--page-size", type=int, default=500)
    p.add_argument("--output-dir", type=Path, required=True)
    p.add_argument("--insecure", action="store_true", help="Disable TLS certificate verification")
    return p.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    if not 1 <= args.page_size <= 10000:
        raise SystemExit("--page-size must be between 1 and 10000")
    if urllib.parse.urlsplit(args.url).username:
        raise SystemExit("credentials in --url are forbidden; use environment variables")
    client = ElasticsearchClient(args.url, args.index, os.getenv("ES_API_KEY", ""), os.getenv("ES_USERNAME", ""), os.getenv("ES_PASSWORD", ""), args.insecure)
    docs = list(client.hits(args.start, args.end, args.service, args.page_size))
    events = [
        event for hit in docs for event in events_from_source(hit.get("_source") or {})
        if not args.service or event["service_name"] == args.service
    ]
    candidates = normalize(events)
    meta = {"from": args.start, "to": args.end, "index": args.index, "service": args.service, "page_size": args.page_size, "documents": len(docs), "events": len(events)}
    report = audit(candidates, meta)
    write_outputs(args.output_dir, candidates, report)
    print(f"exported {len(candidates)} candidates from {len(events)} events in {len(docs)} documents")
    if report["sensitive_input_warning_count"]:
        print(f"warning: {report['sensitive_input_warning_count']} candidate inputs may contain secrets", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
