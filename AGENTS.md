# AGENTS.md

macOS上の開発ツール設定を管理する個人用dotfiles。stow等の自動化ツールは使わず手動管理。

## 言語

すべての応答は日本語で行う。

## 標準と違う慣習

- 配色は `colors/ghost-visor.toml` が唯一のソース。wezterm・nvim・home-manager・lazygit・gh-dash・vim・statusline・`COLOR-SYSTEM.md` は `cd scripts && go run ./cmd/generate-colors` で生成する。派生側を直接編集しない
- Nix / home-manager の反映は `darwin-apply`（`sudo darwin-rebuild switch --flake ~/.config` のラッパー）
- 新規ファイルは `git add` してから rebuild する。flakeはgit追跡ファイルしか見ないため、未追跡だとNix評価が失敗する

## 記録の義務

- コミットは Conventional Commits（`<type>(<scope>): <subject>`）。本文の言語と粒度は `git log` の既存スタイルに合わせる

## 関連ドキュメント

- `CLAUDE.md` — Claude Code向けの指針
- `COLOR-SYSTEM.md` — カラーシステムのガイドライン
- `nvim/CLAUDE.md` — Neovim設定の詳細
