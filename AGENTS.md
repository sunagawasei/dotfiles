# AGENTS.md

このファイルは、このリポジトリでコードを扱う際の AI エージェントへの指針を提供します。

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfiles設定リポジトリです。バージョン管理にはGitを使用し、GitHubでリモート管理されています。

## 言語設定

**重要**: すべての応答は日本語で行ってください。

## 主要な設定ディレクトリ

| ディレクトリ | 用途 |
|-------------|------|
| `nvim/` | Neovim (LazyVim) 設定 |
| `wezterm/` | WezTerm ターミナル設定 |
| `claude/` | Claude Code 設定 |
| `codex/` | OpenAI Codex CLI 設定 |
| `fish/` | Fish Shell 設定 |
| `zsh/` | Zsh 設定 |
| `git/` | Git グローバル設定 |
| `lazygit/` | LazyGit TUI 設定 |

## 重要なパス

- **Neovim**: `nvim/init.lua` および `nvim/lua/`
- **WezTerm**: `wezterm/wezterm.lua` および `wezterm/keybinds.lua`
- **Claude Code**: `claude/settings.json`
- **Codex CLI**: `codex/config.toml` および `codex/AGENTS.md`

## 設定変更のワークフロー

1. 設定ファイルを直接編集
2. 各アプリケーションで動作確認
3. 日本語または英語で説明的なコミットメッセージを作成
4. 変更をコミット

### コミットメッセージの例

- `feat(nvim): Go言語サポートを追加`
- `perf(options): updatetimeを1000msに変更してCPU使用率を改善`
- `docs: 文字ジャンプコマンドの説明を追加`

## アーキテクチャの特徴

- **モジュラー構造**: 各ツールが独立した設定ディレクトリを持つ
- **手動管理**: stowや自動化ツールは使用せず、直接編集
- **統一テーマ**: カスタムカラーによる一貫性
- **日本語サポート**: 設定全体で日本語環境をサポート
- **パフォーマンス重視**: 各ツールで最適化された設定

## 参照ドキュメント

- `COLOR-SYSTEM.md`: カラーシステムのガイドライン
- `KEYMAPS.md`: キーバインド一覧
- `nvim/CLAUDE.md`: Neovim 詳細設定

## 注意事項

- キーバインドや操作を追加した場合は `KEYMAPS.md` に記録する
- カラー選択時は `COLOR-SYSTEM.md` のガイドラインに従う
- プラグインやツール追加時は適切な設定ディレクトリに配置する
