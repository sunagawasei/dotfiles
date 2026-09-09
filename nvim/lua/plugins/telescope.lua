-- .gitignore を無視して検索する。代わりに rgignore でリポジトリ横断のノイズを落とす。
-- telescope は渡したテーブルへ追記するので、呼び出しごとに新しいテーブルを返す
local function rg_args()
  return { "--no-ignore-vcs", "--ignore-file", vim.fn.stdpath("config") .. "/rgignore" }
end

-- telescope の grep_string は文字単位のVisualでしか選択範囲を読まず、行単位・矩形では
-- カーソル下の単語へ落ちる。desc と挙動を一致させるため選択文字列を自分で取り出す
local function visual_selection()
  local lines = vim.fn.getregion(vim.fn.getpos("v"), vim.fn.getpos("."), { type = vim.fn.mode() })
  if #lines > 1 then
    vim.notify("複数行の選択は検索語にできない", vim.log.levels.WARN)
    return nil
  end
  local text = lines[1]
  if not text or text == "" then
    return nil
  end
  return text
end

return {
  "nvim-telescope/telescope.nvim",
  dependencies = { "nvim-lua/plenary.nvim" },
  cmd = "Telescope",
  keys = {
    {
      "<leader>sw",
      function()
        require("telescope.builtin").grep_string({
          cwd = LazyVim.root(),
          word_match = "-w",
          additional_args = rg_args(),
        })
      end,
      mode = "n",
      desc = "Word (Telescope)",
    },
    {
      "<leader>sw",
      function()
        local text = visual_selection()
        if not text then
          return
        end
        require("telescope.builtin").grep_string({
          cwd = LazyVim.root(),
          search = text,
          additional_args = rg_args(),
        })
      end,
      mode = "x",
      desc = "Selection (Telescope)",
    },
    {
      "<leader>ss",
      function()
        require("telescope.builtin").lsp_document_symbols({ symbols = LazyVim.config.get_kind_filter() })
      end,
      desc = "LSP Symbols (Telescope)",
    },
    {
      "<leader>sh",
      function()
        require("telescope.builtin").help_tags()
      end,
      desc = "Help Pages (Telescope)",
    },
  },
}
