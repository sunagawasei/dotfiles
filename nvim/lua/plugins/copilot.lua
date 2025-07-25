return {
  -- GitHub Copilot
  {
    "github/copilot.vim",
    event = "InsertEnter",
    config = function()
      -- Copilot設定
      vim.g.copilot_no_tab_map = true
      vim.g.copilot_assume_mapped = true
      vim.g.copilot_tab_fallback = ""
      
      -- Tab キーでの補完を有効化
      vim.api.nvim_set_keymap("i", "<Tab>", 'copilot#Accept("\\<Tab>")', {
        expr = true,
        replace_keycodes = false
      })
      vim.g.copilot_tab_fallback = ""
    end,
  },
}
