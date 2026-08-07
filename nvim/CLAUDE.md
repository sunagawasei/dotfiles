# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

LazyVimベースのNeovim設定。個人用にカスタマイズされており、日本語コメントを含む。

## 標準と違う慣習

LazyVim extraは`lua/config/lazy.lua`の`{ import = "lazyvim.plugins.extras.*" }`で有効化する。`:LazyExtras`/lazyvim.json UIは未使用のため、UIから足しても反映されない。

## 踏んだ罠

- Terraform lint(terraform_validate)は`lua/plugins/terraform.lua`で無効化してある
- k8s manifestのスキーマ補完は未対応。SchemaStoreはfileMatchベースで素のmanifestを拾わないため、per-bufferでのスキーマ切替が要る

## トラブルシューティング

キーマップ競合確認は`:map`/`:nmap`/`:verbose map`。Neotestの「No tests found」等、テスト実行関連の詳細なトラブルシューティング手順は`../.claude/docs/nvim-troubleshooting.md`参照。
