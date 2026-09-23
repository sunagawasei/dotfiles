---
name: cursor-delegate
description: 隣paneでcursor agentを起動して実装を任せ、claude(メイン)は現状確認と調査に徹する。「隣でcursorに実装させて」「cursorに投げて、claudeは見てるだけでいい」で使う
---

# cursorへの実装委譲

隣のpaneでcursor agentを起動し、指定したタスクを自律実装させる。claude(メイン)側はそのタスクのファイルセットを自分では編集せず、状況確認・調査・質問への回答に徹する。

## 前提

`herdr`スキルをロードしてherdr CLIの基本(pane/agentの区別、IDの扱い)を把握していること。`HERDR_ENV=1`で動いていること(`herdr`スキルの前提チェックに従う)。

## 起動手順

1. `herdr pane current --current`で自分(claude)のworkspace/tab/pane IDを確認する
2. `herdr agent list`で、同じtab内に既にcursorのagentが動いていないか確認する
   - 動いていれば、そのagent名(またはpane id)を使い、手順3(pane起動)は飛ばして手順4へ進む
   - 動いていなければ手順3へ
3. 隣にpaneを作る: `herdr pane split --current --direction right --cwd "$PWD" --no-focus`。レスポンスの`.result.pane.pane_id`を控える
4. そのpaneでcursor agentを起動する: `herdr agent start <agent名> --kind cursor --pane <pane_id>`(既存のagent名と衝突しない名前にする。同一tabで複数cursorを走らせる場合は用途がわかる名前にする)
5. タスクのgoalを送る(`--wait`は付けない。`/goal`は長時間走り続けるため、待つとタイムアウトするか実質固まる):

   ```
   herdr agent prompt <agent名> "/goal <ユーザーから渡されたタスク内容>。完了条件はPR作成までとし、mainへのmergeは行わない。自分が開いたPRがmergeされたことに気づいたら、そのローカル・リモート両方の作業ブランチを削除する。"
   ```

   タスク内容・完了条件はユーザーの依頼に合わせて都度組み立てる。上記の「PR作成まで・mergeしない」は安全側のデフォルトで、ユーザーが別の完了条件を指定したらそちらに従う。

6. 数秒後に`herdr agent get <agent名>`と`herdr agent read <agent名> --source recent-unwrapped --lines 40`で、goalが実際にactiveになった(「Goal active」表示が出た)ことを確認する

定期的な自動確認(ScheduleWakeupによる30分おきの報告等)は既定で仕込まない。進捗はユーザーに聞かれたとき、または後述の手順で都度確認する。

## 進捗を尋ねられたら

1. `herdr pane list --workspace <対象workspace>` / `herdr agent list`で対象paneのagent種別・cwd・ブランチを確認する
2. `herdr pane read <pane_id> --source recent-unwrapped --lines 60〜150`で直近の作業内容・To-do・やり取りを読む
3. 他エージェントとの協業状況を詳しく見るときは、そのworkspaceのagmsg team(paneログに`Sent to codex in team s-...`のように出る)を確認する
   - `bash ~/.agents/skills/agmsg/scripts/team.sh <team>` — ロースター
   - `bash ~/.agents/skills/agmsg/scripts/history.sh <team>` — メッセージ履歴(出力が大きいので絞り込む)
4. 要約してユーザーに報告する(生のログを転載しない)

### 質問・指示を取り次ぐ場合

`herdr agent prompt <pane_id> "<message>" --wait --timeout <ms>`で送る。対象がWorking中だと、メッセージは画面下の「follow-ups」欄に`○`(未送信)のまま溜まり、`--wait`はタイムアウトしやすい。その場合:

1. `herdr pane read <pane_id>`でfollow-upsキューに`○`のメッセージが見えるか確認する
2. `herdr agent send-keys <pane_id> enter`でキューのメッセージを即座に処理へ回す(steer)
3. `herdr agent wait <pane_id> --until idle --timeout <ms>`で完了を待つ

### 同じリポジトリを見ているときは worktree の共有を確認する

対象paneの`cwd`が自分と同じリポジトリだったら、`git worktree list`で**同一worktreeを共有していないか**を確かめる。共有していると、**相手がブランチを切り替えた結果として自分の`git status`のブランチが勝手に変わる**。

- 確認: `git worktree list`が1行しか返らず、かつ`herdr pane list`の`cwd`が自分と同じなら共有している
- 共有していたら、こちらからのブランチ操作・commit・stashは控える(相手の作業中の状態を壊す)
- セッション開始時のgit statusは当てにならない。ブランチ名を使う前に`git branch --show-current`を採り直す

## 「止めて」「閉じて」と言われたら

「止める」と「閉じる」は別の操作。

- **止める(goalの実行を止める。pane・agentは残る)**: `herdr agent send-keys <agent名> C-c`。実行後`herdr agent get <agent名>`で`agent_status`が`done`になったことを確認する
- **閉じる(pane自体を畳む。以後`herdr agent list`から消える)**: `herdr pane close <pane_id>`。再開するには起動手順3(pane split)からやり直す

## メイン(claude)自身がこのrepoでファイル編集・commitする必要が生じたら

このスキルの前提は「cursorが実装し、claudeは調査・確認に徹する」こと。委譲したタスクの範囲でclaude自身が編集する必要が出たら、まずcursor側へ`herdr agent prompt`で伝えて任せられないか検討する。

どうしてもclaude側で書く必要がある場合、cursorと同じworktreeを共有していると、そのままeditすると相手が実行中の`git add`/commit/branch切替に巻き込まれる。`git worktree add ../<repo名>-<用途> -b <branch> origin/main`で自分専用の隔離worktreeを作り、そちらで作業する。マージ・デプロイがGitHub側/CI側で完結する設計であれば、作業完了後に`git worktree remove`するだけで元のworktreeに戻す必要すら無い。
