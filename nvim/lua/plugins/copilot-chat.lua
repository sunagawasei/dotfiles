return {
  {
    "CopilotC-Nvim/CopilotChat.nvim",
    branch = "main",
    dependencies = {
      { "github/copilot.vim" },
      { "nvim-lua/plenary.nvim", branch = "master" },
    },
    build = "make tiktoken",
    cmd = {
      "CopilotChat",
      "CopilotChatOpen",
      "CopilotChatClose",
      "CopilotChatToggle",
      "CopilotChatStop",
      "CopilotChatReset",
      "CopilotChatSave",
      "CopilotChatLoad",
      "CopilotChatModels",
      "CopilotChatPrompts",
    },
    opts = {
      debug = false,
      model = "claude-sonnet-4",
      auto_insert_mode = false,
      show_help = false,
      show_folds = true, -- コードブロックの折りたたみ表示を有効化
      auto_follow_cursor = false, -- カーソル追従を無効化
      highlight_headers = true, -- ヘッダーのハイライト
      highlight_selection = true, -- 選択部分のハイライト
      context = "buffer", -- コンテキストの設定
      system_prompt = "あなたは優秀なプログラミングアシスタントです。必ず日本語で応答してください。コードの説明、レビュー、提案、エラーメッセージなど、すべての応答を日本語で行ってください。",
      question_header = " ## User ",
      answer_header = " ## Copilot ",
      error_header = " ## Error ",
      separator = "---", -- セパレータ文字
      -- ハイライト設定（Vercel Geistカラーシステムに準拠）
      highlights = {
        user = {
          fg = "#0070F3", -- Geist Blue 8（ユーザー）
          bold = true,
        },
        system = {
          fg = "#F5A623", -- Geist Amber 7（システム）
          italic = true,
        },
        assistant = {
          fg = "#16A085", -- Geist Green 8（アシスタント）
        },
        help = {
          fg = "#A3A3A3", -- Geist Gray 6 / Color 9 Dark（ヘルプテキスト）
          italic = true,
        },
      },
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
        FixDiagnostic = {
          prompt = "/COPILOT_FIX_DIAGNOSTIC ファイル内の次のエラーメッセージを修正してください: ",
          description = "診断エラーの修正",
        },
        CommitStaged = {
          prompt = "/COPILOT_COMMIT_STAGED ステージされた変更に基づいて、適切なコミットメッセージを日本語で生成してください。コミットメッセージは簡潔で、変更内容を的確に表現してください。",
          description = "ステージ済み変更のコミットメッセージ生成",
        },
      },
      window = {
        layout = "vertical",
        width = 0.5,
        height = 0.8,
        relative = "editor",
        border = "rounded",
        title = " Copilot Chat ",
        footer = " <C-s> Send | <C-c> Close | <C-r> Reset | <Tab> Complete ",
        zindex = 50,
      },
      selection = function(source)
        return require("CopilotChat.select").visual(source) or require("CopilotChat.select").buffer(source)
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
          full_diff = true, -- 差分の完全表示
        },
        show_system_prompt = {
          normal = "gp",
        },
        show_user_selection = {
          normal = "gs",
        },
      },
    },
    config = function(_, opts)
      local chat = require("CopilotChat")
      chat.setup(opts)
      
      -- CopilotChatのバッファでマークダウンファイルタイプを設定
      vim.api.nvim_create_autocmd("BufEnter", {
        pattern = "copilot-*",
        callback = function()
          -- CopilotChat専用のマークダウンファイルタイプを設定
          vim.bo.filetype = "markdown.copilot-chat"
          -- 折り畳みを有効化（パフォーマンス最適化のためindentベースに変更）
          vim.wo.foldenable = true
          vim.wo.foldmethod = "indent"
          vim.wo.foldlevel = 99
        end,
      })
    end,
    keys = {
      -- 基本操作
      { "<leader>cc", "<cmd>CopilotChatOpen<cr>", desc = "CopilotChat - Open" },
      { "<leader>ct", "<cmd>CopilotChatToggle<cr>", desc = "CopilotChat - Toggle" },
      { "<leader>cs", "<cmd>CopilotChatStop<cr>", desc = "CopilotChat - Stop" },
      { "<leader>cr", "<cmd>CopilotChatReset<cr>", desc = "CopilotChat - Reset" },
      { "<leader>cl", "<cmd>CopilotChatLoad<cr>", desc = "CopilotChat - Load session" },

      -- インタラクティブな質問
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

      -- アクションピッカー
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
        "<leader>cf",
        function()
          local actions = require("CopilotChat.actions")
          require("CopilotChat.integrations.telescope").pick(actions.prompt_actions({
            selection = require("CopilotChat.select").diagnostics,
          }))
        end,
        desc = "CopilotChat - Fix Diagnostic",
      },

      -- 日本語プロンプト用のキーバインド
      { "<leader>cE", "<cmd>CopilotChatExplain<cr>", mode = {"n", "v"}, desc = "CopilotChat - コード説明" },
      { "<leader>cR", "<cmd>CopilotChatReview<cr>", mode = "v", desc = "CopilotChat - コードレビュー" },
      { "<leader>cF", "<cmd>CopilotChatFix<cr>", mode = "v", desc = "CopilotChat - バグ修正" },
      { "<leader>cO", "<cmd>CopilotChatOptimize<cr>", mode = "v", desc = "CopilotChat - コード最適化" },
      { "<leader>cD", "<cmd>CopilotChatDocs<cr>", mode = "v", desc = "CopilotChat - ドキュメント生成" },
      { "<leader>cT", "<cmd>CopilotChatTests<cr>", mode = "v", desc = "CopilotChat - テスト生成" },
      { "<leader>cC", "<cmd>CopilotChatCommit<cr>", desc = "CopilotChat - コミットメッセージ生成" },
    },
  },
}

