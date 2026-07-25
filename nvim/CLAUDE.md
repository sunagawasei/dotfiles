# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

LazyVimベースのNeovim設定。個人用にカスタマイズされており、日本語コメントを含む。

## アーキテクチャ

### プラグイン管理

プラグインは`lua/plugins/`に個別ファイルとして配置：
- LazyVimのデフォルトプラグインを自動的に継承
- LazyVim extraは`lua/config/lazy.lua`の`{ import = "lazyvim.plugins.extras.*" }`で有効化（`:LazyExtras`/lazyvim.json UIは未使用）。有効化中: `ai.copilot` / `ui.mini-indentscope` / lang系（`python`/`go`/`vue`/`helm`/`terraform`/`yaml`/`docker`）
- IaC・設定ファイル対応: `lang.helm`（templates/*.yaml→helm ft+gotmpl treesitter+helm_ls）・`lang.terraform`（terraform-ls+treesitter+terraform_fmt）・`lang.yaml`（yamlls+SchemaStore）・`lang.docker`。Terraform lint(terraform_validate)は`lua/plugins/terraform.lua`で無効化。k8s manifestのスキーマ補完は未対応（SchemaStoreはfileMatchベースで素のmanifestを拾わない＝要per-bufferスキーマ切替）
- カスタムプラグインは各ファイルでテーブルを返す形式で定義

## 開発コマンド

### LSP操作・ファイル検索

LSPナビゲーション（`gd`定義ジャンプ等）、診断表示、コードアクション、Snacks.nvimベースのファイル検索・エクスプローラー（`<leader>fe`/`<leader>ff`）、フォーマット実行（`<leader>cf`）を提供。キーバインド詳細は`../KEYMAPS.md`参照。

## 設定の特徴

### 日本語環境対応
- macOSでのIME自動切り替え（InsertLeave時に英数入力へ）
- UTF-8エンコーディング設定
- 日本語コメントを含むコードベース

### 診断表示
- 赤い波線（underline）を無効化
- インライン診断テキスト（virtual_text）を無効化
- サインカラムの診断アイコンは表示

### カラースキーム
- カスタムカラースキームを使用（プラグイン依存なし）

## カスタマイズ時の注意

1. **プラグイン追加**: `lua/plugins/`に新しいLuaファイルを作成
2. **既存プラグインの設定変更**: 同名のspecでオーバーライド
3. **フォーマッター追加**: `lua/plugins/conform.lua`の`formatters_by_ft`に追加
4. **キーマップ追加**: `lua/config/keymaps.lua`に記述

## トラブルシューティング

キーマップ競合確認は`:map`/`:nmap`/`:verbose map`。Neotestの「No tests found」等、テスト実行関連の詳細なトラブルシューティング手順は`../.claude/docs/nvim-troubleshooting.md`参照。
