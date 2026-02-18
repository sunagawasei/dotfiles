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

-- 外部でファイルが変更されたときに自動的に再読み込み
opt.autoread = true

-- CursorHoldイベントのトリガー時間（ミリ秒）
-- デフォルトは4000ms、CPU使用率を抑えるため1000msに設定
opt.updatetime = 1000

-- LazyVimのformat_on_saveを有効化
vim.g.autoformat = true

-- プロジェクトローカル設定（.nvim.lua）を有効化
-- Neovim 0.9以降は安全：初回読み込み時に確認ダイアログが表示される
vim.o.exrc = true

-- list表示を無効化（タブ文字の>表示を消す）
opt.list = false

-- ステータスラインを画面全体で1つだけ表示（複数ウィンドウでの切り替えを防ぐ）
opt.laststatus = 3

-- フローティングウィンドウのボーダー設定（Neovim 0.10以降）
-- すべてのフローティングウィンドウに角丸ボーダーを追加
if vim.fn.has("nvim-0.10") == 1 then
  vim.o.winborder = "rounded"
end

-- East Asian Ambiguous Width文字の扱い
-- WezTermの設定（treat_east_asian_ambiguous_width_as_wide = false）と一致させる
opt.ambiwidth = "single"
