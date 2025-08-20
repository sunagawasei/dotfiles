-- blink.cmp のコマンドライン補完を有効化
return {
  {
    "saghen/blink.cmp",
    opts = {
      -- コマンドライン補完を有効化
      cmdline = {
        enabled = true,
        -- デフォルトのキーマップを使用
        -- Tab: 補完メニュー表示・最初の項目挿入
        -- Ctrl-n/p: 次/前の項目
        -- Ctrl-y: 確定, Ctrl-e: キャンセル
        completion = {
          menu = {
            -- : コマンドモードでのみ自動表示
            -- / や ? 検索モードでは Tab で手動表示
            auto_show = function(ctx)
              return vim.fn.getcmdtype() == ":"
            end,
          },
        },
      },
    },
  },
}
