# ロール一覧(権限・隔離・pin)

各ロールの権限境界・隔離・モデル指定の正本。フロー全体と不変条件は `claude/CLAUDE.md` の「フロー」節と `flow-rules.md`。

- **メイン(本セッション)**: ユーザーとの対話でのプラン化・統合・検収(要件適合+diff査読+git log+test)・git。権限=**writeの承認・統合・commitの唯一の制御主体**
- **finding-triage(haiku・headless claude-code)**: codex-reviewとの査読ループ(段2・段3、段9)の調停役。codex-reviewのfindingsのうちMedium以上を採用/見送り/別タスクに振り分けてcodex-reviewへ返し、異議が無くなるか3往復に達したら、最終表だけをメインへ送る。Low以下は落としてよく、メインには返さない。独立した設計レビュー・コードレビューは行わない(指摘の振り分けだけを行う)。repoは変更しない。権限=sandboxがrepoへのwriteを拒否し、readはproject dirと継承add-dirに限る。credential path readと外部状態変更の禁止、agmsg message storeへの直接書き込み禁止はrole file規約でのみ抑止する。モデルは`spawn.claude_model.finding-triage: haiku`(alias自動追従)。通常・節約の両モードで同じ。手順は`claude/skills/orchestrate-agents/SKILL.md`「査読ループ」節
- **実装subagent(Claude Code組み込みAgent tool・`claude/agents/impl-worker.md`・sonnet)**: 承認済みsubtaskの**ファイルセットの範囲だけ**編集する。commit/push禁止。完了報告はメイン宛。親セッションの権限をそのまま継承するため、外部write・repo外read・機密情報readの禁止は**契約文でのみ抑止する**(強制境界は無い)。実行中にメインへ問い合わせることはできず、判断が要る場面では変更を加えず`question`として返す
- **codex-review**: プラン査読(承認依頼前の常時ゲート)+段9のdiff査読。`[loop]`ヘッダ付きpacketのときは、findingsをfinding-triageへ送り、その振り分けに「異議なし」か「異議(根拠)」を返す(ループ手順はSKILL.md「査読ループ」節)。**意図一致査読と脆弱性4観点の両方を全hunk対象で担当する**(実装がAnthropic側に寄ったため、cross-vendorの査読はここ1本になる)。権限=read-only。起動はreviewer/research専用のCODEX_HOME(空execpolicy Rules・config.toml無し)を使い、対話用途の広いallow rule(`docker compose`等)を継承しない。networkはghのHTTPSホストのみを許すallowlistが有効なときだけ起動し、allow外ホスト・直接IP・allowlist不通のいずれかでは起動しない(`sunagawasei/agmsg#42`/`#43`)
- **codex-research**: コードベース内・外部ソース読解の横断調査。file:line一覧・構造化データを返す。パッチは作らない。権限=read-only運用。起動時の隔離は上のcodex-reviewと同一(同じlauncher・同じ専用CODEX_HOMEを使うため)
