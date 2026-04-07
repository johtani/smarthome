# アーキテクチャ概要

このドキュメントは Smarthome の全体像を簡潔に示すためのものです。
設定・運用・詳細仕様は [README.md](../README.md) を参照してください。

## コンポーネント

- CLI: サブコマンドを受け付けて家電操作を実行
- Slack Server (`-server`): Slack メンション/Slash Command からサブコマンドを実行
- MCP Server (`-mcp`): MCP ツールとしてサブコマンドを公開
- Cron: 温湿度データを定期収集して InfluxDB に保存

## ディレクトリ責務

- `main.go`: エントリーポイント
- `subcommand/`: サブコマンド定義と実行制御
- `subcommand/action/`: 外部APIごとの具体アクション実装
- `server/slack/`: Slack 連携
- `server/mcp/`: MCP 連携
- `server/cron/`: 定期実行処理
- `internal/`: 共通内部実装（例: OpenTelemetry）
- `config/`: 設定ファイルサンプル

## 実行フロー

1. 設定を読み込む
2. 入力（CLI/Slack/MCP）をサブコマンドに解決する
3. サブコマンドが複数 Action を順次実行する
4. 実行結果を呼び出し元へ返す

## Action のトレース方針

- `Action.Run` のスパン作成は `subcommand/action.StartRunSpan` を使って共通化する
- `StartRunSpan` は `action.args` 属性を付与する
- 新しい Action を追加するときは、`otel.Tracer(...).Start(...)` を直接呼ばずに `StartRunSpan` を使う
