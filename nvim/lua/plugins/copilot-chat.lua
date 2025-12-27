-- CopilotChat.nvim キーマップオーバーライド
-- LazyVimデフォルトの<leader>a*を無効化し、<leader>c*に変更
-- （ClaudeCode.nvimの<leader>a*と競合を避けるため）
return {
  "CopilotC-Nvim/CopilotChat.nvim",
  opts = {
    show_help = false,
    model = "gpt-5.1-codex",
    auto_insert_mode = false,
    system_prompt = [[あなたは優秀なプログラミングアシスタントです。

【最重要ルール】
- すべての回答・説明は必ず日本語で行ってください
- 英語での説明は絶対に禁止です
- コード内のコメントは英語で構いません

このルールは例外なく、すべての質問に適用されます。]],
    prompts = {
      Explain = {
        prompt = "選択されたコードの説明を日本語で書いてください。",
      },
      Review = {
        prompt = "選択されたコードをレビューしてください。問題点や改善点を日本語で指摘してください。",
      },
      Fix = {
        prompt = "このコードには問題があります。問題を修正したコードを日本語の説明付きで書き直してください。",
      },
      Optimize = {
        prompt = "選択されたコードを最適化してください。改善点を日本語で説明してください。",
      },
      Docs = {
        prompt = "選択されたコードにドキュメントコメントを日本語で追加してください。",
      },
      Tests = {
        prompt = "選択されたコードのテストを生成してください。説明は日本語で書いてください。",
      },
    },
  },
  keys = {
    -- LazyVimデフォルト(<leader>a*)を無効化
    { "<leader>aa", false },
    { "<leader>ax", false },
    { "<leader>aq", false },
    { "<leader>ap", false },
    -- 新しいキーマップ(<leader>c*)
    { "<leader>cp", "<cmd>CopilotChatToggle<cr>", desc = "CopilotChat Toggle", mode = { "n", "v" } },
    { "<leader>ce", "<cmd>CopilotChatExplain<cr>", desc = "Explain selection", mode = "v" },
    { "<leader>cr", "<cmd>CopilotChatReview<cr>", desc = "Review selection", mode = "v" },
    {
      "<leader>cq",
      function()
        local input = vim.fn.input("Quick Chat: ")
        if input ~= "" then
          require("CopilotChat").ask(input .. "\n\n（必ず日本語で回答してください）")
        end
      end,
      desc = "Quick Chat",
      mode = { "n", "v" },
    },
    { "<leader>cx", "<cmd>CopilotChatReset<cr>", desc = "Clear chat", mode = { "n", "v" } },
  },
}

