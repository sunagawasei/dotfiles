# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

LazyVimベースのNeovim設定。個人用にカスタマイズされており、日本語コメントを含む。

## アーキテクチャ

### コアフレームワーク
- **LazyVim**: 事前設定済みのNeovimセットアップフレームワーク
- **プラグインマネージャー**: lazy.nvim
- **起動エントリポイント**: `init.lua` → `lua/config/lazy.lua`

### ディレクトリ構造
```
nvim/
├── init.lua              # メインエントリポイント（IME設定、診断設定）
├── lua/
│   ├── config/
│   │   ├── lazy.lua      # lazy.nvimブートストラップとプラグイン読み込み
│   │   ├── keymaps.lua   # カスタムキーマップ
│   │   ├── options.lua   # Neovimオプション設定
│   │   └── autocmds.lua  # 自動コマンド設定
│   ├── plugins/          # 個別プラグイン設定
│   └── utils/            # ユーティリティ関数
├── stylua.toml           # Luaフォーマッター設定
└── lazy-lock.json        # プラグインバージョンロック
```

### プラグイン管理

プラグインは`lua/plugins/`に個別ファイルとして配置：
- LazyVimのデフォルトプラグインを自動的に継承
- `lazyvim.plugins.extras.ai.copilot`エクストラを有効化
- カスタムプラグインは各ファイルでテーブルを返す形式で定義

## 開発コマンド

### Neovim内での操作
```vim
:Lazy               # プラグインマネージャーUI
:Lazy sync          # プラグインの同期・更新
:Lazy update        # プラグインの更新のみ
:LazyHealth         # プラグインのヘルスチェック
:ConformInfo        # フォーマッター情報の表示
```

### コードフォーマット
Conform.nvimで管理。設定されているフォーマッター：
- Lua: `stylua` （設定: stylua.toml）
- Python: `black`, `isort`
- JavaScript/TypeScript: `prettier`
- その他: `lua/plugins/conform.lua`参照

フォーマット実行:
- `<leader>cf`: 現在のファイルをフォーマット（LazyVimデフォルト）

### LSP操作
基本的なLSPナビゲーション：
- `gd`: 定義へジャンプ
- `gr`: 参照箇所を検索（関数が使われている場所を表示）
- `gI`: 実装へジャンプ
- `gy`: 型定義へジャンプ
- `gD`: 宣言へジャンプ
- `K`: ホバー情報表示（定義・ドキュメントをフローティングウィンドウで表示）
- `gK`: シグネチャヘルプ

診断関連：
- `<leader>cd`: 診断を表示
- `gl` / `<leader>ld`: 行の診断をフローティングで表示
- `[d` / `]d`: 前/次の診断へ移動
- `[e` / `]e`: 前/次のエラーへ移動

コード操作：
- `<leader>ca`: コードアクション
- `<leader>cr`: リネーム
- `<leader>cl`: LSP情報表示

VSCode互換：
- VSCodeの`gh`（定義をフローティング表示）は、Neovimでは`K`キーが同等の機能を提供

### ファイル検索・エクスプローラー
Snacks.nvimベース：
- `<leader>fe`: ファイルエクスプローラーを開く
- `<leader>ff`: ファイル検索
- 隠しファイルとgitignoreファイルもデフォルトで表示

## 設定の特徴

### 日本語環境対応
- macOSでのIME自動切り替え（InsertLeave時に英数入力へ）
- UTF-8エンコーディング設定
- 日本語コメントを含むコードベース

### 診断表示
- 赤い波線（underline）を無効化
- インライン診断テキスト（virtual_text）を無効化
- サインカラムの診断アイコンは表示

### カラースキーム
- Vercelテーマを使用（`install.colorscheme = { "vercel" }`）

## カスタマイズ時の注意

1. **プラグイン追加**: `lua/plugins/`に新しいLuaファイルを作成
2. **既存プラグインの設定変更**: 同名のspecでオーバーライド
3. **フォーマッター追加**: `lua/plugins/conform.lua`の`formatters_by_ft`に追加
4. **キーマップ追加**: `lua/config/keymaps.lua`に記述

## 主要プラグイン

### コアプラグイン
- **LazyVim**: Neovimの設定フレームワーク
- **lazy.nvim**: プラグインマネージャー
- **snacks.nvim**: ファイルエクスプローラー、ピッカーなどのユーティリティ集

### テーマ・UI
- **vercel.nvim**: Vercel Geistカラーテーマ（メイン）
- **bufferline.nvim**: タブライン表示
- **lualine.nvim**: ステータスライン
- **noice.nvim**: コマンドライン、メッセージ、ポップアップのUI改善
- **no-neck-pain.nvim**: 中央寄せ表示モード（`<leader>nn`）

### コーディング支援
- **copilot.vim**: GitHub Copilot（AIコード補完、Tabキーで補完）
- **CopilotChat.nvim**: Copilotとの対話型チャット（日本語対応）
  - `<leader>cc`: チャットを開く
  - `<leader>cE`: コード説明（選択モード）
  - `<leader>cR`: コードレビュー（選択モード）
  - **選択範囲の渡し方**:
    - ビジュアルモード（`v`/`V`/`<C-v>`）で選択 → キーマップ実行
    - 選択がない場合は自動的にバッファ全体が対象
  - **ファイル・コンテキストの渡し方**（チャット内で使用）:
    - `#file:path/to/file` - 特定ファイルを追加
    - `#buffer:current` - 現在のバッファ
    - `#buffers:visible` - 表示中の全バッファ
    - `#gitdiff` - Git差分
    - `#diagnostics:current` - 診断情報
    - `#glob:*.lua` - パターンマッチするファイル一覧
    - `#grep:TODO` - ワークスペース内検索
- **claudecode.nvim**: Claude Code統合
  - `<leader>ac`: Claude Codeトグル
  - `<leader>ar`: 前回のセッションを再開
- **blink.cmp**: 補完エンジン（LazyVimデフォルト）
- **conform.nvim**: コードフォーマッター
- **nvim-lint**: リンター統合
- **todo-comments.nvim**: TODOコメントのハイライト

### Git連携
- **gitsigns.nvim**: Git差分表示
- **gitgraph.nvim**: Gitグラフビューア（`<leader>gl`）
- **diffview.nvim**: 差分表示ツール

### エディタ機能
- **toggleterm.nvim**: 統合ターミナル
  - `<C-\>`: ターミナルトグル
  - `<leader>tg`: LazyGit
  - `<leader>tu`: GitUI
  - `<leader>tb`: Btop
- **render-markdown.nvim**: Markdownプレビュー強化（Vercel Geistカラー適用）
- **flash.nvim**: 高速移動・検索
- **which-key.nvim**: キーバインドヘルプ表示
- **nvim-window-picker**: ウィンドウ選択ツール（`<leader>wp`）
- **trim.nvim**: 行末空白の自動削除
- **mini.ai**: テキストオブジェクト拡張
- **mini.pairs**: 括弧の自動ペアリング
- **nvim-ts-autotag**: HTMLタグの自動閉じ

### LSP・構文解析
- **nvim-lspconfig**: LSP設定
- **mason.nvim**: LSPサーバー管理
- **nvim-treesitter**: 構文解析・ハイライト
- **nvim-treesitter-textobjects**: Treesitterベースのテキストオブジェクト
- **lazydev.nvim**: Neovim Lua開発支援
- **ts-comments.nvim**: コメントの構文認識

### ユーティリティ
- **plenary.nvim**: Neovimプラグイン用ライブラリ
- **nui.nvim**: UIコンポーネントライブラリ
- **persistence.nvim**: セッション管理
- **trouble.nvim**: 診断リスト表示
- **grug-far.nvim**: 検索・置換ツール

各プラグインの詳細設定は`lua/plugins/`ディレクトリ内の個別ファイルで管理されています。

## トラブルシューティング

```vim
:checkhealth          # Neovim全体のヘルスチェック
:Lazy log             # プラグイン更新ログ
:messages             # エラーメッセージ履歴
```