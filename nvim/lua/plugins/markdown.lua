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
        -- Normal mode: o/Oでリスト自動継続、リスト外では通常動作にフォールバック
        map("n", "o", function()
          if not require("markdown.list").insert_list_item_below() then
            vim.cmd("normal! o")
          end
        end, opts)
        map("n", "O", function()
          if not require("markdown.list").insert_list_item_above() then
            vim.cmd("normal! O")
          end
        end, opts)
      end,
    },
  },
}
