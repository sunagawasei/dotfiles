-- Options are automatically loaded before lazy.nvim startup
-- Default options that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/options.lua

-- Add any additional options here
-- スペルチェックを無効化
local opt = vim.opt
opt.spell = false

-- IME関連の設定
opt.iminsert = 0  -- 挿入モードでIMEをオフ
opt.imsearch = -1  -- 検索時のIME設定（iminsertの値に従う）

-- ポップアップメニューの透明度を無効化
opt.pumblend = 0  -- ポップアップメニューの透明度を0に（完全に不透明）
opt.winblend = 0  -- ウィンドウの透明度を0に

-- TUI（Terminal UI）でRGBカラーを有効化
opt.termguicolors = true
