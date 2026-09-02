---
paths:
  - "**/colors/**"
  - "**/wezterm/wezterm.lua"
  - "**/nvim/lua/plugins/colorscheme.lua"
  - "**/lazygit/config.yml"
  - "**/*theme*"
  - "**/*color*"
---

# カラーシステム規約

## 単一ソースの原則

**`colors/ghost-visor.toml`を真実の単一ソースとする**

派生設定ファイルを直接編集せず、このファイルを変更して generator を実行してください。

## カラーテーマ

**Ghost Visor** - 紫紺の背景に地平の紫とマゼンタを重ね、シアンを補助アクセントに置くパレット

## 適用範囲

以下の設定ファイルで一貫性のあるカラーを使用：

- WezTerm (`wezterm/wezterm.lua`)
- Neovim (`nvim/lua/plugins/colorscheme.lua`)
- LazyGit (`lazygit/config.yml`)
- Zsh (プロンプトとシンタックスハイライト)

## カラー選択ルール

詳細なガイドラインは`COLOR-SYSTEM.md`を参照してください。

### 背景の階層構造

1. `#0C1226` (Deepest): 外枠、タブバー背景
2. `#1A2340` (Main): エディタ、ターミナル作業領域
3. `#332E56` (Panel): サイドバー、フロート
4. `#1F2745` (UI Layer): ポップアップ、非アクティブ領域
5. `#9385C8` (Selection): アクティブ選択、ハイライト

### テキストの優先順位

1. `#F8FCFD` (Critical): アクティブな強調
2. `#96D7F5` (High): 数値、アクティブ要素
3. `#CDE9F5` (Standard): メインテキスト
4. `#ABA4C4` (Low): 補助情報

## バリデーション必須

inventoryは各pairの出所を`Source: <file>:<line>`で持つため、**色を1つも変えなくても、色を含むファイル
（`home-manager/hunk.nix`・`wezterm/keybinds.lua`等）に行が増減しただけでdriftする**。
色以外の理由でそれらを編集したときも下記を流すこと（怠ると`--check`が後で落ちる）。

カラー変更後は必ず以下を実行して整合性を確認：

```bash
cd scripts
go run ./cmd/generate-colors
go run ./cmd/generate-colors --check
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-cvd-pairs
go run ./cmd/verify-colors
```

### 段8: report-only集合の差分確認

report-onlyは非強制集合であり、below-AAの増減を記録する観測値です。
report-only集合と矛盾するfixture凍結を行わないでください。
beforeは、そのtaskで編集を始める前の作業ツリーから採ります。
HEADで代用すると、未commitの先行task成果と当該taskの差分が混ざるためです。

実行可能な順序は次のとおりです。

1. 編集開始前に作業ツリーを基準ツリーへコピーし、基準ツリーで`verify-colors`を実行します。
   `git show HEAD`から基準を作らないでください。
   例：`before_root=$(mktemp -d); before_report="$before_root/verify.txt"; rsync -a --exclude .git ./ "$before_root/"; (cd "$before_root/scripts" && go run ./cmd/verify-colors > "$before_report")`
2. 編集後の作業ツリーでも`after_report=$(mktemp); (cd scripts && go run ./cmd/verify-colors > "$after_report")`を実行します。
3. 出力の`[REPORT-ONLY BELOW-AA]`行を`ConsumerID profile role`へ正規化し、`sort -u`します。
   例：`sed -n 's/^\[REPORT-ONLY BELOW-AA\] \([^ ]*\) profile=\([^ ]*\) role=\([^ ]*\).*$/\1 \2 \3/p' "$before_report" | sort -u > before.sorted; sed -n 's/^\[REPORT-ONLY BELOW-AA\] \([^ ]*\) profile=\([^ ]*\) role=\([^ ]*\).*$/\1 \2 \3/p' "$after_report" | sort -u > after.sorted`
4. `comm -13 before.sorted after.sorted`で新規、`comm -23 before.sorted after.sorted`で解消を出します。

### 手書きoverrideの出所と再検証

inventoryのoverrideを追加するときは、Source欄へ消費側の`file:line`を必ず書きます。
消費側toolの版とrev（例：`herdr v0.8.0 rev 346411fa21afd297f5ed3b3fa56f9e3fbf7654b7`）も必ず書きます。
pairが静的に決まる前提条件（例：`panel_bg`が`reset`、`surface_dim`が特定tokenへ配線されること）をSource欄へ記録します。
`flake.lock`で消費側toolを`bump`したら、該当overrideを再検証してください。

## 一貫性要件

- 全アプリケーション間でカラーの一貫性を保つ
- 新しいカラー値を追加する場合は `colors/ghost-visor.toml` を更新
- `scripts/cmd/generate-colors/main.go` で派生設定を再生成
- `go run ./cmd/generate-colors --check` で生成漏れがないことを確認
- `scripts/cmd/generate-color-inventory/main.go` で検査inventoryを再生成
- `go run ./cmd/generate-color-inventory --check` でinventoryの生成漏れがないことを確認
- `go run ./cmd/verify-colors` でuse-site pair、描画profile、未知HEX、CVDレポートを統合検証
