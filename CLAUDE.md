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

# nix-darwin / home-manager の反映
sudo darwin-rebuild switch --flake ~/.config#CA-20021145
```

### darwin-rebuild の既知事象

- **新規ファイル（`home-manager/patches/` 等）は `git add` してから rebuild する** — flakeはgit追跡ファイルしか見ないため、未追跡だと "Path ... is not tracked by Git" でNix評価が失敗する
- brew bundle の untrusted tap 警告（社内tap `cycloud-io/*` 等）と `--cleanup` 非推奨警告は**無害・無視してよい**

## コード規約

**コミットメッセージ**: Conventional Commits形式 `<type>(<scope>): <subject>`

- **type**: `feat` / `fix` / `docs` / `refactor` / `perf` / `chore`
- **scope**: `nvim` / `wezterm` / `theme` / `zsh` / `fish` / `git` / `scripts` / `docs` / `claude`
