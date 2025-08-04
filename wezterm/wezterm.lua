local wezterm = require("wezterm")
local mux = wezterm.mux
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
config.use_ime = false

-- ==========================================
-- 起動時のウィンドウ配置設定
-- ==========================================
wezterm.on("gui-startup", function(cmd)
	local screens = wezterm.gui.screens()
	local active_screen = screens.active

	-- アクティブスクリーンの65%の幅、80%の高さで設定
	local ratio_width = 0.65
	local ratio_height = 0.80
	local width = active_screen.width * ratio_width
	local height = active_screen.height * ratio_height

	-- ウィンドウをアクティブスクリーンの中央に配置
	local tab, pane, window = mux.spawn_window(cmd or {
		position = {
			x = active_screen.x + (active_screen.width - width) / 2,
			y = active_screen.y + (active_screen.height - height) / 2,
			origin = "ActiveScreen",
		},
	})

	-- ウィンドウサイズを設定
	window:gui_window():set_inner_size(width, height)
end)

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

-- カーソルスタイル
config.default_cursor_style = "BlinkingBar"

-- ウィンドウの外観
config.window_decorations = "RESIZE"
config.window_background_opacity = 0.75
config.macos_window_background_blur = 15
config.window_frame = {
	inactive_titlebar_bg = "none",
	active_titlebar_bg = "none",
	font = wezterm.font("Geist Mono"),
	font_size = 14,
}
config.window_background_gradient = {
	colors = { "#000000" },
}
config.window_padding = {
	left = 5,
	right = 5,
	top = 0,
	bottom = 5,
}
config.show_new_tab_button_in_tab_bar = false
config.show_close_tab_button_in_tabs = false

-- ==========================================
-- カラーテーマ設定 (Geist Design System)
-- ==========================================
-- 非アクティブペインを暗くする設定
config.inactive_pane_hsb = {
	saturation = 0.6, -- 彩度を40%下げる
	brightness = 0.5, -- 明度を半分にして暗くする
}

config.colors = {
	-- 基本色
	foreground = "#FFFFFF",
	background = "#0a0a0a", -- Ghostty background color

	-- カーソル色
	cursor_bg = "#0070F3", -- Success default (Geist blue)
	cursor_fg = "#FFFFFF", -- White text on blue cursor
	cursor_border = "#0070F3", -- Success default

	-- 選択色
	selection_fg = "#FFFFFF", -- White text on dark background
	selection_bg = "#334155", -- Gray 8 - より暗く読みやすい選択色

	-- スクロールバー・分割線
	scrollbar_thumb = "#444444", -- Accent 6 for dark theme
	split = "#333333", -- Accent 7 for dark theme

	-- ANSI色
	ansi = {
		"#111111", -- Black (Accent 8)
		"#EE0000", -- Red (Error)
		"#29BC9B", -- Green (Cyan Dark - more subdued)
		"#92400E", -- Yellow (Amber 9 - 暗めの黄色)
		"#0070F3", -- Blue (Success)
		"#7928CA", -- Magenta (Violet)
		"#3291FF", -- Cyan (Success Light Blue - clearly distinguishable)
		"#FAFAFA", -- White (Accent 1)
	},
	brights = {
		"#666666", -- Bright Black (Accent 5)
		"#FF1A1A", -- Bright Red (Error Light)
		"#50E3C2", -- Bright Green (Cyan)
		"#D97706", -- Bright Yellow (Amber 7 - 中間の明るさ)
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
	-- tab_bar = {
	-- 	inactive_tab_edge = "none",
	-- },
}

return config

