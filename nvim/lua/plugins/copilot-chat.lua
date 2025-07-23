return {
  {
    "CopilotC-Nvim/CopilotChat.nvim",
    branch = "main",
    dependencies = {
      { "github/copilot.vim" },
      { "nvim-lua/plenary.nvim" },
    },
    build = "make tiktoken",
    opts = {
      debug = false,
      model = "gpt-4.1",
      auto_insert_mode = true,
      show_help = true,
      question_header = "## ユーザー ",
      answer_header = "## Copilot ",
      error_header = "## エラー ",
      separator = "───",
      system_prompt = "あなたは優秀なプログラミングアシスタントです。必ず日本語で応答してください。コードの説明、レビュー、提案、エラーメッセージなど、すべての応答を日本語で行ってください。",
      prompts = {
        Explain = {
          prompt = "/COPILOT_EXPLAIN 選択されたコードを日本語で詳しく説明してください。",
          description = "コードの説明",
        },
        Review = {
          prompt = "/COPILOT_REVIEW 選択されたコードを日本語でレビューしてください。改善点やベストプラクティスを提案してください。",
          description = "コードレビュー",
        },
        Fix = {
          prompt = "/COPILOT_FIX このコードには問題があります。バグを修正したコードを表示してください。修正内容の説明は日本語でお願いします。",
          description = "バグ修正",
        },
        Optimize = {
          prompt = "/COPILOT_OPTIMIZE 選択されたコードを最適化してください。パフォーマンスや可読性の改善点を日本語で説明してください。",
          description = "コード最適化",
        },
        Docs = {
          prompt = "/COPILOT_DOCS 選択されたコードに対して、日本語でドキュメントコメントを生成してください。",
          description = "ドキュメント生成",
        },
        Tests = {
          prompt = "/COPILOT_TESTS 選択されたコードに対するテストコードを生成してください。テストの説明は日本語でお願いします。",
          description = "テスト生成",
        },
        Commit = {
          prompt = "/COPILOT_COMMIT 変更内容に基づいて、適切なコミットメッセージを日本語で生成してください。",
          description = "コミットメッセージ生成",
        },
      },
      window = {
        layout = "vertical",
        width = 0.5,
        height = 0.5,
        relative = "editor",
        border = "rounded",
        title = "Copilot Chat",
        footer = nil,
        zindex = 1,
      },
      selection = function(source)
        return require("CopilotChat.select").buffer(source)
      end,
      mappings = {
        complete = {
          detail = "Use @<Tab> or /<Tab> for options.",
          insert = "<Tab>",
        },
        close = {
          normal = "q",
          insert = "<C-c>",
        },
        reset = {
          normal = "<C-r>",
          insert = "<C-r>",
        },
        submit_prompt = {
          normal = "<CR>",
          insert = "<C-s>",
        },
        accept_diff = {
          normal = "<C-y>",
          insert = "<C-y>",
        },
        yank_diff = {
          normal = "gy",
        },
        show_diff = {
          normal = "gd",
        },
        show_system_prompt = {
          normal = "gp",
        },
        show_user_selection = {
          normal = "gs",
        },
      },
    },
    keys = {
      {
        "<leader>cc",
        "<cmd>CopilotChatOpen<cr>",
        desc = "CopilotChat - Open",
      },
      {
        "<leader>ct",
        "<cmd>CopilotChatToggle<cr>",
        desc = "CopilotChat - Toggle",
      },
      {
        "<leader>cs",
        "<cmd>CopilotChatStop<cr>",
        desc = "CopilotChat - Stop",
      },
      {
        "<leader>cr",
        "<cmd>CopilotChatReset<cr>",
        desc = "CopilotChat - Reset",
      },
      {
        "<leader>ce",
        function()
          local input = vim.fn.input("Ask Copilot: ")
          if input ~= "" then
            require("CopilotChat").ask(input, { selection = require("CopilotChat.select").visual })
          end
        end,
        mode = "v",
        desc = "CopilotChat - Ask about selection",
      },
      {
        "<leader>cq",
        function()
          local input = vim.fn.input("Quick Chat: ")
          if input ~= "" then
            require("CopilotChat").ask(input, { selection = require("CopilotChat.select").buffer })
          end
        end,
        desc = "CopilotChat - Quick chat",
      },
      {
        "<leader>ch",
        function()
          local actions = require("CopilotChat.actions")
          require("CopilotChat.integrations.telescope").pick(actions.help_actions())
        end,
        desc = "CopilotChat - Help actions",
      },
      {
        "<leader>cp",
        function()
          local actions = require("CopilotChat.actions")
          require("CopilotChat.integrations.telescope").pick(actions.prompt_actions())
        end,
        desc = "CopilotChat - Prompt actions",
      },
      {
        "<leader>cd",
        "<cmd>CopilotChatDebugInfo<cr>",
        desc = "CopilotChat - Debug Info",
      },
      {
        "<leader>cf",
        "<cmd>CopilotChatFixDiagnostic<cr>",
        desc = "CopilotChat - Fix Diagnostic",
      },
      {
        "<leader>cl",
        "<cmd>CopilotChatLoad<cr>",
        desc = "CopilotChat - Load session",
      },
      {
        "<leader>cv",
        "<cmd>CopilotChatSave<cr>",
        desc = "CopilotChat - Save session",
      },
      -- 日本語プロンプト用のキーバインド
      {
        "<leader>cE",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_EXPLAIN 選択されたコードを日本語で詳しく説明してください。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - コード説明",
      },
      {
        "<leader>cR",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_REVIEW 選択されたコードを日本語でレビューしてください。改善点やベストプラクティスを提案してください。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - コードレビュー",
      },
      {
        "<leader>cF",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_FIX このコードには問題があります。バグを修正したコードを表示してください。修正内容の説明は日本語でお願いします。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - バグ修正",
      },
      {
        "<leader>cO",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_OPTIMIZE 選択されたコードを最適化してください。パフォーマンスや可読性の改善点を日本語で説明してください。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - コード最適化",
      },
      {
        "<leader>cD",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_DOCS 選択されたコードに対して、日本語でドキュメントコメントを生成してください。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - ドキュメント生成",
      },
      {
        "<leader>cT",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_TESTS 選択されたコードに対するテストコードを生成してください。テストの説明は日本語でお願いします。",
            selection = require("CopilotChat.select").visual,
          })
        end,
        mode = "v",
        desc = "CopilotChat - テスト生成",
      },
      {
        "<leader>cC",
        function()
          require("CopilotChat").ask("", {
            prompt = "/COPILOT_COMMIT 変更内容に基づいて、適切なコミットメッセージを日本語で生成してください。",
            selection = require("CopilotChat.select").buffer,
          })
        end,
        desc = "CopilotChat - コミットメッセージ生成",
      },
    },
  },
}

