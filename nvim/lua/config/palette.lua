-- Abyssal Teal カラーパレット共有モジュール
-- TOMLソース: colors/abyssal-teal.toml
-- 単一ソース原則: このモジュールから全プラグイン設定が色を参照する

local M = {}

M.colors = {
  -- コアカラー
  bg = "#0B0C0C",           -- Main (Pure Black) [core.background]
  darkest_bg = "#111E16",   -- Deepest Shadow [core.darkest_bg]
  panel_bg = "#152A2B",     -- Dark Teal Panel [core.panel_bg]
  dark_shadow = "#1E1E24",  -- UI Shadow [core.ui_shadow]
  gutter_bg = "#13171A",    -- Gutter background [nvim.gutter_bg]
  border = "#4D8F9E",       -- UI Border [teals.border]
  mid_gray = "#525B65",     -- Slate Mid [blues_slates.slate_mid]
  light_gray = "#92A2AB",   -- Dim text [foregrounds.dim]

  -- Visibility improvements
  punctuation_gray = "#7A8599",  -- 句読点 [blues_slates.punctuation_gray]
  comment_gray = "#7A869A",      -- コメント [blues_slates.comment_gray]
  git_blame_gray = "#8A97AD",    -- Git blame [blues_slates.git_blame_gray]
  operator = "#64BBBE",     -- Clear Teal [teals.mid_bright]
  fg = "#CEF5F2",           -- Main Text [foregrounds.main]
  near_white = "#B1F4ED",   -- Brightest [foregrounds.bright]
  highlight_white = "#F2FFFF", -- Purest [ansi.bright_white]

  -- アクセントカラー
  cyan = "#6CD8D3",         -- Vibrant Teal [teals.bright]
  bright_cyan = "#9DDCD9",  -- Heading Cyan [foregrounds.heading]
  magenta = "#936997",      -- Glitch Purple (error accent)
  bright_magenta = "#B4B7CD", -- Cloud Slate [blues_slates.cloud_slate]

  -- 拡張セマンティック
  purple_accent = "#8A99BD", -- Muted Purple (Keyword) [purples.muted_purple]
  success = "#6AB9A8",       -- Success indicator [semantic.success]
  ocean_blue = "#326787",    -- Ocean Blue [blues_slates.ocean_blue]
  lavender = "#CED5E9",      -- Lavender (Constant) [purples.lavender]

  -- ANSI/Terminal (normal)
  ansi_red = "#936997",
  ansi_green = "#349594",
  ansi_yellow = "#CED5E9",
  ansi_blue = "#326787",

  -- ANSI Bright [ansi section] — Step 1: 未定義 terminal_color の修正
  bright_red = "#865F7B",      -- ansi.bright_red (Soft Magenta)
  bright_green = "#64BBBE",    -- ansi.bright_green (Clear Teal)
  bright_yellow = "#B1F4ED",   -- ansi.bright_yellow (Bright Text)
  bright_blue = "#A4ABCB",     -- ansi.bright_blue (Sky Slate)

  -- その他
  white = "#A4ABCB",        -- Sky Slate (Icon) [blues_slates.sky_slate]
  selection = "#64BBBE",    -- Clear Teal [core.selection_bg]
  selection_fg = "#0B0C0C", -- Background for contrast
  string = "#659D9E",       -- Base Teal [semantic.string]

  -- Diff背景カラー
  diff_add_bg = "#0D1F1F",      -- Diff Add背景 [nvim.diff_add_bg]
  diff_change_bg = "#1A141A",   -- Diff Change背景 [nvim.diff_change_bg]
  diff_delete_bg = "#151515",   -- Diff Delete背景 [nvim.diff_delete_bg]
}

return M
