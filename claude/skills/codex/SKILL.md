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

## 重要な制約

- codexサブエージェント内では必ず `codex exec --sandbox read-only "プロンプト"` 形式を使用すること
- 引数なしの `codex` や `codex "プロンプト"` （exec なし）は**絶対に禁止**（インタラクティブモードになりstdinハングする）
- 引数（プロンプト）が空の場合はcodexを起動せずユーザーに確認を求めること
- **Bashツールでcodex execを実行する際は必ず `timeout: 600000`（10分）を指定すること**（デフォルト120秒ではタイムアウトする）
- **`run_in_background: true` は絶対に使用しないこと**（バックグラウンド実行するとセッション終了時にkillされる）

## 手順

1. 引数をタスクプロンプトとして受け取る（空の場合はエラー）
2. codexエージェント（subagent_type: codex）を起動
3. 結果を日本語で提示

## Reference Templates

詳細テンプレートは以下を参照：

- `references/code-review-task.md` — Review prompt templates and `codex exec review` usage
- `references/refactoring-task.md` — Refactoring proposal workflow
- `references/delegation-patterns.md` — Example commands by use case
- `references/agent-prompts.md` — Specialized templates (Architect / Analyzer / Optimizer / Security)
- `references/troubleshooting.md` — Install, auth, config, and error fixes
