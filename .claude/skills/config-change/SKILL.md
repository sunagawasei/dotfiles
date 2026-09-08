---
name: config-change
description: dotfiles設定変更のワークフロー（編集、テスト、コミット、プッシュ）
---

# 設定変更ワークフロー

このスキルは、dotfilesの設定を変更してコミットするまでの標準的なワークフローを定義します。

## ワークフロー手順

### 1. 設定ファイルの直接編集

該当する設定ファイルを直接編集します：

```bash
# 例: Neovim設定の編集
nvim nvim/lua/plugins/plugin-name.lua

# 例: WezTerm設定の編集
nvim wezterm/wezterm.lua

# 例: カラー定義の編集
nvim colors/ghost-visor.toml
```

### 2. 各アプリケーションでの動作確認

変更後、該当するアプリケーションで動作を確認します。

#### Neovim

```vim
:Lazy sync          # プラグインの同期
:source %           # 現在のファイルを再読み込み
:checkhealth        # ヘルスチェック
```

#### WezTerm

- WezTermを再起動して設定を反映
- キーバインドが正しく動作するか確認

#### カラーシステム

```bash
# バリデーションスクリプトで整合性確認
cd scripts
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-colors
```

#### シェル設定（Zsh）

```bash
source ~/.zshrc
```

### 3. 説明的なコミットメッセージの作成

Conventional Commits形式に従い、日本語または英語で記述：

#### 形式

```
<type>(<scope>): <subject>
```

#### Type

- `feat`: 新機能追加
- `fix`: バグ修正
- `docs`: ドキュメント変更
- `refactor`: リファクタリング
- `perf`: パフォーマンス改善
- `chore`: ビルドやツールの変更

#### Scope

- `nvim`: Neovim設定
- `wezterm`: WezTerm設定
- `theme`: カラーテーマ
- `zsh`/`fish`: シェル設定
- `git`: Git設定

#### 例

```
feat(nvim): CodeCompanionプラグインを追加
fix(wezterm): キーバインドの競合を解消
docs: KEYMAPS.mdにターミナル操作を追加
refactor(theme): カラー定義をTOML形式に統一
```

### 4. 変更をコミット

```bash
# ステージング
git add nvim/lua/plugins/plugin-name.lua

# コミット
git commit -m "feat(nvim): 新しいプラグインを追加"
```

または、Claude Codeの`/commit`コマンドを使用（commitサブエージェントが自動的にメッセージを生成）：

```bash
/commit
```

### 5. リモートリポジトリへプッシュ

**プッシュはユーザーの指示があった場合のみ**行う（Claudeが自発的にプッシュしない）：

```bash
git push origin main
```

## ツール別テストコマンド

### Neovim

```vim
:Lazy sync              # プラグイン同期
:ConformInfo            # フォーマッター情報
:checkhealth            # 全体のヘルスチェック
:checkhealth treesitter # Treesitter確認
:checkhealth neotest    # Neotest確認
```

### カラーシステム

```bash
cd scripts
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-colors
```

### Git設定

```bash
git config --list
git status
```

### Raycast拡張機能

```bash
cd raycast/extensions/<extension-name>
npm run lint
npm run build
```

## ベストプラクティス

- **プッシュはユーザー指示があった時のみ**: 自発的にリモートへ同期しない

## トラブルシューティング

### プラグインが読み込まれない

```vim
:Lazy log    # エラーログを確認
:messages    # エラーメッセージを確認
```

### カラーの整合性エラー

```bash
# バリデーションスクリプトで詳細を確認
cd scripts
go run ./cmd/generate-color-inventory
go run ./cmd/generate-color-inventory --check
go run ./cmd/verify-colors
```

### Git コンフリクト

```bash
# 現在の状態を確認
git status

# リモートから最新を取得
git pull origin main

# コンフリクトを解決してコミット（パスを明示する。`git add .` は使わない）
git add <解決したファイル>
git commit -m "fix: コンフリクトを解決"
```

## 関連ドキュメント

- `.claude/rules/color-system.md` - カラーシステム規約
- `.claude/rules/neovim.md` - Neovim設定規約
- `nvim/CLAUDE.md` - Neovim詳細ドキュメント
- `COLOR-SYSTEM.md` - カラーシステムガイドライン
