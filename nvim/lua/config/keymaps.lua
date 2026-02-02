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

local function systemlist_ok(cmd)
  local output = vim.fn.systemlist(cmd)
  if vim.v.shell_error ~= 0 or not output or output[1] == "" then
    return nil
  end
  return output[1]
end

local function get_git_root(path)
  if not path or path == "" then
    return nil
  end

  local dir = path
  if vim.fn.isdirectory(path) == 0 then
    dir = vim.fn.fnamemodify(path, ":h")
  end

  return systemlist_ok({ "git", "-C", dir, "rev-parse", "--show-toplevel" })
end

local function get_git_branch(root)
  return systemlist_ok({ "git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD" })
end

local function get_git_remote(root)
  return systemlist_ok({ "git", "-C", root, "remote", "get-url", "origin" })
end

local function normalize_github_remote(remote)
  if not remote or remote == "" then
    return nil
  end

  local https = remote
  https = https:gsub("^git@github.com:", "https://github.com/")
  https = https:gsub("^ssh://git@github.com/", "https://github.com/")
  https = https:gsub("%.git$", "")

  if not https:match("^https://github.com/") then
    return nil
  end

  return https
end

local function get_repo_relative(path, root)
  if not path or path == "" or not root or root == "" then
    return nil
  end

  if path == root then
    return ""
  end

  local prefix = root .. "/"
  if path:sub(1, #prefix) ~= prefix then
    return nil
  end

  return path:sub(#prefix + 1)
end

local function get_github_url(path)
  local root = get_git_root(path)
  if not root then
    return nil, "Not a git repo"
  end

  local remote = get_git_remote(root)
  local base = normalize_github_remote(remote)
  if not base then
    return nil, "Not a GitHub remote"
  end

  local branch = get_git_branch(root)
  if not branch then
    return nil, "No branch"
  end

  local rel = get_repo_relative(path, root)
  if rel == nil then
    return nil, "Path not in repo"
  end

  local kind = vim.fn.isdirectory(path) == 1 and "tree" or "blob"
  local url = string.format("%s/%s/%s", base, kind, branch)
  if rel ~= "" then
    url = url .. "/" .. rel
  end

  return url
end

local function get_line_range()
  local mode = vim.fn.mode()
  if mode == "v" or mode == "V" or mode == "\22" then
    local start_line = vim.fn.line("v")
    local end_line = vim.fn.line(".")
    if start_line > end_line then
      start_line, end_line = end_line, start_line
    end
    return start_line, end_line
  end

  local line = vim.fn.line(".")
  return line, line
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

-- GitHub URLをクリップボードにコピー
vim.keymap.set("n", "<leader>yg", function()
  local path = get_file_path()
  if not path or path == "" then
    vim.notify("No file path available", vim.log.levels.WARN)
    return
  end

  local url, err = get_github_url(path)
  if not url then
    vim.notify(err or "GitHub URL not available", vim.log.levels.WARN)
    return
  end

  vim.fn.setreg('+', url)
  vim.notify("Copied GitHub URL: " .. url, vim.log.levels.INFO)
end, { desc = "Yank GitHub URL" })

-- GitHub行リンクをクリップボードにコピー
vim.keymap.set({ "n", "v" }, "<leader>yl", function()
  if vim.fn.expand('%:p'):match('^oil://') then
    vim.notify("Line link is not available in oil buffers", vim.log.levels.WARN)
    return
  end

  local path = get_file_path()
  if not path or path == "" then
    vim.notify("No file path available", vim.log.levels.WARN)
    return
  end

  if vim.fn.isdirectory(path) == 1 then
    vim.notify("Line link is only for files", vim.log.levels.WARN)
    return
  end

  local url, err = get_github_url(path)
  if not url then
    vim.notify(err or "GitHub URL not available", vim.log.levels.WARN)
    return
  end

  local start_line, end_line = get_line_range()
  local anchor = start_line == end_line
      and ("#L" .. start_line)
      or ("#L" .. start_line .. "-L" .. end_line)
  local link = url .. anchor

  vim.fn.setreg('+', link)
  vim.notify("Copied GitHub line URL: " .. link, vim.log.levels.INFO)
end, { desc = "Yank GitHub line URL" })

-- インラインヒントの表示/非表示を切り替え
vim.keymap.set("n", "<leader>uh", function()
  vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled())
end, { desc = "Toggle Inlay Hints" })

-- Copilotの有効/無効を切り替え（copilot.lua対応）
vim.keymap.set("n", "<leader>ct", "<cmd>Copilot toggle<cr>", { desc = "Toggle Copilot" })

vim.keymap.set("n", "<leader>cp", "<cmd>CodeCompanionChat Toggle<cr>", { desc = "CodeCompanion Chat Toggle" })

-- ウィンドウZoom（全画面化トグル）
vim.keymap.set({ "n", "t" }, "<C-w>z", function()
  Snacks.zen.zoom()
end, { desc = "Toggle Zoom" })
