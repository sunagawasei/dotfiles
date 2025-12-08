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

    -- 差分表示・タブ管理設定（v0.3.0新機能）
    diff = {
      -- タブ自動命名を有効化
      additional_tab_options = true,

      -- カスタムタブ名フォーマット関数
      tab_name_format = function(filename, change_type)
        -- ファイル名からベース名を抽出
        local base = filename:match("([^/]+)$") or filename

        -- 変更タイプの日本語マッピング
        local type_map = {
          Refactor = "リファクタ",
          Fix = "修正",
          Add = "追加",
          Update = "更新",
          Remove = "削除",
          Diff = "差分"
        }

        local jp_type = type_map[change_type] or change_type

        -- タブ名を生成（例: "[修正] auth.go"）
        return string.format("[%s] %s", jp_type, base)
      end,
    },

  },
  config = function(_, opts)
    -- setup()を呼び出す
    require("claudecode").setup(opts)

    -- 差分表示バッファで行折り返しを有効化
    vim.api.nvim_create_autocmd({ "BufEnter", "BufWinEnter" }, {
      group = vim.api.nvim_create_augroup("claudecode_diff_wrap", { clear = true }),
      callback = function(event)
        local bufnr = event.buf

        -- claudecode.nvimのバッファ変数で確実に検出
        if vim.b[bufnr].claudecode_diff_tab_name then
          vim.opt_local.wrap = true         -- 行折り返しを有効化
          vim.opt_local.linebreak = true    -- 単語単位で折り返し
          vim.opt_local.breakindent = true  -- 折り返し行のインデント保持
          vim.opt_local.showbreak = "↪ "    -- 折り返し行の視覚的マーカー
        end
      end,
    })
  end,
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
      ft = { "neo-tree" },
    },

    -- Diff管理（v0.3.0新機能）
    { "<leader>aa", "<cmd>ClaudeCodeDiffAccept<cr>", desc = "Accept diff changes" },
    { "<leader>ad", "<cmd>ClaudeCodeDiffDeny<cr>", desc = "Deny diff changes" },

  },
}
