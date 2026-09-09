---
paths:
  - "**/nvim/**/*.lua"
  - "**/nvim/init.lua"
---

# Neovim設定規約

## アーキテクチャ

このNeovim設定はLazyVimベースのモジュラー構造です。

詳細なドキュメントは`nvim/CLAUDE.md`を参照してください。

## プラグイン追加規約

### 新しいプラグインの追加

`lua/plugins/`ディレクトリに個別のLuaファイルを作成：

```lua
-- lua/plugins/plugin-name.lua
return {
  "author/plugin-name",
  event = "VeryLazy", -- 遅延読み込み推奨
  opts = {
    -- プラグイン設定
  },
}
```

### 既存プラグインの設定変更

同名のspecでオーバーライド：

```lua
-- lua/plugins/existing-plugin.lua
return {
  "existing/plugin",
  opts = {
    -- カスタム設定
  },
}
```

## フォーマッター追加

`lua/plugins/conform.lua`の`formatters_by_ft`に追加：

```lua
formatters_by_ft = {
  go = { "gofmt", "goimports" },
  lua = { "stylua" },
  -- 新しいファイルタイプを追加
}
```

## キーマップ追加

`lua/config/keymaps.lua`に記述する：

```lua
vim.keymap.set("n", "<leader>xx", function()
  -- 機能の実装
end, { desc = "機能の説明" })
```

## モジュラー構造要件

- 1プラグイン = 1ファイル（明確さのため）
- 複雑な設定はサブディレクトリにファイル分割可能
- 設定の再利用性を重視
- LazyVimのデフォルトを尊重（オーバーライドは最小限に）

## キーマップ追加時の衝突確認

追加前に既存の割り当てをgrepで確認する。特にプラグインspecの`keys`で宣言する場合、
LazyVimや他プラグインが同じキーを宣言していると起動ごとに勝者が変わる（詳細は`nvim/CLAUDE.md`）。

## トラブルシューティング

問題が発生した場合：

```vim
:checkhealth          # 全体のヘルスチェック
:Lazy log             # プラグイン更新ログ
:messages             # エラーメッセージ履歴
```

詳細なトラブルシューティングは`nvim/CLAUDE.md`を参照してください。
