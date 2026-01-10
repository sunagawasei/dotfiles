return {
  "nvim-neo-tree/neo-tree.nvim",
  enabled = false,  -- oil.nvimに移行したため無効化
  opts = {
    -- 通知メッセージを非表示
    enable_diagnostics = false,
    enable_git_status = true,
    window = {
      position = "right", -- 右側から開く
      width = 40,
      mappings = {
        -- LazyVimデフォルトマッピングを復元
        ["l"] = "open",
        ["h"] = "close_node",
        ["<space>"] = "none",
        ["Y"] = {
          function(state)
            local node = state.tree:get_node()
            local path = node:get_id()
            vim.fn.setreg("+", path, "c")
          end,
          desc = "Copy Path to Clipboard",
        },
        ["O"] = {
          function(state)
            require("lazy.util").open(state.tree:get_node().path, { system = true })
          end,
          desc = "Open with System Application",
        },
        ["P"] = { "toggle_preview", config = { use_float = false } },
      },
    },
    filesystem = {
      follow_current_file = {
        enabled = true, -- 現在のファイルを自動的にフォーカス
      },
      use_libuv_file_watcher = true, -- ファイル変更の自動監視
      filtered_items = {
        visible = true, -- 隠しファイルを表示
        hide_dotfiles = false, -- .で始まるファイルを表示
        hide_gitignored = false, -- .gitignoreされたファイルも表示
      },
    },
    -- イベント通知を抑制
    event_handlers = {
      {
        event = "neo_tree_popup_input_ready",
        handler = function()
          vim.cmd("stopinsert")
        end,
      },
    },
  },
  keys = {
    -- LazyVimデフォルトの<leader>feを<leader>eにリマップ
    {
      "<leader>fe",
      false, -- LazyVimデフォルトを無効化
    },
    {
      "<leader>fE",
      false, -- LazyVimデフォルトを無効化
    },
    {
      "<leader>e",
      function()
        local neotree = require("neo-tree.command")

        -- 現在のバッファがneo-treeかどうか判定
        if vim.bo.filetype == "neo-tree" then
          -- neo-treeにフォーカスがある場合: 閉じる
          neotree.execute({ toggle = true, dir = vim.uv.cwd() })
        else
          -- neo-tree以外のバッファにフォーカスがある場合
          -- neo-treeウィンドウが存在するか確認
          local neo_tree_visible = false
          for _, win in ipairs(vim.api.nvim_list_wins()) do
            local buf = vim.api.nvim_win_get_buf(win)
            if vim.bo[buf].filetype == "neo-tree" then
              neo_tree_visible = true
              break
            end
          end

          if neo_tree_visible then
            -- neo-treeが表示されている場合: フォーカスを移す
            neotree.execute({ action = "focus", dir = vim.uv.cwd() })
          else
            -- neo-treeが非表示の場合: 開く
            neotree.execute({ toggle = true, dir = vim.uv.cwd() })
          end
        end
      end,
      desc = "Explorer Neo-tree (toggle/focus)",
    },
    {
      "<leader>E",
      function()
        require("neo-tree.command").execute({ toggle = true, dir = vim.uv.cwd() })
      end,
      desc = "Explorer Neo-tree (cwd)",
    },
    -- Git/Buffer explorerは既存のLazyVimデフォルトを維持
  },
}
