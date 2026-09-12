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

-- ピッカーを開いた時点のカーソル行に最も近いシンボルへ選択を移す on_complete callback。
-- telescope の symbol item は LSP の selectionRange(名前の範囲)なので、宣言行の上に
-- いないときは「カーソル行以前で最後の宣言」に落ちる
local function cursor_follower(row)
  local done = false
  return function(picker)
    if done then
      return
    end
    done = true
    -- 打鍵済みなら選択を動かさない
    if picker.closed or picker:_get_prompt() ~= "" then
      return
    end

    local manager = picker.manager
    local items = {}
    for i = 1, manager and manager:num_results() or 0 do
      local entry = manager:get_entry(i)
      local item = entry and entry.value
      if type(item) == "table" and item.lnum then
        local col = item.col or 1
        items[#items + 1] = {
          index = i,
          pos = { item.lnum, col },
          range = {
            start = { line = item.lnum - 1, character = col - 1 },
            ["end"] = { line = (item.end_lnum or item.lnum) - 1, character = (item.end_col or col) - 1 },
          },
        }
      end
    end

    -- select_index の fallback は pos 昇順を前提にするので、表示順に依存しないよう自分で並べる
    table.sort(items, function(a, b)
      if a.pos[1] ~= b.pos[1] then
        return a.pos[1] < b.pos[1]
      end
      return a.pos[2] < b.pos[2]
    end)

    local idx = require("util.lsp_symbol_cursor").select_index(items, row)
    if idx then
      picker:set_selection(picker:get_row(items[idx].index))
    end
  end
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
        require("telescope.builtin").lsp_document_symbols({
          symbols = LazyVim.config.get_kind_filter(),
          -- 既定の symbol_width = 25 は gopls が返す `(*Receiver).Method` のレシーバで
          -- 埋まりメソッド名が消える。1未満は results ウィンドウ幅に対する比率
          symbol_width = 0.8,
          on_complete = { cursor_follower(vim.api.nvim_win_get_cursor(0)[1] - 1) },
        })
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
  -- 既定の preview_cutoff = 120 だと 120桁未満の幅でプレビューが畳まれる
  opts = { defaults = { layout_config = { horizontal = { preview_cutoff = 1 } } } },
}
