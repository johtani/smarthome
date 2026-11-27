---
name: evaluate-resolvers
description: smarthome の固定データで resolver のモデル・経路・最適化前後を比較評価し、誤起動と評価条件を含めて結果を報告する。
---

# 固定データで resolver を比較する

リポジトリ直下を起点に、[評価バッチの手順](../../../tools/dspy/README.md)から比較目的に合う経路を選ぶ。

- モデル間の評価や成果物生成には `optimize_and_evaluate.py --evaluation-manifest` を使う。確定済み manifest を各モデルに渡し、データ、分割、catalog、prompt version、seed、閾値を揃える。成果物の生成結果と本番採用の判断は区別する。
- 同一モデルで Go 直呼び・DSPy 最適化前・最適化後を比べる場合は `compare_resolvers.py` を使う。manifest が指す train/dev/test と catalog の SHA-256 が一致することを確認してから渡す。同じ配信先・モデルを指定し、推論設定とキャッシュ条件を記録する。学習は train、比較の分母は固定 test とする。
- 確定データが必要な場合は、同じ README の「レビュー済みデータの確定」に従う。レビュー前の候補は正解として扱わない。実ログを含む入力と行別レポートの保存先は Git 管理外の `tmp/` とする。

経路ごとのコマンド・引数の同時正解数と母数を示す。3経路比較では Raw と Final を分け、対象外入力への誤起動数／対象外件数、未解決数、形式エラー数、応答時間中央値を並べる。改善・悪化した例と、母数や測定条件から判断を保留した点を添える。モデル間比較ではバッチの command／args 精度、catalog 違反率、ゲート結果を使い、レポートで測定していない指標は未測定と記す。
