# Issue: 自然言語によるコマンド入力

## 概要
ユーザーが入力した自然言語（例：「エアコンをつけて」「ジャズを流して」）を解析し、適切なサブコマンドを選択・実行する仕組みを導入します。

## 背景・目的
現在、本ツールはコマンド名やエイリアスによる厳密な一致、または編集距離（Levenshtein distance）による曖昧検索をサポートしています。
しかし、より直感的で柔軟な操作を実現するため、LLM（OpenAI APIやllama.cpp）を活用した自然言語解析によるコマンド解決を検討しています。

## 提案されている実装プラン

### 1. 構成案

#### A. LLM Action の作成 (`subcommand/action/llm/`)
- OpenAI互換API（OpenAI, llama.cppなど）を呼び出すクライアントを実装。
- **Structured Outputs** や **Function Calling** を利用し、実行すべきコマンド名と引数をJSON形式で取得する。

#### B. コマンド解決エンジンの拡張 (`subcommand/subcommand.go`)
`Commands.Find` メソッドのロジックを以下のように拡張：
1. **既存の厳密一致/曖昧検索**: 低コストなため、まずこれを試行。
2. **LLMによる解決**: 上記で見つからない場合、LLM Action を呼び出してコマンドを特定。
   - プロンプトには `subcommand.Definition` から生成した利用可能なコマンドリストを含める。

#### C. 設定の追加 (`subcommand/config.go`)
- `LLM_API_KEY`, `LLM_ENDPOINT`, `LLM_MODEL` などの設定項目を追加。

### 2. 実装ステップ
1. **LLM Action の実装**: OpenAI互換APIとの通信基盤を作成。
2. **プロンプト構築ロジック**: `Definition` の情報をプロンプトに変換する機能を実装。
3. **サーバー統合**: `server/slack/handler.go` などの `findAndExec` 内でLLMによる解決を試みるよう変更。

### 3. メリット
- **既存資産の活用**: 既に定義されている `Description`（説明文）をそのままLLMへの指示として利用可能。
- **トレーサビリティ**: OpenTelemetry を使用してLLMの呼び出し時間やトークン消費量を計測可能。

## 注意事項
- **レスポンスタイム**: LLMの応答待ちが発生するため、UX上の考慮（「考え中...」の返信など）が必要。
- **フォールバック**: 不確実な場合は無理に実行せず、確認を促す仕組みにする。
