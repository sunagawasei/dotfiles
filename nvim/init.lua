-- bootstrap lazy.nvim, LazyVim and your plugins
require("config.lazy")

-- パフォーマンス設定の読み込み
local performance = require("config.performance")
performance.optimize_avante()
-- performance.measure_startup_time() -- 必要時にコメントアウト

vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.clipboard = "unnamedplus"

--[[ エンコーディング設定 ]]--
vim.scriptencoding = "utf-8"
vim.opt.encoding = "utf-8" -- Neovimのデフォルト文字コード
vim.opt.fileencoding = "utf-8" -- ファイルの保存時
vim.opt.fileencodings = "utf-8,utf-16,sjis," -- ファイル読み込み時推定

-- IME設定（エラーハンドリング付き）
if vim.fn.has("mac") == 1 then
  vim.opt.ttimeoutlen = 1
  local ime_group = vim.api.nvim_create_augroup("MyIMEGroup", { clear = true })
  vim.api.nvim_create_autocmd("InsertLeave", {
    group = ime_group,
    callback = function()
      pcall(function()
        vim.fn.system('osascript -e "tell application \\"System Events\\" to key code 102"')
      end)
    end,
  })
end
