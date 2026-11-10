---
name: commit
user-invocable: true
description: |
  コミットして、commit、変更を記録して、save my changes。
  Use when the user says "コミットして", "commit", "変更を記録して", "save my changes",
  or when explicitly asked to create a commit.
---

# /commit

メインセッションが直接コミットする(専用のsubagentは使わない)。

## 手順

1. `git status` と `git diff --stat` を確認し、**当該タスクに関係しないファイルが変更済みになっていないか**チェックする。
   - 関係しないファイルがある場合はユーザーに確認してからコミットする
   - とくに、セッション開始時から既に `M` だったファイルはタスクのスコープ外の可能性が高い

2. 自分が変更したファイルだけをpathspecで指定してコミットする(`git commit <path>...`。無関係な未コミット変更を含めない)
3. 結果(コミットSHA・push有無)を日本語で提示

## commit messageの書き方

- 言語は `git log --oneline -20` の既存スタイルに合わせる。英語で書く場合はレビュー用に日本語訳を併記する(commitされる本文は英語のまま)
- **リポジトリの履歴に一度も入っていない状態からの差分を書かない**。「XではなくY」「もうXに依存しない」型の記述は、Xがgitに存在しなければ読む人には何も指さない。現在の状態とその理由だけを書く
- 内部の決定番号(D1、F3等)・別文書の節番号を書かない。リポジトリの外を指す識別子は読む人に伝わらない
- **messageを提示して承認を得てから実行する**。ユーザーの「commit」「go」等の短い指示は、直前にmessageを提示済みの場合にのみ承認として扱う
