# DSPy Resolver Server

`smarthome` の `resolver.mode=dspy` で利用する外部 HTTP resolver の最小実装です。

## API

- `GET /healthz`
- `POST /resolve`

### Request (`POST /resolve`)

```json
{
  "text": "電気をつけて",
  "command_list": "  light on : turn on the light\n  aircon on : turn on the air conditioner",
  "prompt_version": "2026-04-16"
}
```

### Response

```json
{
  "command": "light on",
  "args": "",
  "thought": "..."
}
```

`command` は `command_list` 内の名前に一致するものだけ返します。一致しない場合は空文字列を返します。

## Docker Compose (recommended)

1. 環境変数を設定:

```powershell
$env:OPENAI_API_KEY="<your_api_key>"
```

2. 起動:

```powershell
docker compose -f tools/dspy-resolver/docker-compose.yml up -d --build
```

3. ヘルスチェック:

```powershell
curl http://localhost:8089/healthz
```

`OPENAI_API_KEY` 未設定などで LM 初期化に失敗した場合は `503` を返します。

## Docker (single command)

```powershell
docker build -f tools/dspy-resolver/Dockerfile -t smarthome-dspy-resolver .
docker run --rm -p 8089:8080 `
  -e OPENAI_API_KEY=$env:OPENAI_API_KEY `
  -e MODEL=openai/gpt-4o-mini `
  smarthome-dspy-resolver
```

## smarthome 側設定

`config/config.json` または環境変数で設定します。

```json
{
  "resolver": {
    "mode": "dspy",
    "dspy_endpoint": "http://localhost:8089/resolve",
    "dspy_timeout_seconds": 5
  }
}
```

または環境変数:

- `SMARTHOME_RESOLVER_MODE=dspy`
- `SMARTHOME_RESOLVER_DSPY_ENDPOINT=http://localhost:8089/resolve`
- `SMARTHOME_RESOLVER_DSPY_TIMEOUT_SECONDS=5`

DSPy resolver が未設定または失敗した場合、`smarthome` は legacy resolver (LLM) にフォールバックします。
