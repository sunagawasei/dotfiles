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
            h1 = "#F2FFFF", -- ハイライト白
            h2 = "#E6F1F0", -- ニアホワイト
            h3 = "#D7E2E1", -- 前景
            h4 = "#CDD8D7", -- ライトグレー
            h5 = "#AAB6B5", -- 明るいグレー
            h6 = "#7E8A89", -- 中間グレー
          }

          for i = 1, 6 do
            local bg_name = "RenderMarkdownH" .. i .. "Bg"
            local fg_name = "RenderMarkdownH" .. i
            vim.api.nvim_set_hl(0, bg_name, {
              bg = colors["h" .. i],
              fg = "#0E1210",  -- 背景色を前景に
              bold = true
            })
            vim.api.nvim_set_hl(0, fg_name, {
              fg = colors["h" .. i],
              bold = true
            })
          end

          -- その他のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#2A2F2E" })  -- 濃い影
          vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#2A2F2E", fg = "#D7E2E1" })  -- 濃い影 + 前景
          vim.api.nvim_set_hl(0, "RenderMarkdownLink", { fg = "#5AAFAD", underline = true })  -- cyanアクセント
          vim.api.nvim_set_hl(0, "RenderMarkdownQuote", { fg = "#7E8A89", italic = true })  -- 中間グレー
          vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = "#5AAFAD" })  -- cyanアクセント
          vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = "#3A3F3E" })  -- 分割線
          vim.api.nvim_set_hl(0, "RenderMarkdownTodo", { fg = "#8C83A3" })  -- magentaアクセント
          vim.api.nvim_set_hl(0, "RenderMarkdownImportant", { fg = "#B3A9D1" })  -- bright magenta

          -- テーブル関連のハイライト設定（モノクロ基調）
          vim.api.nvim_set_hl(0, "RenderMarkdownTableHead", { fg = "#E6F1F0", bold = true })  -- ニアホワイト
          vim.api.nvim_set_hl(0, "RenderMarkdownTableRow", { fg = "#D7E2E1" })  -- 前景
          vim.api.nvim_set_hl(0, "RenderMarkdownTableFill", { fg = "#3A3F3E" })  -- 分割線
        end,
      })

      -- 初期化時にもカラー設定を適用
      vim.schedule(function()
        vim.cmd("doautocmd ColorScheme")
      end)
    end,
  },
}
