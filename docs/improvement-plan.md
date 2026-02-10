# スマートホームプロジェクト改善プラン

最終更新: 2026-02-08

## フェーズ1: コード品質の基礎改善

### 1. TODOコメントの解決

- [x] `subcommand/subcommand.go:34` - ignoreError時のエラーメッセージ処理を決定・実装
- [x] `subcommand/subcommand.go:77,184` - Did You Mean機能の候補ソート実装
- [x] `server/mcp/server.go:71` - typeパラメータの特殊処理をより汎用的な設計に変更
- [x] `subcommand/action/kagome/kagome.go:48` - Slack用返信フォーマットの実装検討

### 2. エラーハンドリングの改善

- [x] `subcommand/config.go` - panicをエラー返却に変更
- [x] `server/cron/cron.go` - panicをエラー返却に変更
- [x] `server/mcp/server.go` - panicをエラー返却に変更
- [x] `server/slack/bot-server.go` - panicをエラー返却に変更
- [x] `main.go` - 各呼び出し箇所でエラーハンドリングを追加
- [x] エラーメッセージの一貫性向上（%wを使用したエラーラップ）

### 3. ロギングの構造化

- [x] `fmt.Printf`を構造化ログ（slog）に置き換え
- [x] ログレベル（Debug/Info/Error）の適切な使い分け
- [x] `server/slack/handler.go` の "######### :" プレフィックスを適切なログに変更

## フェーズ2: パフォーマンスと安全性

### 4. キャッシュ実装の最適化

- [x] `subcommand/action/switchbot/client.go` - キャッシュの不要なリセットを削除
- [x] 並行アクセス対策（sync.RWMutex）の追加検討

### 5. HTTPクライアント処理の共通化

- [x] レスポンス処理の重複コード削減
- [x] 共通ヘルパー関数の作成（defer/Close/StatusCodeチェック）

### 6. 定数の整理

- [x] マジックナンバーを定数化
- [x] 設定可能な値は構造体フィールドに移動

## フェーズ3: テストとCI/CD

### 7. テストカバレッジ向上

- [x] `main.go` のテスト追加
- [x] `server/slack/bot-server.go` のテスト追加
- [x] `server/cron/` のテスト追加
- [x] `server/mcp/server.go` のテスト追加

### 8. CI/CDパイプライン強化

- [ ] golangci-lintの追加
- [ ] テストカバレッジレポート追加
- [ ] gosecによるセキュリティスキャン追加

## フェーズ4: ドキュメントと保守性

### 9. Godocコメント追加

- [ ] 公開関数・型・メソッドにドキュメントコメント追加
- [ ] パッケージレベルのdoc.go作成

### 10. 設定管理の改善

- [x] 環境変数での設定オーバーライドサポート
- [x] より詳細なバリデーションエラーメッセージ

---

## 推奨実施順序

**フェーズ1 → フェーズ2 → フェーズ3 → フェーズ4**

各フェーズは独立しているため、必要に応じて選択的に実施可能です。

## 詳細な改善提案

### エラーハンドリングの改善（フェーズ1-2）

現在、設定ファイルの読み込みやcron処理で`panic`を使用していますが、以下の理由でエラー返却に変更すべきです：

- より柔軟なエラーハンドリングが可能
- テストが容易になる
- ライブラリとして利用される場合の安全性向上

### キャッシュの最適化（フェーズ2-4）

`switchbot/client.go`の`CachedClient`実装で、キャッシュが見つからない場合に全体をリセットしています：

```go
// 現在の実装 (74, 90行目)
c.sceneNameCache = map[string]string{}
c.deviceNameCache = map[string]string{}
```

この実装では、キャッシュミス時に既存のキャッシュが全て破棄されます。改善案：

1. キャッシュをリセットせず追加のみ行う
2. 並行アクセスの可能性がある場合は`sync.RWMutex`で保護

### HTTPクライアント処理の共通化（フェーズ2-5）

`owntone/client.go`等で同じパターンが繰り返されています：

```go
defer func(Body io.ReadCloser) {
    _, _ = io.Copy(io.Discard, Body)
    _ = Body.Close()
}(res.Body)
if res.StatusCode != http.StatusXXX {
    return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
}
```

共通ヘルパー関数を作成することで、コードの重複を削減し保守性を向上できます。

### Did You Mean機能の改善（フェーズ1-1）

`subcommand.go:184`のTODOコメントの通り、候補を距離順にソートすることで、より適切な提案が可能になります。

### テストカバレッジ（フェーズ3-7）

現在、以下の重要な部分のテストが不足しています：

- エントリポイント（main.go）
- サーバーモード（slack, cron, mcp）
- 設定ファイルの読み込み・バリデーション

これらのテストを追加することで、リファクタリングや機能追加時の安全性が向上します。

### CI/CDの強化（フェーズ3-8）

現在のGitHub Actionsワークフローは基本的なビルドとテストのみです。以下を追加することを推奨：

```yaml
# 追加推奨項目
- Linter (golangci-lint)
- テストカバレッジレポート (coveralls, codecov等)
- セキュリティスキャン (gosec)
- 依存関係の脆弱性チェック (govulncheck)
```

## 進捗管理

各項目の進捗は、このファイルのチェックボックスを更新することで管理できます。

完了した項目は以下のように更新してください：

```markdown
- [x] 完了した項目
```
