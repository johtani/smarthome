#!/usr/bin/env sh
set -eu

RESOLVER_EVENTS_CSV="${RESOLVER_EVENTS_CSV:-/workspace/tmp/resolver-events/resolver-events.csv}"
WORK_DIR="${WORK_DIR:-/workspace/tmp/dspy}"
MODEL="${MODEL:-openai/gpt-4o-mini}"
COMMAND_CATALOG="${COMMAND_CATALOG:-/workspace/tools/dspy/command_catalog.sample.json}"
MIN_COMMAND_ACCURACY="${MIN_COMMAND_ACCURACY:-0.80}"
MIN_ARG_ACCURACY="${MIN_ARG_ACCURACY:-0.60}"
LM_API_BASE="${LM_API_BASE:-}"
LM_API_KEY="${LM_API_KEY:-}"
LM_MODEL_TYPE="${LM_MODEL_TYPE:-}"
LM_TEMPERATURE="${LM_TEMPERATURE:-}"
LM_MAX_TOKENS="${LM_MAX_TOKENS:-}"

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

set -- \
  --dataset-jsonl "$DATASET_JSONL" \
  --command-catalog "$COMMAND_CATALOG" \
  --model "$MODEL" \
  --report-out "$REPORT_JSON" \
  --min-command-accuracy "$MIN_COMMAND_ACCURACY" \
  --min-arg-accuracy "$MIN_ARG_ACCURACY"

if [ -n "$LM_API_BASE" ]; then
  set -- "$@" --api-base "$LM_API_BASE"
fi
if [ -n "$LM_API_KEY" ]; then
  set -- "$@" --api-key "$LM_API_KEY"
fi
if [ -n "$LM_MODEL_TYPE" ]; then
  set -- "$@" --model-type "$LM_MODEL_TYPE"
fi
if [ -n "$LM_TEMPERATURE" ]; then
  set -- "$@" --temperature "$LM_TEMPERATURE"
fi
if [ -n "$LM_MAX_TOKENS" ]; then
  set -- "$@" --max-tokens "$LM_MAX_TOKENS"
fi

python /opt/dspy/optimize_and_evaluate.py "$@"

echo "batch finished: $REPORT_JSON"
