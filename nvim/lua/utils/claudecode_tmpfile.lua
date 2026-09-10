local M = {}

-- @mention はパスだけを送り、Claude がファイルを読むのはユーザーが
-- プロンプトを送信した時点。nvim の終了は読み取り完了の境界にならないので、
-- 削除は経過時間だけで決める。
local STALE_SECONDS = 24 * 60 * 60
local NAME_ATTEMPTS = 100

local seq = 0
local cleanup_registered = {}

-- group/other にビットが立っていないか。立っていれば他ユーザーが
-- ディレクトリを覗ける。
local function is_private(mode)
  return mode % 64 == 0
end

-- /tmp は他ユーザーが同名の symlink を先置きでき、writefile がそれを辿って
-- 任意ファイルを上書きしうる。実行ユーザーだけが辿れる 0700 の state 配下に置く。
local function ensure_dir(subdir)
  local dir = vim.fn.stdpath("state") .. "/" .. subdir
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

local function sweep(dir, ext)
  local now = os.time()
  for _, path in ipairs(vim.fn.glob(dir .. "/*." .. ext, false, true)) do
    if now - vim.fn.getftime(path) > STALE_SECONDS then
      vim.fn.delete(path)
    end
  end
end

-- 送信ごとと終了時の両方で掃く。どちらも期限超過分しか消さないので、
-- ユーザーが送信前のプロンプトに残している @mention は消えない。
local function register_cleanup(dir, subdir, ext)
  if cleanup_registered[dir] then
    return
  end
  cleanup_registered[dir] = true
  vim.api.nvim_create_autocmd("VimLeavePre", {
    group = vim.api.nvim_create_augroup("claudecode_tmpfile_cleanup_" .. subdir, { clear = true }),
    callback = function()
      sweep(dir, ext)
    end,
  })
end

-- "wx" は O_CREAT|O_EXCL なので、既存ファイルや先置きされた symlink を
-- 辿らずに EEXIST で弾ける。mode も作成時点から 0600 になる。
local function write_new(dir, ext, lines)
  local data = table.concat(lines, "\n") .. "\n"

  for _ = 1, NAME_ATTEMPTS do
    seq = seq + 1
    local path = string.format("%s/%d-%d.%s", dir, vim.fn.getpid(), seq, ext)
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

-- Claude へ渡す一時ファイルを1つ作り、そのパスを返す。
---@param lines string[] 書き込む行
---@param opts { subdir: string, ext: string }
---@return string|nil path, string|nil err
function M.create(lines, opts)
  local dir, dir_err = ensure_dir(opts.subdir)
  if not dir then
    return nil, dir_err
  end

  register_cleanup(dir, opts.subdir, opts.ext)
  sweep(dir, opts.ext)

  return write_new(dir, opts.ext, lines)
end

return M
