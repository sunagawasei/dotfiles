# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 適用コマンド

設定変更を反映するには、フレークルート (`~/.config`) から実行：

```bash
darwin-rebuild switch --flake ~/.config#CA-20021145
```

> この home-manager 設定は nix-darwin の darwinModule として統合されているため、単独の `home-manager switch` は使わず、上記コマンドで適用する。

## アーキテクチャ

フレーク構成 (`~/.config/flake.nix`):
- **nix-darwin** (hostname: `CA-20021145`) + **home-manager** + **nix-homebrew** を統合
- nix-darwin 設定: `~/.config/nix-darwin/configuration.nix`
- home-manager の統合: `~/.config/nix-darwin/home_manager.nix` が `../home-manager/home.nix` を参照

### モジュール構成

`home.nix` がエントリポイントで、以下の7モジュールを import：

| ファイル | 役割 |
|----------|------|
| `packages.nix` | 一般ツール群 + `programs.direnv` (nix-direnv) |
| `dev.nix` | 言語ツールチェーン (Go, Node.js 22, pnpm, Python/uv) |
| `git.nix` | Git 設定 + `programs.delta` |
| `shell.nix` | 環境変数・PATH (XDG Base Dir, EDITOR, CLAUDE_CONFIG_DIR 等) |
| `zsh.nix` | Zinit プラグイン管理、Pure プロンプト、カラー設定 (304行、最大モジュール) |
| `fzf.nix` | fzf 有効化のみ (ZSH 統合は無効化) |
| `cloud.nix` | クラウド/インフラツール (awscli2, google-cloud-sdk, terraform, kubectl 等) |

### パッケージ追加の判断基準

- 一般ツール → `packages.nix`
- 開発言語・ランタイム → `dev.nix`
- クラウド・インフラ → `cloud.nix`

## 注意点

- **fzf の ZSH 統合**: `fzf.nix` で `enableZshIntegration = false` を設定し、`zsh.nix` の Zinit (`fzf-tab`) で管理
- **Zsh 初期化順序**: `zsh.nix` では `lib.mkMerge` / `lib.mkOrder` で複数の `initContent` ブロックの実行順を制御
- **CLAUDECODE ガード**: `zsh.nix` に `[[ -n "$CLAUDECODE" ]]` チェックがあり、Claude Code 内では zoxide を無効化
- **カラーテーマ**: `zsh.nix` の Pure プロンプト・LS_COLORS・FZF_DEFAULT_OPTS・ZSH_HIGHLIGHT_STYLES は Abyssal Teal テーマで統一 (`colors/abyssal-teal.toml` が single source of truth)
