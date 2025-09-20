return {
  "coder/claudecode.nvim",
  dependencies = { "folke/snacks.nvim" },
  lazy = false, -- 遅延ロードを無効化してNeovim起動時に即座に初期化
  priority = 1000, -- 早期にロードしてWebSocketサーバーを自動起動
  opts = {
    -- Claude Code CLI設定
    terminal_cmd = "/Users/s23159/.local/bin/claude --ide",
    auto_start = true, -- SSEサーバーを自動起動

    -- ターミナル設定（v0.3.0形式）
    terminal = {
      provider = "none", -- ターミナル管理を無効化、手動でClaude CLIを起動
      -- WebSocketサーバーは自動起動し、MCPプロトコルは完全に動作
      -- tmux、kitty、別ターミナルウィンドウなどで手動で'claude'コマンドを実行すると自動接続
    },

  },
  config = true, -- setup()を確実に呼び出す（必須）
  keys = {
    { "<leader>a", nil, desc = "AI/Claude Code" },


    -- Claude Code統合の開始・停止
    { "<leader>aI", "<cmd>ClaudeCodeStart<cr>", desc = "Start Claude Code integration" },
    { "<leader>aS", "<cmd>ClaudeCodeStop<cr>", desc = "Stop Claude Code integration" },
    { "<leader>ai", "<cmd>ClaudeCodeStatus<cr>", desc = "Show Claude Code status" },

    -- ファイル・コンテキスト管理
    { "<leader>ab", "<cmd>ClaudeCodeAdd %<cr>", desc = "Add current buffer to Claude" },
    { "<leader>as", "<cmd>ClaudeCodeSend<cr>", mode = "v", desc = "Send selection to Claude" },
    {
      "<leader>at",
      "<cmd>ClaudeCodeTreeAdd<cr>",
      desc = "Add file from tree explorer",
      ft = { "NvimTree", "oil", "mini.files" },
    },

    -- Diff管理（v0.3.0新機能）
    { "<leader>aa", "<cmd>ClaudeCodeDiffAccept<cr>", desc = "Accept diff changes" },
    { "<leader>ad", "<cmd>ClaudeCodeDiffDeny<cr>", desc = "Deny diff changes" },

  },
}
