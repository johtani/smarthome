### プロジェクト概要
このプロジェクトは、家の家電をコマンドライン、Slack、またはMCP（Model Context Protocol）経由で操作するためのGo製ツールです。
SwitchBot、OwnTone、YamahaなどのAPIを統合しています。

### ディレクトリ構造と役割
- `main.go`: エントリーポイント。サーバーモード（Slack, MCP）とコマンドラインモードの切り替えを行います。
- `subcommand/`: 各種操作（エアコンON、音楽再生など）の定義。
  - `subcommand.go`: サブコマンドの基盤となる構造体（`Subcommand`, `Definition`）とコマンド一覧。
  - `config.go`: 設定ファイルの読み込み。
- `subcommand/action/`: 実際のAPI呼び出しなどの最小単位の処理（Action）。
  - `action.go`: `Action` インターフェースの定義。
  - `switchbot/`, `owntone/`, `yamaha/`: 各サービスごとの Action 実装。
- `server/`: サーバー機能。
  - `slack/`: Slack Socket Mode ボットの実装。
  - `mcp/`: MCP サーバーの実装。
  - `cron/`: 定期実行処理（温湿度計のデータ取得など）。
- `internal/otel/`: OpenTelemetryの設定。

### 開発ガイドライン
- **ブランチ作成**:
  - 作業を開始する前に、必ず新しいブランチを作成してください（例: `feature/xxxx`）。
- **新しいコマンドの追加**:
  1. `subcommand/` 直下に `xxxx.go` を作成し、`Definition` とファクトリ関数を実装します。
  2. `subcommand/subcommand.go` の `NewCommands()` 関数に作成した `Definition` を追加します。
- **Actionの利用**:
  - 可能な限り `subcommand/action/` 配下の既存の Action を組み合わせてサブコマンドを構成してください。
  - 新しい API 連携が必要な場合は、新しい Action を作成します。
- **トレーサビリティ**:
  - 各 Action やサブコマンドの実行には OpenTelemetry による計測を組み込んでください。
  - `otel.Tracer("...").Start(ctx, "...")` を使用します。
- **テスト**:
  - 新しい機能や修正には、対応するテストファイル（`*_test.go`）を作成または更新してください。
- **設定**:
  - 設定項目は `subcommand/config.go` の `Config` 構造体に追加し、`config/config.json.sample` も更新してください。
