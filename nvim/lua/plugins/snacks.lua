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
    indent = {
      enabled = false, -- インデントガイドを無効化（全階層表示を防ぐ）
    },
  },
}

