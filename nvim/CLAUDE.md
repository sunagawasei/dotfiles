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

キーマップ競合確認は`:map`/`:nmap`/`:verbose map`。

同じキーを複数のプラグインspecが`keys`で宣言すると、lazy.nvimは最初に登録されたspecのrhsだけを取り込んだloaderを1個作り、残りは`active[id]`に束ねる。どちらが先かは`pairs()`順に依存して起動ごとに変わるため、**衝突を放置すると起動ごとに動作が変わる**。実マッピングのdescがspec側のdescと違っていたらこれを疑う。
