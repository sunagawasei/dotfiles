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
          NormalFloat = { bg = "#2A2F2E", fg = "#D7E2E1" }, -- フローティングウィンドウの背景と文字色（濃い影・氷白）
          FloatBorder = { bg = "#2A2F2E", fg = "#5AAFAD" }, -- フローティングウィンドウの境界線（発光ライン）
          Pmenu = { bg = "#2A2F2E", fg = "#D7E2E1" }, -- ポップアップメニューの背景と文字色
          PmenuSel = { bg = "#2A2F2E", fg = "#F2FFFF" }, -- ポップアップメニューの選択項目（影・bright white）
          PmenuSbar = { bg = "#2A2F2E" }, -- スクロールバーの背景
          PmenuThumb = { bg = "#7E8A89" }, -- スクロールバーのつまみ（中間調）
          -- CJK IME support
          CmpItemAbbrMatch = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemAbbrMatchFuzzy = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemAbbr = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKind = { bg = "#2A2F2E", fg = "#7E8A89" },
          CmpItemMenu = { bg = "#2A2F2E", fg = "#7E8A89" },
          CmpItemKindFunction = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindMethod = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindConstructor = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindClass = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindEnum = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindEvent = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindInterface = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindStruct = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindVariable = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindField = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindProperty = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindEnumMember = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindConstant = { bg = "#2A2F2E", fg = "#96CBD1" },
          CmpItemKindKeyword = { bg = "#2A2F2E", fg = "#5AAFAD" },
          CmpItemKindModule = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindValue = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindUnit = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindText = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindSnippet = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindFile = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindFolder = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindColor = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindReference = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindOperator = { bg = "#2A2F2E", fg = "#D7E2E1" },
          CmpItemKindTypeParameter = { bg = "#2A2F2E", fg = "#D7E2E1" },

          -- Blink.cmp 用のハイライトグループ
          BlinkCmpMenu = { bg = "#2A2F2E", fg = "#D7E2E1" },
          BlinkCmpMenuBorder = { bg = "#2A2F2E", fg = "#5AAFAD" },
          BlinkCmpMenuSelection = { bg = "#2A2F2E", fg = "#F2FFFF" },
          BlinkCmpDoc = { bg = "#2A2F2E", fg = "#D7E2E1" },
          BlinkCmpDocBorder = { bg = "#2A2F2E", fg = "#5AAFAD" },
          BlinkCmpLabel = { bg = "#2A2F2E", fg = "#D7E2E1" },
          BlinkCmpLabelMatch = { bg = "#2A2F2E", fg = "#5AAFAD" },
          BlinkCmpKind = { bg = "#2A2F2E", fg = "#7E8A89" },
          BlinkCmpSource = { bg = "#2A2F2E", fg = "#7E8A89" },
          -- Snacks.nvim Picker/Explorer の透明化
          SnacksPickerNormal = { bg = "none" },
          SnacksPickerBorder = { bg = "none" },
          SnacksPickerNormalFloat = { bg = "none" },
          SnacksPickerFile = { bg = "none" },
          SnacksPickerDir = { bg = "none" },
          SnacksPickerPathHidden = { bg = "none" },
          SnacksPickerBox = { bg = "none" },
          SnacksPickerPrompt = { bg = "none" },
          SnacksPickerMatch = { bg = "none", fg = "#5AAFAD" },
          SnacksPickerList = { bg = "none" },
          SnacksPickerListCursorLine = { bg = "#2A2F2E", fg = "#F2FFFF" },
          SnacksPickerSelection = { bg = "#2A2F2E", fg = "#F2FFFF", bold = true },
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
          
          -- MiniIcons ハイライトグループ - 白色統一表示
          MiniIconsAzure = { fg = "#EDEDED" },   -- 白色統一
          MiniIconsBlue = { fg = "#EDEDED" },    -- 白色統一
          MiniIconsCyan = { fg = "#EDEDED" },    -- 白色統一
          MiniIconsGreen = { fg = "#EDEDED" },   -- 白色統一
          MiniIconsGrey = { fg = "#EDEDED" },    -- 白色統一
          MiniIconsOrange = { fg = "#EDEDED" },  -- 白色統一
          MiniIconsPurple = { fg = "#EDEDED" },  -- 白色統一
          MiniIconsRed = { fg = "#EDEDED" },     -- 白色統一
          MiniIconsYellow = { fg = "#EDEDED" },  -- 白色統一
          
          -- GitSigns ハイライトグループ - モノクロ基調＋アクセント
          GitSignsAdd = { fg = "#5AAFAD" },           -- cyanアクセント: 追加
          GitSignsChange = { fg = "#8C83A3" },        -- magentaアクセント: 変更
          GitSignsDelete = { fg = "#AAB6B5" },        -- 明るいグレー: 削除
          GitSignsAddNr = { fg = "#5AAFAD" },         -- cyanアクセント
          GitSignsChangeNr = { fg = "#8C83A3" },      -- magentaアクセント
          GitSignsDeleteNr = { fg = "#AAB6B5" },      -- 明るいグレー
          GitSignsCurrentLineBlame = { fg = "#7E8A89" }, -- 中間グレー

          -- Diff関連のハイライトグループ
          DiffAdd = { fg = "#5AAFAD" },               -- cyanアクセント
          DiffChange = { fg = "#8C83A3" },            -- magentaアクセント
          DiffDelete = { fg = "#AAB6B5" },            -- 明るいグレー
          DiffText = { fg = "#96CBD1", bold = true }, -- bright cyanアクセント

          -- 構文ハイライト（Cyber Glitch Teal - ネオン系配色）
          -- キーワード（発光ライン - if, const, function など）
          ["@keyword"] = { fg = "#5AAFAD" },           -- 発光ライン（ネオン青緑）
          ["@keyword.function"] = { fg = "#5AAFAD" },  -- 発光ライン
          ["@keyword.operator"] = { fg = "#5AAFAD" },  -- 発光ライン
          ["@keyword.return"] = { fg = "#5AAFAD" },    -- 発光ライン
          ["@keyword.import"] = { fg = "#5AAFAD" },    -- 発光ライン

          -- 文字列（グリッチ紫アクセント）
          ["@string"] = { fg = "#8C83A3" },            -- グリッチ紫

          -- 数値・真偽値（氷白）
          ["@number"] = { fg = "#D7E2E1" },            -- 氷白
          ["@boolean"] = { fg = "#D7E2E1" },           -- 氷白

          -- コメント（中間調）
          ["@comment"] = { fg = "#7E8A89" },

          -- 関数名（発光ハイライト）
          ["@function"] = { fg = "#96CBD1" },          -- 発光ハイライト（明るいシアン）
          ["@function.call"] = { fg = "#96CBD1" },     -- 発光ハイライト
          ["@function.method"] = { fg = "#96CBD1" },   -- 発光ハイライト
          ["@function.method.call"] = { fg = "#96CBD1" }, -- 発光ハイライト

          -- 変数・パラメータ（氷白）
          ["@variable"] = { fg = "#D7E2E1" },
          ["@variable.builtin"] = { fg = "#96CBD1" },  -- 発光ハイライト（組み込み変数）
          ["@parameter"] = { fg = "#D7E2E1" },

          -- 型名（発光ハイライト）
          ["@type"] = { fg = "#96CBD1" },              -- 発光ハイライト
          ["@type.builtin"] = { fg = "#96CBD1" },      -- 発光ハイライト

          -- プロパティ・属性（氷白）
          ["@property"] = { fg = "#D7E2E1" },
          ["@field"] = { fg = "#D7E2E1" },

          -- 定数（氷白）
          ["@constant"] = { fg = "#D7E2E1" },          -- 氷白
          ["@constant.builtin"] = { fg = "#D7E2E1" },  -- 氷白

          -- 演算子・句読点（中間調）
          ["@operator"] = { fg = "#7E8A89" },          -- 中間調
          ["@punctuation"] = { fg = "#7E8A89" },       -- 中間調
          ["@punctuation.bracket"] = { fg = "#7E8A89" }, -- 中間調
          ["@punctuation.delimiter"] = { fg = "#7E8A89" }, -- 中間調

          -- 名前空間・モジュール（氷白）
          ["@namespace"] = { fg = "#D7E2E1" },
          ["@module"] = { fg = "#D7E2E1" },

          -- タグ（HTML/XML用 - 発光ライン）
          ["@tag"] = { fg = "#5AAFAD" },               -- 発光ライン
          ["@tag.attribute"] = { fg = "#96CBD1" },     -- 発光ハイライト
          ["@tag.delimiter"] = { fg = "#7E8A89" },     -- 中間調

          -- TreesitterContext (スティッキースクロール) ハイライトグループ
          TreesitterContext = { bg = "#2A2F2E", fg = "#D7E2E1" }, -- 濃い影背景、前景
          TreesitterContextBottom = {
            underline = true,
            sp = "#2A2F2E" -- 濃い影の下線
          },
          TreesitterContextLineNumber = {
            bg = "#2A2F2E",
            fg = "#7E8A89" -- 中間グレーの行番号
          },
        },
      })
      -- setup()の後にcolorschemeを設定する必要がある
      vim.cmd.colorscheme("vercel")
      
      -- ターミナルカラー設定（モノクロ基調＋アクセント）
      vim.g.terminal_color_0 = "#0E1210"   -- Black: 背景
      vim.g.terminal_color_1 = "#B9C6C5"   -- Red: グレー
      vim.g.terminal_color_2 = "#9AA6A5"   -- Green: グレー
      vim.g.terminal_color_3 = "#C7D2D1"   -- Yellow: グレー
      vim.g.terminal_color_4 = "#7E8A89"   -- Blue: グレー
      vim.g.terminal_color_5 = "#8C83A3"   -- Magenta: グリッチ紫（アクセント）
      vim.g.terminal_color_6 = "#5AAFAD"   -- Cyan: ネオン青緑（アクセント）
      vim.g.terminal_color_7 = "#D7E2E1"   -- White: 前景
      vim.g.terminal_color_8 = "#2A2F2E"   -- Bright Black: 濃い影
      vim.g.terminal_color_9 = "#E6F1F0"   -- Bright Red: ニアホワイト
      vim.g.terminal_color_10 = "#CDD8D7"  -- Bright Green: ライトグレー
      vim.g.terminal_color_11 = "#F2FFFF"  -- Bright Yellow: ハイライト白
      vim.g.terminal_color_12 = "#AAB6B5"  -- Bright Blue: 明るいグレー
      vim.g.terminal_color_13 = "#B3A9D1"  -- Bright Magenta: ブライト紫（アクセント）
      vim.g.terminal_color_14 = "#96CBD1"  -- Bright Cyan: ブライト青緑（アクセント）
      vim.g.terminal_color_15 = "#FFFFFF"  -- Bright White: 純白
      
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
        vim.api.nvim_set_hl(0, "TabLine", { fg = "#7E8A89", bg = "none" })      -- 非アクティブタブ（中間グレー）
        vim.api.nvim_set_hl(0, "TabLineSel", { fg = "#D7E2E1", bg = "none", bold = true })  -- アクティブタブ（前景）
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
        vim.api.nvim_set_hl(0, "SnacksPickerListCursorLine", { bg = "#2A2F2E", fg = "#F2FFFF" })
        vim.api.nvim_set_hl(0, "SnacksPickerSelection", { bg = "#2A2F2E", fg = "#F2FFFF", bold = true })
        vim.api.nvim_set_hl(0, "SnacksPickerCursor", { bg = "#2A2F2E", fg = "#F2FFFF" })
        vim.api.nvim_set_hl(0, "SnacksPickerCursorLine", { bg = "#2A2F2E", fg = "#F2FFFF" })
        
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
        -- テスト結果のステータス色
        vim.api.nvim_set_hl(0, "NeotestPassed", { fg = "#5AAFAD" })     -- cyanアクセント: 成功
        vim.api.nvim_set_hl(0, "NeotestFailed", { fg = "#8C83A3" })     -- magentaアクセント: 失敗
        vim.api.nvim_set_hl(0, "NeotestRunning", { fg = "#D7E2E1" })    -- 前景: 実行中
        vim.api.nvim_set_hl(0, "NeotestSkipped", { fg = "#D7E2E1" })    -- 前景: スキップ
        vim.api.nvim_set_hl(0, "NeotestMarked", { fg = "#5AAFAD" })     -- cyanアクセント: マーク済み
        vim.api.nvim_set_hl(0, "NeotestWinSelect", { fg = "#5AAFAD" })  -- cyanアクセント: ウィンドウ選択

        -- UI要素のハイライト
        vim.api.nvim_set_hl(0, "NeotestAdapterName", { fg = "#7E8A89" }) -- 中間グレー: アダプタ名
        vim.api.nvim_set_hl(0, "NeotestBorder", { fg = "#2A2F2E" })      -- 濃い影: ボーダー
        vim.api.nvim_set_hl(0, "NeotestDir", { fg = "#7E8A89" })         -- 中間グレー: ディレクトリ
        vim.api.nvim_set_hl(0, "NeotestFile", { fg = "#D7E2E1" })        -- 前景: ファイル名
        vim.api.nvim_set_hl(0, "NeotestNamespace", { fg = "#5AAFAD" })   -- cyanアクセント: 名前空間
        vim.api.nvim_set_hl(0, "NeotestIndent", { fg = "#2A2F2E" })      -- 濃い影: インデント
        vim.api.nvim_set_hl(0, "NeotestExpandMarker", { fg = "#7E8A89" }) -- 中間グレー: 展開マーカー

        -- 追加のステータス色（強調バージョン）
        vim.api.nvim_set_hl(0, "NeotestFocused", { fg = "#5AAFAD", bold = true })  -- cyanアクセント: フォーカス
        vim.api.nvim_set_hl(0, "NeotestUnknown", { fg = "#7E8A89" })               -- 中間グレー: 不明
        vim.api.nvim_set_hl(0, "NeotestTarget", { fg = "#5AAFAD" })                -- cyanアクセント: ターゲット

        -- Scrollbar関連のハイライトグループ（透明化から除外）
        vim.api.nvim_set_hl(0, "ScrollbarHandle", { bg = "none", fg = "#2A2F2E" })
        vim.api.nvim_set_hl(0, "ScrollbarSearch", { fg = "#5AAFAD" })  -- cyanアクセント
        vim.api.nvim_set_hl(0, "ScrollbarError", { fg = "#8C83A3" })   -- magentaアクセント
        vim.api.nvim_set_hl(0, "ScrollbarWarn", { fg = "#96CBD1" })    -- bright cyanアクセント
        vim.api.nvim_set_hl(0, "ScrollbarInfo", { fg = "#96CBD1" })    -- bright cyanアクセント
        vim.api.nvim_set_hl(0, "ScrollbarHint", { fg = "#5AAFAD" })    -- cyanアクセント
        vim.api.nvim_set_hl(0, "ScrollbarMisc", { fg = "#7E8A89" })    -- 中間グレー
        vim.api.nvim_set_hl(0, "ScrollbarGitAdd", { fg = "#5AAFAD" })  -- cyanアクセント
        vim.api.nvim_set_hl(0, "ScrollbarGitChange", { fg = "#8C83A3" })  -- magentaアクセント
        vim.api.nvim_set_hl(0, "ScrollbarGitDelete", { fg = "#AAB6B5" })  -- 明るいグレー
        
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

