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

      -- Vercel Geistカラーテーマとの統合
      vim.api.nvim_create_autocmd("ColorScheme", {
        callback = function()
          -- ヘッダーカラーの設定
          local colors = {
            h1 = "#0070f3", -- Vercel Blue
            h2 = "#7928ca", -- Vercel Purple
            h3 = "#ff0080", -- Vercel Pink
            h4 = "#f81ce5", -- Vercel Magenta
            h5 = "#eb367f", -- Vercel Light Pink
            h6 = "#8b5cf6", -- Vercel Light Purple
          }

          for i = 1, 6 do
            local bg_name = "RenderMarkdownH" .. i .. "Bg"
            local fg_name = "RenderMarkdownH" .. i
            vim.api.nvim_set_hl(0, bg_name, {
              bg = colors["h" .. i],
              fg = "#ffffff",
              bold = true
            })
            vim.api.nvim_set_hl(0, fg_name, {
              fg = colors["h" .. i],
              bold = true
            })
          end

          -- その他のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#1a1a1a" })
          vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#2a2a2a", fg = "#e1e1e1" })
          vim.api.nvim_set_hl(0, "RenderMarkdownLink", { fg = "#0070f3", underline = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownQuote", { fg = "#888888", italic = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = "#00d9ff" })
          vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = "#666666" })
          vim.api.nvim_set_hl(0, "RenderMarkdownTodo", { fg = "#ffa500" })
          vim.api.nvim_set_hl(0, "RenderMarkdownImportant", { fg = "#ff6b6b" })
        end,
      })

      -- 初期化時にもカラー設定を適用
      vim.schedule(function()
        vim.cmd("doautocmd ColorScheme")
      end)
    end,
  },
}
