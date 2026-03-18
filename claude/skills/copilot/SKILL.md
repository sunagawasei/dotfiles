---
name: copilot-knowledge
user-invocable: false
description: |
  Background knowledge for Copilot CLI delegation. Provides context for orchestrating
  quick code reviews and analysis via the copilot agent.
  Auto-loaded when Claude considers using the copilot agent.
---

# Copilot 委譲ナレッジ

## 使い分けガイド

| 状況 | 推奨ツール |
|------|----------|
| さっとレビューしてほしい | copilot agent（軽量・高速） |
| エラーの原因を診断したい | copilot agent |
| テストを提案してほしい | copilot agent |
| ドキュメントのレビュー | copilot agent |
| 深いアーキテクチャ分析 | codex agent |
| セキュリティ監査 | codex agent |
| 2回以上失敗したバグ調査 | codex agent |

## Reference Templates

詳細テンプレートは以下を参照：

- `references/code-review.md` — コードレビュー用プロンプトテンプレート
- `references/error-analysis.md` — エラー診断ワークフロー
- `references/test-generation.md` — テスト提案テンプレート
- `references/documentation.md` — ドキュメントレビューテンプレート
