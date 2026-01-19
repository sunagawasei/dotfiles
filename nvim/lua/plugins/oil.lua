return {
  "stevearc/oil.nvim",
  lazy = false,  -- 起動時にoilを開くため遅延読み込みを無効化
  dependencies = { "nvim-tree/nvim-web-devicons" },
  config = function()
    require("oil").setup({
      default_file_explorer = true,
      columns = { "icon" },
      view_options = {
        show_hidden = true,
      },
      keymaps = {
        ["<CR>"] = "actions.select",
        ["<C-v>"] = "actions.select_vsplit",
        ["<C-s>"] = "actions.select_split",
        ["<C-h>"] = "actions.parent",
        ["<C-l>"] = "actions.select",
        ["q"] = "actions.close",
        ["<C-r>"] = "actions.refresh",
      },
    })

    -- 引数なしで起動した場合のみoilを開く
    -- vim.schedule_wrap()を使用してタイミング問題を回避
    -- 参考: https://github.com/stevearc/oil.nvim/issues/268
    vim.api.nvim_create_autocmd("VimEnter", {
      callback = vim.schedule_wrap(function(data)
        -- 引数なしで起動、またはディレクトリを開いた場合のみoilを開く
        if data.file == "" or vim.fn.isdirectory(data.file) ~= 0 then
          require("oil").open()
        end
      end),
    })
  end,
  keys = {
    { "-", "<cmd>Oil<cr>", desc = "Open parent directory" },
    { "<leader>e", "<cmd>Oil<cr>", desc = "Open file explorer" },
  },
}
