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

- **canonical form を必ず使うこと**（hang防止 + 安定実行に必要なフラグをすべて含む）:
  ```bash
  codex exec \
    --sandbox read-only \
    -c approval_policy="never" \
    --skip-git-repo-check \
    --color never \
    --cd "$(pwd)" \
    "<prompt>" \
    < /dev/null
  ```
- `< /dev/null` は**必須**（省略するとstdinブロックでhangする）
- 引数なしの `codex` や `exec` なしの `codex "..."` は**絶対に禁止**（TUI起動でstdinハング）
- 引数（プロンプト）が空の場合はcodexを起動せずユーザーに確認を求めること
- **Bashツールでcodex execを実行する際は必ず `timeout: 600000`（10分）を指定すること**（デフォルト120秒ではタイムアウトする）
- **`run_in_background: true` は絶対に使用しないこと**（バックグラウンド実行するとセッション終了時にkillされる）

## 手順

1. 引数をタスクプロンプトとして受け取る（空の場合はエラー）
2. codexエージェント（subagent_type: codex）を起動
3. 結果を日本語で提示
