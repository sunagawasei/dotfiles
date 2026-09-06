# 委譲運用の経緯・根拠(sonnetサブエージェント時代を含む)

判断規則そのものはグローバル`CLAUDE.md`の「エージェント役割分担」節が正本で、ここには経緯だけを置く。sonnetサブエージェントとSonnet reviewerは2026-08-21の単一フロー化で退役した — 以下は当時の判断の記録。

## 前提（メインがsonnetの場合は委譲しない）

2026-07-13ユーザー指示。メイン自身がsonnetで動いているセッションでsonnetサブエージェントへ委譲すると、同一課金プール・同一モデルのためコスト裁定が効かず、往復オーバーヘッドが純増するだけと判明した。別課金プールのcodex系（codex-impl/codex-research）への委譲は課金プールが別なので、この制約を受けない。

## メインはファイル編集・作成を自分で行わない

2026-07-04ユーザー指示が起点。当初は「Edit/Writeは全部sonnetへ」だったが、2026-07-12〜のTorishima構成導入でコード挙動(logic)を変える編集はcodex-implへ切り出す方針が追加され、2026-07-13にその適用範囲を「新機能・複数ファイル」から「行数・機能性を問わず」へ拡大した。sonnetへの委譲は機械的編集・棚卸し・一括修正・データ収集等、logicに触れない作業に残った。

## 調査・データ収集も自分で行わない

2026-07-06ユーザー指示で全プロジェクト常時適用に。その後の運用で外部リポジトリのソースコード読解も「調査」に含まれることが判明し、2026-07-13「そういう確認はcodexにさせてよ」・2026-07-14「そういう調査はcodexにさせてよ。グローバルでその設定にして」という指摘を受け、GitHub上のソースをraw取得して読む場合も含め、コードを読んで挙動を判定する調査はsonnetでなくcodex-researchへ回す運用に改めた（実例: herdr `integration install` の挙動をGitHub rawソースから調べる件をsonnetに投げたのは誤りだった）。

## Agent toolの`model: sonnet`明示とCLAUDE_CODE_SUBAGENT_MODEL

2026-07-07〜、`~/.config/claude/settings.json`の`env.CLAUDE_CODE_SUBAGENT_MODEL=sonnet`で全サブエージェントをsonnetへ環境変数レベルで強制する保険を追加した。これにより「重要な最終検証・判断だけ`model: opus`をピンポイント投入する」という以前の例外運用（2026-07-07以前）は機能しなくなり、廃止した。opusでの検証が必要な局面は、以後サブエージェントに委譲せずメインセッション自身が直接判断する。

## 委譲対象の拡大（調査に限らない）

2026-07-12〜、棚卸し・一括修正・セッション履歴抽出・issue一括登録など、指示を自己完結パケット化できる定型作業は実装・編集までsonnetにやらせる運用に広げた（ただし実質的な機能実装はcodex-implが先）。

## スポットチェックの実例

「修正したと報告したが未適用」だった実例が2026-07-04に発生。以降、sonnetの「編集した」報告は鵜呑みにせず、grep/存在確認で必ずスポットチェックする運用にした。

## 些末な自己完結タスクの直接処理

2026-07-25の`/insights` frictionで判明。メモリへのノート書きなど、diffだけで正しさが自明かつ単発で完結する極小タスクをsonnetサブエージェントに投げたところ、spawnしたsubagentが返らずtrivialタスクが無限待ちでstallする実害（note-writing task無限待ち）が発生した。以降、この種のタスクはsonnetにもcodexにも委譲せずメインが直接処理する。

## 2026-07-29: Claude 5シリーズ追従とモデル指定方針の明文化

- モデル指定はalias自動追従を正とする(固定model IDは書かない)。2026-07-29時点の解決先: `sonnet`=Sonnet 5(claude-sonnet-5)、`fable`=Fable 5(claude-fable-5)。aliasの解決先は公式仕様として「プロバイダ推奨版へ自動追従」であり、固定したい場合のみフルIDまたは`ANTHROPIC_DEFAULT_*_MODEL`を使う
- `CLAUDE_CODE_SUBAGENT_MODEL`はaliasを受け付けて自動追従し、呼び出し時の`model`パラメータ・agent定義frontmatterより優先される(公式仕様で確認済み)。v2.1.196以降、`inherit`を設定した場合は「未設定」と同じ扱いになり通常のモデル解決(呼び出し時パラメータ→frontmatter→メインモデル)が続行される(旧版はメインモデルへ強制だった)
- 役割分担の経済的根拠(来歴): 別課金プールのcodexへの外注がトークン削減の本質。Anthropicプール内では上位tierメインとsonnetサブエージェントの単価差が委譲の裁定になる

## 2026-09-06: 実装レーンをcodex workerから組み込みsubagentへ

codexへの実装依頼で意図とのズレ・やり直しが多く所要時間が伸びている、というユーザーの体感が動機。実装を`claude/agents/impl-worker.md`(sonnet/xhigh)へ移した。2026-08-21に退役していたsubagentへの実装委譲が、形を変えて復活したことになる。

**動機は体感の改善とAnthropic側への一本化であり、「codexが原因だと確定したから」ではない。** 段2のfable-reviewは「ズレの発生点がmanagerの再パケット化にある可能性が高く、切り分けずに2変数を同時に動かしている」と指摘したが、ユーザー判断で切り分けは実施しない。切替後の効果計測も行わない(体感で運用する)。したがってどちらが効いたかは今後も確定しない。

確定した内容:

- manager/watcherを実装レーンから外す。メインが分割・dispatch・受け入れ検査を持つ。scope-okトークンとfingerprint照合は、codex workerがsandbox内で外から観測できないことへの対処だったので役目が消えた
- 段9の査読はcodexへ集約(意図一致+脆弱性4観点の両方、全hunk対象)。実装がAnthropic側に寄ったためopus-review(Anthropic)は一次査読に使えない。**これは「独立した2つの査読」ではなく「Anthropic authorに対する単一のcross-vendorゲート」**で、観点の独立性は作られない。opus-reviewは高リスク変更の第2意見として休眠
- 手順書にはsubagentレーンだけを書く。codexレーンへ戻すときはgit履歴から読んで適用する。role fileとagmsg configは削除せず残す
- 外部write抑止のフックは導入しない(ユーザー判断)。subagentは親の権限を継承し、`Bash(gh api:*)`等が許可済みのため外部writeは無プロンプトで通る。**remote mutationは事後検知もできない残存リスク**

**cred-split決定の反転**: `.claude/docs/cred-split/E-decision-table.md`は「機械的編集の束ねはcodex-impl、Claude subagent経路(7.21%)を可搬分として外へ出す」を決定として記録している。本変更はこれを反転させ、Anthropicプールの消費が増える(増分の上限は当時の7.21%)。メインOpus / subagent sonnetの単価差は裁定として残る。

**実効model/effortの確認方法**: `claude/projects/<project-slug>/<session>/subagents/agent-<name>-<id>.jsonl`に`"model"`と`"effort"`が残る。メインのtranscriptには`isSidechain`レコードとして現れない。2026-09-06の実測は`claude-sonnet-5`・`effort: xhigh`で、frontmatterの値がそのまま実効値になっていた。`CLAUDE_CODE_SUBAGENT_MODEL`(=sonnet)との優先順位は両方がsonnetのため未判別のまま。**frontmatterで`model`を変えた定義を追加したら、初回dispatch後に必ずこのファイルで確認する**(2026-07-29の記録はenv優先、公式docは v2.1.251以降frontmatter優先とあり、両説が未解決)。
