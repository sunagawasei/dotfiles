---
name: skill-name
description: Brief one-line description of what this skill does
---

# [スキルタイトル]

このスキルは[目的]を実行します。

## 概要

[スキルの簡単な説明を2-3文で記述]

## ディレクトリ構造

```
skills/skill-name/
├── SKILL.md
└── references/
    ├── template.md          # [テンプレートファイル1の説明]
    ├── config.yaml          # [リソースファイルの説明]
    └── example.txt          # [実例ファイルの説明]
```

## ワークフロー

### 1. [ステップ1の名前]

[ステップ1の詳細な手順]

**質問**: "[ユーザーに確認すべき内容]"

### 2. [ステップ2の名前]

[ステップ2の詳細な手順]

**参照**: `references/template.md`でテンプレートを確認

### 3. [ステップ3の名前]

[ステップ3の詳細な手順]

```bash
# references/内のファイルを使用
cp references/template.md target-location/
```

### 4. [ステップ4の名前]

[ステップ4の詳細な手順]

## テンプレートファイル

### template.md

参照: `references/template.md`

[テンプレートの使い方の説明]

### config.yaml

参照: `references/config.yaml`

[設定ファイルの使い方の説明]

## 使用例

### 例1: [シナリオ名]

```bash
/skill-name
```

**Claude Codeの動作**:
1. [ステップ1の説明]
2. [ステップ2の説明]
3. references/template.mdからテンプレートをコピー
4. [ステップ4の説明]

[期待される結果]

### 例2: [別のシナリオ名]

```bash
/skill-name
```

[期待される結果や動作の説明]

## トラブルシューティング

### references/ファイルが見つからない

**症状**: テンプレートファイルが見つからないエラー

**解決方法**:

```bash
# ディレクトリ構造を確認
ls -la .claude/skills/skill-name/references/

# references/ディレクトリを作成
mkdir -p .claude/skills/skill-name/references/
```

### [問題2]

**症状**: [どのような問題が発生するか]

**解決方法**:

[解決手順の説明]

## ベストプラクティス

1. **テンプレートは直接コピー可能に**: references/内のファイルはそのままコピーして使える形式にする
2. **適切なファイル配置**: テンプレートはreferences/、実行可能スクリプトはscripts/に配置
3. **明確なファイル名**: references/内のファイル名は用途が明確にわかるようにする

## 関連スキル

- `/related-skill-1` - [説明]
- `/related-skill-2` - [説明]

## 関連ドキュメント

- `path/to/doc.md` - [説明]
- `references/template.md` - テンプレートファイル
- `references/config.yaml` - 設定ファイル
