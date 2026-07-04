---
name: commit
user-invocable: true
description: |
  コミットして、commit、変更を記録して、save my changes。
  Use when the user says "コミットして", "commit", "変更を記録して", "save my changes",
  or when explicitly asked to create a commit.
---

# /commit

Agent toolを使い、commitサブエージェントを起動してコミットを実行する。

## 手順

1. `git status` と `git diff --stat` を確認し、**当該タスクに関係しないファイルが変更済みになっていないか**チェックする。
   - 関係しないファイルがある場合はユーザーに確認してから commitエージェントを起動する
   - とくに、セッション開始時から既に `M` だったファイルはタスクのスコープ外の可能性が高い

2. Agent toolで commitエージェントを起動する:
   - `subagent_type: commit`
   - `prompt`: ステージングすべきファイルを明示して渡す。なければ「現在の変更をコミットしてください」
3. エージェントの結果を日本語で提示
