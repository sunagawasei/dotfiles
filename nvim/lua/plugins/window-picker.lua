return {
  "s1n7ax/nvim-window-picker",
  version = "2.*",
  keys = {
    {
      "<leader>wp",
      function()
        local picked_window_id = require("window-picker").pick_window()
        if picked_window_id then
          vim.api.nvim_set_current_win(picked_window_id)
        end
      end,
      desc = "Pick window",
    },
  },
  opts = {
    selection_chars = "123456789",
    hint = "floating-big-letter",
    show_prompt = false,
    filter_rules = {
      autoselect_one = true,
      include_current = false,
      bo = {
        filetype = { "neo-tree", "neo-tree-popup", "notify" },
        buftype = { "terminal", "quickfix" },
      },
    },
  },
}