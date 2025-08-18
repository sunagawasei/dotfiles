return {
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = {
          normal = {
            a = { fg = "#FFFFFF", bg = "#3B82F6", gui = "bold" }, -- Blue 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          insert = {
            a = { fg = "#FFFFFF", bg = "#22C55E", gui = "bold" }, -- Green 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          visual = {
            a = { fg = "#FFFFFF", bg = "#A855F7", gui = "bold" }, -- Purple 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          replace = {
            a = { fg = "#FFFFFF", bg = "#EF4444", gui = "bold" }, -- Red 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          command = {
            a = { fg = "#FFFFFF", bg = "#F59E0B", gui = "bold" }, -- Amber 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          terminal = {
            a = { fg = "#FFFFFF", bg = "#14B8A6", gui = "bold" }, -- Teal 6
            b = { fg = "#FFFFFF", bg = "NONE" },
            c = { fg = "#A3A3A3", bg = "NONE" },
          },
          inactive = {
            a = { fg = "#666666", bg = "NONE" },
            b = { fg = "#666666", bg = "NONE" },
            c = { fg = "#666666", bg = "NONE" },
          },
        },
        component_separators = { left = "", right = "" },
        section_separators = { left = "", right = "" },
      },
      sections = {
        lualine_a = {
          {
            "mode",
            fmt = function(str)
              if str == "NORMAL" then
                return "N"
              end
              return str:sub(1, 1)
            end,
          },
        },
        lualine_b = { "branch", { "diff", show_ahead_behind = false } },
        lualine_c = {
          "diagnostics",
          {
            "filename",
            path = 1, -- フルパス表示
          },
        },
        lualine_x = { "filetype" },
        lualine_y = { "location" },
        lualine_z = {},
      },
      inactive_sections = {
        lualine_a = {},
        lualine_b = {},
        lualine_c = {
          {
            "filename",
            path = 1, -- フルパス表示
          },
        },
        lualine_x = {},
        lualine_y = {},
        lualine_z = {},
      },
    },
  },
}