# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 適用コマンド

設定変更を反映するには、フレークルート (`~/.config`) から実行：

```bash
darwin-apply
```

> この home-manager 設定は nix-darwin の darwinModule として統合されているため、単独の `home-manager switch` は使わず、上記コマンドで適用する。

## モジュールの非自明な事情

- `git.nix`: `programs.delta` は lazygit の pager 専用（`enableGitIntegration=false`）。git 本体の差分は `diff.tool`/`difftool.hunk` に回す
- `hunk.nix`: パッケージ本体は flake input `hunk` の `homeManagerModules.default` を `nix-darwin/home_manager.nix` で import して提供する（`home.packages` への直接追加はしない）。`enableGitIntegration=true` で `core.pager` は hunk 側、`enableClaudeIntegration=true` で `~/.claude/skills/hunk-review` を自動リンクする
- `zsh.nix`: 最大モジュール。初期化順序の制御が入っているため下の「注意点」も参照

### パッケージ追加の判断基準

- 一般ツール → `packages.nix`
- 開発言語・ランタイム → `dev.nix`
- クラウド・インフラ → `cloud.nix`

## 注意点

- **fzf の ZSH 統合**: `fzf.nix` で `enableZshIntegration = false` を設定し、`zsh.nix` の Zinit (`fzf-tab`) で管理
- **Zsh 初期化順序**: `zsh.nix` では `lib.mkMerge` / `lib.mkOrder` で複数の `initContent` ブロックの実行順を制御
- **CLAUDECODE ガード**: `zsh.nix` に `[[ -n "$CLAUDECODE" ]]` チェックがあり、Claude Code 内では zoxide を無効化
- **カラーテーマ**: `colors/ghost-visor.toml` を single source of truth とし、`ls` は eza の `theme.yml`、zsh 補完リストは export しないシェルローカル変数で配色する。`LS_COLORS` は eza で `theme.yml` より優先されるため、zsh 初期化時に明示的に unset する
