# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) への指針を提供します。

## リポジトリ概要

macOS上の開発ツール設定を管理する個人用dotfiles設定リポジトリです。バージョン管理にはGitを使用し、GitHubでリモート管理されています。

## 主要な設定

### エディタ

- **Neovim (LazyVim)**: `nvim/` - メインエディタ、詳細設定は `nvim/CLAUDE.md` を参照
- **Zed**: `zed/` - モダンコードエディタ、カスタム設定とVercelテーマ

### ターミナルエミュレーター

- **WezTerm (メイン)**: `wezterm/` - GPU加速、カスタムキーバインド
- **Ghostty (サブ)**: `ghostty/` - macOSネイティブ、軽量

### シェル設定

- **Fish Shell**: `fish/` - メインシェル
- **Zsh**: `zsh/` - 従来シェル、豊富な補完設定
- **Starship**: `starship.toml` - プロンプト設定

### 開発ツール

- **Git**: `git/` - グローバル設定
- **LazyGit**: `lazygit/` - Git TUI設定
- **GitUI**: `gitui/` - Git GUI設定
- **tmux**: `tmux/` - ターミナル多重化
- **Karabiner**: `karabiner/` - キーボード設定
- **Claude Code**: `claude/` - AI統合開発環境、設定ディレクトリは`$CLAUDE_CONFIG_DIR`で指定

### カラーシステム

- **統一テーマ**: Vercel Geistカラースキーム
- **ドキュメント**: `vercel-geist-colors.md`
- **適用範囲**: 全アプリケーション間で一貫性のあるダークテーマ

## 一般的な開発タスク

### Neovim設定

詳細な設定とキーマップは `nvim/CLAUDE.md` を参照してください。LazyVimベースの設定で、Go言語サポート、テスト実行、AI統合などを含みます。

### Raycast拡張機能

`raycast/extensions/` 内の拡張機能を扱う場合：

```bash
npm run build    # ビルド
npm run dev      # 開発モード
npm run lint     # リント
```

### 設定変更のワークフロー

1. 設定ファイルを直接編集
2. 各アプリケーションで動作確認
3. 日本語または英語で説明的なコミットメッセージを作成
4. 変更をコミット

最近のコミット例：

- "feat(nvim): Go言語サポートを追加"
- "perf(options): updatetimeを1000msに変更してCPU使用率を改善"
- "docs: 文字ジャンプコマンドの説明を追加"

### Gitワークフロー

- `main`ブランチで直接作業
- リモートリポジトリ: `git@github.com:sunagawasei/dotfiles.git`
- 定期的にプッシュして同期

## 重要なパス

- **Neovim**: `nvim/init.lua` および `nvim/lua/`
- **WezTerm**: `wezterm/wezterm.lua` および `wezterm/keybinds.lua`
- **Ghostty**: `ghostty/config`
- **Zed**: `zed/settings.json` および `zed/keymap.json`
- **Git**: `git/config` および `git/ignore`
- **Claude Code**: `claude/settings.json` および `.claude/settings.local.json`（プロジェクトレベル）

## アーキテクチャの特徴

- **モジュラー構造**: 各ツールが独立した設定ディレクトリを持つ
- **手動管理**: stowや自動化ツールは使用せず、直接編集
- **統一テーマ**: Vercel Geistカラーによる一貫性
- **日本語サポート**: 設定全体で日本語環境をサポート
- **パフォーマンス重視**: 各ツールで最適化された設定

## Claude Code設定

### 設定ディレクトリ

- **ユーザー設定**: `$CLAUDE_CONFIG_DIR = /Users/s23159/.config/claude`
- **プロジェクト設定**: `.claude/` （各プロジェクトのルートディレクトリ）

### 設定ファイル

- **ユーザーレベル**: `claude/settings.json` - 全プロジェクト共通設定
- **プロジェクトレベル**: `.claude/settings.local.json` - プロジェクト固有設定

## 注意事項

- キーバインドや操作を追加した場合は `KEYMAPS.md` に記録する
- カラー選択時は `vercel-geist-colors.md` のガイドラインに従う
- プラグインやツール追加時は適切な設定ディレクトリに配置する

