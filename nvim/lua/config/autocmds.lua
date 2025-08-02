-- Autocmds are automatically loaded on the VeryLazy event
-- Default autocmds that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/autocmds.lua
--
-- Add any additional autocmds here
-- with `vim.api.nvim_create_autocmd`
--
-- Or remove existing autocmds by their group name (which is prefixed with `lazyvim_` for the defaults)
-- e.g. vim.api.nvim_del_augroup_by_name("lazyvim_wrap_spell")

-- Disable the concealing in some file formats
-- The default conceallevel is 3 in LazyVim
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "json", "jsonc", "markdown" },
  callback = function()
    vim.wo.conceallevel = 0
  end,
})

-- Fix wavy lines issue in VSCode by disabling spell check
vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("lazyvim_wrap_spell", { clear = true }),
  pattern = "*",
  callback = function()
    if vim.g.vscode then
      vim.opt_local.wrap = false
      vim.opt_local.spell = false
    else
      vim.opt_local.wrap = true
      vim.opt_local.spell = true
    end
  end,
})

-- 自動フォーマット設定（Conform.nvimを直接使用）
vim.api.nvim_create_autocmd("BufWritePre", {
  group = vim.api.nvim_create_augroup("FormatOnSave", { clear = true }),
  callback = function(args)
    if vim.g.autoformat then
      require("conform").format({
        bufnr = args.buf,
        timeout_ms = 500,
        lsp_fallback = true,
      })
    end
  end,
})

-- 診断の赤い波線を確実に無効化
vim.api.nvim_create_autocmd({ "VimEnter", "LspAttach", "BufEnter", "BufWinEnter", "FileType" }, {
  callback = function()
    -- 即座に適用
    vim.diagnostic.config({
      underline = false, -- 赤い波線を無効化
      virtual_text = false, -- インライン診断テキストも無効化
      signs = true, -- 左側のサインカラムは表示
      float = {
        border = "rounded",
        source = "always",
      },
      severity_sort = true,
      update_in_insert = false,
    })

    -- 遅延して再適用（他のプラグインによる上書きを防ぐ）
    vim.defer_fn(function()
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
    end, 100)
  end,
})

-- 外部でファイルが変更されたときに自動的にバッファを更新
vim.api.nvim_create_augroup("auto_read", { clear = true })

-- フォーカスを得たとき、バッファに入ったときにファイルの変更をチェック
vim.api.nvim_create_autocmd({ "FocusGained", "BufEnter", "TermClose", "TermLeave" }, {
  group = "auto_read",
  callback = function()
    if vim.o.buftype ~= "nofile" then
      vim.cmd("checktime")
    end
  end,
})

-- カーソルが停止したときもチェック（通常モードと挿入モード両方）
vim.api.nvim_create_autocmd({ "CursorHold", "CursorHoldI" }, {
  group = "auto_read",
  callback = function()
    if vim.o.buftype ~= "nofile" then
      vim.cmd("checktime")
    end
  end,
})

-- ファイルが外部で変更されたときの通知（オプション）
vim.api.nvim_create_autocmd("FileChangedShellPost", {
  group = "auto_read",
  callback = function()
    vim.notify("File changed on disk. Buffer reloaded.", vim.log.levels.INFO)
  end,
})

-- Formatコマンドを作成
vim.api.nvim_create_user_command("Format", function(args)
  local range = nil
  if args.count ~= -1 then
    local end_line = vim.api.nvim_buf_get_lines(0, args.line2 - 1, args.line2, true)[1]
    range = {
      start = { args.line1, 0 },
      ["end"] = { args.line2, end_line:len() },
    }
  end
  require("conform").format({ async = true, lsp_fallback = true, range = range })
end, { range = true })

-- FormatToggleコマンドを作成（自動フォーマットの切り替え）
vim.api.nvim_create_user_command("FormatToggle", function(args)
  if args.bang then
    -- FormatToggle! でグローバル設定を切り替え
    vim.g.autoformat = not vim.g.autoformat
    vim.notify("Autoformat " .. (vim.g.autoformat and "enabled" or "disabled") .. " globally")
  else
    -- FormatToggle でバッファローカル設定を切り替え
    vim.b.autoformat = not vim.b.autoformat
    vim.notify("Autoformat " .. (vim.b.autoformat and "enabled" or "disabled") .. " for this buffer")
  end
end, { bang = true })
