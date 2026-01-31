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
-- Unicode文字幅設定
-- ==========================================
-- Unicode仕様バージョンを最新（14）に設定
config.unicode_version = 14
-- East Asian Ambiguous Width文字を半角として扱う（Box Drawing文字の表示崩れを防ぐ）
config.treat_east_asian_ambiguous_width_as_wide = false
-- グリフのオーバーフローを禁止（整列を保つ）
config.allow_square_glyphs_to_overflow_width = "Never"
-- すべての文字の幅を統一（Box Drawing文字と日本語文字の表示を揃える）
config.cell_width = 1.0

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
	{ family = "GeistMono NF" }, -- メインフォント（Nerd Font版・略名）
	{ family = "GeistMono NF", assume_emoji_presentation = true }, -- Nerd Fontアイコン用（2回目の定義）
	{ family = "IBM Plex Sans JP" }, -- 日本語フォント
	{ family = "Hiragino Kaku Gothic ProN" }, -- 日本語フォント
})
-- フォントサイズを16ポイントに設定
config.font_size = 14.0
-- フォント間の高さ調整を無効化（メトリクスの一貫性を保つ）
config.use_cap_height_to_scale_fallback_fonts = false
-- 行間を少し広げて文字の重なりを防ぐ
config.line_height = 1.1

-- ==========================================
-- ウィンドウ設定
-- ==========================================

-- ウィンドウの外観
-- macOSネイティブボタン付きのタイトルバー（最小化・最大化・閉じる）
config.window_decorations = "INTEGRATED_BUTTONS|RESIZE|MACOS_FORCE_ENABLE_SHADOW"
-- 統合タイトルボタンのスタイルをmacOSネイティブに設定
config.integrated_title_button_style = "MacOsNative"
-- ウィンドウ背景の不透明度（95%）
config.window_background_opacity = 0.95
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
	-- タイトルバーのフォントサイズ（通常のフォントサイズと同期）
	font_size = config.font_size,
}
-- ウィンドウ背景のグラデーション設定（モノクロ背景）
config.window_background_gradient = {
	colors = { "#1A201E" },
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
-- タブの最大幅を60文字に設定
config.tab_max_width = 60
-- タブの閉じるボタンを非表示
config.show_close_tab_button_in_tabs = false

-- ==========================================
-- カラーテーマ設定
-- ==========================================
-- 非アクティブペインを暗くする設定（HSB色空間で調整）
config.inactive_pane_hsb = {
	saturation = 0.2, -- 彩度を80%下げる
	brightness = 0.2, -- 明度を20%にする
}

config.colors = {
	-- 基本色
	foreground = "#D7E2E1", -- 前景色（テキスト色）：モノクロ
	background = "#1A201E", -- 背景色：モノクロ背景

	-- カーソル色
	cursor_bg = "#F2FFFF", -- カーソル背景：ハイライト白
	cursor_fg = "#1A201E", -- カーソル前景：背景色
	cursor_border = "#F2FFFF", -- カーソル境界線：ハイライト白

	-- 選択色
	selection_fg = "#1A201E", -- 選択時のテキスト色：背景色（反転）
	selection_bg = "#5AAFAD", -- 選択時の背景色：アクセントcyan

	-- スクロールバー・分割線
	scrollbar_thumb = "#2A2F2E", -- スクロールバーのつまみ色：濃い影
	split = "#7E8A89", -- ペイン分割線の色：中間グレー（視認性向上）

	-- ANSI色（ターミナルの標準16色）
	ansi = {
		"#1A201E", -- black: 背景色
		"#B9C6C5", -- red → gray
		"#9AA6A5", -- green → gray
		"#C7D2D1", -- yellow → gray
		"#7E8A89", -- blue → gray
		"#8C83A3", -- magenta: グリッチ紫（アクセント）
		"#5AAFAD", -- cyan: ネオン青緑（アクセント）
		"#D7E2E1", -- white: 前景色
	},
	-- 明るいANSI色（太字や高輝度表示用）
	brights = {
		"#2A2F2E", -- bright black: 濃い影
		"#E6F1F0", -- bright red → near white
		"#CDD8D7", -- bright green → light gray
		"#F2FFFF", -- bright yellow → white
		"#AAB6B5", -- bright blue → light gray
		"#B3A9D1", -- bright magenta: ブライト紫（アクセント）
		"#96CBD1", -- bright cyan: ブライト青緑（アクセント）
		"#F2FFFF", -- bright white: ハイライト白
	},

	-- コピーモード色設定
	copy_mode_active_highlight_bg = { Color = "#5AAFAD" }, -- アクティブハイライト背景：アクセントcyan
	copy_mode_active_highlight_fg = { Color = "#1A201E" }, -- アクティブハイライト前景：背景色
	copy_mode_inactive_highlight_bg = { Color = "#2A2F2E" }, -- 非アクティブハイライト背景：濃い影
	copy_mode_inactive_highlight_fg = { Color = "#D7E2E1" }, -- 非アクティブハイライト前景：前景色

	-- クイックセレクト色設定
	quick_select_label_bg = { Color = "#8C83A3" }, -- ラベル背景：アクセントmagenta
	quick_select_label_fg = { Color = "#F2FFFF" }, -- ラベル前景：ハイライト白
	quick_select_match_bg = { Color = "#5AAFAD" }, -- マッチ背景：アクセントcyan
	quick_select_match_fg = { Color = "#1A201E" }, -- マッチ前景：背景色

	-- タブバー設定
	tab_bar = {
		-- タブバーの背景色
		background = "#1A201E",
		-- アクティブタブのスタイル
		active_tab = {
			bg_color = "#3A4543", -- より明るくして視認性向上
			fg_color = "#F2FFFF", -- ハイライト白
			intensity = "Bold", -- 太字
		},
		-- 非アクティブタブのスタイル
		inactive_tab = {
			bg_color = "#1A201E", -- 背景色
			fg_color = "#7E8A89", -- 現在: #AAB6B5 → 少し暗く
		},
		-- 非アクティブタブのホバー時スタイル
		inactive_tab_hover = {
			bg_color = "#2A2F2E", -- 濃い影
			fg_color = "#D7E2E1", -- 前景色
		},
		-- 新規タブボタンのスタイル
		new_tab = {
			bg_color = "#1A201E",
			fg_color = "#7E8A89",
		},
		-- 新規タブボタンのホバー時スタイル
		new_tab_hover = {
			bg_color = "#2A2F2E",
			fg_color = "#D7E2E1",
		},
		-- 非アクティブタブの境界線を追加
		inactive_tab_edge = "#3A3F3E", -- 現在: "none" → 境界線追加
	},
}

-- ==========================================
-- タブのカスタマイズ
-- ==========================================
-- タブIDから決定論的に色を生成（タブ識別用）
local function tab_id_to_color(tab_id)
	local colors = {
		"#5AAFAD", -- cyan（アクセント）
		"#8C83A3", -- magenta（アクセント）
		"#D7E2E1", -- 前景（モノクロ）
		"#7E8A89", -- 中間グレー（モノクロ）
	}
	return colors[(tab_id % #colors) + 1]
end

-- Claude Code状態に応じたインジケータを返す
local function get_claude_status(tab_title)
	if tab_title:find("%[完了%]") then
		return { icon = "✓", color = "#5AAFAD" } -- cyan
	elseif tab_title:find("%[許可待ち%]") then
		return { icon = "!", color = "#8C83A3" } -- magenta
	elseif tab_title:find("%[入力待ち%]") then
		return { icon = "●", color = "#96CBD1" } -- bright cyan
	elseif tab_title:find("%[実行中%]") then
		return { icon = "▶", color = "#D7E2E1" } -- foreground
	end
	return nil
end

-- cwdを取得して短縮表示するヘルパー関数
local function get_short_cwd(tab, max_len)
	local cwd_uri = tab.active_pane.current_working_dir
	if not cwd_uri or not cwd_uri.file_path then
		return nil
	end
	local path = cwd_uri.file_path
	local home = os.getenv("HOME")
	local title
	if path == home then
		title = "~"
	elseif path:sub(1, #home) == home then
		title = "~" .. path:sub(#home + 1)
	else
		title = path
	end
	if #title > max_len then
		title = title:sub(1, max_len) .. "…"
	end
	return title
end

-- タブタイトルのカスタマイズ処理（相対パスを表示）
wezterm.on("format-tab-title", function(tab)
	-- Claude Code状態を確認（tab.tab_titleを優先）
	local tab_title = tab.tab_title or ""
	local claude_status = get_claude_status(tab_title)

	-- Claude状態がある場合は、状態表示を優先
	if claude_status then
		-- cwdを取得（最大30文字）
		local cwd = get_short_cwd(tab, 30) or "~"

		if tab.is_active then
			local line_color = tab_id_to_color(tab.tab_id)
			return {
				{ Foreground = { Color = line_color } },
				{ Text = "█ " },
				{ Foreground = { Color = claude_status.color } },
				{ Text = claude_status.icon .. " " },
				{ Foreground = { Color = "#FFFFFF" } },
				{ Text = cwd .. " " },
			}
		else
			return {
				{ Text = "  " },
				{ Foreground = { Color = claude_status.color } },
				{ Text = claude_status.icon .. " " },
				{ Foreground = { Color = "#7E8A89" } },
				{ Text = cwd .. " " },
			}
		end
	end

	-- Claude状態がない場合は、既存の処理（cwd表示）
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
	if #title > 50 then
		title = title:sub(1, 50) .. "…"
	end

	-- アクティブタブかどうかで表示を分ける
	if tab.is_active then
		-- アクティブタブ: 色付き縦線インジケーター（タブIDから決定論的に色を選択）
		local color = tab_id_to_color(tab.tab_id)
		return {
			{ Foreground = { Color = color } },
			{ Text = "█" },
			{ Foreground = { Color = "#FFFFFF" } },
			{ Text = " " .. title .. " " },
		}
	else
		-- 非アクティブタブ: インジケーターなし
		return "  " .. title .. "  "
	end
end)

-- ==========================================
-- 中央寄せ機能設定 (no-neck-pain風)
-- ==========================================
-- 中央寄せ機能の設定
local centering_config = {
	enabled = false, -- 中央寄せ機能の有効/無効
	content_width_ratio = 0.8, -- コンテンツ幅の比率（80%）
	fullscreen_only = false, -- フルスクリーン時のみ有効化するか
	min_width_for_centering = 1200, -- 中央寄せを適用する最小ウィンドウ幅
	max_content_width = 1400, -- コンテンツの最大幅（px）
	current_preset = "normal", -- 現在のプリセット
	-- プリセット設定
	presets = {
		narrow = { ratio = 0.5, max_width = 1000 }, -- 狭い（50%、最大1000px）
		normal = { ratio = 0.8, max_width = 1400 }, -- 標準（80%、最大1400px）
		wide = { ratio = 0.75, max_width = 1800 }, -- 広い（75%、最大1800px）
		ultra_wide = { ratio = 0.85, max_width = 2200 }, -- 超広い（85%、最大2200px）
	},
}

-- ウィンドウサイズ変更時の中央寄せ処理
wezterm.on("window-resized", function(window, pane)
	-- 中央寄せが無効の場合は何もしない
	if not centering_config.enabled then
		return
	end

	local window_dims = window:get_dimensions()
	local overrides = window:get_config_overrides() or {}

	-- フルスクリーン限定モードかつ非フルスクリーンの場合は何もしない
	if centering_config.fullscreen_only and not window_dims.is_full_screen then
		overrides.window_padding = nil
		window:set_config_overrides(overrides)
		return
	end

	-- ウィンドウ幅が最小幅未満の場合は中央寄せを無効化
	if window_dims.pixel_width < centering_config.min_width_for_centering then
		overrides.window_padding = {
			left = 5,
			right = 5,
			top = 0,
			bottom = 5,
		}
		window:set_config_overrides(overrides)
		return
	end

	-- プリセット設定を取得
	local preset = centering_config.presets[centering_config.current_preset]
	local effective_ratio = preset and preset.ratio or centering_config.content_width_ratio
	local max_width = preset and preset.max_width or centering_config.max_content_width

	-- コンテンツ幅を計算（比率による制限と最大幅による制限の小さい方を採用）
	local ratio_based_width = window_dims.pixel_width * effective_ratio
	local content_width = math.min(ratio_based_width, max_width)

	-- パディングの合計幅を計算（残りの幅を左右に均等分配）
	local padding_total = window_dims.pixel_width - content_width
	local padding_each = math.floor(padding_total / 2)

	-- 最小パディング（5px）を保証
	local padding_left = math.max(padding_each, 5)
	local padding_right = math.max(padding_each, 5)

	local new_padding = {
		left = padding_left,
		right = padding_right,
		top = 0, -- 上部パディングは変更しない
		bottom = 5, -- 下部パディングは元の設定を維持
	}

	overrides.window_padding = new_padding
	window:set_config_overrides(overrides)
end)

-- 中央寄せ機能のトグル関数
local function toggle_centering(window, pane)
	centering_config.enabled = not centering_config.enabled

	-- 現在のウィンドウに対してのみ設定を適用
	if window then
		local overrides = window:get_config_overrides() or {}

		if centering_config.enabled then
			-- 中央寄せ有効化時: window-resizedイベントを発火
			wezterm.emit("window-resized", window, pane)
		else
			-- 中央寄せ無効化時: パディングをデフォルトに戻す
			overrides.window_padding = {
				left = 5,
				right = 5,
				top = 0,
				bottom = 5,
			}
			window:set_config_overrides(overrides)
		end
	end

	-- 状態をステータスに表示
	local status = centering_config.enabled and "ON" or "OFF"
	wezterm.log_info("Center mode: " .. status)
end

-- プリセット切り替え関数
local function cycle_centering_preset(window, pane)
	local preset_order = { "narrow", "normal", "wide", "ultra_wide" }
	local current_index = 1

	-- 現在のプリセットのインデックスを見つける
	for i, preset_name in ipairs(preset_order) do
		if centering_config.current_preset == preset_name then
			current_index = i
			break
		end
	end

	-- 次のプリセットに切り替え（循環）
	local next_index = (current_index % #preset_order) + 1
	centering_config.current_preset = preset_order[next_index]

	-- 中央寄せが有効な場合は再適用
	if centering_config.enabled and window then
		wezterm.emit("window-resized", window, pane)
	end

	-- 状態をログに表示
	local preset = centering_config.presets[centering_config.current_preset]
	wezterm.log_info(
		string.format(
			"Center preset: %s (ratio: %.0f%%, max: %dpx)",
			centering_config.current_preset,
			preset.ratio * 100,
			preset.max_width
		)
	)
end

-- 幅比率を調整する関数
local function adjust_centering_width(window, pane, delta)
	if not centering_config.enabled then
		return
	end

	-- カスタムプリセットを作成または更新
	if centering_config.current_preset ~= "custom" then
		-- 現在のプリセットをベースにカスタムプリセットを作成
		local current_preset = centering_config.presets[centering_config.current_preset]
		centering_config.presets.custom = {
			ratio = current_preset.ratio,
			max_width = current_preset.max_width,
		}
		centering_config.current_preset = "custom"
	end

	-- 比率を調整（0.3から0.95の範囲で制限）
	local custom_preset = centering_config.presets.custom
	custom_preset.ratio = math.max(0.3, math.min(0.95, custom_preset.ratio + delta))

	-- 設定を再適用
	if window then
		wezterm.emit("window-resized", window, pane)
	end

	-- 状態をログに表示
	wezterm.log_info(string.format("Center width: %.0f%% (custom)", custom_preset.ratio * 100))
end

-- 最大幅を調整する関数
local function adjust_centering_max_width(window, pane, delta)
	if not centering_config.enabled then
		return
	end

	-- カスタムプリセットを作成または更新
	if centering_config.current_preset ~= "custom" then
		local current_preset = centering_config.presets[centering_config.current_preset]
		centering_config.presets.custom = {
			ratio = current_preset.ratio,
			max_width = current_preset.max_width,
		}
		centering_config.current_preset = "custom"
	end

	-- 最大幅を調整（800px から 3000px の範囲で制限）
	local custom_preset = centering_config.presets.custom
	custom_preset.max_width = math.max(800, math.min(3000, custom_preset.max_width + delta))

	-- 設定を再適用
	if window then
		wezterm.emit("window-resized", window, pane)
	end

	-- 状態をログに表示
	wezterm.log_info(string.format("Center max width: %dpx (custom)", custom_preset.max_width))
end

-- フルスクリーンモードの切り替え
local function toggle_centering_fullscreen_only(window, pane)
	centering_config.fullscreen_only = not centering_config.fullscreen_only

	-- 中央寄せが有効な場合は再適用
	if centering_config.enabled and window then
		wezterm.emit("window-resized", window, pane)
	end

	local status = centering_config.fullscreen_only and "ON" or "OFF"
	wezterm.log_info("Center fullscreen-only mode: " .. status)
end

-- グローバル関数として公開（キーバインドから使用するため）
wezterm.toggle_centering = toggle_centering
wezterm.cycle_centering_preset = cycle_centering_preset
wezterm.adjust_centering_width = adjust_centering_width
wezterm.adjust_centering_max_width = adjust_centering_max_width
wezterm.toggle_centering_fullscreen_only = toggle_centering_fullscreen_only

-- 設定をエクスポート
return config

