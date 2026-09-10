local M = {}

local STATE_SUBDIR = "claudecode-hunk"
-- @mention はパスだけを送り、Claude がファイルを読むのはユーザーが
-- プロンプトを送信した時点。nvim の終了は読み取り完了の境界にならないので、
-- 削除は経過時間だけで決める。
local STALE_SECONDS = 24 * 60 * 60
local NAME_ATTEMPTS = 100

local seq = 0
local cleanup_registered = false

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

-- group/other にビットが立っていないか。立っていれば他ユーザーが
-- ディレクトリを覗ける。
local function is_private(mode)
  return mode % 64 == 0
end

local function ensure_dir()
  local dir = vim.fn.stdpath("state") .. "/" .. STATE_SUBDIR
  vim.fn.mkdir(dir, "p", tonumber("700", 8))

  -- mkdir の mode は新規作成時しか効かない。既存ディレクトリが緩ければ
  -- 締め直し、締められなければ書き込まずに失敗させる。
  local st = vim.loop.fs_stat(dir)
  if not st then
    return nil, "state ディレクトリを作成できません: " .. dir
  end
  if not is_private(st.mode) then
    vim.loop.fs_chmod(dir, tonumber("700", 8))
    st = vim.loop.fs_stat(dir)
    if not st or not is_private(st.mode) then
      return nil, string.format("state ディレクトリの権限を 0700 にできません: %s", dir)
    end
  end

  return dir
end

local function sweep(dir)
  local now = os.time()
  for _, path in ipairs(vim.fn.glob(dir .. "/*.diff", false, true)) do
    if now - vim.fn.getftime(path) > STALE_SECONDS then
      vim.fn.delete(path)
    end
  end
end

-- 送信ごとと終了時の両方で掃く。どちらも期限超過分しか消さないので、
-- ユーザーが送信前のプロンプトに残している @mention は消えない。
local function register_cleanup(dir)
  if cleanup_registered then
    return
  end
  cleanup_registered = true
  vim.api.nvim_create_autocmd("VimLeavePre", {
    group = vim.api.nvim_create_augroup("claudecode_hunk_cleanup", { clear = true }),
    callback = function()
      sweep(dir)
    end,
  })
end

-- "wx" は O_CREAT|O_EXCL なので、既存ファイルや先置きされた symlink を
-- 辿らずに EEXIST で弾ける。mode も作成時点から 0600 になる。
local function write_new(dir, lines)
  local data = table.concat(lines, "\n") .. "\n"

  for _ = 1, NAME_ATTEMPTS do
    seq = seq + 1
    local path = string.format("%s/%d-%d.diff", dir, vim.fn.getpid(), seq)
    local fd, open_err, open_errname = vim.loop.fs_open(path, "wx", tonumber("600", 8))
    if not fd then
      -- 名前を変えて解決するのは衝突のときだけ。EACCES や ENOSPC を
      -- 100回試して「名前を確保できない」と誤報しない。
      if open_errname ~= "EEXIST" then
        return nil, "一時ファイルを作成できません: " .. (open_err or "unknown error")
      end
    else
      local pos = 0
      while pos < #data do
        local ok, written = pcall(vim.loop.fs_write, fd, data:sub(pos + 1), pos)
        if not ok or type(written) ~= "number" or written <= 0 then
          pcall(vim.loop.fs_close, fd)
          vim.loop.fs_unlink(path)
          return nil, "一時ファイルの書き込みに失敗しました: " .. path
        end
        pos = pos + written
      end

      local ok_close, closed = pcall(vim.loop.fs_close, fd)
      if not ok_close or not closed then
        vim.loop.fs_unlink(path)
        return nil, "一時ファイルを閉じられませんでした: " .. path
      end
      return path
    end
  end

  return nil, string.format("一時ファイル名を %d 回試して確保できませんでした", NAME_ATTEMPTS)
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

  local dir, dir_err = ensure_dir()
  if not dir then
    return nil, dir_err
  end
  register_cleanup(dir)
  sweep(dir)

  local relpath = vim.fn.fnamemodify(file, ":.")
  local path, write_err = write_new(dir, M.diff_text(relpath, hunk))
  if not path then
    return nil, write_err
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
