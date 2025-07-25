return {
  {
    "MeanderingProgrammer/render-markdown.nvim",
    dependencies = { "nvim-treesitter/nvim-treesitter", "echasnovski/mini.nvim" },
    opts = {
      -- Vercel Geist カラーを適用
      highlights = {
        heading = {
          -- 見出しの前景色（Geistカラーを階層的に使用）
          foregrounds = {
            "RenderMarkdownH1",
            "RenderMarkdownH2",
            "RenderMarkdownH3",
            "RenderMarkdownH4",
            "RenderMarkdownH5",
            "RenderMarkdownH6",
          },
          -- 見出しの背景色
          backgrounds = {
            "RenderMarkdownH1Bg",
            "RenderMarkdownH2Bg",
            "RenderMarkdownH3Bg",
            "RenderMarkdownH4Bg",
            "RenderMarkdownH5Bg",
            "RenderMarkdownH6Bg",
          },
        },
      },
    },
    ft = { "markdown" },
    config = function(_, opts)
      require("render-markdown").setup(opts)
      
      -- Vercel Geist カラーでハイライトグループを設定
      -- 見出しの色（階層的に鮮やかさを調整）
      vim.api.nvim_set_hl(0, "RenderMarkdownH1", { fg = "#0070F3", bold = true }) -- Geist Blue 8
      vim.api.nvim_set_hl(0, "RenderMarkdownH1Bg", { bg = "#111111" }) -- Background 2 (Dark)
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH2", { fg = "#7928CA", bold = true }) -- Geist Purple 7
      vim.api.nvim_set_hl(0, "RenderMarkdownH2Bg", { bg = "#111111" })
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH3", { fg = "#0D9488", bold = true }) -- Geist Teal 7
      vim.api.nvim_set_hl(0, "RenderMarkdownH3Bg", { bg = "#111111" })
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH4", { fg = "#F5A623", bold = true }) -- Geist Amber 7
      vim.api.nvim_set_hl(0, "RenderMarkdownH4Bg", { bg = "#111111" })
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH5", { fg = "#16A085", bold = true }) -- Geist Green 8
      vim.api.nvim_set_hl(0, "RenderMarkdownH5Bg", { bg = "#111111" })
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH6", { fg = "#EC4899", bold = true }) -- Geist Pink 6
      vim.api.nvim_set_hl(0, "RenderMarkdownH6Bg", { bg = "#111111" })
      
      -- コードブロックの色
      vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#1A1A1A", fg = "#EDEDED" }) -- Color 2 (Dark) + Color 7 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#262626", fg = "#A3A3A3" }) -- Color 3 (Dark) + Color 9 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeBorder", { fg = "#444444" }) -- Color 5 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeInfo", { fg = "#666666" }) -- Color 6 (Dark)
      
      -- リストとチェックボックス
      vim.api.nvim_set_hl(0, "RenderMarkdownBullet", { fg = "#66C2FF" }) -- Geist Blue 6
      vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = "#666666" }) -- Color 6 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = "#50E3C2" }) -- Geist Green 6
      
      -- 引用符（ネストレベルごとに色を変更）
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote1", { fg = "#3291FF" }) -- Geist Blue 7
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote2", { fg = "#A855F7" }) -- Geist Purple 6
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote3", { fg = "#14B8A6" }) -- Geist Teal 6
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote4", { fg = "#FFCC33" }) -- Geist Amber 6
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote5", { fg = "#50E3C2" }) -- Geist Green 6
      vim.api.nvim_set_hl(0, "RenderMarkdownQuote6", { fg = "#EC4899" }) -- Geist Pink 6
      
      -- テーブル
      vim.api.nvim_set_hl(0, "RenderMarkdownTableHead", { fg = "#FFFFFF", bg = "#262626", bold = true }) -- Color 10 (Dark) + Color 3 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownTableRow", { fg = "#EDEDED", bg = "#1A1A1A" }) -- Color 7 (Dark) + Color 2 (Dark)
      
      -- ColorScheme変更時にハイライトを再適用
      vim.api.nvim_create_autocmd("ColorScheme", {
        pattern = "*",
        callback = function()
          -- 上記のハイライト設定を再実行
          vim.cmd("doautocmd User RenderMarkdownColors")
        end,
        group = vim.api.nvim_create_augroup("RenderMarkdownHighlights", { clear = true }),
      })
    end,
  },
}