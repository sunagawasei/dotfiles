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


return M
