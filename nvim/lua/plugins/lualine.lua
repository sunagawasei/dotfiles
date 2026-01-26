return {
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = {
          normal = {
            a = { fg = "#0E1210", bg = "#7E8A89", gui = "bold" }, -- gray
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          insert = {
            a = { fg = "#0E1210", bg = "#9AA6A5", gui = "bold" }, -- gray
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          visual = {
            a = { fg = "#0E1210", bg = "#B9C6C5", gui = "bold" }, -- gray
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          replace = {
            a = { fg = "#0E1210", bg = "#AAB6B5", gui = "bold" }, -- gray
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          command = {
            a = { fg = "#0E1210", bg = "#C7D2D1", gui = "bold" }, -- gray
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          terminal = {
            a = { fg = "#0E1210", bg = "#5AAFAD", gui = "bold" }, -- Cyan維持
            b = { fg = "#D7E2E1", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
          },
          inactive = {
            a = { fg = "#7E8A89", bg = "NONE" },
            b = { fg = "#7E8A89", bg = "NONE" },
            c = { fg = "#7E8A89", bg = "NONE" },
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
                return "\u{e7c5}" -- Nerd Fonts e7c5 (Vimアイコン)
              elseif str == "INSERT" then
                return "\u{ea73}" -- Nerd Fonts e7a5 (編集アイコン)
              elseif str == "VISUAL" or str == "V-LINE" or str == "V-BLOCK" then
                return "\u{f0489}" -- Nerd Fonts f0489 (選択アイコン)
              elseif str == "REPLACE" then
                return "\u{f06d4}" -- Nerd Fonts f06d4 (置換アイコン)
              elseif str == "COMMAND" then
                return "\u{f4b5}" -- Nerd Fonts e7c4 (ターミナルアイコン)
              elseif str == "TERMINAL" then
                return "\u{e795}" -- Nerd Fonts e795 (ターミナルアイコン)
              else
                return str:sub(1, 1)
              end
            end,
          },
        },
        lualine_b = {
          "branch",
          {
            "diff",
            show_ahead_behind = false,
            colored = true,
            diff_color = {
              added = { fg = "#5AAFAD" }, -- Cyan
              modified = { fg = "#8C83A3" }, -- Magenta
              removed = { fg = "#AAB6B5" }, -- グレー
            },
          },
        },
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

