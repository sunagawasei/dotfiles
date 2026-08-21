---
name: orchestrate-agents
description: 全タスク共通の単一委譲フロー(対話でのプラン起案→fable-review設計ゲート→codexプラン査読→ユーザー承認→manager分割→worker実装→watcher完了監視→メイン統合→コード査読→commit)の運用手順。agmsgの非同期send+Monitor自動再開で回す
---

# 単一委譲フロー

委譲の形態は1本だけ。第二のフロー(旧「直委譲」「roleチームモード」)は持たない。役割分担と権限の正本はグローバル`CLAUDE.md`の「エージェント役割分担」節で、このファイルは**各段の実務手順・パケット書式・診断**の正本。

## 前提

- このセッションのagmsg Monitor(`watch.sh ... --team s-<このセッションのUUID>`)がSessionStartから常駐している
- 宛先workerのbridgeが稼働していること。遅延spawnは自動発火しない(下記「送信の実務」)
- 送信は非同期`send`が既定。`ask --wait`は使わない(codexでask往復が機能しない実績)

## フロー

### 段1 プランニング(メイン+ユーザーの対話)

メインがユーザーとの対話でプラン案を起案する。**この段より前に第三者のチャレンジは無い**ので、疑いはメインが自分で言語化する。プランの確定は段4のユーザー承認で、段1は起案まで。

- 段2へ渡すpacketには**4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須**で載せる。「該当なし」と書くなら理由も書く
- fable-reviewはこの4 fieldの**欠落・空・定型的で実質のない値**をfindingとして返す。自己申告の空洞化に対する検査主体はここだけなので、形だけ埋めて通さない
- 対話は依頼者のフレームに錨を下ろしやすい。**問題設定自体を疑う**役はメインと段2の両方が負う

### 段2 fable-review(設計ゲート)

プランを`fable-review`へ送る。返るのは findings。**ユーザー承認の代替にしない**。

```bash
AGMSG_CLAUDE_PROBE_TIMEOUT=180 ~/.agents/skills/agmsg/scripts/spawn.sh claude-code fable-review \
  --team <team> --project <path> --headless --reviewer
```

**probe timeoutの既定30秒では足りない**(2026-08-21実測): spawnはsandbox probeが相関ツールイベントを全部出すまで待ち、出なければ`rc=124`でfail-closedする。`~/.config`はSessionStart hookに13.5秒かかり、そこへ`fable`/xhighのturnが乗るため既定では落ちる。`AGMSG_CLAUDE_PROBE_TIMEOUT=180`を付ける。

`--reviewer`は明示する(global `spawn.claude_reviewer`のdriftでlayoutが変わらないようにするため)。reviewer layoutはrepo readを許可し、repoのBash/Edit/Writeをdenyし、agmsg storage/teams/runへのwriteを許可する — 返信は成立する。

### 段3 codex(review役)のプラン査読

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

managerへworkerの起動を通知するときは**agmsg登録名をそのまま書く**。driver typeと混ぜると誤配される(2026-08-21実例: 「worker-1をteamにcodexとして登録済み」と書いたのをmanagerが登録名`codex`と読み、review専任の`codex`へ実装を発注した。`codex`が実装を拒否し、managerが直接報告を`[protocol-reject]`して差し戻したので事故は止まった)。

managerは設計判断をしない。workerの設計分岐の質問、`mechanical-only`の再分類要求、watcherのescalationはいずれもメインへ転送し、メインの回答を中継する。

### 段6 worker → watcher(完全性の検査)

workerの完了報告は必ずwatcher宛。watcherが見るのは**証拠・criteria・fingerprintの完全性だけ**で、正しさの承認ではない。不足は`[criteria-query]`で1巡まで、解決しなければmanagerへescalate。

必須field: `[task:<id>]` / `[subtask:<id>]` / diff identity(ファイルセット+git hash) / author metadata(agent・model・vendor・pool) / 検証エビデンス(コマンドと結果)。

**fingerprintの算出法は段5のrecipeが唯一の定義**。そこから外れた式(HEAD OIDの併記、`shasum`)は使わない — 併記すると並行subtask時にwatcherのscoped契約と一致しなくなる。

### 段7 `[team-ready]`(非終端)

全subtaskが`[watcher-done]`になったらmanagerが`[team-ready]`を出す。ここは終端ではない。

### 段8 メインの統合とauthor再分類

メインが統合し、**統合前後のfingerprintを比較**して自分が実質変更したファイル/hunkをAnthropic authorへ再分類する。

### 段9 コード査読ゲート(author-aware)

- **脆弱性4観点(認証/認可境界・secret出力・外部write・dependency advisory)は常に`fable-review`**。authorに依らずこのpassは必ず通す
- **意図一致・正しさの一次査読はauthorで振る**: codex worker(OpenAI)作のhunk → `fable-review`、メイン(Anthropic)作のhunk → `codex`(review役)
- 混在diffは双方へ送り、それぞれ自分の担当hunkだけを査読する。**両方に段8のauthor再分類マップを渡す**
- findingのラベルは`[subtask:<id>]`(worker作)か`[author:main]`(メイン作)。codexもfable-reviewも、自分のfindingにこのラベルを付ける
- dependency advisoryは到達性を疎通確認し、取得できない場合はpassではなく`not checked`と根拠を返させる

### 段10 差し戻し(2経路)

- **worker作のfinding** → managerへ`[team-reopened]`(findingと対象`[subtask:<id>]`を明記)。managerは該当subtaskだけ再オープンし、**基準fingerprintを取り直してworkerとwatcherの両方へ再配布**してから段6へ(watcherへ再送しないと、古い基準との照合で2回目の`[watcher-done]`に永久に到達しない)
- **メイン作hunkのfinding** → メインが直し、段8のfingerprint再計算 → 段9へ再投入

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
| `[author:main]` | fable-review / codex | メイン | findingの帰属ラベル。メインが直す(managerへは渡さない) |

完了サイクルは**dispatch(段5の発注、またはreopen時の再発注)で開き**、`[team-done]`・`[team-reopened]`・`[task-aborted]`のいずれか1つで閉じる。`[team-ready]`は開いているサイクル内のマイルストーンで境界ではない。したがって`[team-done]`は1サイクルに最大1回で、差し戻しを挟んだ2回目はそのreopenが開いたサイクルの1回目にあたる。メイン作hunkのfindingは`[author:main]`ラベルで扱い、workerのsubtaskを誤って再オープンしない。

## 共通の不変条件

- `[task:<id>]`はセッション内で一意。`[subtask:<id>]`はtask内で一意
- 完了済み・破棄済みタスクのstale/duplicate messageは既読化して無視する
- **readiness照合**は「名前がある」ではなく、各nameのregistrationが**期待typeでちょうど1件**であること。対象はmanager・watcher・使用する全worker・fable-review・codex(review役)、および実際にdispatchするcodex-research/grok-research(起動コマンドが一覧にあることは照合の代わりにならない)
- 起動コマンド(モデル等のconfigはグローバル永続なのでコマンドのみ):
  - codex系(manager, watcher, codex-impl, worker-1/2, hard-worker-1, codex, codex-research): `ensure-codex.sh <project> <name>`
  - claude-code系(fable-review): `AGMSG_CLAUDE_PROBE_TIMEOUT=180 spawn.sh claude-code <name> --team <team> --project <path> --headless --reviewer`(probe timeoutは既定30秒では足りない。上記段2参照)
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
- **workerはproject外の$HOME dotfilesを読めない場合がある**(2026-07-16実例)。必要な設定値・rev・パスは最初からパケットに同梱する

### [research]パケット(codex-research / grok-research宛)

- codex-researchは調査専任(reviewer layout: repoはread、書けるのはagmsg配下のみ。**networkは全開**)。返すのはfile:line一覧や構造化データだけで、**パッチは作らせない**
- 自己完結パケット = GOAL / CONSTRAINTS / SCOPE / SCHEMA(期待する出力の構造) / 検品観点 / DO NOT write files・DO NOT パッチ生成
- **grok-researchの併走は公開情報かredacted packetで閉じる問いに限る**。認証・秘密情報・金銭・不可逆操作を扱う調査、security boundary・データ喪失・課金・広範囲migrationの採否を決める調査は出さない(判定基準の正本はグローバル`CLAUDE.md`)。判定はdispatch前に行い根拠を`[task:<id>]`へ記録し、先行結論をblindに渡さない
- 返答は**検品**する。SCHEMA未充足・情報不足は「◯件中◯件で△△が不足」と対象を具体に指摘して差し戻す
- **「バグ/異常を発見した」系の断定は再現条件まで確認してから採用する**(2026-07-27実例: `kustomize build <base>/api`単体での「既知バグ」報告が、親overlayからのビルドでは正しく解決され実機も正常だった)。SCHEMA充足の検品(形式)とは別に、断定の再現性の検品(内容)が必要
- **インクリメンタル調査**: 1トピック=1パケット。各単位を検品してから次へ

### [review]パケット(codex / fable-review宛)

- どちらもread-only。findingsを返すだけで、**fixはメインが適用**する
- 自己完結パケット = `git diff`か対象`file:line`(プラン査読の場合はプラン本文) + 意図 + (ループ時)前回指摘→対応の対応表
- **段2のプラン査読packetは段1の4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須fieldとして含む**。fable-reviewは欠落・空だけでなく**定型的で実質のない値**もfindingにする(「疑う前提: なし」「反対案: 現案維持」で通さない。該当なしには理由を要求する)
- 出力形式: Findings / Required tests / Residual risk / Confidence。severity順・推測は明記
- 段9では**fable-reviewとcodexの両方**に、**subtask別のdiff identity・依存関係・workerの検証結果**と**段8のauthor再分類マップ(メインが書いた/直したファイルとhunkの一覧)**を渡す。両者はマップが自分に割り当てたhunkだけを査読する(マップが無いとラベルを推測で付け、無実のsubtaskが再オープンされる。codexは担当hunkを確定できず査読対象が空になる)

## レビュー収束条件

指摘の反映(再修正)後の差分は再依頼しうるが、**1回で止めるな・延々と回すな**。

- 残るfindingsが Low / nit / 見送り(理由明記) / 別タスク(スコープ外)だけで、substantive(正しさ・設計・回帰)が無い
- 全findingを 採用 / 見送り / 別タスク に振り分けて反映・判断済み
- **再依頼はsubstantiveなfindingが出た巡だけ**。目安は最大2巡、超えるなら残課題を別タスク化して打ち切る
- 各巡で「前回指摘→対応」の対応表をパケットに含める。無返信のときは無限に待たず、bridgeログからfindingsを読んで内容ベースで収束判断する

## 返信が来ない時の診断

bridgeログ確認・stale pidfile・CLI更新後のdespawn→再spawn等、9項目の診断手順は `references/troubleshooting.md`。モデル/effortのper-workerキーと既知エラー対処は `references/model-routing.md`。

## 関連

- `/agmsg` — inbox確認・送信・履歴
- `/commit` — 変更の確定
- グローバル`CLAUDE.md`「エージェント役割分担」— 役割・権限・フロー全体の正本
