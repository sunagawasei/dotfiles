return {
  {
    "voldikss/vim-floaterm",
    event = "VeryLazy",
    lazy = false,
    config = function()
      -- Configure vim-floaterm here later
    end,
  },
  {
    "yetone/avante.nvim",
    event = "VeryLazy",
    lazy = true,
    version = false,
    build = "make BUILD_FROM_SOURCE=false", -- プリビルドバイナリを使用
    cmd = { "AvanteAsk", "AvanteEdit", "AvanteRefresh", "AvanteFocus" },
    config = function()
      require("avante").setup({
        provider = "copilot",
        auto_suggestions_provider = "copilot",
        providers = {
          copilot = {
            endpoint = "https://api.githubcopilot.com",
            proxy = nil,
            allow_insecure = false,
            timeout = 30000, -- 30秒に短縮（元: 10分）
            model = "claude-sonnet-4",
            extra_request_body = {
              temperature = 0,
              max_completion_tokens = 4096, -- トークン数を制限（元: 1000000）
              reasoning_effort = "medium", -- 推論レベルを下げる（元: high）
            },
          },
        },
        -- 動作設定
        behaviour = {
          auto_suggestions = false, -- 自動提案を無効化（パフォーマンス向上）
          auto_set_highlight_group = true,
          auto_set_keymaps = true,
          auto_apply_diff_after_generation = false,
          support_paste_from_clipboard = false,
          minimize_diff = true,
          enable_token_counting = false,
        },

        -- パフォーマンス設定
        performance = {
          max_suggestions = 50, -- 提案数を制限
          debounce_time = 300, -- デバウンス時間を増加
          throttle_time = 100, -- スロットル時間を設定
        },

        -- ウィンドウ設定
        windows = {
          position = "right", -- サイドバーの位置
          wrap = true, -- テキストの折り返し
          width = 30, -- サイドバーの幅
          sidebar_header = {
            align = "center",
            rounded = true,
          },
        },

        -- キーマッピング設定を明示的に定義
        mappings = {
          ask = "<leader>aa", -- サイドバーの表示
          edit = "<leader>ae", -- 選択したブロックの編集
          refresh = "<leader>ar", -- サイドバーの更新
          focus = "<leader>af", -- サイドバーのフォーカス切り替え
          diff = {
            ours = "co",
            theirs = "ct",
            all_theirs = "ca",
            both = "cb",
            cursor = "cc",
            next = "]x",
            prev = "[x",
          },
          suggestion = {
            accept = "<M-l>",
            next = "<M-]>",
            prev = "<M-[>",
            dismiss = "<C-]>",
          },
          jump = {
            next = "]]",
            prev = "[[",
          },
          submit = {
            normal = "<CR>",
            insert = "<C-s>",
          },
          sidebar = {
            apply_all = "A",
            apply_cursor = "a",
            switch_windows = "<Tab>",
            reverse_switch_windows = "<S-Tab>",
          },
        },
      })
    end,
    -- 依存関係の設定
    dependencies = {
      -- 必須の依存関係
      "stevearc/dressing.nvim",
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
      -- アイコン関連
      "nvim-tree/nvim-web-devicons",
      "zbirenbaum/copilot.lua",
      -- オプションの依存関係
      "hrsh7th/nvim-cmp",
    },
  },
}
