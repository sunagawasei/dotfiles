# Copilot Instructions for dotfiles Repository

macOS上の開発ツール設定を管理する個人用dotfiles。stow等の自動化ツールは使わず手動管理。

## 言語設定

すべての応答とコメントは日本語で行う。

## カラーシステム

`colors/ghost-visor.toml` が全ツールのカラー定義の唯一のソース。**個々の色値をこのファイルへ転記しない**（転記は必ず陳腐化する）。実値は常に TOML と `COLOR-SYSTEM.md` を読む。

派生ファイル（wezterm・nvim・home-manager・lazygit・gh-dash・vim・`claude/statusline.sh`・`COLOR-SYSTEM.md`）は生成物なので直接編集しない。TOMLを編集してから生成と検証を流す:

```bash
cd scripts
go run ./cmd/generate-colors
go run ./cmd/generate-colors --check
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-colors
```

色を1つも変えなくても、色を含むファイルの行が増減しただけで inventory は drift する。色以外の理由でそれらを編集したときも上記を流す。

## Nix / home-manager

反映は `darwin-apply`（`sudo darwin-rebuild switch --flake ~/.config` のラッパー）。新規ファイルは `git add` してから rebuild する。flakeはgit追跡ファイルしか見ないため、未追跡だとNix評価が失敗する。

## コミットメッセージ規約

Conventional Commits（`<type>(<scope>): <description>`）。type は `feat` / `fix` / `perf` / `refactor` / `docs` / `chore`、scope は `nvim` / `wezterm` / `theme` / `zsh` / `git` / `scripts` / `docs` / `claude` など。本文の言語と粒度は `git log` の既存スタイルに合わせる。

## 記録の義務

キーバインドや操作を追加・変更したら `KEYMAPS.md` に記録する。

## 関連ドキュメント

- `AGENTS.md` — 全AIエージェント共通の指針
- `CLAUDE.md` — Claude Code向けの指針
- `COLOR-SYSTEM.md` — カラーシステムの詳細
- `KEYMAPS.md` — キーバインド一覧
- `nvim/CLAUDE.md` — Neovim設定の詳細
