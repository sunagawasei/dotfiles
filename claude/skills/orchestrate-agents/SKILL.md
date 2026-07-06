---
name: orchestrate-agents
description: cursor(調査/データ収集)とcodex(レビュー)への委譲をagmsgの非同期send+Monitor自動再開で並行実行する協業ワークフロー
---

# チーム協業ワークフロー

cursor(横断調査・データ収集専任)とcodex(実装後レビュー)への委譲を、`ask --wait`(ブロック待機)ではなく非同期`send`+agmsg Monitorの自動再開で回し、複数タスクを並行して進める。

## 概要

非同期`send`+agmsg Monitor自動再開は、cursor/codexへの依頼の既定の送信方式(2026-07-05〜、単発でも同じ)。このスキルは、複数の独立したタスクを同時に投げて並行させたい場面のための分解・委譲テンプレートで、CLAUDE.mdの役割分担・パケット書式・検証ゲートは一切変えない。送ったらターンを終え、返信はこのセッションが最初から起動しているagmsg Monitor(SessionStartで自動起動する`watch.sh`、session_teamの`s-<UUID>`宛)が拾って通知し、ターンを自動再開する。

cursorはCLAUDE.mdの方針転換(2026-07-04〜)により**横断調査・データ収集専任**。実装ドラフト(パッチ)は作らず、実装は常にClaude自身が書く([agmsg-demo-fable5-x-constellation](https://github.com/fujibee/agmsg-demo-fable5-x-constellation)のデモのgrok役に相当)。

## 使うタイミング

- ユーザーが「チームで協業して」「並行でやって」など、明示的にこのモードを求めたとき
- 複数の独立したサブタスク(例: cursorへの調査依頼と、それとは別系統の実装・確認作業)を同時に進められるとき

単発で1エージェントに1回聞くだけなら、このスキルの分解テンプレートは不要(送信方式は同じ非同期`send`+Monitor)。`ask --wait`(ブロック待機)はWezTermタブのbusy表示を維持したい単発依頼だけのオプション(codexはask往復が機能しない実績があるため対象外)。

## 前提

- このセッションのagmsg Monitor(`watch.sh ... --team s-<このセッションのUUID>`)がSessionStartから常駐している(追加設定不要)。
- cursor/codexは、既存のCLAUDE.md運用と同じくこのセッションのteamに参加済みであること。
- cursor/codexのbridgeプロセスが稼働していること。遅延spawnは自動では発火しない(下記手順2で確認・起動する)。
- CLAUDE.mdの役割分担(cursor=横断調査・データ収集専任・read-only、codex=実装後レビュー・read-only、Claude=唯一の適用者/実装を書く者/最終判断者)は変更しない。

## ワークフロー

### 1. タスクを独立した単位に分解する

並行させられる単位に分ける。依存関係がある場合(例: 実装はcursorの調査結果が検品を通った後、codexのレビューはClaudeの実装完了後)は、その順序を崩さない — 依存先は該当タスクの完了通知を受けてから送る。

### 2. 非同期でパケットを送る(ask ではなく send)

パケットの中身は下記「依頼パケットの鉄則」と同じ自己完結フォーマットを使う。送信は`--wait`を付けない素の`send.sh`で行うが、**送る前に宛先のbridgeが起きているか確認する**:

- **codex宛**: 必ず先に `~/.agents/skills/agmsg/scripts/ensure-codex.sh <project>` を実行して遅延spawnを発火させる(起動済みならno-op。`CLAUDE_CODE_SESSION_ID`が環境に無ければセッションUUIDを明示して渡す)。`send.sh`自体は一方向送信やctrl系がdespawn済みworkerを蘇生させないよう、意図的にこの入口へ配線されていない
- **cursor宛**: `pgrep -f "cursor-bridge.*<team>"`でbridge稼働を確認し、いなければ`spawn.sh`で起動してから送る
- 怠ると依頼は未読のままDBに滞留し、返信が永遠に来ない(実例: 2026-07-05)

```bash
# cursorへの調査依頼(ブロックしない、prefixは [research])
~/.agents/skills/agmsg/scripts/send.sh $TEAM $AGENT cursor "[research] GOAL: ... / CONSTRAINTS: ... / SCOPE: ... / SCHEMA: ... / 検品観点: ..."

# codexへのレビュー依頼(ブロックしない、prefixは [review])
~/.agents/skills/agmsg/scripts/send.sh $TEAM $AGENT codex "[review] <git diff or file:line> / 意図: ... / 前回指摘→対応: ..."
```

diff同梱など長文・quoting事故が起きやすいパケットは、本文を4番目の引数で渡さず、ファイルに組み立ててから `send.sh <team> <from> <to> --stdin < packet.txt` で標準入力から渡す(実例: 2026-07-05のcodexレビュー依頼)。

`$TEAM`/`$AGENT`はこのセッションの既知の値(agmsgスキルのIdentityで確認済みのもの)を使う。

### 3. ターンを終えるか、他の独立作業を続ける

送信直後にやれる別の独立作業があれば進める。なければそのままターンを終えてユーザーに制御を返してよい — ブロックする必要はない。返信はMonitor通知で自動的に拾われる。

### 4. 返信が来たら検品・実装・検証する

Monitor通知でcursor/codexからの返信を受け取ったら、下記「依頼パケットの鉄則」のゲートをそのまま適用する:

Monitor通知は長い返信を途中で切り詰める。全文は `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT body FROM messages WHERE team='<team>' AND from_agent='<agent>' ORDER BY id DESC LIMIT 1;"` で読む(bridgeログより確実)。

- **cursorの調査結果**: 「cursor依頼パケットの鉄則」に従って検品する。不足があれば具体的に指摘して差し戻し、検品を通ったら、その調査結果を起点に**Claude自身が実装を書く**
- **codexのfindings**: 「レビュー収束条件」に従って採用/見送り/別タスク化を判断

複数の返信が並行して届いた場合は、それぞれ独立に処理する(依存関係があるものだけ順序を守る)。

### 5. 収束したらユーザーに要約報告

cursor/codexの出力そのものは転載せず、採用した内容と最終的な変更のみを報告する(CLAUDE.mdの既存方針どおり)。

## 依頼パケットの鉄則

cursor/codexへ送るパケットの書式、Claude側の検品・収束ルール。`ask --wait`の同期依頼でも本スキルの非同期`send`でも共通で適用する。

### cursor依頼パケットの鉄則

- cursor は **read-only**(ファイルは書けない)。コードベース内の横断調査・外部Web/ドキュメント調査の**両方**を任せてよいが、返すのは file:line 一覧や構造化データだけ。**パッチは作らせない**＝実装は必ず Claude 自身が書く。cursor に diff レビューを振らない(codex の領域)
- agmsg: 宛先 cursor・prefix `[research]`。自己完結パケット = GOAL / CONSTRAINTS / SCOPE(対象範囲: コードベース内 or 外部ソース、両方もあり得る) / SCHEMA(期待する出力の構造・項目を明示) / 検品観点(Claude が何を確認するか) / DO NOT write files・DO NOT パッチ生成
- Claude は cursor の返答を**検品**する([agmsg-demo-fable5-x-constellation](https://github.com/fujibee/agmsg-demo-fable5-x-constellation)のデモの Fable→grok 検品ループと同型)。SCHEMA を満たさない・情報が不足している場合は「◯件中◯件で△△が不足」のように対象を具体的に指摘して**差し戻す**。検品を通った調査結果/データを起点に、実装は Claude が自分で書く
- **インクリメンタル調査**: 大きい調査は一括で丸投げにせず、cursor に **1トピック/1論理単位ずつ** 調査させ、各単位を Claude が検品(SCHEMA充足・過不足の即チェック)してから、次の単位へ cursor を進める。巨大な調査のやり直し(トークン浪費)を防ぎ、早期に軌道修正する。codex の深い実装後レビューはこれとは別レイヤー(節目で実施)

### codex依頼パケットの鉄則

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

### 例1: cursorに外部ライブラリ調査を頼みつつ、別ファイルの調査を並行で進める

```
チームで協業して: cursorにライブラリXのAPI仕様調査を頼みながら、私は関連する既存コードの調査を進める
```

cursorへ`[research]`パケットを非同期送信 → 待たずに自分は調査を継続 → cursorの返信がMonitor通知で届いたら検品(不足があれば差し戻し) → 検品を通ったらClaudeが実装を書く → 完了したらcodexへ`[review]`を送って収束させる。

### 例2: cursorの調査完了後、Claudeの実装をcodexにレビューさせつつ次の独立タスクを走らせる

cursorの返信到着(Monitor通知)→検品→Claudeが実装 → 直後にcodexへ`[review]`を非同期送信 → 待たずに別の独立タスクに着手 → codexの返信到着で収束処理。

## ask --wait との使い分け

| 状況 | 方式 |
|---|---|
| 単発・並行を問わず通常の依頼 | 非同期`send` + Monitor再開(既定) |
| 単発でWezTermタブのbusy表示を維持したい | `ask --wait`(オプション。codexはask往復が機能しない実績があるため対象外) |
| codexがread-onlyで返信不能な既知の環境(ネストsandbox等) | 非同期でも返信不能。bridgeログから読む既存フォールバックに従う |

## 注意点

- session_teamが無効、またはcursor/codexが別teamにいる場合はMonitorが拾えない。事前に同じteamへの参加を確認する。
- 非同期化してもClaudeが唯一の書き込み主体である原則は変わらない。cursorの出力を無検証で適用しない(そもそもcursorはパッチを作らない)。
- 大量のタスクを一度に並行させすぎない。依存関係の見落としは収束を遅らせるだけ。

## 返信が来ない時の診断（2026-07-06実例）

1. **bridgeログ確認**: `~/.agents/skills/agmsg/run/<type>-bridge.<team>.<name>.log` を見る。`rc=124`（CLIタイムアウト）+「leaving message unread」なら依頼は未読滞留しており、bridgeプロセス自体も死んでいることが多い
2. **stale pidfile**: bridgeが死んでいるのに `spawn.sh` が「already running (pid N)」と言う場合は、`despawn.sh <team> <from> <name> --force` で登録を掃除してから再spawnする。再spawnしたbridgeは未読メッセージを自動で再処理する
3. **Monitor(watch.sh)が常駐していない場合のfallback**: 返信はDB直読みで取得できる — `sqlite3 ~/.agents/skills/agmsg/db/messages.db "SELECT body FROM messages WHERE team='<team>' AND from_agent='<agent>' ORDER BY id DESC LIMIT 1;"`。到着待ちはバックグラウンドのポーリングループ（10秒間隔でCOUNTを見て、>0で即exit・15分でタイムアウト）にすると、到着時に通知で拾える

## codex-research（調査用codexワーカー）

codexの既定role（`spawn-roles/codex.codex.md`）はreview専任で、どう頼んでも調査を拒否する（2026-07-06実証）。調査をcodex課金プールに流したい場合は、調査用roleで**別名ワーカー**を立てる（review役の`codex`とは共存）:

```bash
~/.agents/skills/agmsg/scripts/spawn.sh codex codex-research --team <team> --project <project> \
  --role-file ~/.agents/skills/agmsg/db/spawn-roles/research.codex.md
```

role file `research.codex.md` は作成済み（read-only調査専任・ghはGET/検索系のみ可・issue/PR作成禁止・URL/番号の捏造禁止）。依頼パケットの書式はcursorの`[research]`鉄則と同じものを使う。

## 関連スキル

- `/agmsg` - inbox確認・送信・履歴
- `/commit` - 変更の確定

## 関連ドキュメント

- グローバル`CLAUDE.md`の「エージェント役割分担」節 - 役割分担の方針・振り分け基準・トークン節約ガード(依頼パケットの書式・検品手順・収束条件はこのファイルの「依頼パケットの鉄則」節を参照)
