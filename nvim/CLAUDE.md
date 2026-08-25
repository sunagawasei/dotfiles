# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## リポジトリ概要

LazyVimベースのNeovim設定。個人用にカスタマイズされており、日本語コメントを含む。

## 標準と違う慣習

LazyVim extraは`lua/config/lazy.lua`の`{ import = "lazyvim.plugins.extras.*" }`で有効化する。`:LazyExtras`/lazyvim.json UIは未使用のため、UIから足しても反映されない。

## 踏んだ罠

- Terraform lint(terraform_validate)は`lua/plugins/terraform.lua`で無効化してある
- k8s manifestのスキーマ補完は未対応。SchemaStoreはfileMatchベースで素のmanifestを拾わないため、per-bufferでのスキーマ切替が要る
- **spec の `keys` に `{ lhs, "" }` (rhs空文字) を書くとプラグインがロードできなくなる**。lazy.nvimはこれを`M.is_nop`と判定してloader callbackを張らないため、そのキーはただのnopマッピングになる。グループラベル目的で書くのは構わないが、それが唯一のトリガーだと`:Lazy load`以外に起動経路が無くなる
- **`{ lhs, false }` による無効化はファイル名順に負ける**。lazy.nvimは`plugins/`をモジュール名順に読み、`false`は以降の再追加を禁止しない。無効化を集約するファイルは`zz-`始まりにして最後に評価させる（`lua/plugins/zz-disabled-keys.lua`）
- **`{ lhs, false }` のmode省略はnモードだけを消す**。lazy側が`mode = {"n","x"}`等で宣言しているキーは、同じmodeを明示しないと片方が残る
- LazyVimの`config/keymaps.lua`が直接張るキー（`[b`/`]b`等）はspecの`false`では消せない。`vim.keymap.del`が要る
- **Nixの`wrapNeovim`は起動時に`--cmd "lua vim.g.loaded_..._provider=0;..."`を`$@`の前に固定注入する**。`vim.v.argv`をargv[2]から素朴に走査してフラグ判定するロジックは、この`--cmd`を常にユーザーのフラグと誤検知する。`--cmd`とその直後の引数1個はペアでスキップする実装が要る（`nvim/lua/plugins/flatten.lua`の`should_nest`参照）
- **flatten.nvim(nvim-remote系プラグイン全般)のhookは種別ごとに実行プロセスが違う**。`should_nest`はguest(新規起動したnvim)側、`pre_open`/`post_open`はhost(親neovim)側でRPC経由で実行される。別プロセスのLua VMなので`_G`グローバル変数でこれらの間の状態共有はできない（常にnilになる）。ウィンドウ特定等の状態は、host側フック同士(`pre_open`↔`post_open`)でのみ`_G`共有する

## トラブルシューティング

キーマップ競合確認は`:map`/`:nmap`/`:verbose map`。

同じキーを複数のプラグインspecが`keys`で宣言すると、lazy.nvimは最初に登録されたspecのrhsだけを取り込んだloaderを1個作り、残りは`active[id]`に束ねる。どちらが先かは`pairs()`順に依存して起動ごとに変わるため、**衝突を放置すると起動ごとに動作が変わる**。実マッピングのdescがspec側のdescと違っていたらこれを疑う。
