# Smarthomeプロジェクト - AI Agent用プロンプト（プロジェクト全体説明）

## プロジェクト概要

このプロジェクトは、家電をコマンドラインで操作するためのGoアプリケーションです。OwnTone、SwitchBot、Yamaha Extended ControlなどのAPIを統合し、スマートホーム環境を一元管理します。

## 主な機能

1. **コマンドラインツール**: サブコマンド形式で家電を操作
2. **Slack Bot**: Socket Modeを使用したSlack連携
3. **MCP Server**: Model Context Protocol対応サーバー
4. **定期実行処理**: 温湿度データのInfluxDB保存（10分間隔）
5. **OpenTelemetry対応**: トレースデータの送信

## アーキテクチャ

### ディレクトリ構造

```
smarthome/
├── main.go                    # エントリーポイント
├── subcommand/                # サブコマンド定義
│   ├── subcommand.go         # サブコマンドの基本構造
│   ├── config.go             # 設定ファイル読み込み
│   ├── start-meeting.go      # 各種サブコマンド実装
│   └── action/               # アクション実装
│       ├── action.go         # Actionインターフェース
│       ├── owntone/          # OwnTone操作
│       ├── switchbot/        # SwitchBot操作
│       └── yamaha/           # Yamaha機器操作
├── server/                   # サーバーモード
│   ├── slack/               # Slack Bot実装
│   ├── mcp/                 # MCPサーバー実装
│   └── cron/                # 定期実行処理
├── internal/                # 内部パッケージ
│   └── otel/               # OpenTelemetry設定
└── config/                  # 設定ファイル
    ├── config.json          # メイン設定
    └── slack.json           # Slack設定

```

### 主要コンポーネント

#### 1. Subcommand（サブコマンド）

**責務**: ユーザーの要求を複数のActionに変換して実行

**構造**:
```go
type Subcommand struct {
    Definition      // 名前、説明、Factory関数
    actions []action.Action  // 実行するアクションのリスト
    ignoreError bool         // エラーを無視するか
}
```

**重要ポイント**:
- サブコマンドは複数のActionを順次実行
- `ignoreError=true`の場合、エラーが発生してもスキップして次のActionを実行
- Did You Mean機能によるコマンド補正（Levenshtein距離で類似コマンドを提案）
- ショートネーム対応（例: "start meeting" → "sm"）

**登録されているサブコマンド例**:
- `start meeting`: 音楽停止 → 照明ON → アンプOFF
- `finish meeting`: その逆の操作
- `start music`: 音楽再生関連の操作
- `search and play`: 音楽検索して再生

#### 2. Action（アクション）

**責務**: 個別の家電操作を実行

**インターフェース**:
```go
type Action interface {
    Run(ctx context.Context, args string) (string, error)
}
```

**種類**:
- **OwnToneアクション**: 音楽再生、一時停止、プレイリスト変更、検索
- **SwitchBotアクション**: シーン実行、デバイスコマンド、温湿度取得
- **Yamahaアクション**: 電源操作、ボリューム調整、シーン設定
- **特殊アクション**:
  - `HelpAction`: ヘルプ表示
  - `NoopAction`: 何もしない（テスト用）
  - `TokenizeAction`: Kagome（形態素解析）

#### 3. Config（設定管理）

**設定ファイル**: `config/config.json`

**構造**:
```go
type Config struct {
    Owntone   owntone.Config   // OwnTone API設定
    Switchbot switchbot.Config // SwitchBot API設定
    Yamaha    yamaha.Config    // Yamaha API設定
    Influxdb  influxdb.Config  // InfluxDB設定
    Commands  Commands         // 利用可能なコマンド一覧
}
```

**バリデーション**:
- 各APIの必須パラメータチェック
- URL形式の検証
- トークン・シークレットの存在確認

#### 4. Server Modes（サーバーモード）

##### Slack Bot
- **起動方法**: `smarthome -server`
- **Socket Mode**: WebSocketによるリアルタイム通信
- **対応形式**:
  - メンション: `@bot start music`
  - Slash Command: `/start-music`
- **Cron統合**: サーバー起動時に定期処理も開始

##### MCP Server
- **起動方法**: `smarthome -mcp`
- **プロトコル**: Model Context Protocol (stdio)
- **機能**: 全サブコマンドをMCPツールとして公開
- **引数対応**: Definitionに定義されたArgsを自動的にMCPパラメータに変換

##### Cron処理
- **間隔**: 10分ごと
- **処理内容**: SwitchBot温湿度計のデータをInfluxDBに保存
- **対象デバイス**: `server/cron/record-temperature.go`で指定

### データフロー

#### コマンドライン実行の流れ
```
1. main.go::run()
   ↓
2. config読み込み (subcommand.LoadConfig)
   ↓
3. サブコマンド検索 (Commands.Find)
   ↓
4. Subcommand初期化 (Definition.Init)
   ↓
5. Subcommand実行 (Subcommand.Exec)
   ↓
6. 各Action順次実行 (Action.Run)
   ↓
7. 結果表示
```

#### Slack Bot実行の流れ
```
1. Slack Event受信
   ↓
2. メンション or Slash Command判定
   ↓
3. メッセージパース（ボットID除去）
   ↓
4. サブコマンド検索・実行（コマンドラインと同じ）
   ↓
5. Slackに結果返信
```

#### MCP Tool実行の流れ
```
1. MCP Call Tool Request受信
   ↓
2. ツール名からDefinition特定
   ↓
3. 引数を文字列に変換
   ↓
4. Subcommand実行（コマンドラインと同じ）
   ↓
5. MCP Tool Result返却
```

### エラーハンドリング戦略

**基本方針**:
- `panic`は使用せず、エラーを返却
- エラーラッピング（`fmt.Errorf("%w", err)`）で文脈を追加
- 呼び出し元で適切に処理

**Subcommandのエラー処理**:
```go
if s.ignoreError && err != nil {
    // エラーをスキップして続行
    msg = fmt.Sprintf("skip error\t %v\n", err)
} else if err != nil {
    // エラーを返して停止
    return "", err
}
```

### OpenTelemetry統合

**環境変数**:
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLPコレクターのエンドポイント
- `OTEL_SERVICE_NAME`: サービス名（デフォルト: "smarthome"）

**トレース対象**:
- Subcommand実行
- 各Action実行
- HTTP通信（OwnTone、SwitchBot、Yamaha）

**スパン属性**:
```go
span.SetAttributes(
    attribute.String("subcommand.name", s.Name),
    attribute.String("subcommand.args", args),
)
```

## 依存ライブラリ

**主要な外部ライブラリ**:
- `github.com/slack-go/slack`: Slack API/Socket Mode
- `github.com/mark3labs/mcp-go`: MCP実装
- `github.com/hbollon/go-edlib`: Levenshtein距離（Did You Mean）
- `github.com/ikawaha/kagome`: 形態素解析
- `go.opentelemetry.io/otel`: OpenTelemetry
- `github.com/robfig/cron/v3`: Cron処理

## 設計思想

### 1. 拡張性
- 新しいサブコマンドは`Definition`を追加するだけ
- 新しいActionは`Action`インターフェースを実装するだけ
- 新しいAPIは`action/`配下に新しいパッケージを作成

### 2. 再利用性
- Subcommandは複数のActionを組み合わせて構成
- 同じActionを異なるSubcommandで再利用可能
- Client実装（owntone.Client等）は複数のActionで共有

### 3. テスタビリティ
- インターフェースベースの設計
- 依存性注入（Factory関数でClientを注入）
- モックの作成が容易

### 4. 統一インターフェース
- コマンドライン、Slack、MCPで同じSubcommand/Actionを使用
- 実行方法が異なっても処理ロジックは共通

## セキュリティ考慮事項

1. **認証情報の管理**
   - 設定ファイルは`.gitignore`に追加
   - トークン・シークレットはJSONファイルで管理
   - 環境変数での設定は未対応（改善案あり）

2. **API通信**
   - SwitchBot: HMAC認証（Sign/nonce生成）
   - Slack: Bot Token + App Token
   - HTTPS通信

3. **入力検証**
   - 設定ファイルのバリデーション実装
   - ユーザー入力はサブコマンド検索時にチェック

## パフォーマンス特性

1. **キャッシュ**
   - SwitchBot: デバイス名・シーン名をキャッシュ
   - 初回アクセス時にAPI呼び出し、以降はメモリ参照

2. **並行処理**
   - 現在は各Actionを順次実行
   - 並行実行の可能性は今後の改善案

3. **API呼び出し**
   - 各ActionごとにHTTPリクエスト
   - リトライ機構は未実装

## テスト状況

**カバレッジ**:
- `subcommand/action/`: 高いカバレッジ
- `main.go`、`server/`: カバレッジ不足

**テストパターン**:
- HTTPクライアントのモック（`httptest.Server`）
- 正常系・異常系の両方をテスト

## 既知の課題と改善計画

詳細は`docs/improvement-plan.md`を参照。

**重要な課題**:
1. TODOコメントの解決
2. ロギングの構造化（slog導入）
3. HTTPクライアント処理の共通化
4. テストカバレッジ向上
5. CI/CDパイプライン強化

## 開発のベストプラクティス

1. **新しいAPIを追加する場合**
   - `subcommand/action/<api名>/`パッケージを作成
   - `Client`構造体とコンストラクタを実装
   - 各操作を`Action`として実装

2. **新しいSubcommandを追加する場合**
   - `subcommand/<command-name>.go`ファイルを作成
   - `NewXXXDefinition()`関数を実装
   - `NewCommands()`に追加

3. **エラーハンドリング**
   - `panic`は使わない
   - `fmt.Errorf("%w", err)`でラッピング
   - 文脈情報を追加

4. **テスト作成**
   - `*_test.go`ファイルに実装
   - `httptest.Server`でモック作成
   - テーブル駆動テストを推奨

## 参考リンク

- [OwnTone API](https://owntone.github.io/owntone-server/json-api/)
- [SwitchBot API](https://github.com/OpenWonderLabs/SwitchBotAPI)
- [Yamaha Extended Control](https://github.com/rsc-dev/pyamaha/blob/master/doc/YXC_API_Spec_Basic_v2.0.pdf)
- [Model Context Protocol](https://modelcontextprotocol.io/)
