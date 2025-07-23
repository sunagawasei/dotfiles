-- bootstrap lazy.nvim, LazyVim and your plugins
require("config.lazy")

-- パフォーマンス設定の読み込み
-- local performance = require("config.performance")
-- performance.measure_startup_time() -- 必要時にコメントアウト

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

-- スペルチェックを常に無効化
vim.opt.spell = false
vim.api.nvim_create_autocmd("VimEnter", {
  callback = function()
    vim.opt.spell = false
  end,
})

-- 診断の赤い波線を無効化（最優先で適用）
vim.diagnostic.config({
  underline = false,
  virtual_text = false,
  signs = true,
  float = {
    border = "rounded",
    source = "always",
  },
  severity_sort = true,
  update_in_insert = false,
})

-- スペルチェックの波線も無効化（念のため）
vim.o.spell = false
vim.cmd([[highlight clear SpellBad]])
vim.cmd([[highlight clear SpellCap]])
vim.cmd([[highlight clear SpellRare]])
vim.cmd([[highlight clear SpellLocal]])
