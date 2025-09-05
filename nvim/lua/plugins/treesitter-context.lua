return {
  "nvim-treesitter/nvim-treesitter-context",
  event = { "BufReadPost", "BufNewFile" },
  enabled = true,  -- 明示的に有効化
  dependencies = {
    "nvim-treesitter/nvim-treesitter",
  },
  opts = {
    enable = true,        -- プラグインを有効化
    max_lines = 3,        -- 表示する最大行数
    min_window_height = 0, -- 最小ウィンドウ高さ（0で制限なし）
    line_numbers = true,  -- 行番号を表示
    multiline_threshold = 1, -- 複数行コンテキストのしきい値
    trim_scope = "outer", -- 表示範囲の調整（outer/inner）
    mode = "cursor",      -- カーソルに基づいてコンテキストを表示
    separator = nil,      -- セパレーター文字（nilでデフォルト）
    zindex = 20,          -- ウィンドウのレイヤー順位
    on_attach = nil,      -- アタッチ時のコールバック
    
    -- パフォーマンス設定
    throttle = true,      -- スクロール時のスロットリング有効
    
    -- パターン設定（どの構文要素を表示するか）
    patterns = {
      default = {
        "class",
        "function",
        "method",
        "for",
        "while",
        "if",
        "switch",
        "case",
        "interface",
        "struct",
        "enum",
      },
      go = {
        "func", 
        "type_declaration",
        "type_spec",
        "method_declaration",
        "function_declaration",
        "if_statement",
        "for_statement",
        "switch_statement",
        "select_statement",
        "struct_type",
        "interface_type",
      },
      lua = {
        "chunk",
        "function_call",
        "function_definition",
        "if_statement",
        "for_statement",
        "while_statement",
        "repeat_statement",
        "do_statement",
        "table_constructor",
      },
      python = {
        "function_definition",
        "class_definition",
        "for_statement",
        "while_statement", 
        "if_statement",
        "with_statement",
        "try_statement",
      },
      javascript = {
        "function",
        "function_declaration",
        "function_expression",
        "arrow_function",
        "method_definition",
        "class_declaration",
        "for_statement",
        "for_in_statement",
        "while_statement",
        "if_statement",
        "switch_statement",
        "try_statement",
        "object",
      },
      typescript = {
        "function",
        "function_declaration", 
        "function_expression",
        "arrow_function",
        "method_definition",
        "class_declaration",
        "interface_declaration",
        "type_alias_declaration",
        "for_statement",
        "for_in_statement",
        "while_statement",
        "if_statement",
        "switch_statement",
        "try_statement",
      },
    },
  },
  config = function(_, opts)
    -- Vercel Geistカラーテーマに準拠したハイライト設定
    vim.api.nvim_set_hl(0, "TreesitterContext", {
      bg = "#1a1a1a",     -- 背景色（Geist Gray 9）
      fg = "#ededed",     -- テキスト色（Geist Gray 1）
    })
    
    vim.api.nvim_set_hl(0, "TreesitterContextLineNumber", {
      bg = "#1a1a1a",     -- 背景色（Geist Gray 9）
      fg = "#999999",     -- 行番号色（Geist Gray 6）
    })
    
    vim.api.nvim_set_hl(0, "TreesitterContextSeparator", {
      fg = "#333333",     -- セパレーター色（Geist Gray 7）
    })
    
    -- プラグインの初期化
    require("treesitter-context").setup(opts)
    
    -- キーマップ設定
    vim.keymap.set("n", "[c", function()
      require("treesitter-context").go_to_context(vim.v.count1)
    end, { silent = true, desc = "Go to context" })
  end,
}
