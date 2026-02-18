# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfiles設定リポジトリです。バージョン管理にはGitを使用し、GitHubでリモート管理されています。

## アーキテクチャの特徴

- **モジュラー構造**: 各ツールが独立した設定ディレクトリを持つ
- **手動管理**: stowや自動化ツールは使用せず、直接編集
- **統一テーマ**: カスタムAbyssal Tealカラーによる一貫性
- **日本語サポート**: 設定全体で日本語環境をサポート
- **パフォーマンス重視**: 各ツールで最適化された設定

## 主要技術スタック

- **エディタ**: Neovim (LazyVim) - 詳細は `nvim/CLAUDE.md`
- **ターミナル**: WezTerm, Fish/Zsh, Starship
- **Git**: LazyGit, GitUI, GitHub連携
- **AI**: Claude Code, Cursor CLI
- **カラーシステム**: Abyssal Teal (32色パレット)

## 重要なパス

- `nvim/` - Neovim設定 (詳細: `nvim/CLAUDE.md`)
- `wezterm/` - WezTerm設定
- `colors/abyssal-teal.toml` - 統一カラー定義
- `COLOR-SYSTEM.md` - カラーシステムドキュメント
- `KEYMAPS.md` - キーバインドリファレンス

## リモートリポジトリ

- GitHub: `git@github.com:sunagawasei/dotfiles.git`
- メインブランチ: `main`

## 関連ドキュメント

設定の詳細やワークフローについては、`.claude/` ディレクトリ内のファイルを参照してください:

- **ルール**: `.claude/rules/` - ファイル種別ごとの規約（自動適用）
- **スキル**: `.claude/skills/` - 再利用可能な手順（`/config-change`, `/color-validation`）
- **ドキュメント**: `.claude/docs/` - 詳細な技術資料
