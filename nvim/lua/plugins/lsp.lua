return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      -- LazyVimの診断設定
      diagnostics = {
        underline = false, -- 赤い波線を完全に無効化
        virtual_text = false, -- インライン診断テキストも無効化
        signs = true, -- 左側のサインカラムは表示
        float = {
          border = "rounded",
          source = "always",
        },
        severity_sort = true,
        update_in_insert = false,
      },
    },
  },
}