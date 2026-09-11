return {
  "akinsho/toggleterm.nvim",
  version = "*",
  event = "VeryLazy",
  config = function()
    -- 番号付きターミナルの上限(lazygit=99 / hunk=98 / hunk_staged=96 とは別枠)
    local MAX_NUMBERED_TERMINALS = 9

    require("toggleterm").setup({
      size = function(term)
        if term.direction == "horizontal" then
          return vim.o.lines - vim.o.cmdheight - 2
        elseif term.direction == "vertical" then
          return vim.o.columns
        end
      end,
      float_opts = {
        border = "curved",
        width = function() return math.floor(vim.o.columns * 0.95) end,
        height = function() return math.floor(vim.o.lines * 0.95) end,
      },
      open_mapping = nil, -- カスタムキーバインド使用
      hide_numbers = true,
      shade_terminals = true,
      shading_factor = 2,
      start_in_insert = true,
      insert_mappings = true,
      terminal_mappings = true,
      persist_size = true,
      persist_mode = false,
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
        -- LazyGit/Hunkを除外：番号付きターミナルのみ対象
        if term.id >= 1 and term.id <= MAX_NUMBERED_TERMINALS and term:is_open() then
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
      hidden = true,
      count = 99,
      env = {
        NVIM = vim.v.servername, -- $NVIMを明示的に渡してneovim-remoteで開けるようにする
        LG_CONFIG_FILE = table.concat({
          vim.fn.expand("~/.config/lazygit/config.yml"),
          vim.fn.expand("~/.config/lazygit/nvim-hide.yml"),
        }, ","),
      },
      float_opts = {
        border = "curved",
        width = function() return math.floor(vim.o.columns * 0.95) end,
        height = function() return math.floor(vim.o.lines * 0.95) end,
      },
      on_open = function(term)
        vim.cmd("startinsert!")
        vim.api.nvim_buf_set_keymap(term.bufnr, "n", "q", "<cmd>close<CR>", { noremap = true, silent = true })
        -- Ctrl+GをLazyGitにパススルー（グローバルのtモードマッピングを上書き）
        vim.api.nvim_buf_set_keymap(term.bufnr, "t", "<C-g>", "<C-g>", { noremap = true, silent = true })
      end,
      on_close = function(term)
        -- 通常バッファでstartinsert!するとwhich-keyのModeChangedサイクルが増える
        vim.schedule(function()
          if vim.bo.buftype == "terminal" then
            vim.cmd("startinsert!")
          end
        end)
      end,
    })

    _G.lazygit_toggle = function()
      lazygit:toggle()
    end

    _G.lazygit_hide = function()
      if lazygit:is_open() and lazygit:is_focused() then
        lazygit:close()
        return 0
      end
      error("lazygit is not the focused window")
    end

    -- Hunk統合(AIエージェント差分レビュー用TUI)
    local function new_hunk_terminal(cmd, count)
      return Terminal:new({
        cmd = cmd,
        direction = "float",
        hidden = true,
        count = count,
        env = {
          NVIM = vim.v.servername,
          -- Hunkの`e`キー(open file in $EDITOR)を、lazygitのnvim-remoteプリセット同様
          -- 既存Neovimインスタンスのバッファとして開くようリダイレクトする
          EDITOR = vim.fn.stdpath("config") .. "/bin/hunk-nvim-editor/nvim",
        },
        float_opts = {
          border = "curved",
          width = function() return math.floor(vim.o.columns * 0.95) end,
          height = function() return math.floor(vim.o.lines * 0.95) end,
        },
        on_open = function(term)
          vim.cmd("startinsert!")
          vim.api.nvim_buf_set_keymap(term.bufnr, "n", "q", "<cmd>close<CR>", { noremap = true, silent = true })
          -- hunkバッファ限定: フォーカス位置とレビューノートをClaudeへ@参照で送る
          vim.api.nvim_buf_set_keymap(term.bufnr, "t", "<C-a>", "", {
            noremap = true,
            silent = true,
            callback = function()
              require("utils.hunk_notes").send(term, cmd)
            end,
          })
        end,
        on_close = function(term)
          vim.schedule(function()
            if vim.bo.buftype == "terminal" then
              vim.cmd("startinsert!")
            end
          end)
        end,
      })
    end

    local hunk = new_hunk_terminal("hunk diff", 98)
    local hunk_staged = new_hunk_terminal("hunk diff --staged", 96)

    _G.hunk_toggle = function()
      hunk:toggle()
    end

    _G.hunk_staged_toggle = function()
      hunk_staged:toggle()
    end

    _G.hunk_hide = function()
      for _, term in ipairs({ hunk, hunk_staged }) do
        if term:is_open() and term:is_focused() then
          term:close()
          return 0
        end
      end
      error("hunk is not the focused window")
    end

    _G.toggle_last_terminal = function()
      if lazygit:is_open() and lazygit:is_focused() then lazygit:close() end
      if hunk:is_open() and hunk:is_focused() then hunk:close() end
      if hunk_staged:is_open() and hunk_staged:is_focused() then hunk_staged:close() end
      vim.cmd("ToggleTerm")
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

    -- 番号付きターミナルのうち存在するものだけを昇順で返す
    -- (lazygit=99 / hunk=98 / hunk_staged=96 はサイクル対象外)
    local function numbered_terminal_ids()
      local ids = {}
      for _, term in ipairs(require("toggleterm.terminal").get_all(true)) do
        if term.id >= 1 and term.id <= MAX_NUMBERED_TERMINALS then table.insert(ids, term.id) end
      end
      return ids
    end

    -- 指定ターミナルへ切り替える(トグルせず必ず表示・フォーカスする)
    _G.focus_terminal = function(id)
      local terms = require("toggleterm.terminal")
      if _G.terminal_mode == "single" then
        for _, term in ipairs(terms.get_all(true)) do
          if term.id ~= id and term.id >= 1 and term.id <= MAX_NUMBERED_TERMINALS and term:is_open() then
            term:close()
          end
        end
      end
      local term = terms.get(id, true)
      if term and term:is_open() then
        term:focus()
      else
        vim.cmd(id .. "ToggleTerm direction=" .. _G.terminal_direction)
      end
    end

    -- ターミナルサイクル切り替え(step=1で次、-1で前)
    _G.cycle_terminal = function(step)
      local ids = numbered_terminal_ids()
      if #ids == 0 then
        _G.focus_terminal(1)
        return
      end

      -- 起点はカレントバッファ、無ければ最後にフォーカスしたターミナル
      local terms = require("toggleterm.terminal")
      local current = vim.b.toggle_number
      if not current then
        local last = terms.get_last_focused()
        current = last and last.id or nil
      end

      local index = nil
      for i, id in ipairs(ids) do
        if id == current then index = i end
      end
      -- 起点が特定できないときは開いているターミナルへ戻すだけにする
      if not index then
        for _, id in ipairs(ids) do
          local term = terms.get(id, true)
          if term and term:is_open() then
            _G.focus_terminal(id)
            return
          end
        end
        _G.focus_terminal(ids[1])
        return
      end

      _G.focus_terminal(ids[(index - 1 + (step or 1)) % #ids + 1])
    end

    -- 未使用の最小番号でターミナルを新規作成する
    _G.new_terminal = function()
      local used = {}
      for _, term in ipairs(require("toggleterm.terminal").get_all(true)) do
        used[term.id] = true
      end
      for id = 1, MAX_NUMBERED_TERMINALS do
        if not used[id] then
          _G.toggle_smart_terminal(id)
          return
        end
      end
      vim.notify("ターミナルは最大" .. MAX_NUMBERED_TERMINALS .. "個までです", vim.log.levels.WARN)
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
      function() _G.toggle_smart_terminal(1) end,
      mode = "n",
      desc = "Terminal 1",
    },
    {
      "<leader>t1",
      [[<C-\><C-n><cmd>lua _G.toggle_smart_terminal(1)<CR>]],
      mode = "t",
      desc = "Terminal 1",
    },
    {
      "<leader>t2",
      function() _G.toggle_smart_terminal(2) end,
      mode = "n",
      desc = "Terminal 2",
    },
    {
      "<leader>t2",
      [[<C-\><C-n><cmd>lua _G.toggle_smart_terminal(2)<CR>]],
      mode = "t",
      desc = "Terminal 2",
    },
    {
      "<leader>t3",
      function() _G.toggle_smart_terminal(3) end,
      mode = "n",
      desc = "Terminal 3",
    },
    {
      "<leader>t3",
      [[<C-\><C-n><cmd>lua _G.toggle_smart_terminal(3)<CR>]],
      mode = "t",
      desc = "Terminal 3",
    },
    -- ターミナル間の切り替え(ノーマルのみ。tモードでは ] [ をシェルへ渡す)
    {
      "]t",
      function() _G.cycle_terminal(1) end,
      mode = "n",
      desc = "Next Terminal",
    },
    {
      "[t",
      function() _G.cycle_terminal(-1) end,
      mode = "n",
      desc = "Previous Terminal",
    },
    -- 新規ターミナル
    {
      "<leader>tn",
      function() _G.new_terminal() end,
      mode = "n",
      desc = "New Terminal",
    },
    {
      "<leader>tn",
      [[<C-\><C-n><cmd>lua _G.new_terminal()<CR>]],
      mode = "t",
      desc = "New Terminal",
    },
    -- 最後のターミナルトグル
    -- tモードは<C-\><C-n>でterminal normalに出てからクローズ。
    -- これにより閉じ時のModeChangedパターンがnt:nとなり、which-keyが正しく再アタッチできる。
    {
      "<C-/>",
      function() _G.toggle_last_terminal() end,
      mode = "n",
      desc = "Toggle Terminal (Last, shell only)",
    },
    {
      "<C-/>",
      [[<C-\><C-n><cmd>lua _G.toggle_last_terminal()<CR>]],
      mode = "t",
      desc = "Toggle Terminal (Last, shell only, exit insert first)",
    },
    {
      "<C-_>",
      function() _G.toggle_last_terminal() end,
      mode = "n",
      desc = "Toggle Terminal (Last, shell only)",
    },
    {
      "<C-_>",
      [[<C-\><C-n><cmd>lua _G.toggle_last_terminal()<CR>]],
      mode = "t",
      desc = "Toggle Terminal (Last, shell only, exit insert first)",
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

    -- Hunk (diff review)
    {
      "<leader>gr",
      function()
        _G.hunk_toggle()
      end,
      mode = { "n" },
      desc = "Hunk (diff review)",
    },
    {
      "<leader>gV",
      function()
        _G.hunk_staged_toggle()
      end,
      mode = { "n" },
      desc = "Hunk (staged diff review)",
    },

    -- ターミナルモード操作
    {
      "<Esc><Esc>",
      [[<C-\><C-n>]],
      mode = "t",
      desc = "Exit terminal mode",
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
  },
}
