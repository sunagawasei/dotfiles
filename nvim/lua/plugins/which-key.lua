return {
  "folke/which-key.nvim",
  opts = {
    -- 遅延を無効化
    delay = 0,
    -- プラグインの設定
    plugins = {
      marks = true, -- shows a list of your marks on ' and `
      registers = true, -- shows your registers on " in NORMAL or <C-r> in INSERT mode
      spelling = {
        enabled = false, -- スペルチェックを無効化
      },
      -- the presets plugin, adds help for a bunch of default keybindings in Neovim
      -- No actual key bindings are created
      presets = {
        operators = false, -- adds help for operators like d, y, ...
        motions = false, -- adds help for motions
        text_objects = false, -- help for text objects triggered after entering an operator
        windows = false, -- default bindings on <c-w>
        nav = false, -- misc bindings to work with windows
        z = false, -- bindings for folds, spelling and others prefixed with z
        g = false, -- bindings for prefixed with g
      },
    },
    -- window設定
    win = {
      -- border = "none", -- none, single, double, shadow
      padding = { 1, 2 }, -- extra window padding [top/bottom, right/left]
      wo = {
        winblend = 0, -- 透明度を0に
      },
    },
    -- レイアウト設定
    layout = {
      spacing = 3, -- spacing between columns
      align = "left", -- align columns left, center or right
    },
  },
}