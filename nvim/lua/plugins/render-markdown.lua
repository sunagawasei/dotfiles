return {
  {
    "MeanderingProgrammer/render-markdown.nvim",
    opts = {
      -- LazyVimプリセットを使用
      preset = "lazy",
      -- デフォルトでレンダリングを無効化
      enabled = false,
    },
    ft = { "markdown", "gitcommit" },
    keys = {
      {
        "<leader>mr",
        "<cmd>RenderMarkdown toggle<cr>",
        desc = "Toggle Markdown Rendering",
      },
      {
        "<leader>mR",
        function()
          require("render-markdown").setup({})
        end,
        desc = "Reload Render Markdown",
      },
    },
    config = function(_, opts)
      require("render-markdown").setup(opts)

      -- モノクロ基調＋アクセントカラーテーマとの統合
      vim.api.nvim_create_autocmd("ColorScheme", {
        callback = function()
          -- ヘッダーカラーの設定（モノクロ階調）
          local colors = {
            h1 = "#A4E4E0", -- Highlight (2.7%)
            h2 = "#CEF5F2", -- Foreground (6.9%)
            h3 = "#659D9E", -- Teal Mid (2.7%)
            h4 = "#ABB0CC", -- Blue Light (1.4%)
            h5 = "#92A2AB", -- Blue/Gray (1.6%)
            h6 = "#586269", -- Mid Gray (2.5%)
          }

          for i = 1, 6 do
            local bg_name = "RenderMarkdownH" .. i .. "Bg"
            local fg_name = "RenderMarkdownH" .. i
            vim.api.nvim_set_hl(0, bg_name, {
              bg = colors["h" .. i],
              fg = "#132018",  -- Background (63.4%)
              bold = true
            })
            vim.api.nvim_set_hl(0, fg_name, {
              fg = colors["h" .. i],
              bold = true
            })
          end

          -- その他のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#1E2A2D" })  -- Panel Background (3.7%)
          vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#1E2A2D", fg = "#CEF5F2" })
          vim.api.nvim_set_hl(0, "RenderMarkdownLink", { fg = "#5ABDBC", underline = true })  -- Primary Accent (1.7%)
          vim.api.nvim_set_hl(0, "RenderMarkdownQuote", { fg = "#586269", italic = true })  -- Mid Gray
          vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = "#4A8778" })  -- Teal Green (Success)
          vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = "#29595C" })  -- Border
          vim.api.nvim_set_hl(0, "RenderMarkdownTodo", { fg = "#926894" })  -- Purple Muted (Error/Alert)
          vim.api.nvim_set_hl(0, "RenderMarkdownImportant", { fg = "#CED5E9" })  -- Lavender (Warning)

          -- テーブル関連のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownTableHead", { fg = "#A4E4E0", bold = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownTableRow", { fg = "#CEF5F2" })
          vim.api.nvim_set_hl(0, "RenderMarkdownTableFill", { fg = "#29595C" })  -- ボーダー
        end,
      })

      -- 初期化時にもカラー設定を適用
      vim.schedule(function()
        vim.cmd("doautocmd ColorScheme")
      end)
    end,
  },
}