---
name: parallel-handoff
description: 複数のタスクを git worktree と herdr workspace に1つずつ割り当てて並列に進め、このセッションが中央で受入れ・統合・片付けをする。「タスクを分けて別workspaceで並列に進めて」「中央で回して」「handoffして中央集権で動かして」で使う。1タスクを投げっぱなしにするなら handoff、1タスクを段1〜11で通すなら orchestrate-agents。Requires HERDR_ENV=1.
---

# 並列 handoff の中央運用

このセッションが中央になり、各タスクは専用の worktree・branch・herdr workspace で別の Claude Code セッションに任せる。中央は自分では実装せず、分割・起動・受入れ・統合・片付けを行う。各 workspace は自分の codex で査読ループを回す。

委譲フロー(段1〜11)の正本はグローバル CLAUDE.md と `orchestrate-agents` のまま。この skill は、それを複数 workspace で並列に回すときの段取りと、速度のために緩める点を定める。

## 1. 分割

- タスクに `T<番号>` と、何を直すかが分かる名前を付ける。issue 番号だけの名前にしない。
  - 良い例: `T3 #24+#47: Cursor が claude と誤登録される/再開した worker の返信が拒否される問題を、セッション判定の作り直しで直す`
- 同時に走らせるかは、変更ファイルが重なるかだけで決めない。共有 lib を通じて同じ契約(生死判定・lock など)を使うタスクは、作業は並列でよいが、受入れと統合は直列にする。
- 他タスクが作り直す仕組みの上に乗るタスク(例: T3 の生死判定を使う T4)は、前のタスクの統合後に起動する。
- 同時に走らせる workspace は CPU に合わせて 4〜6 本までにする。負荷は `uptime` で見る。

## 2. 起動

各タスクについて、中央の最新 main から worktree を切り、workspace を作り、claude を起動して依頼文を送る。

```bash
git -C <repo> worktree add -q -b <branch> ../<repo>-wt-<task> main
pane=$(herdr workspace create --cwd <worktree> --label "<repo> T<n>: <短い説明>" | grep -o '"pane_id":"[^"]*"' | head -1 | cut -d'"' -f4)
sleep 5   # 作成直後は shell の準備が間に合わず agent start が失敗する
herdr agent start <name> --kind claude --pane "$pane" --timeout 90000
herdr agent prompt "$pane" "$(cat <packet>)" --wait --until working --timeout 15000
```

- 依頼文は「タスク固有の部分」と `references/rules.md` の共通ルールを連結して作る。
- 依頼文には、調べ済みの事実(原因の file:line、upstream の SHA、関係する issue)を書く。workspace 側に同じ調査をさせない。
- 以後の指定は agent 名ではなく pane ID で行う。herdr の agent 名は自動の ID に変わることがある。

## 3. 各 workspace が回すもの

`references/rules.md` のとおり。要点は次の5つ。

- プラン → codex プラン査読(最大2巡で打ち切る)→ 自分で承認 → 実装 → codex コード査読(High が残るときだけ再依頼)。
- fable-review の振り分けは使わない。指摘の採否は workspace が決め、理由を報告に1行ずつ書く。
- テストは変更に対する新規・修正テストだけ。全テストは流さない。
- 繰り返し実行で安定性を確かめるのは最大5回まで。CPU を空回しする負荷生成は禁止。
- 段9 が収束したら、中央の pane へ `herdr agent prompt <中央 pane> "[T<n> review-done] ..."` で報告して待つ。

## 4. 中央の監視

- 各 workspace の報告は、中央のプロンプトに直接届く。それとは別に、状態変化を Monitor で拾う。

```bash
while true; do st=$(herdr agent list | grep -o '"agent_status":"[a-z]*","[^}]*"pane_id":"<pane>"' | grep -o 'status":"[a-z]*' | cut -d'"' -f3); [ -n "$st" ] && [ "$st" != working ] && { echo "<pane>=$st"; break; }; sleep 15; done
```

- 作業中の agent に送った prompt はキューに入り、`--wait` は timeout を返す。timeout は未着弾の証拠ではない。`herdr agent read <pane>` で着弾を確かめる。
- 進行表示は `<repo>/.claude/tasks/<名前>.json` を更新する(task-board mod が読む)。形式は `references/rules.md` 末尾を見る。`.claude/` が gitignore 済みであることを先に確かめる。

## 5. 中央の受入れと統合

受入れは git の確認だけで行う。

1. `git status --porcelain` が空で、`git diff --name-only main..<branch>` に申告外のファイルが無い。
2. 中央の main へ rebase する(`git -C <worktree> rebase main`)。衝突が大きいときは、その workspace に rebase と解消を依頼する。
3. `git branch -f main <branch>` で fast-forward し、`git log --oneline -1 main` で確かめる。

テストの再実行は、時間依存のテスト(lock の競合、watcher の待ち合わせなど)が受入れ条件の中心にあるときだけ、1回流す。落ちたら元の main で同じテストを1回流し、そちらでも落ちれば既存の問題として記録して進める。再実行のループはしない。

統合したら、すぐに片付ける。

```bash
git -C <repo> worktree remove <worktree>
git -C <repo> merge-base --is-ancestor <branch> main && git -C <repo> branch -D <branch>
herdr workspace close <workspace_id>
```

`git branch -d` は今 checkout している branch と比べて判定するので、main に入っていても「未 merge」と出る。main の祖先であることを `merge-base --is-ancestor` で確かめてから `-D` で消す。各 workspace が作った一時 worktree も片付けさせる。

## 6. 締め切りと終了

- 締め切りを決めたら全 workspace に伝える。締め切りで到達点(commit 済みの範囲・未完の内容)を報告させて止める。
- 間に合わないタスクは統合しない。branch と worktree を残し、issue(新規か既存へのコメント)に、残っていることと再開の条件を書く。
- 統合を見送る基準: 査読の High が残っている、全セッションや送信可否のように影響範囲が広いのに確認が不足している、他タスクと同じファイルを大きく変えていて今入れると rebase のやり直しを生む。
- 終了時は、台帳・issue 下書き・タスク表示を repo の `.claude/tasks/` に置き、memory に再開手順を残す。

## 7. 落とし穴

- バックグラウンドの Bash は600秒で打ち切られる。長いテストは `nohup ... & disown` で切り離し、完了ファイルを Monitor で待つ。Monitor の上限は30分なので、切れたら張り直す。
- zsh では変数が単語に分割されない。`kill $pids` や関数に渡したオプション列は1つの引数になる。`xargs` か配列を使う。
- gh のラベル名は `category: test` のようにコロンの後に空白を含むことがある。`gh label list` で確かめる。
- workspace に負荷試験をさせると、空回しのループが孤児で残り、マシン全体が詰まる。親が1の `while :; do :; done` を `ps` で探して止める。
- codex の agmsg 返信が本文空で届くことがある。そのときは `run/codex-bridge.<team>.codex.log` の該当 turn から本文を読む。
- 中央は自分の session team に join されていないことがある(SessionStart の時点で team が無かった場合)。worker を立てたら `whoami.sh` と team の config で自分の登録を確かめる。

## 関連

- `handoff`: 1タスクを投げっぱなしで委譲する。
- `orchestrate-agents`: 1タスクを段1〜11で通す手順の正本。
- `herdr`: herdr CLI の文法。
