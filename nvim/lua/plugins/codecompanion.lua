return {
  "olimorris/codecompanion.nvim",
  version = "^18.0.0",
  dependencies = {
    "nvim-lua/plenary.nvim",
    "nvim-treesitter/nvim-treesitter",
  },
  opts = {
    strategies = {
      chat = {
        adapter = "copilot",
      },
      inline = {
        adapter = "copilot",
      },
      cmd = {
        adapter = "copilot",
      },
    },
    adapters = {
      copilot = function()
        return require("codecompanion.adapters").extend("copilot", {
          schema = {
            model = {
              default = "gpt-5.2",
            },
          },
          opts = {
            send_open_buffers = true,
          },
        })
      end,
    },
    display = {
      chat = {
        icons = {
          buffer_sync_all = "󰪴 ",
          buffer_sync_diff = " ",
          chat_fold = " ",
          tool_pending = " ",
          tool_in_progress = " ",
          tool_failure = " ",
          tool_success = " ",
        },
        show_settings = false,
        show_token_count = true,
        show_context = false, -- Context表示を非表示
        show_header_separator = true, -- ヘッダー区切り線を表示
        separator = "─", -- メッセージ間の区切り文字
        show_tools_processing = true, -- ツール実行中のローディング表示
        auto_scroll = true, -- 自動スクロール
        fold_context = true, -- コンテキストの折りたたみ
        show_reasoning = true, -- 推論過程を表示
        fold_reasoning = true, -- 推論過程を折りたたみ
      },
    },
    interactions = {
      chat = {
        roles = {
          llm = function(adapter)
            return "CodeCompanion (" .. adapter.formatted_name .. ")"
          end,
          user = "Me",
        },
      },
    },
    opts = {
      send_code = true,
      use_default_actions = true,
      use_default_prompt_library = true,
    },
    -- Claude Codeのような自動ファイル探索を実現
    default_tools = {
      "read_file",
      "file_search",
      "grep_search",
      "get_changed_files",
      "list_code_usages",
      "insert_edit_into_file",
    },
    auto_submit_success = true,
  },
}
