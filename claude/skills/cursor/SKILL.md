---
name: cursor-knowledge
user-invocable: false
description: |
  Background knowledge for Cursor Agent CLI delegation. Provides delegation patterns
  and context for orchestrating structured analysis via the cursor agent.
  Auto-loaded when Claude considers using the cursor agent.
---

# Cursor 委譲ナレッジ

## 委譲判断フローチャート

```
タスク受信
    │
    ▼
┌─────────────────────────┐
│ 明示的な cursor 指示？   │
└───────────┬─────────────┘
    ┌───────┴───────┐
    │ Yes          │ No
    ▼              ▼
  委譲        ┌─────────────────────────┐
              │ 構造化計画出力が必要？  │
              └───────────┬─────────────┘
              ┌───────────┴───────────┐
              │ Yes                   │ No
              ▼                       ▼
            委譲              ┌─────────────────────────┐
                              │ 複数ファイル分析？       │
                              └───────────┬─────────────┘
                              ┌───────────┴───────────┐
                              │ Yes                   │ No
                              ▼                       ▼
                            委譲              ┌─────────────────────────┐
                                              │ 設計比較・トレードオフ？ │
                                              └───────────┬─────────────┘
                                              ┌───────────┴───────────┐
                                              │ Yes                   │ No
                                              ▼                       ▼
                                            委譲              copilot / codex / Claude Code
```

## 使い分けガイド

cursor の差別化は**出力形式**：構造化されたステップ付き計画が必要なときに使う。

| 状況 | 推奨ツール |
|------|----------|
| さっとレビューしてほしい | copilot（軽量・高速） |
| エラーの原因を診断したい | copilot |
| テストを提案してほしい | copilot |
| 実装計画を立てたい | **cursor（plan mode）** |
| 2つのアプローチを比較したい | **cursor（plan mode）** |
| リファクタリング戦略を立てたい | **cursor（plan mode）** |
| モジュール間の依存関係を整理したい | **cursor（plan mode）** |
| 特定ファイルについて質問したい | **cursor（ask mode）** |
| 深いアーキテクチャ分析 | codex |
| セキュリティ監査 | codex |
| 2回以上失敗したバグ調査 | codex |
| パフォーマンス最適化分析 | codex |

## 委譲しないケース

| ケース | 理由 |
|--------|------|
| 単純なCRUD操作 | 定型作業、計画不要 |
| 単一ファイルのクイックレビュー | copilot の方が速い |
| セキュリティ重要度の高い監査 | codex の方が深い |
| 繰り返し失敗しているバグ調査 | codex に委譲すべき |

## フォールバックチェーン

- cursor が失敗した場合:
  - 軽量タスク → copilot にフォールバック
  - 深い分析タスク → codex にフォールバック

## Reference Templates

詳細テンプレートは以下を参照：

- `references/implementation-plan.md` — 実装計画プロンプトテンプレート（plan mode）
- `references/architecture-analysis.md` — アーキテクチャ分析ワークフロー（plan mode）
- `references/design-comparison.md` — 設計比較テンプレート（plan/ask mode）
- `references/troubleshooting.md` — インストール、認証、エラー解決
