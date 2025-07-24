return {
  {
    "tiesen243/vercel.nvim",
    lazy = false,
    priority = 1000,
    config = function()
      require("vercel").setup({
        theme = "dark", -- dark テーマを設定
        transparent = false,
        italics = {
          comments = false,
          keywords = false,
          functions = false,
          strings = false,
          variables = false,
        },
        overrides = {
          -- IMEやフローティングウィンドウの設定
          NormalFloat = { bg = "#0a0a0a", fg = "#ededed" }, -- フローティングウィンドウの背景と文字色
          FloatBorder = { bg = "#0a0a0a", fg = "#444444" }, -- フローティングウィンドウの境界線
          Pmenu = { bg = "#1a1a1a", fg = "#ededed" }, -- ポップアップメニューの背景と文字色
          PmenuSel = { bg = "#2a2a2a", fg = "#ffffff" }, -- ポップアップメニューの選択項目
          PmenuSbar = { bg = "#1a1a1a" }, -- スクロールバーの背景
          PmenuThumb = { bg = "#3a3a3a" }, -- スクロールバーのつまみ
          -- CJK IME support
          CmpItemAbbrMatch = { bg = "#1a1a1a", fg = "#569CD6" },
          CmpItemAbbrMatchFuzzy = { bg = "#1a1a1a", fg = "#569CD6" },
          CmpItemAbbr = { bg = "#1a1a1a", fg = "#ededed" },
          CmpItemKind = { bg = "#1a1a1a", fg = "#C586C0" },
          CmpItemMenu = { bg = "#1a1a1a", fg = "#C792EA" },
          CmpItemKindFunction = { bg = "#1a1a1a", fg = "#C586C0" },
          CmpItemKindMethod = { bg = "#1a1a1a", fg = "#C586C0" },
          CmpItemKindConstructor = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindClass = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindEnum = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindEvent = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindInterface = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindStruct = { bg = "#1a1a1a", fg = "#ffbb00" },
          CmpItemKindVariable = { bg = "#1a1a1a", fg = "#9CDCFE" },
          CmpItemKindField = { bg = "#1a1a1a", fg = "#9CDCFE" },
          CmpItemKindProperty = { bg = "#1a1a1a", fg = "#9CDCFE" },
          CmpItemKindEnumMember = { bg = "#1a1a1a", fg = "#9CDCFE" },
          CmpItemKindConstant = { bg = "#1a1a1a", fg = "#9CDCFE" },
          CmpItemKindKeyword = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindModule = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindValue = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindUnit = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindText = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindSnippet = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindFile = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindFolder = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindColor = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindReference = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindOperator = { bg = "#1a1a1a", fg = "#D4D4D4" },
          CmpItemKindTypeParameter = { bg = "#1a1a1a", fg = "#D4D4D4" },
        }
      })
      -- setup()の後にcolorschemeを設定する必要がある
      vim.cmd.colorscheme("vercel")
    end,
  },
}
