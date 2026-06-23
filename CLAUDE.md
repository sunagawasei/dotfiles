# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfilesリポジトリ。stowなどの自動化ツールは使わず手動管理。

- GitHub: `git@github.com:sunagawasei/dotfiles.git` / メインブランチ: `main`

## アーキテクチャ

### カラーシステム（単一ソース原則）

`colors/abyssal-teal.toml` が **唯一の真実のソース**。全アプリの配色はここから派生する（wezterm / nvim / lazygit / zsh）。詳細は `COLOR-SYSTEM.md`。

### Neovim設定（LazyVim）

`nvim/init.lua` → `nvim/lua/config/lazy.lua` がエントリポイント。プラグインは `nvim/lua/plugins/` に1プラグイン1ファイルで配置。詳細は `nvim/CLAUDE.md`。

## 開発コマンド

```bash
# カラーバリデーション（カラー関連ファイル変更後は必須）
cd scripts && go run validate-colors.go

# Raycast拡張機能
cd raycast/extensions/<name> && npm run lint && npm run build
```

## コード規約

**コミットメッセージ**: Conventional Commits形式 `<type>(<scope>): <subject>`

- **type**: `feat` / `fix` / `docs` / `refactor` / `perf` / `chore`
- **scope**: `nvim` / `wezterm` / `theme` / `zsh` / `fish` / `git` / `scripts` / `docs` / `claude`
