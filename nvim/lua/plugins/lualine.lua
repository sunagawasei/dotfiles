-- Color Mapping: See colors/ghost-visor.toml [semantic] section
local p = require("config.palette").colors

local mode_colors = {
  normal = p.subdued_fg,      -- foregrounds.subdued (mode badge background)
  insert = p.success,         -- semantic.success (green for creative action)
  visual = p.operator,        -- semantic.operator (teal bright for selection)
  replace = p.lavender,       -- semantic.warning (lavender for caution)
  command = p.purple_accent,  -- semantic.keyword (blue for commands)
  terminal = p.string,        -- semantic.string (teal mid for shell)
  inactive = p.light_gray,    -- foregrounds.dim (透過バー上で AA を満たす最も控えめな色)
}

return {
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = {
          normal = {
            a = { fg = p.darkest_bg, bg = mode_colors.normal, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          insert = {
            a = { fg = p.darkest_bg, bg = mode_colors.insert, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          visual = {
            a = { fg = p.darkest_bg, bg = mode_colors.visual, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          replace = {
            a = { fg = p.darkest_bg, bg = mode_colors.replace, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          command = {
            a = { fg = p.darkest_bg, bg = mode_colors.command, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          terminal = {
            a = { fg = p.darkest_bg, bg = mode_colors.terminal, gui = "bold" },
            b = { fg = p.fg, bg = "none" },
            c = { fg = p.light_gray, bg = "none" },
          },
          inactive = {
            a = { fg = mode_colors.inactive, bg = "none" },
            b = { fg = mode_colors.inactive, bg = "none" },
            c = { fg = mode_colors.inactive, bg = "none" },
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
              added = { fg = p.success },
              modified = { fg = p.lavender },
              removed = { fg = p.magenta },
            },
          },
        },
        lualine_c = {
          "diagnostics",
          {
            "filename",
            path = 3,            -- 短縮相対パス
            shorting_target = 40,
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
            path = 3,            -- 短縮相対パス
            shorting_target = 40,
          },
        },
        lualine_x = {},
        lualine_y = {},
        lualine_z = {},
      },
    },
  },
}
