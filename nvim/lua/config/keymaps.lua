-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here

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

-- 診断メッセージコピー関数
local function copy_diagnostics()
  local diagnostics = vim.diagnostic.get(0, { lnum = vim.fn.line('.') - 1 })
  if #diagnostics == 0 then
    vim.notify("No diagnostics on current line", vim.log.levels.INFO)
    return
  end

  local messages = {}
  for _, diag in ipairs(diagnostics) do
    table.insert(messages, diag.message)
  end

  local text = table.concat(messages, "\n")
  vim.fn.setreg('+', text)
  vim.notify("Diagnostic copied to clipboard", vim.log.levels.INFO)
end

-- 診断コピー
vim.keymap.set("n", "<leader>dc", copy_diagnostics, { desc = "Copy diagnostics" })
vim.keymap.set("n", "gy", copy_diagnostics, { desc = "Yank diagnostics" })

-- ファイル保存（Normal、Insert、Visual モード）
vim.keymap.set({ "n", "i", "v" }, "<C-s>", "<cmd>w<cr><esc>", { desc = "Save file" })

-- ファイル全体をコピー（Normal モード）
vim.keymap.set("n", "<C-c>", "<cmd>%y+<cr>", { desc = "Copy entire file" })

-- ファイルパスを取得
local function get_file_path()
  local bufname = vim.fn.expand('%:p')

  -- oil.nvimのバッファの場合はカーソル位置のエントリのパスを取得
  if bufname:match('^oil://') then
    local ok, oil = pcall(require, "oil")
    if ok then
      local entry = oil.get_cursor_entry()
      local dir = oil.get_current_dir()
      if entry and dir then
        return dir .. entry.name
      end
    end
    -- フォールバック: oil://プレフィックスを削除
    return bufname:gsub('^oil://', '')
  end

  return bufname
end

-- ファイルパスをクリップボードにコピー
vim.keymap.set("n", "<leader>yp", function()
  local path = get_file_path()
  if path and path ~= "" then
    vim.fn.setreg('+', path)
    vim.notify("Copied absolute path: " .. path, vim.log.levels.INFO)
  else
    vim.notify("No file path available", vim.log.levels.WARN)
  end
end, { desc = "Yank absolute path" })

vim.keymap.set("n", "<leader>yr", function()
  local path = get_file_path()
  if path and path ~= "" then
    local relative = vim.fn.fnamemodify(path, ':.')
    vim.fn.setreg('+', relative)
    vim.notify("Copied relative path: " .. relative, vim.log.levels.INFO)
  else
    vim.notify("No file path available", vim.log.levels.WARN)
  end
end, { desc = "Yank relative path" })

vim.keymap.set("n", "<leader>yf", function()
  local path = get_file_path()
  if path and path ~= "" then
    local filename = vim.fn.fnamemodify(path, ':t')
    vim.fn.setreg('+', filename)
    vim.notify("Copied filename: " .. filename, vim.log.levels.INFO)
  else
    vim.notify("No file path available", vim.log.levels.WARN)
  end
end, { desc = "Yank filename" })

-- インラインヒントの表示/非表示を切り替え
vim.keymap.set("n", "<leader>uh", function()
  vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled())
end, { desc = "Toggle Inlay Hints" })

-- Copilotの有効/無効を切り替え（copilot.lua対応）
vim.keymap.set("n", "<leader>ct", "<cmd>Copilot toggle<cr>", { desc = "Toggle Copilot" })

-- ウィンドウZoom（全画面化トグル）
vim.keymap.set({ "n", "t" }, "<C-w>z", function()
  Snacks.zen.zoom()
end, { desc = "Toggle Zoom" })
