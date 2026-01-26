return {
  "stevearc/oil.nvim",
  lazy = false,  -- 起動時にoilを開くため遅延読み込みを無効化
  dependencies = { "nvim-tree/nvim-web-devicons" },
  config = function()
    local detail = true  -- 詳細表示状態の管理
    require("oil").setup({
      default_file_explorer = true,
      delete_to_trash = true,              -- ファイル削除時にゴミ箱へ移動
      skip_confirm_for_simple_edits = true, -- 単一操作の確認をスキップ
      watch_for_changes = true,            -- 外部変更を自動検出
      columns = { "icon", "permissions", "size", "mtime" },
      view_options = {
        show_hidden = true,
      },
      keymaps = {
        ["<CR>"] = "actions.select",
        ["<C-v>"] = "actions.select_vsplit",
        ["<C-x>"] = "actions.select_split",
        ["<C-h>"] = "actions.parent",
        ["<C-l>"] = "actions.select",
        ["q"] = "actions.close",
        ["<C-r>"] = "actions.refresh",
        ["gd"] = {
          desc = "Toggle file detail view",
          callback = function()
            detail = not detail
            if detail then
              require("oil").set_columns({ "icon", "permissions", "size", "mtime" })
            else
              require("oil").set_columns({ "icon" })
            end
          end,
        },
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
