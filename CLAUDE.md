# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfilesリポジトリ。stowなどの自動化ツールは使わず手動管理。

- GitHub: `git@github.com:sunagawasei/dotfiles.git` / メインブランチ: `main`

## 重要なコマンド

### カラー整合性バリデーション

カラー関連ファイルを変更した後は必ず実行：

```bash
cd scripts && go run validate-colors.go
```

### Raycast拡張機能

```bash
cd raycast/extensions/<extension-name>
npm run lint
npm run build
```

### Neovim内操作（詳細は `nvim/CLAUDE.md`）

```vim
:Lazy sync          # プラグイン同期
:checkhealth        # ヘルスチェック
:ConformInfo        # フォーマッター情報
```

## アーキテクチャ

### カラーシステム（単一ソース原則）

`colors/abyssal-teal.toml` が **唯一の真実のソース**。全アプリの配色はここから派生する。

```
colors/abyssal-teal.toml  ←── 変更はここだけ
  ↓
wezterm/wezterm.lua        # ANSI 16色、背景・前景色
nvim/lua/plugins/colorscheme.lua  # 構文ハイライト、UI要素
lazygit/config.yml         # テーマカラー
zsh/.zshrc                 # プロンプト・シンタックスハイライト
```

カラー変更後は必ず `scripts/validate-colors.go` でバリデーション実行。詳細は `COLOR-SYSTEM.md`。

### Neovim設定（LazyVim）

`nvim/init.lua` → `nvim/lua/config/lazy.lua` がエントリポイント。プラグインは `nvim/lua/plugins/` に1プラグイン1ファイルで配置。詳細は `nvim/CLAUDE.md`。

### scriptsディレクトリ

Goで書かれたユーティリティスクリプト群（カラーバリデーション、コントラストチェック等）。`go.mod`/`go.sum` で管理。

## コミットメッセージ規約

Conventional Commits形式：`<type>(<scope>): <subject>`

**type**: `feat` / `fix` / `docs` / `refactor` / `perf` / `chore`

**scope**: `nvim` / `wezterm` / `theme` / `zsh` / `fish` / `git` / `scripts` / `docs`

```
feat(nvim): CodeCompanionプラグインを追加
fix(wezterm): キーバインドの競合を解消
feat(theme): 背景色を#0B0C0Cに変更
```

Claude Codeによる変更には `Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>` を追加。

## スキル

- `/config-change`: 設定変更→テスト→コミット→プッシュの標準ワークフロー
- `/color-validation`: カラー整合性検証の詳細手順

## 言語選択

スクリプト・CLIツールは **Go言語を優先**。Python/Bash/Node.jsはユーザーが明示的に指定した場合のみ。

## 関連ドキュメント

- `nvim/CLAUDE.md` - Neovim詳細（プラグイン一覧、キーマップ、トラブルシューティング）
- `COLOR-SYSTEM.md` - カラーパレット全定義と色選択ルール
- `KEYMAPS.md` - キーバインドリファレンス
- `.claude/rules/` - ファイル種別ごとの規約（color-system.md, golang.md, neovim.md）
