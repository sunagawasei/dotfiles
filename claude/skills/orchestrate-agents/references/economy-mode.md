# 節約モード

Anthropicプランの消費を抑える運用。段3(プラン査読の指摘振り分け)は通常モードのfable-review(fable)からdesign-review(opus)へ、段5(実装)はimpl-worker(sonnet)からメインセッション自身へ、段9(diff査読)の指摘に対する対応要否判断はdesign-review(opus)へ、それぞれ担当を替える。委譲フローは段1〜11のまま1本で、担当者だけが入れ替わる。別フロー・別skillではない。

**段5をメインが自分で実装するため、段6(受け入れ検査)・段8(author再分類)は成立しない** — 実装主体と検査主体が同一になり、最初からauthor:mainとして段9へ進む。段9の査読はcodex(review役)のままで、メインが書いたコードをcodexが査読する構図はvendorが異なる(Anthropic/ChatGPT)ため、通常モードと同じ独立性を保つ。

## 人員配置

| 段 | 通常モード | 節約モード | 課金プール(節約モード時) |
|---|---|---|---|
| 段2 プラン査読 | codex(review役) | codex(review役)(変更なし) | ChatGPT |
| 段3 指摘の振り分け | fable-review (claude-code, fable) | design-review (claude-code, opus) | Anthropic |
| 段5 実装 | impl-worker (Agent tool, sonnet) | メインセッション自身(委譲しない) | メインの実行環境に準ずる |
| 段9 diff査読 | codex(review役) | codex(review役)(変更なし) | ChatGPT |
| 段9指摘の対応要否判断 | メインが段10の振り分けの中で兼ねる | design-review (claude-code, opus) が採用/見送りを判断 | Anthropic |
| 段10 修正の担当 | メインが振り分け(3経路) | メインが直す(実装主体が単一のため振り分け先は1つ) | メインの実行環境に準ずる |
| メイン | Opus | 変更なし(実行環境は起動時にユーザーが選ぶ。固定しない) | 不定 |

段11の検収とcommitは通常モードと同一。

## 発動条件

ユーザーの明示指示があるときだけ切り替える。タスクの途中でモードを切り替えない。

## 起動手順

`ensure-*`スクリプトは既存workerの生存確認を兼ねる(生きていれば再実行してもno-op)。

### design-review(claude-code、段3・段9指摘の対応要否判断)

```bash
AGMSG_CLAUDE_PROBE_TIMEOUT=150 ~/.agents/skills/agmsg/scripts/ensure-headless.sh claude-code <project> design-review
```

probe timeoutの既定30秒では足りない(SKILL.md段3節、troubleshooting.md項14)。値はリテラルで書く(troubleshooting.md項10)。config設定済み: `spawn.claude_model.design-review: opus` / `spawn.claude_effort.design-review: high` / `spawn.claude_turn_timeout.design-review: 1800` / `spawn.claude_inherit_add_dirs.design-review: true`。read-onlyはグローバル既定`spawn.claude_reviewer: true`で担保する。

### readiness照合

design-reviewはagmsg経由でdispatchする相手なので、SKILL.mdの通常規約をそのまま適用する: registrationが期待typeでちょうど1件あること、かつ当該spawn世代以降の返信実績があること。`ensure-*`はsession team不在でもexit 0のno-opになるため、exit codeだけを準備完了の証拠にしない。

design-reviewは追加で、`spawn.claude_implementer.design-review`が未設定(または`false`)であることと、生成済みsettings.jsonのdenyWriteにprojectパスが含まれることを確認する(layout差はregistration照合では検出できない)。

### role fileの世代

role fileはspawn時に`run/`配下へコピーされ、稼働中のworkerには反映されない(SKILL.md段3節)。role fileを変えた後の初回起動はdespawn→respawnする。

## 段10

design-reviewが「対応する」と判断した指摘は、メインが自分で直して段9へ再投入する(実装主体が単一のため、通常モードのような複数経路への振り分けは発生しない)。design-reviewが「見送り」と判断した指摘は、理由を記録してcommitに進む。

## 通常モードへ戻す

使ったworkerをdespawnする。

```bash
~/.agents/skills/agmsg/scripts/despawn.sh <team> claude design-review
```

despawn前にin-flight dispatchを棚卸しする(同名workerを後から再spawnすると旧dispatchが再駆動される。SKILL.md「共通の不変条件」節)。SessionEnd teardownでもsession teamのheadless worker全員が回収される。

## 残存リスク

- design-reviewが段3(プラン指摘の振り分け)と段9指摘の対応要否判断の両方を兼ねる。同じ観点の見落としがある場合、両方の判断に影響しうる
- メインが実装も検査も兼ねるため、通常モードの段6が想定する「申告外ファイルの突き合わせ」のような第三者チェックが働かない。段9のcodex査読が唯一の外部チェックになる
- cursor経由・codex経由がAnthropicプランを消費しないという前提は未測定の推論
- 段3・段9対応要否判断のdesign-reviewはAnthropicプールのまま。Fableの半額になるだけで、プール自体は移らない
