# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 適用コマンド

設定変更を反映するには、フレークルート (`~/.config`) から実行：

```bash
darwin-apply
```

`darwin-apply`（`../home-manager/packages.nix`）は `sudo darwin-rebuild switch --flake ~/.config`
のラッパー。flake attrを省略しているため hostname（`scutil --get LocalHostName`）で解決される。
**Claudeもこれを実行できる**: permission ruleは deny→ask→allow の順で評価され specificity が
効かないため、`Bash(sudo:*)` deny を残したまま sudo を含む形で例外は作れない。sudo を含まない
コマンド名にして `Bash(darwin-apply)` のみ allow することで、他の sudo は一律拒否のまま維持している。
引数は受け取らない（渡すと exit 2）。sudoers 側の NOPASSWD は `configuration.nix` の
`security.sudo.extraConfig` で呼び出し形の完全一致のみ許可している。

`zsh.nix` のエイリアス：
- `nupdate` — `nix flake update` + `darwin-apply`（パッケージ更新 + 適用）

> home-manager は nix-darwin の darwinModule として統合されているため、単独の `home-manager switch` は使わない。

## アーキテクチャ

フレーク定義 (`~/.config/flake.nix`) が起点：

```
~/.config/flake.nix
  └── darwinConfigurations."<hostname>"    ← 現在は "CA-20038442"（PC交換時は改名する）
        ├── ./nix-darwin/configuration.nix        ← システム設定（このディレクトリ）
        ├── home-manager.darwinModules             ← home_manager.nix → ../home-manager/home.nix
        └── nix-homebrew.darwinModules             ← homebrew.nix
```

| ファイル | 役割 |
|----------|------|
| `../flake.nix` | inputs (nixpkgs-unstable, home-manager, nix-darwin, nix-homebrew) / outputs 定義 |
| `configuration.nix` | macOS システム設定（Touch ID sudo、Finder、zsh `ZDOTDIR` 初期化、unfree 許可リスト） |
| `home_manager.nix` | home-manager モジュール統合（`useGlobalPkgs`/`useUserPackages`、エントリポイント指定） |
| `homebrew.nix` | nix-homebrew 設定、brews/casks リスト、GitHub トークン注入スクリプト |

## パッケージ追加の判断基準

- **Nix パッケージ（一般ツール・言語）** → `../home-manager/` 配下の適切なモジュール（詳細は `../home-manager/CLAUDE.md`）
- **Homebrew cask（GUI アプリ）** → `homebrew.nix` の `casks` リスト
- **Homebrew brew（nixpkgs にない CLI）** → `homebrew.nix` の `brews` リスト

## 注意点

- **Homebrew GitHub トークン**: `homebrew.nix` には `activationScript` のオーバーライドがある。`nix-darwin` の activation は `#!/usr/bin/env -i bash` で環境変数を消去するため、`HOMEBREW_GITHUB_API_TOKEN` を直接渡せない。`gh auth token` をユーザーセッションから動的取得して注入する回避策を実装済み（触らない）。
- **`nix.enable = false`**: Nix デーモン管理は nix-darwin に委ねず別管理。
- **アーキテクチャ**: `aarch64-darwin`（Apple Silicon）固定。
