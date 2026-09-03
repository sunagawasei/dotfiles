local M = {}

-- ラッパーの適用状態と原関数をラップ対象と同じテーブルに置く。
-- 別の場所に置くと、片方だけreloadされたときに二重適用・二重通知になる。
local STATE_KEY = "_dotfiles_send_notify"

function M.label_for(file_path, start_line, end_line)
  local label = vim.fn.fnamemodify(file_path, ":~:.")
  if start_line and end_line then
    return string.format("%s:L%d-L%d", label, start_line + 1, end_line + 1)
  end
  return label
end

function M.notify(label, connected)
  if connected then
    vim.notify("Claude送信: " .. label, vim.log.levels.INFO)
    return
  end
  local cc = require("claudecode")
  local timeout = cc.state and cc.state.config and cc.state.config.queue_timeout
  local msg
  if type(timeout) == "number" then
    msg = string.format(
      "Claude未接続 — %s を保留(%s秒以内に接続しなければ破棄されます)",
      label,
      tostring(timeout / 1000)
    )
  else
    msg = string.format("Claude未接続 — %s を保留(接続しなければ破棄されます)", label)
  end
  vim.notify(msg, vim.log.levels.WARN)
end

-- 送信通知のラッパーを1度だけ適用する。claudecode.lua の config から呼ぶ。
function M.install()
  local cc = require("claudecode")
  if cc[STATE_KEY] then
    return
  end
  local orig = cc.send_at_mention
  cc[STATE_KEY] = { orig = orig }
  cc.send_at_mention = function(file_path, start_line, end_line, ...)
    local connected = cc.is_claude_connected()
    local ok, err = orig(file_path, start_line, end_line, ...)
    if ok then
      -- closure に束ねず都度 require する。helper だけを reload しても
      -- 両経路が同じ版の文面を使うため。
      local mod = require("utils.claudecode_send")
      mod.notify(mod.label_for(file_path, start_line, end_line), connected)
    end
    return ok, err
  end
end

-- 表示名を指定して送る。ラッパーを迂回して原関数を呼ぶので通知はここで1回だけ出す。
function M.send(file_path, opts)
  opts = opts or {}
  local cc = require("claudecode")
  local state = cc[STATE_KEY]
  local send = state and state.orig or cc.send_at_mention
  local connected = cc.is_claude_connected()
  local ok, err = send(file_path, opts.start_line, opts.end_line, opts.context)
  if ok then
    M.notify(opts.label or M.label_for(file_path, opts.start_line, opts.end_line), connected)
  end
  return ok, err
end

return M
