return {
  "lewis6991/gitsigns.nvim",
  event = { "BufReadPre", "BufNewFile" },
  opts = {
    -- Git差分とBlame表示をデフォルトで有効化
    signcolumn = true,   -- Git差分記号を表示
    numhl = false,       -- 行番号ハイライトはOFF
    linehl = false,      -- 行ハイライトはOFF
    word_diff = false,   -- 単語差分はOFF
    current_line_blame = true, -- Blame表示をON

    -- Blame表示のオプション設定
    current_line_blame_opts = {
      virt_text = true,              -- 仮想テキストで表示
      virt_text_pos = 'eol',         -- 行末に表示
      delay = 500,                   -- カーソル移動後500msで表示(ちらつき防止)
      ignore_whitespace = true,      -- 空白のみの変更は無視
      virt_text_priority = 100,      -- 表示優先度
      use_focus = true,              -- アクティブウィンドウでのみ表示
    },

    -- Blame表示のフォーマット
    current_line_blame_formatter = '<author>, <author_time:%Y-%m-%d> - <summary>',

    -- その他の基本設定
    attach_to_untracked = true,
    watch_gitdir = { interval = 1000, follow_files = true },
    sign_priority = 6,
    update_debounce = 100,
    status_formatter = nil,
    max_file_length = 40000,
    
    -- scrollbar連携用のコールバック
    _on_attach_pre = function(_, callback)
      require("scrollbar.handlers.gitsigns").setup()
      callback()
    end,
  },
  keys = {
    -- HunkプレビューとBlame詳細表示
    { "<leader>hp", ":Gitsigns preview_hunk<CR>", desc = "Preview Git hunk" },
    { "<leader>hb", ":Gitsigns blame_line<CR>", desc = "Git blame line" },
  },
}
