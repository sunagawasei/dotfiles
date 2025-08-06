return {
  {
    "tiesen243/vercel.nvim",
    lazy = false,
    priority = 1000,
    config = function()
      require("vercel").setup({
        theme = "dark", -- dark テーマを設定
        transparent = true,
        italics = {
          comments = false,
          keywords = false,
          functions = false,
          strings = false,
          variables = false,
        },
        overrides = {
          -- IMEやフローティングウィンドウの設定
          NormalFloat = { bg = "#1a1a1a", fg = "#ededed" }, -- フローティングウィンドウの背景と文字色（視認性向上）
          FloatBorder = { bg = "#1a1a1a", fg = "#0070f3" }, -- フローティングウィンドウの境界線（Vercel Blue）
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
          -- Snacks.nvim Picker/Explorer の透明化
          SnacksPickerNormal = { bg = "none" },
          SnacksPickerBorder = { bg = "none" },
          SnacksPickerNormalFloat = { bg = "none" },
          SnacksPickerFile = { bg = "none" },
          SnacksPickerDir = { bg = "none" },
          SnacksPickerPathHidden = { bg = "none" },
          SnacksPickerBox = { bg = "none" },
          SnacksPickerPrompt = { bg = "none" },
          SnacksPickerMatch = { bg = "none" },
          SnacksPickerList = { bg = "none" },
          SnacksPickerListCursorLine = { bg = "none" },
          SnacksPickerPathIgnored = { bg = "none" },
          -- 一般的なサイドバー関連のハイライトグループ
          NeoTreeNormal = { bg = "none" },
          NeoTreeNormalNC = { bg = "none" },
          NvimTreeNormal = { bg = "none" },
          NvimTreeNormalNC = { bg = "none" },
          
          -- SnacksPickerTreeがLineNrにリンクされているため
          SnacksPickerTree = { bg = "none" },
          LineNr = { bg = "none" },
        },
      })
      -- setup()の後にcolorschemeを設定する必要がある
      vim.cmd.colorscheme("vercel")
      
      -- 包括的な透明化設定
      local function set_transparent_bg()
        -- 基本的な背景
        vim.api.nvim_set_hl(0, "Normal", { bg = "none" })
        vim.api.nvim_set_hl(0, "NormalNC", { bg = "none" })
        -- NormalFloatは透明化しない（ホバーウィンドウの視認性のため）
        -- vim.api.nvim_set_hl(0, "NormalFloat", { bg = "none" })
        vim.api.nvim_set_hl(0, "EndOfBuffer", { bg = "none" })
        vim.api.nvim_set_hl(0, "SignColumn", { bg = "none" })
        vim.api.nvim_set_hl(0, "VertSplit", { bg = "none" })
        vim.api.nvim_set_hl(0, "WinSeparator", { bg = "none" })
        
        -- ステータスライン・タブライン
        vim.api.nvim_set_hl(0, "StatusLine", { bg = "none" })
        vim.api.nvim_set_hl(0, "StatusLineNC", { bg = "none" })
        vim.api.nvim_set_hl(0, "TabLine", { bg = "none" })
        vim.api.nvim_set_hl(0, "TabLineFill", { bg = "none" })
        
        -- 行番号
        vim.api.nvim_set_hl(0, "LineNr", { bg = "none" })
        vim.api.nvim_set_hl(0, "CursorLineNr", { bg = "none" })
        
        -- Snacks関連のすべて
        local snacks_groups = vim.fn.getcompletion("Snacks", "highlight")
        for _, group in ipairs(snacks_groups) do
          vim.api.nvim_set_hl(0, group, { bg = "none" })
        end
      end
      
      -- 初回実行
      set_transparent_bg()
      
      -- ColorSchemeイベントでも実行
      vim.api.nvim_create_autocmd("ColorScheme", {
        pattern = "*",
        callback = set_transparent_bg,
      })
    end,
  },
}

