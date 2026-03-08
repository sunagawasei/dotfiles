---
name: add-skill
description: ディレクトリ構造とテンプレートを持つ新しいスキルを作成
---

# スキル追加

このスキルは、新しいスキルファイルを作成し、Claude Codeが再利用可能な手順を実行できるようにします。

## スキルとは

- **呼び出し可能**: `/skill-name`で明示的に呼び出し
- **手続き型**: ステップバイステップのワークフローを定義
- **再利用可能**: 共通の操作を標準化
- **参照ファイル**: オプションでテンプレートやリソースを含む

## ワークフロー

### 1. スキルのスコープを決定

**質問**: このスキルはグローバル（全プロジェクト）ですか、プロジェクト固有ですか？

- **グローバル**: `/Users/s23159/.config/claude/skills/`
  - 例: プロジェクト初期化、ルール追加、一般的なワークフロー
- **プロジェクト固有**: `<project>/.claude/skills/`
  - 例: プロジェクト固有のビルド手順、デプロイワークフロー

### 2. スキル名の決定

**質問**: "スキル名は何ですか？"

#### 命名規則

- **kebab-case**（小文字、ハイフン区切り）
- **説明的**: スキルの目的が明確
- **動詞形**: アクションを表す

#### 良い名前の例

```
init-project         # プロジェクト初期化
add-rule             # ルール追加
config-change        # 設定変更
color-validation     # カラーバリデーション
playwright-cli       # Playwright CLI実行
review-pr            # プルリクエストレビュー
deploy-production    # 本番デプロイ
```

#### 避けるべき名前

```
# ❌ 広すぎる
utils                # 何をするのか不明確

# ❌ CamelCaseやsnake_case
InitProject          # kebab-caseを使用
init_project         # kebab-caseを使用

# ❌ 名詞のみ
validation           # 動詞形を使用: validate-code
```

### 3. 簡潔な説明の作成

**質問**: "このスキルは何をしますか？（1行で）"

#### 説明のガイドライン

- **1行**: 50-80文字
- **目的を明確に**: 何を達成するか
- **動詞で始める**: 「〜を作成」「〜を実行」など

#### 良い説明の例

```yaml
description: 新しいプロジェクト用のCLAUDE.mdドキュメントを確立されたパターンに従って初期化
description: 適切なfrontmatterと構造を持つ新しいルールを作成
description: dotfiles設定変更をコミット・プッシュするワークフローを実行
description: カラー定義をWCAG 2.1 AA基準で検証
```

### 4. 参照ファイルの必要性

**質問**: "参照ファイル（テンプレート、リソース）が必要ですか？"

#### references/ディレクトリを作成する場合

- **テンプレート**: コピー可能なファイルテンプレート
- **リソース**: スキルが参照する設定ファイル
- **サンプル**: 実例やスニペット

#### 実例

```
# テンプレート
add-rule/references/template.md

# 複数テンプレート
add-skill/references/template.md
add-skill/references/template-with-references.md

# リソースファイル
playwright-cli/references/playwright.config.ts
```

### 5. ディレクトリ構造作成

#### 基本構造（references/なし）

```
.claude/skills/[skill-name]/
└── SKILL.md
```

#### 拡張構造（references/あり）

```
.claude/skills/[skill-name]/
├── SKILL.md
└── references/
    ├── template.md
    └── resource.txt
```

### 6. SKILL.mdの作成

#### 基本テンプレート

参照: `references/template.md`

```yaml
---
name: skill-name
description: Brief one-line description
---

# [スキルタイトル]

このスキルは[目的]を実行します。

## 概要

[スキルの簡単な説明]

## ワークフロー

### 1. [ステップ1]

[詳細な手順]

### 2. [ステップ2]

[詳細な手順]

## 使用例

### 例1: [シナリオ]

```bash
/skill-name
```

[期待される結果]

## トラブルシューティング

### [問題]

[解決方法]

## 関連スキル

- `/related-skill` - [説明]
```

#### Frontmatter仕様

```yaml
---
name: skill-name          # 必須: kebab-case
description: One line     # 必須: 簡潔な説明
allowed-tools: Tool(*)    # オプション: ツール制限
---
```

### 7. 内容の充実

#### 必須セクション

1. **概要**: スキルの目的と使用タイミング
2. **ワークフロー**: ステップバイステップの手順
3. **使用例**: 具体的なシナリオと実行例

#### オプションセクション

- **トラブルシューティング**: よくある問題と解決方法
- **関連スキル**: 関連する他のスキル
- **関連ドキュメント**: 追加の参考資料

### 8. テスト

作成したスキルが正しく動作するか確認：

```bash
# スキルが認識されるか確認
/skill-name

# Claude Codeがワークフローを正しく実行できるか確認
```

## テンプレート

### 基本テンプレート

参照: `references/template.md`

### references/付きテンプレート

参照: `references/template-with-references.md`

## 実例

### 実例1: シンプルなワークフロースキル

場所: `/Users/s23159/.config/claude/skills/config-change/SKILL.md`

```yaml
---
name: config-change
description: dotfiles設定変更をコミット・プッシュするワークフローを実行
---

# 設定変更ワークフロー

このスキルは、dotfiles設定変更を一貫したプロセスでGitにコミットし、リモートにプッシュします。

## ワークフロー

### 1. 変更内容の確認

```bash
git status
git diff
```

### 2. ステージング

関連ファイルをステージング

### 3. コミット

Conventional Commits形式でコミット

### 4. プッシュ

リモートリポジトリにプッシュ
```

### 実例2: references/付きスキル

場所: `/Users/s23159/.config/claude/skills/add-rule/SKILL.md`

```
add-rule/
├── SKILL.md
└── references/
    └── template.md          # コピー可能なルールテンプレート
```

**SKILL.mdから参照**:
```markdown
### 6. ルールファイル作成

#### 基本構造

参照: `references/template.md`
```

### 実例3: 複雑な手順を持つスキル

場所: `/Users/s23159/.config/claude/skills/init-project/SKILL.md`

```yaml
---
name: init-project
description: 新しいプロジェクト用のCLAUDE.mdドキュメントを確立されたパターンに従って初期化
---

# プロジェクト初期化

## ワークフロー

### 1. プロジェクトタイプの検出

現在のディレクトリを分析してプロジェクトタイプを判断

### 2. プロジェクトスコープの明確化

ユーザーに質問:
- プロジェクトの主な目的は？
- 主要な技術スタック/言語は？
- 特別な開発ワークフローは？

### 3. CLAUDE.md作成

タイプに基づいたテンプレートを選択して作成

### 4. .claude/ディレクトリ提案

複雑なルールやワークフローがある場合は提案
```

## スキルタイプ別ガイド

### ワークフロースキル

**目的**: 複数ステップの操作を自動化

**構造**:
- 明確なステップ番号
- 各ステップの詳細手順
- エラーハンドリング

**実例**: `config-change`, `deploy-production`

### テンプレート生成スキル

**目的**: 新しいファイルやプロジェクトを作成

**構造**:
- ユーザーから情報を収集
- テンプレートから生成
- references/にテンプレートを配置

**実例**: `init-project`, `add-rule`, `add-skill`

### バリデーションスキル

**目的**: データや設定を検証

**構造**:
- 検証基準の定義
- チェック項目のリスト
- エラーメッセージと修正方法

**実例**: `color-validation`

### ツール実行スキル

**目的**: 外部ツールやCLIを実行

**構造**:
- ツールの設定
- 実行パラメータ
- 結果の解釈

**実例**: `playwright-cli`

## トラブルシューティング

### スキルが認識されない

```bash
# スキル一覧を確認
ls -la /Users/s23159/.config/claude/skills/
ls -la .claude/skills/

# Frontmatterを確認
cat .claude/skills/[skill-name]/SKILL.md
```

**チェック項目**:
- [ ] `SKILL.md`ファイル名が正しい（大文字）
- [ ] Frontmatterが存在する（`---`で囲まれている）
- [ ] `name:`と`description:`が定義されている
- [ ] ディレクトリ名とname:が一致している

### Frontmatter構文エラー

```yaml
# ❌ 間違い
---
name: skill name  # スペースは使えない
description: Very long description that spans multiple lines without proper YAML syntax
---

# ✅ 正しい
---
name: skill-name
description: Concise one-line description
---
```

### ワークフローが不明確

**問題**: ステップが曖昧で実行できない

**解決方法**:
1. 各ステップを具体的に記述
2. コマンド例を含める
3. 期待される結果を明記
4. エラーハンドリングを追加

**例**:

```markdown
### 悪い例 ❌

1. ファイルを確認する

### 良い例 ✅

1. ファイルの状態を確認

```bash
git status
```

期待される結果: 変更されたファイルの一覧が表示される
```

## ベストプラクティス

### 1. 単一責任原則

各スキルは1つの明確な目的を持つ

```
# ✅ 良い例
add-rule         # ルール追加のみ
add-skill        # スキル追加のみ

# ❌ 悪い例
add-everything   # 複数の異なる機能
```

### 2. 明確なワークフロー

ステップを番号付きで明確に

```markdown
## ワークフロー

### 1. 準備

[詳細]

### 2. 実行

[詳細]

### 3. 検証

[詳細]
```

### 3. 実例を含める

実際の使用例を示す

```markdown
## 使用例

### 例1: Go言語プロジェクト

```bash
/init-project
```

Claude Codeが質問:
1. "プロジェクトの主な目的は？" → "CLI tool"
2. "主要な技術スタック/言語は？" → "Go 1.21+"
```

### 4. references/の適切な使用

テンプレートは直接コピー可能な形式で

```markdown
# references/template.mdの内容

---
name: skill-name
description: Description
---

# [スキルタイトル]

[コピー可能なテンプレート]
```

### 5. 関連スキル・ドキュメントへのリンク

関連情報へのポインタを提供

```markdown
## 関連スキル

- `/init-project` - プロジェクトにCLAUDE.mdを初期化
- `/add-rule` - プロジェクトにルールを追加

## 関連ドキュメント

- `.claude/docs/git-workflow.md` - Gitワークフロー詳細
```

## 関連スキル

- `/init-project` - プロジェクトにCLAUDE.mdを初期化
- `/add-rule` - プロジェクトにルールを追加
