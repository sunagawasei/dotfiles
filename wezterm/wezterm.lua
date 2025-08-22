-- WezTermモジュールを読み込む
local wezterm = require("wezterm")
-- マルチプレクサ（ウィンドウ・タブ管理）モジュールを読み込む
local mux = wezterm.mux
-- 設定ビルダーを初期化（型安全な設定を作成）
local config = wezterm.config_builder()

-- ==========================================
-- キーバインド設定
-- ==========================================
-- keybinds.luaファイルからキー設定を読み込む
config.keys = require("keybinds").keys
-- キーテーブル（モード別のキー設定）を読み込む
config.key_tables = require("keybinds").key_tables
-- デフォルトのキーバインドを無効化（カスタム設定のみ使用）
config.disable_default_key_bindings = true
-- リーダーキーをCtrl+qに設定（tmuxのプレフィックスキーのような役割）
config.leader = { key = "q", mods = "CTRL", timeout_milliseconds = 1000 }

-- ==========================================
-- 基本設定
-- ==========================================
-- 設定ファイルの変更を自動的に検知して再読み込み
config.automatically_reload_config = true
-- IME（日本語入力）を無効化
config.use_ime = false
-- WebGPUレンダリングエンジンを使用（最高のパフォーマンス）
config.front_end = "WebGpu"
-- スクロールバックバッファを3500行に設定
config.scrollback_lines = 3500

-- ==========================================
-- 起動時のウィンドウ配置設定
-- ==========================================
-- GUI起動時のイベントハンドラを設定
wezterm.on("gui-startup", function(cmd)
	-- 利用可能なスクリーン情報を取得
	local screens = wezterm.gui.screens()
	-- 現在アクティブなスクリーンを取得
	local active_screen = screens.active

	-- アクティブスクリーンの85%の幅、90%の高さで設定
	local ratio_width = 0.65
	local ratio_height = 1
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
-- フォントの設定（フォールバック機能付き）
config.font = wezterm.font_with_fallback({
	"Geist Mono", -- メインフォント（Vercelのモノスペースフォント）
	"IBM Plex Sans JP", -- 日本語フォント（メイン）
	"Hiragino Kaku Gothic ProN", -- 日本語フォント（フォールバック）
})
-- フォントサイズを16ポイントに設定
config.font_size = 16.0

-- ==========================================
-- ウィンドウ設定
-- ==========================================

-- ウィンドウの外観
-- macOSネイティブボタン付きのタイトルバー（最小化・最大化・閉じる）
config.window_decorations = "INTEGRATED_BUTTONS|RESIZE|MACOS_FORCE_ENABLE_SHADOW"
-- 統合タイトルボタンのスタイルをmacOSネイティブに設定
config.integrated_title_button_style = "MacOsNative"
-- ウィンドウ背景の不透明度（75%）
config.window_background_opacity = 0.85
-- macOSのぼかし効果の強度
config.macos_window_background_blur = 30
-- ウィンドウフレームの設定
config.window_frame = {
	-- 非アクティブ時のタイトルバー背景（透明）
	inactive_titlebar_bg = "none",
	-- アクティブ時のタイトルバー背景（透明）
	active_titlebar_bg = "none",
	-- タイトルバーのフォント
	font = wezterm.font("Geist Mono"),
	-- タイトルバーのフォントサイズ
	font_size = 14,
}
-- ウィンドウ背景のグラデーション設定（単色の黒）
config.window_background_gradient = {
	colors = { "#000000" },
}
-- ウィンドウ内側の余白設定
config.window_padding = {
	left = 5, -- 左余白
	right = 5, -- 右余白
	top = 0, -- 上余白（タブバーに密着）
	bottom = 5, -- 下余白
}
-- タブバーを有効化
config.enable_tab_bar = true
-- タブバーの新規タブボタンを非表示
config.show_new_tab_button_in_tab_bar = false
-- タブインデックス番号を非表示
config.show_tab_index_in_tab_bar = false
-- タブの最大幅を20文字に設定
config.tab_max_width = 20
-- タブの閉じるボタンを非表示
config.show_close_tab_button_in_tabs = false

-- ==========================================
-- カラーテーマ設定 (Geist Design System)
-- ==========================================
-- 非アクティブペインを暗くする設定（HSB色空間で調整）
config.inactive_pane_hsb = {
	saturation = 0.7, -- 彩度を30%下げる（色を薄くする）
	brightness = 0.6, -- 明度を60%にして暗くする（少し明るめ）
}

config.colors = {
	-- 基本色
	foreground = "#FFFFFF", -- 前景色（テキスト色）：白
	background = "#0a0a0a", -- 背景色：Ghosttyと同じ暗い黒

	-- カーソル色
	cursor_bg = "#F8F8F8", -- カーソル背景：オフホワイト
	cursor_fg = "#1A1A1A", -- カーソル前景：ダークグレー
	cursor_border = "#666666", -- カーソル境界線：ミディアムグレー

	-- 選択色
	selection_fg = "#FFFFFF", -- 選択時のテキスト色：白
	selection_bg = "#2563EB", -- 選択時の背景色：Blue 7（より鮮明な青）

	-- スクロールバー・分割線
	scrollbar_thumb = "#444444", -- スクロールバーのつまみ色：Accent 6
	split = "#333333", -- ペイン分割線の色：Accent 7

	-- ANSI色（ターミナルの標準16色）
	ansi = {
		"#333333", -- 黒：Accent 7（より明るいグレー）
		"#DC2626", -- 赤：Red 7（落ち着いた赤）
		"#16A34A", -- 緑：Green 7（落ち着いた緑）
		"#D97706", -- 黄：Amber 7（落ち着いた琥珀）
		"#2563EB", -- 青：Blue 7（落ち着いた青）
		"#9333EA", -- マゼンタ：Purple 7（落ち着いた紫）
		"#0D9488", -- シアン：Teal 7（落ち着いたティール）
		"#FAFAFA", -- 白：Accent 1
	},
	-- 明るいANSI色（太字や高輝度表示用）
	brights = {
		"#A3A3A3", -- 明るい黒：Color 6（zsh-autosuggestions用に明るく調整）
		"#B91C1C", -- 明るい赤：Red 8（より暗い赤）
		"#15803D", -- 明るい緑：Green 8（より暗い緑）
		"#B45309", -- 明るい黄：Amber 8（より暗い琥珀）
		"#1D4ED8", -- 明るい青：Blue 8（より暗い青）
		"#7C3AED", -- 明るいマゼンタ：Purple 8（より暗い紫）
		"#0F766E", -- 明るいシアン：Teal 8（より暗いティール）
		"#FFFFFF", -- 明るい白：純白
	},

	-- コピーモード色設定
	copy_mode_active_highlight_bg = { Color = "#0070F3" }, -- アクティブハイライト背景：Vercel青
	copy_mode_active_highlight_fg = { Color = "#FFFFFF" }, -- アクティブハイライト前景：白
	copy_mode_inactive_highlight_bg = { Color = "#444444" }, -- 非アクティブハイライト背景：グレー
	copy_mode_inactive_highlight_fg = { Color = "#FFFFFF" }, -- 非アクティブハイライト前景：白

	-- クイックセレクト色設定
	quick_select_label_bg = { Color = "#7928CA" }, -- ラベル背景：紫
	quick_select_label_fg = { Color = "#FFFFFF" }, -- ラベル前景：白
	quick_select_match_bg = { Color = "#0070F3" }, -- マッチ背景：Vercel青
	quick_select_match_fg = { Color = "#FFFFFF" }, -- マッチ前景：白

	-- タブバー設定
	tab_bar = {
		-- タブバーの背景色
		background = "#0a0a0a",
		-- アクティブタブのスタイル
		active_tab = {
			bg_color = "#1A1A1A", -- 少し明るい背景
			fg_color = "#FFFFFF", -- 白いテキスト
			intensity = "Bold", -- 太字
		},
		-- 非アクティブタブのスタイル
		inactive_tab = {
			bg_color = "#0a0a0a", -- 暗い背景
			fg_color = "#666666", -- グレーのテキスト
		},
		-- 非アクティブタブのホバー時スタイル
		inactive_tab_hover = {
			bg_color = "#111111", -- 少し明るい背景
			fg_color = "#A3A3A3", -- 明るいグレー
		},
		-- 新規タブボタンのスタイル
		new_tab = {
			bg_color = "#0a0a0a",
			fg_color = "#666666",
		},
		-- 新規タブボタンのホバー時スタイル
		new_tab_hover = {
			bg_color = "#111111",
			fg_color = "#A3A3A3",
		},
		-- 非アクティブタブの境界線を非表示
		inactive_tab_edge = "none",
	},
}

-- ==========================================
-- タブのカスタマイズ
-- ==========================================
-- タブタイトルのカスタマイズ処理（相対パスを表示）
wezterm.on("format-tab-title", function(tab)
	-- 現在の作業ディレクトリのURIを取得
	local cwd_uri = tab.active_pane.current_working_dir
	local title = ""
	
	if cwd_uri and cwd_uri.file_path then
		local path = cwd_uri.file_path
		local home = os.getenv("HOME")
		
		-- ホームディレクトリの場合は「~」を表示
		if path == home then
			title = "~"
		-- ホームディレクトリ配下の場合は相対パスを表示
		elseif path:sub(1, #home) == home then
			title = "~" .. path:sub(#home + 1)
		-- それ以外はフルパスを表示
		else
			title = path
		end
	else
		-- パスが取得できない場合はペインのタイトルを使用
		title = tab.active_pane.title
	end

	-- タイトルの長さを制限
	if #title > 16 then
		title = title:sub(1, 16) .. "…"
	end

	-- アクティブタブかどうかで表示を分ける
	if tab.is_active then
		-- アクティブタブ: 明るくて目立つ表示
		return "  " .. title .. "  "
	else
		-- 非アクティブタブ: 控えめな表示
		return "  " .. title .. "  "
	end
end)

-- 設定をエクスポート
return config
