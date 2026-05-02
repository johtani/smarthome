# DSPy Resolver Server

`smarthome` の外部 HTTP resolver 最小実装です。

- コマンド解決（`resolver.mode=dspy`）: `POST /resolve`
- Music Intent 解決（`search and play` 強化）: `POST /resolve-music-intent`

## API

- `GET /healthz`
- `POST /resolve`
- `POST /resolve-music-intent`

### Request (`POST /resolve`)

```json
{
  "text": "電気をつけて",
  "command_list": "  light on : turn on the light\n  aircon on : turn on the air conditioner",
  "prompt_version": "2026-04-16"
}
```

### Response (`POST /resolve`)

```json
{
  "command": "light on",
  "args": "",
  "thought": "..."
}
```

`command` は `command_list` 内の名前に一致するものだけ返します。一致しない場合は空文字列を返します。

### Request (`POST /resolve-music-intent`)

```json
{
  "text": "宇多田ヒカルのFirst Loveかけて"
}
```

### Response (`POST /resolve-music-intent`)

```json
{
  "artist_candidates": ["宇多田ヒカル"],
  "track_candidates": ["First Love"],
  "genre_candidates": [],
  "must_terms": [],
  "confidence": 0.92,
  "ambiguous": false,
  "reason": "artist and track detected",
  "model": "openai/gpt-4o-mini"
}
```

## Docker Compose (recommended)

利用可能な主な環境変数:

```powershell
# 共通
$env:MODEL="openai/gpt-4o-mini"

# OpenAI利用時
$env:OPENAI_API_KEY="<your_api_key>"

# ローカルOpenAI互換(例: llm-swap)利用時
$env:LM_API_BASE="http://host.docker.internal:11434/v1"
$env:LM_API_KEY="local-dummy-key"
$env:LM_MODEL_TYPE="chat"
# 任意
$env:LM_TEMPERATURE="0.2"
$env:LM_MAX_TOKENS="512"

# OpenTelemetry利用時
$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"
$env:OTEL_SERVICE_NAME="smarthome-dspy-resolver"
# llm-swap向けHTTP request body previewを明示的に記録する場合のみ
$env:DSPY_RESOLVER_TRACE_INCLUDE_HTTP_BODY="true"
```

起動:

```powershell
docker compose -f tools/dspy-resolver/docker-compose.yml up -d --build
```

ヘルスチェック:

```powershell
curl http://localhost:18089/healthz
```

`/healthz` は `model` に加えて `api_base` / `model_type` / `temperature` / `max_tokens` / `api_key_source` を返します。  
`LM_TEMPERATURE` または `LM_MAX_TOKENS` が不正値の場合を含め、LM 初期化に失敗した場合は `503` を返します。

Dockerコンテナからホスト上のローカルLLMへ接続する場合、`localhost` ではなく `host.docker.internal` を使ってください。

## OpenTelemetry

dspy-resolver は OpenTelemetry に対応しています。`OTEL_EXPORTER_OTLP_ENDPOINT` または
`OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` を設定すると OTLP/HTTP で trace を送信します。
未設定の場合、FastAPI と HTTP client の instrumentation は有効ですが exporter は作成しないため、
ローカル起動はこれまで通り動作します。

主な span/event:

- FastAPI instrumentation による `/resolve` / `/resolve-music-intent` の server span
- `dspy.resolve.command`
- `dspy.resolve.music_intent`
- `dspy.request` event
- `dspy.response` event
- `llm.request` event（LiteLLM input callback で取得できた場合）
- DSPy/LiteLLM 配下の `requests` / `httpx` outgoing HTTP span

記録する主な属性:

- `dspy.lm.model`
- `dspy.lm.api_base`
- `dspy.lm.model_type`
- `dspy.lm.temperature`
- `dspy.lm.max_tokens`
- `resolver.prompt_version`
- `command_catalog.count`
- 各入力・出力文字列の長さ

ユーザー入力や catalog 本文はデフォルトでは trace に載せず、長さだけを記録します。
内容の先頭を確認したい場合のみ、次を設定してください。

```powershell
$env:DSPY_RESOLVER_TRACE_INCLUDE_INPUT="true"
```

この場合も preview は先頭 256 文字に制限されます。

llm-swap に送信される HTTP request body もデフォルトでは trace に載せません。
`LM_API_BASE` 宛ての outgoing request に限って body の先頭を確認したい場合のみ、次を設定してください。

```powershell
$env:DSPY_RESOLVER_TRACE_INCLUDE_HTTP_BODY="true"
```

この場合の preview は先頭 2048 文字に制限されます。
HTTP client の hook から body が読めない場合は、LiteLLM input callback で取得できた変換済み LLM 入力を
`llm.request` event の `llm.request_body.preview` に記録します。
この preview は API key 類を `[redacted]` に置換し、先頭 4096 文字に制限されます。

## Docker (single command)

```powershell
docker build -f tools/dspy-resolver/Dockerfile -t smarthome-dspy-resolver .
docker run --rm -p 18089:8080 `
  -e MODEL=$env:MODEL `
  -e OPENAI_API_KEY=$env:OPENAI_API_KEY `
  -e LM_API_BASE=$env:LM_API_BASE `
  -e LM_API_KEY=$env:LM_API_KEY `
  -e LM_MODEL_TYPE=$env:LM_MODEL_TYPE `
  -e LM_TEMPERATURE=$env:LM_TEMPERATURE `
  -e LM_MAX_TOKENS=$env:LM_MAX_TOKENS `
  smarthome-dspy-resolver
```

## smarthome 側設定

### コマンド解決（既存）

```json
{
  "resolver": {
    "mode": "dspy",
    "dspy_endpoint": "http://localhost:18089/resolve",
    "dspy_timeout_seconds": 5
  }
}
```

### Music Intent 解決（今回追加）

```json
{
  "owntone": {
    "music_intent_endpoint": "http://localhost:18089/resolve-music-intent",
    "music_intent_timeout_seconds": 5,
    "music_intent_confidence_threshold": 0.75
  }
}
```

または環境変数:

- `SMARTHOME_OWNTONE_MUSIC_INTENT_ENDPOINT=http://localhost:18089/resolve-music-intent`
- `SMARTHOME_OWNTONE_MUSIC_INTENT_TIMEOUT_SECONDS=5`
- `SMARTHOME_OWNTONE_MUSIC_INTENT_CONFIDENCE_THRESHOLD=0.75`

Resolver が未設定または失敗した場合、`smarthome` は既存検索経路へフォールバックします。
