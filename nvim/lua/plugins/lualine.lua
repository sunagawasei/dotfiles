-- Color Mapping: See colors/abyssal-teal.toml [semantic] section
local mode_colors = {
  normal = "#586269",   -- semantic.comment (subdued for idle state)
  insert = "#4A8778",   -- semantic.success (green for creative action)
  visual = "#5ABDBC",   -- semantic.function (teal bright for selection)
  replace = "#CED5E9",  -- semantic.warning (lavender for caution)
  command = "#54688D",  -- semantic.keyword (blue for commands)
  terminal = "#659D9E", -- semantic.string (teal mid for shell)
  inactive = "#586269", -- semantic.comment (dim for background)
}

return {
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = {
          normal = {
            a = { fg = "#0E1210", bg = mode_colors.normal, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          insert = {
            a = { fg = "#0E1210", bg = mode_colors.insert, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          visual = {
            a = { fg = "#0E1210", bg = mode_colors.visual, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          replace = {
            a = { fg = "#0E1210", bg = mode_colors.replace, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          command = {
            a = { fg = "#0E1210", bg = mode_colors.command, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          terminal = {
            a = { fg = "#0E1210", bg = mode_colors.terminal, gui = "bold" },
            b = { fg = "#CEF5F2", bg = "NONE" },
            c = { fg = "#586269", bg = "NONE" },
          },
          inactive = {
            a = { fg = mode_colors.inactive, bg = "NONE" },
            b = { fg = mode_colors.inactive, bg = "NONE" },
            c = { fg = mode_colors.inactive, bg = "NONE" },
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
              added = { fg = "#5ABDBC" },   -- semantic.function (teal bright)
              modified = { fg = "#CED5E9" }, -- semantic.warning (lavender)
              removed = { fg = "#926894" },  -- semantic.error (purple muted)
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

