return {
  "folke/snacks.nvim",
  opts = {
    dashboard = {
      enabled = false,  -- oil.nvim起動のためdashboardを無効化
    },
    terminal = {
      enabled = true,
    },
    lazygit = {
      enabled = true,
      configure = false,  -- 自動テーマ設定を無効化（パフォーマンス改善）
      config = {
        -- os = { editPreset = "nvim-remote" },  -- nvr未インストールのため無効化
        gui = { nerdFontsVersion = "3" },
      },
      win = {
        style = "lazygit",
        wo = {
          scrolloff = 0,  -- スクロールオフセットを無効化してパフォーマンス改善
        },
      },
    },
    picker = {
      sources = {
        explorer = {
          hidden = true,
          ignored = true,
          exclude = { ".git", ".DS_Store" },
        },
        files = {
          hidden = true,
          ignored = true,
        },
      },
      -- カスタムアクションの定義
      actions = {
        copy_diagnostic = function(picker)
          local item = picker:current()
          if not item then
            vim.notify("No item selected", vim.log.levels.INFO)
            return
          end

          local text_to_copy = nil

          -- 診断情報がある場合
          if item.diagnostics and #item.diagnostics > 0 then
            local messages = {}
            for _, diag in ipairs(item.diagnostics) do
              table.insert(messages, diag.message)
            end
            text_to_copy = table.concat(messages, "\n")
          -- LSP結果やファイル内容がある場合
          elseif item.text then
            text_to_copy = item.text
          elseif item.value then
            text_to_copy = item.value
          -- ファイルパスがある場合は行内容を取得
          elseif item.file and item.lnum then
            local ok, lines = pcall(vim.fn.readfile, item.file, "", item.lnum)
            if ok and lines and lines[item.lnum] then
              text_to_copy = lines[item.lnum]
            end
          end

          if text_to_copy then
            vim.fn.setreg('+', text_to_copy)
            vim.notify("Copied to clipboard: " .. text_to_copy:sub(1, 50) .. (text_to_copy:len() > 50 and "..." or ""), vim.log.levels.INFO)
          else
            vim.notify("Nothing to copy", vim.log.levels.WARN)
          end
        end,
      },
      -- picker内のキーマップ設定
      win = {
        input = {
          keys = {
            -- 診断コピー用キーマップ
            ["<C-y>"] = { "copy_diagnostic", mode = { "n", "i" } },
            ["gy"] = { "copy_diagnostic", mode = { "n" } },
          },
        },
      },
    },
    bigfile = {
      enabled = false, -- bigfile機能を完全に無効化
    },
    indent = {
      enabled = false, -- インデントガイドを無効化（全階層表示を防ぐ）
    },
    zen = {
      enabled = true, -- Zoom機能を有効化
    },
  },
  keys = {
    -- 既存のカスタムターミナル（維持）
    { "<leader>gg", function() Snacks.lazygit() end, desc = "Lazygit" },
    { "<leader>gk", function() Snacks.terminal("keifu") end, desc = "Keifu (Git Graph)" },

    -- 位置別ターミナル切り替え
    {
      "<leader>tr",
      function() Snacks.terminal(nil, { win = { position = "right" } }) end,
      desc = "Terminal (Right)",
    },
    {
      "<leader>tb",
      function() Snacks.terminal(nil, { win = { position = "bottom" } }) end,
      desc = "Terminal (Bottom)",
    },
    {
      "<leader>tf",
      function() Snacks.terminal(nil, { win = { position = "float" } }) end,
      desc = "Terminal (Float)",
    },

    -- ターミナルモードでのESCキーバインド
    { "<Esc><Esc>", [[<C-\><C-n>]], mode = "t", desc = "Exit terminal mode" },
    { "<C-q>", [[<C-\><C-n>]], mode = "t", desc = "Exit terminal mode (quick)" },
    { "jk", [[<C-\><C-n>]], mode = "t", desc = "Exit terminal mode (jk)" },

    -- ターミナルモードでのウィンドウ移動
    { "<C-h>", [[<Cmd>wincmd h<CR>]], mode = "t", desc = "Go to left window" },
    { "<C-j>", [[<Cmd>wincmd j<CR>]], mode = "t", desc = "Go to lower window" },
    { "<C-k>", [[<Cmd>wincmd k<CR>]], mode = "t", desc = "Go to upper window" },
    { "<C-l>", [[<Cmd>wincmd l<CR>]], mode = "t", desc = "Go to right window" },
  },
}

