# 節約モード

Anthropicプランの消費を抑えるため、段2(設計レビュー)と段5(実装)の担当を替える運用。段5はChatGPTプールのcodex-implへ移る。段2はAnthropicプールのまま、Fableの半額のOpusへ下げる。委譲フローは段1〜11のまま1本で、担当者だけが入れ替わる。別フロー・別skillではない。

**このモードは正しさより節約を優先する。** 段9の査読はcodexのままなので、codex-implが書いたコードをcodexが一次査読する。「同一vendorが自分の系列の成果を一次査読する配置を作らない」という原則にこのモードは反する。2026-09-12にユーザーが受容を明示した例外で、節約モードの中だけで成立する。通常モードはこの例外を持たない。

## 人員配置

| 段 | 通常モード | 節約モード | 課金プール(節約モード時) |
|---|---|---|---|
| 段2 設計レビュー+段3指摘の収束 | fable-review (claude-code, fable) | design-review (claude-code, opus) | Anthropic |
| 段3 プラン査読 | codex(review役) | codex(review役)(変更なし) | ChatGPT |
| 段5 実装 | impl-worker (Agent tool, sonnet) | codex-impl (agmsg, ChatGPT) | ChatGPT |
| 段9 diff査読 | codex(review役) | codex(review役)(変更なし) | ChatGPT |
| メイン | Opus | Opus(変更なし) | Anthropic |

段6の受け入れ検査、段8の統合、段10の差し戻し、段11の検収とcommitは通常モードと同一。

## 削減の実測値

2026-09-05〜09-12の実測(対象は`~/.config`のセッション、API list price換算)。
メイン(opus-5)が72.9%、fable-review(fable-5.1)が14.6%、実装subagent(sonnet-5)が10.9%。
段2をOpusへ寄せると段2が約7%に下がり、段5をcodex-implへ移すと10.9%が外部プールへ出る。
合計で約18%。メインの72.9%はこのモードでは減らない。

プランの使用量上限がこのドル換算に比例するかは未確認。上の比率はAPI list price換算の消費割合であって、プラン上限の消費割合そのものではない。

## 発動条件

ユーザーの明示指示があるときだけ切り替える。タスクの途中でモードを切り替えない。

## 起動手順

`ensure-*`スクリプトは既存workerの生存確認を兼ねる(生きていれば再実行してもno-op)。

### design-review(claude-code、段2)

```bash
AGMSG_CLAUDE_PROBE_TIMEOUT=150 ~/.agents/skills/agmsg/scripts/ensure-headless.sh claude-code <project> design-review
```

probe timeoutの既定30秒では足りない(SKILL.md段2節、troubleshooting.md項14)。値はリテラルで書く(troubleshooting.md項10)。config設定済み: `spawn.claude_model.design-review: opus` / `spawn.claude_effort.design-review: high` / `spawn.claude_turn_timeout.design-review: 1800` / `spawn.claude_inherit_add_dirs.design-review: true`。read-onlyはグローバル既定`spawn.claude_reviewer: true`で担保する。

### codex-impl(agmsg、段5)

```bash
CLAUDE_CODE_SESSION_ID=<このセッションのUUIDをリテラルで> ~/.agents/skills/agmsg/scripts/ensure-codex.sh <project> codex-impl
```

単独のsimple commandで呼ぶ。セッションUUIDはシェル変数にせずリテラルで書く(troubleshooting.md項10)。config設定済み: `spawn.codex_model.codex-impl: gpt-5.6-sol` / `spawn.codex_effort.codex-impl: xhigh` / `spawn.codex_turn_timeout.codex-impl: 3600` / `spawn.codex_client_name.codex-impl: codex_cli` / `spawn.codex_implementer.codex-impl: true`。

subtaskパケットの送信は`send.sh <team> claude codex-impl --stdin < packet.txt`。完了報告はclaude宛に返る(manager/watcherは介さない)。

パケットには`scope-ok:<task-id>/<subtask-id>`(そのsubtaskを名指しするトークン。メインが承認済みプランと照合して発行する)を必ず含める。無い場合codex-implは書き込まない契約になっている。**これは自然言語の契約で、トークンを検査する実行可能コードは無い。** `spawn.codex_implementer.codex-impl: true`によりcodex-implはlayout上repoを書ける。逸脱を捕まえる手段は段6のメインによる突き合わせだけで、その検出範囲は下記「残存リスク」のとおり限定的。

### readiness照合

design-reviewとcodex-implは、どちらもagmsg経由でdispatchする相手なので、SKILL.mdの通常規約をそのまま適用する: 各nameのregistrationが期待typeでちょうど1件あること、かつ当該spawn世代以降の返信実績があること。`ensure-*`はsession team不在でもexit 0のno-opになるため、exit codeだけを準備完了の証拠にしない。

design-reviewは追加で、`spawn.claude_implementer.design-review`が未設定(または`false`)であることと、生成済みsettings.jsonのdenyWriteにprojectパスが含まれることを確認する(layout差はregistration照合では検出できない)。

### role fileの世代

role fileはspawn時に`run/`配下へコピーされ、稼働中のworkerには反映されない(SKILL.md段2節)。role fileを変えた後の初回起動はdespawn→respawnする。

## 段10の差し戻し

codex-impl作ファイルのfindingは`send.sh <team> claude codex-impl --stdin`で差し戻す(実装subagentのSendMessageとは経路が違う)。findingに加えて`scope-ok`トークンを再発行して同梱する。セッションを跨いで文脈が失われている場合は再spawnし、元のsubtaskパケット(ゴール・制約・ファイルセット)も同梱する。

## 通常モードへ戻す

使ったworkerをdespawnし、次のタスクで通常モードのworker(fable-review・impl-worker)を起動し直す。

```bash
~/.agents/skills/agmsg/scripts/despawn.sh <team> claude design-review
~/.agents/skills/agmsg/scripts/despawn.sh <team> claude codex-impl
```

despawn前にin-flight dispatchを棚卸しする(同名workerを後から再spawnすると旧dispatchが再駆動される。SKILL.md「共通の不変条件」節)。SessionEnd teardownでもsession teamのheadless worker全員が回収される。

## 残存リスク

- codex-impl作コードをcodexが一次査読する。同一vendorなので、同じ前提を共有した見落としは捕まらない。これがこのモードの代償
- codex-implの書き込み範囲は契約のみで、機械的な境界は無い。申告外の新規ファイルは段6の突き合わせでは捕まらない。`git diff --stat`は未追跡ファイルを列挙しないため、宣言外に新規ファイルを作って報告から省かれると段6を通る(2026-09-12実測)。**これは実装subagent(impl-worker)でも同じで、このモードが新しく作るリスクではない** — impl-workerも親セッションの権限をそのまま継承し、外部write・repo外readの禁止は契約文でのみ抑止される(SKILL.md「実装subagentのDO NOT」節)。段6の検出力そのものの改善は本モードの範囲外
- cursor経由・codex経由がAnthropicプランを消費しないという前提は未測定の推論
- 段2のdesign-reviewはAnthropicプールのまま。Fableの半額になるだけで、プール自体は移らない
