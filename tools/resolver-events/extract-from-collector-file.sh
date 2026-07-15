#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  bash tools/resolver-events/extract-from-collector-file.sh --input <resolver-traces.jsonl> --output <resolver-events.csv>

Options:
  --input, -i   OTel Collector file exporter JSONL path
  --output, -o  Output CSV path
  --help, -h    Show this help
USAGE
}

input_path=""
output_csv=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i)
      input_path="${2:-}"
      shift 2
      ;;
    --output|-o)
      output_csv="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$input_path" || -z "$output_csv" ]]; then
  usage >&2
  exit 2
fi

if [[ ! -f "$input_path" ]]; then
  echo "input file not found: $input_path" >&2
  exit 1
fi

mkdir -p "$(dirname "$output_csv")"

python3 - "$input_path" "$output_csv" <<'PY'
import csv
import json
import sys

input_path = sys.argv[1]
output_csv = sys.argv[2]

event_names = {"resolver.decision", "resolver.execution", "resolver.feedback"}
fieldnames = [
    "timestamp",
    "service_name",
    "trace_id",
    "event_name",
    "resolver_schema_version",
    "resolver_request_id",
    "resolver_channel",
    "resolver_input_text_hash",
    "resolver_path",
    "resolver_mode",
    "llm_model",
    "resolver_resolved_command",
    "resolver_resolved_args",
    "resolver_execution_status",
    "feedback_label",
    "feedback_correction",
    "resolver_did_you_mean_command",
]


def attr_value(attrs, key):
    for attr in attrs or []:
        if attr.get("key") != key:
            continue
        value = attr.get("value") or {}
        for value_key in ("stringValue", "intValue", "doubleValue", "boolValue"):
            if value_key in value and value[value_key] is not None:
                return str(value[value_key])
    return ""


rows = []
with open(input_path, "r", encoding="utf-8") as f:
    for line_no, line in enumerate(f, start=1):
        line = line.strip()
        if not line:
            continue
        try:
            doc = json.loads(line)
        except json.JSONDecodeError as exc:
            raise SystemExit(f"failed to parse JSON on line {line_no}: {exc}") from exc

        for resource_span in doc.get("resourceSpans", []) or []:
            service_name = attr_value(
                (resource_span.get("resource") or {}).get("attributes"),
                "service.name",
            )
            for scope_span in resource_span.get("scopeSpans", []) or []:
                for span in scope_span.get("spans", []) or []:
                    trace_id = str(span.get("traceId", ""))
                    for event in span.get("events", []) or []:
                        event_name = str(event.get("name", ""))
                        if event_name not in event_names:
                            continue

                        attrs = event.get("attributes") or []
                        rows.append(
                            {
                                "timestamp": str(event.get("timeUnixNano", "")),
                                "service_name": service_name,
                                "trace_id": trace_id,
                                "event_name": event_name,
                                "resolver_schema_version": attr_value(attrs, "resolver.schema_version"),
                                "resolver_request_id": attr_value(attrs, "resolver.request_id"),
                                "resolver_channel": attr_value(attrs, "resolver.channel"),
                                "resolver_input_text_hash": attr_value(attrs, "resolver.input_text_hash"),
                                "resolver_path": attr_value(attrs, "resolver.path"),
                                "resolver_mode": attr_value(attrs, "resolver.mode"),
                                "llm_model": attr_value(attrs, "llm.model"),
                                "resolver_resolved_command": attr_value(attrs, "resolver.resolved_command"),
                                "resolver_resolved_args": attr_value(attrs, "resolver.resolved_args"),
                                "resolver_execution_status": attr_value(attrs, "resolver.execution_status"),
                                "feedback_label": attr_value(attrs, "feedback.label"),
                                "feedback_correction": attr_value(attrs, "feedback.correction"),
                                "resolver_did_you_mean_command": attr_value(attrs, "resolver.did_you_mean_command"),
                            }
                        )

with open(output_csv, "w", encoding="utf-8", newline="") as f:
    writer = csv.DictWriter(f, fieldnames=fieldnames)
    writer.writeheader()
    writer.writerows(rows)

print(f"exported rows: {len(rows)}")
print(f"output: {output_csv}")
PY
