# 依頼文に付ける共通ルール

各タスクの依頼文の末尾に、次をそのまま付ける。`<中央 pane>` は herdr 方式のときだけ、中央セッションの pane ID に置き換える。

```text
## 運用ルール(中央セッション=pane <中央 pane> の claude が統括。ユーザーは事前承認済み)
- cwd は専用 git worktree。他の worktree(特に主 checkout)を編集・branch 切替しない。
- 委譲フローは orchestrate-agents skill に従い、この workspace 自身で回す: プラン起案 → 自分の codex でプラン査読(最大2巡で打ち切り) → 指摘の採否は自分で決める(fable-review は使わない) → 自分で承認 → 実装(自分で書くか impl-worker subagent) → 自分の codex でコード査読(再依頼は High が残るときのみ。Medium 以下は直して終わり)。採否の理由は報告に1行ずつ書く。
- プラン本文は (1) 変更ファイル一覧+目的 (2) 時系列 (3) 受入テスト一覧 に絞り、5 field(疑う前提 / 反対案 / その帰結 / 未解決の問い / 指摘が生じうる経路の列挙)を付ける。プランは $TMPDIR 配下に置き、repo 内に置かない。
- worker 起動: 自分の session team(s-<自分の CLAUDE_CODE_SESSION_ID>)。team が無い/自分が未登録なら先に join.sh で登録する。**Claude Code の agent team の teammate として動く場合は Monitor を張らず**、codex に依頼したら返信を foreground の Bash until ループで待つ(`until history.sh s-<sid> <自分の名前> 5 | grep -q "codex-<名前> →"; do sleep 20; done` を timeout 540000 で実行し、返信が来ないなら再実行する)。idle の teammate は自分の Monitor の通知では起きないため(sunagawasei/agmsg#60)。herdr の別 workspace(別プロセス)で動く場合は従来どおり Monitor で watch.sh を張ってよい(timeout_ms 1800000、切れたら張り直し)。codex は `CLAUDE_CODE_SESSION_ID=<sid リテラル> ~/.agents/skills/agmsg/scripts/ensure-codex.sh <worktree path> codex` を単独の simple command で実行する。
- codex の返信が本文空で届いたら、~/.agents/skills/agmsg/run/codex-bridge.<team>.codex.log の該当 turn から本文を読む。
- `git stash` は使わない(全 worktree で共有され、他の worktree に混入する)。変更前の比較は `git worktree add --detach` した別 worktree で取る。`pkill -f` のようにパターンで他のプロセスを巻き込む kill は使わず、自分が起動した pid を指定する。codex の名前は `codex-<自分のタスク名>` にし、`codex` では ensure-codex を実行しない。
- commit はこの branch へのローカル commit のみ。push・PR・GitHub 操作は禁止。
- live install(~/.agents/skills/agmsg 等の実運用中の配置)への反映・install.sh 実行・既存 worker の despawn は禁止。HOME を書き換えない。
- テストは変更に対する新規・修正テストだけ。全テストは中央も流さない。繰り返し実行の確認は最大5回。CPU を空回しする負荷生成は禁止。
- 終了前に、自分で作った一時 worktree を git worktree remove で片付ける。
- 報告: コード査読が収束したら中央へ `[T<番号> review-done] <commit 一覧と要点>` を送る。teammate は `SendMessage(to: lead)`、herdr 方式は `herdr agent prompt <中央 pane> "..."`(本文にダブルクォート・バッククォートを含めない)。判断に困る/blocked/error のときも同形式で送り、中央の指示を待つ。
```

# タスク表示の JSON 形式

`<repo>/.claude/tasks/<名前>.json`。中央だけが書く。

```json
{
  "updated_at": "2026-10-05T00:30+09:00",
  "session": "<タスク群の名前> (<中央 pane>)",
  "tasks": [
    {"id": "T3", "title": "T3 #24+#47: <何を直すかが分かる説明>", "status": "implementing", "ws": "wCM"},
    {"id": "T5", "title": "T5 #27+#50: <説明>", "status": "merged", "ws": null}
  ]
}
```

status は `implementing` / `fixing` / `waiting (<何>待ち)` / `queued` / `merged` / `deferred (<理由や branch>)` を使う。
