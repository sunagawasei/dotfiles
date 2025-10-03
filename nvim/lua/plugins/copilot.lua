return {
  -- GitHub Copilot (copilot.vimを無効化してLazyVimのcopilot.luaを使用)
  {
    "github/copilot.vim",
    enabled = false,  -- copilot.vimを無効化
  },

  -- copilot.luaの設定をオーバーライド（デフォルトオフ設定）
  {
    "zbirenbaum/copilot.lua",
    enabled = false,
    opts = {
      -- 補完をデフォルトで無効化
      suggestion = {
        enabled = false,           -- 起動時は無効
        auto_trigger = false,      -- 自動表示無効
        keymap = {
          accept = "<Tab>",        -- Tabで補完を受け入れ
          accept_word = false,
          accept_line = false,
          next = "<M-]>",         -- Alt+]で次の提案
          prev = "<M-[>",         -- Alt+[で前の提案
          dismiss = "<C-]>",      -- Ctrl+]で提案を閉じる
        },
      },
      -- パネルも無効化
      panel = { enabled = false },
      filetypes = {
        markdown = true,
        help = true,
      },
    },
    keys = {
      -- Copilot補完のトグル
      {
        "<leader>ct",
        function()
          -- suggestion機能の有効/無効を切り替え
          local copilot = require("copilot.suggestion")
          local current_state = require("copilot.config").get().suggestion.enabled

          if current_state then
            copilot.dismiss()
            require("copilot.config").set({
              suggestion = { enabled = false, auto_trigger = false }
            })
            vim.notify("Copilot suggestions disabled", vim.log.levels.INFO)
          else
            require("copilot.config").set({
              suggestion = { enabled = true, auto_trigger = true }
            })
            vim.notify("Copilot suggestions enabled", vim.log.levels.INFO)
          end
        end,
        desc = "Toggle Copilot Suggestion",
      },
    },
  },
}
