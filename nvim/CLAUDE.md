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

### LSP操作・ファイル検索
LSPナビゲーション（`gd`定義ジャンプ等）、診断表示、コードアクション、Snacks.nvimベースのファイル検索・エクスプローラー（`<leader>fe`/`<leader>ff`）を提供。キーバインド詳細は`../KEYMAPS.md`参照。

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
- カスタムカラースキームを使用（プラグイン依存なし）

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
- **カスタムカラースキーム**: 独自のカラーテーマ（プラグイン依存なし）
- **bufferline.nvim**: タブライン表示
- **lualine.nvim**: ステータスライン
- **noice.nvim**: コマンドライン、メッセージ、ポップアップのUI改善
- **no-neck-pain.nvim**: 中央寄せ表示モード（`<leader>nn`）

### コーディング支援
- **copilot.vim**: GitHub Copilot（AIコード補完、Tabキーで補完）
- **CopilotChat.nvim**: Copilotとの対話型チャット（日本語対応）。キーバインド・コンテキスト指定構文は`../KEYMAPS.md`参照
- **claudecode.nvim**: Claude Code統合。キーバインドは`../KEYMAPS.md`参照
- **blink.cmp**: 補完エンジン（LazyVimデフォルト）
- **conform.nvim**: コードフォーマッター
- **nvim-lint**: リンター統合
- **todo-comments.nvim**: TODOコメントのハイライト

### テスト実行（Neotest）
- **neotest**: テスト実行フレームワーク（VSCode Test Explorer相当）
- **neotest-golang**: Go言語テストアダプタ（サブテスト、テーブルテスト、レース検出、カバレッジ、DAP連携対応）

キーバインド一覧は`../KEYMAPS.md`参照。「No tests found」等のトラブルシューティングは`../.claude/docs/nvim-troubleshooting.md`参照。

### Git連携
- **gitsigns.nvim**: Git差分表示

### エディタ機能
- **toggleterm.nvim**: 統合ターミナル（最も人気のあるプラグイン）。番号付きターミナル、方向切り替え、サイズ変更、REPL選択範囲送信、LazyGit統合など。キーバインドは`../KEYMAPS.md`参照
- **render-markdown.nvim**: Markdownプレビュー強化（カスタムカラー適用）
- **which-key.nvim**: キーバインドヘルプ表示
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

### 基本的なデバッグコマンド
```vim
:checkhealth          # Neovim全体のヘルスチェック
:Lazy log             # プラグイン更新ログ
:messages             # エラーメッセージ履歴
```

キーマップ競合確認は`:map`/`:nmap`/`:verbose map`。Neotestの「No tests found」等、テスト実行関連の詳細なトラブルシューティング手順は`../.claude/docs/nvim-troubleshooting.md`参照。
