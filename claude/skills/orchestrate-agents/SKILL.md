---
name: orchestrate-agents
description: 全タスク共通の単一委譲フロー(対話でのプラン起案→fable-review設計ゲート→codexプラン査読→ユーザー承認→メイン分割→実装subagent→メイン受け入れ検査→統合→codexコード査読→commit)の運用手順。査読役へはagmsgの非同期send+Monitor自動再開、実装はClaude Code組み込みのAgent tool
---

# 単一委譲フロー

委譲の形態は1本だけ。第二のフロー(旧「直委譲」「roleチームモード」)は持たない。役割分担と権限の正本はグローバル`CLAUDE.md`の「エージェント役割分担」節で、このファイルは**各段の実務手順・パケット書式・診断**の正本。

## 前提

- このセッションのagmsg Monitor(`watch.sh ... --team s-<このセッションのUUID>`)がSessionStartから常駐している。**SessionStart時点でそのプロジェクトのteamが無いと常駐しない**(session-start.shは既存registrationから解決するため)。teamを作ったら`ps aux | grep 'watch.sh <このセッションのUUID>'`で確認し、無ければMonitorツールから`watch.sh <session_id> <project_path> <agent_type> <active_name> --team <team>`を自分で張る(pidfileは`run/watch.<session_id>.<n>.pid`)
- **自前のpollingループでMonitorを代用しない**。history.sh/inbox.shを叩くループは(a)inbox.shが既読化して返信を消費する (b)宛先フィルタの正規表現が取りこぼす、の2つで壊れる。2026-09-01の実例: `(codex|fable-review|...) → main`のパターンが`codex-research → main`にマッチせず、返信到着に9時間気づかなかった。**このときMonitorプロセス自体は生きており、「Monitorが死んだ」という一次診断も誤りだった** — 無反応時はプロセス生死とフィルタの両方を疑う
- 宛先workerのbridgeが稼働していること。遅延spawnは自動発火しない(下記「送信の実務」)
- 送信は非同期`send`が既定。`ask --wait`は使わない(codexでask往復が機能しない実績)

## フロー

### 段1 プランニング(メイン+ユーザーの対話)

メインがユーザーとの対話でプラン案を起案する。**この段より前に第三者のチャレンジは無い**ので、疑いはメインが自分で言語化する。プランの確定は段4のユーザー承認で、段1は起案まで。

- 段2へ渡すpacketには**4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須**で載せる。「該当なし」と書くなら理由も書く
- fable-reviewはこの4 fieldの**欠落・空・定型的で実質のない値**をfindingとして返す。自己申告の空洞化に対する検査主体はここだけなので、形だけ埋めて通さない
- 対話は依頼者のフレームに錨を下ろしやすい。**問題設定自体を疑う**役はメインと段2の両方が負う

### 段2 fable-review(設計ゲート)

プランを`fable-review`へ送る。返るのは findings。**ユーザー承認の代替にしない**。認証・秘密情報を含むプランは送付前に該当部分をredactする(redactすると議論不成立ならF3の規約どおり段2を省く)。

```bash
AGMSG_CLAUDE_PROBE_TIMEOUT=150 ~/.agents/skills/agmsg/scripts/ensure-headless.sh claude-code <path> fable-review
```

**probe timeoutは既定30秒では足りない**(`_spawn.sh`の`AGMSG_CLAUDE_PROBE_TIMEOUT:-30`)。fableは30秒でprobeの全tool eventを出し切れず`rc=124`でfail-closedする(2026-08-29実測)。turn timeoutと違い**per-nameのconfigキーは存在せず、env varでのみ指定する**ため、起動は常に上記の環境変数込みの1行で行う。値は`$VAR`にせずリテラルで書く(troubleshooting.md項10)。失敗時の診断はtroubleshooting.md項14。

driverはclaude-code。read-onlyはreviewer layout(グローバル既定`spawn.claude_reviewer: true`で担保。per-nameキーは存在しないためグローバルキーで運用する)。**repo writeはsandbox+spawn probeで強制されるが、repo read(project配下=`~/.config`全体+継承add-dir)はBash経由で開く**。credential path read・外部状態変更(認証済みCLIの変更系操作・network write)の禁止はrole file規約でのみ抑止し、sandboxは強制しない(強制境界ではないことを認識のうえ運用する選択)。**sandboxのwrite denyはrepoに限られ、agmsg message store/team registration/run状態はBashから技術的に書け、`$SKILL_DIR`全体(全team・全project分)がread可能** — これも規約でのみ抑止する残余リスクとして扱う(F1/F2)。network egressは未検証(`not checked`、F6)。モデルは`spawn.claude_model.fable-review: fable`(alias、自動追従)。effortは`spawn.claude_effort.fable-review: high`。claude-code driverにはcursorのようなmodel-audit機構が無く、実効モデルの動的な機械照合はできない。記録は静的設定値に留め、モデル自身への自己申告は検証証跡として扱わない(残存リスクとして記録するだけに留める)。

**turn timeoutの既定300秒では足りない**: `spawn.claude_turn_timeout.fable-review: 1800`で個別設定する(per-nameキーが存在する)。add-dir継承は`spawn.claude_inherit_add_dirs.fable-review: true`で明示する(グローバル既定はoff)。

**layout差はreadiness照合では検出できない**: spawn前後で`spawn.claude_implementer.fable-review`が未設定(または`false`)であること、生成済みsettings.jsonのdenyWriteにprojectパスが含まれることを確認する(F4)。

**rollback範囲は3点セット**: `git checkout`で戻るのはrole fileのみ。agmsgのspawn状態(despawn→旧cursor版`fable-review.cursor.md`で再spawn)とconfig(`spawn.claude_model.fable-review`等のkeyをcursor向け設定へ手動で戻す)は別途の操作が要る。3点いずれか1つの復旧漏れは新driverと旧configの不整合(dead-letterや誤ったlayoutでのspawn)を招くため、rollback時は3点を1セットで実行し、実行後に段2のreadiness照合(spawn世代を跨いだ返信実績の無効化含む)をやり直す。

**role fileの編集は稼働中workerに反映されない**: claude-code/cursor/codexいずれのdriverも`_spawn.sh`がspawn時にrole fileを`run/`配下へ`cp`し(`<type>-bridge.<team>.<name>.role`)、bridgeはそのスナップショットだけを毎ターン読む。`db/spawn-roles/<name>.<type>.md`(= `claude/agmsg-roles/`のsymlink先)を直接編集しても、稼働中のworkerはspawn時点の内容のまま動き続ける(意図的な設計: source側の編集/削除でlive workerを不意に変えないため)。role file変更を反映するには、その名前のworkerをdespawn→再spawnし、readiness照合(probe)をやり直す。

### 段3 codex(review役)のプラン査読

段2(fable-review)を実施した場合、その査読を完了した最終返信(Findings/Required tests/Residual risk/Confidence形式のもの。findingsゼロの場合もこの形式で明示)を受け取ってから送る(ACK・受領通知・途中報告・blocked通知はこの形式を持たないため対象外)。**workerのturnは完走したがagmsg配送だけ失敗した場合に限り、bridgeログからの代替を認める**。代替として有効なのは、(a)**現行taskと同一teamの**bridgeログでfable-review自身が送信したエントリ(team一致・送信者一致)、(b)現行の段2査読依頼のdispatch時刻(agmsg message storeのtimestamp)より後、(c)fable-reviewの直近respawnより後(runログのspawn時刻。respawnが一度も無い場合はfable-review初回spawn時刻より後)、(d)完了形式を備え内容が現行プランの査読であること、の4条件をすべて満たすentryだけ。**この4条件は規約レベルの信頼に依る照合であって、bridgeが強制する境界ではない**(agmsg message store等はfable-review自身がBashから書ける。既知の残余リスク)。**条件を満たすentryがログに無ければ代替の出番はなく、troubleshooting.mdの診断→解消できなければユーザーへ報告して指示を仰ぐ経路一本になる**(段3の先行送信で回避しない)。段2の返信前に段3へ査読依頼を送らない。段2省略(**redactすると議論が成立しない場合**)は、段2への査読依頼を送信する前に決定した場合に限る。依頼送信後のtimeout・無返信・失敗は省略に再分類できない(respawnした場合は返信実績が無効化されるため再probeをやり直す)。

渡すプランには段2の指摘への対応を反映する。段3packetの必須fieldとして、段2の全findingについてfinding ID・振り分け(採用/見送り/別タスク)・見送りと別タスクは理由を一対一で列挙する(段2省略時は「段2省略(redactすると議論不成立、依頼送信前に決定)」と明記)。**認証・秘密情報を含むfindingは理由を`redacted(理由: 認証/秘密)`で代替でき、これはpacket不備に当たらない。redactedを使ったfindingは、内容(理由の具体)を平文開示せずfinding ID・重大度・振り分け(採用/見送り/別タスク)・解決状態(対応済み/未対応)を段4のユーザー承認時と段11の検収報告の両方に明記する(ユーザーは元のfable-review返信にagmsg履歴から直接アクセスできる)**。一覧から落ちたfindingがあればpacket不備として扱う。この対応は初回送信に適用し、段3findingsを受けた再送(収束ループ)は下記「レビュー収束条件」に従う。

承認依頼前の常時ゲート。書式は下記「[review]パケットの鉄則」。

### 段4 ユーザー承認

**このゲートを通る前に、実装subagentへdispatchを1件も出さない。** 承認対象はサブタスク方針を含むプラン全体。

### 段5 メインの分割とdispatch

メインが承認済みプランをサブタスクへ分割し、実装subagent(`claude/agents/impl-worker.md`・sonnet/xhigh)へdispatchする。managerは介さない。

**並列で走らせるときの2条件**(どちらも守れないなら直列にする):

- **ファイルセットを重複させない**。repo全体のformatter、codegen、同じDBやportを使うtestのような副作用も競合するので、ソースのパスが分かれているだけで安心しない
- **dispatch中はメインが同じrepoを編集しない**。メインの並行編集があると、完了報告と`git diff`の突き合わせで帰属が確定しなくなる

**subagentは実行中にメインへ問い合わせられない**。判断が要る場面に当たったら変更を加えず`question`として返る契約になっている。メインが回答し、`SendMessage`で同じsubagentを継続させる。

**権限確認の要るコマンドではstallしない**(2026-09-06実測): 背景で動くsubagentがask対象のコマンド(`Bash(rm -rf *)`等)に当たると、プロンプトは**メインセッション側に`from the <agent> agent`として出る**。ユーザーが承認すればsubagentはそのまま続行し、subagent側からは普通に成功したようにしか見えない。2026-07-25のnote-writing無限待ちのような沈黙は起きない。拒否した場合の挙動は未確認。

### 段6 メインの受け入れ検査

完了報告を受けたら、メインが次を確認する。

- **申告と実体の突き合わせ**: 完了報告の変更ファイル一覧と`git diff --stat`を比較する。**申告外のパスに差分があれば、そのsubagentの逸脱として差し戻す**。並列時は「誰の変更か」を集約diffから復元できないため、説明のつかない差分が1つでもあればゲートを止める
- **検証の再現**: 報告された検証コマンドを1-2本スポット再実行する。「検証したと報告したが未実施」はこの環境では他に検出手段が無い(2026-07-04前例)
- **git mutationの検出**: dispatch前後で`git rev-parse HEAD`・`git show-ref`・`git status --porcelain`を比較する。`git log`だけでは既存commitのpushや`gh`経由のremote mutationは検出できない(remote mutationは事後検知できない残存リスク)

`question`で返った場合、**partial changeが残っている可能性がある**。subagentに自己判断でrollbackさせない(共有worktreeなので他の変更を壊す)。partial changeは変更ファイル一覧に明記させ、扱いはメインが決める。

### 段7 全subtask通過(非終端)

全subtaskが受け入れ検査を通ったら段8へ。ここは終端ではない。

### 段8 メインの統合とauthor再分類

メインが統合し、自分が実質変更したファイル/hunkを「メイン作」へ再分類する。author再分類マップは**「メイン作 / subagent作」の2区分**で、用途はfindingの差し戻し先の判定。

### 段9 コード査読ゲート

**codex(review役)が全hunkを査読する** — 意図一致・正しさと、脆弱性4観点(認証/認可境界・secret出力・外部write・dependency advisory)の両方。

- 実装がAnthropic側(メイン・実装subagent)に寄ったため、cross-vendorの査読はcodex1本になる。**これは「独立した2つの査読」ではなく「Anthropic authorに対する単一のcross-vendorゲート」**であり、観点の独立性は作られない。高リスク変更で第2の目が要るとメインが判断したときだけopus-reviewを起こす
- 段8のauthor再分類マップを渡す。**マップの用途は差し戻し先の判定で、査読対象は常に全hunk**
- findingのラベルは`[subtask:<id>]`(subagent作)・`[author:main]`(メイン作)・`[design-level]`(承認済み設計自体の欠陥。対応する単一hunkが無ければfile:line欄に`(design-level, no single hunk)`と明記)
- **依存を変えるdiffのadvisory確認はメインの明示責務**。段9パケットの必須fieldに確認結果と根拠を含める(欠落時はcommit不可)
- 段1〜11を通らない原子的編集の例外は、資格要件で認証・security boundary・課金・外部write・依存関係を変えないことが担保されるため脆弱性4観点の対象外
- fable-reviewは段2の設計レビューと段3の対応可否収束を担当し、段9には関与しない

### 段9査読者フォールバックチェーン(codex応答不能時)

1. **主**: codex(review役)
2. **フォールバック**: grok-review(cursor・config key `spawn.cursor_model.grok-review: cursor-grok-4.6-xhigh`、暫定運用)。codex同等の役割(意図一致査読+脆弱性4観点)を代行する
   - **grok-reviewのreadiness照合は常に厳格基準**: 期待token受信・エラー応答でない・pinしたmodel identityが一致、の3点。初回spawn時に不成功ならgrok-reviewは使用不能と判定し項4へ。復帰判定時に不成功なら主は未復旧と判定しフォールバックを継続する
   - role file(`grok-review.cursor.md`)にreadiness probeへの応答例外を記載済み
   - **レビュー本体の成功基準**: 「[review]パケットの鉄則」の出力形式(Findings/Required tests/Residual risk/Confidence。clean reviewはFindingsなしを含む)を主が返すこと。返さない場合はspawn/probe時エラー・usage-limit応答・timeout・無応答・部分応答・形式不正・異常終了のいずれであってもすべて不成功とみなしフォールバックへ切替。**切替時は部分査読を継ぎ足さず段9を最初からやり直す**(遅着した主の応答とフォールバック応答を二重採用しない)
   - **停止条件**: grok-reviewの誤陰性・査読形式不良を1件でも観測した時点でフォールバックを停止する。停止状態はセッションをまたいで有効とし(このファイルまたはCLAUDE.mdへの追記で記録)、定義した再有効化手順(ユーザー承認を得て本節を書き換える)以外では再開しない
3. **フォールバック不能クラス**: 最終outbound payload全体(diff本文+補足説明+ログ+添付コンテキスト)に秘密の実値(credential・token等)が含まれると判断される場合、grok-review(cursor系)は対象外。機械的secret scannerは実装せず、メインの目視確認に依る(誤判定は残存リスクとして受容)。この場合はcodex単独で、codexも不能なら項4
4. **段9ブロック**: 主・フォールバック双方が使用不能な場合。段9を無言スキップせずタスクをブロックしてユーザーに報告する。**ユーザーの明示的waiverがある場合のみ例外**とし、waiverは3条件を満たす: (a)対象diffをrepository identity・base tree・commit対象の完全なstaged treeを含むfingerprintに固定し、commit直前に再照合する (b)「段9成功」ではなく「段9未実施・ユーザーwaiver」として検収記録に残す (c)waiver後に対象diffが変更されたら旧waiverは無効・再承認必須

外部送信の安全境界: cursor harness経由の査読役(grok-review・opus-review)へは、ユーザーが受容したデータ境界としてprivate diff送付を許容する(2026-08-27ユーザー確認済み)。調査役(grok-research等)への最小化義務(未公開コード断片を含めない)とは別軸。ユーザーが受容を撤回した時点で即時無効。

**grok-reviewとgrok-researchは名前が1語しか違わないが権限が異なる別worker**: grok-review(査読役、private diff送付許容)とgrok-research(調査役、最小化義務でsecret・未公開コード断片禁止)を混同しない。`[review]`パケットの宛先は必ず`grok-review`。

### 段10 差し戻し(3経路)

- **subagent作のfinding**(`[subtask:<id>]`ラベル) → `SendMessage`で当該subagentへ差し戻す。**完了済みのsubagentも名前かagentIdで再開でき、前ターンの文脈を保持している**(2026-09-06実測。パケットに書いた識別子と、前ターンにしか出ていないコマンド出力の両方を答えられた)。**セッションを跨いで文脈が失われている場合は新規spawnし、findingに加えて元のsubtaskパケット(ゴール・制約・ファイルセット)を同梱する**(findingだけ渡すと、設計意図を知らないagentがその行だけ直す)
- **メイン作hunkのfinding**(`[author:main]`ラベル) → メインが直して段9へ再投入
- **設計レベルのfinding**(`[design-level]`ラベル) → メインが理由を明記して該当taskを終端し、新規`[task:<id>]`を発行して段1から再起動する(旧taskとの関連は理由欄で相互参照)。既存taskの延命はしない

### 段11 検収→commit

段9のfindingが全て解消したら、**メインが最終diffを固定し、必須検証を自分で再実行する**。差し戻しで内容が変わった後のdiffが、査読を通したdiffと同一であることを確認するため。watcherもfingerprintも無いので、ここが唯一の「検証済み状態と最終成果物を結び付けるゲート」になる。

検収の中身: `git -C <repo> status` / `git diff`で(1)要件適合・プラン逸脱 (2)正しさ・エッジ・回帰リスク を読む。`git log`と`git show-ref`で勝手commit・branch mutationがないことを確認する。

ユーザー確認後にcommitする。検収失敗・ユーザー差し戻しは段10の該当経路へ戻る。終端は3つだけ: **メインのcommit** / **ユーザーの中止** / **taskの中止**(変更不要・実行不能・段取り不備と確定した場合。メインが理由を記録し、走っているsubagentを`TaskStop`で停止して終端する)。

### 実装レーンの2形態(subagent / team)

既定は**subagent**。teamを使うのは次のどちらかに当たるときだけ。

- サブタスク同士が**作業中に**情報をやり取りする必要がある(片方の中間結果をもう片方が引き取る)
- 3人以上を並行させ、共有タスクリストで進行管理する規模になる

**この環境ではteammateは別プロセスにならない**(2026-09-06実測。`~/.claude/teams/<team>/config.json`が`backendType: "in-process"`。別ペインで立てるにはtmuxかiTerm2が要り、WezTermは対象外)。「独立した文脈で並行に考えさせる」も「あとから呼び戻して継続させる」もsubagentで取れるので、上の2つに当たらないならsubagentを選ぶ。

- 有効化は`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`(`claude/settings.json`、git管理外)
- `claude/agents/impl-worker.md`の定義はsubagentとteammateの両方でそのまま使える。ただしteammateとして起動すると`skills`は定義側の指定が無視されproject設定から読まれる
- 制約: 1セッションに1チームまで / teammateはさらにteammateを立てられない / `/resume`でteammateは復元されない / spawn時にteammateごとの権限モードを指定できない
- 状態は`~/.claude/teams/<team>/config.json`、共有タスクは`~/.claude/tasks/<team>/`、受信箱は`teams/<team>/inboxes/<name>.json`
- 停止は`TaskStop`にteammate名を渡す

### 実効model/effortの確認

subagent・teammateとも、実効値は`claude/projects/<project-slug>/<session>/subagents/agent-<name>-<id>.jsonl`に残る。`"model"`と`"effort"`をgrepする。メインのtranscriptには`isSidechain`レコードとして現れないので、そちらを探しても見つからない(2026-09-06実測)。

frontmatterの`model`/`effort`はそのまま実効値になる(同日実測で`claude-sonnet-5`・`effort: xhigh`)。`CLAUDE_CODE_SUBAGENT_MODEL`との優先順位は、両方がsonnetのため未判別。**frontmatterで`model`を変えた定義を追加したら、初回dispatch後にこのファイルで実効値を必ず確認する**。

## findingラベル

| ラベル | 出す | 受ける | 意味 |
|---|---|---|---|
| `[subtask:<id>]` | codex / grok-review | メイン | subagent作hunkのfinding。SendMessageで当該subagentへ差し戻す |
| `[author:main]` | codex / grok-review | メイン | メイン作hunkのfinding。メインが直す |
| `[design-level]` | codex / grok-review | メイン | 承認済み設計自体の欠陥。file:lineが無い場合は`(design-level, no single hunk)`と明記。段10の第3経路(task中止→新task発行)へ |

完了サイクルは**段5のdispatch(またはreopen時の再dispatch)で開き**、段11の検収完了・段10の差し戻し・taskの中止のいずれか1つで閉じる。段7は境界ではなくサイクル内のマイルストーン。

## 共通の不変条件

- `[task:<id>]`はセッション内で一意。`[subtask:<id>]`はtask内で一意
- 完了済み・破棄済みタスクのstale/duplicate messageは既読化して無視する
- **agmsg相手のreadiness照合**は「名前がある」ではなく、各nameのregistrationが**期待typeでちょうど1件、かつ当該セッションで返信実績があること**(dead-letterはregistration照合だけでは検出できない)。**新規spawn直後で返信実績がまだ無い場合はtrivialなprobeパケットを1通送り、その応答到達をもって返信実績とする**。**respawn(despawn→再spawn)した場合、respawn前の返信実績は無効**として扱い、必ずrespawn後に新規probeを送り直す。対象はfable-review・codex(review役)、および実際にdispatchするcodex-research/grok-research/grok-review。**実装subagentはagmsgに乗らないのでこの照合の対象外**
- 起動コマンド(モデル等のconfigはグローバル永続なのでコマンドのみ):
  - codex系(codex, codex-research): `ensure-codex.sh <project> <name>`
  - claude-code系(fable-review): `AGMSG_CLAUDE_PROBE_TIMEOUT=150 ensure-headless.sh claude-code <project> fable-review`(既定30秒ではfableがprobeを出し切れずrc=124でfail-closedする。probe timeoutはper-nameのconfigキーが無くenv varのみ。model/effort/turn timeoutはper-nameのconfigキーで固定。`spawn.claude_turn_timeout.fable-review: 1800`を設定済み。段2参照)
  - cursor系(grok-review、codexフォールバック時のみ): `AGMSG_CURSOR_BRIDGE_TURN_TIMEOUT=1800 ensure-headless.sh cursor <project> grok-review`
  - cursor系(opus-review、第2意見が要るときのみ): `AGMSG_CURSOR_BRIDGE_TURN_TIMEOUT=1800 ensure-headless.sh cursor <project> opus-review`(既定180秒ではopus:highの査読が切れる)
  - cursor系(grok-research): `ensure-headless.sh cursor <project> <name>`
  - role fileは`db/spawn-roles/<name>.<type>.md`の規約名で自動解決される
- SessionEnd teardownでsession teamのheadless worker全員が回収される。次セッションでは必要roleをspawnし直す(config永続なので同モデルで立つ)
- **despawn前にin-flight dispatchを棚卸しする**。同名workerを後から再spawnすると旧dispatchが再駆動される(2026-08-01実例)
- turn実行中のworkerをkill/teardownするとそのturnは丸ごと失われ、誰も再駆動しない(respawnはモデル側threadを引き継がない)
- 実装subagentの生存判定に既読フラグやログのmtimeを使わない。Agent toolの完了通知が唯一の完了シグナル

## 送信の実務

- **codex系宛は必ず先に** `~/.agents/skills/agmsg/scripts/ensure-codex.sh <project> [worker名]`(起動済みならno-op)。怠ると依頼は未読のままDBに滞留し返信が来ない(2026-07-05実例)
- **`ensure-codex.sh`/`spawn.sh`は単独のsimple commandで呼ぶ**。セッションUUIDは`CLAUDE_CODE_SESSION_ID=<リテラルUUID>`と直書きする(シェル変数にするとsandbox除外が効かず、codexが自分のseatbeltを張れずspawnが失敗する)。for/`;`/`&&`/パイプ/コマンド置換の中に入れない
- **本文は例外なくファイルに組み立てて `send.sh <team> <from> <to> --stdin < packet.txt`**。4番目の引数に直接書かない。本文にバッククォートが1つでもあるとコマンド置換が発火し、**その部分が欠落したまま送信が成功する**(2026-08-14に1セッションで3回。短文でも起きる)
- 送信後は自分の送信文面をDBからgrepし、重要な文字列(コマンド名・設定値・file:line)が残っているか確認する
- **既読フラグは生存判定にも沈黙判定にも使えない**(2026-08-14実証)。codex系workerはturn実行中に新規メッセージを読まないため、作業中のworker宛は未読が溜まって当然。逆に既読でも送信先を間違えていれば返信は来ない
- **agmsg workerが無言のときの一次診断は bridge ログ(`run/<type>-bridge.<team>.<name>.log`)の mtime と最終行**。10分以上更新が無ければ異常を疑う。**ログのlifecycle行はbridge死亡時に凍るので、それだけでbusyと判断しない — `.meta`の`pid=`を`ps -p`で照合する**
- **worker側の送信先typoは silent fail する**(2026-08-14実例)。無言のときはteamを絞らず`from_agent`だけで横断検索する
- Monitor通知は長い返信を切り詰める。全文はDBから読む
- worker/subagentの出力そのものは転載せず、採用した内容と最終的な変更のみを報告する

## パケットの鉄則

### subtaskパケット(メイン→実装subagent)

自己完結が絶対条件(subagentはパケット本文とrepoしか見ない。メインの会話文脈は渡らない): `[task:<id>]`+`[subtask:<id>]` / ゴール / 制約 / **スコープのファイルセット** / 受入条件 / 検証コマンド / 報告書式。

DO NOTを明記: git commit/push禁止・ファイルセット外の変更禁止・無断の設計変更禁止・未検証の完了報告禁止・**外部write禁止(GitHub issue/PR/comment/reviewの作成・変更、gh POST/PATCH/DELETE、その他の対外mutation)**・**repo外read禁止と機密情報read禁止**。例外はユーザーの明示指示をパケットが引用している場合だけで、その場合も対象の操作を限定して書く。

**これらは契約文でのみ抑止される**。subagentは親セッションの権限をそのまま継承し、`claude/settings.json`のallowに`Bash(gh api:*)`等が入っているため、外部writeは技術的には無プロンプトで通る。codex workerのimplementer layoutが持っていたnetwork遮断・repo外read禁止に相当する強制境界は無い(2026-09-06、ユーザーがフック導入を見送ったうえでの受容)。remote mutationは**事後検知もできない**。

**質問プロトコル**: subagentは実行中にメインへ問い合わせられない。判断が要る場面では**変更を加えずに**`question`として完了報告を返す。書き込み後に判断点が見つかった場合は、**自己判断でrollbackさせない**(共有worktreeなので他の変更を壊す)。partial changeを変更ファイル一覧に明記させ、扱いはメインが決める。

**完了報告の必須field**: 結果種別(`done|blocked|question`) / 変更ファイル一覧(repo相対パス) / 検証(実行したコマンドと実際の出力)。`question`・`blocked`では何が決められないかと選択肢・各々の帰結。

**外部情報が要る作業**: subagentはnetworkを使えるが、委譲先で叩けない外部CLI/APIについては**実出力サンプルをパケットに貼る**(2026-08-12 herdr-titleで確立)。推測で実装しfakeでテストすると**テストは全passするのに実機で動かない**。実例: `herdr tab list`のtabsは`result.tabs`配下だがworkerはトップレベルと推測し、`json.Unmarshal`が成功して空スライスを返し機能が黙って無効化された。併せて「認識できない形式はエラーにする(errなし空を返さない)」をテスト要件に含める。

**環境依存failを含むテストスイート**は、委譲前にベースラインのfail一覧を採取して`baseline-failures.txt`に置き、ゲートを「対象モジュール全green+`comm -13 baseline after`で新規failゼロ」にする。

### [research]パケット(codex-research / grok-research宛)

- codex-researchは調査専任(reviewer layout: repoはread、書けるのはagmsg配下のみ。**networkは全開**)。返すのはfile:line一覧や構造化データだけで、**パッチは作らせない**
- 自己完結パケット = GOAL / CONSTRAINTS / SCOPE / SCHEMA(期待する出力の構造) / 検品観点 / DO NOT write files・DO NOT パッチ生成
- **grok-researchの併走は公開情報かredacted packetで閉じる問いに限る**。認証・秘密情報・金銭・不可逆操作を扱う調査、security boundary・データ喪失・課金・広範囲migrationの採否を決める調査は出さない(判定基準の正本はグローバル`CLAUDE.md`)。判定はdispatch前に行い根拠を`[task:<id>]`へ記録し、先行結論をblindに渡さない
- 返答は**検品**する。SCHEMA未充足・情報不足は「◯件中◯件で△△が不足」と対象を具体に指摘して差し戻す
- **「バグ/異常を発見した」系の断定は再現条件まで確認してから採用する**(2026-07-27実例: `kustomize build <base>/api`単体での「既知バグ」報告が、親overlayからのビルドでは正しく解決され実機も正常だった)。SCHEMA充足の検品(形式)とは別に、断定の再現性の検品(内容)が必要
- **インクリメンタル調査**: 1トピック=1パケット。各単位を検品してから次へ

### [review]パケット(codex / fable-review宛。段9のフォールバック時はcodexをgrok-reviewに読み替える)

- いずれもread-only。findingsを返すだけで、**fixはメインが適用**する(subagent作hunkのfixはSendMessageで当該subagentへ差し戻す)
- 自己完結パケット = `git diff`か対象`file:line`(プラン査読の場合はプラン本文) + 意図 + (ループ時)前回指摘→対応の対応表
- **段2のプラン査読packetは段1の4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須fieldとして含む**。fable-reviewは欠落・空だけでなく**定型的で実質のない値**もfindingにする(「疑う前提: なし」「反対案: 現案維持」で通さない。該当なしには理由を要求する)。再送(反復査読)では「最初の未査読プラン」ではなく「前回findingsへの対応」として評価する
- **段3(codex)のプラン査読packetは、段2を実施した場合、段2findingsのfinding ID・振り分け(採用/見送り/別タスク)・見送りと別タスクは理由を一対一で列挙する**(段2省略時は「段2省略(redactすると議論不成立、依頼送信前に決定)」と明記。認証・秘密情報を含むfindingは理由を`redacted(理由: 認証/秘密)`で代替可、packet不備に当たらない。redactedを使ったfindingは、内容(理由の具体)を平文開示せずfinding ID・重大度・振り分け(採用/見送り/別タスク)・解決状態(対応済み/未対応)を段4のユーザー承認時と段11の検収報告の両方に明記する)。欠落は順序違反の可視化点として扱う
- 出力形式: Findings / Required tests / Residual risk / Confidence。severity順・推測は明記・各findingにIDを付す
- 段9では**codex(フォールバック中はgrok-review)**に、**subtask別のdiff identity・依存関係・subagentの検証結果**・**承認済みプラン本文**・**dependency advisoryの確認結果と根拠(メイン記入。欠落時はcommit不可)**と**段8のauthor再分類マップ(メイン作 / subagent作の2区分)**を渡す。**マップの用途はfindingの差し戻し先の判定で、査読対象は常に全hunk**(マップが無いとラベルを推測で付け、無実のsubtaskが差し戻される)

## レビュー収束条件

指摘の反映(再修正)後の差分は再依頼しうるが、**1回で止めるな・延々と回すな**。

- 残るfindingsが Low / nit / 見送り(理由明記) / 別タスク(スコープ外)だけで、substantive(正しさ・設計・回帰)が無い
- 全findingを 採用 / 見送り / 別タスク に振り分けて反映・判断済み
- **再依頼はsubstantiveなfindingが出た巡だけ**。目安は最大2巡、超えるなら残課題を別タスク化して打ち切る
- 各巡で「前回指摘→対応」の対応表をパケットに含める。無返信のときは無限に待たず、bridgeログからfindingsを読んで内容ベースで収束判断する
- **同じ箇所で3巡以上続くときは、修正が「一般化した」つもりの決め打ちになっていないか疑う**。2026-09-01の実例: DST時刻の扱いで「spring-forwardの1時間を拒否」→「fall-backの1時間を両方受理」→「移行幅を仮定せずoffsetから導出」と3巡し、すべて『1時間の移行しか無い』という同じ暗黙の定数が原因だった。修正のたびに「この定数・刻み幅・件数の仮定はどこから来たか」を1行で言語化し、言語化できなければまだ一般化できていない

## 返信が来ない時の診断

agmsg worker(査読・調査役)向け。bridgeログ確認・stale pidfile・CLI更新後のdespawn→再spawn等、診断手順は `references/troubleshooting.md`。モデル/effortのper-workerキーと既知エラー対処は `references/model-routing.md`。実装subagentはagmsgを経由しないのでこれらの対象外で、無言のときはAgent toolの完了通知と`/tasks`を見る。

## 関連

- `/agmsg` — inbox確認・送信・履歴
- `/commit` — 変更の確定
- グローバル`CLAUDE.md`「エージェント役割分担」— 役割・権限・フロー全体の正本
