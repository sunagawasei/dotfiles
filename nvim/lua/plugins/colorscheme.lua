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
          
          -- Blink.cmp 用のハイライトグループ
          BlinkCmpMenu = { bg = "#1a1a1a", fg = "#ededed" },
          BlinkCmpMenuBorder = { bg = "#1a1a1a", fg = "#0070f3" },
          BlinkCmpMenuSelection = { bg = "#1E40AF", fg = "#ffffff" },
          BlinkCmpDoc = { bg = "#1a1a1a", fg = "#ededed" },
          BlinkCmpDocBorder = { bg = "#1a1a1a", fg = "#0070f3" },
          BlinkCmpLabel = { bg = "#1a1a1a", fg = "#ededed" },
          BlinkCmpLabelMatch = { bg = "#1a1a1a", fg = "#0070f3" },
          BlinkCmpKind = { bg = "#1a1a1a", fg = "#C586C0" },
          BlinkCmpSource = { bg = "#1a1a1a", fg = "#C792EA" },
          -- Snacks.nvim Picker/Explorer の透明化
          SnacksPickerNormal = { bg = "none" },
          SnacksPickerBorder = { bg = "none" },
          SnacksPickerNormalFloat = { bg = "none" },
          SnacksPickerFile = { bg = "none" },
          SnacksPickerDir = { bg = "none" },
          SnacksPickerPathHidden = { bg = "none" },
          SnacksPickerBox = { bg = "none" },
          SnacksPickerPrompt = { bg = "none" },
          SnacksPickerMatch = { bg = "none", fg = "#0070f3" },
          SnacksPickerList = { bg = "none" },
          SnacksPickerListCursorLine = { bg = "#1E40AF", fg = "#ffffff" },
          SnacksPickerSelection = { bg = "#1E40AF", fg = "#ffffff", bold = true },
          SnacksPickerPathIgnored = { bg = "none" },
          -- 一般的なサイドバー関連のハイライトグループ
          NeoTreeNormal = { bg = "none" },
          NeoTreeNormalNC = { bg = "none" },
          NvimTreeNormal = { bg = "none" },
          NvimTreeNormalNC = { bg = "none" },
          
          -- SnacksPickerTreeがLineNrにリンクされているため
          SnacksPickerTree = { bg = "none" },
          LineNr = { bg = "none" },
          
          -- ClaudeCode関連のハイライトグループ
          ClaudeCodeNormal = { bg = "none" },
          ClaudeCodeNormalFloat = { bg = "none" },
          ClaudeCodeFloatBorder = { bg = "none" },
          ClaudeCodeWinBar = { bg = "none" },
          ClaudeCodeWinBarNC = { bg = "none" },
          ClaudeCodeStatusLine = { bg = "none" },
          ClaudeCodeStatusLineNC = { bg = "none" },
          
          -- MiniIcons ハイライトグループ - Vercel Geist カラー
          MiniIconsAzure = { fg = "#3B82F6" },   -- Blue 6 - 関数、メソッド
          MiniIconsBlue = { fg = "#2563EB" },    -- Blue 7 - Lua、設定ファイル
          MiniIconsCyan = { fg = "#14B8A6" },    -- Teal 6 - ライセンス、特殊ファイル
          MiniIconsGreen = { fg = "#22C55E" },   -- Green 6 - 成功、初期化ファイル
          MiniIconsGrey = { fg = "#64748B" },    -- Gray 6 - システムファイル、Ansible
          MiniIconsOrange = { fg = "#F59E0B" },  -- Amber 6 - 警告、変更ログ
          MiniIconsPurple = { fg = "#A855F7" },  -- Purple 6 - Git設定、YAML
          MiniIconsRed = { fg = "#EF4444" },     -- Red 6 - エラー、重要な設定
          MiniIconsYellow = { fg = "#FBBF24" },  -- Amber 5 - README、Docker
        },
      })
      -- setup()の後にcolorschemeを設定する必要がある
      vim.cmd.colorscheme("vercel")
      
      -- ターミナルカラー設定（Vercel Geistカラーシステム準拠）
      vim.g.terminal_color_0 = "#0F172A"   -- Gray 10 (黒系)
      vim.g.terminal_color_1 = "#EF4444"   -- Red 6 (赤)
      vim.g.terminal_color_2 = "#22C55E"   -- Green 6 (緑)
      vim.g.terminal_color_3 = "#F59E0B"   -- Amber 6 (黄)
      vim.g.terminal_color_4 = "#3B82F6"   -- Blue 6 (青)
      vim.g.terminal_color_5 = "#A855F7"   -- Purple 6 (マゼンタ)
      vim.g.terminal_color_6 = "#14B8A6"   -- Teal 6 (シアン)
      vim.g.terminal_color_7 = "#E2E8F0"   -- Gray 3 (明るいグレー - 補完メニューの文字色)
      vim.g.terminal_color_8 = "#64748B"   -- Gray 6 (暗いグレー)
      vim.g.terminal_color_9 = "#DC2626"   -- Red 7 (明るい赤)
      vim.g.terminal_color_10 = "#16A34A"  -- Green 7 (明るい緑)
      vim.g.terminal_color_11 = "#D97706"  -- Amber 7 (明るい黄)
      vim.g.terminal_color_12 = "#2563EB"  -- Blue 7 (明るい青)
      vim.g.terminal_color_13 = "#9333EA"  -- Purple 7 (明るいマゼンタ)
      vim.g.terminal_color_14 = "#0D9488"  -- Teal 7 (明るいシアン)
      vim.g.terminal_color_15 = "#F8FAFC"  -- Gray 1 (白系)
      
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
        
        -- Snacks関連のハイライトグループ（選択項目は除外）
        local snacks_groups = vim.fn.getcompletion("Snacks", "highlight")
        for _, group in ipairs(snacks_groups) do
          -- 選択項目関連のハイライトは透明化しない
          if not string.match(group, "CursorLine") and 
             not string.match(group, "Selection") and 
             not string.match(group, "Cursor") then
            vim.api.nvim_set_hl(0, group, { bg = "none" })
          end
        end
        
        -- Snacks Picker選択項目のハイライトを明示的に設定
        vim.api.nvim_set_hl(0, "SnacksPickerListCursorLine", { bg = "#1E40AF", fg = "#ffffff" })
        vim.api.nvim_set_hl(0, "SnacksPickerSelection", { bg = "#1E40AF", fg = "#ffffff", bold = true })
        vim.api.nvim_set_hl(0, "SnacksPickerCursor", { bg = "#1E40AF", fg = "#ffffff" })
        vim.api.nvim_set_hl(0, "SnacksPickerCursorLine", { bg = "#1E40AF", fg = "#ffffff" })
        
        -- ClaudeCode関連のハイライトグループ
        local claudecode_groups = vim.fn.getcompletion("ClaudeCode", "highlight")
        for _, group in ipairs(claudecode_groups) do
          vim.api.nvim_set_hl(0, group, { bg = "none" })
        end
        
        -- 追加の可能性があるClaudeCode関連グループ
        vim.api.nvim_set_hl(0, "ClaudeCodeNormal", { bg = "none" })
        vim.api.nvim_set_hl(0, "ClaudeCodeNormalFloat", { bg = "none" })
        vim.api.nvim_set_hl(0, "ClaudeCodeFloatBorder", { bg = "none" })
        vim.api.nvim_set_hl(0, "ClaudeCodeWinBar", { bg = "none" })
        vim.api.nvim_set_hl(0, "ClaudeCodeWinBarNC", { bg = "none" })
        
        -- MiniIconsのハイライトグループは背景透明化しない（アイコンの色を保持）
        -- これらのグループは前景色のみ設定し、背景は透明のままにする
      end
      
      -- 初回実行
      set_transparent_bg()
      
      -- 複数のイベントで透明化を実行
      vim.api.nvim_create_autocmd({
        "ColorScheme",
        "VimEnter",
        "UIEnter",
        "BufWinEnter",
      }, {
        pattern = "*",
        callback = function()
          -- 少し遅延させて確実に適用
          vim.defer_fn(set_transparent_bg, 10)
        end,
      })
      
      -- ClaudeCodeウィンドウが開いた時にも適用
      vim.api.nvim_create_autocmd("User", {
        pattern = "ClaudeCode*",
        callback = function()
          vim.defer_fn(set_transparent_bg, 50)
        end,
      })
    end,
  },
}

