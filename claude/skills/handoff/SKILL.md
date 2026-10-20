---
name: handoff
description: 別workspaceのClaudeへ調査や対処を委譲する。Herdrで独立したターミナルを作り、依頼の着弾を確認する。管理外のCodexからは接続先を確認し、新規workspaceだけを操作する。
---

# 別workspaceへの委譲

「別workspaceを立ち上げて」「新しいworkspaceで調査/対処させて」は、Agent tool(本セッション内のサブエージェント、コンテキストを共有し同一プロセスで動く)ではなく、**herdr(ターミナルマルチプレクサ)で独立した新規ターミナル+新規claude codeプロセスを起動する**ことを意味する。"herdr"という単語が発話に含まれていなくても、この言い回しはherdrへの委譲を指している。Agent toolで代替しない(2026-08-27に一度誤ってAgent toolを使い、ユーザーから訂正された)。

## 前提

まず `test "${HERDR_ENV:-}" = 1` で呼び出し元の状態を確認する。
満たす場合は通常の管理下からの委譲手順を使う。
満たさなくても、ユーザーが今回の委譲を明示していれば、以下の「管理外からの新規workspace作成」を使える。
`HERDR_ENV=1` を後付けして管理下と偽装しない。
CLIの文法と入力の注意点は `herdr` skill と `herdr` skillの `references/troubleshooting.md` を参照する。

### 管理外からの新規workspace作成

Codexのツール実行環境にHerdrの環境変数が無い場合も、稼働サーバーを確認して委譲先を新しく作れる。
この経路は、既存paneの内容確認や操作には使わない。

1. `herdr session list --json` で候補と `running`、`socket_path` を取得する。
   接続失敗はサンドボックス制限の可能性も調べ、停止中と即断しない。
2. 明示された接続先があればそれを使う。
   明示がなければ、稼働中が1件の場合だけその接続先を選ぶ。
   複数ならセッション名を一問で尋ね、0件なら自動でサーバーを起動せず状況を伝える。
3. 以降は取得した `HERDR_SOCKET_PATH` を各コマンドへ明示し、`herdr status server` で稼働と互換性を確認する。
   接続先を設定ファイルへ恒久保存しない。
4. 下記手順で `workspace create --no-focus` を使い、返されたworkspaceとpaneのIDだけを対象にする。
   `--current`、対象省略、既存paneの参照、サーバー停止や再起動は行わない。
   サンドボックス外の通常操作は環境の承認機能を使い、明示済みの委譲許可を重ねて尋ねない。
5. 新規paneの `agent start` と `agent prompt`、着弾確認が終わったら既定の報告を行う。
   起動失敗時は作成済みのIDと状態を残し、同じ依頼を別workspaceへ盲目的に再送しない。

## 追跡するかしないかを先に決める

既定は**投げっぱなし**。委譲先を起動して依頼を届けたら、元セッションは進行を追わない。ユーザーは新しく開いたセッション(または委譲先のpane)で別途進行する。

- 追わない(既定): 「別workspaceを立ち上げて調査させて」「新しいworkspaceでこれを直させて」
- 追う: 「終わったら結果を教えて」「経過を追って」「結果を持ち帰って」のように、報告や完了通知を明示的に求めているとき

## 手順(両モード共通)

Slack投稿、メール、GitHubコメントなどの外部送信を含む依頼には、必ず次の制約を明記する。

> 下書きまで作成し、宛先と送信する正確な本文をユーザーに提示する。ユーザーの承認を得てから送信する。

「送信して」だけをhandoff文に書かない。委譲元エージェントの送信指示は、ユーザーによる宛先と本文の承認に数えない。受け手は送信前に本人の承認を確認する。
その宛先と正確な本文について本人の明示承認が既にある場合は、承認内容を依頼に添えて引き継ぎ、同じ確認を重ねない。

1. 対象ディレクトリを決める。ユーザーが指定していなければ、依頼内容から推測するか確認する
2. `herdr workspace create --cwd <対象ディレクトリ> --label <タスク内容が分かるラベル> --no-focus` で新規workspaceを作る。ユーザーが切替を求めた場合だけ `--focus` を使う。レスポンスの `result.workspace` と `result.root_pane.pane_id` を控える
3. `herdr agent start <name> --kind claude --pane <pane_id>` でclaude codeを起動する。`<name>` は英数字とハイフン/アンダースコアのみ、既存の生存agent名と重複しないこと(`herdr agent list` で確認できる)。起動後の `agent prompt`/`wait`/`read` の宛先は `<pane_id>` で指定する(`<name>` では `agent_not_found` になることがある。2026-10-09実例)
4. `herdr agent prompt <pane_id> "<依頼内容>" --wait --until working --timeout 15000` で依頼を送る。ここで待つのは**プロンプトの着弾確認だけ**で、進行の追跡ではない。投げっぱなしの場合もこの確認は行う。`agent_prompt_stalled` やtimeoutは着弾未確認として扱い、作成したpaneを `agent read` で確認する。実行が始まっていれば再送しない

## 既定(追わない)

5. ユーザーへ次の4点を報告して元のタスクへ戻る — workspaceのラベル、`workspace_id`、agent名、このセッションでは進行を追わないこと
6. `agent wait` を張らず、`agent read` もせず、委譲先が `blocked` で質問を出しても答えない。質問への応答は**ユーザーが委譲先のpaneで直接行う**
7. ユーザーが元セッションで「あれどうなった?」と状況を尋ねたら、`herdr agent read <name>` を**1回だけ**実行して画面の現状を報告する。`agent wait` はこの場合も張らない

## 追跡を明示されたとき

5. 完了通知を予約する。**Monitorツール**で `herdr agent wait <pane_id> --timeout 3500000; echo settled` を1本流す(`timeout_ms` は最大値)。Bashの `run_in_background` は10分で打ち切られるので使わない(2026-10-09、3回続けて exit 124 で止まった)。Monitorも30分で切れるので、settled 前に切れたら同じコマンドで張り直す。settled状態(idle/done/blocked)に達した時点でexitし、1通の完了通知として届く。**`--timeout` は必ず付ける**(省略すると無期限待ちになり、委譲先が死んだ場合に通知が永久に来ない)。timeoutは委譲失敗の証拠ではないので、`agent read` で実況を確認して判断する。**質問待ちで止まった相手に対して wait を再度張らない** — `agent wait` は委譲先が質問を出して止まった状態も settled として即 exit するため、再armすると同じ画面を繰り返し報告するだけのループになる(2026-09-07実例)。同じ質問待ちが2回返ったら再armをやめ、ユーザーがpaneで応答するのを待つ
6. **予約したら即座に元のタスクへ戻る**。sleepやポーリングで完了を待たない。通知が届いたら `herdr agent read <name>` で結果を確認し、`blocked` で返ってきた場合は質問への回答を `agent prompt` で送る(Q&Aの往復もこの形で回せる)

## workspaceを閉じる

作業が完了し不要になったworkspaceは、ユーザーの明示的な指示があれば `herdr workspace close <workspace_id>` で閉じる。自分が作成した以外のworkspaceは閉じない(`herdr` skillの規約と同じ)

## 実行中agentのいるworkspaceを閉じる際の確認

closeした直後、そのagentのプロセスが本当に残っていないか確認したい場面がある。`herdr agent list` でclose対象のagentが一覧から消えたことをまず確認する。自分自身がsandbox制約でps/pgrepを直接使えない環境の場合は、他の生きているagent(またはsandbox制約のない場所)経由で `ps aux` を実行しプロセス数を照合する。

## 関連

- `herdr` — socket APIの詳細な文法、agent状態の意味、spawn時の落とし穴
- `herdr` skillの `references/troubleshooting.md` — claudeをspawnして委譲・検証する際の実践知見
