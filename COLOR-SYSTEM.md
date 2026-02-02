# カラーシステムドキュメント

このドキュメントは、dotfilesで実際に使用しているカラーパレット、色選択ルール、および各設定ファイルの色設定一覧を定義します。

---

## 1. カラーパレット定義

### コアカラー（モノクロ階調）

| 名称 | HEX | RGB | 用途 |
|-----|-----|-----|-----|
| 背景 | `#1A201E` | 26,32,30 | メイン背景 |
| 濃い影 | `#2A2F2E` | 42,47,46 | 選択行、ホバー |
| 分割線 | `#3A3F3E` | 58,63,62 | ボーダー |
| 中間グレー | `#7E8A89` | 126,138,137 | コメント、補助 |
| 明るいグレー | `#AAB6B5` | 170,182,181 | 削除行 |
| ライトグレー | `#CDD8D7` | 205,216,215 | 演算子 |
| 前景 | `#D7E2E1` | 215,226,225 | メインテキスト |
| ニアホワイト | `#E6F1F0` | 230,241,240 | 強調 |
| ハイライト白 | `#F2FFFF` | 242,255,255 | 最強調 |

### アクセントカラー

| 名称 | HEX | RGB | 用途 |
|-----|-----|-----|-----|
| Cyan | `#5AAFAD` | 90,175,173 | キーワード、成功、追加 |
| Bright Cyan | `#96CBD1` | 150,203,209 | 関数、型 |
| Magenta | `#8C83A3` | 140,131,163 | 文字列、変更 |
| Bright Magenta | `#B3A9D1` | 179,169,209 | 強調紫 |

### ANSI Red/Green/Yellow/Blue（グレースケール化）

| 色 | HEX | RGB | 用途 |
|-----|-----|-----|-----|
| ANSI Red → Gray | `#B9C6C5` | 185,198,197 | ANSI 1 |
| ANSI Green → Gray | `#9AA6A5` | 154,166,165 | ANSI 2 |
| ANSI Yellow → Gray | `#C7D2D1` | 199,210,209 | ANSI 3 |
| ANSI Blue → Gray | `#7E8A89` | 126,138,137 | ANSI 4 |

### Bright色（グレースケール化）

| 色 | HEX | RGB | 用途 |
|-----|-----|-----|-----|
| Bright Red → Near White | `#E6F1F0` | 230,241,240 | ANSI 9 |
| Bright Green → Light Gray | `#CDD8D7` | 205,216,215 | ANSI 10 |
| Bright Yellow → White | `#F2FFFF` | 242,255,255 | ANSI 11 |
| Bright Blue → Light Gray | `#AAB6B5` | 170,182,181 | ANSI 12 |

### ANSI 16色（WezTerm/Neovim共通）

モノクロ基調＋アクセントの統一ANSI 16色マッピング:

```
 0: #1A201E (Black)           8: #7E8A89 (Bright Black - 視認性向上)
 1: #B9C6C5 (Red→Gray)        9: #E6F1F0 (Bright Red→Near White)
 2: #9AA6A5 (Green→Gray)     10: #CDD8D7 (Bright Green→Light Gray)
 3: #C7D2D1 (Yellow→Gray)    11: #F2FFFF (Bright Yellow→White)
 4: #7E8A89 (Blue→Gray)      12: #AAB6B5 (Bright Blue→Light Gray)
 5: #8C83A3 (Magenta)        13: #B3A9D1 (Bright Magenta)
 6: #5AAFAD (Cyan)           14: #96CBD1 (Bright Cyan)
 7: #D7E2E1 (White)          15: #F2FFFF (Bright White)
```

---

## 2. 色選択ルール

### 背景系の階調

```
#1A201E → #2A2F2E → #3A3F3E
(最暗)    (中間)    (境界線)
```

- **`#1A201E`**: メインの背景色
- **`#2A2F2E`**: 選択行、ホバー状態
- **`#3A3F3E`**: ボーダー、分割線

### テキスト系の階調

```
#7E8A89 → #D7E2E1 → #F2FFFF
(補助)    (通常)    (強調)
```

- **`#7E8A89`**: コメント、非アクティブな要素
- **`#D7E2E1`**: 通常のテキスト
- **`#F2FFFF`**: 重要な要素の強調

### 意味的色分け

| 意味 | 色 | カラーコード | 使用例 |
|------|-----|-------------|--------|
| 追加/成功 | Cyan | `#5AAFAD` | Git追加行、成功メッセージ |
| 変更/警告 | Magenta | `#8C83A3` | Git変更行、文字列リテラル |
| 削除 | 明るいグレー | `#AAB6B5` | Git削除行 |
| エラー | Magenta | `#8C83A3` | エラーメッセージ、診断 |
| 警告 | Bright Magenta | `#B3A9D1` | 警告メッセージ |
| 情報 | Bright Cyan | `#96CBD1` | 情報メッセージ |
| ヒント | Cyan | `#5AAFAD` | ヒントメッセージ |

### エディタモード色（グレースケール化）

- **Normal**: Gray (`#7E8A89`)
- **Insert**: Gray (`#9AA6A5`)
- **Visual**: Gray (`#B9C6C5`)
- **Replace**: Gray (`#AAB6B5`)
- **Command**: Gray (`#C7D2D1`)
- **Terminal**: Cyan (`#5AAFAD`) - アクセント維持

---

## 3. 設定ファイル一覧

### ターミナル

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `wezterm/wezterm.lua` | `config.colors` | ANSI 16色、カーソル、選択範囲 |
| `wezterm/wezterm.lua` | タブバー色 | タブの背景、前景、新規タブボタン |

**主な設定内容**:
- ANSI Black: `#1A201E` (背景)
- ANSI 1-4: グレースケール (`#B9C6C5`, `#9AA6A5`, `#C7D2D1`, `#7E8A89`)
- ANSI Cyan: `#5AAFAD` (アクセント維持)
- ANSI Magenta: `#8C83A3` (アクセント維持)
- ANSI White: `#D7E2E1` (前景)
- Bright 9-12: グレースケール (`#E6F1F0`, `#CDD8D7`, `#F2FFFF`, `#AAB6B5`)
- カーソル: `#F2FFFF`
- 選択範囲前景: `#F2FFFF` (Bright White)
- 選択範囲背景: `#5AAFAD` (Cyan)
- モード色: 全てグレースケール

### Cursor CLI

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `cursor/cli-config.json` | エディタ設定 | Vimモード、行番号表示 |
| WezTerm ANSI色に準拠 | カラーテーマ | ターミナルエミュレーターの色設定を使用 |

**主な設定内容**:
- カラーテーマ: WezTermのANSI 16色設定に準拠
- 背景色: `#1A201E` (WezTermから継承)
- 前景色: `#D7E2E1` (WezTermから継承)
- Vimモード: 有効
- 行番号表示: 有効

**注意**: Cursor CLIは独自のカラーテーマ設定をサポートしていません。
すべての色設定はターミナルエミュレーター（WezTerm）のANSI色設定から継承されます。

### Neovim

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `nvim/lua/plugins/colorscheme.lua` | ハイライトグループ | 構文、UI、診断、Git差分 |
| `nvim/lua/plugins/lualine.lua` | `custom_theme` | モード別ステータスライン色 |
| `nvim/lua/plugins/scrollbar.lua` | ハンドル、マーカー | スクロールバー、検索/診断マーカー |
| `nvim/lua/plugins/render-markdown.lua` | ヘッダー、要素 | Markdownレンダリング色 |

**主な設定内容**:
- Normal背景: `#1A201E`
- 選択行: `#2A2F2E`
- コメント: `#7E8A89`
- キーワード: `#5AAFAD` (Cyan)
- 関数: `#96CBD1` (Bright Cyan)
- 文字列: `#8C83A3` (Magenta)
- モード色: 全てグレースケール (Normal: `#7E8A89`, Insert: `#9AA6A5`, etc.)
- Diff追加: `#5AAFAD` (Cyan)
- Diff変更: `#8C83A3` (Magenta)
- Diff削除: `#AAB6B5` (グレー)
- 診断エラー: `#8C83A3` (Magenta)
- 診断警告: `#B3A9D1` (Bright Magenta)
- 診断情報: `#96CBD1` (Bright Cyan)

### プロンプト

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `starship.toml` | `format` 各要素 | ディレクトリ、Git、言語インジケータ |

**主な設定内容**:
- ディレクトリ: Cyan (`#5AAFAD`)
- Gitブランチ: Bright Magenta (`#B3A9D1`)
- 言語バージョン: 前景 (`#D7E2E1`)

### Git TUI

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `lazygit/config.yml` | `gui.theme` | LazyGit全体のテーマ設定 |

**主な設定内容**:
- 選択行背景: `#2A2F2E`
- 選択行テキスト: Cyan (`#5AAFAD`)
- 未ステージ: Bright Magenta (`#B3A9D1`)
- デフォルトテキスト: `#D7E2E1`

### シェル

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `zsh/.zshrc` | `LS_COLORS` | ディレクトリ一覧の色 |
| `zsh/.zshrc` | `FZF_DEFAULT_OPTS` | ファジーファインダーの色 |

**主な設定内容**:
- FZFプロンプト: Cyan (`#5AAFAD`)
- FZF選択: `#2A2F2E`
- FZF境界線: `#3A3F3E`

### Vim

| ファイル | 設定項目 | 説明 |
|---------|---------|------|
| `vim/vimrc` | `highlight` 定義 | Vim構文ハイライト |

---

## 4. 色の一貫性を保つためのガイドライン

### 新しい設定を追加する場合

1. **背景色が必要な場合**: コアカラーの階調から選択
   - 最暗背景: `#1A201E`
   - インタラクティブ要素: `#2A2F2E`
   - 境界線: `#3A3F3E`

2. **テキスト色が必要な場合**: テキスト系階調から選択
   - 補助情報: `#7E8A89`
   - 通常テキスト: `#D7E2E1`
   - 強調: `#F2FFFF`

3. **意味のある色分けが必要な場合**: アクセントカラーのみ使用
   - 成功/追加: Cyan (`#5AAFAD`)
   - 変更/文字列: Magenta (`#8C83A3`)
   - エラー: Magenta (`#8C83A3`)
   - 警告: Bright Magenta (`#B3A9D1`)
   - 情報: Bright Cyan (`#96CBD1`)
   - ヒント: Cyan (`#5AAFAD`)

4. **モード色が必要な場合**: グレースケールを使用
   - 各モード（Normal, Insert, Visual等）は異なるグレー階調を使用
   - TerminalモードのみCyan (`#5AAFAD`) を使用

### カラーコードの記述形式

- **HEX形式を優先**: `#0E1210`
- **RGB形式（必要な場合）**: `14,18,16`
- **大文字を使用**: `#0E1210` (小文字 `#0e1210` ではなく)

### 設定ファイルを変更した場合

このドキュメントも同時に更新して、色設定の一貫性を維持してください。

---

## 参考リンク

- [CLAUDE.md](./CLAUDE.md) - リポジトリ全体の設計方針
