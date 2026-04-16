#!/usr/bin/env sh
set -eu

RESOLVER_EVENTS_CSV="${RESOLVER_EVENTS_CSV:-/workspace/tmp/resolver-events/resolver-events.csv}"
WORK_DIR="${WORK_DIR:-/workspace/tmp/dspy}"
MODEL="${MODEL:-openai/gpt-4o-mini}"
COMMAND_CATALOG="${COMMAND_CATALOG:-/workspace/tools/dspy/command_catalog.sample.json}"
MIN_COMMAND_ACCURACY="${MIN_COMMAND_ACCURACY:-0.80}"
MIN_ARG_ACCURACY="${MIN_ARG_ACCURACY:-0.60}"

if [ ! -f "$RESOLVER_EVENTS_CSV" ]; then
  echo "resolver events csv not found: $RESOLVER_EVENTS_CSV" >&2
  exit 1
fi

mkdir -p "$WORK_DIR"

DATASET_JSONL="$WORK_DIR/dataset.jsonl"
REPORT_JSON="$WORK_DIR/report.json"

python /opt/dspy/prepare_dataset.py \
  --input-csv "$RESOLVER_EVENTS_CSV" \
  --output-jsonl "$DATASET_JSONL" \
  --min-row-per-request 2

python /opt/dspy/optimize_and_evaluate.py \
  --dataset-jsonl "$DATASET_JSONL" \
  --command-catalog "$COMMAND_CATALOG" \
  --model "$MODEL" \
  --report-out "$REPORT_JSON" \
  --min-command-accuracy "$MIN_COMMAND_ACCURACY" \
  --min-arg-accuracy "$MIN_ARG_ACCURACY"

echo "batch finished: $REPORT_JSON"
