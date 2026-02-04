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
          -- ヘッダーカラーの設定 (Expanded Abyssal Teal)
          local colors = {
            h1 = "#9DDCD9", -- Heading Cyan
            h2 = "#CEF5F2", -- Main Text
            h3 = "#64BBBE", -- Clear Teal
            h4 = "#A4ABCB", -- Sky Slate
            h5 = "#B4B7CD", -- Cloud Slate
            h6 = "#525B65", -- Slate Mid
          }

          for i = 1, 6 do
            local bg_name = "RenderMarkdownH" .. i .. "Bg"
            local fg_name = "RenderMarkdownH" .. i
            vim.api.nvim_set_hl(0, bg_name, {
              bg = colors["h" .. i],
              fg = "#0B0C0C",  -- Background (Pure Black)
              bold = true
            })
            vim.api.nvim_set_hl(0, fg_name, {
              fg = colors["h" .. i],
              bold = true
            })
          end

          -- その他のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#152A2B" })  -- Dark Teal Panel
          vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#152A2B", fg = "#B1F4ED" })
          vim.api.nvim_set_hl(0, "RenderMarkdownLink", { fg = "#6CD8D3", underline = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownQuote", { fg = "#525B65", italic = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = "#349594" }) -- Deep Sea Teal
          vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = "#275D62" }) -- UI Border
          vim.api.nvim_set_hl(0, "RenderMarkdownTodo", { fg = "#936997" }) -- Glitch Purple
          vim.api.nvim_set_hl(0, "RenderMarkdownImportant", { fg = "#CED5E9" }) -- Lavender

          -- テーブル関連のハイライト設定
          vim.api.nvim_set_hl(0, "RenderMarkdownTableHead", { fg = "#9DDCD9", bold = true })
          vim.api.nvim_set_hl(0, "RenderMarkdownTableRow", { fg = "#CEF5F2" })
          vim.api.nvim_set_hl(0, "RenderMarkdownTableFill", { fg = "#275D62" })  -- ボーダー
        end,
      })

      -- 初期化時にもカラー設定を適用
      vim.schedule(function()
        vim.cmd("doautocmd ColorScheme")
      end)
    end,
  },
}
