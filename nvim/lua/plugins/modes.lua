return {
  {
    "mvllow/modes.nvim",
    version = "v0.2.*",
    event = { "BufReadPost", "BufNewFile" },
    opts = {
      colors = {
        -- Vercel Geist カラーパレット準拠
        copy = "#FBBF24", -- Amber 5: コピー操作
        delete = "#EF4444", -- Red 6: 削除操作
        insert = "#22C55E", -- Green 6: 挿入モード
        visual = "#0891B2", -- Cyan 6: ビジュアルモード
        replace = "#A855F7", -- Purple 6: リプレイスモード
      },
      line_opacity = 0.35,
      set_cursor = true,
      set_cursorline = true,
      set_number = true,
      ignore_filetypes = {
        "NvimTree",
        "neo-tree",
        "TelescopePrompt",
        "snacks_picker",
        "toggleterm",
        "lazy",
        "mason",
      },
    },
  },
}
