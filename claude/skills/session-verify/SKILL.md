---
name: session-verify
description: 過去セッションで導入した変更・設定・修正が現在も機能しているかを裏取り調査する。Usage: /session-verify <session-id> [観点]
---

# セッション裏取り調査

過去のセッションIDを指定して「そのセッションで導入した変更が今も機能しているか」を検証する。

## ワークフロー

### 1. 対象セッションの特定

```bash
ls ~/.config/claude/projects/*/<session-id>.jsonl
```

- プロジェクトが分かっていれば直接 `~/.config/claude/projects/<slug>/<session-id>.jsonl`

### 2. 当時の変更の抽出

- jsonlから「ユーザーの指示」と「適用された変更」を抽出する:
  - ユーザー発言: `jq -r 'select(.type=="user") | .message.content | if type=="string" then . else ([.[] | select(.type=="text") | .text] | join("\n")) end' <file>.jsonl`
  - 変更対象: Edit/Writeツールの `file_path`、設定キー、コミットハッシュ
- **トークン節約**: 抽出・要約はsonnetサブエージェント（`model: sonnet`）に委譲してよい。判定は本セッションで行う

### 3. 現状との突き合わせ

- 変更されたファイル・設定キー・hookが現在も存在するか、内容が維持されているかを実物で確認する
- 挙動の検証が要るもの（hook・ルール遵守など）は、対象セッション**以降**のjsonlで実際に使われた/効いた形跡を確認する
- git管理下なら `git log -- <file>` でその後の変更履歴も見る

### 4. 報告

- 判定: **機能している / 劣化・部分的 / 巻き戻っている** のいずれかと根拠（file:line、後続セッションの実例）
- 巻き戻り・劣化があれば、修正案を提示（勝手に直さずユーザーの判断を仰ぐ）
