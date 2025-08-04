-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here

-- エクスプローラーのキーマッピングをカスタマイズ
-- エクスプローラーが開いている場合はフォーカス、フォーカスされている場合は閉じる、閉じている場合は開く
local function smart_explorer_focus(opts)
  local Snacks = require("snacks")
  opts = opts or {}
  
  -- 既存のエクスプローラーピッカーを取得
  local explorer_pickers = Snacks.picker.get({ source = "explorer" })
  
  -- エクスプローラーが存在する場合
  for _, picker in pairs(explorer_pickers) do
    if picker:is_focused() then
      -- フォーカスされている場合は閉じる
      picker:close()
      return
    else
      -- フォーカスされていない場合はフォーカスを移動
      picker:focus()
      return
    end
  end
  
  -- エクスプローラーが存在しない場合は新規作成
  if #explorer_pickers == 0 then
    Snacks.picker.explorer(opts)
  end
end

-- キーマッピングの設定
vim.keymap.set("n", "<leader>e", function() 
  smart_explorer_focus()
end, { desc = "Explorer Snacks (root dir)" })

vim.keymap.set("n", "<leader>E", function() 
  smart_explorer_focus({ cwd = vim.fn.getcwd() })
end, { desc = "Explorer Snacks (cwd)" })

-- フォーマット用のキーマップ
vim.keymap.set({ "n", "v" }, "<leader>fm", function()
  require("conform").format({
    async = true,
    lsp_fallback = true,
  })
end, { desc = "Format file or range" })

-- 診断表示用のキーマップ
vim.keymap.set("n", "gl", vim.diagnostic.open_float, { desc = "Line diagnostics" })
vim.keymap.set("n", "<leader>ld", vim.diagnostic.open_float, { desc = "Line diagnostics" })

-- 診断間の移動
vim.keymap.set("n", "[d", vim.diagnostic.goto_prev, { desc = "Previous diagnostic" })
vim.keymap.set("n", "]d", vim.diagnostic.goto_next, { desc = "Next diagnostic" })
vim.keymap.set("n", "[e", function() vim.diagnostic.goto_prev({ severity = vim.diagnostic.severity.ERROR }) end, { desc = "Previous error" })
vim.keymap.set("n", "]e", function() vim.diagnostic.goto_next({ severity = vim.diagnostic.severity.ERROR }) end, { desc = "Next error" })

-- 診断リスト
vim.keymap.set("n", "<leader>dl", vim.diagnostic.setloclist, { desc = "Diagnostic location list" })
vim.keymap.set("n", "<leader>dq", vim.diagnostic.setqflist, { desc = "Diagnostic quickfix list" })

-- ファイル保存（Normal、Insert、Visual モード）
vim.keymap.set({ "n", "i", "v" }, "<C-s>", "<cmd>w<cr><esc>", { desc = "Save file" })

-- ファイル全体をコピー（Normal モード）
vim.keymap.set("n", "<C-c>", "<cmd>%y+<cr>", { desc = "Copy entire file" })

