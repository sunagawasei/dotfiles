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
cd scripts && go run ./cmd/verify-cvd-pairs
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

## 段8: report-only集合の差分確認

report-onlyは非強制集合であり、below-AAの増減を記録するための観測値です。
fixtureを凍結してreport-onlyの集合を強制してはいけません。
比較のbeforeは、そのtaskで編集を始める前の作業ツリーから採ります。
HEADで代用すると、未commitの先行task成果と当該taskの差分が混ざるためです。

実行順序は次のとおりです。

1. 編集開始前に作業ツリー全体を基準ツリーへコピーし、基準ツリーで`verify-colors`を実行する。
   `git show HEAD`から基準を作ってはならない。
   例：`before_root=$(mktemp -d); before_report="$before_root/verify.txt"; rsync -a --exclude .git ./ "$before_root/"; (cd "$before_root/scripts" && go run ./cmd/verify-colors > "$before_report")`
2. 編集後の作業ツリーでも`after_report=$(mktemp); (cd scripts && go run ./cmd/verify-colors > "$after_report")`を実行する。
3. 各出力の`[REPORT-ONLY BELOW-AA]`行を、`ConsumerID profile role`の3列へ正規化し、`sort -u`する。
   例：`sed -n 's/^\[REPORT-ONLY BELOW-AA\] \([^ ]*\) profile=\([^ ]*\) role=\([^ ]*\).*$/\1 \2 \3/p' "$before_report" | sort -u > before.sorted; sed -n 's/^\[REPORT-ONLY BELOW-AA\] \([^ ]*\) profile=\([^ ]*\) role=\([^ ]*\).*$/\1 \2 \3/p' "$after_report" | sort -u > after.sorted`
4. `comm -13 before.sorted after.sorted`で新規、`comm -23 before.sorted after.sorted`で解消を出す。
   beforeとafterの両方を同じ`ConsumerID profile role`形式にしてから比較する。

## 手書きoverrideの出所と再検証

inventoryのoverrideを追加するときは、Source欄へ消費側の`file:line`を必ず書きます。
同じSource欄へ消費側toolの版とrev（例：`herdr v0.8.0 rev 346411fa21afd297f5ed3b3fa56f9e3fbf7654b7`）を必ず記録します。
pairが静的に決まる前提条件（例：`panel_bg`が`reset`、`surface_dim`が特定tokenへ配線されること）もSource欄へ書きます。
`flake.lock`で消費側toolを`bump`したら、該当overrideのSource、実装、pairを再検証してください。

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
version = "2.0.0"

[core]
background = "#1A2340"        # メイン背景
darkest_bg = "#0C1226"        # 最暗背景
panel_bg = "#332E56"          # パネル背景
selection_bg = "#9385C8"      # 選択強調

[foregrounds]
main = "#CDE9F5"              # メインテキスト
bright = "#96D7F5"            # ブライトテキスト

[teals]
bright = "#58CAF8"            # Visor Glow Cyan
mid_bright = "#92BFD9"        # Operator Blue
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
