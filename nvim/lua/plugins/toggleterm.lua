return {
  "akinsho/toggleterm.nvim",
  version = "*",
  event = "VeryLazy",
  config = function()
    require("toggleterm").setup({
      size = function(term)
        if term.direction == "horizontal" then
          return 15
        elseif term.direction == "vertical" then
          return vim.o.columns * 0.4
        end
      end,
      open_mapping = nil, -- カスタムキーバインド使用
      hide_numbers = true,
      shade_terminals = true,
      shading_factor = 2,
      start_in_insert = true,
      insert_mappings = true,
      terminal_mappings = true,
      persist_size = true,
      persist_mode = true,
      direction = "horizontal",
      close_on_exit = true,
      shell = vim.o.shell,
      highlights = {
        Normal = {
          link = "Normal",
        },
        NormalFloat = {
          link = "NormalFloat",
        },
        FloatBorder = {
          link = "FloatBorder",
        },
      },
      winbar = {
        enabled = true, -- winbarを有効化してターミナル名を表示
      },
    })

    -- グローバル変数でモード管理
    _G.terminal_mode = "single" -- "single" or "side-by-side"

    -- グローバル変数で方向管理
    _G.terminal_direction = "horizontal" -- "horizontal", "vertical", or "float"

    -- モード切り替え関数
    _G.toggle_terminal_mode = function()
      if _G.terminal_mode == "single" then
        _G.terminal_mode = "side-by-side"
        vim.notify("Terminal Mode: Side-by-Side (横並び表示)", vim.log.levels.INFO)
      else
        _G.terminal_mode = "single"
        vim.notify("Terminal Mode: Single (1つだけ表示)", vim.log.levels.INFO)
      end
    end

    -- 方向切り替え関数
    _G.change_terminal_direction = function(direction)
      -- 入力検証
      local valid_directions = { horizontal = true, vertical = true, float = true }
      if not valid_directions[direction] then
        vim.notify("Invalid direction: " .. direction, vim.log.levels.ERROR)
        return
      end

      -- 既に同じ方向の場合は何もしない
      if _G.terminal_direction == direction then
        vim.notify("Already in " .. direction .. " mode", vim.log.levels.INFO)
        return
      end

      -- toggleterm APIからすべてのターミナルを取得
      local terms = require("toggleterm.terminal")
      local all_terminals = terms.get_all()

      -- 開いているターミナルのIDを記録（番号付きターミナルのみ）
      local open_terminals = {}
      for _, term in pairs(all_terminals) do
        -- LazyGit/Keifuを除外：番号付きターミナル（1/2/3）のみ対象
        if term.id >= 1 and term.id <= 3 and term:is_open() then
          table.insert(open_terminals, term.id)
          term:close()
        end
      end

      -- 方向を変更
      _G.terminal_direction = direction

      -- 開いていたターミナルを新しい方向で再度開く
      for _, term_id in ipairs(open_terminals) do
        vim.cmd(term_id .. "ToggleTerm direction=" .. direction)
      end

      -- 通知
      local direction_names = {
        horizontal = "Horizontal (下部横分割)",
        vertical = "Vertical (右側縦分割)",
        float = "Float (フローティング)",
      }
      vim.notify("Terminal Direction: " .. direction_names[direction], vim.log.levels.INFO)
    end

    -- サイクル切り替え関数
    _G.cycle_terminal_direction = function()
      local cycle = {
        horizontal = "vertical",
        vertical = "float",
        float = "horizontal",
      }
      local next_direction = cycle[_G.terminal_direction]
      _G.change_terminal_direction(next_direction)
    end

    -- スマートターミナルトグル
    _G.toggle_smart_terminal = function(count)
      local terms = require("toggleterm.terminal")
      local all_terminals = terms.get_all()

      if _G.terminal_mode == "side-by-side" then
        -- 横並びモード：そのままトグル
        vim.cmd(count .. "ToggleTerm direction=" .. _G.terminal_direction)
      else
        -- 単一モード：他を閉じる
        for _, term in pairs(all_terminals) do
          if term.id ~= count and term:is_open() then
            term:close()
          end
        end
        vim.cmd(count .. "ToggleTerm direction=" .. _G.terminal_direction)
      end
    end

    -- LazyGit統合
    local Terminal = require("toggleterm.terminal").Terminal
    local lazygit = Terminal:new({
      cmd = "lazygit",
      direction = "float",
      float_opts = {
        border = "curved",
        width = math.floor(vim.o.columns * 0.95),
        height = math.floor(vim.o.lines * 0.95),
      },
      on_open = function(term)
        vim.cmd("startinsert!")
        vim.api.nvim_buf_set_keymap(term.bufnr, "n", "q", "<cmd>close<CR>", { noremap = true, silent = true })
      end,
      on_close = function(term)
        vim.cmd("startinsert!")
      end,
    })

    _G.lazygit_toggle = function()
      lazygit:toggle()
    end

    -- Keifu統合
    local keifu = Terminal:new({
      cmd = "keifu",
      direction = "float",
      float_opts = {
        border = "curved",
        width = math.floor(vim.o.columns * 0.9),
        height = math.floor(vim.o.lines * 0.9),
      },
      on_open = function(term)
        vim.cmd("startinsert!")
      end,
    })

    _G.keifu_toggle = function()
      keifu:toggle()
    end

    -- ターミナルウィンドウのサイズを変更
    _G.resize_terminal = function(delta)
      local current_win = vim.api.nvim_get_current_win()
      local current_height = vim.api.nvim_win_get_height(current_win)
      local new_height = math.max(5, current_height + delta) -- 最小5行
      vim.api.nvim_win_set_height(current_win, new_height)
    end

    -- 特定サイズに設定
    _G.set_terminal_size = function(size)
      local current_win = vim.api.nvim_get_current_win()
      vim.api.nvim_win_set_height(current_win, size)
    end

    -- ターミナルサイクル切り替え
    _G.cycle_terminal = function()
      local current_buf = vim.api.nvim_get_current_buf()
      local current_id = vim.b[current_buf].toggle_number
      if not current_id then return end
      local next_id = (current_id % 3) + 1
      _G.toggle_smart_terminal(next_id)
    end

    -- ターミナルを最大化
    _G.maximize_terminal = function()
      local current_win = vim.api.nvim_get_current_win()
      local max_height = vim.o.lines - vim.o.cmdheight - 2 -- コマンドライン、ステータスラインを除く
      vim.api.nvim_win_set_height(current_win, max_height)
    end
  end,
  keys = {
    -- 番号付きターミナル
    {
      "<leader>t1",
      function()
        _G.toggle_smart_terminal(1)
      end,
      mode = { "n", "t" },
      desc = "Terminal 1",
    },
    {
      "<leader>t2",
      function()
        _G.toggle_smart_terminal(2)
      end,
      mode = { "n", "t" },
      desc = "Terminal 2",
    },
    {
      "<leader>t3",
      function()
        _G.toggle_smart_terminal(3)
      end,
      mode = { "n", "t" },
      desc = "Terminal 3",
    },

    -- ターミナルモード切り替え
    {
      "<leader>tm",
      function()
        _G.toggle_terminal_mode()
      end,
      mode = { "n" },
      desc = "Toggle Terminal Mode (Single/Side-by-Side)",
    },

    -- ターミナル方向切り替え
    {
      "<leader>th",
      function()
        _G.change_terminal_direction("horizontal")
      end,
      mode = { "n", "t" },
      desc = "Terminal Direction: Horizontal",
    },
    {
      "<leader>tv",
      function()
        _G.change_terminal_direction("vertical")
      end,
      mode = { "n", "t" },
      desc = "Terminal Direction: Vertical",
    },
    {
      "<leader>tf",
      function()
        _G.change_terminal_direction("float")
      end,
      mode = { "n", "t" },
      desc = "Terminal Direction: Float",
    },
    {
      "<leader>tD",
      function()
        _G.cycle_terminal_direction()
      end,
      mode = { "n", "t" },
      desc = "Cycle Terminal Direction (H→V→F)",
    },

    -- 最後のターミナルトグル
    {
      "<C-/>",
      "<cmd>ToggleTerm<CR>",
      mode = { "n", "t" },
      desc = "Toggle Terminal (Last)",
    },
    {
      "<C-_>",
      "<cmd>ToggleTerm<CR>",
      mode = { "n", "t" },
      desc = "Toggle Terminal (Last)",
    },

    -- LazyGit
    {
      "<leader>gg",
      function()
        _G.lazygit_toggle()
      end,
      mode = { "n" },
      desc = "LazyGit",
    },

    -- Keifu
    {
      "<leader>gk",
      function()
        _G.keifu_toggle()
      end,
      mode = { "n" },
      desc = "Keifu (Git Graph)",
    },

    -- ターミナルサイクル切り替え
    {
      "<C-g>",
      function() _G.cycle_terminal() end,
      mode = "t",
      desc = "Cycle to next terminal",
    },

    -- ターミナルモード操作
    {
      "<Esc><Esc>",
      [[<C-\><C-n>]],
      mode = "t",
      desc = "Exit terminal mode",
    },
    {
      "<C-q>",
      [[<C-\><C-n>]],
      mode = "t",
      desc = "Exit terminal mode (quick)",
    },
    {
      "<C-h>",
      [[<Cmd>wincmd h<CR>]],
      mode = "t",
      desc = "Go to left window",
    },
    {
      "<C-j>",
      [[<Cmd>wincmd j<CR>]],
      mode = "t",
      desc = "Go to lower window",
    },
    {
      "<C-k>",
      [[<Cmd>wincmd k<CR>]],
      mode = "t",
      desc = "Go to upper window",
    },
    {
      "<C-l>",
      [[<Cmd>wincmd l<CR>]],
      mode = "t",
      desc = "Go to right window",
    },

    -- 追加機能
    {
      "<leader>ta",
      "<cmd>ToggleTermToggleAll<CR>",
      mode = { "n" },
      desc = "Toggle All Terminals",
    },

    -- 選択範囲送信（REPL機能）
    {
      "<leader>ts",
      "<cmd>ToggleTermSendVisualSelection<CR>",
      mode = "v",
      desc = "Send selection to terminal",
    },
    {
      "<leader>tl",
      "<cmd>ToggleTermSendVisualLines<CR>",
      mode = "v",
      desc = "Send lines to terminal",
    },
    {
      "<leader>ts",
      "<cmd>ToggleTermSendCurrentLine<CR>",
      mode = "n",
      desc = "Send current line to terminal",
    },

    -- ターミナル名前付け
    {
      "<leader>tn",
      "<cmd>ToggleTermSetName<CR>",
      mode = { "n" },
      desc = "Set terminal name",
    },

    -- ターミナル選択UI
    {
      "<leader>tS",
      "<cmd>TermSelect<CR>",
      mode = { "n" },
      desc = "Select terminal",
    },

    -- ターミナルサイズ変更（Altキー）
    {
      "<M-k>",
      function()
        _G.resize_terminal(1)
      end,
      mode = { "n", "t" },
      desc = "Increase terminal height by 1",
    },
    {
      "<M-j>",
      function()
        _G.resize_terminal(-1)
      end,
      mode = { "n", "t" },
      desc = "Decrease terminal height by 1",
    },
    {
      "<M-K>",
      function()
        _G.resize_terminal(5)
      end,
      mode = { "n", "t" },
      desc = "Increase terminal height by 5",
    },
    {
      "<M-J>",
      function()
        _G.resize_terminal(-5)
      end,
      mode = { "n", "t" },
      desc = "Decrease terminal height by 5",
    },

    -- ターミナルサイズプリセット（Leaderキー）
    {
      "<leader>t+",
      function()
        _G.maximize_terminal()
      end,
      mode = { "n", "t" },
      desc = "Maximize terminal",
    },
    {
      "<leader>t-",
      function()
        _G.set_terminal_size(10)
      end,
      mode = { "n", "t" },
      desc = "Set terminal to small size (10 lines)",
    },
  },
}
