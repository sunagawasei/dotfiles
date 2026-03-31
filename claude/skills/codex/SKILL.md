---
name: codex-knowledge
user-invocable: false
description: |
  Background knowledge for Codex CLI delegation. Provides delegation patterns and context
  for orchestrating Codex analysis via the codex agent.
  Auto-loaded when Claude considers using the codex agent.
---

# Codex 委譲ナレッジ

## 委譲判断フローチャート

```
タスク受信
    │
    ▼
┌─────────────────────────┐
│ 明示的な Codex 指示？    │
└───────────┬─────────────┘
    ┌───────┴───────┐
    │ Yes          │ No
    ▼              ▼
  委譲        ┌─────────────────────────┐
              │ 複雑度チェック           │
              └───────────┬─────────────┘
              ┌───────────┴───────────┐
              │ Yes                   │ No
              ▼                       ▼
            委譲              ┌─────────────────────────┐
                              │ 失敗チェック（2回以上）  │
                              └───────────┬─────────────┘
                              ┌───────────┴───────────┐
                              │ Yes                   │ No
                              ▼                       ▼
                            委譲              ┌─────────────────────────┐
                                              │ 品質・セキュリティ要件  │
                                              └───────────┬─────────────┘
                                              ┌───────────┴───────────┐
                                              │ Yes                   │ No
                                              ▼                       ▼
                                            委譲              Claude Code で実行
```

## 使い分けガイド（3エージェント）

| 状況 | 推奨ツール |
|------|----------|
| クイックレビュー・エラー診断 | copilot agent |
| 実装計画・設計比較・依存関係整理 | cursor agent（構造化計画出力） |
| 深いアーキテクチャ分析 | codex agent |
| セキュリティ監査 | codex agent |
| 2回以上失敗したバグ調査 | codex agent |
| パフォーマンス最適化 | codex agent |

## 委譲しないケース

| ケース | 理由 |
|--------|------|
| 単純な CRUD 操作 | 定型作業，深い分析不要 |
| 小規模なバグ修正（初回） | まず Claude Code で試行 |
| ドキュメント更新のみ | 創造性より正確性重視 |
| フォーマット・リント修正 | 機械的な処理 |
| 実装計画が必要 | cursor agent に委譲 |

## Reference Templates

詳細テンプレートは以下を参照：

- `references/code-review-task.md` — Review prompt templates and `codex exec review` usage
- `references/refactoring-task.md` — Refactoring proposal workflow
- `references/delegation-patterns.md` — Decision flowchart with example commands by use case
- `references/agent-prompts.md` — Specialized templates (Architect / Analyzer / Optimizer / Security)
- `references/troubleshooting.md` — Install, auth, config, and error fixes
