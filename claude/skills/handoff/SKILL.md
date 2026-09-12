---
name: handoff
description: ユーザーが「別workspaceを立ち上げて調査/対処させて」「新しいworkspaceでやらせて」のように言ったときに使う。Agent tool(本セッション内のサブエージェント)ではなく、herdr経由で独立した新規ターミナル+新規claude codeプロセスを起動して作業を委譲する。Requires HERDR_ENV=1.
---

# 別workspaceへの委譲

「別workspaceを立ち上げて」「新しいworkspaceで調査/対処させて」は、Agent tool(本セッション内のサブエージェント、コンテキストを共有し同一プロセスで動く)ではなく、**herdr(ターミナルマルチプレクサ)で独立した新規ターミナル+新規claude codeプロセスを起動する**ことを意味する。"herdr"という単語が発話に含まれていなくても、この言い回しはherdrへの委譲を指している。Agent toolで代替しない(2026-08-27に一度誤ってAgent toolを使い、ユーザーから訂正された)。

## 前提

`test "${HERDR_ENV:-}" = 1` を確認する。満たさなければherdr管理下のpaneで動いていないので、その旨を伝えて停止する。CLIの正確な文法・落とし穴(長文プロンプトのペースト扱い・完了検知等)は `herdr` skill と `herdr` skillの `references/troubleshooting.md` が正本。このskillはそこへの「トリガーの翻訳」と典型手順のショートカットに留める。

## 追跡するかしないかを先に決める

既定は**投げっぱなし**。委譲先を起動して依頼を届けたら、元セッションは進行を追わない。ユーザーは新しく開いたセッション(または委譲先のpane)で別途進行する。

- 追わない(既定): 「別workspaceを立ち上げて調査させて」「新しいworkspaceでこれを直させて」
- 追う: 「終わったら結果を教えて」「経過を追って」「結果を持ち帰って」のように、報告や完了通知を明示的に求めているとき

## 手順(両モード共通)

1. 対象ディレクトリを決める。ユーザーが指定していなければ、依頼内容から推測するか確認する
2. `herdr workspace create --cwd <対象ディレクトリ> --label <タスク内容が分かるラベル>` で新規workspaceを作る。レスポンスの `result.root_pane.pane_id` を控える
3. `herdr agent start <name> --kind claude --pane <pane_id>` でclaude codeを起動する。`<name>` は英数字とハイフン/アンダースコアのみ、既存の生存agent名と重複しないこと(`herdr agent list` で確認できる)
4. `herdr agent prompt <name> "<依頼内容>" --wait --until working --timeout 15000` で依頼を送る。ここで待つのは**プロンプトの着弾確認だけ**で、進行の追跡ではない。投げっぱなしの場合もこの確認は行う。`agent_prompt_stalled` やtimeoutが返ったらプロンプトが届いていないので、`agent read` で状況を見て再送する

## 既定(追わない)

5. ユーザーへ次の4点を報告して元のタスクへ戻る — workspaceのラベル、`workspace_id`、agent名、このセッションでは進行を追わないこと
6. `agent wait` を張らず、`agent read` もせず、委譲先が `blocked` で質問を出しても答えない。質問への応答は**ユーザーが委譲先のpaneで直接行う**
7. ユーザーが元セッションで「あれどうなった?」と状況を尋ねたら、`herdr agent read <name>` を**1回だけ**実行して画面の現状を報告する。`agent wait` はこの場合も張らない

## 追跡を明示されたとき

5. 完了通知を予約する。Bashツールの `run_in_background` で `herdr agent wait <name> --timeout 1800000` を1本流す。settled状態(idle/done/blocked)に達した時点でexitし、1通の完了通知として届く。**`--timeout` は必ず付ける**(省略すると無期限待ちになり、委譲先が死んだ場合に通知が永久に来ない)。timeoutは委譲失敗の証拠ではないので、`agent read` で実況を確認して判断する。**質問待ちで止まった相手に対して wait を再度張らない** — `agent wait` は委譲先が質問を出して止まった状態も settled として即 exit するため、再armすると同じ画面を繰り返し報告するだけのループになる(2026-09-07実例)。同じ質問待ちが2回返ったら再armをやめ、ユーザーがpaneで応答するのを待つ
6. **予約したら即座に元のタスクへ戻る**。sleepやポーリングで完了を待たない。通知が届いたら `herdr agent read <name>` で結果を確認し、`blocked` で返ってきた場合は質問への回答を `agent prompt` で送る(Q&Aの往復もこの形で回せる)

## workspaceを閉じる

作業が完了し不要になったworkspaceは、ユーザーの明示的な指示があれば `herdr workspace close <workspace_id>` で閉じる。自分が作成した以外のworkspaceは閉じない(`herdr` skillの規約と同じ)

## 実行中agentのいるworkspaceを閉じる際の確認

closeした直後、そのagentのプロセスが本当に残っていないか確認したい場面がある。`herdr agent list` でclose対象のagentが一覧から消えたことをまず確認する。自分自身がsandbox制約でps/pgrepを直接使えない環境の場合は、他の生きているagent(またはsandbox制約のない場所)経由で `ps aux` を実行しプロセス数を照合する。

## 関連

- `herdr` — socket APIの詳細な文法、agent状態の意味、spawn時の落とし穴
- `herdr` skillの `references/troubleshooting.md` — claudeをspawnして委譲・検証する際の実践知見
