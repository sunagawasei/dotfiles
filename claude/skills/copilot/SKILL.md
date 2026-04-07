---
name: copilot
user-invocable: true
description: |
  GitHub Copilot CLIでコード分析を実行する。
  Usage: /copilot "プロンプト"
  Auto-loaded when Claude considers using the copilot agent.
---

# /copilot

ユーザーの引数をタスクプロンプトとして、copilotサブエージェントを起動する。

## 重要な制約

- copilotサブエージェント内のBashツールで `run_in_background: true` は**絶対に使用しないこと**（バックグラウンド実行するとセッション進行時にkillされる）

## 手順

1. 引数をタスクプロンプトとして受け取る
2. copilotエージェント（subagent_type: copilot）を起動
3. 結果を日本語で提示

## Reference Templates

詳細テンプレートは以下を参照：

- `references/code-review.md` — コードレビュー用プロンプトテンプレート
- `references/error-analysis.md` — エラー診断ワークフロー
- `references/test-generation.md` — テスト提案テンプレート
- `references/documentation.md` — ドキュメントレビューテンプレート
