# Neovim キーマップ一覧

このドキュメントは、Neovim設定で定義されているカスタムキーマップをまとめたものです。

## 基本操作

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<C-s>` | ファイルを保存 | n, i, v |
| `<C-c>` | ファイル全体をクリップボードにコピー | n |

## ファイルエクスプローラー

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>e` | エクスプローラーを開く/フォーカス/閉じる（ルートディレクトリ） | n |
| `<leader>E` | エクスプローラーを開く/フォーカス/閉じる（現在のディレクトリ） | n |
| `<leader>fe` | ファイルエクスプローラー（LazyVimデフォルト） | n |

## コードフォーマット

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>fm` | ファイルまたは選択範囲をフォーマット | n, v |
| `<leader>cf` | コードフォーマット（LazyVimデフォルト） | n, v |

## 診断（Diagnostics）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `gl` | 現在行の診断をフローティングウィンドウで表示 | n |
| `<leader>ld` | 現在行の診断を表示 | n |
| `<leader>cd` | 診断を表示（LazyVimデフォルト） | n |
| `[d` | 前の診断へ移動 | n |
| `]d` | 次の診断へ移動 | n |
| `[e` | 前のエラーへ移動 | n |
| `]e` | 次のエラーへ移動 | n |
| `<leader>dl` | 診断をロケーションリストに追加 | n |
| `<leader>dq` | 診断をQuickfixリストに追加 | n |

## LSP ナビゲーション（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `gd` | 定義へジャンプ | n |
| `gr` | 参照箇所を検索 | n |
| `gI` | 実装へジャンプ | n |
| `gy` | 型定義へジャンプ | n |
| `gD` | 宣言へジャンプ | n |
| `K` | ホバー情報表示（定義・ドキュメント） | n |
| `gK` | シグネチャヘルプ | n |
| `<leader>ca` | コードアクション | n, v |
| `<leader>cr` | リネーム | n |
| `<leader>cl` | LSP情報表示 | n |

## ターミナル

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<C-\>` | ターミナルをトグル | n, t |
| `<leader>tf` | フローティングターミナル | n |
| `<leader>th` | 水平分割ターミナル | n |
| `<leader>tv` | 垂直分割ターミナル | n |
| `<leader>tg` | LazyGitを開く | n |
| `<leader>tu` | GitUIを開く | n |
| `<leader>tb` | Btopを開く | n |

## Git

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>gl` | Gitグラフを表示 | n |
| `<leader>gg` | LazyGit（LazyVimデフォルト） | n |
| `<leader>gG` | LazyGit（現在のファイル履歴） | n |
| `<leader>gb` | Git blame | n |
| `<leader>gB` | Git browse | n |

## AI アシスタント

### GitHub Copilot
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<Tab>` | Copilotの提案を受け入れ | i |
| `<leader>cc` | CopilotChatを開く | n |
| `<leader>ct` | CopilotChatをトグル | n |
| `<leader>cs` | CopilotChatを停止 | n |
| `<leader>cr` | CopilotChatをリセット | n |
| `<leader>cE` | コードを説明（選択範囲） | v |
| `<leader>cR` | コードレビュー（選択範囲） | v |
| `<leader>cF` | 修正提案（選択範囲） | v |
| `<leader>cO` | 最適化提案（選択範囲） | v |
| `<leader>cD` | ドキュメント生成（選択範囲） | v |
| `<leader>cT` | テストケース生成（選択範囲） | v |

### Claude Code
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>ac` | Claude Codeをトグル | n |
| `<leader>af` | Claude Codeにフォーカス | n |
| `<leader>ar` | 前回のセッションを再開 | n |
| `<leader>aC` | セッションを継続 | n |
| `<leader>ab` | 現在のバッファを追加 | n |
| `<leader>as` | 選択範囲をClaude Codeに送信 | v |
| `<leader>aa` | Diffを承認 | n |
| `<leader>ad` | Diffを拒否 | n |

## Markdown

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>mp` | Markdownプレビューをトグル | n |
| `<leader>ms` | Markdownプレビューを停止 | n |

## ウィンドウ管理

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>wp` | ウィンドウピッカーで移動 | n |
| `<leader>nn` | No Neck Pain（中央寄せ表示）をトグル | n |
| `<C-h>` | 左のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-j>` | 下のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-k>` | 上のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-l>` | 右のウィンドウへ移動（LazyVimデフォルト） | n |
| `<leader>w-` | 水平分割（LazyVimデフォルト） | n |
| `<leader>w\|` | 垂直分割（LazyVimデフォルト） | n |
| `<leader>wd` | ウィンドウを削除（LazyVimデフォルト） | n |

## 検索・置換（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>ff` | ファイル検索 | n |
| `<leader>fg` | Grep検索 | n |
| `<leader>fb` | バッファ検索 | n |
| `<leader>fh` | ヘルプタグ検索 | n |
| `<leader>fc` | 設定ファイル検索 | n |
| `<leader>fk` | キーマップ検索 | n |
| `<leader>fr` | 最近開いたファイル | n |
| `<leader>sr` | 検索と置換（Grug-far） | n, v |

## バッファ操作（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<S-h>` | 前のバッファ | n |
| `<S-l>` | 次のバッファ | n |
| `[b` | 前のバッファ | n |
| `]b` | 次のバッファ | n |
| `<leader>bb` | バッファを切り替え | n |
| `<leader>bd` | バッファを削除 | n |
| `<leader>bD` | バッファを強制削除 | n |

## その他のユーティリティ

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>l` | Lazy（プラグインマネージャー） | n |
| `<leader>L` | LazyVim Changelog | n |
| `<leader>qq` | 終了 | n |
| `<leader>ui` | UIのインデントガイドをトグル | n |
| `<leader>ul` | 行番号をトグル | n |
| `<leader>uL` | 相対行番号をトグル | n |
| `<leader>uw` | 折り返しをトグル | n |

## ノート

- `<leader>` キーはデフォルトでスペースキーに設定されています
- `n` = Normal mode, `i` = Insert mode, `v` = Visual mode, `t` = Terminal mode
- LazyVimのデフォルトキーマップの詳細は[公式ドキュメント](https://www.lazyvim.org/keymaps)を参照してください
