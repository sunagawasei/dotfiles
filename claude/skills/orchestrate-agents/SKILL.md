---
name: orchestrate-agents
description: codex-research(調査)・codex-impl(自走実装)・codex(オンデマンド査読)への委譲をagmsgの非同期send+Monitor自動再開で回す、Claude+codex系で完結する協業ワークフロー。壁打ち→プラン→[implement]→Q&A→検収の実装委譲フローを含む
---

# チーム協業ワークフロー

codex-research(コードベース内調査・データ収集)・codex-impl(実質的な機能実装の自走)・codex(オンデマンド査読)への委譲を、`ask --wait`(ブロック待機)ではなく非同期`send`+agmsg Monitorの自動再開で回す。2026-07-12〜、実質的な実装はcodex-implへ委譲し、メイン(Fable)は起案者として質問対応と検収を担う。同日cursorはオーケストレーションから外れた(調査はcodex-researchへ。以後Claude+codex系で完結)。役割分担の全体はグローバルCLAUDE.mdの「エージェント役割分担」。

## 概要

非同期`send`+agmsg Monitor自動再開は、codex系ワーカー(codex-research/codex-impl/codex)への依頼の既定の送信方式(2026-07-05〜、単発でも同じ)。このスキルは、複数の独立したタスクを同時に投げて並行させたい場面のための分解・委譲テンプレートで、CLAUDE.mdの役割分担・パケット書式・検証ゲートは一切変えない。送ったらターンを終え、返信はこのセッションが最初から起動しているagmsg Monitor(SessionStartで自動起動する`watch.sh`、session_teamの`s-<UUID>`宛)が拾って通知し、ターンを自動再開する。

調査役(codex-research)はパッチを作らない([agmsg-demo-fable5-x-constellation](https://github.com/fujibee/agmsg-demo-fable5-x-constellation)のデモのgrok役に相当)。実質的な実装はcodex-implが書き、Claudeはプラン・質問応答・検収を担う(下記「codex-impl自走実装ワークフロー」)。cursorは2026-07-12にオーケストレーションから外した(agmsgのcursor統合自体は機能として残存)。

## 使うタイミング

- ユーザーが「チームで協業して」「並行でやって」など、明示的にこのモードを求めたとき
- 複数の独立したサブタスク(例: codex-researchへの調査依頼と、それとは別系統の実装・確認作業)を同時に進められるとき
- 実質的な機能実装の委譲(codex-impl)は標準フロー(ユーザーの明示不要)。本skillの「codex-impl自走実装ワークフロー」に従う

単発で1エージェントに1回聞くだけなら、このスキルの分解テンプレートは不要(送信方式は同じ非同期`send`+Monitor)。`ask --wait`(ブロック待機)は使わない(codexでask往復が機能しない実績があり、cursor離脱で実質廃止)。

## 前提

- このセッションのagmsg Monitor(`watch.sh ... --team s-<このセッションのUUID>`)がSessionStartから常駐している(追加設定不要)。
- codex系ワーカーは、既存のCLAUDE.md運用と同じくこのセッションのteamに参加済みであること。
- 宛先workerのbridgeプロセスが稼働していること。遅延spawnは自動では発火しない(下記手順2で確認・起動する)。
- CLAUDE.mdの役割分担(codex-research=コードベース内調査・データ収集・パッチ不可 / codex-impl=実質的実装の自走・対象repoにwrite・commit禁止 / codex=オンデマンド査読・read-only / Claude=起案者・検収者・適用/commitの唯一の主体)は変更しない。

## ワークフロー

### 1. タスクを独立した単位に分解する

並行させられる単位に分ける。依存関係がある場合(例: プラン・実装委譲はcodex-researchの調査結果が検品を通った後、検収はcodex-implの[done]受領後)は、その順序を崩さない — 依存先は該当タスクの完了通知を受けてから送る。

### 2. 非同期でパケットを送る(ask ではなく send)

パケットの中身は下記「依頼パケットの鉄則」と同じ自己完結フォーマットを使う。送信は`--wait`を付けない素の`send.sh`で行うが、**送る前に宛先のbridgeが起きているか確認する**:

- **codex系宛(codex/codex-research/codex-impl共通)**: 必ず先に `~/.agents/skills/agmsg/scripts/ensure-codex.sh <project> [worker名]` を実行して遅延spawnを発火させる(起動済みならno-op。worker名省略時は`codex`。role fileは規約名`db/spawn-roles/<worker名>.codex.md`が自動解決される。`CLAUDE_CODE_SESSION_ID`が環境に無ければセッションUUIDを明示して渡す)。`send.sh`自体は一方向送信やctrl系がdespawn済みworkerを蘇生させないよう、意図的にこの入口へ配線されていない
- 怠ると依頼は未読のままDBに滞留し、返信が永遠に来ない(実例: 2026-07-05)
- **送信後は既読(read_at)を確認してから「依頼中」と報告する**: `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT id, read_at IS NOT NULL FROM messages WHERE team='<team>' AND from_agent='claude' ORDER BY id DESC LIMIT 1;"` をスポット実行。send 自体は bridge が死んでいても成功するため、既読確認なしの「依頼済み」報告は空振りに気づけない(実例: 2026-07-11、bridge死亡で未読滞留のままユーザーに指摘された)。数分待って未読なら despawn --force → ensure-codex で bridge を入れ替える(未読は自動再処理される)

```bash
# codex-researchへの調査依頼(ブロックしない、prefixは [research])
~/.agents/skills/agmsg/scripts/send.sh $TEAM $AGENT codex-research "[research] GOAL: ... / CONSTRAINTS: ... / SCOPE: ... / SCHEMA: ... / 検品観点: ..."

# codexへのオンデマンド査読依頼(ブロックしない、prefixは [review])
~/.agents/skills/agmsg/scripts/send.sh $TEAM $AGENT codex "[review] <git diff or file:line> / 意図: ... / 前回指摘→対応: ..."
```

diff同梱など長文・quoting事故が起きやすいパケットは、本文を4番目の引数で渡さず、ファイルに組み立ててから `send.sh <team> <from> <to> --stdin < packet.txt` で標準入力から渡す(実例: 2026-07-05のcodexレビュー依頼)。

`$TEAM`/`$AGENT`はこのセッションの既知の値(agmsgスキルのIdentityで確認済みのもの)を使う。

### 3. ターンを終えるか、他の独立作業を続ける

送信直後にやれる別の独立作業があれば進める。なければそのままターンを終えてユーザーに制御を返してよい — ブロックする必要はない。返信はMonitor通知で自動的に拾われる。

### 4. 返信が来たら検品・実装・検証する

Monitor通知でcodex系ワーカーからの返信を受け取ったら、下記「依頼パケットの鉄則」のゲートをそのまま適用する:

Monitor通知は長い返信を途中で切り詰める。全文は `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT body FROM messages WHERE team='<team>' AND from_agent='<agent>' ORDER BY id DESC LIMIT 1;"` で読む(bridgeログより確実)。

- **codex-researchの調査結果**: 「[research]パケットの鉄則」に従って検品する。不足があれば具体的に指摘して差し戻し、検品を通ったら、その調査結果を起点にプラン・[implement]パケットを組む
- **codexのfindings**: 「レビュー収束条件」に従って採用/見送り/別タスク化を判断

複数の返信が並行して届いた場合は、それぞれ独立に処理する(依存関係があるものだけ順序を守る)。

### 5. 収束したらユーザーに要約報告

codex系ワーカーの出力そのものは転載せず、採用した内容と最終的な変更のみを報告する(CLAUDE.mdの既存方針どおり)。

## codex-impl自走実装ワークフロー（2026-07-12〜）

実質的な機能実装(新機能・refactor・複数ファイル変更)はcodex-implに自走させ、メイン(Fable)は起案者として質問対応と検収を担う。codex-implはimplementer layout(cwd=対象repo+workspace-write=repo書き込み可・network off)のheadless codexワーカー(gpt-5.6-sol)。些細な編集(1行config・typo等)はこのフローに乗せない(従来どおりsonnet/メイン直接)。

### 1. 壁打ち→プラン確定

ユーザーと要件を壁打ちし、プランを固める(分岐点はAskUserQuestion)。プランは[implement]パケットにそのまま載る粒度で: 背景・要件・制約・完了条件・検証手順(実行すべきテスト/ビルド)。

### 2. codex-implの確認・起動

```bash
pgrep -f "codex-bridge.*<team>.*codex-impl" >/dev/null || \
  CLAUDE_CODE_SESSION_ID=<session-uuid> ~/.agents/skills/agmsg/scripts/ensure-codex.sh <対象repo> codex-impl
```

configの`spawn.codex_implementer.codex-impl: true`によりensure-codex.sh経由でもimplementer layoutが適用され、role fileは規約名`db/spawn-roles/codex-impl.codex.md`が自動解決される。明示spawnする場合:

```bash
~/.agents/skills/agmsg/scripts/spawn.sh codex codex-impl --team <team> --project <対象repo> --headless --implementer
```

cwd=対象repoなので**プロジェクト単位でspawnする**。別repoの実装を頼むときは`despawn.sh <team> claude codex-impl`してから対象repoで再spawn。

### 3. [implement]パケットを送る

「[implement]パケットの鉄則」の書式で組み立て、`send.sh <team> claude codex-impl --stdin < packet.txt`(非同期)。送ったらターンを終えてよい(返信はMonitorが拾う)。

### 4. Q&Aループ(実装中の質問対応)

codex-implは判断に迷うと質問を返信してターンを終える(role fileで強制)。Monitor通知で受けたら:

- 起案者の権限・知識で答えられる質問は**即答**して`send.sh`で返信(ターン往復で実装が継続する)
- ユーザー判断が要る論点はAskUserQuestionにかけてから回答を返信
- 返信するまでcodex-implは止まっている。放置しない

### 5. [done]受領→検収(Fable単独)

[done]報告を鵜呑みにせず実物を確認する:

- `git -C <repo> status` / `git diff`を読み、(1)要件適合・プラン逸脱 (2)正しさ・エッジ・回帰リスク を査読
- `git log`で勝手commitがないことを確認
- 報告された検証(テスト等)を1-2コマンドでスポット再現

差し戻しは「指摘→対応」対応表付きで修正指示を送る。**収束は目安最大2巡**(「レビュー収束条件」を検収に読み替え): substantive(正しさ・設計・回帰)な指摘が残る巡だけ差し戻し、Low/nitのみなら自分で微修正するか見送り理由を明記して閉じる。

### 6. 検収通過→diff提示→commit

ユーザーに実際の差分を提示し、承認後にcommit(グローバルCLAUDE.mdのcommit規約)。codex-implの出力そのものは転載せず、採用結果だけ報告する。

## 依頼パケットの鉄則

codex系ワーカーへ送るパケットの書式、Claude側の検品・収束ルール。

### [research]パケットの鉄則（codex-research宛）

- codex-research は**調査専任**(consultant layout: repoはread、書けるのはagmsg配下のみ=実質read-only)。返すのは file:line 一覧や構造化データだけ。**パッチは作らせない**＝実装は codex-impl が書く。role file は `db/spawn-roles/codex-research.codex.md`(gh はGET/検索系のみ可・issue/PR作成禁止・URL/番号の捏造禁止)
- agmsg: 宛先 codex-research・prefix `[research]`。自己完結パケット = GOAL / CONSTRAINTS / SCOPE / SCHEMA(期待する出力の構造・項目を明示) / 検品観点(Claude が何を確認するか) / DO NOT write files・DO NOT パッチ生成
- **sonnet班との併走を推奨**(2026-07-12ユーザー指示): 中〜大の調査テーマは、コードベース内をcodex-research・外部Web/GitHub/ライブラリ仕様をsonnetサブエージェントに同時に投げて両面から掘り、Claudeが突き合わせて検品する
- Claude は返答を**検品**する。SCHEMA を満たさない・情報が不足している場合は「◯件中◯件で△△が不足」のように対象を具体的に指摘して**差し戻す**。検品を通った調査結果/データを起点にプラン・[implement]パケットを組む
- **インクリメンタル調査**: 大きい調査は一括で丸投げにせず、**1トピック/1論理単位ずつ**調査させ、各単位を検品(SCHEMA充足・過不足の即チェック)してから次の単位へ進める。巨大な調査のやり直し(トークン浪費)を防ぎ、早期に軌道修正する

### [implement]パケットの鉄則（codex-impl宛）

- 宛先codex-impl・prefix `[implement]`。**自己完結**が絶対条件(codex-implはパケット本文とrepoしか見ない): プラン全文(背景1段落・要件・制約・完了条件)/ 検証手順(実行すべきテスト・ビルドコマンド)/ 報告書式([done]: 変更ファイルのfile:line一覧・プランとの対応・実行した検証と結果・逸脱と残課題)/ 質問プロトコル(迷ったら実装を止めて`send.sh <team> codex-impl claude "..."`で質問: 何を実装中か・選択肢・推奨案)
- DO NOTを明記: git commit/push禁止・対象repo外の変更禁止・無断の設計変更禁止・未検証の完了報告禁止
- 長文は`--stdin`で送る(quoting事故防止)
- turn timeoutは3600秒(config済み)。「区切りの良いところまで進めて途中経過を報告してよい」と書くと長タスクの往復が安定する
- networkはoff(workspace-write既定)。外部情報が要る作業は、必要な情報をパケットに同梱するか、事前にsonnet班で収集して渡す

### codex依頼パケットの鉄則

（2026-07-12〜: 標準フローの検収はFableが担うため、codex査読はオンデマンド。依頼時の書式は従来どおり以下）

- codex は **read-only**。findings を返すだけで、**fix は Claude が適用**する。依頼は review/verify/findings/test-plan のみ。codex に計画を振らない
- agmsg: 宛先 codex・prefix `[review]`。自己完結パケット = `git diff` か対象 `file:line` ＋ 意図 ＋(ループ時)前回指摘→対応の対応表(codex はメッセージ本文しか見ない)
- 出力形式: Findings / Required tests / Residual risk / Confidence。severity 順・推測は明記
- 対象は **実質的な実装のみ**。1行修正など些細な編集はレビュー不要

### レビュー収束条件

**実質的な**実装(機能・タスク完了の節目。1行修正など些細な編集は除く)をしたら、その diff を codex に非同期`send`してレビューを受ける(codexはask往復が機能しない実績があるため`ask`は使わない)。指摘の反映(再修正)も「実装」なので反映後の差分を再依頼しうるが、**1回で止めるな・延々と回すな**。能動的に妥協点を見出して打ち切る。

収束条件(いずれか満たせば完了とみなす):
- 残る findings が **Low / nit / 「見送り(理由明記)」/ 「別タスク(スコープ外)」だけ** で、substantive(Med/High＝正しさ・設計・回帰に関わる)な findings が無い。
- 全 finding を **採用 / 見送り(理由明記) / 別タスク(スコープ外)** に振り分けて反映・判断済み。採用が codex 自身の提案なら、その反映は些末扱いで再依頼不要。
- **再依頼は substantive な finding が出た巡だけ**。Low だけの巡は 1 巡で閉じる。新規性のない同深刻度の繰り返しは収束。**目安は最大2巡**、超えるなら残課題を「別タスク」化して打ち切る。

各巡では「前回指摘 → 対応」の対応表をパケットに含める。**ask が timeout / 無返信のときは無限に待たず**、bridge ログ(`run/<type>-bridge.<team>.<name>.log`)から findings を読んで内容ベースで収束判断する。

## 使用例

### 例1: codex-researchにコードベース内調査を頼みつつ、別タスクを並行で進める

```
チームで協業して: codex-researchに既存コード内のXの使用箇所・依存関係の棚卸しを頼みながら、私は外部ライブラリYのAPI仕様をsonnetサブエージェントに調査させる
```

codex-researchへ`[research]`パケット(コードベース内スコープ)を非同期送信 → 待たずに自分は外部ライブラリ調査(sonnetサブエージェント+WebFetch/WebSearchまたはcontext7 MCP)を進める → codex-researchの返信がMonitor通知で届いたら検品(不足があれば差し戻し) → 検品を通ったらプランを固めてcodex-implへ`[implement]`を送り、[done]をFableが検収する。

### 例2: codex-researchの調査完了後、codex-implへ実装を委譲しつつ次の独立タスクを走らせる

codex-researchの返信到着(Monitor通知)→検品→プラン確定→codex-implへ`[implement]`を非同期送信→待たずに別の独立タスクに着手→[done]到着で検収。

## ask --wait との使い分け

| 状況 | 方式 |
|---|---|
| 単発・並行を問わず通常の依頼 | 非同期`send` + Monitor再開(既定) |
| `ask --wait`(ブロック待機) | **使わない**。codexはask往復が機能しない実績があり、cursor離脱(2026-07-12)で実質的な用途が消えた |
| codexがread-onlyで返信不能な既知の環境(ネストsandbox等) | 非同期でも返信不能。bridgeログから読む既存フォールバックに従う |

## 注意点

- session_teamが無効、または宛先workerが別teamにいる場合はMonitorが拾えない。事前に同じteamへの参加を確認する。
- 非同期化してもcommit・適用の最終主体はClaudeである原則は変わらない。codex-researchの調査もcodex-implの[done]も無検証で受け入れない(検品・検収ゲートを通す)。
- 大量のタスクを一度に並行させすぎない。依存関係の見落としは収束を遅らせるだけ。

## 返信が来ない時の診断（2026-07-06実例）

1. **bridgeログ確認**: `~/.agents/skills/agmsg/run/<type>-bridge.<team>.<name>.log` を見る。`rc=124`（CLIタイムアウト）+「leaving message unread」なら依頼は未読滞留しており、bridgeプロセス自体も死んでいることが多い
2. **stale pidfile**: bridgeが死んでいるのに `spawn.sh` が「already running (pid N)」と言う場合は、`despawn.sh <team> <from> <name> --force` で登録を掃除してから再spawnする。再spawnしたbridgeは未読メッセージを自動で再処理する
3. **Monitor(watch.sh)が常駐していない場合のfallback**: 返信はDB直読みで取得できる — `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT body FROM messages WHERE team='<team>' AND from_agent='<agent>' ORDER BY id DESC LIMIT 1;"`。到着待ちはバックグラウンドのポーリングループ（10秒間隔でCOUNTを見て、>0で即exit・15分でタイムアウト）にすると、到着時に通知で拾える
4. **CLI更新後は必ずdespawn→再spawn**: bridgeは起動時のCLIバイナリを掴み続けるため、codex CLI等を更新しても既存bridgeには反映されない（実例: 2026-07-10、旧CLIが「gpt-5.6-sol requires a newer version of Codex」の400で全turn失敗し続けた。ログ上はturn completed with errorが並ぶ）。`despawn.sh <team> <from> codex --force` → `ensure-codex.sh` で入れ替える。なお `ensure-codex.sh` は生存確認を兼ねる（生きていれば "already running" のno-op）
5. **既読消化された依頼は再送**: 壊れたbridgeが依頼を既読処理してしまった場合、再spawn後の自動再処理は未読のみが対象なので、該当依頼は `send.sh` で再送する。既読状態は `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT id, read_at IS NOT NULL FROM messages WHERE team='<team>' AND from_agent='claude' ORDER BY id DESC LIMIT 3;"` で確認できる

## codex-research（調査用codexワーカー）

codexの既定role（`spawn-roles/codex.codex.md`）はreview専任で、どう頼んでも調査を拒否する（2026-07-06実証）。調査は**別名ワーカーcodex-research**に送る（review役`codex`・実装役`codex-impl`と共存）:

```bash
~/.agents/skills/agmsg/scripts/ensure-codex.sh <project> codex-research
```

role file は `db/spawn-roles/codex-research.codex.md`（規約名でensure-codex.shから自動解決）。依頼書式は上記「[research]パケットの鉄則」を使う。

## codexワーカーのモデル/effort振り分け（2026-07-08〜）

agmsg configのper-workerキー（`spawn.codex_model.<name>` / `spawn.codex_effort.<name>`、codex headless限定）により、ワーカー名がモデル+effortのプリセットになっている。タスク難易度の判定はコードで自動化せず、依頼側（Claude）が適切な名前のワーカーへ送ることで実現する。

| タスク | 宛先ワーカー | モデル/effort |
|---|---|---|
| 実質的な機能実装の自走 | codex-impl | gpt-5.6-sol / xhigh（global継承）・turn_timeout 3600s・implementer layout（cwd=対象repo+workspace-write） |
| 実装後レビュー（オンデマンド） | codex | gpt-5.6-sol / xhigh（global継承） |
| 調査・軽微な確認・大量列挙 | codex-research | gpt-5.6-terra / xhigh（global継承。遅いと感じたら `spawn.codex_effort.codex-research: high` へ下げる） |
| 大規模・設計横断の節目レビュー | codex-deep（一時spawn→使い捨て） | gpt-5.6-sol / max（configキー設定済み） |

codex-deepは常駐させず、必要時にspawnし終わったらdespawnする:

```bash
~/.agents/skills/agmsg/scripts/spawn.sh codex codex-deep --team <team> --project <project> \
  --role-file ~/.agents/skills/agmsg/db/spawn-roles/codex.codex.md   # review役を流用
# 使用後: ~/.agents/skills/agmsg/scripts/despawn.sh <team> claude codex-deep
```

注意:
- 重いモード（max/ultra）への自動エスカレーションはしない。gpt-5.6系のクレジット消費倍率が未公開のため、明示的に選んだ時だけ使う
- gpt-5.6系はプレビュー段階。モデルIDが無効化・改名されたら該当configキー（実効config: `~/.agents/skills/agmsg/db/config.yaml`）を更新して戻す（振り分けの仕組み自体はモデル非依存）
- **gpt-5.6-sol が bridge 経由で `400 "requires a newer version of Codex"` になる場合**: 原因はCLI版ではなく、app-server `initialize` の `clientInfo.name` に対する server-side gate（bridgeの既定名 `agmsg-codex-bridge` が sol の first-party allowlist に弾かれる）。per-worker キー `spawn.codex_client_name.<name>: codex_cli` を設定して first-party 名を名乗らせると解消する（既定は従来名のまま。sol を使う worker にのみ設定）。CLI を最新stableに上げても・再認証しても直らない（2026-07-10確認）。純正 `codex exec` は別API面のため通るので、exec成功をbridge成功の証拠にしないこと。

## 関連スキル

- `/agmsg` - inbox確認・送信・履歴
- `/commit` - 変更の確定

## 関連ドキュメント

- グローバル`CLAUDE.md`の「エージェント役割分担」節 - 役割分担の方針・振り分け基準・トークン節約ガード(依頼パケットの書式・検品手順・収束条件はこのファイルの「依頼パケットの鉄則」節を参照)
