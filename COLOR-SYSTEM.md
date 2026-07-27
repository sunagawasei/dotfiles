# カラーシステムドキュメント

このドキュメントは、dotfilesで使用する **Ghost Visor** テーマのカラーパレット、色選択ルール、および生成対象を定義します。

**カラー定義の単一ソース**: [`colors/ghost-visor.toml`](./colors/ghost-visor.toml)

---

## 1. カラーパレット定義

壁紙（Ghost in the Shell ep.01 frame）の実測色を基に、暗いインディゴ背景と visor-glow のシアン／マゼンタを組み合わせています。

### コア・背景系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メイン背景 | `#202A42` | エディタ、ターミナル作業領域 |
| 最暗背景 | `#141B2D` | タブバー、外枠 |
| パネル背景 | `#324664` | サイドバー、フロート |
| UIシャドウ | `#242F48` | 非アクティブ領域、ポップアップ |
| 選択強調 | `#5199C2` | 選択範囲、アクティブ状態 |

### テキスト・強調系

| 名称 | HEX | 用途 |
|-----|-----|-----|
| メインテキスト | `#CDE9F5` | 通常文字 |
| 最高強調 | `#F8FCFD` | カーソルテキスト、重要 |
| ブライトテキスト | `#9FDBF7` | 数値、定数、アクティブタブ |
| ヘディング/パス | `#88CBEA` | ディレクトリ名、見出し |
| ディムテキスト | `#ADBAC9` | 補助テキスト |
| Subdued | `#949DB0` | 非アクティブ前景、低強調テキスト |

### スペクトラム・アクセント

| グループ | 色 | HEX | 用途 |
|---------|-----|-----|-----|
| **Cyan** | Success | `#62C9C2` | 成功インジケーター |
| | Visor Glow | `#58CAF8` | 関数、コマンド |
| | Selection Blue | `#92BFD9` | 演算子、ステータス |
| | String Blue | `#9ABED3` | 文字列、情報 |
| | UI Border | `#45799D` | 境界線、分割線 |
| **Slate** | Git Blame Gray | `#B1B9CD` | Git blame、補助情報 |
| | Comment Gray | `#B1B9CA` | コメント |
| | Punctuation Gray | `#B3B9C7` | 句読点 |
| | Slate Mid | `#4A4953` | 意図的減光専用（Conceal、Flash backdrop、打ち消し線、背景塗り） |
| | Sky Slate | `#AFB8D9` | 型、オプション |
| **Purple** | Error Purple | `#CEADE1` | エラー、重要警告 |
| | Lavender | `#D0D4F0` | 警告、定数、ラベル |
| | Muted Purple | `#B4B6DA` | キーワード、予約語 |

### ANSI 16色（拡張マッピング）

<!-- BEGIN GENERATED COLORS -->
```
 0: #141B2D (Black)           8: #5E5A63 (Bright Black)
 1: #C67F9E (Red)             9: #D38AA6 (Bright Red)
 2: #2FA5A0 (Green)          10: #52C4BC (Bright Green)
 3: #C2B08D (Yellow)         11: #DFBE90 (Bright Yellow)
 4: #6393D7 (Blue)           12: #A8B2D6 (Bright Blue)
 5: #D96DC0 (Magenta)        13: #EBC1F8 (Bright Magenta)
 6: #58CAF8 (Cyan)           14: #9FDBF7 (Bright Cyan)
 7: #CDE9F5 (White)          15: #F8FCFD (Bright White)
```
<!-- END GENERATED COLORS -->

---

## 2. 色選択ルール

### 背景の階層構造

1. **`#141B2D` (Deepest)**: 外枠、タブバー背景
2. **`#202A42` (Main)**: エディタ、ターミナル作業領域
3. **`#324664` (Panel)**: サイドバー、フロート
4. **`#242F48` (UI Layer)**: ポップアップ、非アクティブ領域
5. **`#5199C2` (Selection)**: アクティブ選択、ハイライト

### テキストの優先順位

1. **`#F8FCFD` (Critical)**: アクティブな強調
2. **`#9FDBF7` (High)**: 定数、アクティブ要素
3. **`#CDE9F5` (Standard)**: メインテキスト
4. **`#ADBAC9` (Low)**: 補助情報
5. **`#B1B9CD` (Auxiliary)**: Git blame、補助的な情報
6. **`#B1B9CA` (Comment)**: コメント
7. **`#B3B9C7` (Punctuation)**: 句読点
8. **`#949DB0` (Subdued)**: 非アクティブ前景、低強調テキスト
9. **`#4A4953` (Intentional Dim)**: Conceal、Flash backdrop、打ち消し線、背景塗り

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
