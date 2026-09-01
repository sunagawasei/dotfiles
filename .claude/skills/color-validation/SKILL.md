---
name: color-validation
description: colors/ghost-visor.tomlを単一ソースとしたカラー生成、生成drift確認、use-site pair・未知HEX・CVD統合検証の手順
---

# カラー整合性バリデーション

このスキルは、`colors/ghost-visor.toml`を単一ソースとして、全設定ファイルへのカラー生成と整合性検証を行う手順を定義します。

## 概要

カラーシステムは**生成方式**です。`colors/ghost-visor.toml`を編集し、`scripts/cmd/generate-colors/main.go`を実行すると、WezTerm・Neovim・home-manager・LazyGit・vim・statusline・COLOR-SYSTEM.mdの派生ファイルが一括で更新されます。手書きで複数ファイルを同期する運用ではありません。

## 主ワークフロー

### 1. 単一ソースを編集

```bash
nvim colors/ghost-visor.toml
```

### 2. 生成を実行

```bash
cd scripts && go run ./cmd/generate-colors
cd scripts && go run ./cmd/generate-color-inventory
```

これにより以下が生成・更新されます：

- `nvim/lua/config/palette.lua`（丸ごと生成）
- `wezterm/colors.lua`（丸ごと生成）
- `home-manager/colors.nix`（丸ごと生成）
- `lazygit/config.yml`（マーカーブロック置換）
- `vim/vimrc`（cterm256近似値を自動計算してマーカーブロック置換）
- `claude/statusline.sh`（HEXを10進RGBへ自動変換してマーカーブロック置換）
- `COLOR-SYSTEM.md`（パレット表をマーカーブロック置換）

### 3. 生成漏れがないことを確認

```bash
cd scripts && go run ./cmd/generate-colors --check
cd scripts && go run ./cmd/generate-color-inventory --check
cd scripts && go run ./cmd/verify-colors
```

差分がある場合は非0 exitで失敗する。**コミット前は必ずこれをpassさせる**。

`verify-colors` は use-site pair ごとの truecolor AA 判定、cterm fallback と環境依存pairの report-only 表示、waiver理由、未知HEX、ANSI CVDレポートを統合する。exit 0はenforced subsetの合格だけを意味し、実画面の全箇所がAAであることは保証しない。

### 4. zshへの反映（別経路）

`zsh/.zshrc`はNix store symlinkのため、上記generateの対象ではなく`darwin-rebuild switch --flake ~/.config`で反映する。**switch前はverify-colorsのUNKNOWN HEX LITERALSがzsh/.zshrcに対してexpected failする場合がある**（symlink先が旧内容のため）。switch未実施時のfailを不整合と誤認しない。

## 未知HEX検査

`verify-colors`のUNKNOWN HEX LITERALSセクションは、生成物・設定ファイル中のHEXリテラルがパレットに存在するか検査する。生成対象ファイルにTOML由来でない未知のHEXリテラルが紛れている場合はpolicy NGとしてexit 1になる。

```bash
cd scripts && go run ./cmd/verify-colors
```

## CVD・コントラストの数値検証

`verify-colors`はuse-site pair単位のWCAGコントラストと、ANSI 16色全ペアのCVDシミュレーションを出力する。

```bash
cd scripts && go run ./cmd/verify-colors
```

- 不透明な実使用背景を持つenforced pairは4.5:1をexit-code gateにする
- 環境依存pairとcterm fallbackはreport-onlyとして数値を表示する
- ANSI 16色全ペアのCVD ΔEは独自スクリーニング基準としてreport-onlyで表示する

## 不整合の修正手順

### 1. エラーメッセージの確認

`generate-colors --check`または`verify-colors`のエラーが表示された場合、以下を確認：

- どのファイルで差分/エラーが発生したか
- `colors/ghost-visor.toml`の現在値は何か

### 2. 単一ソースの確認

```bash
grep "background" colors/ghost-visor.toml
```

### 3. 再生成・再検証

```bash
cd scripts && go run ./cmd/generate-colors
cd scripts && go run ./cmd/generate-colors --check
cd scripts && go run ./cmd/generate-color-inventory
cd scripts && go run ./cmd/generate-color-inventory --check
cd scripts && go run ./cmd/verify-colors
```

生成対象ファイルを手で直接編集しない。差分は必ず`colors/ghost-visor.toml`側を直して再生成で解消する。

## カラー定義の構造

### colors/ghost-visor.toml の構造

```toml
[metadata]
name = "Ghost Visor"
version = "1.1.0"

[core]
background = "#202A42"        # メイン背景
darkest_bg = "#141B2D"        # 最暗背景
panel_bg = "#324664"          # パネル背景
selection_bg = "#5199C2"      # 選択強調

[foregrounds]
main = "#CDE9F5"              # メインテキスト
bright = "#9FDBF7"            # ブライトテキスト

[teals]
bright = "#58CAF8"            # Visor Glow Cyan
mid_bright = "#92BFD9"        # Selection Blue
standard = "#9ABED3"          # String Blue
```

## 対象設定ファイル

生成・検証が扱う設定ファイル：

1. **WezTerm** (`wezterm/colors.lua`) — ANSI 16色・背景色・前景色・UI要素（generate対象）
2. **Neovim** (`nvim/lua/config/palette.lua`) — 構文ハイライト・UI要素・診断表示（generate対象）
3. **home-manager** (`home-manager/colors.nix`) — Nix側の色定義（generate対象）
4. **LazyGit** (`lazygit/config.yml`) — テーマカラー・ボーダー・選択色（マーカーブロック）
5. **vim** (`vim/vimrc`) — cterm256近似（マーカーブロック）
6. **Claude Code** (`claude/statusline.sh`) — 10進RGB（マーカーブロック）
7. **Zsh** (`zsh/.zshrc`) — Nix store symlink経由、darwin-rebuild switchで反映

### statusline固有の制約

- `claude/statusline.sh`の`# BEGIN/END GENERATED COLORS: ANSI`と`: SEGMENTS`の2ブロックは**直接編集しない**。生成元は`scripts/cmd/generate-colors/main.go`の`statuslineANSITemplate`/`statuslineSegmentsTemplate`
- `verify-colors`のinventory抽出（`scripts/internal/verifycolors/sourceinventory/statusline.go`）は色変数名を正規表現で拾うため、**変数名は`C_[A-Z]+`（アンダースコア不可）**。バーの未使用部分などの意図的低コントラスト色は**`C_[A-Z]+TRACK`**の命名に従う必要がある。TRACK色は4.5:1を強制しないreport-onlyのsurface pairとして登録される
- バー描画のような**汎用ヘルパー関数は生成ブロックの外**に定義する（生成ブロック内にはトークン参照を含む色の判定だけを置く）。既存例は`build_meter`/`format_remaining`

## トラブルシューティング

### バリデーション/生成スクリプトが見つからない

```bash
ls -la scripts/cmd/generate-colors/main.go scripts/cmd/generate-color-inventory/main.go scripts/cmd/verify-colors/main.go
```

### Goがインストールされていない

```bash
go version

# インストールされていない場合はNix (home-manager) で追加する
# home-manager/dev.nix 等の home.packages に go を追加して
# darwin-rebuild switch --flake ~/.config で適用（brew installは使わない）
```

### zshだけ反映されない

`darwin-rebuild switch --flake ~/.config`未実施が原因である可能性が高い。生成物側の問題ではないため、他ファイルのgenerate/checkを再実行して切り分けない。

## ベストプラクティス

### カラー変更のワークフロー

```bash
# 1. 単一ソースを編集
nvim colors/ghost-visor.toml

# 2. 生成
cd scripts && go run ./cmd/generate-colors
cd scripts && go run ./cmd/generate-color-inventory

# 3. 生成漏れ確認
cd scripts && go run ./cmd/generate-colors --check
cd scripts && go run ./cmd/generate-color-inventory --check
cd scripts && go run ./cmd/verify-colors

# 4. 問題なければコミット
git add colors/ wezterm/ nvim/ home-manager/ lazygit/ vim/ claude/statusline.sh COLOR-SYSTEM.md
git commit -m "feat(theme): カラー定義を更新"
```

### 新しいカラーの追加

1. `colors/ghost-visor.toml`に定義を追加
2. `cd scripts && go run ./cmd/generate-colors`で派生ファイルへ反映
3. `go run ./cmd/generate-color-inventory`で検査inventoryへ反映
4. 両generatorの`--check`で生成漏れがないことを確認
5. `go run ./cmd/verify-colors`でuse-site pair、未知HEX、CVDを統合検証

## 関連ドキュメント

- `.claude/rules/color-system.md` - カラーシステム規約
- `COLOR-SYSTEM.md` - カラーシステムガイドライン
- `colors/ghost-visor.toml` - カラー定義の単一ソース
- `scripts/cmd/generate-colors/main.go` - 生成スクリプト（本体）
- `scripts/cmd/generate-color-inventory/main.go` - use-site pair inventory生成スクリプト
- `scripts/cmd/verify-colors/main.go` - use-site pair・描画profile・未知HEX・CVDの統合検証
- `.claude/skills/config-change/SKILL.md` - 設定変更ワークフロー
