# DSPy Batch Pipeline

Issue #159 Phase D 向けの、オフライン最適化 + 評価ゲートの最小構成です。

## Files

- `prepare_dataset.py`
  - ResolverイベントCSVから、DSPy用JSONLデータセットを生成
- `optimize_and_evaluate.py`
  - DSPyで最適化し、オフライン評価を実行
  - しきい値判定（gate）結果をJSONで出力
- `run-batch.ps1`
  - 上記2ステップをまとめたバッチ実行ラッパー
- `Dockerfile` / `run-batch.sh`
  - Dockerで同じバッチ処理を実行するための構成
- `command_catalog.sample.json`
  - コマンド一覧サンプル（実運用では最新定義に置き換える）
- `requirements.txt`
  - Python依存

## 1) Setup

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r tools/dspy/requirements.txt
```

## 2) Dataset Build

```powershell
python tools/dspy/prepare_dataset.py `
  --input-csv .\tmp\resolver-events\resolver-events.csv `
  --output-jsonl .\tmp\dspy\dataset.jsonl `
  --min-row-per-request 2
```

入力CSVには少なくとも以下の列が必要です。

- `resolver_request_id`
- `event_name`
- `resolver_resolved_command`
- `resolver_resolved_args`
- `feedback_label`
- `feedback_correction`

次の列は任意ですが、モデル・プロンプト別の評価と再現性確保のため、resolverイベント抽出時に保持することを推奨します。

- `llm_model`
- `resolver_prompt_version`
- `resolver_artifact_version`
- `resolver_dataset_version`

学習用の自然文は `input_text` 列を優先します。未設定の場合はスキップされます。

## 3) Optimize + Evaluate

### OpenAI

```powershell
python tools/dspy/optimize_and_evaluate.py `
  --dataset-jsonl .\tmp\dspy\dataset.jsonl `
  --command-catalog .\tools\dspy\command_catalog.sample.json `
  --model openai/gpt-4o-mini `
  --report-out .\tmp\dspy\report.json `
  --min-command-accuracy 0.80 `
  --min-arg-accuracy 0.60
```

### OpenAI-compatible local LM

`tools/dspy-resolver` と同じ `LM_*` 設定を利用できます。CLI 引数を指定した場合は環境変数より優先されます。

```powershell
python tools/dspy/optimize_and_evaluate.py `
  --dataset-jsonl .\tmp\dspy\dataset.jsonl `
  --command-catalog .\tools\dspy\command_catalog.sample.json `
  --model openai/qwen2.5:14b `
  --api-base http://localhost:11434/v1 `
  --api-key local-dummy-key `
  --model-type chat `
  --temperature 0.2 `
  --max-tokens 512 `
  --report-out .\tmp\dspy\report.json `
  --min-command-accuracy 0.80 `
  --min-arg-accuracy 0.60
```

`report.json` には以下を出力します。

- `baseline` / `optimized` の精度
- `baseline` / `optimized` の利用元モデル・prompt version別精度（`breakdown`）
- `gate_passed`
- 失敗した評価ケース（最大20件）

## 4) Scheduled Batch

`run-batch.ps1` を Task Scheduler / cron 相当で定期実行してください。

ポイント:

- 本番反映は `gate_passed=true` の場合のみ
- オンライン学習はしない（バッチで再最適化）

## 5) Docker

### Build

```powershell
docker build -f tools/dspy/Dockerfile -t smarthome-dspy-batch .
```

### Run

```powershell
docker run --rm `
  -e OPENAI_API_KEY=$env:OPENAI_API_KEY `
  -e MODEL=openai/gpt-4o-mini `
  -e RESOLVER_EVENTS_CSV=/workspace/tmp/resolver-events/resolver-events.csv `
  -e WORK_DIR=/workspace/tmp/dspy `
  -v ${PWD}:/workspace `
  smarthome-dspy-batch
```

ローカルの OpenAI 互換サーバーを使う場合、Docker コンテナからホストへ到達するために `host.docker.internal` を使います。

```powershell
docker run --rm `
  -e MODEL=openai/qwen2.5:14b `
  -e LM_API_BASE=http://host.docker.internal:11434/v1 `
  -e LM_API_KEY=local-dummy-key `
  -e LM_MODEL_TYPE=chat `
  -e LM_TEMPERATURE=0.2 `
  -e LM_MAX_TOKENS=512 `
  -e RESOLVER_EVENTS_CSV=/workspace/tmp/resolver-events/resolver-events.csv `
  -e WORK_DIR=/workspace/tmp/dspy `
  -v ${PWD}:/workspace `
  smarthome-dspy-batch
```

必要に応じて以下も上書きできます。

- `COMMAND_CATALOG`
- `MIN_COMMAND_ACCURACY`
- `MIN_ARG_ACCURACY`
- `LM_API_BASE`
- `LM_API_KEY`
- `LM_MODEL_TYPE`
- `LM_TEMPERATURE`
- `LM_MAX_TOKENS`

LM 設定の優先順位:

- API key: `--api-key` / wrapper parameter -> `LM_API_KEY` -> `OPENAI_API_KEY`
- API base: `--api-base` / wrapper parameter -> `LM_API_BASE`
- Model type: `--model-type` / wrapper parameter -> `LM_MODEL_TYPE` -> `chat`
- Temperature: `--temperature` / wrapper parameter -> `LM_TEMPERATURE`
- Max tokens: `--max-tokens` / wrapper parameter -> `LM_MAX_TOKENS`
