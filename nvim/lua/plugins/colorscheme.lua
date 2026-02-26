return {
  -- LazyVimのカラースキーム自動適用を無効化（プラグインは維持）
  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = function() end, -- 何もしない関数で上書き
    },
  },
  {
    dir = vim.fn.stdpath("config"),
    name = "custom-colorscheme",
    -- 独自カラースキーム（プラグイン依存なし）
    lazy = false,
    priority = 1000,
    config = function()
      -- カラーパレット定義 — TOMLソース: colors/abyssal-teal.toml
      local colors = require("config.palette").colors

      -- ハイライトグループの設定
      local highlights = {
        -- 基本UI
        Normal = { bg = colors.bg, fg = colors.fg },
        NormalNC = { bg = colors.bg, fg = colors.fg },
        EndOfBuffer = { bg = colors.bg },
        SignColumn = { bg = "none" },
        VertSplit = { bg = colors.bg },
        WinSeparator = { bg = colors.bg },

        -- 基本UI - 追加分
        Directory = { fg = colors.cyan },
        Title = { fg = colors.bright_cyan, bold = true },
        Special = { fg = colors.mid_gray },
        Identifier = { fg = colors.fg },
        Statement = { fg = colors.purple_accent },
        PreProc = { fg = colors.cyan },
        Type = { fg = colors.white },
        Constant = { fg = colors.lavender },
        String = { fg = colors.string },
        Number = { fg = colors.near_white },
        Boolean = { fg = colors.near_white },
        Function = { fg = colors.cyan },
        Keyword = { fg = colors.purple_accent },
        Operator = { fg = colors.operator },
        Comment = { fg = colors.comment_gray },  -- より明るく読みやすく

        -- IMEやフローティングウィンドウの設定
        NormalFloat = { bg = colors.dark_shadow, fg = colors.fg },
        FloatBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        Pmenu = { bg = colors.dark_shadow, fg = colors.fg },
        PmenuSel = { bg = colors.dark_shadow, fg = colors.highlight_white },
        PmenuSbar = { bg = colors.dark_shadow },
        PmenuThumb = { bg = colors.mid_gray },

        -- CJK IME support
        CmpItemAbbrMatch = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemAbbrMatchFuzzy = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemAbbr = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKind = { bg = colors.dark_shadow, fg = colors.mid_gray },
        CmpItemMenu = { bg = colors.dark_shadow, fg = colors.mid_gray },
        CmpItemKindFunction = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindMethod = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindConstructor = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindClass = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindEnum = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindEvent = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindInterface = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindStruct = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindVariable = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindField = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindProperty = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindEnumMember = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindConstant = { bg = colors.dark_shadow, fg = colors.bright_cyan },
        CmpItemKindKeyword = { bg = colors.dark_shadow, fg = colors.cyan },
        CmpItemKindModule = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindValue = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindUnit = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindText = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindSnippet = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindFile = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindFolder = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindColor = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindReference = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindOperator = { bg = colors.dark_shadow, fg = colors.fg },
        CmpItemKindTypeParameter = { bg = colors.dark_shadow, fg = colors.fg },

        -- Blink.cmp 用のハイライトグループ
        BlinkCmpMenu = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpMenuBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpMenuSelection = { bg = colors.dark_shadow, fg = colors.highlight_white },
        BlinkCmpDoc = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpDocBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpLabel = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpLabelMatch = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpKind = { bg = colors.dark_shadow, fg = colors.mid_gray },
        BlinkCmpSource = { bg = colors.dark_shadow, fg = colors.mid_gray },

        -- Snacks.nvim Picker/Explorer の透明化
        SnacksPickerNormal = { bg = "none" },
        SnacksPickerBorder = { bg = "none" },
        SnacksPickerNormalFloat = { bg = "none" },
        SnacksPickerFile = { bg = "none" },
        SnacksPickerDir = { bg = "none" },
        SnacksPickerPathHidden = { bg = "none" },
        SnacksPickerBox = { bg = "none" },
        SnacksPickerPrompt = { bg = "none" },
        SnacksPickerMatch = { bg = "none", fg = colors.cyan },
        SnacksPickerList = { bg = "none" },
        SnacksPickerListCursorLine = { bg = colors.dark_shadow, fg = colors.highlight_white },
        SnacksPickerSelection = { bg = colors.dark_shadow, fg = colors.highlight_white, bold = true },
        SnacksPickerPathIgnored = { bg = "none" },

        -- 一般的なサイドバー関連のハイライトグループ
        NeoTreeNormal = { bg = "none" },
        NeoTreeNormalNC = { bg = "none" },
        NvimTreeNormal = { bg = "none" },
        NvimTreeNormalNC = { bg = "none" },
        SnacksPickerTree = { bg = "none" },
        LineNr = { bg = "none" },
        CursorLineNr = { bg = "none", fg = colors.highlight_white, bold = true },

        -- ClaudeCode関連のハイライトグループ
        ClaudeCodeNormal = { bg = "none" },
        ClaudeCodeNormalFloat = { bg = "none" },
        ClaudeCodeFloatBorder = { bg = "none" },
        ClaudeCodeWinBar = { bg = "none" },
        ClaudeCodeWinBarNC = { bg = "none" },
        ClaudeCodeStatusLine = { bg = "none" },
        ClaudeCodeStatusLineNC = { bg = "none" },

        -- MiniIcons ハイライトグループ - 白色統一表示
        MiniIconsAzure = { fg = colors.white },
        MiniIconsBlue = { fg = colors.white },
        MiniIconsCyan = { fg = colors.white },
        MiniIconsGreen = { fg = colors.white },
        MiniIconsGrey = { fg = colors.white },
        MiniIconsOrange = { fg = colors.white },
        MiniIconsPurple = { fg = colors.white },
        MiniIconsRed = { fg = colors.white },
        MiniIconsYellow = { fg = colors.white },

        -- GitSigns ハイライトグループ - モノクロ基調＋アクセント
        GitSignsAdd = { fg = colors.success },
        GitSignsChange = { fg = colors.magenta },
        GitSignsDelete = { fg = colors.light_gray },
        GitSignsAddNr = { fg = colors.success },
        GitSignsChangeNr = { fg = colors.magenta },
        GitSignsDeleteNr = { fg = colors.light_gray },
        GitSignsCurrentLineBlame = { fg = colors.git_blame_gray },  -- コメントより明るく

        -- Diff関連のハイライトグループ
        DiffAdd = { fg = colors.cyan, bg = "#0D1F1F" },
        DiffChange = { fg = colors.magenta, bg = "#1A141A" },
        DiffDelete = { fg = colors.light_gray, bg = "#151515" },
        DiffText = { fg = colors.bright_cyan, bold = true },

        -- 構文ハイライト（Cyber Glitch Teal - ネオン系配色）
        ["@keyword"] = { fg = colors.purple_accent },
        ["@keyword.function"] = { fg = colors.purple_accent },
        ["@keyword.operator"] = { fg = colors.purple_accent },
        ["@keyword.return"] = { fg = colors.purple_accent },
        ["@keyword.import"] = { fg = colors.cyan },
        ["@string"] = { fg = colors.string },
        ["@number"] = { fg = colors.fg },
        ["@boolean"] = { fg = colors.fg },
        ["@comment"] = { fg = colors.comment_gray },  -- Commentと統一
        ["@function"] = { fg = colors.bright_cyan },
        ["@function.call"] = { fg = colors.bright_cyan },
        ["@function.method"] = { fg = colors.bright_cyan },
        ["@function.method.call"] = { fg = colors.bright_cyan },
        ["@variable"] = { fg = colors.fg },
        ["@variable.builtin"] = { fg = colors.bright_cyan },
        ["@parameter"] = { fg = colors.fg },
        ["@type"] = { fg = colors.bright_cyan },
        ["@type.builtin"] = { fg = colors.bright_cyan },
        ["@property"] = { fg = colors.fg },
        ["@field"] = { fg = colors.fg },
        ["@constant"] = { fg = colors.fg },
        ["@constant.builtin"] = { fg = colors.fg },
        ["@operator"] = { fg = colors.operator },
        ["@punctuation"] = { fg = colors.punctuation_gray },
        ["@punctuation.bracket"] = { fg = colors.punctuation_gray },
        ["@punctuation.delimiter"] = { fg = colors.punctuation_gray },
        ["@punctuation.delimiter.yaml"] = { fg = colors.light_gray }, -- YAML list marker visibility
        ["@namespace"] = { fg = colors.fg },
        ["@module"] = { fg = colors.fg },
        ["@tag"] = { fg = colors.cyan },
        ["@tag.attribute"] = { fg = colors.bright_cyan },
        ["@tag.delimiter"] = { fg = colors.mid_gray },

        -- Treesitter Markup（Markdown用）
        ["@markup.heading"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.1"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.2"] = { fg = colors.near_white, bold = true },
        ["@markup.heading.3"] = { fg = colors.fg, bold = true },
        ["@markup.heading.4"] = { fg = colors.operator, bold = true },
        ["@markup.heading.5"] = { fg = colors.light_gray, bold = true },
        ["@markup.heading.6"] = { fg = colors.mid_gray, bold = true },
        ["@markup.list"] = { fg = colors.light_gray },
        ["@markup.list.markdown"] = { fg = colors.light_gray },
        ["@markup.list.checked"] = { fg = colors.cyan },
        ["@markup.list.unchecked"] = { fg = colors.light_gray },
        ["@markup.link"] = { fg = colors.cyan, underline = true },
        ["@markup.link.label"] = { fg = colors.cyan },
        ["@markup.link.url"] = { fg = colors.mid_gray, underline = true },
        ["@markup.raw"] = { fg = colors.fg },
        ["@markup.raw.markdown_inline"] = { bg = colors.dark_shadow, fg = colors.fg },
        ["@markup.raw.block"] = { fg = colors.fg },
        ["@markup.strong"] = { fg = colors.highlight_white, bold = true },
        ["@markup.italic"] = { fg = colors.fg, italic = true },
        ["@markup.strikethrough"] = { fg = colors.mid_gray, strikethrough = true },
        ["@markup.quote"] = { fg = colors.light_gray, italic = true },
        ["@punctuation.special.markdown"] = { fg = colors.light_gray },

        -- Oil.nvim固有グループ
        OilDir = { fg = colors.cyan },
        OilDirIcon = { fg = colors.cyan },
        OilLink = { fg = colors.magenta },
        OilFile = { fg = colors.fg },

        -- TreesitterContext (スティッキースクロール) ハイライトグループ
        TreesitterContext = { bg = colors.dark_shadow, fg = colors.fg },
        TreesitterContextBottom = { underline = true, sp = colors.dark_shadow },
        TreesitterContextLineNumber = { bg = colors.dark_shadow, fg = colors.mid_gray },

        -- 診断関連（bufferline、LSP診断などが参照）
        DiagnosticError = { fg = colors.magenta },
        DiagnosticWarn = { fg = colors.bright_magenta },
        DiagnosticInfo = { fg = colors.bright_cyan },
        DiagnosticHint = { fg = colors.cyan },
        WarningMsg = { fg = colors.bright_magenta },
        ErrorMsg = { fg = colors.magenta },

        -- 検索・マッチ関連
        Search = { bg = colors.dark_shadow, fg = colors.cyan },
        IncSearch = { bg = colors.cyan, fg = colors.bg },
        CurSearch = { bg = colors.bright_cyan, fg = colors.bg },
        Substitute = { bg = colors.magenta, fg = colors.bg },
        MatchParen = { bg = colors.border, fg = colors.highlight_white, bold = true },

        -- カーソル・選択関連
        Visual = { bg = colors.selection, fg = colors.selection_fg },
        VisualNOS = { bg = colors.selection, fg = colors.selection_fg },
        Select = { bg = colors.selection, fg = colors.selection_fg },
        CursorLine = { bg = colors.dark_shadow },
        CursorColumn = { bg = colors.dark_shadow },
        Cursor = { bg = colors.fg, fg = colors.bg },
        lCursor = { bg = colors.fg, fg = colors.bg },
        CursorIM = { bg = colors.cyan, fg = colors.bg },
        TermCursor = { bg = colors.fg },
        TermCursorNC = { bg = colors.mid_gray },
        ColorColumn = { bg = colors.dark_shadow },

        -- LSP診断拡張（VirtualText、Underline、Sign、Floating）
        DiagnosticVirtualTextError = { fg = colors.magenta },
        DiagnosticVirtualTextWarn = { fg = colors.bright_magenta },
        DiagnosticVirtualTextInfo = { fg = colors.bright_cyan },
        DiagnosticVirtualTextHint = { fg = colors.cyan },
        DiagnosticUnderlineError = { sp = colors.magenta, undercurl = true },
        DiagnosticUnderlineWarn = { sp = colors.bright_magenta, undercurl = true },
        DiagnosticUnderlineInfo = { sp = colors.bright_cyan, undercurl = true },
        DiagnosticUnderlineHint = { sp = colors.cyan, undercurl = true },
        DiagnosticSignError = { fg = colors.magenta },
        DiagnosticSignWarn = { fg = colors.bright_magenta },
        DiagnosticSignInfo = { fg = colors.bright_cyan },
        DiagnosticSignHint = { fg = colors.cyan },
        DiagnosticFloatingError = { fg = colors.magenta },
        DiagnosticFloatingWarn = { fg = colors.bright_magenta },
        DiagnosticFloatingInfo = { fg = colors.bright_cyan },
        DiagnosticFloatingHint = { fg = colors.cyan },

        -- LSP参照・インレイヒント
        LspReferenceText = { bg = colors.dark_shadow },
        LspReferenceRead = { bg = colors.dark_shadow },
        LspReferenceWrite = { bg = colors.dark_shadow, bold = true },
        LspSignatureActiveParameter = { fg = colors.cyan, bold = true },
        LspCodeLens = { fg = colors.mid_gray },
        LspInlayHint = { fg = colors.mid_gray, italic = true },

        -- Folding関連
        Folded = { bg = colors.dark_shadow, fg = colors.mid_gray },
        FoldColumn = { bg = "none", fg = colors.border },

        -- Flash.nvim
        FlashBackdrop = { fg = colors.mid_gray },
        FlashMatch = { bg = colors.dark_shadow, fg = colors.cyan },
        FlashCurrent = { bg = colors.cyan, fg = colors.bg },
        FlashLabel = { bg = colors.magenta, fg = colors.bg, bold = true },
        FlashPrompt = { fg = colors.fg },
        FlashPromptIcon = { fg = colors.cyan },

        -- Which-Key.nvim
        WhichKey = { fg = colors.cyan },
        WhichKeyGroup = { fg = colors.bright_cyan },
        WhichKeyDesc = { fg = colors.fg },
        WhichKeySeparator = { fg = colors.border },
        WhichKeyFloat = { bg = "none" },
        WhichKeyBorder = { fg = colors.border },
        WhichKeyValue = { fg = colors.mid_gray },

        -- Noice.nvim
        NoiceCmdline = { fg = colors.fg },
        NoiceCmdlineIcon = { fg = colors.cyan },
        NoiceCmdlineIconSearch = { fg = colors.cyan },
        NoiceCmdlinePopup = { bg = "none" },
        NoiceCmdlinePopupBorder = { fg = colors.cyan },
        NoiceCmdlinePopupTitle = { fg = colors.cyan },
        NoiceConfirm = { bg = "none" },
        NoiceConfirmBorder = { fg = colors.cyan },
        NoiceMini = { bg = "none" },
        NoicePopup = { bg = "none" },
        NoicePopupBorder = { fg = colors.cyan },
        NoicePopupmenu = { bg = "none" },
        NoicePopupmenuBorder = { fg = colors.cyan },
        NoicePopupmenuMatch = { fg = colors.cyan, bold = true },
        NoicePopupmenuSelected = { bg = colors.dark_shadow },
        NoiceVirtualText = { fg = colors.mid_gray },

        -- Notify（noice.nvimが使用）
        NotifyERRORBorder = { fg = colors.magenta },
        NotifyWARNBorder = { fg = colors.bright_magenta },
        NotifyINFOBorder = { fg = colors.bright_cyan },
        NotifyDEBUGBorder = { fg = colors.mid_gray },
        NotifyTRACEBorder = { fg = colors.border },
        NotifyERRORIcon = { fg = colors.magenta },
        NotifyWARNIcon = { fg = colors.bright_magenta },
        NotifyINFOIcon = { fg = colors.bright_cyan },
        NotifyDEBUGIcon = { fg = colors.mid_gray },
        NotifyTRACEIcon = { fg = colors.border },
        NotifyERRORTitle = { fg = colors.magenta },
        NotifyWARNTitle = { fg = colors.bright_magenta },
        NotifyINFOTitle = { fg = colors.bright_cyan },
        NotifyDEBUGTitle = { fg = colors.mid_gray },
        NotifyTRACETitle = { fg = colors.border },
        NotifyERRORBody = { fg = colors.fg },
        NotifyWARNBody = { fg = colors.fg },
        NotifyINFOBody = { fg = colors.fg },
        NotifyDEBUGBody = { fg = colors.fg },
        NotifyTRACEBody = { fg = colors.fg },

        -- Trouble.nvim
        TroubleNormal = { bg = "none" },
        TroubleText = { fg = colors.fg },
        TroubleSource = { fg = colors.mid_gray },
        TroubleCode = { fg = colors.mid_gray },
        TroubleLocation = { fg = colors.mid_gray },
        TroubleFile = { fg = colors.cyan },
        TroubleFoldIcon = { fg = colors.border },
        TroubleCount = { fg = colors.magenta, bold = true },
        TroubleError = { fg = colors.magenta },
        TroubleWarning = { fg = colors.bright_magenta },
        TroubleHint = { fg = colors.cyan },
        TroubleInformation = { fg = colors.bright_cyan },

        -- その他ビルトイングループ
        NonText = { fg = colors.border },
        SpecialKey = { fg = colors.border },
        Whitespace = { fg = colors.border },
        Conceal = { fg = colors.mid_gray },
        Question = { fg = colors.cyan },
        MoreMsg = { fg = colors.cyan },
        ModeMsg = { fg = colors.fg, bold = true },
        WildMenu = { bg = colors.dark_shadow, fg = colors.highlight_white },
        QuickFixLine = { bg = colors.dark_shadow },
        WinBar = { bg = "none", fg = colors.fg },
        WinBarNC = { bg = "none", fg = colors.mid_gray },

        -- Lazy.nvim
        LazyButton = { bg = "none", fg = colors.mid_gray },
        LazyButtonActive = { bg = colors.selection, fg = colors.highlight_white, bold = true },
        LazyH1 = { fg = colors.highlight_white, bold = true },
        LazyH2 = { fg = colors.cyan, bold = true },
        LazySpecial = { fg = colors.cyan },
        LazyCommit = { fg = colors.mid_gray },
        LazyCommitType = { fg = colors.cyan },
        LazyDimmed = { fg = colors.mid_gray },
        LazyProp = { fg = colors.light_gray },
        LazyValue = { fg = colors.bright_magenta },
        LazyLocal = { fg = colors.cyan },
        LazyProgressDone = { fg = colors.cyan },
        LazyProgressTodo = { fg = colors.border },
        LazyReasonCmd = { fg = colors.cyan },
        LazyReasonEvent = { fg = colors.bright_cyan },
        LazyReasonFt = { fg = colors.bright_cyan },
        LazyReasonKeys = { fg = colors.cyan },
        LazyReasonPlugin = { fg = colors.cyan },
        LazyReasonStart = { fg = colors.cyan },
      }

      -- ハイライトグループを適用
      for group, attrs in pairs(highlights) do
        vim.api.nvim_set_hl(0, group, attrs)
      end

      -- ターミナルカラー設定（統一ANSI 16色）
      vim.g.terminal_color_0 = colors.bg              -- Black: 背景
      vim.g.terminal_color_1 = colors.ansi_red        -- Red → gray
      vim.g.terminal_color_2 = colors.ansi_green      -- Green → gray
      vim.g.terminal_color_3 = colors.ansi_yellow     -- Yellow → gray
      vim.g.terminal_color_4 = colors.ansi_blue       -- Blue → gray
      vim.g.terminal_color_5 = colors.magenta         -- Magenta: グリッチ紫（アクセント）
      vim.g.terminal_color_6 = colors.cyan            -- Cyan: ネオン青緑（アクセント）
      vim.g.terminal_color_7 = colors.fg              -- White: 前景
      vim.g.terminal_color_8 = colors.dark_shadow     -- Bright Black: 濃い影
      vim.g.terminal_color_9 = colors.bright_red      -- Bright Red → near white
      vim.g.terminal_color_10 = colors.bright_green   -- Bright Green → light gray
      vim.g.terminal_color_11 = colors.bright_yellow  -- Bright Yellow → white
      vim.g.terminal_color_12 = colors.bright_blue    -- Bright Blue → light gray
      vim.g.terminal_color_13 = colors.bright_magenta -- Bright Magenta
      vim.g.terminal_color_14 = colors.bright_cyan    -- Bright Cyan
      vim.g.terminal_color_15 = colors.highlight_white -- Bright White: ハイライト白

      -- 包括的な透明化設定
      local function set_transparent_bg()
        -- 基本的な背景
        vim.api.nvim_set_hl(0, "Normal", { bg = "none" })
        vim.api.nvim_set_hl(0, "NormalNC", { bg = "none" })
        vim.api.nvim_set_hl(0, "EndOfBuffer", { bg = "none" })
        vim.api.nvim_set_hl(0, "SignColumn", { bg = "none" })
        vim.api.nvim_set_hl(0, "VertSplit", { bg = "none" })
        vim.api.nvim_set_hl(0, "WinSeparator", { bg = "none" })

        -- ステータスライン・タブライン
        vim.api.nvim_set_hl(0, "StatusLine", { bg = "none" })
        vim.api.nvim_set_hl(0, "StatusLineNC", { bg = "none" })
        vim.api.nvim_set_hl(0, "TabLine", { fg = colors.mid_gray, bg = "none" })
        vim.api.nvim_set_hl(0, "TabLineSel", { fg = colors.fg, bg = "none", bold = true })
        vim.api.nvim_set_hl(0, "TabLineFill", { bg = "none" })

        -- 行番号
        vim.api.nvim_set_hl(0, "LineNr", { bg = "none" })
        vim.api.nvim_set_hl(0, "CursorLineNr", { bg = "none", fg = colors.highlight_white, bold = true })
        vim.api.nvim_set_hl(0, "FoldColumn", { bg = "none", fg = colors.border })

        -- Snacks関連のハイライトグループ（選択項目は除外）
        local snacks_groups = vim.fn.getcompletion("Snacks", "highlight")
        for _, group in ipairs(snacks_groups) do
          if not string.match(group, "CursorLine") and
             not string.match(group, "Selection") and
             not string.match(group, "Cursor") then
            vim.api.nvim_set_hl(0, group, { bg = "none" })
          end
        end

        -- Snacks Picker選択項目のハイライトを明示的に設定
        vim.api.nvim_set_hl(0, "SnacksPickerListCursorLine", { bg = colors.dark_shadow, fg = colors.highlight_white })
        vim.api.nvim_set_hl(0, "SnacksPickerSelection", { bg = colors.dark_shadow, fg = colors.highlight_white, bold = true })
        vim.api.nvim_set_hl(0, "SnacksPickerCursor", { bg = colors.dark_shadow, fg = colors.highlight_white })
        vim.api.nvim_set_hl(0, "SnacksPickerCursorLine", { bg = colors.dark_shadow, fg = colors.highlight_white })

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

        -- Neotestのハイライトグループ設定（モノクロ基調＋アクセント）
        vim.api.nvim_set_hl(0, "NeotestPassed", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "NeotestFailed", { fg = colors.magenta })
        vim.api.nvim_set_hl(0, "NeotestRunning", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestSkipped", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestMarked", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "NeotestWinSelect", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "NeotestAdapterName", { fg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "NeotestBorder", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestDir", { fg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "NeotestFile", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestNamespace", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "NeotestIndent", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestExpandMarker", { fg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "NeotestFocused", { fg = colors.cyan, bold = true })
        vim.api.nvim_set_hl(0, "NeotestUnknown", { fg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "NeotestTarget", { fg = colors.cyan })

        -- Scrollbar関連のハイライトグループ
        vim.api.nvim_set_hl(0, "ScrollbarHandle", { bg = "none", fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "ScrollbarSearch", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "ScrollbarError", { fg = colors.magenta })
        vim.api.nvim_set_hl(0, "ScrollbarWarn", { fg = colors.bright_magenta })
        vim.api.nvim_set_hl(0, "ScrollbarInfo", { fg = colors.bright_cyan })
        vim.api.nvim_set_hl(0, "ScrollbarHint", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "ScrollbarMisc", { fg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "ScrollbarGitAdd", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "ScrollbarGitChange", { fg = colors.magenta })
        vim.api.nvim_set_hl(0, "ScrollbarGitDelete", { fg = colors.light_gray })
      end

      -- 初回実行
      set_transparent_bg()

      -- 複数のイベントで透明化を実行
      vim.api.nvim_create_autocmd({
        "ColorScheme",
        "UIEnter",
      }, {
        pattern = "*",
        callback = function()
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
