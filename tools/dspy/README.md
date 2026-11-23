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
- `regression_cases.jsonl`
  - 明示的な対象選択など、モデル変更時にも維持する固定回帰ケース

## Shared resolver program

本番HTTP resolverとオフライン最適化は、どちらも
`tools/dspy_common/resolver_program.py` の単一段 `ResolverProgram` / `ResolveSignature` を使用します。

単一段構成を採用する理由は、本番の1回のLM呼び出しとAPI入出力を維持しながら、オフラインで最適化したものと
同じプログラムを本番で実行できるためです。意図解析・command選択・引数生成を別々に呼び出す4段構成は使用しません。
最適化成果物はDSPy programの状態と`manifest.json`の組として保存され、本番resolverは
model / prompt / program schemaが一致し、固定testセットの評価ゲートを通過したものだけをロードします。

## モデル別成果物の生成と採用

データはseed固定でtrain/dev/test（既定60/20/20）へ分割され、最適化に使用しないtestセットの
command accuracy / args accuracy / unresolved率 / catalog違反率で採用を判定します。
`--dataset-version`を省略すると入力JSONLのSHA-256からversionを生成します。

```powershell
# LFM用
python tools/dspy/optimize_and_evaluate.py `
  --dataset-jsonl .\tmp\dspy\dataset.jsonl `
  --command-catalog .\tools\dspy\command_catalog.sample.json `
  --model openai/lfm2.5-2.6B `
  --api-base http://192.168.2.240:8080/v1 `
  --prompt-version resolver-v2 `
  --dataset-version reviewed-2026-08-11 `
  --artifact-version lfm-resolver-v1 `
  --artifact-out .\tmp\dspy\artifacts\lfm2.5-2.6B `
  --report-out .\tmp\dspy\lfm-report.json

# Qwen用はmodelと出力先を変更して同じデータセットで実行
```

ゲート未達の場合はreportだけを保存し、成果物は出力しません。`manifest.json`にはmodel、prompt、
program schema、dataset、artifact version、optimizer設定、baseline/optimized評価を記録します。

本番切替では成果物ディレクトリをresolverへread-only mountし、次を同時に設定します。

```powershell
$env:MODEL="openai/lfm2.5-2.6B"
$env:DSPY_PROMPT_VERSION="resolver-v2"
$env:DSPY_ARTIFACTS_HOST_DIR=".\tmp\dspy\artifacts\lfm2.5-2.6B"
docker compose -f tools/dspy-resolver/docker-compose.yml up -d --build
```

`/healthz`の`artifact_loaded`、`artifact_version`、`dataset_version`を確認してから段階的に切り替えます。
model・prompt・schema不一致、欠損、破損、ゲート未達では未最適化の共通Signatureへ自動フォールバックし、
理由を`artifact_load_error`へ表示します。ロールバックは以前の成果物ディレクトリへ戻して再起動します。

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
  --prompt-version offline-eval-v1 `
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
  --prompt-version offline-eval-v1 `
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
- `PROMPT_VERSION`

LM 設定の優先順位:

- API key: `--api-key` / wrapper parameter -> `LM_API_KEY` -> `OPENAI_API_KEY`
- API base: `--api-base` / wrapper parameter -> `LM_API_BASE`
- Model type: `--model-type` / wrapper parameter -> `LM_MODEL_TYPE` -> `chat`
- Temperature: `--temperature` / wrapper parameter -> `LM_TEMPERATURE`
- Max tokens: `--max-tokens` / wrapper parameter -> `LM_MAX_TOKENS`

## レビュー済みデータの確定

リポジトリ直下で、Python 3.11 以降と本プロジェクト用の Go を使います。入力はレビュー画面が出力した `resolver-reviewed/v1` です。実ログ・グループ対応表・出力は Git 管理外の `tmp/` に保存します。

```powershell
$dir = "tmp/resolver-events/example"
python tools/dspy/finalize_dataset.py `
  --reviewed "$dir/reviewed.jsonl" --group-map "$dir/groups.json" `
  --config-dir config --output-dir "$dir/evaluation-v1" `
  --model openai/qwen3.6-35b-a3b-mtp --model openai/lfm2.5-2.6B `
  --prompt-version resolver-v2 --seed 42 --min-command-accuracy 0.80 --min-arg-accuracy 0.60 `
  --from 2026-01-01T00:00:00Z --to 2026-02-01T00:00:00Z
```

`--from` と `--to` には収集時の開始日時と終了日時（終了を含まない）を指定します。
`--group-map` は任意です。たとえば `{"group-a":"artist-1","group-b":"artist-1"}` とすると、同じ対象の別表記を一緒に分割します。固定例の言い換えには、対応する `regression_cases.jsonl` の `request_id` を割り当てます。表記の正規化だけで見つからない言い換えは、この対応表で指定してください。

確定処理は Go のコマンド定義（指定ディレクトリのマクロを含む）で必須引数・prefix・enum を検証します。判断不能の正解はコマンドと引数を両方空文字にします。保留・未レビュー・矛盾はエラー、明示的な除外は件数に記録します。エラー時は新規出力先に `validation.json` と `validation.md` だけを残します。修正後は新しい出力先で再実行してください。

同一の正規化入力は1件にまとめ、固定例と同じグループを通常分割から除きます。残りのグループを seed で順序付け、6:2:2 を目安に train/dev/test に割り当てます。test では実ログを優先し、判断不能の実ログがあれば最低1グループを含めます。件数・出典は検証レポート、重複や回帰例に重なる入力の ID は JSON で確認できます。

評価環境に `tools/dspy/requirements.txt` をインストールし、各モデルの API 接続先を設定して実行します。両方の実行に同じマニフェストを渡すと、ファイルの SHA-256 を照合した上で、保存済みの分割・catalog・prompt version・seed・閾値を使用します。

```powershell
python tools/dspy/optimize_and_evaluate.py --evaluation-manifest "$dir/evaluation-v1/evaluation-manifest.json" --model openai/qwen3.6-35b-a3b-mtp --api-base $env:QWEN_API_BASE --report-out "$dir/qwen.json"
python tools/dspy/optimize_and_evaluate.py --evaluation-manifest "$dir/evaluation-v1/evaluation-manifest.json" --model openai/lfm2.5-2.6B --api-base $env:LFM_API_BASE --report-out "$dir/lfm.json"
```

`dataset.jsonl` は3分割を連結したものです。`regression.jsonl` の固定例は独立しており、通常の評価精度の分母に含めません。
