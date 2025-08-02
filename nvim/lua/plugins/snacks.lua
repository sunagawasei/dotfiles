return {
  "folke/snacks.nvim",
  opts = {
    picker = {
      sources = {
        explorer = {
          hidden = true,
          ignored = true,
          exclude = { ".git", ".DS_Store" },
        },
        files = {
          hidden = true,
          ignored = true,
        },
      },
    },
    bigfile = {
      enabled = false, -- bigfile機能を完全に無効化
    },
  },
}

