# teammate 版(Claude Code の agent team で並列に回す)

herdr の workspace の代わりに、タスクごとに teammate(in-process の別文脈)を立てる。段取りは SKILL.md と同じで、違いだけを書く。

- **起動**: タスクごとに `git worktree add -b <branch> ../<repo>-wt-<task> main` を切り、Agent tool(`subagent_type: general-purpose`、`name` 付き、`run_in_background: true`)で teammate を起こす。依頼文は「タスク固有の部分」+ `rules.md` の共通ルール。書き込みが重なるタスクは必ず worktree を分ける。
- **報告**: teammate → lead は `SendMessage`。lead の `SendMessage` は idle の teammate を起こす。
- **返信待ち**: idle の teammate は自分の Monitor の通知では起きない(sunagawasei/agmsg#60 で再現確認)。teammate は codex の返信を foreground の `until` ループで待つ(`rules.md`)。lead は全 codex 返信を1本の Monitor(`history.sh` を20秒ごと)で拾い、止まっている teammate へ `SendMessage` で中継する。Monitor は30分で切れるので張り直す。
- **止まっているかの確認**: `ListAgents` の `idle` と、各 teammate の transcript の最終更新時刻(`<project>/<session>/subagents/agent-a<name>-*.jsonl`)を突き合わせる。codex の返信(`history.sh` の `●` 未読)があるのに teammate が `idle` なら、中継する。transcript が10分以上止まっていても、foreground で長い bats を流していれば正常: `ps` の `bats-exec-file` と、その出力ファイルの更新時刻で確かめる(実例: 14ファイルの bats で約18分、load average 27)。
- **終えた teammate は閉じる**: 統合したら `TaskStop`(`task_id` に teammate 名)で停止する。停止後に届く idle 通知は、直前の報告の繰り返しなので無視してよい。codex worker は teammate の停止では片付かない。全 teammate の終了後に、`CLAUDE_CODE_SESSION_ID=<中央の sid> despawn.sh <session team> lead <codex 名>` を名前ごとに実行する(消えた worktree の worker は `status=forced` で止まる)。他プロジェクトの bridge(`ps` で `--project` が別)は触らない。
- **受入れ**: `SKILL.md` §5 と同じ。teammate が `review-done` を送らないまま止まることがある。codex の最終返信が High なしで、worktree が clean なら、lead が受入れ検査をして統合してよい。
- **片付け**: 追加の指示を送った teammate の worktree は、その完了報告まで消さない。消す前に `git status --porcelain` で未commit の変更を見る(統合と追加修正が重なって、未commit のテストを失いかけた実績)。
- **再起動**: teammate は `/resume` で復元されない。worktree と branch は残る。未統合 branch と、codex の返信の位置を memory に書いてから再起動する。
- **同時数**: CPU に合わせて4〜6。
