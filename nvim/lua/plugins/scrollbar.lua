return {
  "petertriho/nvim-scrollbar",
  event = { "BufReadPost", "BufNewFile" },
  dependencies = {
    "lewis6991/gitsigns.nvim",
  },
  config = function()
    local colors = {
      -- Vercel Geistカラーシステムに準拠
      handle = "#333333",        -- スクロールバーのハンドル
      search = "#0070f3",        -- 検索結果（Vercel Blue）
      error = "#EF4444",         -- エラー（Red 6）
      warn = "#F59E0B",          -- 警告（Amber 6）
      info = "#3B82F6",          -- 情報（Blue 6）
      hint = "#14B8A6",          -- ヒント（Teal 6）
      misc = "#A855F7",          -- その他（Purple 6）
      -- Git関連（GitSignsと統一）
      GitAdd = "#4ADE80",       -- 追加（Green 5）
      GitChange = "#60A5FA",    -- 変更（Blue 5）
      GitDelete = "#F87171",    -- 削除（Red 5）
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
