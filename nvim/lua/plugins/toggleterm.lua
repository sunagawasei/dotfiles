return {
  "akinsho/toggleterm.nvim",
  version = "*",
  event = "VeryLazy",
  config = function()
    require("toggleterm").setup({
      -- サイズ設定
      size = function(term)
        if term.direction == "horizontal" then
          return 15
        elseif term.direction == "vertical" then
          return vim.o.columns * 0.4
        end
      end,
      
      -- 基本設定
      open_mapping = [[<c-\>]], -- デフォルトのトグルキー
      hide_numbers = true, -- ターミナルでの行番号を非表示
      shade_filetypes = {},
      shade_terminals = true,
      shading_factor = 2, -- ターミナルの背景を暗くする
      start_in_insert = true,
      insert_mappings = true, -- インサートモードでもマッピングを有効化
      terminal_mappings = true, -- ターミナルモードでもマッピングを有効化
      persist_size = true,
      persist_mode = true, -- ターミナルモードを記憶
      direction = "float", -- 'vertical' | 'horizontal' | 'tab' | 'float'
      close_on_exit = true, -- 終了時に自動的に閉じる
      shell = vim.o.shell, -- デフォルトシェルを使用
      
      -- フロートウィンドウの設定
      float_opts = {
        border = "curved", -- 'single' | 'double' | 'shadow' | 'curved'
        width = function()
          return math.floor(vim.o.columns * 0.8)
        end,
        height = function()
          return math.floor(vim.o.lines * 0.8)
        end,
        winblend = 0,
        highlights = {
          border = "Normal",
          background = "Normal",
        },
      },
      
      -- ウィンドウバーの設定
      winbar = {
        enabled = false,
      },
    })
    
    -- カスタムターミナル関数
    local Terminal = require("toggleterm.terminal").Terminal
    
    -- lazygit用のターミナル
    local lazygit = Terminal:new({
      cmd = "lazygit",
      direction = "float",
      float_opts = {
        border = "double",
      },
      on_open = function(term)
        vim.cmd("startinsert!")
        vim.api.nvim_buf_set_keymap(term.bufnr, "n", "q", "<cmd>close<CR>", {noremap = true, silent = true})
      end,
    })
    
    -- gitui用のターミナル
    local gitui = Terminal:new({
      cmd = "gitui",
      direction = "float",
      float_opts = {
        border = "double",
      },
    })
    
    -- btop用のターミナル
    local btop = Terminal:new({
      cmd = "btop",
      direction = "float",
      float_opts = {
        border = "double",
      },
    })
    
    -- カスタムコマンドの作成
    vim.api.nvim_create_user_command("LazyGit", function()
      lazygit:toggle()
    end, {})
    
    vim.api.nvim_create_user_command("GitUI", function()
      gitui:toggle()
    end, {})
    
    vim.api.nvim_create_user_command("Btop", function()
      btop:toggle()
    end, {})
    
    -- グローバル関数として登録（キーマッピング用）
    _G.toggleterm_lazygit = function()
      lazygit:toggle()
    end
    
    _G.toggleterm_gitui = function()
      gitui:toggle()
    end
    
    _G.toggleterm_btop = function()
      btop:toggle()
    end
  end,
  
  -- キーマッピング
  keys = {
    -- 基本的なトグル
    { "<c-\\>", "<cmd>ToggleTerm<cr>", desc = "Toggle terminal" },
    { "<leader>tf", "<cmd>ToggleTerm direction=float<cr>", desc = "Terminal float" },
    { "<leader>th", "<cmd>ToggleTerm direction=horizontal<cr>", desc = "Terminal horizontal" },
    { "<leader>tv", "<cmd>ToggleTerm direction=vertical<cr>", desc = "Terminal vertical" },
    
    -- カスタムターミナル
    { "<leader>tg", function() _G.toggleterm_lazygit() end, desc = "LazyGit" },
    { "<leader>tu", function() _G.toggleterm_gitui() end, desc = "GitUI" },
    { "<leader>tb", function() _G.toggleterm_btop() end, desc = "Btop" },
    
    -- ターミナルモードでのキーマッピング
    { "<esc><esc>", [[<C-\><C-n>]], mode = "t", desc = "Exit terminal mode" },
    { "<C-h>", [[<Cmd>wincmd h<CR>]], mode = "t", desc = "Go to left window" },
    { "<C-j>", [[<Cmd>wincmd j<CR>]], mode = "t", desc = "Go to lower window" },
    { "<C-k>", [[<Cmd>wincmd k<CR>]], mode = "t", desc = "Go to upper window" },
    { "<C-l>", [[<Cmd>wincmd l<CR>]], mode = "t", desc = "Go to right window" },
  },
}