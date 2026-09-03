---
name: orchestrate-agents
description: 全タスク共通の単一委譲フロー(対話でのプラン起案→fable-review設計ゲート→codexプラン査読→ユーザー承認→manager分割→worker実装→watcher完了監視→メイン統合→コード査読→commit)の運用手順。agmsgの非同期send+Monitor自動再開で回す
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

**このゲートを通る前に、manager/worker宛の実装パケットを1件も出さない。** 承認対象はサブタスク方針を含むプラン全体。

### 段5 manager(分割・発注)

`ensure-codex.sh <project> manager`で起動し、`[task:<id>]`付きの自己完結パケット(ゴール/制約/完了条件/検収観点)を送る。managerは受入条件を**同文で**worker と watcher の両方へ配る。

分割後のsubtaskは、メインが**承認済みプランの範囲内**であることを確認して初めてrepo writeが有効になる。範囲外ならメインが差し戻し、必要ならユーザーへ再承認を取る。

**発注ゲート**: managerは発注前に`[scope-check]`(分割一覧+各subtaskのファイルセット)をメインへ出し、メインが承認済みプランの範囲内と確認した分だけ`[scope-ok]`を返す。managerのdispatchパケットには`scope-ok:<task-id>/<subtask-id>`トークン(subtask単位)が入り、これがworkerのrepo write許可の根拠になる(自分のsubtask idと一致するトークンが無ければworkerは書かずに差し戻す。task単位のトークンでは、保留したsubtaskや後から発明されたsubtaskを止められない)。watcher宛のコピーには受入条件に加えて**ファイルセットと基準fingerprint**を入れる。

**fingerprintの算出法(ここが定義の1箇所)**: workerはcommitできないのでcommit hashは使えない。基準・提出とも次の2つを組で使う。
- 変更有無: `git -C <repo> status --porcelain -- <そのsubtaskのファイルセット>`の出力そのもの(**必ずファイルセットに絞る**。絞らないと並行subtaskの作業中変更が混ざり、帰属不能な不一致でstallする)
- 内容: ファイルセットの各パスについて、存在すれば`git -C <repo> hash-object <path>`、存在しなければリテラル`absent`。**存在しないパスに`hash-object`を実行しない**(fatalになる。新規作成予定のパスはdispatch時が`absent`、削除するパスは提出時が`absent`、renameは旧パス`absent`+新パスhashで表れる)
`shasum`は使わない(worker sandboxでlibperl.dylibの読み込みを拒否され失敗する。2026-08-21実測)。ハッシュ関数が別途必要なら`openssl dgst -sha256`。

**fingerprintを取る間、メインは同じrepoを編集しない**(2026-08-21実測): メインが並行編集していると`status --porcelain`と`hash-object`の値が数分おきに変わり、workerの作業前後ペアが必ず不一致になる。watcherは契約どおり差し戻すので、原因はworkerでなく段取りにある。メイン側の編集が続く間はdispatchしない、または対象repoを分ける。

**段9のfindings対応でメインが編集した直後にworkerのcycleが閉じると、fingerprintは必ず不一致になる**(2026-08-30実測)。上の「メインは編集しない」は段5〜7の話で、段10でメイン作hunkのfindingを直す間はメインの編集が不可避なため回避できない。このときwatcherはbaselineと提出の差異を**観測として提示し、正否判断をせずメインへescalateする**。メインがディスク上の内容を読んで帰属を確定し、自分の変更であればworkerの逸脱として扱わない。workerに説明責任を負わせない。

managerへworkerの起動を通知するときは**agmsg登録名をそのまま書く**。driver typeと混ぜると誤配される(2026-08-21実例: 「worker-1をteamにcodexとして登録済み」と書いたのをmanagerが登録名`codex`と読み、review専任の`codex`へ実装を発注した。`codex`が実装を拒否し、managerが直接報告を`[protocol-reject]`して差し戻したので事故は止まった)。

メインは`[task:<id>]`パケットまたは`[scope-ok]`の応答で**dispatch可能なworker名を明示的に列挙する**。managerは起動状況を知らないため、列挙が無いと未起動のworker名へ発注する。agmsgの送信は宛先未登録でもsilent failし、managerが`--force`で再送しても配送されない。滞留したdispatchはDBに残り、後で同名workerをspawnすると再駆動される(「共通の不変条件」の「despawn前にin-flight dispatchを棚卸しする」と同じ機序)。実例: 2026-08-31、メインがmanager/watcher/worker-1の3体だけを起動・probeしたところ、managerがcodex-impl/hard-worker-1/worker-2へも発注し`--force`で再送した。task-abort後もそれらへの停止通知は送られず、滞留dispatchが残った。

managerは設計判断をしない。**workerの設計分岐の質問はfable-reviewへ転送し、その回答を最終決定としてworkerへ中継する**(メインの承認は挟まない)。`mechanical-only`の再分類要求とwatcherのescalationは、従来どおりメインへ転送し、メインの回答を中継する。

### 段6 worker → watcher(完全性の検査)

workerの完了報告は必ずwatcher宛。watcherが見るのは**証拠・criteria・fingerprintの完全性だけ**で、正しさの承認ではない。不足は`[criteria-query]`で1巡まで、解決しなければmanagerへescalate。

必須field: `[task:<id>]` / `[subtask:<id>]` / diff identity(ファイルセット+git hash) / author metadata(agent・model・vendor・pool) / 検証エビデンス(コマンドと結果)。

**fingerprintの算出法は段5のrecipeが唯一の定義**。そこから外れた式(HEAD OIDの併記、`shasum`)は使わない — 併記すると並行subtask時にwatcherのscoped契約と一致しなくなる。

### 段7 `[team-ready]`(非終端)

全subtaskが`[watcher-done]`になったらmanagerが`[team-ready]`を出す。ここは終端ではない。

### 段8 メインの統合とauthor再分類

メインが統合し、**統合前後のfingerprintを比較**して自分が実質変更したファイル/hunkをAnthropic authorへ再分類する。

### 段9 コード査読ゲート(author-aware)

- **脆弱性4観点(認証/認可境界・secret出力・外部write・dependency advisory)は常に`opus-review`が全hunkを対象に担当**。authorに依らずこのpassは必ず通す。**opus-reviewが応答不能な場合はgrok-reviewへフォールバックする**(下記「段9査読者フォールバックチェーン」参照)
- **意図一致・正しさの一次査読はauthorで振る**: codex worker(OpenAI)作のhunk → `opus-review`(フォールバック中は`grok-review`)、メイン(Anthropic)作のhunk → `codex`(review役)
- 混在diffでは**opus-review(またはフォールバック中はgrok-review)に全hunkを渡す**。**両方に段8のauthor再分類マップを渡す**が、マップが制限するのは意図一致査読(worker作hunkに限定)だけで、脆弱性4観点は常に全hunk対象(全hunkが単一authorでも、単一要素のマップとして明示する)
- findingのラベルは`[subtask:<id>]`(worker作)・`[author:main]`(メイン作)・`[design-level]`(承認済み設計自体の欠陥。**opus-review/grok-review・codexどちらも自分が担当したhunkから見つけたら使う**。対応する単一hunkが無ければfile:line欄に`(design-level, no single hunk)`と明記)。codexもopus-review/grok-reviewも、自分のfindingにこのラベルを付ける
- dependency advisoryは到達性を疎通確認し、取得できない場合はpassではなく`not checked`と根拠を返させる。cursorのopus-review/grok-reviewはShell denyのためこの確認が構造的にできない。**依存を変えるdiffのadvisory確認はメインの明示責務**とし、段9パケットの必須fieldに確認結果と根拠を含める(欠落時はcommit不可)
- **workerがツールチェーンを持たない構成では、契約を変えない自明なコンパイルエラーはメインが直して`[author:main]`に再分類する**。enum caseの綴り違い、`return`漏れ、import漏れ、非Sendable値の隔離境界越えなど、修正が一意で挙動と公開契約を変えないもの。差し戻しの往復コストが査読価値を上回るため。直した内容は段9でcodex(review役)へ回し、workerとwatcherには「メインが直したファイルは今後workerに変更させない」と明示する
- fable-reviewは段2の設計レビュー専任で段9には関与しない

### 段9査読者フォールバックチェーン(opus-review応答不能時)

2026-08-27、opus-reviewがCursor teamのusage limitに到達し使用不能になった実例を機に導入(復旧予定日はエラー文言に依る)。

1. **主**: opus-review(cursor・`claude-opus-5-thinking-high`)
2. **フォールバック**: grok-review(cursor・config key `spawn.cursor_model.grok-review: cursor-grok-4.6-xhigh`、暫定運用)。opus-review同等の役割(意図一致査読+脆弱性4観点)を代行する
   - **grok-review(査読役)に関するreadiness照合は常に厳格基準を適用する**: 初回spawn時のreadiness照合、およびフォールバック中のopus-review復帰判定の両方で、期待token受信・エラー応答でない・pinしたmodel identityが一致、の3点を満たすことを求める。初回spawn時に不成功ならgrok-reviewは使用不能と判定し項5(段9ブロック)へ。復帰判定時に不成功なら主は未復旧と判定しフォールバックを継続する。**manager・watcher等の他workerの通常readiness照合は既存規約どおり応答到達で足りる**(厳格化はopus-review/grok-reviewの査読役に限る)
   - role file(`grok-review.cursor.md`)にreadiness probeへの明示的な応答例外を記載済み(「standing roleは査読リクエストに対するもので、trivialなprobeには短い応答で答えてよい」)。この例外を明記しない限りmodelがstanding roleを厳密に守った回に不成功判定になりうるため、role file変更で対応した(SKILL.mdの記述だけでは担保できない)
   - **レビュー本体の成功基準**: 「[review]パケットの鉄則」の出力形式(Findings/Required tests/Residual risk/Confidence。clean reviewはFindingsなしを含む)を主が返すこと。これを返さない場合はspawn/probe時エラー・査読中のusage-limit応答・timeout・無応答・部分応答・形式不正・異常終了のいずれであってもすべて不成功とみなし、フォールバックへ切替。**切替時は部分査読を継ぎ足さず段9を最初からやり直す**(遅着した主の応答とフォールバック応答を二重採用しない)
   - **停止条件**: grok-reviewの誤陰性・査読形式不良を1件でも観測した時点でフォールバックを停止する。停止状態はセッションをまたいで有効とし(このファイルまたはCLAUDE.mdへの追記で記録)、定義した再有効化手順(ユーザー承認を得て本節を書き換える)以外では再開しない
3. **フォールバック不能クラス**: 最終outbound payload全体(diff本文+補足説明+ログ+添付コンテキスト)に秘密の実値(credential・token等)が含まれると判断される場合。機械的secret scannerは実装せず、既存opus-review運用と同水準のメイン目視確認に依る(誤判定は残存リスクとして受容)。該当する場合、主・フォールバックいずれもcursor系のため対象外
4. **段9ブロック(waiver可)**: **3のフォールバック不能クラスに該当する場合に限る**(cursor系全体が対象外で主・フォールバックどちらも使えないケース)。**ユーザーの明示的waiverがある場合のみ例外**とし、waiverは次の3条件を満たす: (a)対象diffをrepository identity・base tree・commit対象の完全なstaged treeを含むfingerprintに固定し、commit直前に再照合する (b)「段9成功」ではなく「段9未実施・ユーザーwaiver」として検収記録に残す(段9パスとして扱わない) (c)waiver後に対象diffが変更されたら旧waiverは無効・再承認必須
5. **段9ブロック(waiver不可)**: 3に該当せず(秘密の実値を含まない)、かつ主・フォールバック双方が使用不能な場合。段9を無言スキップせずタスクをブロックしてユーザーに報告。waiverの対象ではない(通常の二重障害はブロック一択)

外部送信の安全境界: cursor harness経由の査読役(opus-review・grok-review)へは、ユーザーが受容したデータ境界としてprivate diff送付を許容する(2026-08-27ユーザー確認済み)。調査役(grok-research等)への最小化義務(未公開コード断片を含めない)とは別軸。対象はCLAUDE.mdが適用される全project、個別repositoryのローカル外部送信禁止規約があればそれを優先、ユーザーが受容を撤回した時点で即時無効。

**grok-reviewとgrok-researchは名前が1語しか違わないが権限が異なる別worker**: grok-review(査読役、private diff送付許容)とgrok-research(調査役、最小化義務でsecret・未公開コード断片禁止)を混同しない。`[review]`パケットの宛先は必ず`grok-review`。フォールバック発生直後はgrok-researchが既にready(返信実績あり)でもgrok-reviewが未spawn/未probeの窓が生じうるため、この窓でprivate diffをgrok-researchへ誤送しない — dispatch前に宛先名を確認する。role fileの新規追加(`claude/agmsg-roles/grok-review.cursor.md`)はCLAUDE.md/SKILL.mdの変更と同一commitに含める(role file無しでのspawnはcredential非読取・probe例外等の規約が乗らない)。

未解決のまま残す前提: usage limitの機序(モデル別クォータかteam契約構造か)は未検証。Cursor外のAnthropicフォールバック(headless claude-code経由の別ロール)は将来課題として保留。Grok4.6の4観点較正はコマンドインジェクション検出1サンプルのみ(恒久品質保証ではない)。

### 段10 差し戻し(3経路)

- **worker作のfinding** → managerへ`[team-reopened]`(findingと対象`[subtask:<id>]`を明記)。managerは該当subtaskだけ再オープンし、**基準fingerprintを取り直してworkerとwatcherの両方へ再配布**してから段6へ(watcherへ再送しないと、古い基準との照合で2回目の`[watcher-done]`に永久に到達しない)
- **メイン作hunkのfinding** → メインが直し、段8のfingerprint再計算 → 段9へ再投入
- **設計レベルのfinding**(`[design-level]`ラベル。承認済み設計自体の欠陥が段9で判明した場合) → メインが`[task-abort]`で理由を明記して該当taskを終端(`[task-aborted]`) → 新規`[task:<id>]`を発行して段1から再起動(旧taskとの関連は理由欄で相互参照)。既存taskの延命はしない

### 段11 `[team-done]`→検収→commit

段9のfindingが全て解消したら**メインがmanagerへ`[findings-resolved]`を送り**、managerはそれを受けて`[team-done]`を**1完了サイクルに1回だけ**出す(自発的には出さない)。メインが実物を検収し、ユーザー確認後にcommitする。検収失敗・ユーザー差し戻しは段10の該当経路へ戻り、解消後に再び`[team-done]`。終端は3つだけ: **メインのcommit** / **ユーザーの中止** / **`[task-aborted]`**(調査や実装の結果「変更不要」と確定した場合、実行不能と確定した場合、段取りの誤りで測定できない場合。メインが理由を明記してmanagerへ送り、managerはsubtask破棄とworker/watcherへの通知を済ませて`[task-aborted]`を1通返す。`[team-ready]`も`[team-done]`も出さない)。

検収の中身: `git -C <repo> status` / `git diff`で(1)要件適合・プラン逸脱 (2)正しさ・エッジ・回帰リスク を読む。`git log`で勝手commitがないことを確認。報告された検証を1-2コマンドでスポット再現。

## タグの producer / consumer

| タグ | 出す | 受ける | 意味 |
|---|---|---|---|
| `[scope-check]` | manager | メイン | 分割一覧の範囲確認依頼(発注前) |
| `[scope-ok]` | メイン | manager | 承認済みプランの範囲内と確認した subtask の列挙 |
| `[criteria-query]` | watcher | worker | 完全性の不足の指摘(1巡まで) |
| `[watcher-done]` | watcher | manager | 完全性の検査を満たした(正しさの承認ではない) |
| `[team-ready]` | manager | メイン | 全subtaskが`[watcher-done]`(非終端) |
| `[team-reopened]` | メイン | manager | worker作findingの差し戻し(対象subtask明記) |
| `[findings-resolved]` | メイン | manager | 段9の全finding解消。`[team-done]`の唯一のトリガー |
| `[team-done]` | manager | メイン | 完了宣言。1完了サイクルに1回 |
| `[task-abort]` | メイン | manager | 中止要求。dispatch前・dispatch後・`[team-ready]`後のいつでも出せる |
| `[task-aborted]` | manager | メイン | 中止の終端(subtask破棄とworker/watcherへの停止通知を済ませたことの報告)。開いているサイクルもこれで閉じる |
| `[author:main]` | opus-review / codex | メイン | findingの帰属ラベル。メインが直す(managerへは渡さない) |
| `[design-level]` | opus-review / codex | メイン | 承認済み設計自体の欠陥(担当hunkから発見時)。file:lineが無い場合は`(design-level, no single hunk)`と明記。段10の第3経路(task-abort→新task発行)へ |

完了サイクルは**dispatch(段5の発注、またはreopen時の再発注)で開き**、`[team-done]`・`[team-reopened]`・`[task-aborted]`のいずれか1つで閉じる。`[team-ready]`は開いているサイクル内のマイルストーンで境界ではない。したがって`[team-done]`は1サイクルに最大1回で、差し戻しを挟んだ2回目はそのreopenが開いたサイクルの1回目にあたる。メイン作hunkのfindingは`[author:main]`ラベルで扱い、workerのsubtaskを誤って再オープンしない。

## 共通の不変条件

- `[task:<id>]`はセッション内で一意。`[subtask:<id>]`はtask内で一意
- 完了済み・破棄済みタスクのstale/duplicate messageは既読化して無視する
- **readiness照合**は「名前がある」ではなく、各nameのregistrationが**期待typeでちょうど1件、かつ当該セッションで返信実績があること**(dead-letterはregistration照合だけでは検出できない)。**新規spawn直後で返信実績がまだ無い場合はtrivialなprobeパケットを1通送り、その応答到達をもって返信実績とする**(probeは通常のtask dispatchとして数えない。循環依存を避けるための最小手順)。**respawn(despawn→再spawn)した場合、respawn前の返信実績は無効**として扱い、必ずrespawn後に新規probeを送り直す(旧instanceの応答をもって新instanceをreadyと誤判定しない)。対象はmanager・watcher・使用する全worker・fable-review・opus-review・codex(review役)、および実際にdispatchするcodex-research/grok-research(起動コマンドが一覧にあることは照合の代わりにならない)。**grok-reviewはopus-reviewフォールバック時のみdispatchするため、フォールバックが発生した回に限りreadiness照合の対象に加える**
- 起動コマンド(モデル等のconfigはグローバル永続なのでコマンドのみ):
  - codex系(manager, watcher, codex-impl, worker-1/2, hard-worker-1, codex, codex-research): `ensure-codex.sh <project> <name>`
  - claude-code系(fable-review): `AGMSG_CLAUDE_PROBE_TIMEOUT=150 ensure-headless.sh claude-code <project> fable-review`(既定30秒ではfableがprobeを出し切れずrc=124でfail-closedする。probe timeoutはper-nameのconfigキーが無くenv varのみ。model/effort/turn timeoutはper-nameのconfigキーで固定。既定300秒では設計レビューに不足するため`spawn.claude_turn_timeout.fable-review: 1800`を設定済み。上記段2参照)
  - cursor系(opus-review): `AGMSG_CURSOR_BRIDGE_TURN_TIMEOUT=1800 ensure-headless.sh cursor <project> opus-review`(既定180秒ではopus:highの査読が切れる)
  - cursor系(grok-review、opus-reviewフォールバック時のみ): `AGMSG_CURSOR_BRIDGE_TURN_TIMEOUT=1800 ensure-headless.sh cursor <project> grok-review`
  - cursor系(grok-research): `ensure-headless.sh cursor <project> <name>`
  - role fileは`db/spawn-roles/<name>.<type>.md`の規約名で自動解決される
- SessionEnd teardownでsession teamのheadless worker全員が回収される。次セッションでは必要roleをspawnし直す(config永続なので同モデルで立つ)
- **despawn前にin-flight dispatchを棚卸しする**。同名workerを後から再spawnすると旧dispatchが再駆動される(2026-08-01実例: 解体済みチームのサブタスクが再spawn後のworkerで蘇生し、幽霊タスクにcycleを浪費した)
- turn実行中のworkerをkill/teardownするとそのturnは丸ごと失われ、誰も再駆動しない(respawnはモデル側threadを引き継がない)
- worker/コンサルの生存判定に既読フラグを使わない(下記「送信の実務」)

## 送信の実務

- **codex系宛は必ず先に** `~/.agents/skills/agmsg/scripts/ensure-codex.sh <project> [worker名]`(起動済みならno-op)。怠ると依頼は未読のままDBに滞留し返信が来ない(2026-07-05実例)
- **`ensure-codex.sh`/`spawn.sh`は単独のsimple commandで呼ぶ**。セッションUUIDは`CLAUDE_CODE_SESSION_ID=<リテラルUUID>`と直書きする(シェル変数にするとsandbox除外が効かず、codexが自分のseatbeltを張れずspawnが失敗する)。for/`;`/`&&`/パイプ/コマンド置換の中に入れない
- **本文は例外なくファイルに組み立てて `send.sh <team> <from> <to> --stdin < packet.txt`**。4番目の引数に直接書かない。本文にバッククォートが1つでもあるとコマンド置換が発火し、**その部分が欠落したまま送信が成功する**(2026-08-14に1セッションで3回。短文でも起きる)
- 送信後は自分の送信文面をDBからgrepし、重要な文字列(コマンド名・設定値・file:line)が残っているか確認する
- **既読フラグは生存判定にも沈黙判定にも使えない**(2026-08-14実証)。codex系workerはturn実行中に新規メッセージを読まないため、作業中のworker宛は未読が溜まって当然。逆に既読でも送信先を間違えていれば返信は来ない
- **workerが無言のときの一次診断は bridge ログ(`run/<type>-bridge.<team>.<name>.log`)の mtime と最終行**。10分以上更新が無ければ異常を疑う。**ログのlifecycle行はbridge死亡時に凍るので、それだけでbusyと判断しない — `.meta`の`pid=`を`ps -p`で照合する**
- **worker側の送信先typoは silent fail する**(2026-08-14実例)。無言のときはteamを絞らず`from_agent`だけで横断検索する
- Monitor通知は長い返信を切り詰める。全文はDBから読む
- workerの出力そのものは転載せず、採用した内容と最終的な変更のみを報告する

## パケットの鉄則

### subtaskパケット(manager→worker)

自己完結が絶対条件(workerはパケット本文とrepoしか見ない): `[task:<id>]`+`[subtask:<id>]` / ゴール / 制約 / **スコープのファイルセット** / 受入条件(watcherへ配るものと同文) / 検証コマンド / 報告書式(必須fieldは段6) / 質問プロトコル(迷ったら止めてmanagerへ質問)。

DO NOTを明記: git commit/push禁止・ファイルセット外の変更禁止・無断の設計変更禁止・未検証の完了報告禁止・**外部write禁止(GitHub issue/PR/comment/reviewの作成・変更、gh POST/PATCH/DELETE、その他の対外mutation)**。implementer layoutのnetwork遮断に頼らず契約側でも塞ぐ。例外はユーザーの明示指示をパケットが引用している場合だけで、その場合も対象の操作を限定して書く。

`mechanical-only`パケットの宛先は**codex-implだけ**(停止・再分類の契約をrole fileに持つのがcodex-implのみ)。他のworkerが受け取ったら再分類要求を返して着手しない。

`mechanical-only`と書けるのは、挙動・公開契約・認証・security boundary・課金・外部write・依存関係を変えない、correct-by-inspectionな編集だけ。workerは判断が要る場面に当たったら再分類を要求して止まる。

### network offワーカーへ実装を委譲する前の準備

実装worker(implementer layout)はnetworkが遮断されている。外部情報が要る作業は、必要な情報をパケットへ同梱するか、事前にcodex-research(network全開)で収集して渡す。

- **委譲先で叩けない外部CLI/APIは、実出力サンプルをパケットに貼る**(2026-08-12 herdr-titleで確立)。推測で実装しfakeでテストすると**テストは全passするのに実機で動かない**。実例: `herdr tab list`のtabsは`result.tabs`配下だがworkerはトップレベルと推測し、`json.Unmarshal`が成功して空スライスを返し機能が黙って無効化された。併せて「認識できない形式はエラーにする(errなし空を返さない)」をテスト要件に含める
- **Go**: `spawn.codex_extra_fs_roots`で`~/go/pkg/mod=read`(+`~/Library/Caches/go-build=write`)が付与済みなら、depsがhostのmodule cacheに揃う限りvendor焼き込みは不要。cacheに無い依存・キー未設定・他言語は従来どおりrepo内に焼き込み、オフラインビルドが通ることを事前検証する。build cache writeが未付与なら`GOCACHE=${TMPDIR:-/tmp}/...`を検証コマンドに明記する
- **Rust/cargo**: `cargo vendor <出力dir> > .cargo/config.toml`で依存を焼き込む。**出力先を既存の`vendor/`にしない**(tracked path依存を上書き破壊する。herdrで実例)。toolchainがnix devShell由来なら`nix print-dev-env > dev-env.sh`を焼き出してworkerには`source dev-env.sh`させる。`CARGO_HOME=$PWD/.cargo-home`+`CARGO_NET_OFFLINE=true`
- 環境依存failを含むテストスイートは、委譲前にベースラインのfail一覧を採取して`baseline-failures.txt`に置き、ゲートを「対象モジュール全green+`comm -13 baseline after`で新規failゼロ」にする
- workerのsandboxで原理的に通らない検証(unix socket bind・$HOME配下書込み・git config読取等)は環境偽装で戦わせず、メインの非sandbox環境での再実行に切り替える
- **workerのsandboxはprojectディレクトリ外をreadできない**(add-dirで渡した別repoも含む)。別repoのファイルとの突き合わせ(byte一致検証等)を受入条件に含めるとworkerが`Operation not permitted`で停止する。該当ファイルはメインが対象repo内へ配置し、workerへはsha256等の期待値だけを渡して検証をハッシュ一致に置き換える(2026-09-01実測。配置したファイルのauthorはメインになる)
- **workerはproject外の$HOME dotfilesを読めない場合がある**(2026-07-16実例)。必要な設定値・rev・パスは最初からパケットに同梱する

### [research]パケット(codex-research / grok-research宛)

- codex-researchは調査専任(reviewer layout: repoはread、書けるのはagmsg配下のみ。**networkは全開**)。返すのはfile:line一覧や構造化データだけで、**パッチは作らせない**
- 自己完結パケット = GOAL / CONSTRAINTS / SCOPE / SCHEMA(期待する出力の構造) / 検品観点 / DO NOT write files・DO NOT パッチ生成
- **grok-researchの併走は公開情報かredacted packetで閉じる問いに限る**。認証・秘密情報・金銭・不可逆操作を扱う調査、security boundary・データ喪失・課金・広範囲migrationの採否を決める調査は出さない(判定基準の正本はグローバル`CLAUDE.md`)。判定はdispatch前に行い根拠を`[task:<id>]`へ記録し、先行結論をblindに渡さない
- 返答は**検品**する。SCHEMA未充足・情報不足は「◯件中◯件で△△が不足」と対象を具体に指摘して差し戻す
- **「バグ/異常を発見した」系の断定は再現条件まで確認してから採用する**(2026-07-27実例: `kustomize build <base>/api`単体での「既知バグ」報告が、親overlayからのビルドでは正しく解決され実機も正常だった)。SCHEMA充足の検品(形式)とは別に、断定の再現性の検品(内容)が必要
- **インクリメンタル調査**: 1トピック=1パケット。各単位を検品してから次へ

### [review]パケット(codex / fable-review / opus-review宛。フォールバック時はopus-reviewをgrok-reviewに読み替える)

- 3者ともread-only。findingsを返すだけで、**fixはメインが適用**する
- 自己完結パケット = `git diff`か対象`file:line`(プラン査読の場合はプラン本文) + 意図 + (ループ時)前回指摘→対応の対応表
- **段2のプラン査読packetは段1の4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須fieldとして含む**。fable-reviewは欠落・空だけでなく**定型的で実質のない値**もfindingにする(「疑う前提: なし」「反対案: 現案維持」で通さない。該当なしには理由を要求する)。再送(反復査読)では「最初の未査読プラン」ではなく「前回findingsへの対応」として評価する
- **段3(codex)のプラン査読packetは、段2を実施した場合、段2findingsのfinding ID・振り分け(採用/見送り/別タスク)・見送りと別タスクは理由を一対一で列挙する**(段2省略時は「段2省略(redactすると議論不成立、依頼送信前に決定)」と明記。認証・秘密情報を含むfindingは理由を`redacted(理由: 認証/秘密)`で代替可、packet不備に当たらない。redactedを使ったfindingは、内容(理由の具体)を平文開示せずfinding ID・重大度・振り分け(採用/見送り/別タスク)・解決状態(対応済み/未対応)を段4のユーザー承認時と段11の検収報告の両方に明記する)。欠落は順序違反の可視化点として扱う
- 出力形式: Findings / Required tests / Residual risk / Confidence。severity順・推測は明記・各findingにIDを付す
- 段9では**opus-review(フォールバック中はgrok-review)とcodexの両方**に、**subtask別のdiff identity・依存関係・workerの検証結果**・**承認済みプラン本文**・**dependency advisoryの確認結果と根拠(メイン記入。欠落時はcommit不可)**と**段8のauthor再分類マップ(メインが書いた/直したファイルとhunkの一覧。全hunkが単一authorでも単一要素のマップとして明示する)**を渡す。**codexの意図一致査読、およびopus-review/grok-reviewの意図一致査読(スコープ1)はマップが割り当てたhunkだけが対象**(マップが無いとラベルを推測で付け、無実のsubtaskが再オープンされる。codexは担当hunkを確定できず査読対象が空になる)。**opus-review/grok-reviewの脆弱性4観点(スコープ2)は常に全hunkが対象でマップに制限されない**

## レビュー収束条件

指摘の反映(再修正)後の差分は再依頼しうるが、**1回で止めるな・延々と回すな**。

- 残るfindingsが Low / nit / 見送り(理由明記) / 別タスク(スコープ外)だけで、substantive(正しさ・設計・回帰)が無い
- 全findingを 採用 / 見送り / 別タスク に振り分けて反映・判断済み
- **再依頼はsubstantiveなfindingが出た巡だけ**。目安は最大2巡、超えるなら残課題を別タスク化して打ち切る
- 各巡で「前回指摘→対応」の対応表をパケットに含める。無返信のときは無限に待たず、bridgeログからfindingsを読んで内容ベースで収束判断する
- **同じ箇所で3巡以上続くときは、修正が「一般化した」つもりの決め打ちになっていないか疑う**。2026-09-01の実例: DST時刻の扱いで「spring-forwardの1時間を拒否」→「fall-backの1時間を両方受理」→「移行幅を仮定せずoffsetから導出」と3巡し、すべて『1時間の移行しか無い』という同じ暗黙の定数が原因だった。修正のたびに「この定数・刻み幅・件数の仮定はどこから来たか」を1行で言語化し、言語化できなければまだ一般化できていない

## 返信が来ない時の診断

bridgeログ確認・stale pidfile・CLI更新後のdespawn→再spawn等、9項目の診断手順は `references/troubleshooting.md`。モデル/effortのper-workerキーと既知エラー対処は `references/model-routing.md`。

## 関連

- `/agmsg` — inbox確認・送信・履歴
- `/commit` — 変更の確定
- グローバル`CLAUDE.md`「エージェント役割分担」— 役割・権限・フロー全体の正本
