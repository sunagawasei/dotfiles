---
name: cursor
user-invocable: true
description: |
  Cursor Agent CLIでコード分析を実行する。
  Usage: /cursor "プロンプト"
  Auto-loaded when Claude considers using the cursor agent.
---

# /cursor

ユーザーの引数をタスクプロンプトとして、cursorサブエージェントを起動する。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. cursorエージェント（subagent_type: cursor）を起動
3. 結果を日本語で提示

## Reference Templates

詳細テンプレートは以下を参照：

- `references/implementation-plan.md` — 実装計画プロンプトテンプレート（plan mode）
- `references/architecture-analysis.md` — アーキテクチャ分析ワークフロー（plan mode）
- `references/design-comparison.md` — 設計比較テンプレート（plan/ask mode）
- `references/troubleshooting.md` — インストール、認証、エラー解決
