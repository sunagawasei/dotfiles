return {
  {
    "tadmccorkle/markdown.nvim",
    ft = "markdown",
    keys = {
      { "<leader>mt", "<cmd>MDTaskToggle<cr>", desc = "Toggle Checkbox" },
    },
    opts = {
      on_attach = function(bufnr)
        local map = vim.keymap.set
        local opts = { buffer = bufnr }
        -- Insert mode: Enterでリスト自動継続（リスト外では通常のEnter）
        map("i", "<CR>", "<cmd>MDListItemBelow<cr>", opts)
        -- Normal mode: o/Oでリスト自動継続
        map("n", "o", "<cmd>MDListItemBelow<cr>", opts)
        map("n", "O", "<cmd>MDListItemAbove<cr>", opts)
      end,
    },
  },
}
