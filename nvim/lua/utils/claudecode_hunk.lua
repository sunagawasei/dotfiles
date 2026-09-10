local M = {}

local TMPFILE_OPTS = { subdir = "claudecode-hunk", ext = "diff" }

-- gitsigns の create_hunk と同じ式。公開 API の get_hunks() は vend を返さない。
local function hunk_vend(hunk)
  return hunk.added.start + math.max(hunk.added.count - 1, 0)
end

-- 先頭の delete hunk は added.start が 0、末尾の delete hunk は EOF+1 になり、
-- どちらもカーソルが乗れない行番号なので個別に判定する。
function M.find_hunk(hunks, lnum, line_count)
  for _, hunk in ipairs(hunks or {}) do
    local vend = hunk_vend(hunk)
    if lnum == 1 and hunk.added.start == 0 and vend == 0 then
      return hunk
    elseif lnum == line_count and hunk.added.start == line_count + 1 then
      return hunk
    elseif hunk.added.start <= lnum and vend >= lnum then
      return hunk
    end
  end
end

local CTRL_ESCAPES = {
  ["\a"] = "\\a",
  ["\b"] = "\\b",
  ["\f"] = "\\f",
  ["\n"] = "\\n",
  ["\r"] = "\\r",
  ["\t"] = "\\t",
  ["\v"] = "\\v",
}

-- git と同じ pathname quoting。囲まないとタブや改行を含むファイル名で
-- ヘッダ行が分割され、unified diff として読めなくなる。
function M.quote_path(path)
  if not path:find('[%c"\\]') then
    return path
  end
  local escaped = path:gsub('[\\"]', "\\%0"):gsub("%c", function(c)
    return CTRL_ESCAPES[c] or string.format("\\%03o", string.byte(c))
  end)
  return '"' .. escaped .. '"'
end

-- Claude 側で言語判定とファイル特定ができるよう、git 由来の diff と同じ
-- ヘッダ3行を付ける。
function M.diff_text(relpath, hunk)
  local a, b = M.quote_path("a/" .. relpath), M.quote_path("b/" .. relpath)
  local lines = {
    "diff --git " .. a .. " " .. b,
    "--- " .. a,
    "+++ " .. b,
    hunk.head,
  }
  vim.list_extend(lines, hunk.lines)
  return lines
end

-- 送信対象のファイルを作るところまで。送信は M.send が行う。
---@return { path: string, label: string }|nil, string|nil
function M.build(bufnr, lnum)
  local ok_gs, gs = pcall(require, "gitsigns")
  if not ok_gs then
    return nil, "gitsigns.nvim is not loaded"
  end

  local file = vim.api.nvim_buf_get_name(bufnr)
  if file == "" then
    return nil, "無名バッファには hunk がありません"
  end

  local hunks = gs.get_hunks(bufnr)
  if not hunks then
    return nil, "gitsigns がこのバッファに attach していません"
  end

  local hunk = M.find_hunk(hunks, lnum, vim.api.nvim_buf_line_count(bufnr))
  if not hunk then
    return nil, "カーソル位置に hunk がありません"
  end

  local relpath = vim.fn.fnamemodify(file, ":.")
  local path, err = require("utils.claudecode_tmpfile").create(M.diff_text(relpath, hunk), TMPFILE_OPTS)
  if not path then
    return nil, err
  end

  return { path = path, label = string.format("%s の hunk %s", relpath, hunk.head) }
end

function M.send()
  local built, err = M.build(vim.api.nvim_get_current_buf(), vim.api.nvim_win_get_cursor(0)[1])
  if not built then
    vim.notify(err or "unknown error", vim.log.levels.WARN)
    return
  end

  local ok, send_err = require("utils.claudecode_send").send(built.path, {
    context = "git-hunk",
    label = built.label,
  })
  if not ok then
    vim.notify("Claude送信に失敗: " .. (send_err or "unknown"), vim.log.levels.ERROR)
  end
end

return M
