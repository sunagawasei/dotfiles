---
name: codex-copilot
user-invocable: true
description: |
  codexとcopilotを並列起動してコード分析を実行する。
  Usage: /codex-copilot "プロンプト"
---

# /codex-copilot

ユーザーの引数をタスクプロンプトとして、codexとcopilotの2エージェントを並列起動する。

## 重要な制約

- **全エージェント共通:**
  - `run_in_background: true` は絶対に使用しない（セッション終了時やコンテキスト進行時にkillされる原因になる）
  - すべてのエージェントがフォアグラウンドで完了するまで待機し、結果を統合すること
- **codexエージェント固有:**
  - `codex exec` 実行時は必ず `timeout: 600000` を指定する
  - `< /dev/null` を末尾に付けること（stdinハング防止）
  - `-c approval_policy="never"` を付けること（対話承認待ちハング防止）
  - 詳細な canonical form は `claude/agents/codex.md` §Execution を参照

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 2つのエージェントを**同一メッセージ内で並列に**起動:
   - codexエージェント（subagent_type: codex）
   - copilotエージェント（subagent_type: copilot）
3. 両方の結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
