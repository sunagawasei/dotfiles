---
name: herdr-english-reply
argument-hint: "[workspace_number]"
description: >-
  別workspaceのAIエージェントとの会話で、相手の最新返信を読んだうえでユーザーの返信案を
  平易でカジュアルな英語に添削・英訳し、代替表現とニュアンス差を示す。Use when the user wants
  English reply help / 英文添削 / 日本語からの英訳 for a chat with an AI agent running in
  another herdr workspace, or invokes this skill with a workspace number.
---

# herdr English Reply

別workspaceで動くAIエージェント宛の返信を、相手の最新メッセージを踏まえてカジュアルな英語に整える。添削結果だけでなく、選べる表現とそのニュアンス差も出す。

前提は `HERDR_ENV=1`。未設定なら対象paneを読めないので、その旨を伝えて停止する。herdrコマンドの詳細は `herdr` skillに従う。

## ワークフロー

### 1. workspace番号を受け取る

`/herdr-english-reply <番号>`（例: `/herdr-english-reply 3`）のようにARGUMENTSへworkspace番号（`herdr workspace list` の `number`）が渡されていれば、それを使ってすぐ手順2へ進む。

ARGUMENTSが空、または番号として読み取れない場合は、起動直後は**何も読まない**まま、ユーザーが番号を送ってくるのを待つ。番号を解決したら「準備できた、返信案を送って」とだけ返す。

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

### 4. 添削または英訳する

入力が英語なら添削、日本語なら英訳する。どちらの場合も出力フォーマットは同じ。ユーザーが英作文できなくて日本語をそのまま送ってくるのは想定内なので、英語で書き直すよう促さない。

**返信したい本文なのか、直前の提案へのコメントなのかを取り違えない。** 短い発言は両方の意味になりうる（例: 「そのままでいいや」は「Replyはこの案でいいから送る内容として使って」の意にも「直前の質問への回答」の意にも読める）。`"..."` や `「...」` で囲まれた文字列は常に「返信したい内容そのもの」として扱う。囲みがなく判別できない場合は、対象paneへの返信案として扱ってよいか一言確認してから進める。

**平易な英語を最優先する。** ユーザーが自分で読めて、次回は自分で書ける語彙に落とす。

- 中高レベルの語で言えることに難しい語を使わない（`utilize`→`use`、`ascertain`→`check`、`in the event that`→`if`）
- 凝ったイディオム・句動詞・比喩表現を持ち込まない。直球で言う
- 一文を短く切る。関係代名詞で伸ばすより2文に割る
- 省略形・命令形でよい。`I would appreciate it if you could` 系の枕は落とす
- ビジネスメール調・教科書英語にしない。逆にスラングや皮肉を足すのも禁止
- コード識別子・パス・技術用語・エラーメッセージは原文のまま残す
- 「よしなに」「いい感じに」のような曖昧語は具体的な英語に落とす。落とし方が複数あるなら断定せず選択肢として出す

## 出力フォーマット

説明は日本語、返信本文は英語。`Reply` はそのまま貼れる完成文をコードブロックで出す。

- **Context** — 相手の最新返信の要点を1〜3文。長い引用はしない
- **Reply** — コピペ用の英文
- **Notes** — 直した点と、代替表現を1〜2個（各1行でニュアンス差を添える）。代替案にも平易さの基準を同じく適用する

「もっとカジュアルに」「短く」「この単語は残して」などの指示が来たら、同じフォーマットで出し直す。

`Reply` の文構造・文法を聞かれたら、都度Context/Reply/Notesの形式には拘らず、該当箇所を要素分解して日本語で説明する（例: 決まり文句なら由来と直訳、`動詞+目的語+補語`のような型なら型の名前と他の用例）。平易な英語を保つ方針と同様、文法用語も必要最小限にする。

## 禁止事項

**対象paneへ勝手に送信しない。** `pane send-text` / `pane run` は、ユーザーが明示的に送信を頼んだときだけ実行する。添削の完了は送信の許可ではない。

## 使用例

`/herdr-english-reply 3` または ユーザー: `3` → workspace 3 のエージェントpaneを解決して待機。

ユーザー: `それで進めて。終わったら結果だけ見せて`

→ paneを読み、Context「実装方針に合意済みで、テストを回すか確認してきている」／Reply `Sounds good, go ahead. Just show me the results when you're done.`／Notes「別案 `OK, please go ahead. Send me the results at the end.`（少し落ち着いた言い方）」
