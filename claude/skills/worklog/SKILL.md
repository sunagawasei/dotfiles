---
name: worklog
description: "worklog CLIでタスクごとの稼働時間を計測する。`/worklog start [作業内容]` で開始、`/worklog stop [HH:MM]` で停止。「タスク宣言して」「ws1にタスク宣言」「作業停止を報告」でも使う。過去の同一・類似タスクがあれば名前とタグを再利用する。"
argument-hint: "[start|stop] [作業内容 | HH:MM]"
---

# worklog start / stop

このセッションのエージェントが`worklog` CLIを直接実行して稼働時間を記録する。常に`--json`を付ける(TTYが無い環境で対話プロンプトが開けないため)。CLIの仕様全般は`~/poc/worklog/.claude/skills/worklog-operation/SKILL.md`が正本。

`worklog new`/`switch`は稼働中の別タスクを自動停止する。宣言の報告に停止された旧タスク名(`prev_project`)を含める。

## start [作業内容]

作業内容を省略したときは、依頼文から決める。

1. 稼働状況: `worklog status --json`
2. 開始時刻(HH:MM): 依頼を受けた時刻をセッションjsonlから取る。同じproject dirで別セッションが動いていると最新ファイルが自分のものとは限らないので、依頼の文面が出るかで確認する。

   ```bash
   cd ~/.config/claude/projects/<project-dir> && f=$(ls -t *.jsonl | head -1)
   jq -r 'select(.type=="user" and (.isMeta|not) and (.isSidechain|not)) | select(.message.content|type=="string" or (.[0].type!="tool_result")) | "\(.timestamp|sub("\\.[0-9]+Z$";"Z")|fromdateiso8601+32400|strftime("%Y-%m-%d %H:%M JST")) \((.message.content|if type=="string" then . else map(select(.type=="text").text)|join(" ") end)|gsub("\\s+";" ")|.[0:50])"' "$f" | tail -10
   ```

   `<project-dir>`は作業ディレクトリのgit rootの`/`を`-`に置換した名前(`/Users/s23159/.config` → `-Users-s23159--config`)。複数repoで1つのproject dirを共有している場合(例: containerregistry)はこの名前にならない。dirが見つからなければ `grep -l '<依頼文の一部>' ~/.config/claude/projects/*/*.jsonl` でセッションファイルを特定する。
3. 過去タスクの検索: 作業内容から識別性の高い語(issue番号`#219`、チケットID、固有名詞)を選び、語ごとに実行して結果をマージする。検索対象はログのファイル名(プロジェクト名)の部分一致だけ。

   ```bash
   ~/poc/worklog/.claude/scripts/find-past-project.sh "<keyword>"
   ```

   出力は `日付<TAB>プロジェクト名<TAB>最終タグID`(新しい順)。
4. 名前とタグを決める:

   | 検索結果 | 扱い |
   |---|---|
   | 同一タスクと言える(issue番号・チケットIDが一致、または名前が一致) | 名前と最終タグIDをそのまま使う。確認しない |
   | 類似のみ | 上位3件と「新規」を`AskUserQuestion`で見せて選ばせる |
   | 0件 | 新規。下の「新規の命名」へ |

5. 開始: `worklog new --json "<プロジェクト名>" <タグID> <HH:MM>`
6. `worklog status --json` で稼働中のタスクが宣言した内容になったか確認し、プロジェクト名・タグ名・開始時刻・`prev_project`を報告する。

### 新規の命名

- GitHub issue起点: `gh issue view <番号> --repo <owner>/<repo> --json title`で実タイトルを取り、`#<番号> <タイトル>`にする。URLスラッグはタイトルと一致しないので名前に使わない
- タグ: `worklog tag list --json`で一覧を取る。同じリポジトリの直近タスクのタグを候補にして、プロジェクト名と合わせて`AskUserQuestion`で確認する。推測で確定しない
- 名前は1〜255文字。`/` `\` `*` `?` `"` `<` `>` `|` `..` と制御文字は使えない

## stop [HH:MM]

ユーザーが停止を明示したときだけ実行する。「セッションを終える」「wrapup」「/exit」は作業の停止を意味しないので、それだけでは止めない。

`worklog stop --json [HH:MM]`。作業を終えた時刻が今でなければ`HH:MM`を渡す。画像から読んだ時刻は実行前に`AskUserQuestion`で確認する。

## ログの直し方

CLIに編集コマンドは無い。ログファイルを直接書き換える。

- 場所: `~/.local/state/worklog/YYYY/MM/DD/<プロジェクト名>.log`。1行1件で `YYYY/MM/DD HH:MM:SS<TAB>start|stop<TAB>タグID`
- 稼働中のタスクは同じ日のディレクトリの `current`(`<プロジェクト名><TAB>タグID`)が指す。開始時刻はログの最後の `start` 行から取られるので、`current` は名前を変えるときだけ直す
- 書き換える前にファイルをscratchpadへコピーし、書き換えた後に `worklog timeline --json` で区間を確認する

| 直したいこと | 書き換え |
|---|---|
| 誤って止めて再開した(区間が途切れた・重なった) | 誤った `stop` 行と再開の `start` 行を消す |
| 開始・停止の時刻がずれた | その行の時刻を書き換える |
| 名前を誤った | ログファイルをリネームする。稼働中なら `current` の名前も書き換える |
