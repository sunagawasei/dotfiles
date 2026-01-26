return {
  "petertriho/nvim-scrollbar",
  event = { "BufReadPost", "BufNewFile" },
  dependencies = {
    "lewis6991/gitsigns.nvim",
  },
  config = function()
    local colors = {
      -- カスタムカラーシステムに準拠
      handle = "#2A2F2E",        -- スクロールバーのハンドル（COLOR-SYSTEM.md準拠: 濃い影）
      search = "#5AAFAD",        -- 検索結果（Cyan）
      error = "#8C83A3",         -- エラー（Magenta）
      warn = "#B3A9D1",          -- 警告（Bright Magenta）
      info = "#96CBD1",          -- 情報（Bright Cyan）
      hint = "#5AAFAD",          -- ヒント（Cyan）
      misc = "#7E8A89",          -- その他（Gray）
      -- Git関連（GitSignsと統一）
      GitAdd = "#5AAFAD",       -- 追加（Cyan）
      GitChange = "#8C83A3",    -- 変更（Magenta）
      GitDelete = "#AAB6B5",    -- 削除（グレー）
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
        cursor = true,     -- カーソル位置表示
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
