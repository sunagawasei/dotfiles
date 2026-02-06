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

# Claude Code

## モデル切り替え

`cmd+option+p`

# zsh

## suspenndを戻すコマンド

`fg`
