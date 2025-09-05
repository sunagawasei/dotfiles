return {
  "lewis6991/gitsigns.nvim",
  event = { "BufReadPre", "BufNewFile" },
  opts = {
    -- デフォルトでサインカラムを非表示
    signcolumn = false,  -- デフォルトOFF
    numhl = false,       -- 行番号ハイライトもOFF
    linehl = false,      -- 行ハイライトもOFF
    word_diff = false,   -- 単語差分もOFF
    current_line_blame = false, -- Blame表示もOFF
    
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
    -- メイントグル: サインカラムの表示/非表示
    { "<leader>ug", ":Gitsigns toggle_signs<CR>", desc = "Toggle Git signs" },
    
    -- 追加のトグル機能
    { "<leader>ub", ":Gitsigns toggle_current_line_blame<CR>", desc = "Toggle Git blame line" },
    { "<leader>un", ":Gitsigns toggle_numhl<CR>", desc = "Toggle Git number highlight" },
    { "<leader>ul", ":Gitsigns toggle_linehl<CR>", desc = "Toggle Git line highlight" },
    { "<leader>uw", ":Gitsigns toggle_word_diff<CR>", desc = "Toggle Git word diff" },
    { "<leader>ud", ":Gitsigns toggle_deleted<CR>", desc = "Toggle Git deleted" },
    
    -- プレビュー機能（トグルではない）
    { "<leader>hp", ":Gitsigns preview_hunk<CR>", desc = "Preview Git hunk" },
    { "<leader>hb", ":Gitsigns blame_line<CR>", desc = "Git blame line" },
  },
}
