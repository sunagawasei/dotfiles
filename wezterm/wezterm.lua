local wezterm = require("wezterm")
local config = wezterm.config_builder()

-- ==========================================
-- キーバインド設定
-- ==========================================
config.keys = require("keybinds").keys
config.key_tables = require("keybinds").key_tables
config.disable_default_key_bindings = true
config.leader = { key = "q", mods = "CTRL", timeout_milliseconds = 2000 }

-- ==========================================
-- 基本設定
-- ==========================================
config.automatically_reload_config = true
config.use_ime = true
config.ime_preedit_rendering = "Builtin"

-- macOSのIME関連設定
config.macos_forward_to_ime_modifier_mask = "SHIFT|CTRL"
config.send_composed_key_when_left_alt_is_pressed = false
config.send_composed_key_when_right_alt_is_pressed = false

-- ==========================================
-- フォント設定
-- ==========================================
config.font = wezterm.font_with_fallback({
	"Geist Mono",
	"Hiragino Kaku Gothic ProN",
	"Noto Sans CJK JP",
})
config.font_size = 16.0

-- ==========================================
-- ウィンドウ設定
-- ==========================================
-- ウィンドウサイズ
config.initial_cols = 41
config.initial_rows = 111

-- ウィンドウの外観
config.window_background_opacity = 0.85
config.macos_window_background_blur = 20
config.window_decorations = "RESIZE"
config.window_frame = {
	inactive_titlebar_bg = "none",
	active_titlebar_bg = "none",
}
config.window_background_gradient = {
	colors = { "#111111" }, -- Geist Accent 8 for consistency
}
config.window_padding = {
	left = 20,
	right = 20,
	top = 5,
	bottom = 5,
}

-- ==========================================
-- タブバー設定
-- ==========================================
config.show_new_tab_button_in_tab_bar = false
config.show_close_tab_button_in_tabs = false
config.hide_tab_bar_if_only_one_tab = false
config.tab_bar_at_bottom = false
config.use_fancy_tab_bar = false
config.tab_max_width = 100

-- ==========================================
-- カラーテーマ設定 (Geist Design System)
-- ==========================================
config.colors = {
	-- 基本色
	foreground = "#FFFFFF",
	background = "#111111", -- Accent 8 for dark background

	-- カーソル色
	cursor_bg = "#0070F3", -- Success default (Geist blue)
	cursor_fg = "#FFFFFF", -- White text on blue cursor
	cursor_border = "#0070F3", -- Success default

	-- 選択色
	selection_fg = "#FFFFFF", -- White text on dark background
	selection_bg = "#0761D1", -- Success dark for better contrast

	-- スクロールバー・分割線
	scrollbar_thumb = "#444444", -- Accent 6 for dark theme
	split = "#333333", -- Accent 7 for dark theme

	-- IMEの変換候補ウィンドウ用の設定
	compose_cursor = "#FF6B6B",

	-- ANSI色
	ansi = {
		"#111111", -- Black (Accent 8)
		"#EE0000", -- Red (Error)
		"#29BC9B", -- Green (Cyan Dark - more subdued)
		"#F5A623", -- Yellow (Warning)
		"#0070F3", -- Blue (Success)
		"#7928CA", -- Magenta (Violet)
		"#3291FF", -- Cyan (Success Light Blue - clearly distinguishable)
		"#FAFAFA", -- White (Accent 1)
	},
	brights = {
		"#666666", -- Bright Black (Accent 5)
		"#FF1A1A", -- Bright Red (Error Light)
		"#50E3C2", -- Bright Green (Cyan)
		"#F7B955", -- Bright Yellow (Warning Light)
		"#3291FF", -- Bright Blue (Success Light)
		"#F81CE5", -- Bright Magenta (Purple highlight)
		"#0070F3", -- Bright Cyan (Success Blue - bold and clear)
		"#FFFFFF", -- Bright White
	},

	-- コピーモード色
	copy_mode_active_highlight_bg = { Color = "#0070F3" },
	copy_mode_active_highlight_fg = { Color = "#FFFFFF" },
	copy_mode_inactive_highlight_bg = { Color = "#444444" },
	copy_mode_inactive_highlight_fg = { Color = "#FFFFFF" },

	-- クイックセレクト色
	quick_select_label_bg = { Color = "#7928CA" },
	quick_select_label_fg = { Color = "#FFFFFF" },
	quick_select_match_bg = { Color = "#0070F3" },
	quick_select_match_fg = { Color = "#FFFFFF" },

	-- タブバー
	tab_bar = {
		inactive_tab_edge = "none",
	},
}

return config
