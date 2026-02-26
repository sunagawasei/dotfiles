return {
  "petertriho/nvim-scrollbar",
  event = { "BufReadPost", "BufNewFile" },
  dependencies = {
    "lewis6991/gitsigns.nvim",
  },
  config = function()
    local p = require("config.palette").colors
    local colors = {
      handle = p.dark_shadow,   -- core.ui_shadow
      search = p.operator,      -- semantic.operator
      error = p.magenta,        -- semantic.error
      warn = p.lavender,        -- semantic.warning
      info = p.cyan,            -- semantic.info
      hint = p.operator,        -- semantic.operator
      misc = p.mid_gray,        -- semantic.comment
      GitAdd = p.operator,      -- semantic.operator
      GitChange = p.magenta,    -- semantic.error
      GitDelete = p.light_gray, -- foregrounds.dim
    }

    require("scrollbar").setup({
      show = true,
      show_in_active_only = false,
      handle = {
        text = " ",
        blend = 0,  -- 透明度（0 = 不透明）
        color = colors.handle,
      },
      marks = {
        -- 診断マーク
        Error = { color = colors.error, text = { "-", "=" } },
        Warn = { color = colors.warn, text = { "-", "=" } },
        Info = { color = colors.info, text = { "-", "=" } },
        Hint = { color = colors.hint, text = { "-", "=" } },
        Misc = { color = colors.misc, text = { "-", "=" } },
        -- 検索マーク
        Search = { color = colors.search, text = { "-", "=" } },
        -- Git統合
        GitAdd = { color = colors.GitAdd, text = "┆" },
        GitChange = { color = colors.GitChange, text = "┆" },
        GitDelete = { color = colors.GitDelete, text = "▁" },
      },
      excluded_buftypes = {
        "terminal",
        "nofile",
        "quickfix",
        "prompt",
      },
      excluded_filetypes = {
        "NvimTree",
        "neo-tree",
        "dashboard",
        "lazy",
        "mason",
        "TelescopePrompt",
        "SnacksPickerPrompt",
        "SnacksExplorer",
      },
      handlers = {
        cursor = false,    -- カーソル位置表示（パフォーマンス最適化のため無効化）
        diagnostic = true, -- 診断情報表示
        handle = true,     -- スクロールハンドル表示
        search = false,    -- 検索結果表示（hlslens未使用のためメッセージ抑制）
        gitsigns = true,   -- Git情報表示（要gitsigns.nvim）
      },
    })

    -- Gitsigns統合はgitsigns.luaで管理
    -- scrollbarのhandler登録のみここで実行
    require("scrollbar.handlers.gitsigns").setup()
  end,
}
