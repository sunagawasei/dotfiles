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

-- Auto format on save
vim.api.nvim_create_autocmd("BufWritePre", {
  pattern = "*",
  callback = function(args)
    require("conform").format({ bufnr = args.buf })
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
