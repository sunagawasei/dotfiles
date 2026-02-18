# カラーシステムドキュメント

このドキュメントは、dotfilesで実際に使用している **Abyssal Teal (深淵の青緑) Expanded** テーマのカラーパレット、色選択ルール、および各設定ファイルの色設定一覧を定義します。

**カラー定義の単一ソース**: [`colors/abyssal-teal.toml`](./colors/abyssal-teal.toml)

---

## 1. カラーパレット定義（拡張版）

元画像を32色に詳細分解し、UIの各階層により細かいニュアンスを与えました。

### コア・背景系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メイン背景 | `#0B0C0C` | 基本背景（純黒に近い） |
| 最暗背景 | `#111E16` | タブバー、ガター |
| パネル背景 | `#152A2B` | サイドバー、フロート |
| UIシャドウ | `#31364C` | 非アクティブ領域、ポップアップ |
| 選択強調 | `#64BBBE` | 選択範囲、アクティブ状態 |

### テキスト・強調系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メインテキスト | `#CEF5F2` | 通常文字 |
| 最高強調 | `#F2FFFF` | カーソルテキスト、重要 |
| ブライトテキスト | `#B1F4ED` | 数値、定数、アクティブタブ |
| ヘディング/パス | `#9DDCD9` | ディレクトリ名、見出し |
| ディムテキスト | `#92A2AB` | 補助テキスト |

### スペクトラム・アクセント

| グループ | 色 | HEX | 用途 |
|---------|-----|-----|-----|
| **Teal** | Success Teal | `#6AB9A8` | 成功インジケーター（AAA level） |
| | Vibrant Teal | `#6CD8D3` | 関数、コマンド |
| | Clear Teal | `#64BBBE` | 演算子、ステータス |
| | Base Teal | `#659D9E` | 文字列、情報 |
| | UI Border | `#4D8F9E` | 境界線、分割線（視認性改善版） |
| **Slate** | Git Blame Gray | `#8A97AD` | Git blame、補助情報 |
| | Comment Gray | `#7A869A` | コメント（視認性改善版） |
| | Punctuation Gray | `#7A8599` | 句読点（視認性改善版） |
| | Slate Mid | `#525B65` | 非アクティブ（特定用途のみ） |
| | Sky Slate | `#A4ABCB` | 型、オプション |
| **Purple** | Glitch Purple | `#936997` | エラー、重要警告 |
| | Lavender | `#CED5E9` | 警告、定数、ラベル |
| | Muted Purple | `#8A99BD` | キーワード、予約語（視認性改善版） |

### ANSI 16色（拡張マッピング）

```
 0: #111E16 (Black)           8: #525B65 (Bright Black)
 1: #936997 (Red)             9: #865F7B (Bright Red)
 2: #349594 (Green)          10: #64BBBE (Bright Green)
 3: #CED5E9 (Yellow)         11: #B1F4ED (Bright Yellow)
 4: #326787 (Blue)           12: #A4ABCB (Bright Blue)
 5: #584EA2 (Magenta)        13: #B4B7CD (Bright Magenta)
 6: #6CD8D3 (Cyan)           14: #9DDCD9 (Bright Cyan)
 7: #CEF5F2 (White)          15: #F2FFFF (Bright White)
```

---

## 2. 色選択ルール

### 背景の階層構造

1. **`#111E16` (Deepest)**: 外枠、タブバー背景
2. **`#0B0C0C` (Main)**: エディタ、ターミナル作業領域（純黒に近い）
3. **`#152A2B` (Panel)**: サイドバー、フロート
4. **`#31364C` (UI Layer)**: ポップアップ、非アクティブ領域
5. **`#64BBBE` (Selection)**: アクティブ選択、ハイライト

### テキストの優先順位

1. **`#F2FFFF` (Critical)**: アクティブな強調
2. **`#B1F4ED` (High)**: 定数、アクティブ要素
3. **`#CEF5F2` (Standard)**: メインテキスト
4. **`#92A2AB` (Low)**: 補助情報
5. **`#8A97AD` (Auxiliary)**: Git blame、補助的な情報
6. **`#7A869A` (Comment)**: コメント（視認性改善版）
7. **`#7A8599` (Punctuation)**: 句読点（視認性改善版）
8. **`#525B65` (Inactive)**: 非アクティブ要素（意図的に暗い）

---

## 3. 設定ファイル一覧

各設定ファイルは `colors/abyssal-teal.toml` の定義に基づいています。

### 統合状況

*   **WezTerm**: 32色分解に基づいたフルパレット適用
*   **Neovim**: 構文ハイライトに Muted Purple (Keyword), Clear Teal (Operator) 等を追加
*   **Zsh**: コマンド、パス、オプションに個別アクセントを割り当て
*   **Starship**: ヘディング・アクセントを最適化
*   **LazyGit**: ボーダーと選択色のコントラストを調整

### バリデーション

変更後は必ず以下のスクリプトで整合性を確認してください：

```bash
cd scripts && go run validate-colors.go
```

---

## 参考リンク

- [colors/abyssal-teal.toml](./colors/abyssal-teal.toml) - 拡張定義ソース
- [CLAUDE.md](./CLAUDE.md) - 設計方針
