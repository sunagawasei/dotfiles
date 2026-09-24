-- snacks.nvim(Snacks.gh)のgh_actionsメニューから呼ばれる、
-- PRのdiffをhunkで開くアクションの実処理。登録は plugins/snacks.lua 側。
local M = {}

-- toggletermのid空間は番号付きターミナル(1-9)とlazygit=99/hunk=98/hunk_staged=96が
-- 使用中。PRごとにcmdが変わるためTerminal:newを毎回呼ぶ必要があり、固定idだと
-- 2回目以降は最初のインスタンス(古いcmd)がそのまま返ってくる(toggleterm仕様)。
local next_terminal_id = 900

local function open_terminal(cmd, on_exit)
  next_terminal_id = next_terminal_id + 1
  require("toggleterm.terminal").Terminal
    :new({
      cmd = cmd,
      id = next_terminal_id,
      direction = "float",
      hidden = true,
      env = {
        EDITOR = vim.fn.stdpath("config") .. "/bin/hunk-nvim-editor/nvim",
      },
      float_opts = {
        border = "curved",
        width = function() return math.floor(vim.o.columns * 0.95) end,
        height = function() return math.floor(vim.o.lines * 0.95) end,
      },
      on_open = function(term)
        vim.cmd("startinsert!")
        vim.api.nvim_buf_set_keymap(term.bufnr, "n", "q", "<cmd>close<CR>", { noremap = true, silent = true })
      end,
      on_exit = on_exit,
    })
    :open()
end

---@param repo string owner/name
---@param number integer
function M.open_pr_diff(repo, number)
  local result = vim.system({ "gh", "pr", "diff", tostring(number), "-R", repo }, { text = true }):wait()
  if result.code ~= 0 then
    vim.notify("gh pr diff が失敗しました: " .. (result.stderr or ""), vim.log.levels.ERROR)
    return
  end

  local tmpfile = vim.fn.tempname()
  vim.fn.writefile(vim.split(result.stdout, "\n", { plain = true }), tmpfile)

  -- hunk patch はstdinからも読めるが、パイプ経由だとstdinがパイプに専有されて
  -- ターミナルからのキー入力を受け取れなくなる(TUI操作が効かない)。ファイル引数で回避する。
  open_terminal(string.format("hunk patch %s", vim.fn.shellescape(tmpfile)), function()
    vim.schedule(function() os.remove(tmpfile) end)
  end)
end

return M
