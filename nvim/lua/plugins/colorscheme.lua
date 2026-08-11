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
        Directory = { fg = colors.teal_bright },
        Title = { fg = colors.foreground_heading, bold = true },
        Special = { fg = colors.punctuation_gray },
        Identifier = { fg = colors.syntax_variable },
        Statement = { fg = colors.syntax_violet },
        PreProc = { fg = colors.teal_bright },
        Type = { fg = colors.syntax_type },
        Constant = { fg = colors.syntax_constant },
        String = { fg = colors.string },
        Number = { fg = colors.syntax_number },
        Boolean = { fg = colors.syntax_constant },
        Function = { fg = colors.syntax_function },
        Keyword = { fg = colors.syntax_violet },
        Operator = { fg = colors.operator },
        Comment = { fg = colors.comment_gray },  -- より明るく読みやすく

        -- IMEやフローティングウィンドウの設定
        NormalFloat = { bg = colors.dark_shadow, fg = colors.fg },
        FloatBorder = { bg = colors.dark_shadow, fg = colors.teal_bright },
        Pmenu = { bg = colors.dark_shadow, fg = colors.fg },
        PmenuSel = { bg = colors.dark_shadow, fg = colors.highlight_white },
        PmenuSbar = { bg = colors.dark_shadow },
        PmenuThumb = { bg = colors.mid_gray },

        -- Blink.cmp 用のハイライトグループ
        BlinkCmpMenu = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpMenuBorder = { bg = colors.dark_shadow, fg = colors.teal_bright },
        BlinkCmpMenuSelection = { bg = colors.dark_shadow, fg = colors.highlight_white },
        BlinkCmpDoc = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpDocBorder = { bg = colors.dark_shadow, fg = colors.teal_bright },
        BlinkCmpLabel = { bg = colors.dark_shadow, fg = colors.fg },
        BlinkCmpLabelMatch = { bg = colors.dark_shadow, fg = colors.teal_bright },
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
        SnacksPickerMatch = { bg = "none", fg = colors.teal_bright },
        SnacksPickerList = { bg = "none" },
        SnacksPickerListCursorLine = { bg = colors.selection, fg = colors.selection_fg, bold = true },
        SnacksPickerSelection      = { bg = colors.selection, fg = colors.selection_fg, bold = true },
        SnacksPickerPathIgnored = { bg = "none" },

        -- 一般的なサイドバー関連のハイライトグループ
        SnacksPickerTree = { bg = "none" },
        LineNr = { bg = "none" },
        CursorLineNr = { bg = "none", fg = colors.highlight_white, bold = true },

        -- MiniIcons ハイライトグループ - 白色統一表示
        MiniIconsAzure = { fg = colors.sky_slate },
        MiniIconsBlue = { fg = colors.sky_slate },
        MiniIconsCyan = { fg = colors.sky_slate },
        MiniIconsGreen = { fg = colors.sky_slate },
        MiniIconsGrey = { fg = colors.sky_slate },
        MiniIconsOrange = { fg = colors.sky_slate },
        MiniIconsPurple = { fg = colors.sky_slate },
        MiniIconsRed = { fg = colors.sky_slate },
        MiniIconsYellow = { fg = colors.sky_slate },

        -- GitSigns ハイライトグループ - モノクロ基調＋アクセント
        GitSignsAdd = { fg = colors.git_added },
        GitSignsChange = { fg = colors.git_changed },
        GitSignsDelete = { fg = colors.git_deleted },
        GitSignsAddNr = { fg = colors.git_added },
        GitSignsChangeNr = { fg = colors.git_changed },
        GitSignsDeleteNr = { fg = colors.git_deleted },
        GitSignsCurrentLineBlame = { fg = colors.git_blame_gray },  -- コメントより明るく

        -- Diff関連のハイライトグループ
        DiffAdd = { fg = colors.git_added, bg = colors.diff_add_bg },
        DiffChange = { fg = colors.git_changed, bg = colors.diff_change_bg },
        DiffDelete = { fg = colors.git_deleted, bg = colors.diff_delete_bg },
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
        ["@keyword.import"] = { fg = colors.teal_bright },
        ["@string"] = { fg = colors.string },
        ["@number"] = { fg = colors.syntax_number },
        ["@boolean"] = { fg = colors.syntax_constant },
        ["@comment"] = { fg = colors.comment_gray },  -- Commentと統一
        ["@function"] = { fg = colors.syntax_function },
        ["@function.call"] = { fg = colors.syntax_function },
        ["@function.method"] = { fg = colors.syntax_function },
        ["@function.method.call"] = { fg = colors.syntax_function },
        ["@function.builtin"] = { fg = colors.syntax_function },
        ["@function.macro"] = { fg = colors.syntax_function },
        ["@variable"] = { fg = colors.syntax_variable },
        ["@variable.builtin"] = { fg = colors.syntax_builtin_variable },
        ["@variable.member"] = { fg = colors.syntax_variable },
        ["@variable.parameter"] = { fg = colors.syntax_variable },
        ["@variable.parameter.builtin"] = { fg = colors.syntax_builtin_variable },
        ["@parameter"] = { fg = colors.syntax_variable },
        ["@type"] = { fg = colors.syntax_type },
        ["@type.builtin"] = { fg = colors.syntax_type },
        ["@property"] = { fg = colors.syntax_variable },
        ["@field"] = { fg = colors.syntax_variable },
        ["@constant"] = { fg = colors.syntax_constant },
        ["@constant.builtin"] = { fg = colors.syntax_constant },
        ["@operator"] = { fg = colors.operator },
        ["@punctuation"] = { fg = colors.punctuation_gray },
        ["@punctuation.bracket"] = { fg = colors.punctuation_gray },
        ["@punctuation.delimiter"] = { fg = colors.punctuation_gray },
        ["@punctuation.delimiter.yaml"] = { fg = colors.light_gray }, -- YAML list marker visibility
        -- YAML: 値の型で色分け（キー=cyan / 数値=lavender / 真偽・null=syntax violet。文字列は@string既定を維持）
        ["@property.yaml"] = { fg = colors.teal_bright },
        ["@boolean.yaml"] = { fg = colors.syntax_violet },
        ["@number.yaml"] = { fg = colors.lavender },
        ["@constant.builtin.yaml"] = { fg = colors.syntax_violet },
        ["@namespace"] = { fg = colors.fg },
        ["@module"] = { fg = colors.fg },
        ["@tag"] = { fg = colors.teal_bright },
        ["@tag.attribute"] = { fg = colors.foreground_heading },
        ["@tag.delimiter"] = { fg = colors.subdued_fg },

        -- Treesitter Markup（Markdown用）
        ["@markup.heading"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.1"] = { fg = colors.highlight_white, bold = true },
        ["@markup.heading.2"] = { fg = colors.foreground_bright, bold = true },
        ["@markup.heading.3"] = { fg = colors.fg, bold = true },
        ["@markup.heading.4"] = { fg = colors.operator, bold = true },
        ["@markup.heading.5"] = { fg = colors.light_gray, bold = true },
        ["@markup.heading.6"] = { fg = colors.subdued_fg, bold = true },
        -- Markdown 見出し背景: teal グラデーション（treesitter + render-markdown.nvim 共有）
        ["@markup.heading.1.markdown"] = { bg = colors.foreground_heading, fg = colors.bg, bold = true },
        ["@markup.heading.2.markdown"] = { bg = colors.teal_bright,        fg = colors.bg, bold = true },
        ["@markup.heading.3.markdown"] = { bg = colors.operator,       fg = colors.bg, bold = true },
        ["@markup.heading.4.markdown"] = { bg = colors.fg,             fg = colors.bg, bold = true },
        ["@markup.heading.5.markdown"] = { bg = colors.cloud_slate,        fg = colors.bg, bold = true },
        ["@markup.heading.6.markdown"] = { bg = colors.light_gray,     fg = colors.bg, bold = true },
        -- render-markdown.nvim: インライン編集レンダリング用ハイライト
        -- 見出し行背景: @markup.heading.<N>.markdown.bg と同じ teal グラデーション（palette 参照）
        RenderMarkdownH1Bg = { bg = colors.foreground_heading },
        RenderMarkdownH2Bg = { bg = colors.teal_bright },
        RenderMarkdownH3Bg = { bg = colors.operator },
        RenderMarkdownH4Bg = { bg = colors.fg },
        RenderMarkdownH5Bg = { bg = colors.cloud_slate },
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
        RenderMarkdownTableHead = { fg = colors.foreground_heading, bold = true },
        RenderMarkdownTableRow  = { fg = colors.fg },
        -- リスト・チェックボックス
        RenderMarkdownBullet    = { fg = colors.punctuation_gray },
        RenderMarkdownChecked   = { fg = colors.ansi_green },
        RenderMarkdownUnchecked = { fg = colors.light_gray },
        RenderMarkdownTodo      = { fg = colors.ui_accent_fg },
        -- リンク・クォート
        RenderMarkdownLink  = { fg = colors.teal_bright, underline = true },
        RenderMarkdownQuote = { fg = colors.light_gray, italic = true },
        -- アラート（重大度ごとのsemantic aliasを使用）
        RenderMarkdownInfo    = { fg = colors.diagnostic_info,  bold = true },
        RenderMarkdownHint    = { fg = colors.diagnostic_hint,  bold = true },
        RenderMarkdownSuccess = { fg = colors.success,        bold = true },
        RenderMarkdownWarn    = { fg = colors.diagnostic_warn,  bold = true },
        RenderMarkdownError   = { fg = colors.diagnostic_error, bold = true },
        ["@markup.list"] = { fg = colors.light_gray },
        ["@markup.list.markdown"] = { fg = colors.light_gray },
        ["@markup.list.checked"] = { fg = colors.teal_bright },
        ["@markup.list.unchecked"] = { fg = colors.light_gray },
        ["@markup.link"] = { fg = colors.teal_bright, underline = true },
        ["@markup.link.label"] = { fg = colors.teal_bright },
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
        OilDir = { fg = colors.teal_bright },
        OilDirIcon = { fg = colors.teal_bright },
        OilLink = { fg = colors.syntax_violet },
        OilFile = { fg = colors.fg },

        -- TreesitterContext (スティッキースクロール) ハイライトグループ
        TreesitterContext = { bg = colors.dark_shadow, fg = colors.fg },
        TreesitterContextBottom = { underline = true, sp = colors.dark_shadow },
        TreesitterContextLineNumber = { bg = colors.dark_shadow, fg = colors.subdued_fg },

        -- 診断関連（bufferline、LSP診断などが参照）
        DiagnosticError = { fg = colors.diagnostic_error },
        DiagnosticWarn = { fg = colors.diagnostic_warn },
        DiagnosticInfo = { fg = colors.diagnostic_info },
        DiagnosticHint = { fg = colors.diagnostic_hint },
        WarningMsg = { fg = colors.diagnostic_warn },
        ErrorMsg = { fg = colors.diagnostic_error },

        -- 検索・マッチ関連
        Search = { bg = colors.dark_shadow, fg = colors.teal_bright },
        IncSearch = { bg = colors.teal_bright, fg = colors.bg },
        CurSearch = { bg = colors.foreground_heading, fg = colors.bg },
        Substitute = { bg = colors.ui_target_bg, fg = colors.bg },
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
        DiagnosticVirtualTextError = { fg = colors.diagnostic_error },
        DiagnosticVirtualTextWarn = { fg = colors.diagnostic_warn },
        DiagnosticVirtualTextInfo = { fg = colors.diagnostic_info },
        DiagnosticVirtualTextHint = { fg = colors.diagnostic_hint },
        DiagnosticUnderlineError = { sp = colors.diagnostic_error, undercurl = true },
        DiagnosticUnderlineWarn = { sp = colors.diagnostic_warn, undercurl = true },
        DiagnosticUnderlineInfo = { sp = colors.diagnostic_info, undercurl = true },
        DiagnosticUnderlineHint = { sp = colors.diagnostic_hint, undercurl = true },
        DiagnosticSignError = { fg = colors.diagnostic_error },
        DiagnosticSignWarn = { fg = colors.diagnostic_warn },
        DiagnosticSignInfo = { fg = colors.diagnostic_info },
        DiagnosticSignHint = { fg = colors.diagnostic_hint },
        DiagnosticFloatingError = { fg = colors.diagnostic_error },
        DiagnosticFloatingWarn = { fg = colors.diagnostic_warn },
        DiagnosticFloatingInfo = { fg = colors.diagnostic_info },
        DiagnosticFloatingHint = { fg = colors.diagnostic_hint },

        -- LSP参照・インレイヒント
        LspReferenceText = { bg = colors.dark_shadow },
        LspReferenceRead = { bg = colors.dark_shadow },
        LspReferenceWrite = { bg = colors.dark_shadow, bold = true },
        LspSignatureActiveParameter = { fg = colors.teal_bright, bold = true },
        LspCodeLens = { fg = colors.subdued_fg },
        LspInlayHint = { fg = colors.subdued_fg, italic = true },

        -- Folding関連
        Folded = { bg = colors.dark_shadow, fg = colors.subdued_fg },
        FoldColumn = { bg = "none", fg = colors.border },

        -- Which-Key.nvim
        WhichKey = { fg = colors.teal_bright },
        WhichKeyGroup = { fg = colors.foreground_heading },
        WhichKeyDesc = { fg = colors.fg },
        WhichKeySeparator = { fg = colors.border },
        WhichKeyFloat = { bg = "none" },
        WhichKeyBorder = { fg = colors.border },
        WhichKeyValue = { fg = colors.subdued_fg },

        -- Noice.nvim
        NoiceCmdline = { fg = colors.fg },
        NoiceCmdlineIcon = { fg = colors.teal_bright },
        NoiceCmdlineIconSearch = { fg = colors.teal_bright },
        NoiceCmdlinePopup = { bg = "none" },
        NoiceCmdlinePopupBorder = { fg = colors.teal_bright },
        NoiceCmdlinePopupTitle = { fg = colors.teal_bright },
        NoiceConfirm = { bg = "none" },
        NoiceConfirmBorder = { fg = colors.teal_bright },
        NoiceMini = { bg = "none" },
        NoicePopup = { bg = "none" },
        NoicePopupBorder = { fg = colors.teal_bright },
        NoicePopupmenu = { bg = "none" },
        NoicePopupmenuBorder = { fg = colors.teal_bright },
        NoicePopupmenuMatch = { fg = colors.teal_bright, bold = true },
        NoicePopupmenuSelected = { bg = colors.dark_shadow },
        NoiceVirtualText = { fg = colors.subdued_fg },

        -- Notify（noice.nvimが使用）
        NotifyERRORBorder = { fg = colors.diagnostic_error },
        NotifyWARNBorder = { fg = colors.diagnostic_warn },
        NotifyINFOBorder = { fg = colors.diagnostic_info },
        NotifyDEBUGBorder = { fg = colors.subdued_fg },
        NotifyTRACEBorder = { fg = colors.border },
        NotifyERRORIcon = { fg = colors.diagnostic_error },
        NotifyWARNIcon = { fg = colors.diagnostic_warn },
        NotifyINFOIcon = { fg = colors.diagnostic_info },
        NotifyDEBUGIcon = { fg = colors.subdued_fg },
        NotifyTRACEIcon = { fg = colors.border },
        NotifyERRORTitle = { fg = colors.diagnostic_error },
        NotifyWARNTitle = { fg = colors.diagnostic_warn },
        NotifyINFOTitle = { fg = colors.diagnostic_info },
        NotifyDEBUGTitle = { fg = colors.subdued_fg },
        NotifyTRACETitle = { fg = colors.border },
        NotifyERRORBody = { fg = colors.fg },
        NotifyWARNBody = { fg = colors.fg },
        NotifyINFOBody = { fg = colors.fg },
        NotifyDEBUGBody = { fg = colors.fg },
        NotifyTRACEBody = { fg = colors.fg },

        -- その他ビルトイングループ
        NonText = { fg = colors.border },
        SpecialKey = { fg = colors.border },
        Whitespace = { fg = colors.border },
        Conceal = { fg = colors.mid_gray },
        Question = { fg = colors.teal_bright },
        MoreMsg = { fg = colors.teal_bright },
        ModeMsg = { fg = colors.fg, bold = true },
        WildMenu = { bg = colors.dark_shadow, fg = colors.highlight_white },
        QuickFixLine = { bg = colors.dark_shadow },
        WinBar = { bg = "none", fg = colors.fg },
        WinBarNC = { bg = "none", fg = colors.subdued_fg },

        -- Lazy.nvim
        LazyButton = { bg = "none", fg = colors.subdued_fg },
        LazyButtonActive = { bg = colors.selection, fg = colors.darkest_bg, bold = true },
        LazyH1 = { fg = colors.highlight_white, bold = true },
        LazyH2 = { fg = colors.teal_bright, bold = true },
        LazySpecial = { fg = colors.teal_bright },
        LazyCommit = { fg = colors.subdued_fg },
        LazyCommitType = { fg = colors.teal_bright },
        LazyDimmed = { fg = colors.mid_gray },
        LazyProp = { fg = colors.light_gray },
        LazyValue = { fg = colors.cloud_slate },
        LazyLocal = { fg = colors.teal_bright },
        LazyProgressDone = { fg = colors.teal_bright },
        LazyProgressTodo = { fg = colors.border },
        LazyReasonCmd = { fg = colors.teal_bright },
        LazyReasonEvent = { fg = colors.foreground_heading },
        LazyReasonFt = { fg = colors.foreground_heading },
        LazyReasonKeys = { fg = colors.teal_bright },
        LazyReasonPlugin = { fg = colors.teal_bright },
        LazyReasonStart = { fg = colors.teal_bright },
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
        vim.api.nvim_set_hl(0, "NeotestPassed", { fg = colors.teal_bright })
        vim.api.nvim_set_hl(0, "NeotestFailed", { fg = colors.diagnostic_error })
        vim.api.nvim_set_hl(0, "NeotestRunning", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestSkipped", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestMarked", { fg = colors.teal_bright })
        vim.api.nvim_set_hl(0, "NeotestWinSelect", { fg = colors.teal_bright })
        vim.api.nvim_set_hl(0, "NeotestAdapterName", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestBorder", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestDir", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestFile", { fg = colors.fg })
        vim.api.nvim_set_hl(0, "NeotestNamespace", { fg = colors.teal_bright })
        vim.api.nvim_set_hl(0, "NeotestIndent", { fg = colors.dark_shadow })
        vim.api.nvim_set_hl(0, "NeotestExpandMarker", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestFocused", { fg = colors.teal_bright, bold = true })
        vim.api.nvim_set_hl(0, "NeotestUnknown", { fg = colors.subdued_fg })
        vim.api.nvim_set_hl(0, "NeotestTarget", { fg = colors.teal_bright })

        -- Scrollbar関連のハイライトグループ
        vim.api.nvim_set_hl(0, "ScrollbarHandle", { bg = colors.mid_gray })
        vim.api.nvim_set_hl(0, "ScrollbarSearch", { fg = colors.teal_bright })
        vim.api.nvim_set_hl(0, "ScrollbarError", { fg = colors.diagnostic_error })
        vim.api.nvim_set_hl(0, "ScrollbarWarn", { fg = colors.diagnostic_warn })
        vim.api.nvim_set_hl(0, "ScrollbarInfo", { fg = colors.diagnostic_info })
        vim.api.nvim_set_hl(0, "ScrollbarHint", { fg = colors.diagnostic_hint })
        vim.api.nvim_set_hl(0, "ScrollbarMisc", { fg = colors.light_gray })
        vim.api.nvim_set_hl(0, "ScrollbarGitAdd", { fg = colors.git_added })
        vim.api.nvim_set_hl(0, "ScrollbarGitChange", { fg = colors.git_changed })
        vim.api.nvim_set_hl(0, "ScrollbarGitDelete", { fg = colors.git_deleted })
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
