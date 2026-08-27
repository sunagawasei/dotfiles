---
name: workspace-delegate
description: ユーザーが「別workspaceを立ち上げて調査/対処させて」「新しいworkspaceでやらせて」のように言ったときに使う。Agent tool(本セッション内のサブエージェント)ではなく、herdr経由で独立した新規ターミナル+新規claude codeプロセスを起動して作業を委譲する。Requires HERDR_ENV=1.
---

# 別workspaceへの委譲

「別workspaceを立ち上げて」「新しいworkspaceで調査/対処させて」は、Agent tool(本セッション内のサブエージェント、コンテキストを共有し同一プロセスで動く)ではなく、**herdr(ターミナルマルチプレクサ)で独立した新規ターミナル+新規claude codeプロセスを起動する**ことを意味する。"herdr"という単語が発話に含まれていなくても、この言い回しはherdrへの委譲を指している。Agent toolで代替しない(2026-08-27に一度誤ってAgent toolを使い、ユーザーから訂正された)。

## 前提

`test "${HERDR_ENV:-}" = 1` を確認する。満たさなければherdr管理下のpaneで動いていないので、その旨を伝えて停止する。CLIの正確な文法・落とし穴(長文プロンプトのペースト扱い・完了検知等)は `herdr` skill と `herdr` skillの `references/troubleshooting.md` が正本。このskillはそこへの「トリガーの翻訳」と典型手順のショートカットに留める。

## 手順

1. 対象ディレクトリを決める。ユーザーが指定していなければ、依頼内容から推測するか確認する
2. `herdr workspace create --cwd <対象ディレクトリ> --label <タスク内容が分かるラベル>` で新規workspaceを作る。レスポンスの `result.root_pane.pane_id` を控える
3. `herdr agent start <name> --kind claude --pane <pane_id>` でclaude codeを起動する。`<name>` は英数字とハイフン/アンダースコアのみ、既存の生存agent名と重複しないこと(`herdr agent list` で確認できる)
4. `herdr agent prompt <name> "<依頼内容>" --wait --timeout <ms>` で依頼を送る。重い調査は120000〜300000ms程度を目安にする
5. `herdr agent read <name>` で結果を確認する。追加の指示が必要なら再度 `agent prompt` で送る(Q&Aの往復もこの形で回せる)
6. 作業が完了し不要になったworkspaceは、ユーザーの明示的な指示があれば `herdr workspace close <workspace_id>` で閉じる。自分が作成した以外のworkspaceは閉じない(`herdr` skillの規約と同じ)

## 実行中agentのいるworkspaceを閉じる際の確認

closeした直後、そのagentのプロセスが本当に残っていないか確認したい場面がある。`herdr agent list` でclose対象のagentが一覧から消えたことをまず確認する。自分自身がsandbox制約でps/pgrepを直接使えない環境の場合は、他の生きているagent(またはsandbox制約のない場所)経由で `ps aux` を実行しプロセス数を照合する。

## 関連

- `herdr` — socket APIの詳細な文法、agent状態の意味、spawn時の落とし穴
- `herdr` skillの `references/troubleshooting.md` — claudeをspawnして委譲・検証する際の実践知見
