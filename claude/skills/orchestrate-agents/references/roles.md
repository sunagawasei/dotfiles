# ロール一覧(権限・隔離・pin)

各ロールの権限境界・隔離・モデル指定の正本。フロー全体と不変条件は `claude/CLAUDE.md` の「フロー」節と `flow-rules.md`。

- **メイン(本セッション)**: ユーザーとの対話でのプラン化・統合・検収(要件適合+diff査読+git log+test)・git。権限=**writeの承認・統合・commitの唯一の制御主体**
- **design-review(Opus 5.5・headless cursor)**: codex査読(段2)指摘への対応可否収束(段3)と、codexコード査読(段9)指摘への対応可否収束を担当。独立した設計レビュー・コードレビューは行わない(指摘の振り分けだけを行う)。findings-onlyでrepo不変。権限=cursor_readonly(Write/Shell deny。readはcredential denylist。projectが`~/.config`だと`gh`/`gcloud`/`cursor`/`codex`の設定dirはworkspace内でdenyされない。credential path readと外部状態変更の禁止はrole file規約でのみ抑止)。モデルは`spawn.cursor_model.design-review: claude-opus-5-5-medium`(label監査あり。実測`Claude Opus 5.5 300K Medium`)。Cursorチーム枠を消費しAnthropicのclaude-code枠は使わない(2026-11-17にclaude-code版から移行)。cursor harness経由なので秘密の実値を含むpayloadは送らない。通常・節約の両モードで同じ。手順は`claude/skills/orchestrate-agents/SKILL.md`段3。fable-reviewは休眠(`model-routing.md`「休眠ロール」)
- **休眠ロール(fable-review・opus-review・manager/watcher・codex-impl)**: 平常時は起動しない。fable-reviewは段3・段9の振り分けをdesign-review(opus)へ移して休眠。opus-reviewは高リスク変更でメインが第2意見を要すると判断したときだけ起こす。manager/watcherは実装レーンをcodex workerへ戻すときだけ、codex-implは節約モードの段5だけで使う。権限・pin・復帰手順: `claude/skills/orchestrate-agents/references/model-routing.md`「休眠ロール」
- **grok-review(Grok4.6・headless cursor)**: codex(review役)応答不能時の段9フォールバック(暫定運用)。codex同等の役割(意図一致査読+脆弱性4観点)を代行する。実装がAnthropic側に寄った現構成では、cross-vendorの査読を保つ唯一の代替経路。findings-onlyでrepo不変。権限=cursor_readonly。トリガー・復帰・停止条件・waiver・fallback不能クラスの判定は`claude/skills/orchestrate-agents/SKILL.md`「段9査読者フォールバックチェーン」節が正本
- **実装subagent(Claude Code組み込みAgent tool・`claude/agents/impl-worker.md`・sonnet)**: 承認済みsubtaskの**ファイルセットの範囲だけ**編集する。commit/push禁止。完了報告はメイン宛。親セッションの権限をそのまま継承するため、外部write・repo外read・機密情報readの禁止は**契約文でのみ抑止する**(強制境界は無い)。実行中にメインへ問い合わせることはできず、判断が要る場面では変更を加えず`question`として返す
- **codex(review役)**: プラン査読(承認依頼前の常時ゲート)+段9のdiff査読。**意図一致査読と脆弱性4観点の両方を全hunk対象で担当する**(実装がAnthropic側に寄ったため、cross-vendorの査読はここ1本になる)。権限=read-only。起動はreviewer/research専用のCODEX_HOME(空execpolicy Rules・config.toml無し)を使い、対話用途の広いallow rule(`docker compose`等)を継承しない。networkはghのHTTPSホストのみを許すallowlistが有効なときだけ起動し、allow外ホスト・直接IP・allowlist不通のいずれかでは起動しない(`sunagawasei/agmsg#42`/`#43`)
- **codex-research**: コードベース内・外部ソース読解の横断調査。file:line一覧・構造化データを返す。パッチは作らない。権限=read-only運用。起動時の隔離は上のcodex(review役)と同一(同じlauncher・同じ専用CODEX_HOMEを使うため)
- **grok-research(Grok4.6・headless cursor)**: 公開情報とredacted packetに限った第二の調査経路。権限=read-only
