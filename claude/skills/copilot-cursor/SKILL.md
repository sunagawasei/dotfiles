---
name: copilot-cursor
user-invocable: true
description: |
  copilotとcursorを並列起動してコード分析を実行する。
  Usage: /copilot-cursor "プロンプト"
---

# /copilot-cursor

ユーザーの引数をタスクプロンプトとして、copilotとcursorの2エージェントを並列起動する。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 2つのエージェントを**同一メッセージ内で並列に**起動:
   - copilotエージェント（subagent_type: copilot）
   - cursorエージェント（subagent_type: cursor）
3. 両方の結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
