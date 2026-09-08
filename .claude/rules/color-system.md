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

### report-only集合の差分確認・手書きoverrideの出所

手順の正本は `.claude/skills/color-validation/SKILL.md`（「段8: report-only集合の差分確認」「手書きoverrideの出所と再検証」）。
report-onlyは非強制集合なので、fixture凍結で強制しない。

## 一貫性要件

- 全アプリケーション間でカラーの一貫性を保つ
- 新しいカラー値を追加する場合は `colors/ghost-visor.toml` を更新
- `scripts/cmd/generate-colors/main.go` で派生設定を再生成
- `go run ./cmd/generate-colors --check` で生成漏れがないことを確認
- `scripts/cmd/generate-color-inventory/main.go` で検査inventoryを再生成
- `go run ./cmd/generate-color-inventory --check` でinventoryの生成漏れがないことを確認
- `go run ./cmd/verify-colors` でuse-site pair、描画profile、未知HEX、CVDレポートを統合検証
