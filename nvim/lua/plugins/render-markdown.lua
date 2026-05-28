return {
  {
    "MeanderingProgrammer/render-markdown.nvim",
    opts = {
      preset = "lazy",
      enabled = true, -- 編集時インラインレンダリングをデフォルト有効
      heading = {
        width = "block", -- 背景を見出し文字幅に制限（デフォルト "full" = 画面右端まで）
      },
    },
    ft = { "markdown", "gitcommit" },
    keys = {
      {
        "<leader>mr",
        "<cmd>RenderMarkdown toggle<cr>",
        ft = "markdown",
        desc = "Toggle Inline Markdown Render",
      },
      {
        "<leader>mR",
        "<cmd>RenderMarkdown reload<cr>",
        ft = "markdown",
        desc = "Reload Inline Markdown Render",
      },
    },
  },
}
