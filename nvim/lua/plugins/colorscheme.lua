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
      -- カラーパレット定義 — TOMLソース: colors/ghost-visor.toml
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
        Special = { fg = colors.punctuation_gray },
        Identifier = { fg = colors.fg },
        Statement = { fg = colors.syntax_violet },
        PreProc = { fg = colors.cyan },
        Type = { fg = colors.white },
        Constant = { fg = colors.lavender },
        String = { fg = colors.string },
        Number = { fg = colors.near_white },
        Boolean = { fg = colors.near_white },
        Function = { fg = colors.cyan },
        Keyword = { fg = colors.syntax_violet },
        Operator = { fg = colors.operator },
        Comment = { fg = colors.comment_gray },  -- より明るく読みやすく

        -- IMEやフローティングウィンドウの設定
        NormalFloat = { bg = colors.dark_shadow, fg = colors.fg },
        FloatBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        Pmenu = { bg = colors.dark_shadow, fg = colors.fg },
        PmenuSel = { bg = colors.dark_shadow, fg = colors.highlight_white },
        PmenuSbar = { bg = colors.dark_shadow },
        PmenuThumb = { bg = colors.mid_gray },

        -- Blink.cmp 用のハイライトグループ
        BlinkCmpMenu = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpMenuBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpMenuSelection = { bg = colors.dark_shadow, fg = colors.highlight_white },
        BlinkCmpDoc = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpDocBorder = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpLabel = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpLabelMatch = { bg = colors.dark_shadow, fg = colors.cyan },
        BlinkCmpKind = { bg = colors.dark_shadow, fg = colors.subdued_fg },
        BlinkCmpSource = { bg = colors.dark_shadow, fg = colors.subdued_fg },

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
        SnacksPickerListCursorLine = { bg = colors.selection, fg = colors.selection_fg, bold = true },
        SnacksPickerSelection      = { bg = colors.selection, fg = colors.selection_fg, bold = true },
        SnacksPickerPathIgnored = { bg = "none" },

        -- 一般的なサイドバー関連のハイライトグループ
        SnacksPickerTree = { bg = "none" },
        LineNr = { bg = "none" },
        CursorLineNr = { bg = "none", fg = colors.highlight_white, bold = true },

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
        DiffAdd = { fg = colors.cyan, bg = colors.diff_add_bg },
        DiffChange = { fg = colors.magenta, bg = colors.diff_change_bg },
        DiffDelete = { fg = colors.light_gray, bg = colors.diff_delete_bg },
        DiffText = { fg = colors.highlight_white, bg = colors.diff_change_inline_bg, bold = true }, -- bg明示で変更語の可読性確保

        -- word_diff（行内差分）ハイライト - 行背景より明るいtint＋明色文字で可読性確保
        GitSignsAddInline = { fg = colors.highlight_white, bg = colors.diff_add_inline_bg },
        GitSignsChangeInline = { fg = colors.highlight_white, bg = colors.diff_change_inline_bg },
        GitSignsDeleteInline = { fg = colors.highlight_white, bg = colors.diff_delete_inline_bg },

        -- 構文ハイライト（Cyber Glitch Teal - ネオン系配色）
        ["@keyword"] = { fg = colors.syntax_violet },
        ["@keyword.function"] = { fg = colors.syntax_violet },
        ["@keyword.operator"] = { fg = colors.syntax_violet },
        ["@keyword.return"] = { fg = colors.syntax_violet },
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
        -- YAML: 値の型で色分け（キー=cyan / 数値=lavender / 真偽・null=syntax violet。文字列は@string既定を維持）
        ["@property.yaml"] = { fg = colors.cyan },
        ["@boolean.yaml"] = { fg = colors.syntax_violet },
        ["@number.yaml"] = { fg = colors.lavender },
        ["@constant.builtin.yaml"] = { fg = colors.syntax_violet },
        ["@namespace"] = { fg = colors.fg },
        ["@module"] = { fg = colors.fg },
        ["@tag"] = { fg = colors.cyan },
        ["@tag.attribute"] = { fg = colors.bright_cyan },
        ["@tag.delimiter"] = { fg = colors.subdued_fg },

        -- Treesitter Markup（Markdown用）
        ["@markup.heading"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.1"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.2"] = { fg = colors.near_white, bold = true },
        ["@markup.heading.3"] = { fg = colors.fg, bold = true },
        ["@markup.heading.4"] = { fg = colors.operator, bold = true },
        ["@markup.heading.5"] = { fg = colors.light_gray, bold = true },
        ["@markup.heading.6"] = { fg = colors.subdued_fg, bold = true },
        -- Markdown 見出し背景: teal グラデーション（treesitter + render-markdown.nvim 共有）
        ["@markup.heading.1.markdown"] = { bg = colors.bright_cyan,    fg = colors.bg, bold = true },
        ["@markup.heading.2.markdown"] = { bg = colors.cyan,           fg = colors.bg, bold = true },
        ["@markup.heading.3.markdown"] = { bg = colors.operator,       fg = colors.bg, bold = true },
        ["@markup.heading.4.markdown"] = { bg = colors.fg,             fg = colors.bg, bold = true },
        ["@markup.heading.5.markdown"] = { bg = colors.bright_magenta, fg = colors.bg, bold = true },
        ["@markup.heading.6.markdown"] = { bg = colors.light_gray,     fg = colors.bg, bold = true },
        -- render-markdown.nvim: インライン編集レンダリング用ハイライト
        -- 見出し行背景: @markup.heading.<N>.markdown.bg と同じ teal グラデーション（palette 参照）
        RenderMarkdownH1Bg = { bg = colors.bright_cyan },
        RenderMarkdownH2Bg = { bg = colors.cyan },
        RenderMarkdownH3Bg = { bg = colors.operator },
        RenderMarkdownH4Bg = { bg = colors.fg },
        RenderMarkdownH5Bg = { bg = colors.bright_magenta },
        RenderMarkdownH6Bg = { bg = colors.light_gray },
        -- 見出しアイコン: teal 背景上で読みやすい黒
        RenderMarkdownH1 = { fg = colors.bg, bold = true },
        RenderMarkdownH2 = { fg = colors.bg, bold = true },
        RenderMarkdownH3 = { fg = colors.bg, bold = true },
        RenderMarkdownH4 = { fg = colors.bg, bold = true },
        RenderMarkdownH5 = { fg = colors.bg, bold = true },
        RenderMarkdownH6 = { fg = colors.bg, bold = true },
        -- コードブロック
        RenderMarkdownCode       = { bg = colors.panel_bg },
        RenderMarkdownCodeInline = { bg = colors.panel_bg, fg = colors.fg },
        -- テーブル
        RenderMarkdownTableHead = { fg = colors.bright_cyan, bold = true },
        RenderMarkdownTableRow  = { fg = colors.fg },
        -- リスト・チェックボックス
        RenderMarkdownBullet    = { fg = colors.punctuation_gray },
        RenderMarkdownChecked   = { fg = colors.ansi_green },
        RenderMarkdownUnchecked = { fg = colors.light_gray },
        RenderMarkdownTodo      = { fg = colors.magenta },
        -- リンク・クォート
        RenderMarkdownLink  = { fg = colors.cyan, underline = true },
        RenderMarkdownQuote = { fg = colors.light_gray, italic = true },
        -- アラート（Info=bright_cyan / Warn=bright_magenta / Error=magenta）
        RenderMarkdownInfo    = { fg = colors.bright_cyan,    bold = true },
        RenderMarkdownHint    = { fg = colors.cyan,           bold = true },
        RenderMarkdownSuccess = { fg = colors.success,        bold = true },
        RenderMarkdownWarn    = { fg = colors.bright_magenta, bold = true },
        RenderMarkdownError   = { fg = colors.magenta,        bold = true },
        ["@markup.list"] = { fg = colors.light_gray },
        ["@markup.list.markdown"] = { fg = colors.light_gray },
        ["@markup.list.checked"] = { fg = colors.cyan },
        ["@markup.list.unchecked"] = { fg = colors.light_gray },
        ["@markup.link"] = { fg = colors.cyan, underline = true },
        ["@markup.link.label"] = { fg = colors.cyan },
        ["@markup.link.url"] = { fg = colors.subdued_fg, underline = true },
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
        TreesitterContextLineNumber = { bg = colors.dark_shadow, fg = colors.subdued_fg },

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
        CursorIM = { bg = colors.fg, fg = colors.bg },
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
        LspCodeLens = { fg = colors.subdued_fg },
        LspInlayHint = { fg = colors.subdued_fg, italic = true },

        -- Folding関連
        Folded = { bg = colors.dark_shadow, fg = colors.subdued_fg },
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
        WhichKeyValue = { fg = colors.subdued_fg },

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
        NoiceVirtualText = { fg = colors.subdued_fg },

        -- Notify（noice.nvimが使用）
        NotifyERRORBorder = { fg = colors.magenta },
        NotifyWARNBorder = { fg = colors.bright_magenta },
        NotifyINFOBorder = { fg = colors.bright_cyan },
        NotifyDEBUGBorder = { fg = colors.subdued_fg },
        NotifyTRACEBorder = { fg = colors.border },
        NotifyERRORIcon = { fg = colors.magenta },
        NotifyWARNIcon = { fg = colors.bright_magenta },
        NotifyINFOIcon = { fg = colors.bright_cyan },
        NotifyDEBUGIcon = { fg = colors.subdued_fg },
        NotifyTRACEIcon = { fg = colors.border },
        NotifyERRORTitle = { fg = colors.magenta },
        NotifyWARNTitle = { fg = colors.bright_magenta },
        NotifyINFOTitle = { fg = colors.bright_cyan },
        NotifyDEBUGTitle = { fg = colors.subdued_fg },
        NotifyTRACETitle = { fg = colors.border },
        NotifyERRORBody = { fg = colors.fg },
        NotifyWARNBody = { fg = colors.fg },
        NotifyINFOBody = { fg = colors.fg },
        NotifyDEBUGBody = { fg = colors.fg },
        NotifyTRACEBody = { fg = colors.fg },

        -- Trouble.nvim
        TroubleNormal = { bg = "none" },
        TroubleText = { fg = colors.fg },
        TroubleSource = { fg = colors.subdued_fg },
        TroubleCode = { fg = colors.subdued_fg },
        TroubleLocation = { fg = colors.subdued_fg },
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
        WinBarNC = { bg = "none", fg = colors.subdued_fg },

        -- Lazy.nvim
        LazyButton = { bg = "none", fg = colors.subdued_fg },
        LazyButtonActive = { bg = colors.selection, fg = colors.darkest_bg, bold = true },
        LazyH1 = { fg = colors.highlight_white, bold = true },
        LazyH2 = { fg = colors.cyan, bold = true },
        LazySpecial = { fg = colors.cyan },
        LazyCommit = { fg = colors.subdued_fg },
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
      vim.g.terminal_color_0 = colors.ansi_black             -- ANSI Black (ansi.black)
      vim.g.terminal_color_1 = colors.ansi_red               -- ANSI Red (ansi.red)
      vim.g.terminal_color_2 = colors.ansi_green             -- ANSI Green (ansi.green)
      vim.g.terminal_color_3 = colors.ansi_yellow            -- ANSI Yellow (ansi.yellow)
      vim.g.terminal_color_4 = colors.ansi_blue              -- ANSI Blue (ansi.blue)
      vim.g.terminal_color_5 = colors.ansi_magenta           -- ANSI Magenta (ansi.magenta)
      vim.g.terminal_color_6 = colors.ansi_cyan              -- ANSI Cyan (ansi.cyan)
      vim.g.terminal_color_7 = colors.ansi_white             -- ANSI White (ansi.white)
      vim.g.terminal_color_8 = colors.ansi_bright_black      -- ANSI Bright Black (ansi.bright_black)
      vim.g.terminal_color_9 = colors.ansi_bright_red        -- ANSI Bright Red (ansi.bright_red)
      vim.g.terminal_color_10 = colors.ansi_bright_green     -- ANSI Bright Green (ansi.bright_green)
      vim.g.terminal_color_11 = colors.ansi_bright_yellow    -- ANSI Bright Yellow (ansi.bright_yellow)
      vim.g.terminal_color_12 = colors.ansi_bright_blue      -- ANSI Bright Blue (ansi.bright_blue)
      vim.g.terminal_color_13 = colors.ansi_bright_magenta   -- ANSI Bright Magenta (ansi.bright_magenta)
      vim.g.terminal_color_14 = colors.ansi_bright_cyan      -- ANSI Bright Cyan (ansi.bright_cyan)
      vim.g.terminal_color_15 = colors.ansi_bright_white     -- ANSI Bright White (ansi.bright_white)

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
        vim.api.nvim_set_hl(0, "TabLine", { fg = colors.subdued_fg, bg = "none" })
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
        vim.api.nvim_set_hl(0, "SnacksPickerListCursorLine", { bg = colors.selection, fg = colors.selection_fg, bold = true })
        vim.api.nvim_set_hl(0, "SnacksPickerSelection",      { bg = colors.selection, fg = colors.selection_fg, bold = true })
        vim.api.nvim_set_hl(0, "SnacksPickerCursor",         { bg = colors.selection, fg = colors.selection_fg })
        vim.api.nvim_set_hl(0, "SnacksPickerCursorLine",     { bg = colors.selection, fg = colors.selection_fg, bold = true })

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
        vim.api.nvim_set_hl(0, "NeotestAdapterName", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestBorder", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestDir", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestFile", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestNamespace", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "NeotestIndent", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestExpandMarker", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestFocused", { fg = colors.cyan, bold = true })
        vim.api.nvim_set_hl(0, "NeotestUnknown", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestTarget", { fg = colors.cyan })

        -- Scrollbar関連のハイライトグループ
        vim.api.nvim_set_hl(0, "ScrollbarHandle", { bg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "ScrollbarSearch", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "ScrollbarError", { fg = colors.magenta })
        vim.api.nvim_set_hl(0, "ScrollbarWarn", { fg = colors.bright_magenta })
        vim.api.nvim_set_hl(0, "ScrollbarInfo", { fg = colors.bright_cyan })
        vim.api.nvim_set_hl(0, "ScrollbarHint", { fg = colors.cyan })
        vim.api.nvim_set_hl(0, "ScrollbarMisc", { fg = colors.light_gray })
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
