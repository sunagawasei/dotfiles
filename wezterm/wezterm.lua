-- WezTermモジュールを読み込む
local wezterm = require("wezterm")
local colors = require("colors")
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
-- Pure prompt (sindresorhus/pure) のバックグラウンドgit fetchによるpane閉じ警告を抑制
config.skip_close_confirmation_for_processes_named = {
	"bash", "sh", "zsh", "fish", "tmux", "nu", "cmd.exe", "pwsh.exe", "powershell.exe",
	"git",
}
-- デフォルト作業ディレクトリを~/.configに設定
config.default_cwd = wezterm.home_dir .. "/.config"

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
		cwd = wezterm.home_dir .. "/.config",
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
	{ family = "GeistMono NF", assume_emoji_presentation = true }, -- Nerd Fontアイコン用
	{ family = "Osaka" }, -- 日本語フォント
	{ family = "Hiragino Kaku Gothic ProN" }, -- 日本語フォント（フォールバック）
})
-- フォントサイズ（OsakaMonoビットマップサイズ15pxに合わせる）
config.font_size = 15.0
-- フォント間の高さ調整を無効化（メトリクスの一貫性を保つ）
config.use_cap_height_to_scale_fallback_fonts = false
-- 行間を少し広げて文字の重なりを防ぐ
config.line_height = 1.1

-- ==========================================
-- Geist Pixel Font Rules (Semantic Pixelation)
-- ==========================================
-- Maps italic styling to pixel fonts for UI-only rendering
-- Terminal content remains in Geist Mono for readability
-- NOTE: Geist Pixel Squareフォントが未インストールのため一時的に無効化
--[[
config.font_rules = {
	-- When italic is requested, use Geist Pixel instead
	{
		intensity = "Normal",
		italic = true,
		font = wezterm.font_with_fallback({
			{
				family = "Geist Pixel Square",
				-- WOFF2 format with absolute path
				-- WezTerm may support WOFF2 directly
			},
			{ family = "IBM Plex Sans JP" }, -- CJK fallback
			{ family = "GeistMono NF" }, -- Ultimate fallback
		}),
	},
	-- Bold + Italic → Bold Geist Pixel
	{
		intensity = "Bold",
		italic = true,
		font = wezterm.font_with_fallback({
			{
				family = "Geist Pixel Square",
				weight = "Bold",
			},
			{ family = "IBM Plex Sans JP", weight = "Bold" },
			{ family = "GeistMono NF", weight = "Bold" },
		}),
	},
}
--]]

-- ==========================================
-- ウィンドウ設定
-- ==========================================

-- ウィンドウの外観
-- macOSネイティブボタンを非表示にし、リサイズ機能だけを残す
config.window_decorations = "RESIZE"
-- ウィンドウ背景の不透明度（88%）
config.window_background_opacity = 0.88
-- macOSのぼかし効果の強度
config.macos_window_background_blur = 30
-- ウィンドウフレームの設定
config.window_frame = {
	-- 非アクティブ時のタイトルバー背景（透明）
	inactive_titlebar_bg = "none",
	-- アクティブ時のタイトルバー背景（透明）
	active_titlebar_bg = "none",
	-- タブバーのフォント（本体と同じGeistMono NF。視認性のため太字に）
	font = wezterm.font_with_fallback({
		{ family = "GeistMono NF", weight = "Bold" },
		{ family = "GeistMono NF", weight = "Bold", assume_emoji_presentation = true },
		{ family = "Osaka" }, -- 日本語フォント（本体と同じフォールバック）
		{ family = "Hiragino Kaku Gothic ProN" },
	}),
	-- タブバーのフォントサイズ（本体15.0より大きい17.0。リテラル固定なので本体 config.font_size とは連動しない）
	font_size = 17.0,
}
-- ウィンドウ背景のグラデーション設定（モノクロ背景）
config.window_background_gradient = {
	colors = { colors.core.background },
}
-- ウィンドウ内側の余白設定
config.window_padding = {
	left = 12, -- 左余白
	right = 12, -- 右余白
	top = 10, -- 上余白（ボーダレスウィンドウのため余裕を持たせる）
	bottom = 0, -- 下余白（タブバーに密着）
}
-- タブバーを有効化
config.enable_tab_bar = true
-- タブバーを下に配置
config.tab_bar_at_bottom = true
-- fancyタブバーを使用（タブバー専用フォントサイズ指定のため。format-tab-titleは引き続き有効）
config.use_fancy_tab_bar = true
-- タブが1つの時はタブバーを非表示
config.hide_tab_bar_if_only_one_tab = true
-- タブバーの新規タブボタンを非表示
config.show_new_tab_button_in_tab_bar = false
-- タブインデックス番号を非表示
config.show_tab_index_in_tab_bar = false
-- タブの最大幅（見切れ防止のため広め）
config.tab_max_width = 100
-- タブの閉じるボタンを非表示
config.show_close_tab_button_in_tabs = false

-- ==========================================
-- カラーテーマ設定
-- ==========================================
-- 非アクティブペインを暗くする設定（HSB色空間で調整）
config.inactive_pane_hsb = {
	saturation = 0.9, -- 彩度を10%下げる（可読性を保つ）
	brightness = 0.7, -- 明度を70%にする（軽い調光）
}

config.colors = {
	-- 基本色
	foreground = colors.foregrounds.main, -- Main Foreground
	background = colors.wezterm.background, -- Main Background

	-- カーソル色（前景色と統一）
	cursor_bg = colors.foregrounds.main, -- Main Foreground
	cursor_fg = colors.core.background,
	cursor_border = colors.foregrounds.main,

	-- 選択色
	selection_fg = colors.core.selection_fg, -- Background color for text
	selection_bg = colors.wezterm.selection_bg, -- High Visibility

	-- スクロールバー・分割線
	scrollbar_thumb = colors.teals.border, -- UI Border
	split = colors.wezterm.split,

	-- IME入力中カーソル（通常カーソルと統一）
	compose_cursor = colors.foregrounds.main,
	-- ビジュアルベル（控えめフラッシュ）
	visual_bell = colors.teals.dark_accent,

	-- ANSI色
	ansi = {
		colors.ansi.black,
		colors.ansi.red,
		colors.ansi.green,
		colors.ansi.yellow,
		colors.ansi.blue,
		colors.ansi.magenta,
		colors.ansi.cyan,
		colors.ansi.white,
	},
	-- 明るいANSI色
	brights = {
		colors.ansi.bright_black,
		colors.ansi.bright_red,
		colors.ansi.bright_green,
		colors.ansi.bright_yellow,
		colors.ansi.bright_blue,
		colors.ansi.bright_magenta,
		colors.ansi.bright_cyan,
		colors.ansi.bright_white,
	},

	-- コピーモード色設定
	copy_mode_active_highlight_bg = { Color = colors.teals.bright },
	copy_mode_active_highlight_fg = { Color = colors.core.background },
	copy_mode_inactive_highlight_bg = { Color = colors.core.panel_bg },
	copy_mode_inactive_highlight_fg = { Color = colors.foregrounds.main },

	-- クイックセレクト色設定
	quick_select_label_bg = { Color = colors.ansi.red },
	quick_select_label_fg = { Color = colors.ansi.bright_white },
	quick_select_match_bg = { Color = colors.teals.bright },
	quick_select_match_fg = { Color = colors.core.background },

	-- タブバー設定
	tab_bar = {
		background = "none",
		active_tab = {
			bg_color = colors.wezterm.active_tab,
			fg_color = colors.wezterm.active_tab_fg,
			intensity = "Bold",
		},
		inactive_tab = {
			bg_color = colors.wezterm.inactive_tab, -- pill化でタブ境界を明確化（旧 none は透明で境界不可視）
			fg_color = colors.wezterm.inactive_tab_fg,
		},
		inactive_tab_hover = {
			bg_color = colors.teals.dark_accent,
			fg_color = colors.foregrounds.main,
		},
		new_tab = {
			bg_color = "none",
			fg_color = colors.wezterm.inactive_tab_fg,
		},
		new_tab_hover = {
			bg_color = colors.wezterm.inactive_tab,
			fg_color = colors.foregrounds.main,
		},
		inactive_tab_edge = colors.teals.border,
	},
}

-- ==========================================
-- タブのカスタマイズ
-- ==========================================
-- アクティブタブの固定色（ghost-visor: slate_dark / bright）
local ACTIVE_TAB_BG = colors.wezterm.active_tab
local ACTIVE_TAB_FG = colors.wezterm.active_tab_fg

-- 稼働中に表示する点字スピナーのコマ（純正と同じ点字系・U+280B..）。
-- Claude Code はタイトルのスピナーをコマ送り更新しないため、wezterm側で自前にアニメーションさせる。
local spinner_cps = { 0x280B, 0x2819, 0x2839, 0x2838, 0x283C, 0x2834, 0x2826, 0x2827, 0x2807, 0x280F }
local spinner_idx = 0

-- Claude Code 純正タイトル(OSC 2)の先頭グリフから状態インジケータを返す。
-- Claude Code はタイトルを「<グリフ> <タスク要約>」形式で設定する:
--   稼働中    = 点字スピナー U+2800–U+28FF (UTF-8: E2 A0–A3 xx)
--   待機/完了 = ✳ 等の Dingbats (UTF-8: E2 9C/9D/9E xx)
-- 誤検出を抑えるため「E2先頭 ＋ グリフ直後が空白(=要約区切り) ＋ 点字/Dingbats」のみ採用し、
-- 矢印・罫線・bullet 等の他E2記号(vim/lazygit/btop 等)は Claude 状態とみなさない。
-- 待機は ✓ を表示、稼働は自前の点字スピナーをアニメーション表示する。
local function get_claude_status(pane_title)
	local b1, b2 = pane_title:byte(1, 2)
	if b1 ~= 0xE2 or b2 == nil then
		return nil -- 先頭が3バイトUTF-8シンボル(Claudeのグリフ)でなければ対象外
	end
	local b4 = pane_title:byte(4)
	if b4 ~= nil and b4 ~= 0x20 then
		return nil -- グリフ直後が空白でない(=「<グリフ> <要約>」形式でない)なら対象外
	end
	if b2 >= 0xA0 and b2 <= 0xA3 then
		-- 稼働中: 自前スピナーをコマ送り（update-statusで全タブ同期。非アクティブタブも動く）
		local frame = utf8.char(spinner_cps[(spinner_idx % #spinner_cps) + 1])
		return { icon = frame, color = colors.teals.bright }
	end
	if b2 == 0x9C or b2 == 0x9D or b2 == 0x9E then
		return { icon = "✓", color = colors.foregrounds.bright } -- 待機/完了: Dingbats(✳ 等) → チェックマーク
	end
	return nil -- それ以外のE2記号は Claude 状態とみなさない
end

-- Claude Code 純正タイトルから先頭グリフ(3バイト)と続く空白を除いた要約を返す
local function strip_claude_glyph(pane_title)
	return (pane_title:sub(4):gsub("^%s+", ""))
end

-- paneのcwdを取得して短縮表示するヘルパー関数
local function short_cwd(pane, max_len)
	local cwd_uri = pane.current_working_dir
	if not cwd_uri or not cwd_uri.file_path then
		return nil
	end
	local path = cwd_uri.file_path
	local home = os.getenv("HOME")
	local title
	if home and path == home then
		title = "~"
	elseif home and path:sub(1, #home) == home then
		title = "~" .. path:sub(#home + 1)
	else
		title = path
	end
	return wezterm.truncate_right(title, max_len)
end

-- 1つのpaneの表示セグメントを返す: アイコン(非Claudeはnil), アイコン色, テキスト
local function pane_segment(pane, max_text)
	local ptitle = pane.title or ""
	local status = get_claude_status(ptitle)
	if status then
		local summary = strip_claude_glyph(ptitle)
		if summary == "" then
			summary = short_cwd(pane, max_text) or "~"
		end
		return status.icon, status.color, wezterm.truncate_right(summary, max_text)
	end
	return nil, nil, short_cwd(pane, max_text) or wezterm.truncate_right(ptitle, max_text)
end

-- タブタイトルのカスタマイズ処理（斜めシェイプ＋このタブの全paneの状態を併記）
wezterm.on("format-tab-title", function(tab, tabs, panes, config, hover)
	-- 背景色の定義
	local bar_bg = "none"
	local hover_bg = colors.teals.dark_accent

	-- 非アクティブもpill背景にして境界を明確化（colors.tab_bar.inactive_tab と一致）。
	-- 旧 "none" では左右パワーライン形状が透明で描かれ境界が消えていた。
	local bg = colors.wezterm.inactive_tab
	local fg = colors.wezterm.inactive_tab_fg

	if tab.is_active then
		bg = ACTIVE_TAB_BG
		fg = ACTIVE_TAB_FG
	elseif hover then
		bg = hover_bg
		fg = colors.foregrounds.main
	end

	-- このタブの全paneを対象に（分割時は全paneのタイトルを併記）
	local tab_panes = tab.panes or { tab.active_pane }
	-- pane数で要約幅を配分（単一は長め、複数は短く）。tab_max_width が最終的な上限
	local per = (#tab_panes > 1) and math.max(18, math.floor(80 / #tab_panes)) or 84

	local items = {
		{ Background = { Color = bar_bg } },
		{ Foreground = { Color = bg } },
		{ Text = "" }, -- 左端の斜めシェイプ
		{ Background = { Color = bg } },
		{ Foreground = { Color = fg } },
		{ Text = " " },
	}

	for i, p in ipairs(tab_panes) do
		if i > 1 then
			table.insert(items, { Foreground = { Color = fg } })
			table.insert(items, { Text = " │ " }) -- pane区切り
		end
		local icon, icon_color, text = pane_segment(p, per)
		if icon then
			-- アクティブタブは背景が色付きで状態色と被るため、高コントラストのfgで描画
			local ic = tab.is_active and fg or icon_color
			table.insert(items, { Foreground = { Color = ic } })
			table.insert(items, { Text = icon .. " " })
		end
		table.insert(items, { Foreground = { Color = fg } })
		table.insert(items, { Text = text })
	end

	table.insert(items, { Foreground = { Color = fg } })
	table.insert(items, { Text = " " })
	table.insert(items, { Background = { Color = bar_bg } })
	table.insert(items, { Foreground = { Color = bg } })
	table.insert(items, { Text = "" }) -- 右端の斜めシェイプ
	return items
end)

-- 稼働中(点字スピナー)のpaneが1つでも存在するか走査する。
-- 無ければスピナーを進めず、format-tab-title の出力が不変になるため
-- wezterm が GPU 再描画をスキップでき、待機時の常時再描画コストを抑えられる。
local function any_pane_working()
	local ok, found = pcall(function()
		for _, w in ipairs(wezterm.mux.all_windows()) do
			for _, t in ipairs(w:tabs()) do
				for _, p in ipairs(t:panes()) do
					local b1, b2 = (p:get_title() or ""):byte(1, 2)
					if b1 == 0xE2 and b2 ~= nil and b2 >= 0xA0 and b2 <= 0xA3 then
						return true
					end
				end
			end
		end
		return false
	end)
	if not ok then
		return true -- 走査失敗時は安全側（アニメ継続）
	end
	return found
end

-- タブバーを定期再描画し、稼働中スピナーをアニメーションさせる。
-- 稼働paneが在るときだけ spinner_idx を進める（待機時は出力不変＝再描画スキップ）。
config.status_update_interval = 200
wezterm.on("update-status", function(window, pane)
	if any_pane_working() then
		spinner_idx = spinner_idx + 1
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
			left = 12,
			right = 12,
			top = 10,
			bottom = 0,
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

	-- 最小パディング（12px）を保証
	local padding_left = math.max(padding_each, 12)
	local padding_right = math.max(padding_each, 12)

	local new_padding = {
		left = padding_left,
		right = padding_right,
		top = 10, -- 上部パディングを維持
		bottom = 0, -- 下部パディングはタブバーに密着
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
				left = 12,
				right = 12,
				top = 10,
				bottom = 0,
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

-- 文字選択UI色設定（config.colorsの外でトップレベルに設定）
config.char_select_bg_color = colors.wezterm.active_tab
config.char_select_fg_color = colors.wezterm.active_tab_fg

-- 設定をエクスポート
return config
