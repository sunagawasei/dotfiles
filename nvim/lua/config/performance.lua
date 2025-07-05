-- パフォーマンス診断設定
local M = {}

-- 起動時間測定
function M.measure_startup_time()
  local start_time = vim.fn.reltime()
  
  vim.api.nvim_create_autocmd("VimEnter", {
    callback = function()
      local startup_time = vim.fn.reltimestr(vim.fn.reltime(start_time))
      print("Neovim startup time: " .. startup_time .. " seconds")
    end,
  })
end

-- プラグイン読み込み時間測定
function M.profile_plugins()
  vim.cmd("profile start /tmp/nvim-profile.log")
  vim.cmd("profile func *")
  vim.cmd("profile file *")
end

-- avante.nvim専用のパフォーマンス設定
function M.optimize_avante()
  -- バッファサイズ制限
  vim.api.nvim_create_autocmd("FileType", {
    pattern = "Avante",
    callback = function()
      -- ローカルオプションでパフォーマンス向上
      vim.opt_local.syntax = "off" -- シンタックスハイライトを無効
      vim.opt_local.spell = false -- スペルチェックを無効
      vim.opt_local.wrap = false -- 折り返しを無効
    end,
  })
end

return M
