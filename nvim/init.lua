-- Leader key設定を最優先で実行（lazy.nvim読み込み前に必須）
vim.g.mapleader = " "
vim.g.maplocalleader = "\\"

-- bootstrap lazy.nvim, LazyVim and your plugins
require("config.lazy")

vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.clipboard = "unnamedplus"

--[[ エンコーディング設定 ]]--
vim.scriptencoding = "utf-8"
vim.opt.encoding = "utf-8" -- Neovimのデフォルト文字コード
-- vim.opt.fileencoding = "utf-8" -- ファイルの保存時（バッファ単位の設定のためここでは設定しない）
vim.opt.fileencodings = "utf-8,utf-16,sjis," -- ファイル読み込み時推定

-- IME設定（エラーハンドリング付き）
if vim.fn.has("mac") == 1 then
  vim.opt.ttimeoutlen = 0  -- ESCキーの応答を即座に（デフォルト: 50ms）
  vim.opt.timeoutlen = 300  -- キーマッピングのタイムアウト時間を短縮（デフォルト: 1000ms）
end
