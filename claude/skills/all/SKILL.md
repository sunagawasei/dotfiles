---
name: all
user-invocable: true
description: |
  codex・copilot・cursorの3エージェントを並列起動してコード分析を実行する。
  Usage: /all "プロンプト"
---

# /all

ユーザーの引数をタスクプロンプトとして、3エージェントを並列起動する。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 3つのエージェントを**同一メッセージ内で並列に**起動:
   - codexエージェント（subagent_type: codex）
   - copilotエージェント（subagent_type: copilot）
   - cursorエージェント（subagent_type: cursor）
3. 3つの結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
