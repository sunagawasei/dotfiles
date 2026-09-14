local M = {}

-- claude CLI は CLAUDE_CODE_SSE_PORT と一致する lock があればそれを選ぶ。
-- 同じ herdr workspace の pane が同じ値を持つので、この nvim と隣の claude が
-- 1 対 1 で結びつく。導出は shell 側の 1 箇所に置き、ここでは読むだけにする
-- (同じハッシュを二重に持つと片方の変更で静かに外れる)。
function M.pinned_port()
  local raw = os.getenv("CLAUDE_CODE_SSE_PORT")
  if not raw or raw == "" then
    return nil
  end

  local port = tonumber(raw)
  if not port or port ~= math.floor(port) or port < 1 or port > 65535 then
    return nil, string.format("CLAUDE_CODE_SSE_PORT がポート番号として不正です: %s", raw)
  end

  return port
end

-- libuv の bind() は使用中のポートでも成功し、エラーは listen() まで遅延する。
-- 空きの判定には listen まで通す必要がある。
local function listenable(port)
  local probe = vim.loop.new_tcp()
  if not probe then
    return false
  end
  local ok = probe:bind("127.0.0.1", port) and probe:listen(1, function() end)
  probe:close()
  return ok and true or false
end

local function warn(msg)
  -- setup() より前に呼ばれるため、UI が出てから通知する。
  vim.schedule(function()
    vim.notify(msg, vim.log.levels.WARN)
  end)
end

-- 固定ポートを使える場合だけ port_range を返す。nil を返した呼び出し元は
-- claudecode.nvim の既定範囲をそのまま使う。
---@return { min: integer, max: integer }|nil
function M.range_for_pinned_port()
  local port, err = M.pinned_port()
  if err then
    warn(err .. " — 既定のポート範囲で起動します")
    return nil
  end
  if not port then
    return nil
  end

  -- 範囲を 1 点にすると、埋まっていた場合に claudecode.nvim の
  -- find_available_port が nil を返してサーバごと起動しない。
  if not listenable(port) then
    warn(
      string.format(
        "CLAUDE_CODE_SSE_PORT=%d は使用中です。既定のポート範囲で起動しますが、この Neovim は同じ workspace の claude と自動で結びつきません",
        port
      )
    )
    return nil
  end

  return { min = port, max = port }
end

local isolated = false

-- claude CLI はポートが一致しないと lock の workspaceFolders と cwd で照合し、
-- 同じディレクトリを開いた別 workspace の Neovim を掴む。cwd の祖先になり得ない
-- 文字列を書いて、この照合を不成立にする。
---@return boolean ok, string|nil err
function M.isolate_lockfile_workspace()
  local workspace_id = os.getenv("HERDR_WORKSPACE_ID")
  if not workspace_id or workspace_id == "" then
    return true
  end
  if isolated then
    return true
  end

  local ok, lockfile = pcall(require, "claudecode.lockfile")
  if not ok then
    return false, "claudecode.lockfile を読み込めませんでした"
  end

  local original = lockfile.get_workspace_folders
  if type(original) ~= "function" then
    return false, "claudecode.lockfile.get_workspace_folders が見つかりません"
  end

  lockfile.get_workspace_folders = function(...)
    local tagged = {}
    for _, folder in ipairs(original(...)) do
      table.insert(tagged, folder .. "#herdr=" .. workspace_id)
    end
    return tagged
  end
  isolated = true

  return true
end

return M
