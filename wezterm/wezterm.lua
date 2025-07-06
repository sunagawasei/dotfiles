local wezterm = require 'wezterm'
local config = wezterm.config_builder()

config.keys = require("keybinds").keys
config.key_tables = require("keybinds").key_tables
config.disable_default_key_bindings = true

config.leader = { key = "q", mods = "CTRL", timeout_milliseconds = 2000 }

config.automatically_reload_config = true

-- フォントサイズを16に設定
config.font_size = 16

-- デフォルトウィンドウサイズを設定
config.initial_cols = 48
config.initial_rows = 62

config.use_ime = true
config.window_background_opacity = 0.85
config.macos_window_background_blur = 20
config.window_decorations = "RESIZE"
config.hide_tab_bar_if_only_one_tab = true
config.window_frame = {
   inactive_titlebar_bg = "none",
   active_titlebar_bg = "none",
 }
config.window_background_gradient = {
   colors = { "#111111" },  -- Geist Accent 8 for consistency
 }
config.show_new_tab_button_in_tab_bar = false
config.show_close_tab_button_in_tabs = false
config.colors = {
   -- Geist Design System Colors
   -- Primary colors
  --  foreground = "#000000",  -- Primary text (light theme)
  --  background = "#FFFFFF",  -- Background (light theme)
   
   -- For dark theme appearance, using reversed colors
   foreground = "#FFFFFF",
   background = "#111111",  -- Accent 8 for dark background
   
   -- Cursor colors
   cursor_bg = "#0070F3",    -- Success default (Geist blue)
   cursor_fg = "#FFFFFF",    -- White text on blue cursor
   cursor_border = "#0070F3", -- Success default
   
   -- Selection colors (adjusted for dark theme)
   selection_fg = "#FFFFFF",  -- White text on dark background
   selection_bg = "#0761D1",  -- Success dark for better contrast
   
   -- Scrollbar and split (adjusted for dark theme)
   scrollbar_thumb = "#444444", -- Accent 6 for dark theme
   split = "#333333",           -- Accent 7 for dark theme
   
   -- ANSI colors using Geist palette (improved)
   ansi = {
     "#111111",  -- Black (Accent 8)
     "#EE0000",  -- Red (Error)
     "#29BC9B",  -- Green (Cyan Dark - more subdued)
     "#F5A623",  -- Yellow (Warning)
     "#0070F3",  -- Blue (Success)
     "#7928CA",  -- Magenta (Violet)
     "#3291FF",  -- Cyan (Success Light Blue - clearly distinguishable)
     "#FAFAFA",  -- White (Accent 1)
   },
   brights = {
     "#666666",  -- Bright Black (Accent 5)
     "#FF1A1A",  -- Bright Red (Error Light)
     "#50E3C2",  -- Bright Green (Cyan)
     "#F7B955",  -- Bright Yellow (Warning Light)
     "#3291FF",  -- Bright Blue (Success Light)
     "#F81CE5",  -- Bright Magenta (Purple highlight)
     "#0070F3",  -- Bright Cyan (Success Blue - bold and clear)
     "#FFFFFF",  -- Bright White
   },
   
   -- Copy mode colors (adjusted for dark theme)
   copy_mode_active_highlight_bg = { Color = "#0070F3" },
   copy_mode_active_highlight_fg = { Color = "#FFFFFF" },
   copy_mode_inactive_highlight_bg = { Color = "#444444" },  -- Accent 6
   copy_mode_inactive_highlight_fg = { Color = "#FFFFFF" },
   
   -- Quick select colors
   quick_select_label_bg = { Color = "#7928CA" },
   quick_select_label_fg = { Color = "#FFFFFF" },
   quick_select_match_bg = { Color = "#0070F3" },
   quick_select_match_fg = { Color = "#FFFFFF" },
   
   tab_bar = {
     inactive_tab_edge = "none",
   },
 }
local SOLID_LEFT_ARROW = wezterm.nerdfonts.ple_lower_right_triangle
local SOLID_RIGHT_ARROW = wezterm.nerdfonts.ple_upper_left_triangle

wezterm.on("format-tab-title", function(tab, tabs, panes, config, hover, max_width)
   -- Geist theme colors for tabs
   local background = "#444444"  -- Accent 6 for inactive tabs
   local foreground = "#FFFFFF"
   local edge_background = "none"

   if tab.is_active then
     background = "#0070F3"  -- Geist Success blue for active tab
     foreground = "#FFFFFF"
   end
   local edge_foreground = background
   local title = "   " .. wezterm.truncate_right(tab.active_pane.title, max_width - 1) .. "   "

   return {
     { Background = { Color = edge_background } },
     { Foreground = { Color = edge_foreground } },
     { Text = SOLID_LEFT_ARROW },
     { Background = { Color = background } },
     { Foreground = { Color = foreground } },
     { Text = title },
     { Background = { Color = edge_background } },
     { Foreground = { Color = edge_foreground } },
     { Text = SOLID_RIGHT_ARROW },
   }
 end)
return config
