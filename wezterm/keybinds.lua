local wezterm = require("wezterm")
local act = wezterm.action

-- Show which key table is active in the status area
wezterm.on("update-right-status", function(window, pane)
	local name = window:active_key_table()
	if name then
		local elements = {}

		-- モードごとに異なる色とアイコンを設定（abyssal-teal.toml準拠）
		if name == "copy_mode" then
			-- Copy mode: semantic.comment
			table.insert(elements, { Background = { Color = "#525B65" } })
			table.insert(elements, { Foreground = { Color = "#FFFFFF" } })
			table.insert(elements, { Text = " 󰆏 COPY MODE " }) -- NerdFont copy icon
		elseif name == "resize_pane" then
			-- Resize mode: semantic.string
			table.insert(elements, { Background = { Color = "#659D9E" } })
			table.insert(elements, { Foreground = { Color = "#FFFFFF" } })
			table.insert(elements, { Text = "  RESIZE " }) -- NerdFont resize icon
		elseif name == "pane_navigation" then
			-- Pane Navigation mode: foregrounds.dim
			table.insert(elements, { Background = { Color = "#92A2AB" } })
			table.insert(elements, { Foreground = { Color = "#111E16" } })
			table.insert(elements, { Text = "  PANE NAV " }) -- NerdFont window icon
		elseif name == "search_mode" then
			-- Search mode: semantic.operator
			table.insert(elements, { Background = { Color = "#64BBBE" } })
			table.insert(elements, { Foreground = { Color = "#111E16" } })
			table.insert(elements, { Text = "  SEARCH " }) -- NerdFont search icon
		else
			-- その他のモード: semantic.keyword
			table.insert(elements, { Background = { Color = "#5F698E" } })
			table.insert(elements, { Foreground = { Color = "#FFFFFF" } })
			table.insert(elements, { Text = " TABLE: " .. name .. " " })
		end

		window:set_right_status(wezterm.format(elements))
	else
		window:set_right_status("")
	end
end)

return {
	keys = {
		-- Claude Code /set-terminal で追加されたShift+Enterキーバインド
		{ key = "Enter", mods = "SHIFT", action = wezterm.action { SendString = "\x1b\r" } },
		-- Ctrl+Zを無効化（サスペンド防止）
		{ key = "z", mods = "CTRL", action = act.Nop },
		{
			-- workspaceの切り替え
			key = "w",
			mods = "LEADER",
			action = act.ShowLauncherArgs({ flags = "WORKSPACES", title = "Select workspace" }),
		},
		{
			--workspaceの名前変更
			key = "$",
			mods = "LEADER",
			action = act.PromptInputLine({
				description = "(wezterm) Set workspace title:",
				action = wezterm.action_callback(function(win, pane, line)
					if line then
						wezterm.mux.rename_workspace(wezterm.mux.get_active_workspace(), line)
					end
				end),
			}),
		},
		{
			key = "W",
			mods = "LEADER|SHIFT",
			action = act.PromptInputLine({
				description = "(wezterm) Create new workspace:",
				action = wezterm.action_callback(function(window, pane, line)
					if line then
						window:perform_action(
							act.SwitchToWorkspace({
								name = line,
							}),
							pane
						)
					end
				end),
			}),
		},
		-- コマンドパレット表示
		{ key = "p", mods = "SUPER", action = act.ActivateCommandPalette },
		-- 新しいウィンドウを開く
		{
			key = "n",
			mods = "SUPER",
			action = wezterm.action_callback(function(window, pane)
				-- 利用可能なスクリーン情報を取得
				local screens = wezterm.gui.screens()
				-- 現在アクティブなスクリーンを取得
				local active_screen = screens.active

				-- アクティブスクリーンの85%の幅、90%の高さで設定
				local ratio_width = 0.85
				local ratio_height = 1
				local width = active_screen.width * ratio_width
				local height = active_screen.height * ratio_height

				-- ウィンドウをアクティブスクリーンの中央に配置して新しいウィンドウを作成
				local tab, pane, new_window = wezterm.mux.spawn_window({
					position = {
						x = active_screen.x + (active_screen.width - width) / 2,
						y = active_screen.y + (active_screen.height - height) / 2,
						origin = "ActiveScreen",
					},
				})

				-- ウィンドウサイズを設定
				new_window:gui_window():set_inner_size(width, height)
			end),
		},
		-- Tab移動
		{ key = "]", mods = "SUPER|SHIFT", action = act.ActivateTabRelative(1) },
		{ key = "[", mods = "SUPER|SHIFT", action = act.ActivateTabRelative(-1) },
		-- Tab移動 (Ctrl+Tab) - 外部キーボード互換用
		{ key = "Tab", mods = "CTRL", action = act.ActivateTabRelative(1) },
		{ key = "Tab", mods = "CTRL|SHIFT", action = act.ActivateTabRelative(-1) },
		-- Tab入れ替え
		{ key = "{", mods = "LEADER", action = act({ MoveTabRelative = -1 }) },
		-- Tab新規作成
		{ key = "t", mods = "SUPER", action = act({ SpawnTab = "CurrentPaneDomain" }) },
		-- Tabを閉じる
		{ key = "w", mods = "SUPER", action = act({ CloseCurrentTab = { confirm = true } }) },
		{ key = "}", mods = "LEADER", action = act({ MoveTabRelative = 1 }) },

		-- コピーモード (CTRL+SHIFT+X)
		{ key = "X", mods = "CTRL|SHIFT", action = act.ActivateCopyMode },
		-- コピー
		{ key = "c", mods = "SUPER", action = act.CopyTo("Clipboard") },
		-- 貼り付け
		{ key = "v", mods = "SUPER", action = act.PasteFrom("Clipboard") },

		-- Pane作成 leader + r or d
		{ key = "d", mods = "LEADER", action = act.SplitVertical({ domain = "CurrentPaneDomain" }) },
		{ key = "r", mods = "LEADER", action = act.SplitHorizontal({ domain = "CurrentPaneDomain" }) },
		-- Paneを閉じる leader + x
		{ key = "x", mods = "LEADER", action = act({ CloseCurrentPane = { confirm = true } }) },
		-- Pane移動 leader + hlkj
		{ key = "h", mods = "LEADER", action = act.ActivatePaneDirection("Left") },
		{ key = "l", mods = "LEADER", action = act.ActivatePaneDirection("Right") },
		{ key = "k", mods = "LEADER", action = act.ActivatePaneDirection("Up") },
		{ key = "j", mods = "LEADER", action = act.ActivatePaneDirection("Down") },
		-- Pane選択
		{ key = "p", mods = "LEADER", action = act.PaneSelect },
		-- 選択中のPaneのみ表示
		{ key = "z", mods = "LEADER", action = act.TogglePaneZoomState },
		-- ターミナルクリア
		{ key = "c", mods = "LEADER", action = act.SendString('\x0c') },

		-- フォントサイズ切替
		{ key = "+", mods = "SUPER", action = act.IncreaseFontSize },
		{ key = "-", mods = "SUPER", action = act.DecreaseFontSize },
		-- フォントサイズのリセット
		{ key = "0", mods = "SUPER", action = act.ResetFontSize },

		-- タブ切替 Cmd + 数字
		{ key = "1", mods = "SUPER", action = act.ActivateTab(0) },
		{ key = "2", mods = "SUPER", action = act.ActivateTab(1) },
		{ key = "3", mods = "SUPER", action = act.ActivateTab(2) },
		{ key = "4", mods = "SUPER", action = act.ActivateTab(3) },
		{ key = "5", mods = "SUPER", action = act.ActivateTab(4) },
		{ key = "6", mods = "SUPER", action = act.ActivateTab(5) },
		{ key = "7", mods = "SUPER", action = act.ActivateTab(6) },
		{ key = "8", mods = "SUPER", action = act.ActivateTab(7) },
		{ key = "9", mods = "SUPER", action = act.ActivateTab(-1) },

		-- コマンドパレット
		{ key = "p", mods = "SHIFT|CTRL", action = act.ActivateCommandPalette },
		-- 設定再読み込み
		{ key = "r", mods = "SHIFT|CTRL", action = act.ReloadConfiguration },
		-- キーテーブル用
		{ key = "s", mods = "LEADER", action = act.ActivateKeyTable({ name = "resize_pane", one_shot = false }) },
		{
			key = "a",
			mods = "LEADER",
			action = act.ActivateKeyTable({ name = "activate_pane", timeout_milliseconds = 1000 }),
		},

		-- バックスラッシュ入力用 (¥の代わりに\を入力)
		{ key = "¥", mods = "NONE", action = act.SendString("\\") },
		{ key = "¥", mods = "SHIFT", action = act.SendString("|") },

		-- Shift+Enterで改行を挿入
		{ key = "Enter", mods = "SHIFT", action = wezterm.action.SendString("\n") },

		-- Quick Select mode
		{ key = "Space", mods = "CTRL|SHIFT", action = act.QuickSelect },

		-- ペインナビゲーションモード (Leader + q)
		{ key = "q", mods = "LEADER", action = act.ActivateKeyTable({ name = "pane_navigation", one_shot = false }) },

		-- よく使うペイン移動用の直接キーバインド
		{ key = "H", mods = "CTRL|SHIFT", action = act.ActivatePaneDirection("Left") },
		{ key = "L", mods = "CTRL|SHIFT", action = act.ActivatePaneDirection("Right") },
		{ key = "K", mods = "CTRL|SHIFT", action = act.ActivatePaneDirection("Up") },
		{ key = "J", mods = "CTRL|SHIFT", action = act.ActivatePaneDirection("Down") },

		-- ペイン回転
		{ key = "R", mods = "LEADER|SHIFT", action = act.RotatePanes("Clockwise") },

		-- ==========================================
		-- 中央寄せ機能のキーバインド (no-neck-pain風)
		-- ==========================================
		-- 中央寄せ機能のオン/オフ切り替え (Leader + m)
		{
			key = "m",
			mods = "LEADER",
			action = wezterm.action_callback(function(window, pane)
				wezterm.toggle_centering(window, pane)
			end),
		},
		-- プリセット切り替え (Leader + M)
		{
			key = "M",
			mods = "LEADER|SHIFT",
			action = wezterm.action_callback(function(window, pane)
				wezterm.cycle_centering_preset(window, pane)
			end),
		},
		-- 中央寄せ幅を狭くする (Leader + -)
		{
			key = "-",
			mods = "LEADER",
			action = wezterm.action_callback(function(window, pane)
				wezterm.adjust_centering_width(window, pane, -0.05)  -- 5%ずつ調整
			end),
		},
		-- 中央寄せ幅を広くする (Leader + +) - JISキーボード対応
		{
			key = "+",
			mods = "LEADER|SHIFT",
			action = wezterm.action_callback(function(window, pane)
				wezterm.adjust_centering_width(window, pane, 0.05)  -- 5%ずつ調整
			end),
		},
		-- 最大幅を狭くする (Leader + ,)
		{
			key = ",",
			mods = "LEADER",
			action = wezterm.action_callback(function(window, pane)
				wezterm.adjust_centering_max_width(window, pane, -100)  -- 100pxずつ調整
			end),
		},
		-- 最大幅を広くする (Leader + .)
		{
			key = ".",
			mods = "LEADER",
			action = wezterm.action_callback(function(window, pane)
				wezterm.adjust_centering_max_width(window, pane, 100)  -- 100pxずつ調整
			end),
		},
		-- フルスクリーン限定モードの切り替え (Leader + f)
		{
			key = "f",
			mods = "LEADER",
			action = wezterm.action_callback(function(window, pane)
				wezterm.toggle_centering_fullscreen_only(window, pane)
			end),
		},
	},
	-- キーテーブル
	-- https://wezfurlong.org/wezterm/config/key-tables.html
	key_tables = {
		-- Paneサイズ調整 leader + s
		resize_pane = {
			{ key = "h", action = act.AdjustPaneSize({ "Left", 1 }) },
			{ key = "l", action = act.AdjustPaneSize({ "Right", 1 }) },
			{ key = "k", action = act.AdjustPaneSize({ "Up", 1 }) },
			{ key = "j", action = act.AdjustPaneSize({ "Down", 1 }) },

			-- Cancel the mode by pressing escape
			{ key = "Enter", action = "PopKeyTable" },
		},
		activate_pane = {
			{ key = "h", action = act.ActivatePaneDirection("Left") },
			{ key = "l", action = act.ActivatePaneDirection("Right") },
			{ key = "k", action = act.ActivatePaneDirection("Up") },
			{ key = "j", action = act.ActivatePaneDirection("Down") },
		},
		-- copyモード leader + [
		copy_mode = {
			-- 移動
			{ key = "h", mods = "NONE", action = act.CopyMode("MoveLeft") },
			{ key = "j", mods = "NONE", action = act.CopyMode("MoveDown") },
			{ key = "k", mods = "NONE", action = act.CopyMode("MoveUp") },
			{ key = "l", mods = "NONE", action = act.CopyMode("MoveRight") },
			-- 最初と最後に移動
			{ key = "^", mods = "SHIFT", action = act.CopyMode("MoveToStartOfLineContent") },
			{ key = "$", mods = "SHIFT", action = act.CopyMode("MoveToEndOfLineContent") },
			-- 左端に移動
			{ key = "0", mods = "NONE", action = act.CopyMode("MoveToStartOfLine") },
			{ key = "o", mods = "NONE", action = act.CopyMode("MoveToSelectionOtherEnd") },
			{ key = "O", mods = "NONE", action = act.CopyMode("MoveToSelectionOtherEndHoriz") },
			--
			{ key = ";", mods = "NONE", action = act.CopyMode("JumpAgain") },
			-- 単語ごと移動
			{ key = "w", mods = "NONE", action = act.CopyMode("MoveForwardWord") },
			{ key = "b", mods = "NONE", action = act.CopyMode("MoveBackwardWord") },
			{ key = "e", mods = "NONE", action = act.CopyMode("MoveForwardWordEnd") },
			{ key = "Tab", mods = "NONE", action = act.CopyMode("MoveForwardWord") },
			{ key = "Tab", mods = "SHIFT", action = act.CopyMode("MoveBackwardWord") },
			-- ジャンプ機能 t f
			{ key = "t", mods = "NONE", action = act.CopyMode({ JumpForward = { prev_char = true } }) },
			{ key = "f", mods = "NONE", action = act.CopyMode({ JumpForward = { prev_char = false } }) },
			{ key = "T", mods = "NONE", action = act.CopyMode({ JumpBackward = { prev_char = true } }) },
			{ key = "F", mods = "NONE", action = act.CopyMode({ JumpBackward = { prev_char = false } }) },
			-- 一番下へ
			{ key = "G", mods = "NONE", action = act.CopyMode("MoveToScrollbackBottom") },
			-- 一番上へ
			{ key = "g", mods = "NONE", action = act.CopyMode("MoveToScrollbackTop") },
			-- viweport
			{ key = "H", mods = "NONE", action = act.CopyMode("MoveToViewportTop") },
			{ key = "L", mods = "NONE", action = act.CopyMode("MoveToViewportBottom") },
			{ key = "M", mods = "NONE", action = act.CopyMode("MoveToViewportMiddle") },
			-- スクロール
			{ key = "b", mods = "CTRL", action = act.CopyMode("PageUp") },
			{ key = "f", mods = "CTRL", action = act.CopyMode("PageDown") },
			{ key = "d", mods = "CTRL", action = act.CopyMode({ MoveByPage = 0.5 }) },
			{ key = "u", mods = "CTRL", action = act.CopyMode({ MoveByPage = -0.5 }) },
			-- 範囲選択モード
			{ key = "v", mods = "NONE", action = act.CopyMode({ SetSelectionMode = "Cell" }) },
			{ key = "v", mods = "CTRL", action = act.CopyMode({ SetSelectionMode = "Block" }) },
			{ key = "V", mods = "NONE", action = act.CopyMode({ SetSelectionMode = "Line" }) },
			-- コピー（copy modeは継続）
			{ key = "y", mods = "NONE", action = act.CopyTo("Clipboard") },

			-- 検索機能（vimライク）
			{ key = "/", mods = "NONE", action = act.CopyMode("EditPattern") },
			{ key = "?", mods = "NONE", action = act.CopyMode("EditPattern") },
			{ key = "n", mods = "NONE", action = act.CopyMode("NextMatch") },
			{ key = "N", mods = "NONE", action = act.CopyMode("PriorMatch") },

			-- [ キーでcopy modeを終了（vimライク）
			{ key = "[", mods = "NONE", action = act.CopyMode("Close") },

			-- コピーモードを終了
			{
				key = "Enter",
				mods = "NONE",
				action = act.Multiple({ { CopyTo = "ClipboardAndPrimarySelection" }, { CopyMode = "Close" } }),
			},
			-- 検索とセレクションをクリア（Copy Modeは継続）
			{ key = "Escape", mods = "NONE", action = act.Multiple({ act.CopyMode("ClearPattern"), act.CopyMode("ClearSelectionMode") }) },
			{ key = "c", mods = "CTRL", action = act.CopyMode("Close") },
			{ key = "q", mods = "NONE", action = act.CopyMode("Close") },
		},
		-- ペインナビゲーションモード (Leader + q)
		pane_navigation = {
			{ key = "h", action = act.ActivatePaneDirection("Left") },
			{ key = "l", action = act.ActivatePaneDirection("Right") },
			{ key = "k", action = act.ActivatePaneDirection("Up") },
			{ key = "j", action = act.ActivatePaneDirection("Down") },
			-- 数字キーでペインを直接選択
			{ key = "1", action = act.ActivatePaneByIndex(0) },
			{ key = "2", action = act.ActivatePaneByIndex(1) },
			{ key = "3", action = act.ActivatePaneByIndex(2) },
			{ key = "4", action = act.ActivatePaneByIndex(3) },
			{ key = "5", action = act.ActivatePaneByIndex(4) },
			{ key = "6", action = act.ActivatePaneByIndex(5) },
			{ key = "7", action = act.ActivatePaneByIndex(6) },
			{ key = "8", action = act.ActivatePaneByIndex(7) },
			{ key = "9", action = act.ActivatePaneByIndex(8) },
			-- ペインサイズ調整
			{ key = "H", mods = "SHIFT", action = act.AdjustPaneSize({ "Left", 5 }) },
			{ key = "L", mods = "SHIFT", action = act.AdjustPaneSize({ "Right", 5 }) },
			{ key = "K", mods = "SHIFT", action = act.AdjustPaneSize({ "Up", 5 }) },
			{ key = "J", mods = "SHIFT", action = act.AdjustPaneSize({ "Down", 5 }) },
			-- ペイン回転
			{ key = "r", action = act.RotatePanes("Clockwise") },
			{ key = "R", mods = "SHIFT", action = act.RotatePanes("CounterClockwise") },
			-- ペインを閉じる
			{ key = "x", action = act.CloseCurrentPane({ confirm = true }) },
			-- ペインをズーム
			{ key = "z", action = act.TogglePaneZoomState },
			-- モードを終了
			{ key = "Enter", action = "PopKeyTable" },
			{ key = "Escape", action = "PopKeyTable" },
			{ key = "q", action = "PopKeyTable" },
		},
	},
}
