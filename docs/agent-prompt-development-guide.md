# Smarthomeプロジェクト - AI Agent用プロンプト（開発ガイド）

このガイドは、Smarthomeプロジェクトに新しい機能を追加する際の具体的な手順を説明します。

## 前提知識

このガイドを読む前に、`agent-prompt-project-overview.md`でプロジェクト全体のアーキテクチャを理解してください。

## 開発の基本フロー

1. 要件定義: どのような操作を実装するか決定
2. Action実装: 個別の家電操作を実装
3. Subcommand実装: Actionを組み合わせたコマンドを作成
4. テスト作成: ユニットテストを実装
5. 動作確認: 実際の環境でテスト

## 新しいサブコマンドの追加

### ケーススタディ: "start meeting" の実装

`subcommand/start-meeting.go`を例に、新しいサブコマンドの追加方法を説明します。

#### Step 1: ファイル作成

`subcommand/`ディレクトリに新しいファイルを作成します。

命名規則: `<command-name>.go`（スペースは"-"に変換）

```bash
subcommand/start-meeting.go
```

#### Step 2: Definition関数の実装

```go
package subcommand

const StartMeetingCmd = "start meeting"

func NewStartMeetingDefinition() Definition {
    return Definition{
        Name:        StartMeetingCmd,
        Description: "Actions before starting meeting",
        Factory:     NewStartMeetingSubcommand,
    }
}
```

**ポイント**:
- `Name`: コマンド名（スペース区切りOK）
- `Description`: ヘルプに表示される説明
- `Factory`: Subcommandを生成する関数
- `shortnames`: ショートネームを設定する場合は追加（例: `[]string{"sm"}`）
- `Args`: 引数が必要な場合は追加

#### Step 3: Factory関数の実装

```go
func NewStartMeetingSubcommand(definition Definition, config Config) Subcommand {
    // 各APIのクライアントを初期化
    owntoneClient := owntone.NewClient(config.Owntone)
    switchbotClient := switchbot.NewClient(config.Switchbot)
    yamahaClient := yamaha.NewClient(config.Yamaha)

    return Subcommand{
        Definition: definition,
        actions: []action.Action{
            // 実行したいActionを順番に並べる
            owntone.NewPauseAction(owntoneClient),
            switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneId),
            yamaha.NewPowerOffAction(yamahaClient),
        },
        ignoreError: true,  // エラーが発生しても続行する場合はtrue
    }
}
```

**ignoreErrorの使い分け**:
- `true`: 一部の操作が失敗しても他の操作を続行したい場合（例: 既に停止している音楽を停止）
- `false`: エラーが発生したら即座に停止したい場合

#### Step 4: Commands構造体に登録

`subcommand/subcommand.go`の`NewCommands()`関数に追加します。

```go
func NewCommands() Commands {
    return Commands{
        Definitions: []Definition{
            // ... 既存のDefinition
            NewStartMeetingDefinition(),  // ← 追加
            // ... 他のDefinition
        },
    }
}
```

### 引数を持つSubcommandの実装

#### ケーススタディ: "search and play" の実装

```go
func NewSearchAndPlayMusicCmdDefinition() Definition {
    return Definition{
        Name:        SearchAndPlayMusicCmd,
        Description: "Search tracks and play",
        Factory:     NewSearchAndPlayMusicSubcommand,
        Args: []Arg{
            {
                Name:        "query",
                Description: "Search query for music",
                Required:    true,  // 必須引数
            },
        },
    }
}
```

**引数の利用**:
- Actionの`Run(ctx, args string)`で`args`として渡される
- `args`は文字列として受け取るため、必要に応じてパース

#### Enumを持つ引数の例

```go
Args: []Arg{
    {
        Name:        "type",
        Description: "Tokenizer type",
        Required:    true,
        Enum:        []string{"ipa", "uni", "neologd"},
    },
}
```

**MCPサーバーでの効果**:
- MCP Tool定義に自動的に反映
- Enumは選択肢として提供される

## 新しいActionの追加

### ケーススタディ: OwnTone操作の実装

#### Step 1: パッケージ構造

```
subcommand/action/owntone/
├── client.go          # APIクライアント
├── client_test.go     # クライアントのテスト
├── play.go            # 再生Action
├── play_test.go       # 再生Actionのテスト
├── pause.go           # 一時停止Action
└── ...                # その他のAction
```

#### Step 2: Client実装

```go
// subcommand/action/owntone/client.go
package owntone

import (
    "github.com/johtani/smarthome/subcommand/action/internal"
)

type Client struct {
    config Config
    internal.HttpAction  // 共通のHTTP処理を継承
}

type Config struct {
    Url string `json:"url"`
}

func (c Config) Validate() error {
    if c.Url == "" {
        return fmt.Errorf("OwnTone.url is required")
    }
    return nil
}

func NewClient(config Config) Client {
    return Client{
        config:     config,
        HttpAction: internal.NewHttpAction(),
    }
}
```

**ポイント**:
- `Config`構造体: API接続に必要な情報
- `Validate()`メソッド: 設定の妥当性チェック
- `internal.HttpAction`: 共通のHTTP処理を再利用

#### Step 3: Action実装

```go
// subcommand/action/owntone/play.go
package owntone

import (
    "context"
    "fmt"
    "go.opentelemetry.io/otel"
)

type PlayAction struct {
    client Client
}

func NewPlayAction(client Client) PlayAction {
    return PlayAction{client: client}
}

func (a PlayAction) Run(ctx context.Context, args string) (string, error) {
    // OpenTelemetryトレース
    ctx, span := otel.Tracer("action").Start(ctx, "owntone.PlayAction")
    defer span.End()

    // API呼び出し
    err := a.client.Play(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to play: %w", err)
    }

    return "Music started", nil
}
```

**ポイント**:
- `Action`インターフェースを実装（`Run`メソッド）
- OpenTelemetryでトレース
- エラーはラッピングして返す

#### Step 4: ClientのAPI呼び出しメソッド実装

```go
// subcommand/action/owntone/client.go
func (c Client) Play(ctx context.Context) error {
    ctx, span := otel.Tracer("client").Start(ctx, "owntone.Client.Play")
    defer span.End()

    url := fmt.Sprintf("%s/api/player/play", c.config.Url)

    // PUTリクエスト送信
    res, err := c.HttpAction.Put(ctx, url, "", "application/json")
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    defer c.HttpAction.CloseResponse(res)

    // ステータスコードチェック
    if res.StatusCode != http.StatusNoContent {
        return fmt.Errorf("unexpected status code: %d", res.StatusCode)
    }

    return nil
}
```

**HttpActionの活用**:
- `Put(ctx, url, body, contentType)`: PUT リクエスト
- `Get(ctx, url)`: GET リクエスト
- `Post(ctx, url, body, contentType)`: POST リクエスト
- `CloseResponse(res)`: レスポンスの適切なクローズ

## 新しいAPI統合の追加

### ケーススタディ: 新しいスマートホームAPI "NewDevice" の統合

#### Step 1: パッケージ作成

```bash
mkdir subcommand/action/newdevice
```

#### Step 2: Config定義

```go
// subcommand/config.go に追加
type Config struct {
    Owntone   owntone.Config   `json:"Owntone"`
    Switchbot switchbot.Config `json:"Switchbot"`
    Yamaha    yamaha.Config    `json:"Yamaha"`
    NewDevice newdevice.Config `json:"NewDevice"`  // ← 追加
    Influxdb  influxdb.Config  `json:"Influxdb"`
    Commands  Commands
}

func (c Config) validate() error {
    // ... 既存のバリデーション
    err = c.NewDevice.Validate()  // ← 追加
    if err != nil {
        errs = append(errs, err.Error())
    }
    // ...
}
```

#### Step 3: NewDeviceパッケージ実装

```go
// subcommand/action/newdevice/client.go
package newdevice

import (
    "context"
    "fmt"
    "github.com/johtani/smarthome/subcommand/action/internal"
)

type Config struct {
    ApiUrl    string `json:"api_url"`
    ApiKey    string `json:"api_key"`
    DeviceId  string `json:"device_id"`
}

func (c Config) Validate() error {
    if c.ApiUrl == "" {
        return fmt.Errorf("NewDevice.api_url is required")
    }
    if c.ApiKey == "" {
        return fmt.Errorf("NewDevice.api_key is required")
    }
    return nil
}

type Client struct {
    config Config
    internal.HttpAction
}

func NewClient(config Config) Client {
    return Client{
        config:     config,
        HttpAction: internal.NewHttpAction(),
    }
}

// API操作メソッド
func (c Client) TurnOn(ctx context.Context) error {
    // 実装
}

func (c Client) TurnOff(ctx context.Context) error {
    // 実装
}
```

#### Step 4: Action実装

```go
// subcommand/action/newdevice/turn-on.go
package newdevice

import (
    "context"
    "go.opentelemetry.io/otel"
)

type TurnOnAction struct {
    client Client
}

func NewTurnOnAction(client Client) TurnOnAction {
    return TurnOnAction{client: client}
}

func (a TurnOnAction) Run(ctx context.Context, args string) (string, error) {
    ctx, span := otel.Tracer("action").Start(ctx, "newdevice.TurnOnAction")
    defer span.End()

    err := a.client.TurnOn(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to turn on: %w", err)
    }

    return "NewDevice turned on", nil
}
```

#### Step 5: Subcommandで利用

```go
// subcommand/new-command.go
func NewNewCommandSubcommand(definition Definition, config Config) Subcommand {
    newdeviceClient := newdevice.NewClient(config.NewDevice)

    return Subcommand{
        Definition: definition,
        actions: []action.Action{
            newdevice.NewTurnOnAction(newdeviceClient),
        },
        ignoreError: false,
    }
}
```

#### Step 6: 設定ファイル更新

```json
// config/config.json
{
  "Owntone": { ... },
  "Switchbot": { ... },
  "Yamaha": { ... },
  "NewDevice": {
    "api_url": "https://api.newdevice.com",
    "api_key": "your-api-key",
    "device_id": "device-123"
  },
  "Influxdb": { ... }
}
```

## テストの作成

### テストの基本構造

```go
// subcommand/action/owntone/play_test.go
package owntone

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestPlayAction_Run(t *testing.T) {
    // テストケース定義
    tests := []struct {
        name           string
        statusCode     int
        wantErr        bool
        wantErrMessage string
        wantMessage    string
    }{
        {
            name:        "success",
            statusCode:  http.StatusNoContent,
            wantErr:     false,
            wantMessage: "Music started",
        },
        {
            name:           "server error",
            statusCode:     http.StatusInternalServerError,
            wantErr:        true,
            wantErrMessage: "unexpected status code",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // モックサーバー作成
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // リクエストの検証
                if r.Method != http.MethodPut {
                    t.Errorf("expected PUT, got %s", r.Method)
                }
                if r.URL.Path != "/api/player/play" {
                    t.Errorf("unexpected path: %s", r.URL.Path)
                }

                // レスポンス返却
                w.WriteHeader(tt.statusCode)
            }))
            defer server.Close()

            // テスト対象の準備
            config := Config{Url: server.URL}
            client := NewClient(config)
            action := NewPlayAction(client)

            // 実行
            msg, err := action.Run(context.Background(), "")

            // 検証
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if err != nil && tt.wantErrMessage != "" {
                if !strings.Contains(err.Error(), tt.wantErrMessage) {
                    t.Errorf("expected error containing %q, got %q", tt.wantErrMessage, err.Error())
                }
            }
            if msg != tt.wantMessage {
                t.Errorf("expected message %q, got %q", tt.wantMessage, msg)
            }
        })
    }
}
```

### テストのベストプラクティス

1. **テーブル駆動テスト**: 複数のケースを配列で定義
2. **httptest.Server**: HTTPクライアントのモック
3. **リクエスト検証**: メソッド、URL、ヘッダー、ボディをチェック
4. **エラーメッセージ検証**: `strings.Contains`で部分一致チェック

### Subcommandのテスト

```go
// subcommand/subcommand_test.go
func TestSubcommand_Exec(t *testing.T) {
    // Subcommandの作成
    def := Definition{
        Name:        "test command",
        Description: "Test",
    }

    // モックActionの作成
    mockAction := &MockAction{
        runFunc: func(ctx context.Context, args string) (string, error) {
            return "success", nil
        },
    }

    subcommand := Subcommand{
        Definition:  def,
        actions:     []action.Action{mockAction},
        ignoreError: false,
    }

    // 実行
    msg, err := subcommand.Exec(context.Background(), "")

    // 検証
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if msg != "success" {
        t.Errorf("expected 'success', got %q", msg)
    }
}
```

## よくある実装パターン

### パターン1: 複数のAPIを順次呼び出し

```go
func NewComplexSubcommand(definition Definition, config Config) Subcommand {
    owntoneClient := owntone.NewClient(config.Owntone)
    switchbotClient := switchbot.NewClient(config.Switchbot)

    return Subcommand{
        Definition: definition,
        actions: []action.Action{
            // 順番に実行される
            owntone.NewPauseAction(owntoneClient),
            switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneId),
            owntone.NewClearQueueAction(owntoneClient),
        },
        ignoreError: false,  // 1つでも失敗したら停止
    }
}
```

### パターン2: 引数に応じた動的な処理

```go
type ConditionalAction struct {
    client Client
}

func (a ConditionalAction) Run(ctx context.Context, args string) (string, error) {
    // 引数をパース
    if strings.HasPrefix(args, "type:") {
        actionType := strings.TrimPrefix(args, "type:")
        // typeに応じた処理
        switch actionType {
        case "ipa":
            return a.processIpa(ctx)
        case "uni":
            return a.processUni(ctx)
        default:
            return "", fmt.Errorf("unknown type: %s", actionType)
        }
    }

    // デフォルト処理
    return a.processDefault(ctx, args)
}
```

### パターン3: APIレスポンスの加工

```go
func (a DisplayAction) Run(ctx context.Context, args string) (string, error) {
    // APIからデータ取得
    playlists, err := a.client.GetPlaylists(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to get playlists: %w", err)
    }

    // ユーザーフレンドリーな形式に変換
    var builder strings.Builder
    builder.WriteString("Available Playlists:\n")
    for _, playlist := range playlists {
        builder.WriteString(fmt.Sprintf("- %s (ID: %s)\n", playlist.Name, playlist.ID))
    }

    return builder.String(), nil
}
```

### パターン4: キャッシュの利用

```go
type CachedClient struct {
    config      Config
    cache       map[string]string
    cacheLoaded bool
    internal.HttpAction
}

func (c *CachedClient) GetDeviceName(ctx context.Context, deviceId string) (string, error) {
    // キャッシュ初期化
    if !c.cacheLoaded {
        if err := c.loadCache(ctx); err != nil {
            return "", err
        }
        c.cacheLoaded = true
    }

    // キャッシュから取得
    if name, ok := c.cache[deviceId]; ok {
        return name, nil
    }

    return "", fmt.Errorf("device not found: %s", deviceId)
}

func (c *CachedClient) loadCache(ctx context.Context) error {
    // APIから全デバイス取得
    devices, err := c.fetchAllDevices(ctx)
    if err != nil {
        return err
    }

    // キャッシュに格納
    c.cache = make(map[string]string)
    for _, device := range devices {
        c.cache[device.ID] = device.Name
    }

    return nil
}
```

## エラーハンドリングのパターン

### パターン1: エラーラッピング

```go
func (a SomeAction) Run(ctx context.Context, args string) (string, error) {
    result, err := a.client.DoSomething(ctx)
    if err != nil {
        // 文脈情報を追加してラッピング
        return "", fmt.Errorf("failed to do something: %w", err)
    }

    return result, nil
}
```

### パターン2: 詳細なエラー情報

```go
func (c Client) ApiCall(ctx context.Context) error {
    res, err := c.HttpAction.Get(ctx, url)
    if err != nil {
        return fmt.Errorf("HTTP request failed (url=%s): %w", url, err)
    }
    defer c.HttpAction.CloseResponse(res)

    if res.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d, url: %s, headers: %v",
            res.StatusCode, url, res.Header)
    }

    return nil
}
```

### パターン3: 部分的な成功の扱い

```go
func NewRobustSubcommand(definition Definition, config Config) Subcommand {
    return Subcommand{
        Definition: definition,
        actions: []action.Action{
            action1,
            action2,
            action3,
        },
        ignoreError: true,  // エラーが起きても続行
    }
}

// Subcommand.Execでは以下のように処理される:
// - action1が失敗 → エラーメッセージを記録し、action2へ
// - action2が成功 → 結果を記録し、action3へ
// - action3が失敗 → エラーメッセージを記録
// - 全ての結果を結合して返却
```

## OpenTelemetryの活用

### トレースの追加

```go
func (a MyAction) Run(ctx context.Context, args string) (string, error) {
    // トレース開始
    ctx, span := otel.Tracer("action").Start(ctx, "mypackage.MyAction")
    defer span.End()

    // 属性の追加
    span.SetAttributes(
        attribute.String("args", args),
        attribute.String("operation", "my_operation"),
    )

    // 処理実行
    result, err := a.doWork(ctx)
    if err != nil {
        // エラー記録
        span.RecordError(err)
        return "", err
    }

    // 結果を属性として記録
    span.SetAttributes(attribute.String("result", result))

    return result, nil
}
```

### ネストしたトレース

```go
func (c Client) ComplexOperation(ctx context.Context) error {
    ctx, span := otel.Tracer("client").Start(ctx, "ComplexOperation")
    defer span.End()

    // Step 1
    if err := c.step1(ctx); err != nil {
        return err
    }

    // Step 2
    if err := c.step2(ctx); err != nil {
        return err
    }

    return nil
}

func (c Client) step1(ctx context.Context) error {
    // このトレースはComplexOperationの子スパンになる
    ctx, span := otel.Tracer("client").Start(ctx, "step1")
    defer span.End()

    // 処理
    return nil
}
```

## デバッグのヒント

### ログ出力

現在は`fmt.Printf`を使用していますが、将来的には構造化ログ（slog）への移行が推奨されます。

```go
// 現在の方法
fmt.Printf("Processing device: %s\n", deviceId)

// 推奨される方法（改善計画に含まれる）
slog.Info("processing device", "device_id", deviceId)
```

### Slackでのデバッグ

Slack Botモードでは、設定ファイル`config/slack.json`の`debug`フラグを有効にすると詳細なログが出力されます。

```json
{
  "bot_token": "xoxb-...",
  "app_token": "xapp-...",
  "debug": true
}
```

### MCPサーバーのデバッグ

MCPサーバーは`WithLogging()`オプションで起動しているため、標準エラー出力にログが出力されます。

```go
server.NewMCPServer(
    "Smart Home MCP",
    "0.1.0",
    server.WithLogging(),  // ログ出力有効
)
```

## 設定ファイルの管理

### サンプルファイルの更新

新しい設定項目を追加した場合は、サンプルファイルも更新してください。

```json
// config/config.json.sample
{
  "Owntone": {
    "url": "http://localhost:3689"
  },
  "Switchbot": {
    "token": "YOUR_SWITCHBOT_TOKEN",
    "secret": "YOUR_SWITCHBOT_SECRET",
    "light_scene_id": "YOUR_SCENE_ID"
  },
  "NewDevice": {
    "api_url": "https://api.newdevice.com",
    "api_key": "YOUR_API_KEY",
    "device_id": "YOUR_DEVICE_ID"
  }
}
```

### 必須項目と任意項目

```go
func (c Config) Validate() error {
    var errs []string

    // 必須項目
    if c.ApiUrl == "" {
        errs = append(errs, "api_url is required")
    }

    // 任意項目（デフォルト値を設定）
    if c.Timeout == 0 {
        c.Timeout = 30  // デフォルト30秒
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errs, ", "))
    }
    return nil
}
```

## コードレビューのチェックリスト

新しいコードをコミットする前に、以下を確認してください：

- [ ] `Action`インターフェースを実装している
- [ ] OpenTelemetryトレースを追加している
- [ ] エラーは`fmt.Errorf("%w", err)`でラッピングしている
- [ ] `panic`を使用していない
- [ ] テストを作成している
- [ ] 設定項目のバリデーションを実装している
- [ ] サンプル設定ファイルを更新している
- [ ] README.mdを更新している（必要な場合）
- [ ] `NewCommands()`にDefinitionを登録している

## トラブルシューティング

### 問題: "Sorry, I cannot understand what you want..."

**原因**: サブコマンドが見つからない

**解決方法**:
1. `NewCommands()`に`Definition`を追加したか確認
2. コマンド名のスペルミスがないか確認
3. Did You Mean機能で提案されたコマンドを確認

### 問題: "設定の読み込みに失敗"

**原因**: `config/config.json`が不正

**解決方法**:
1. JSONの構文エラーがないか確認
2. 必須項目が全て設定されているか確認
3. `Validate()`メソッドでエラーメッセージを確認

### 問題: HTTPリクエストが失敗する

**原因**: APIエンドポイントの問題

**解決方法**:
1. URLが正しいか確認
2. ネットワーク接続を確認
3. APIトークン・認証情報を確認
4. `httptest.Server`を使ってモックテストを作成し、問題を切り分け

### 問題: Slack Botが反応しない

**原因**: 設定またはメンション形式の問題

**解決方法**:
1. `config/slack.json`の設定を確認
2. メンション形式を確認（`@bot start music`）
3. Slash Commandの場合は登録されているか確認
4. `debug: true`にしてログを確認

## 参考実装

プロジェクト内の実装例：

- **シンプルなAction**: `subcommand/action/owntone/pause.go`
- **複雑なAction**: `subcommand/action/owntone/search.go`
- **キャッシュを使うAction**: `subcommand/action/switchbot/client.go`
- **引数を使うSubcommand**: `subcommand/search-and-play.go`
- **複数Actionを組み合わせ**: `subcommand/start-meeting.go`

## まとめ

Smarthomeプロジェクトの開発は以下のステップで進めます：

1. **要件定義**: どのような操作を実装するか明確にする
2. **設計**: 既存のActionを使うか、新しいActionを作るか決定
3. **実装**: Action → Subcommand → 設定 の順に実装
4. **テスト**: ユニットテストを作成し、動作確認
5. **ドキュメント**: README.mdやサンプル設定を更新

このガイドに従うことで、プロジェクトのアーキテクチャに沿った一貫性のあるコードを書くことができます。
