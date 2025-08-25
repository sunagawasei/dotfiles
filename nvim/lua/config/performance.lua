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

-- CPU使用率の高いプロセスを特定するヘルパー
function M.check_performance()
  vim.cmd("messages clear")
  print("=== Performance Check ===")
  print("updatetime: " .. vim.o.updatetime .. "ms")
  print("Current buffer: " .. vim.fn.expand("%:p"))
  print("LSP clients: ")
  local clients = vim.lsp.get_active_clients()
  for _, client in ipairs(clients) do
    print("  - " .. client.name)
  end
  print("Loaded plugins: " .. #vim.tbl_keys(package.loaded))
  print("Autocommands: ")
  local autocmds = vim.api.nvim_get_autocmds({})
  local event_counts = {}
  for _, autocmd in ipairs(autocmds) do
    event_counts[autocmd.event] = (event_counts[autocmd.event] or 0) + 1
  end
  for event, count in pairs(event_counts) do
    if count > 5 then  -- 5個以上のautocmdがあるイベントのみ表示
      print("  - " .. event .. ": " .. count)
    end
  end
end

-- パフォーマンスコマンドの登録
vim.api.nvim_create_user_command("PerfCheck", function()
  M.check_performance()
end, { desc = "Check Neovim performance metrics" })

vim.api.nvim_create_user_command("PerfProfile", function()
  M.profile_plugins()
  print("Profiling started. Run :profile stop to finish and check /tmp/nvim-profile.log")
end, { desc = "Start profiling Neovim" })

return M
