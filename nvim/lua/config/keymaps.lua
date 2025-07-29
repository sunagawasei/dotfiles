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

