# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfiles設定リポジトリです。バージョン管理にはGitを使用しています（ローカルのみ、リモートなし）。

## 主要な設定

### Neovim (LazyVim)
- 場所: `nvim/`
- フレームワーク: LazyVim（事前設定済みのNeovimセットアップ）
- Luaフォーマット: `nvim/stylua.toml`でstyluaを使用

### Weztermターミナル（メイン）
- 場所: `wezterm/`
- 主要設定ファイル: `wezterm/wezterm.lua`
- キーバインド設定: `wezterm/keybinds.lua`
- テーマ: カスタムVercelダークカラー

### Ghosttyターミナル（サブ使用）
- 場所: `ghostty/`
- 主要設定ファイル: `ghostty/config`
- キーバインドリファレンス: `ghostty/command.md`
- テーマ: カスタムVercelダークカラー

### カラーシステム
- ドキュメント: `vercel-geist-colors.md`
- アプリケーション間で一貫したVercel Geistカラースキーム
- 10色のカラースケール（Gray、Blue、Red、Amber、Green、Teal、Purple、Pink）

### シェル設定
- Starshipプロンプト: `starship.toml`
- Fishシェル: `fish/`
- シェルユーティリティ: `shell/`

## 一般的な開発タスク

### Raycast拡張機能
`raycast/extensions/`内のRaycast拡張機能を扱う場合：
```bash
# 拡張機能のビルド
npm run build  # または ray build

# 開発モード
npm run dev    # または ray develop

# コードのリント
npm run lint   # または ray lint

# リントエラーの修正
npm run fix-lint  # または ray lint --fix
```

### 設定変更
1. 設定ファイルを直接編集
2. 各アプリケーションで変更をテスト
3. 説明的な日本語または英語のメッセージでコミット
4. 最近のコミットスタイルの例：
   - "キーバインド設定を追加し、^wでコマンドラインを編集できるようにしました"
   - "プレフィックスキーをCtrl+qに変更し、C-bを解除"
   - "chore: `raycast/extensions`を除外"

### Gitワークフロー
- `main`ブランチで直接作業
- リモートリポジトリは設定されていない
- バックアップファイル（*.bak）は一部のツールによって自動的に作成される

## 重要なパス
- Neovim設定: `nvim/init.lua`および`nvim/lua/`
- Wezterm設定: `wezterm/wezterm.lua`および`wezterm/keybinds.lua`
- Ghostty設定: `ghostty/config`
- ターミナルテーマ: Vercel Geistカラーベース
- Raycast拡張機能: `raycast/extensions/*/`

## アーキテクチャの注意点
- 集中化されたビルドシステムやセットアップスクリプトなし
- 各アプリケーションが独自の設定を管理
- 手動での設定デプロイ（stowやシンボリックリンク管理なし）
- 設定全体で日本語サポート
- Vercelのデザインシステムを使用した一貫性のあるダークテーマ