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

### テスト実行（Neotest）
- **neotest**: テスト実行フレームワーク（VSCode Test Explorer相当）
- **neotest-golang**: Go言語テストアダプタ（サブテスト、テーブルテスト対応）

#### 基本テスト実行キーバインド：
- `<leader>tr`: カーソル位置のテストを実行（最も近いテスト）
- `<leader>tt`: 現在のファイルのすべてのテストを実行
- `<leader>tT`: プロジェクト全体のテストを実行
- `<leader>tl`: 最後に実行したテストを再実行
- `<leader>td`: カーソル位置のテストをデバッグ実行

#### 拡張テスト実行キーバインド：
- `<leader>tp`: パッケージ/ディレクトリ単位でテスト実行
- `<leader>tf`: 失敗したテストのみ再実行（詳細出力付き）
- `<leader>ta`: プロジェクト全体のテスト実行（詳細出力付き）
- `<leader>tA`: プロジェクト全体のテスト実行（レース検出付き）
- `<leader>tc`: 最寄りのテストをカバレッジ測定付きで実行
- `<leader>tC`: 現在ファイルをカバレッジ測定付きで実行

#### テスト結果・出力操作：
- `<leader>ts`: テスト結果サマリーを表示/非表示
- `<leader>to`: テスト出力を表示
- `<leader>tO`: テスト出力パネルを切り替え
- `<leader>tx`: 実行中のテストを停止
- `<leader>tq`: テストquickfixウィンドウを閉じる
- `<leader>t?`: テストステータスを開く
- `<leader>tw`: テストウォッチモード切り替え

#### Go言語テスト機能：
- **サブテスト実行**: `t.Run("subtest", ...)`の個別実行
- **テーブルテスト対応**: テストケース単位での実行
- **レース検出**: `-race`フラグによる並行処理バグ検出
- **カバレッジ測定**: `-cover`フラグによるコードカバレッジ
- **詳細出力**: `-v`フラグによるテスト実行詳細表示
- **Testifyサポート**: Testifyフレームワークのテストスイート対応
- **デバッグ統合**: DAP（Debug Adapter Protocol）連携

#### テスト結果の視覚表示：
- **インラインマーカー**: テスト関数の横に成功/失敗アイコン表示
- **カラーコード**: Vercel Geistカラーによる状態表示
  - ✓ 成功（緑）、✗ 失敗（赤）、◐ 実行中（青）、○ スキップ（黄）
- **サマリーパネル**: ツリー形式でのテスト構造・結果表示
- **出力パネル**: テスト実行ログとエラー詳細

### Git連携
- **gitsigns.nvim**: Git差分表示

### エディタ機能
- **toggleterm.nvim**: 統合ターミナル
  - `<C-\>`: ターミナルトグル
  - `<leader>tg`: LazyGit
  - `<leader>tu`: GitUI
  - `<leader>tb`: Btop
- **render-markdown.nvim**: Markdownプレビュー強化（Vercel Geistカラー適用）
- **flash.nvim**: 高速移動・検索
  - `s`: Flash jump（ラベル付きジャンプ）
  - `S`: Flash Treesitter（構文要素ジャンプ）
  - `f`/`F`: 前方/後方文字検索（`t`/`T`は無効化済み）
  - `;`/`,`: 文字検索の繰り返し/逆方向
  - **注意**: `t`/`T`キーは無効化（Neotestキーバインドとの競合回避）
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

### 基本的なデバッグコマンド
```vim
:checkhealth          # Neovim全体のヘルスチェック
:Lazy log             # プラグイン更新ログ
:messages             # エラーメッセージ履歴
```

### キーマップの競合確認
```vim
:map <leader>tr       # 特定のキーマップを確認
:nmap <leader>t       # ノーマルモードのキーマップを確認
:verbose map s        # キーマップの詳細情報を確認
```

### テスト実行の問題

#### "No tests found" エラーの対処法
```vim
:TSInstall go gomod gosum gowork  # Treesitterパーサーを再インストール
:checkhealth treesitter           # Treesitterの状態確認
:checkhealth neotest              # Neotestの診断
:lua print(vim.inspect(require("neotest").summary.get_adapters())) # アダプター確認
```

#### キーマップが動作しない場合
- `<leader>tr`が動作しない場合：Flash.nvimの`t`キー設定を確認
- 設定ファイル：`lua/plugins/flash-config.lua`で`t`/`T`キーを無効化済み
- 遅延読み込みが原因の場合：プラグインが読み込まれているか確認

#### Goテスト検出の問題
- **プロジェクトルート**: `go.mod`ファイルがプロジェクトルートに存在するか確認
- **ファイル名**: `*_test.go`パターンに従っているか確認
- **関数名**: `Test*`または`Benchmark*`で始まっているか確認
- **作業ディレクトリ**: Neovimをプロジェクトルートから起動しているか確認

#### デバッグコマンド
```vim
:messages                          # 起動時のデバッグメッセージを確認
:lua vim.print(require("neotest").state.get_adapters()) # アダプター一覧
:lua vim.print(require("neotest").state.adapter_id("neotest-golang")) # 特定アダプタID
:checkhealth neotest              # Neotestヘルスチェック
:Neotest summary                  # テストサマリーを表示
```

#### 起動時の確認事項
Neovimを起動後、`:messages`で以下のメッセージが表示されるかを確認：
```
neotest-golang loaded successfully
Neotest configured with 1 adapter(s)
Adapter 1: neotest-golang
```

メッセージが表示されない場合は、アダプタの読み込みに失敗しています。

#### 段階的な動作確認手順
1. **Neovimの再起動**
   ```bash
   # プロジェクトルートから起動（go.modがあるディレクトリ）
   cd /path/to/go/project
   nvim
   ```

2. **アダプタ読み込みの確認**
   ```vim
   :messages  # 起動メッセージを確認
   ```

3. **Goテストファイルを開く**
   ```vim
   :e *_test.go  # テストファイルを開く
   ```

4. **テスト関数にカーソルを移動**
   - `Test*`で始まる関数の中にカーソルを置く

5. **テスト実行**
   ```vim
   <leader>tr  # 最寄りのテストを実行
   ```

6. **結果確認**
   - 「No tests found」が表示されなければ成功
   - テストが実行され、結果が表示される

### プラグイン設定の確認
- Go言語サポート：`lua/config/lazy.lua`で`extras.lang.go`が有効
- テスト設定：`lua/plugins/neotest.lua`でneotest-golang設定
- Flash設定：`lua/plugins/flash-config.lua`でキーマップカスタマイズ
