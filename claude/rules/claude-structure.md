---
paths:
  - "**/.claude/**"
  - "**/CLAUDE.md"
  - "**/claude/rules/**"
  - "**/claude/skills/**"
  - "**/claude/docs/**"
  - "**/claude/agents/**"
---

# Claude Code構造規約

このルールは、Claude Code構造ファイル（rules、skills、agents、docs、CLAUDE.md）を編集する際に自動的に適用されます。

グローバル設定（`~/.claude/`。`$CLAUDE_CONFIG_DIR` で変更可）は全プロジェクト共通、プロジェクト固有設定（`<project>/.claude/`）は個別プロジェクト用。両者とも`rules/` `skills/` `docs/`の同じ構造を持つ。

## クイックリファレンス

### Rules（ルール）

**目的**: `paths:`のパターンに基づいて自動適用されるガイドライン（frontmatter必須）

**ポイント**:
- kebab-case命名（例: `golang.md`, `commit-messages.md`）
- 簡潔に保つ（目安は末尾「ベストプラクティス」参照）

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
- kebab-case命名（例: `session-harvest`, `split-commits`）
- ステップバイステップのワークフロー
- グローバル（`claude/skills/`）への新規追加時は`.gitignore`のallowlist（`!claude/skills/<name>/`）登録が必須。`claude/skills/*`が既定ignoreのため、未登録だと黙って未追跡になる（`git status --short`で`??`と出ることを確認）

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

**目的**: このリポジトリで作業するClaude Codeへの指針。網羅的なドキュメントではない

**削る / 残す判定基準**:
- **削る（コードベースから導出可能）**: ディレクトリ構造・技術スタックや依存一覧・標準的なbuild/test/lintコマンド・ソースからコピーしたAPIシグネチャ/型/スキーマ・アーキ概要・モデルが既に従う一般的ベストプラクティス・lint設定やCIが機械的に強制しているルール
- **残す（コードベースから導出できない）**: gotchaと失敗契約（「Xすると黙ってYになる」）・設計判断とその理由・言語やツールの既定と異なる非標準の慣習・エージェントへの安全上の禁止事項・リポジトリの作法（ブランチ命名・PR規約・コミット形式）・ドメイン用語集・推測できないbuild/testコマンド（非標準スクリプト・必須フラグ・環境変数）・他所にあるコンテキストへのポインタ

**構造**（gotcha中心）:
```markdown
# CLAUDE.md

## リポジトリ概要

[このリポジトリが何のためのものか。1〜3行]

## 踏んだ罠

[「Xすると黙ってYになる」型の失敗契約]

## 標準と違う慣習

[言語/ツールの既定と異なる非標準の慣習。コードだけ読むと誤解する箇所]

## 推測できないコマンド

[非標準スクリプト・必須フラグ・環境変数・事前準備。標準的な呼び出しは書かない]

## 安全上の禁止事項

[触ってはいけないファイル、してはいけない操作]

## 関連ドキュメント

[詳細ドキュメントへのポインタ]
```

**書かない**: ディレクトリツリー、依存一覧、標準コマンド、アーキ概要、一般的ベストプラクティス — これらは`ls`とマニフェストから導出できるため

**ポイント**:
- Frontmatterなし
- 簡潔に保つ（目安は末尾「ベストプラクティス」参照）
- 詳細は`.claude/`に分離

### CLAUDE.local.md（個人ローカル用）

**目的**: プロジェクト固有だが git にコミットしない個人メモ（ローカル環境のプロファイル名・個人パス等）

**ポイント**:
- リポジトリルートに置く。`CLAUDE.md` と並んで自動読込される（公式サポート、[docs/memory](https://code.claude.com/docs/en/memory)）
- **`.gitignore` への追加は手動**（自動では ignore されない）
- `.claude/CLAUDE.md` は root `CLAUDE.md` と同一の「Project instructions」スコープ（チーム共有・コミット前提）であり、ローカル専用の置き場としては使わない

## 各ファイル種別の骨格

### Rule

```yaml
---
paths:
  - "**/*.go"
  - "**/go.mod"
---

# ルールタイトル

[内容]
```

### Skill

```
skills/[name]/SKILL.md
skills/[name]/references/   # オプション
```

```yaml
---
name: skill-name
description: 一行説明
---

# スキルタイトル
```

### Doc

```markdown
# ドキュメントタイトル

[内容]
```

## ベストプラクティス

1. **単一責任**: 1ファイル = 1つの規約/ワークフロー
2. **簡潔に**: 読む側が必要な判断をできる最小限に保つ。詳細は`references/`や`docs/`へ切り出す（Docsは長文OK・制限なし）
3. **例示は最小限**: 骨格が伝わる1例に留める。良い例/悪い例の対比や網羅的なサンプルは探索空間を狭めるので置かない
4. **テストする**: 作成後、必ず動作確認
5. **DRY原則**: 重複を避け、参照を使用
6. **段階的開示**: 基本→詳細の順で情報提供
