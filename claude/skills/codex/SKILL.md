---
name: codex
user-invocable: true
description: |
  OpenAI Codex CLIでコード分析を実行する。
  Usage: /codex "プロンプト"
  Auto-loaded when Claude considers using the codex agent.
---

# /codex

ユーザーの引数をタスクプロンプトとして、codexサブエージェントを起動する。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. codexエージェント（subagent_type: codex）を起動
3. 結果を日本語で提示

## Reference Templates

詳細テンプレートは以下を参照：

- `references/code-review-task.md` — Review prompt templates and `codex exec review` usage
- `references/refactoring-task.md` — Refactoring proposal workflow
- `references/delegation-patterns.md` — Example commands by use case
- `references/agent-prompts.md` — Specialized templates (Architect / Analyzer / Optimizer / Security)
- `references/troubleshooting.md` — Install, auth, config, and error fixes
