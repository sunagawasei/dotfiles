return {
  {
    "MeanderingProgrammer/render-markdown.nvim",
    enabled = false, -- Markdownレンダリングを無効化
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
        code = {
          "RenderMarkdownCode",
          "RenderMarkdownCodeInline",
        },
      },
      -- 見出しのスタイル設定
      heading = {
        -- アイコンとパディングを追加
        icons = { "◉ ", "● ", "○ ", "◆ ", "◇ ", "▸ " },
        signs = { "󰌕" },
        width = "full",
        left_pad = 1,
        right_pad = 1,
        -- アンダーラインを追加（H1とH2のみ）
        border = true,
        border_virtual = true,
        above = "▔",
        below = "▁",
      },
      -- コードブロック設定
      code = {
        -- エンブレムを有効化
        enabled = true,
        -- アイコンとハイライト
        sign = true,
        -- スタイル設定（言語タブ付き）
        style = "language",
        -- ボーダー設定（厚めのボーダー）
        border = "thick",
        -- 位置設定
        position = "left",
        -- パディング設定
        left_margin = 0,
        left_pad = 2,
        right_pad = 2,
        -- 幅設定
        width = "full",
        -- 言語表示設定
        language_pad = 2,
        language_name = true,
        -- デリミタの非表示化
        conceal_delimiters = false,
        -- 最小幅
        min_width = 0,
        -- ボーダー文字
        above = "▄",
        below = "▀",
        -- ハイライトグループを明示的に指定
        highlight = "RenderMarkdownCode",
        highlight_inline = "RenderMarkdownCodeInline",
        -- アイコン設定
        icons = {
          -- 言語別アイコン
          bash = "",
          c = "",
          cpp = "",
          css = "",
          go = "",
          html = "",
          java = "",
          javascript = "",
          json = "",
          lua = "",
          markdown = "",
          python = "",
          rust = "",
          typescript = "",
          vim = "",
          yaml = "",
          -- デフォルトアイコン
          default = "",
        },
      },
    },
    ft = { "markdown", "copilot-chat" },
    config = function(_, opts)
      require("render-markdown").setup(opts)
      
      -- Vercel Geist カラーでハイライトグループを設定
      -- 見出しの色（階層的に鮮やかさを調整）
      vim.api.nvim_set_hl(0, "RenderMarkdownH1", { fg = "#3291FF", bold = true }) -- Geist Blue 7 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH1Bg", { bg = "#1A1A1A" }) -- Color 2 (Dark) - より明るい背景
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH2", { fg = "#A855F7", bold = true }) -- Geist Purple 6 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH2Bg", { bg = "#262626" }) -- Color 3 (Dark) - 段階的に明るく
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH3", { fg = "#14B8A6", bold = true }) -- Geist Teal 6 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH3Bg", { bg = "#1A1A1A" }) -- Color 2 (Dark)
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH4", { fg = "#FFCC33", bold = true }) -- Geist Amber 6 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH4Bg", { bg = "#262626" }) -- Color 3 (Dark)
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH5", { fg = "#50E3C2", bold = true }) -- Geist Green 6 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH5Bg", { bg = "#1A1A1A" }) -- Color 2 (Dark)
      
      vim.api.nvim_set_hl(0, "RenderMarkdownH6", { fg = "#F472B6", bold = true }) -- Geist Pink 5 (より明るく)
      vim.api.nvim_set_hl(0, "RenderMarkdownH6Bg", { bg = "#262626" }) -- Color 3 (Dark)
      
      -- コードブロックの色（nocombineで強制適用）
      vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- Color 3 (Dark) + Color 7 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeBlock", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- コードブロック全体
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeInline", { bg = "#333333", fg = "#A3A3A3", nocombine = true }) -- Color 4 (Dark) + Color 9 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeBorder", { fg = "#666666", bg = "#262626", nocombine = true }) -- Color 6 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeInfo", { fg = "#A3A3A3", bg = "#333333", nocombine = true }) -- Color 9 (Dark)
      vim.api.nvim_set_hl(0, "RenderMarkdownCodeLang", { fg = "#66C2FF", bg = "#333333", bold = true, nocombine = true }) -- 言語表示（Geist Blue 6）
      
      -- 追加のコードブロック関連ハイライト（nocombineで強制適用）
      vim.api.nvim_set_hl(0, "CodeBlock", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- 一般的なコードブロック
      vim.api.nvim_set_hl(0, "@markup.raw.block.markdown", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- Treesitter コードブロック
      vim.api.nvim_set_hl(0, "@markup.raw.markdown_inline", { bg = "#333333", fg = "#A3A3A3", nocombine = true }) -- Treesitter インラインコード
      vim.api.nvim_set_hl(0, "@text.literal", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- Treesitter リテラルテキスト
      vim.api.nvim_set_hl(0, "@markup.raw", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- Treesitter 生テキスト
      vim.api.nvim_set_hl(0, "RenderMarkdownCodePre", { bg = "#262626", fg = "#EDEDED", nocombine = true }) -- プリフォーマットテキスト
      
      -- render-markdown.nvimが使用する可能性のある他のハイライトグループ
      vim.api.nvim_set_hl(0, "markdownCode", { bg = "#262626", fg = "#EDEDED", nocombine = true })
      vim.api.nvim_set_hl(0, "markdownCodeBlock", { bg = "#262626", fg = "#EDEDED", nocombine = true })
      vim.api.nvim_set_hl(0, "markdownCodeDelimiter", { fg = "#666666", nocombine = true })
      
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
      
      -- FileTypeとBufEnterでもハイライトを適用
      local function apply_markdown_highlights()
        -- コードブロックのハイライトを再適用
        vim.api.nvim_set_hl(0, "RenderMarkdownCode", { bg = "#262626", fg = "#EDEDED", nocombine = true })
        vim.api.nvim_set_hl(0, "RenderMarkdownCodeBlock", { bg = "#262626", fg = "#EDEDED", nocombine = true })
        vim.api.nvim_set_hl(0, "@markup.raw.block.markdown", { bg = "#262626", fg = "#EDEDED", nocombine = true })
        vim.api.nvim_set_hl(0, "@text.literal", { bg = "#262626", fg = "#EDEDED", nocombine = true })
        vim.api.nvim_set_hl(0, "@markup.raw", { bg = "#262626", fg = "#EDEDED", nocombine = true })
      end
      
      vim.api.nvim_create_autocmd({ "FileType", "BufEnter" }, {
        pattern = { "markdown", "copilot-chat" },
        callback = apply_markdown_highlights,
        group = vim.api.nvim_create_augroup("RenderMarkdownCodeHighlights", { clear = true }),
      })
    end,
  },
}