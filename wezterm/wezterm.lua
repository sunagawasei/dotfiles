local wezterm = require 'wezterm'
local config = {}

-- Weztermの新しい設定システムを使用
if wezterm.config_builder then
  config = wezterm.config_builder()
end

-- フォントサイズを16に設定
config.font_size = 16

-- デフォルトウィンドウサイズを設定
config.initial_cols = 80
config.initial_rows = 24

return config
