---
name: init-project
description: 新しいプロジェクト用のCLAUDE.mdドキュメントを確立されたパターンに従って初期化
---

# プロジェクト初期化

このスキルは、新しいプロジェクトにCLAUDE.mdを作成し、Claude Codeが効果的にコードベースを理解できるようにします。

## ワークフロー

### 1. プロジェクトタイプの検出

現在のディレクトリを分析してプロジェクトタイプを判断：

```bash
# Go project
ls go.mod go.sum

# Node.js project
ls package.json

# Python project
ls pyproject.toml setup.py requirements.txt

# Dotfiles
ls .zshrc .bashrc
```

### 2. プロジェクトスコープの明確化

ユーザーに以下を質問：

- **プロジェクトの主な目的は？** (例: Web app、CLI tool、Library、Configuration)
- **主要な技術スタック/言語は？**
- **特別な開発ワークフローは？**

### 3. CLAUDE.md作成

#### 基本テンプレート

```markdown
# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

## リポジトリ概要

[プロジェクトの説明を1-2文で記述]

## アーキテクチャ

### 技術スタック
- **言語**: [主要言語]
- **フレームワーク**: [使用フレームワーク]
- **主要ライブラリ**: [重要な依存関係]

### ディレクトリ構造
```
[プロジェクト]/
├── [主要ディレクトリ1]/     # 説明
├── [主要ディレクトリ2]/     # 説明
└── [設定ファイル]           # 説明
```

## 開発コマンド

### セットアップ
```bash
# 依存関係のインストール
[インストールコマンド]
```

### 開発
```bash
# 開発サーバー起動
[開発コマンド]

# テスト実行
[テストコマンド]

# ビルド
[ビルドコマンド]
```

## 重要なパス

- `[重要ファイル1]` - [説明]
- `[重要ファイル2]` - [説明]
- `[重要ディレクトリ]/` - [説明]

## 関連ドキュメント

[プロジェクト固有のドキュメントがあればリンク]
```

### 4. プロジェクトタイプ別のカスタマイズ

#### Go Project

```markdown
## 開発コマンド

### セットアップ
```bash
# 依存関係のインストール
go mod download
```

### 開発
```bash
# アプリケーション実行
go run main.go

# テスト実行
go test ./...

# ビルド
go build -o bin/app
```

## テスト

- テストファイル: `*_test.go`パターン
- カバレッジ: `go test -cover ./...`
- ベンチマーク: `go test -bench=. ./...`
```

#### Node.js/TypeScript Project

```markdown
## 開発コマンド

### セットアップ
```bash
# 依存関係のインストール
npm install
# または
pnpm install
```

### 開発
```bash
# 開発サーバー起動
npm run dev

# テスト実行
npm test

# ビルド
npm run build

# リント
npm run lint
```

## スクリプト

`package.json`のscriptsセクションを参照
```

#### Python Project

```markdown
## 開発コマンド

### セットアップ
```bash
# 仮想環境作成
python -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate  # Windows

# 依存関係のインストール
pip install -r requirements.txt
```

### 開発
```bash
# アプリケーション実行
python main.py

# テスト実行
pytest

# リント
flake8 .
```
```

#### Dotfiles Project

```markdown
## リポジトリ概要

個人用設定ファイル（dotfiles）を管理するリポジトリです。

## 主要な設定

### エディタ
- **[エディタ名]**: `[ディレクトリ]/` - [説明]

### シェル
- **[シェル名]**: `[ディレクトリ]/` - [説明]

### ツール
- **[ツール名]**: `[ディレクトリ]/` - [説明]

## セットアップ

[インストール/シンボリックリンク作成手順]

## カスタマイズ

[設定変更の方法]
```

### 5. .claude/ディレクトリ提案

プロジェクトに複雑なルールやワークフローがある場合：

```bash
# プロジェクトレベル設定ディレクトリを作成
mkdir -p .claude/rules
mkdir -p .claude/skills
mkdir -p .claude/docs
```

そして、CLAUDE.mdに以下を追加：

```markdown
## 関連ドキュメント

プロジェクト固有の規約やワークフローについては、`.claude/`ディレクトリ内のファイルを参照してください：

- **ルール**: `.claude/rules/` - ファイル種別ごとの規約
- **スキル**: `.claude/skills/` - 再利用可能な手順
- **ドキュメント**: `.claude/docs/` - 詳細な技術資料
```

## 実例

### 実例1: シンプルなCLI Tool（Go）

```markdown
# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

## リポジトリ概要

コマンドラインツールで、[機能の説明]を提供します。

## アーキテクチャ

### 技術スタック
- **言語**: Go 1.21+
- **CLI Framework**: cobra

### ディレクトリ構造
```
cli-tool/
├── cmd/          # コマンド定義
├── internal/     # 内部パッケージ
├── main.go       # エントリーポイント
└── go.mod        # モジュール定義
```

## 開発コマンド

### セットアップ
```bash
go mod download
```

### 開発
```bash
# ツール実行
go run main.go [command]

# テスト
go test ./...

# ビルド
go build -o bin/tool
```

## 重要なパス

- `cmd/` - コマンド定義
- `internal/` - 内部ロジック
- `main.go` - エントリーポイント
```

### 実例2: Dotfiles（現在のリポジトリ）

参照: `/Users/s23159/.config/CLAUDE.md`

## ベストプラクティス

1. **簡潔に保つ**: CLAUDE.mdは50行以内を目標
2. **重要な情報に焦点**: 開発者が最初に知る必要があることのみ
3. **詳細は分離**: 複雑な規約は`.claude/`に移動
4. **実例を含める**: コマンド例を具体的に記述
5. **最新に保つ**: プロジェクト構造変更時に更新

## トラブルシューティング

### CLAUDE.mdが長すぎる

詳細な情報を`.claude/docs/`に移動：

```bash
mkdir -p .claude/docs
# 詳細ドキュメントを作成
nvim .claude/docs/architecture.md
```

CLAUDE.mdからリンク：

```markdown
## 関連ドキュメント

- 詳細なアーキテクチャ: `.claude/docs/architecture.md`
```

### プロジェクト固有のルールが多い

`.claude/rules/`にルールファイルを作成：

```bash
/add-rule
```

## 関連スキル

- `/add-rule` - プロジェクトにルールを追加
- `/add-skill` - プロジェクトにスキルを追加
