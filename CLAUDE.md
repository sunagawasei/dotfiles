# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfilesリポジトリ。stowなどの自動化ツールは使わず手動管理。

## アーキテクチャ

### カラーシステム（単一ソース原則）

`colors/ghost-visor.toml` が **唯一の真実のソース**。全アプリの配色はここから派生する（wezterm / nvim / lazygit / zsh）。詳細は `COLOR-SYSTEM.md`。

カラー関連ファイルを変更したら `color-validation` skill の検証手順を必ず通す。

### herdrのローカルパッチ

herdr(ターミナルマルチプレクサ)は flake input のソースへローカルパッチを当てて配備する。パッチ、`home-manager/patches/`、`home-manager/herdr.nix` を触るときは `claude/skills/herdr-deploy/SKILL.md` の配備手順を必ず通す。commitしただけでは稼働バイナリは変わらない。

## 開発コマンド

```bash
# Claude Code hooks（Goソース変更後は必ず両方）
cd claude/hooks && go test ./... && go build -o . ./...

# nix-darwin / home-manager の反映（ユーザー・Claude共通。Claudeも実行できる）
darwin-apply
```

`darwin-apply` は `sudo darwin-rebuild switch --flake ~/.config` のラッパー。flake attrを
省略しているため、hostname（`scutil --get LocalHostName`）でdarwin-rebuildが解決する。
`Bash(sudo:*)` denyに掛からない名前にすることでClaudeにも開放している（詳細は`nix-darwin/CLAUDE.md`）。

### darwin-rebuild の既知事象

- **新規ファイル（`home-manager/patches/` 等）は `git add` してから rebuild する** — flakeはgit追跡ファイルしか見ないため、未追跡だと "Path ... is not tracked by Git" でNix評価が失敗する
- `--cleanup` 非推奨警告は無害・無視してよい。**brew bundle の untrusted tap `Error: Refusing to load formula ...`は無害ではない** — `darwin-rebuild switch`全体をそこで中断させ、home-manager側の変更（activation）が一切適用されないまま終わる（2026-07-07実例）。`/run/current-system`のリンク先ハッシュが更新されているかで検知できる。`brew trust <tap>`（または`brew trust --formula <formula>`）で解消してから再実行する
- **Pkg型のcaskは先に手動インストールしてから`casks`へ追加する** — 未インストール状態だとactivationの非対話`brew bundle`がsudoプロンプトで止まる。`brew install --cask <name>`を対話実行して通しておけば、以降のactivationは既インストールのno-op扱いになる。判別は`brew info --cask --json=v2 <name>`のartifactsに`pkg`があるか（該当: `karabiner-elements` / `azookey` / `microsoft-remote-desktop`）
- **sudo不要で評価・ビルドのみ検証できる**: `darwin-rebuild build --flake ~/.config` — システムには適用せず、Nix評価エラーやビルド失敗を先に潰せる。実際に適用する前の下調べに使う
- **commit前のbuildとcommit後のbuild/switchではsystemハッシュが変わる**（flake self revがdarwin-version.jsonに焼き込まれるため。dirty treeとcommit済みtreeでも変わる）。パッケージ反映の検証はsystemハッシュの一致比較ではなく `nix-store -qR /run/current-system | grep <pkg>` → `nix-store -q --deriver <path>` でパッケージ単位のderivationを確認する。単に反映有無を見るだけなら `command -v <cmd>` + `realpath` がより直接的（`useUserPackages = true` でもユーザーパッケージは`/run/current-system`のclosureに含まれる。「含まれないので前者は使えない」という指摘は実測で誤り）

## コード規約

**コミットメッセージ**: Conventional Commits形式 `<type>(<scope>): <subject>`

- **type**: `feat` / `fix` / `docs` / `refactor` / `perf` / `chore`
- **scope**: `nvim` / `wezterm` / `theme` / `zsh` / `fish` / `git` / `scripts` / `docs` / `claude`
