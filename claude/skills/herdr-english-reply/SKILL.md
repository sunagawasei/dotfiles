---
name: herdr-english-reply
description: >-
  別workspaceのAIエージェントとの会話で、相手の最新返信を読んだうえでユーザーの返信案を
  カジュアルな英語に添削し、代替表現とニュアンス差を示す。Use when the user wants English
  reply help / 英文添削 for a chat with an AI agent running in another herdr workspace,
  or invokes this skill with a workspace number.
---

# herdr English Reply

別workspaceで動くAIエージェント宛の返信を、相手の最新メッセージを踏まえてカジュアルな英語に整える。添削結果だけでなく、選べる表現とそのニュアンス差も出す。

前提は `HERDR_ENV=1`。未設定なら対象paneを読めないので、その旨を伝えて停止する。herdrコマンドの詳細は `herdr` skillに従う。

## ワークフロー

### 1. workspace番号を待つ

起動直後は**何も読まない**。ユーザーが対象のworkspace番号（`herdr workspace list` の `number`）を送ってくるのを待つ。番号が来たら解決して「準備できた、返信案を送って」とだけ返す。

### 2. 対象paneを解決する

**添削のたびに毎回引き直す。** paneやworkspaceが閉じるとidは詰められて再利用されるため、前回のidをそのまま使うと別セッションを読む。

```bash
herdr workspace list
herdr pane list --workspace <workspace_id>
```

- `number` 一致のエントリから `workspace_id` を取る。初回に控えた `workspace_id` と食い違ったら黙って進めず、`label` を添えて対象が変わっていないか確認する
- pane一覧のうち `agent` フィールドを持つpaneが対象（nvim等のpaneにはこのフィールドがない）。複数あれば pane_id / agent / `terminal_title_stripped` を並べて選ばせる。0件ならその旨を伝えて聞き直す

### 3. 相手の最新返信を読む

```bash
herdr pane read <pane_id> --source recent-unwrapped --lines 80
```

TUI装飾・spinner・入力プロンプト行は捨て、エージェント側の**最新の実質的な発言**だけを文脈にする。途中で切れていたら `--lines` を増やす。まだ何も出力がない場合は文脈なしで添削してよいが、そのことを明示する。

### 4. 添削する

入力は日本語でも英語でもよい。ユーザーの意図を保ったまま、AIエージェント相手のチャットとして自然な英語にする。

- 省略形・命令形・短い文でよい。`I would appreciate it if you could` 系の枕は落とす
- ビジネスメール調・教科書英語にしない。逆にスラングや皮肉を足すのも禁止
- コード識別子・パス・技術用語・エラーメッセージは原文のまま残す
- 「よしなに」「いい感じに」のような曖昧語は具体的な英語に落とす。落とし方が複数あるなら断定せず選択肢として出す

## 出力フォーマット

説明は日本語、返信本文は英語。`Reply` はそのまま貼れる完成文をコードブロックで出す。

- **Context** — 相手の最新返信の要点を1〜3文。長い引用はしない
- **Reply** — コピペ用の英文
- **Notes** — 直した点と、代替表現を1〜2個（各1行でニュアンス差を添える）

「もっとカジュアルに」「短く」「この単語は残して」などの指示が来たら、同じフォーマットで出し直す。

## 禁止事項

**対象paneへ勝手に送信しない。** `pane send-text` / `pane run` は、ユーザーが明示的に送信を頼んだときだけ実行する。添削の完了は送信の許可ではない。

## 使用例

ユーザー: `3` → workspace 3 のエージェントpaneを解決して待機。

ユーザー: `それで進めて。終わったら結果だけ見せて`

→ paneを読み、Context「実装方針に合意済みで、テストを回すか確認してきている」／Reply `Sounds good, go ahead. Just show me the results when you're done.`／Notes「別案 `Cool, proceed — ping me with the outcome.`（もう少し軽く、急かさない感じ）」
