# スマートホームプロジェクト改善プラン

最終更新: 2026-02-10

## 1. 現在の優先課題

### フェーズ3: CI/CDパイプライン強化
- [x] golangci-lintの導入
    - 静的解析によるコード品質の自動チェック
- [x] テストカバレッジレポートの自動生成
    - GitHub Actionsでの可視化（Coveralls, Codecov等）
- [x] セキュリティスキャンの実施
    - `gosec` による脆弱性診断の自動化
- [x] 依存関係の脆弱性チェック
    - `govulncheck` の導入

### フェーズ4: ドキュメントと保守性
- [ ] Godocコメントの整備
    - 公開関数、型、メソッドへの適切なドキュメント追加
- [ ] パッケージレベルのドキュメント作成
    - 各主要ディレクトリへの `doc.go` 追加による構造の明確化

---

## 2. 完了済みの項目

### フェーズ1: コード品質の基礎改善
- [x] TODOコメントの解決
    - `subcommand/subcommand.go`: ignoreError時の処理実装
    - `subcommand/subcommand.go`: Did You Mean機能のソート実装
    - `server/mcp/server.go`: typeパラメータ処理の汎用化
    - `subcommand/action/kagome/kagome.go`: Slack用返信フォーマット検討
- [x] エラーハンドリングの改善
    - `panic` を廃止し、適切なエラー返却（`error`）に変更
    - `%w` を用いたエラーラップによるトレーサビリティ向上
- [x] ロギングの構造化
    - `fmt.Printf` から `slog` への移行
    - ログレベルの適切な使い分け

### フェーズ2: パフォーマンスと安全性
- [x] キャッシュ実装の最適化
    - `switchbot/client.go`: 不要なキャッシュリセットの削除
    - 並行アクセスを考慮した設計
- [x] HTTPクライアント処理の共通化
    - ヘルパー関数の導入による重複コードの削減
    - ボディの確実なクローズ処理の徹底
- [x] 定数の整理
    - マジックナンバーの定数化と設定値の構造化

### フェーズ3: テストの拡充
- [x] テストカバレッジの向上
    - `main.go`, `server/slack`, `server/cron`, `server/mcp` 等の主要コンポーネントへのテスト追加

### フェーズ4: 設定管理
- [x] 設定管理の改善
    - 環境変数による設定オーバーライドのサポート
    - バリデーションエラーメッセージの充実

---

## 3. 詳細な改善提案（継続中）

### CI/CDの強化（フェーズ3）

現在のGitHub Actionsワークフローを拡張し、以下のステップを追加することを推奨します：

```yaml
# 追加推奨ステップ例
- name: Run Linter
  uses: golangci/golangci-lint-action@v6

- name: Security Scan
  run: go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

- name: Vulnerability Check
  run: go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

### ドキュメントの整備（フェーズ4）

プロジェクトの規模が拡大しているため、Goの標準的なドキュメント形式（Godoc）に準拠することで、新規参画者や将来のメンテナンスが容易になります。特に `action` パッケージや `subcommand` のインターフェース定義には詳細な説明が必要です。
