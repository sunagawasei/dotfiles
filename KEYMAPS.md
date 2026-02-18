# Neovim

## 診断全般のジャンプ（warning/error/info含む）

- ]d: 次の診断へ移動
- [d: 前の診断へ移動

## コメントアウト

`gcc`

## 折り畳み開閉

- za: 折り畳みの開閉（トグル）
- zM: すべての折り畳みを閉じる
- zR: すべての折り畳みを開く

## 囲み文字

1. 囲み文字を追加 (gsa)

- gsaiw" - カーソル位置の単語を""で囲む
- v (ビジュアルモード) → 選択 → gsa"

2. 囲み文字を削除 (gsd)
   操作: gsd"
3. 囲み文字を置換 (gsr)
   操作: gsr"'

## GitHubのリンクをコピーするキーバインド

- <leader>yg：ファイルのGitHub URLをコピー（ノーマルモード, oil）
- <leader>yl：行リンクをコピー（ノーマル/ビジュアル。選択範囲があると範囲リンク）

## gitのコミットや編集ユーザーなどの詳細を表示する

- <leader>hp: Hunkプレビュー（差分確認）
- <leader>hb: Blame詳細表示（コミット情報のポップアップ）

## page up/down

ctrlを押しながらfで1ページ進みます。forwardのfです。ctrlを押しながらbで 1ページ戻ります。

## 括弧

- ib/ab ()括弧
- iB/aB {}括弧

## ターミナル操作

### ウィンドウ間移動

- Ctrl+h: 左のウィンドウへ移動
- Ctrl+j: 下のウィンドウへ移動
- Ctrl+k: 上のウィンドウへ移動
- Ctrl+l: 右のウィンドウへ移動

※ ノーマルモード・ターミナルモード両方で使用可能

### toggletrm

- <leader>ts: 選択範囲をターミナルに送信
    - コマンド: ToggleTermSendVisualSelection
    - 使い方: ビジュアルモード（v）で範囲を選択してから実行
- <leader>tl: 選択行をターミナルに送信
    - コマンド: ToggleTermSendVisualLines
    - 使い方: ビジュアルラインモード（V）で行を選択してから実行

## scratch buffer

- <leader>+.

# Claude Code

## モデル切り替え

`cmd+option+p`

# zsh

## suspenndを戻すコマンド

`fg`
