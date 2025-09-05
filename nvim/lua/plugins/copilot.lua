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
        replace_keycodes = false,
      })
      vim.g.copilot_tab_fallback = ""
      
      -- Alt+L で次の単語を受け入れる（右移動の L、Karabiner競合回避）
      vim.api.nvim_set_keymap("i", "<M-l>", "<Plug>(copilot-accept-word)", {
        silent = true,
      })
      
      -- Ctrl+K で次の行を受け入れる
      vim.api.nvim_set_keymap("i", "<C-K>", "<Plug>(copilot-accept-line)", {
        silent = true,
      })
      
      -- Ctrl+] で補完を拒否/キャンセル（Copilotデフォルト）
      vim.api.nvim_set_keymap("i", "<C-]>", "<Plug>(copilot-dismiss)", {
        silent = true,
      })
      
      -- Alt+] で次の提案を表示
      vim.api.nvim_set_keymap("i", "<M-]>", "<Plug>(copilot-next)", {
        silent = true,
      })
      
      -- Alt+[ で前の提案を表示
      vim.api.nvim_set_keymap("i", "<M-[>", "<Plug>(copilot-previous)", {
        silent = true,
      })
    end,
  },
}
