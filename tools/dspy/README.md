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

学習用の自然文は `input_text` 列を優先します。未設定の場合はスキップされます。

## 3) Optimize + Evaluate

```powershell
python tools/dspy/optimize_and_evaluate.py `
  --dataset-jsonl .\tmp\dspy\dataset.jsonl `
  --command-catalog .\tools\dspy\command_catalog.sample.json `
  --model openai/gpt-4o-mini `
  --report-out .\tmp\dspy\report.json `
  --min-command-accuracy 0.80 `
  --min-arg-accuracy 0.60
```

`report.json` には以下を出力します。

- `baseline` / `optimized` の精度
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

必要に応じて以下も上書きできます。

- `COMMAND_CATALOG`
- `MIN_COMMAND_ACCURACY`
- `MIN_ARG_ACCURACY`
