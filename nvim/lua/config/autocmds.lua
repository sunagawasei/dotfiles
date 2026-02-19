-- Autocmds are automatically loaded on the VeryLazy event
-- Default autocmds that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/autocmds.lua
--
-- Add any additional autocmds here
-- with `vim.api.nvim_create_autocmd`
--
-- Or remove existing autocmds by their group name (which is prefixed with `lazyvim_` for the defaults)
-- e.g. vim.api.nvim_del_augroup_by_name("lazyvim_wrap_spell")

-- LazyVimのspell設定を無効化
vim.api.nvim_del_augroup_by_name("lazyvim_wrap_spell")

-- テキストファイルとMarkdownで折り返しを有効化（スペルチェックなし）
vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("custom_text_wrap", { clear = true }),
  pattern = { "text", "plaintex", "markdown", "gitcommit" },
  callback = function()
    vim.opt_local.wrap = true           -- 行折り返しを有効化
    vim.opt_local.linebreak = false     -- 画面幅ベースで折り返し（日本語対応）
    vim.opt_local.breakindent = true    -- 折り返し行のインデント保持
    vim.opt_local.showbreak = "> "      -- 折り返し行マーカー（日本語対応）
  end,
})

-- Disable the concealing in some file formats
-- The default conceallevel is 3 in LazyVim
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "json", "jsonc", "markdown" },
  callback = function()
    vim.wo.conceallevel = 0
  end,
})


-- 自動フォーマット設定（Conform.nvimを直接使用）
vim.api.nvim_create_autocmd("BufWritePre", {
  group = vim.api.nvim_create_augroup("FormatOnSave", { clear = true }),
  callback = function(args)
    -- グローバル設定とバッファローカル設定の両方をチェック
    if vim.g.autoformat then
      -- vim.b.autoformat が明示的に false でない場合のみフォーマット
      if vim.b.autoformat ~= false then
        require("conform").format({
          bufnr = args.buf,
          timeout_ms = 500,
          lsp_fallback = true,
        })
      end
    end
  end,
})

-- 外部でファイルが変更されたときに自動的にバッファを更新
vim.api.nvim_create_augroup("auto_read", { clear = true })

-- フォーカスを得たときのみファイルの変更をチェック（パフォーマンス最適化）
vim.api.nvim_create_autocmd({ "FocusGained" }, {
  group = "auto_read",
  callback = function()
    if vim.o.buftype ~= "nofile" then
      vim.cmd("checktime")
    end
  end,
})

-- カーソルが停止したときもチェック（updatetimeを長くしたので頻度が下がる）
-- パフォーマンス最適化のため無効化（FocusGainedで既にチェック済み）
-- vim.api.nvim_create_autocmd({ "CursorHold" }, {
--   group = "auto_read",
--   callback = function()
--     if vim.o.buftype ~= "nofile" then
--       vim.cmd("checktime")
--     end
--   end,
-- })

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

-- ターミナルバッファでTreesitter foldexprを無効化（パフォーマンス最適化）
-- Gemini調査で判明した最大のボトルネック：
-- Treesitterがターミナル出力（構造化されていないテキスト）を
-- 構文解析しようとして、開閉/スクロール時に大きなラグが発生
vim.api.nvim_create_autocmd("TermOpen", {
  pattern = "term://*",
  callback = function()
    vim.opt_local.foldmethod = "manual"
    vim.opt_local.foldexpr = "0"
  end,
  desc = "Disable Treesitter foldexpr in terminal buffers for performance",
})

-- ターミナルモード（tモード）でもleader keyマッピングを使えるようにする
-- 背景: terminal_mappings=trueはtogglterm固有のマッピングのみ有効にするため、
-- LazyVimのグローバルleader keyマッピングはtモードでは動作しない。
-- ここでstopinsertしてからSnacksを呼び出すことで、ターミナルモードのままでも動作する。
vim.api.nvim_create_autocmd("TermOpen", {
  pattern = "term://*toggleterm#*",
  callback = function()
    local opts = { buffer = 0, silent = true }
    vim.keymap.set("t", "<leader>ff", function()
      vim.cmd("stopinsert")
      Snacks.picker.files()
    end, vim.tbl_extend("force", opts, { desc = "Find Files (from terminal)" }))
    vim.keymap.set("t", "<leader><space>", function()
      vim.cmd("stopinsert")
      Snacks.picker.files()
    end, vim.tbl_extend("force", opts, { desc = "Find Files (from terminal)" }))
    vim.keymap.set("t", "<leader>fg", function()
      vim.cmd("stopinsert")
      Snacks.picker.grep()
    end, vim.tbl_extend("force", opts, { desc = "Grep (from terminal)" }))
    vim.keymap.set("t", "<leader>fb", function()
      vim.cmd("stopinsert")
      Snacks.picker.buffers()
    end, vim.tbl_extend("force", opts, { desc = "Buffers (from terminal)" }))
    vim.keymap.set("t", "<leader>e", function()
      vim.cmd("stopinsert")
      Snacks.picker.explorer()
    end, vim.tbl_extend("force", opts, { desc = "Explorer (from terminal)" }))
  end,
  desc = "Enable leader key mappings in toggleterm terminal mode",
})

-- CursorLineをファイルエクスプローラーで無効化（パフォーマンス最適化）
-- Codex調査で判明したNeovim既知バグ（neovim/neovim#8159）への対策：
-- CursorLineハイライトが下方向カーソル移動を劇的に遅延させる
-- 上方向は高速、下方向のみ遅い非対称的な問題
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "neo-tree", "NvimTree", "oil" },
  callback = function()
    vim.opt_local.cursorline = false
  end,
  desc = "Disable cursorline in file explorers for performance",
})
