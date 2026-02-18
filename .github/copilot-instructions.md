# Copilot Instructions for dotfiles Repository

このリポジトリは、macOS上の開発環境を統一されたカスタムダークテーマで管理する個人用dotfiles設定です。

## 言語設定

**すべての応答とコメントは日本語で行ってください。**

## リポジトリ構造

手動管理型のdotfilesリポジトリで、stowなどの自動化ツールは使用していません。各ツールの設定は独立したディレクトリで管理されており、`~/.config/` 配下に直接配置されます。

主要なディレクトリ:
- `nvim/` - Neovim (LazyVim) 設定
- `wezterm/` - WezTerm ターミナルエミュレーター設定
- `zsh/` - Zsh シェル設定
- `git/` - Git グローバル設定
- `lazygit/` - LazyGit TUI 設定
- `claude/`, `codex/`, `gemini/` - AIアシスタント設定
- `colors/` - カラーパレット定義

## カラーシステム（重要）

### 単一ソース原則

**`colors/abyssal-teal.toml`** が全ツールのカラー定義の唯一のソースです。色を変更または追加する際は、必ずこのファイルを更新してください。

### Abyssal Teal テーマ

- **背景**: `#0B0C0C` (純黒に近い)
- **メインテキスト**: `#CEF5F2` (明るい青緑)
- **アクセント**: Teal (`#6CD8D3`, `#64BBBE`) と Purple (`#8A99BD`, `#CED5E9`)
- **選択範囲**: `#64BBBE` (高視認性Teal)

### カラー変更のワークフロー

1. `colors/abyssal-teal.toml` を編集
2. 各設定ファイル (WezTerm, Neovim等) に反映
3. バリデーションを実行:
   ```bash
   cd scripts && go run validate-colors.go
   ```

詳細は `COLOR-SYSTEM.md` を参照してください。

## コミットメッセージ規約

Conventional Commitsスタイルを採用（日本語または英語）:

**形式:** `<type>(<scope>): <description>`

**タイプ:**
- `feat` - 新機能追加
- `fix` - バグ修正
- `perf` - パフォーマンス改善
- `refactor` - リファクタリング
- `docs` - ドキュメント変更
- `chore` - 設定変更、依存関係更新

**スコープ例:**
- `nvim`, `wezterm`, `zsh`, `git`, `lazygit`, `theme`, `ui`, `cursor`

**実例（最近のコミットから）:**
```
feat(nvim): Snacksターミナル設定を拡張し、toggletermを削除
fix(nvim): ターミナルの横並び表示を修正（stack=falseを設定）
perf(theme): ANSI色とWezTermの選択色を調整して視認性を改善
docs(cursor): ネットワークアクセス設定をallowlistからallに変更
chore(gitignore): Gemini CLI設定ディレクトリを除外
```

変更内容を具体的に記述し、「何を」「なぜ」変更したかが分かるようにしてください。

## 設定変更時の注意事項

### 1. キーバインドの追加・変更

キーバインドを追加または変更した場合は、`KEYMAPS.md` を更新してください。

### 2. Neovim プラグイン管理

LazyVimベースの設定で、プラグイン設定は `nvim/lua/plugins/` に分割されています。

**ディレクトリ構造:**
- `nvim/init.lua` - エントリポイント（IME設定、診断設定）
- `nvim/lua/config/` - コア設定
  - `lazy.lua` - lazy.nvimブートストラップ
  - `keymaps.lua` - カスタムキーマップ
  - `options.lua` - Neovimオプション
  - `autocmds.lua` - 自動コマンド
- `nvim/lua/plugins/` - プラグイン別設定（1ファイル1プラグイン）

**主要プラグイン:**
- `codecompanion.lua` - AI統合（Claude API経由のコード支援）
- `copilot.lua` - GitHub Copilot統合
- `lsp.lua` - Language Server Protocol設定
- `conform.lua` - フォーマッター設定（stylua、prettier等）
- `gitsigns.lua` - Git差分表示とBlame機能
- `oil.lua` - ファイラー（vim-like操作）
- `snacks.lua` - ターミナル、通知、その他UI機能

**プラグイン追加時の手順:**
1. `nvim/lua/plugins/` に新しいLuaファイルを作成（例: `my-plugin.lua`）
2. LazyVimのプラグイン定義形式でテーブルを返す
3. 必要に応じて `nvim/CLAUDE.md` を更新

### 3. AIアシスタント設定

複数のAIアシスタント向けドキュメントが存在します:

- `AGENTS.md` - 全AIエージェント共通の基本指針
- `CLAUDE.md` - Claude Code (claude.ai/code) 向け詳細情報
- `GEMINI.md` - Gemini CLI 向け情報
- `nvim/CLAUDE.md` - Neovim設定の詳細

AIアシスタント向けの重要な変更があれば、該当ファイルを更新してください。

## アーキテクチャの特徴

### モジュラー構造

各ツールが独立した設定ディレクトリを持ち、相互依存はありません。これにより、個別のツール設定を柔軟に変更できます。

### パフォーマンス重視

- Neovim: `updatetime=1000` で CPU 使用率を最適化
- WezTerm: GPU加速とマルチプレクサ機能
- 遅延ロードとモジュラープラグイン管理

### 日本語環境サポート

- Neovim: 日本語IMEサポート
- WezTerm: 日本語フォント設定
- ドキュメント: 日本語優先

## 重要なパスとファイル

### 設定ファイル
- **Neovim**: `nvim/init.lua`, `nvim/lua/config/`, `nvim/lua/plugins/`
  - フォーマッター: `nvim/stylua.toml` (Lua用)
  - プラグインロック: `nvim/lazy-lock.json`
- **WezTerm**: `wezterm/wezterm.lua`, `wezterm/keybinds.lua`
- **Zsh**: `zsh/.zshrc`
- **Starship**: `starship.toml`
- **Git**: `git/config`, `git/ignore`
- **LazyGit**: `lazygit/config.yml`
- **Cursor CLI**: `cursor/cli-config.json`

### カラーシステム
- **定義ファイル**: `colors/abyssal-teal.toml` (32色拡張パレット)
- **バリデーター**: `scripts/validate-colors.go`

### ドキュメント
- **AI指示**: `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, `nvim/CLAUDE.md`
- **キーマップ**: `KEYMAPS.md`
- **カラーガイド**: `COLOR-SYSTEM.md`
- **Cursor設定**: `cursor/SETTINGS.md`, `cursor/KEYMAPS.md`

## Git ワークフロー

- **メインブランチ**: `main` で直接作業
- **リモート**: `git@github.com:sunagawasei/dotfiles.git`
- 機能追加や大規模な変更は個別にコミット
- 定期的にリモートと同期

## テストとバリデーション

### カラー整合性チェック

カラー変更後は必ず整合性を検証してください：

```bash
cd scripts
go run validate-colors.go
```

**チェック内容:**
- `colors/abyssal-teal.toml` で定義された色が各設定ファイルで正しく使用されているか
- 未定義の色が使用されていないか（`#000000`, `#FFFFFF`, `#F2FFFF` は許可）
- コントラスト比が適切か

**依存パッケージ:**
- Go 1.16以上
- `github.com/pelletier/go-toml/v2` (go.mod管理)

### Neovim 設定チェック

**プラグイン管理:**
```vim
:Lazy               # プラグインマネージャーUI
:Lazy sync          # プラグインの同期・更新
:LazyHealth         # プラグインのヘルスチェック
```

**LSPとフォーマッター:**
```vim
:checkhealth        # 全体的な健全性チェック
:ConformInfo        # フォーマッター情報の表示
:LspInfo            # LSP接続状態の確認
```

### 設定リロード

- **Neovim**: `:source %` (現在のファイル) または `:source $MYVIMRC` (init.lua) または再起動
- **WezTerm**: 設定は自動リロード（`wezterm.lua`保存時）
- **Zsh**: `source ~/.zshrc` または新しいシェルを起動

## 関連ドキュメント

- `README.md` - リポジトリ概要
- `COLOR-SYSTEM.md` - カラーシステムの詳細
- `KEYMAPS.md` - キーバインド一覧
- `nvim/CLAUDE.md` - Neovim設定の詳細
- `cursor/SETTINGS.md` - Cursor CLI設定ガイド
