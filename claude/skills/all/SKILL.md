---
name: all
user-invocable: true
description: |
  codex・copilot・cursorの3エージェントを並列起動してコード分析を実行する。
  Usage: /all "プロンプト"
---

# /all

ユーザーの引数をタスクプロンプトとして、3エージェントを並列起動する。

## 重要な制約

- **全エージェント共通:**
  - `run_in_background: true` は絶対に使用しない（セッション終了時やコンテキスト進行時にkillされる原因になる）
  - 3つのエージェントすべてがフォアグラウンドで完了するまで待機し、結果を統合すること
- **codexエージェント固有:**
  - `codex exec` 実行時は必ず `timeout: 600000` を指定する

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 3つのエージェントを**同一メッセージ内で並列に**起動:
   - codexエージェント（subagent_type: codex）
   - copilotエージェント（subagent_type: copilot）
   - cursorエージェント（subagent_type: cursor）
3. 3つの結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
