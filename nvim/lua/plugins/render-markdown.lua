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
      local function apply_colors()
        local p = require("config.palette").colors

        -- ヘッダーカラーの設定 (Expanded Abyssal Teal)
        local h_colors = {
          h1 = p.bright_cyan,    -- Heading Cyan
          h2 = p.fg,             -- Main Text
          h3 = p.operator,       -- Clear Teal
          h4 = p.white,          -- Sky Slate
          h5 = p.bright_magenta, -- Cloud Slate
          h6 = p.mid_gray,       -- Slate Mid
        }

        for i = 1, 6 do
          local bg_name = "RenderMarkdownH" .. i .. "Bg"
          local fg_name = "RenderMarkdownH" .. i
          vim.api.nvim_set_hl(0, bg_name, {
            bg = h_colors["h" .. i],
            fg = p.bg,   -- サーフェス統一: panel_bg背景上のテキストは bg で contrast確保
            bold = true,
          })
          vim.api.nvim_set_hl(0, fg_name, {
            fg = h_colors["h" .. i],
            bold = true,
          })
        end

        -- サーフェス色の統一 (Step 10): code block は panel_bg で一貫
        vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = p.panel_bg })
        vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = p.panel_bg, fg = p.near_white })
        vim.api.nvim_set_hl(0, "RenderMarkdownLink", { fg = p.cyan, underline = true })
        vim.api.nvim_set_hl(0, "RenderMarkdownQuote", { fg = p.light_gray, italic = true })
        vim.api.nvim_set_hl(0, "RenderMarkdownChecked", { fg = p.ansi_green })
        vim.api.nvim_set_hl(0, "RenderMarkdownUnchecked", { fg = p.light_gray })
        vim.api.nvim_set_hl(0, "RenderMarkdownTodo", { fg = p.magenta })
        vim.api.nvim_set_hl(0, "RenderMarkdownImportant", { fg = p.lavender })

        -- テーブル関連のハイライト設定
        vim.api.nvim_set_hl(0, "RenderMarkdownTableHead", { fg = p.bright_cyan, bold = true })
        vim.api.nvim_set_hl(0, "RenderMarkdownTableRow", { fg = p.fg })
        vim.api.nvim_set_hl(0, "RenderMarkdownTableFill", { fg = p.border })
      end

      vim.api.nvim_create_autocmd("ColorScheme", {
        callback = apply_colors,
      })

      -- 初期化時にもカラー設定を適用
      vim.schedule(function()
        vim.cmd("doautocmd ColorScheme")
      end)
    end,
  },
}
