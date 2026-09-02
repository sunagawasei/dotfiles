# カラーシステムドキュメント

このドキュメントは、dotfilesで使用する **Ghost Visor** テーマのカラーパレット、色選択ルール、および生成対象を定義します。

**カラー定義の単一ソース**: [`colors/ghost-visor.toml`](./colors/ghost-visor.toml)

---

## 1. カラーパレット定義

壁紙（Ghost in the Shell、夜景スカイライン）の地平線帯の実測色を基に、紫紺の背景に地平の紫とマゼンタを重ねています。シアンは関数・コマンドの識別のため補助のアクセントに置いています。

### コア・背景系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メイン背景 | `#1A2340` | エディタ、ターミナル作業領域 |
| 最暗背景 | `#0C1226` | タブバー、外枠 |
| パネル背景 | `#332E56` | サイドバー、フロート |
| UIシャドウ | `#1F2745` | 非アクティブ領域、ポップアップ |
| 選択強調 | `#9385C8` | 選択範囲、アクティブ状態 |

### テキスト・強調系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メインテキスト | `#CDE9F5` | 通常文字 |
| 最高強調 | `#F8FCFD` | カーソルテキスト、重要 |
| ブライトテキスト | `#96D7F5` | 数値、アクティブタブ |
| ヘディング/パス | `#B79AD0` | ディレクトリ名、見出し |
| ディムテキスト | `#ABA4C4` | 補助テキスト |
| Subdued | `#9E97B8` | 非アクティブ前景、低強調テキスト |

### スペクトラム・アクセント

| グループ | 色 | HEX | 用途 |
|---------|-----|-----|-----|
| **Cyan** | Success | `#76D6C4` | 成功インジケーター |
| | Visor Glow | `#58CAF8` | 関数、コマンド |
| | Operator Blue | `#92BFD9` | 演算子、ステータス |
| | String Blue | `#9ABED3` | 文字列、情報 |
| | UI Border | `#4C4C74` | 境界線、分割線 |
| **Slate** | Git Blame Gray | `#A8A5C2` | Git blame、補助情報 |
| | Comment Gray | `#A5A2BC` | コメント |
| | Punctuation Gray | `#AEACC4` | 句読点 |
| | Slate Mid | `#453C5F` | 意図的減光専用（Conceal、Flash backdrop、打ち消し線、背景塗り） |
| | Sky Slate | `#AFB8D9` | 型、オプション |
| **Purple** | Error Rose | `#E2839A` | エラー、重要警告 |
| | Lavender | `#D0D4F0` | 警告、定数、ラベル |
| | Muted Purple | `#B79AD0` | キーワード、予約語 |

### ANSI 16色（拡張マッピング）

<!-- BEGIN GENERATED COLORS -->
```
 0: #0C1226 (Black)           8: #565273 (Bright Black)
 1: #C97F9E (Red)             9: #E2839A (Bright Red)
 2: #2FA5A0 (Green)          10: #76D6C4 (Bright Green)
 3: #CAB083 (Yellow)         11: #DFBE90 (Bright Yellow)
 4: #7089D6 (Blue)           12: #AFB8D9 (Bright Blue)
 5: #C86FBE (Magenta)        13: #E6BCF4 (Bright Magenta)
 6: #58CAF8 (Cyan)           14: #96D7F5 (Bright Cyan)
 7: #CDE9F5 (White)          15: #F8FCFD (Bright White)
```
<!-- END GENERATED COLORS -->

---

## 2. 色選択ルール

### 背景の階層構造

1. **`#0C1226` (Deepest)**: 外枠、タブバー背景
2. **`#1A2340` (Main)**: エディタ、ターミナル作業領域
3. **`#332E56` (Panel)**: サイドバー、フロート
4. **`#1F2745` (UI Layer)**: ポップアップ、非アクティブ領域
5. **`#9385C8` (Selection)**: アクティブ選択、ハイライト

### テキストの優先順位

1. **`#F8FCFD` (Critical)**: アクティブな強調
2. **`#96D7F5` (High)**: 数値、アクティブ要素
3. **`#CDE9F5` (Standard)**: メインテキスト
4. **`#ABA4C4` (Low)**: 補助情報
5. **`#A8A5C2` (Auxiliary)**: Git blame、補助的な情報
6. **`#A5A2BC` (Comment)**: コメント
7. **`#AEACC4` (Punctuation)**: 句読点
8. **`#9E97B8` (Subdued)**: 非アクティブ前景、低強調テキスト
9. **`#453C5F` (Intentional Dim)**: Conceal、Flash backdrop、打ち消し線、背景塗り

---

## 3. 設定ファイル一覧

各設定ファイルは `colors/ghost-visor.toml` の定義に基づいています。

### 統合状況

*   **WezTerm**: ANSI 16色とタブ・選択色を生成モジュールから参照
*   **Neovim**: 生成パレットを構文ハイライトとUIに適用
*   **Zsh**: コマンド、パス、オプションに個別アクセントを割り当て
*   **Pure**: path・git・prompt各要素に Ghost Visor を適用
*   **LazyGit**: ボーダーと選択色のコントラストを調整

### バリデーション

色を変更する場合は `colors/ghost-visor.toml` を編集し、派生設定を生成して整合性を確認してください：

```bash
cd scripts
go run ./cmd/generate-colors
go run ./cmd/generate-colors --check
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-colors
```

---

## 参考リンク

- [colors/ghost-visor.toml](./colors/ghost-visor.toml) - 拡張定義ソース
- [CLAUDE.md](./CLAUDE.md) - 設計方針
