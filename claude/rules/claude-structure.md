---
paths:
  - "**/.claude/**"
  - "**/CLAUDE.md"
  - "**/claude/rules/**"
  - "**/claude/skills/**"
  - "**/claude/docs/**"
---

# Claude Code構造規約

このルールは、Claude Code構造ファイル（rules、skills、docs、CLAUDE.md）を編集する際に自動的に適用されます。

## クイックリファレンス

### Rules（ルール）

**目的**: ファイルパターンに基づいて自動適用されるガイドライン

**構造**:
```yaml
---
paths:
  - "**/*.ext"
  - "**/pattern/**"
trigger: "optional-keyword"
---

# ルール名

[ガイドライン内容]
```

**ポイント**:
- Frontmatter必須（YAMLフォーマット）
- `paths:`は配列形式（`-`で始まる）
- kebab-case命名（例: `golang.md`, `commit-messages.md`）
- 50-100行程度の簡潔な内容

### Skills（スキル）

**目的**: `/skill-name`で呼び出し可能な手続き型ワークフロー

**構造**:
```yaml
---
name: skill-name
description: Brief one-line description
---

# スキルタイトル

## 概要
## ワークフロー
## 使用例
```

**ポイント**:
- ディレクトリ構造: `skills/[name]/SKILL.md`
- オプション: `references/`サブディレクトリでテンプレート提供
- kebab-case命名（例: `init-project`, `add-rule`）
- ステップバイステップのワークフロー

### Docs（ドキュメント）

**目的**: 包括的な参照ドキュメント

**構造**:
```markdown
# ドキュメントタイトル

[詳細な説明]
```

**ポイント**:
- Frontmatterなし（純粋なMarkdown）
- 長文OK（詳細な説明向け）
- kebab-case命名（例: `tool-configurations.md`, `git-workflow.md`）

### CLAUDE.md

**目的**: プロジェクト概要とClaude Codeへの指針

**構造**:
```markdown
# CLAUDE.md

## リポジトリ概要
## アーキテクチャ
## 開発コマンド
## 重要なパス
## 関連ドキュメント
```

**ポイント**:
- Frontmatterなし
- 簡潔に保つ（50行以内推奨）
- 詳細は`.claude/`に分離

## 確立されたパターンに従う

### ✅ 良いパターン

#### Rule作成時

```yaml
---
paths:
  - "**/*.go"
  - "**/go.mod"
---

# Go言語規約

**コード生成時は、明示的に指定がない限りGolangを使用してください**

## プロジェクト構造

- `go.mod`: Goモジュール定義
```

#### Skill作成時

```
skills/config-change/
├── SKILL.md
└── references/          # オプション
    └── template.md
```

```yaml
---
name: config-change
description: dotfiles設定変更のワークフロー
---

# 設定変更ワークフロー

## ワークフロー

### 1. 変更内容の確認
### 2. ステージング
### 3. コミット
```

#### Doc作成時

```markdown
# Gitワークフロー

このドキュメントは、dotfilesリポジトリでのGit運用方法を定義します。

## ブランチ戦略
## コミットワークフロー
## プッシュ頻度
```

### ❌ 避けるべきパターン

#### Frontmatter構文エラー

```yaml
# ❌ 間違い
---
paths:
- "**/*.go"  # インデントが足りない
---

# ✅ 正しい
---
paths:
  - "**/*.go"  # スペース2つでインデント
---
```

#### 広すぎるPathパターン

```yaml
# ❌ 避ける
---
paths:
  - "**/*"  # 全てのファイルに適用される
---

# ✅ 適切
---
paths:
  - "**/*.go"
  - "**/go.mod"
---
```

#### 絶対パス使用

```yaml
# ❌ 避ける
---
paths:
  - "/Users/s23159/project/src/main.go"  # 環境依存
---

# ✅ 適切
---
paths:
  - "**/src/**/*.go"  # 相対パターン
---
```

#### スキル名の不適切な形式

```
# ❌ 避ける
InitProject         # CamelCase
init_project        # snake_case
project             # 曖昧すぎる

# ✅ 適切
init-project        # kebab-case、明確
```

## よくある落とし穴

### 1. YAML構文エラー

**問題**: Frontmatterが正しく認識されない

**チェック項目**:
- [ ] 開始と終了の`---`が存在する
- [ ] `paths:`の後にコロン（`:`）がある
- [ ] 各パスが`-`で始まる
- [ ] インデントがスペース2つ

### 2. Pathパターンがマッチしない

**問題**: ルールが期待通りに適用されない

**確認方法**:
```bash
# パターンをテスト
find . -path "**/pattern/**" -type f

# glob展開をテスト（zsh）
echo **/*.go
```

### 3. スキルが認識されない

**問題**: `/skill-name`で呼び出せない

**チェック項目**:
- [ ] ファイル名が`SKILL.md`（大文字）
- [ ] Frontmatterに`name:`と`description:`がある
- [ ] ディレクトリ名とname:が一致
- [ ] kebab-case命名

### 4. ファイル配置が間違っている

**正しい配置**:

グローバル設定:
```
/Users/s23159/.config/claude/
├── rules/
├── skills/
└── docs/
```

プロジェクト固有設定:
```
<project>/.claude/
├── rules/
├── skills/
└── docs/
```

**間違った配置**:
```
# ❌ 避ける
/Users/s23159/.config/.claude/    # グローバル設定には使わない
```

## 実例参照

### 既存のRules

- `/Users/s23159/.config/.claude/rules/golang.md` - 基本的なルール
- `/Users/s23159/.config/.claude/rules/commit-messages.md` - trigger付きルール
- `/Users/s23159/.config/.claude/rules/color-system.md` - 複雑なルール

### 既存のSkills

- `/Users/s23159/.config/claude/skills/init-project/SKILL.md` - 複雑なワークフロー
- `/Users/s23159/.config/claude/skills/add-rule/SKILL.md` - references/付き
- `/Users/s23159/.config/.claude/skills/config-change/SKILL.md` - シンプルなワークフロー

### 既存のDocs

- `/Users/s23159/.config/.claude/docs/tool-configurations.md` - 包括的ドキュメント
- `/Users/s23159/.config/.claude/docs/git-workflow.md` - ワークフロー詳細

## テンプレート

### 新しいRuleを追加

```bash
/add-rule
```

テンプレート: `/Users/s23159/.config/claude/skills/add-rule/references/template.md`

### 新しいSkillを追加

```bash
/add-skill
```

テンプレート:
- `/Users/s23159/.config/claude/skills/add-skill/references/template.md`
- `/Users/s23159/.config/claude/skills/add-skill/references/template-with-references.md`

### 新しいプロジェクトを初期化

```bash
/init-project
```

## ベストプラクティス

1. **単一責任**: 1ファイル = 1つの規約/ワークフロー
2. **簡潔に**: Rules/Skillsは50-100行、Docsは制限なし
3. **実例を含める**: 良い例と悪い例を示す
4. **テストする**: 作成後、必ず動作確認
5. **DRY原則**: 重複を避け、参照を使用
6. **段階的開示**: 基本→詳細の順で情報提供

## 参考

- `/init-project` - 新しいプロジェクトのCLAUDE.md作成
- `/add-rule` - 新しいルール追加
- `/add-skill` - 新しいスキル追加
