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

    -- diff表示設定
    diff_opts = {
      open_in_new_tab = true,           -- 新規タブでdiffを開く（作業中バッファを汚さない）
      hide_terminal_in_new_tab = true,  -- 新規タブでターミナル非表示（diffに集中）
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
    -- claudecode.nvim は visual 選択の終端列を byte 単位の inclusive index として扱う。
    -- 終端がマルチバイト文字の場合に欠けた UTF-8 を送らないよう、
    -- setup() 前に抽出と送信を補正する。
    local utf8_ok, utf8 = pcall(require, "utils.claudecode_utf8")
    if utf8_ok then
      utf8.patch_selection_function("get_visual_selection", utf8.wrap_selection_result)
      utf8.patch_selection_function("get_visual_selection_from_marks", utf8.wrap_selection_result)
      utf8.patch_selection_function("send_selection_update", utf8.wrap_send_selection_update)
    else
      vim.notify("claudecode.nvim UTF-8 patch could not be loaded", vim.log.levels.WARN)
    end

    -- closeAllDiffTabs が Diffview など「自分で開いた diff」まで閉じてしまう問題への対処。
    -- 元ハンドラは diff モードの全ウィンドウを無差別に閉じるため、Diffview の左右ペイン
    -- （vim ネイティブの diff モード）が Claude Code の submit のたびに巻き込まれて閉じる。
    -- claudecode が追跡する diff のみを閉じるハンドラに差し替える。
    -- setup() 内の register_all() は登録時にモジュールの .handler をコピーするため、
    -- setup() の「前」にモジュール側を差し替えれば、登録時（再接続での再登録も含む）に確実に拾われる。
    local cad_ok, close_all = pcall(require, "claudecode.tools.close_all_diff_tabs")
    if cad_ok then
      close_all.handler = function(_)
        local closed = 0
        local diff_ok, diff = pcall(require, "claudecode.diff")
        if diff_ok then
          closed = diff.close_all_diffs("closeAllDiffTabs tool") or 0
        end
        return {
          content = { { type = "text", text = "CLOSED_" .. closed .. "_DIFF_TABS" } },
        }
      end
    end

    -- setup()を呼び出す（この時点で上記の差し替え済みハンドラが登録される）
    require("claudecode").setup(opts)

    -- 差分表示バッファで行折り返しを有効化
    vim.api.nvim_create_autocmd({ "BufEnter", "BufWinEnter" }, {
      group = vim.api.nvim_create_augroup("claudecode_diff_wrap", { clear = true }),
      callback = function(event)
        local bufnr = event.buf

        -- claudecode.nvimのバッファ変数で確実に検出
        if vim.b[bufnr].claudecode_diff_tab_name then
          vim.opt_local.wrap = true         -- 行折り返しを有効化
          vim.opt_local.linebreak = false   -- 画面幅ベースで折り返し（日本語対応）
          vim.opt_local.breakindent = true  -- 折り返し行のインデント保持
          vim.opt_local.showbreak = "> "    -- 折り返し行マーカー（日本語対応）
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
      ft = { "oil" },
    },

    -- Diff管理（v0.3.0新機能）
    { "<leader>aa", "<cmd>ClaudeCodeDiffAccept<cr>", desc = "Accept diff changes" },
    { "<leader>ad", "<cmd>ClaudeCodeDiffDeny<cr>", desc = "Deny diff changes" },

  },
}
