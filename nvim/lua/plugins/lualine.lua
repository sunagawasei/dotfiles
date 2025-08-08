return {
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = {
          normal = {
            a = { fg = "#FFFFFF", bg = "#3B82F6", gui = "bold" }, -- Blue 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          insert = {
            a = { fg = "#FFFFFF", bg = "#22C55E", gui = "bold" }, -- Green 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          visual = {
            a = { fg = "#FFFFFF", bg = "#A855F7", gui = "bold" }, -- Purple 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          replace = {
            a = { fg = "#FFFFFF", bg = "#EF4444", gui = "bold" }, -- Red 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          command = {
            a = { fg = "#FFFFFF", bg = "#F59E0B", gui = "bold" }, -- Amber 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          terminal = {
            a = { fg = "#FFFFFF", bg = "#14B8A6", gui = "bold" }, -- Teal 6
            b = { fg = "#FFFFFF", bg = "#1A1A1A" },
            c = { fg = "#A3A3A3", bg = "#111111" },
          },
          inactive = {
            a = { fg = "#666666", bg = "#1A1A1A" },
            b = { fg = "#666666", bg = "#111111" },
            c = { fg = "#666666", bg = "#111111" },
          },
        },
      },
    },
  },
}