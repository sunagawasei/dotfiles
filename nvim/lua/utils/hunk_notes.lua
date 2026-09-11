-- hunk(ターミナル差分レビューTUI)のフォーカス位置とレビューノートを
-- Claude Code へ@参照として渡す。hunkへの書き戻しは行わない。
local M = {}

local TMPFILE_OPTS = { subdir = "claudecode-hunk-notes", ext = "md" }
local VALID_COMMANDS = { ["hunk diff"] = true, ["hunk diff --staged"] = true }
-- CLI呼び出しを3回以上に増やすなら上限を再評価する
local SESSION_LIST_TIMEOUT_MS = 1500
local LABEL_MAX_CHARS = 60

local function is_present(v)
  return v ~= nil and v ~= vim.NIL
end

local function is_nonempty_string(v)
  return type(v) == "string" and v ~= ""
end

local function is_nonneg_int(v)
  return type(v) == "number" and v >= 0 and v == math.floor(v)
end

-- 要素数2・両要素とも0以上の整数のrangeだけを受理する。
local function normalize_range(range)
  if type(range) ~= "table" or not vim.islist(range) or #range ~= 2 then
    return nil, "range の形が不正です(型不正)"
  end
  for _, v in ipairs(range) do
    if not is_nonneg_int(v) then
      return nil, "range の要素が0以上の整数ではありません(型不正)"
    end
  end
  return { range[1], range[2] }
end

-- focusは4フィールドを組として扱う。受理するのは
-- (A) 4つ揃う (B) selectedFilePathと2つのrangeが3つとも欠落、の2状態だけ。
-- 存在判定はselectedFilePathの有無を使う(selectedHunkIndexは差分が空でも0が残るため使わない)。
local function normalize_focus(state)
  local has_path = is_present(state.selectedFilePath)
  local has_old = is_present(state.selectedHunkOldRange)
  local has_new = is_present(state.selectedHunkNewRange)
  local present_count = (has_path and 1 or 0) + (has_old and 1 or 0) + (has_new and 1 or 0)

  if present_count == 0 then
    return nil -- no-focus (受理)
  end
  if present_count < 3 then
    return nil, "focus のフィールドが一部だけ欠落しています(field欠落)"
  end

  if not is_nonempty_string(state.selectedFilePath) then
    return nil, "selectedFilePath が非空文字列ではありません(型不正)"
  end
  local old_range, old_err = normalize_range(state.selectedHunkOldRange)
  if not old_range then
    return nil, "selectedHunkOldRange が不正です: " .. old_err
  end
  local new_range, new_err = normalize_range(state.selectedHunkNewRange)
  if not new_range then
    return nil, "selectedHunkNewRange が不正です: " .. new_err
  end
  if not is_present(state.selectedHunkIndex) then
    return nil, "selectedHunkIndex がありません(field欠落)"
  end
  if not is_nonneg_int(state.selectedHunkIndex) then
    return nil, "selectedHunkIndex が0以上の整数ではありません(型不正)"
  end

  return {
    filePath = state.selectedFilePath,
    oldRange = old_range,
    newRange = new_range,
    hunkIndex = state.selectedHunkIndex,
  }
end

local function normalize_note(raw)
  if type(raw) ~= "table" then
    return nil, "note要素がtableではありません(型不正)"
  end
  if not is_nonempty_string(raw.filePath) then
    return nil, "note.filePath が非空文字列ではありません(field欠落または型不正)"
  end
  if type(raw.body) ~= "string" then
    return nil, "note.body が string ではありません(field欠落または型不正)"
  end

  local note = { filePath = raw.filePath, body = raw.body }
  local has_new = is_present(raw.newRange)
  local has_old = is_present(raw.oldRange)
  if not has_new and not has_old then
    return nil, "note の newRange / oldRange が両方とも欠落しています(field欠落)"
  end
  if has_new then
    local r, err = normalize_range(raw.newRange)
    if not r then
      return nil, "note.newRange が不正です: " .. err
    end
    note.newRange = r
  end
  if has_old then
    local r, err = normalize_range(raw.oldRange)
    if not r then
      return nil, "note.oldRange が不正です: " .. err
    end
    note.oldRange = r
  end
  return note
end

-- reviewNoteCountはoptional metadata。配列長が唯一の正で、
-- 存在するときだけ配列長との一致を検査する。
local function normalize_notes(state)
  if type(state.reviewNotes) ~= "table" or not vim.islist(state.reviewNotes) then
    return nil, "reviewNotes が配列ではありません(field欠落または型不正)"
  end

  if is_present(state.reviewNoteCount) then
    if not is_nonneg_int(state.reviewNoteCount) then
      return nil, "reviewNoteCount が0以上の整数ではありません(型不正)"
    end
    if state.reviewNoteCount ~= #state.reviewNotes then
      return nil, "reviewNoteCount が reviewNotes の配列長と一致しません(型不正)"
    end
  end

  local notes = {}
  for i, raw in ipairs(state.reviewNotes) do
    local note, err = normalize_note(raw)
    if not note then
      return nil, string.format("reviewNotes[%d] が不正です: %s", i - 1, err)
    end
    table.insert(notes, note)
  end
  return notes
end

-- pidの型検査はM.parse_sessions(全要素対象、matchの前提)側の責務。
-- ここではmatch_sessionが選んだ1件の残りのfieldだけを検査する。
---@param raw table `hunk session list --json`の生セッション要素(pid検査済み)
---@return table|nil session, string|nil err
function M.normalize_session(raw)
  if type(raw) ~= "table" then
    return nil, "session要素がtableではありません(型不正)"
  end
  if not is_nonempty_string(raw.sessionId) then
    return nil, "session.sessionId が非空文字列ではありません(field欠落または型不正)"
  end
  if not is_nonempty_string(raw.repoRoot) then
    return nil, "session.repoRoot が非空文字列ではありません(field欠落または型不正)"
  end
  if is_present(raw.title) and type(raw.title) ~= "string" then
    return nil, "session.title が string ではありません(型不正)"
  end
  local title = is_present(raw.title) and raw.title or nil

  if type(raw.snapshot) ~= "table" or type(raw.snapshot.state) ~= "table" then
    return nil, "snapshot.state がありません(field欠落。hunk の版上げの可能性)"
  end
  local state = raw.snapshot.state

  local focus, focus_err = normalize_focus(state)
  if focus_err then
    return nil, focus_err
  end

  local notes, notes_err = normalize_notes(state)
  if not notes then
    return nil, notes_err
  end

  return {
    sessionId = raw.sessionId,
    pid = raw.pid,
    repoRoot = raw.repoRoot,
    title = title,
    focus = focus,
    notes = notes,
  }
end

-- `hunk session list --json` の応答テキストを検証する最小限の第1段。
-- pid照合に必要な形(rootがtable・sessionsが配列・各要素がtableでpidがnumber)
-- だけを見る。他要素の中身(snapshot.state等)はここでは検査しない。
-- 対象外セッションの型崩れで対象セッションの処理まで落とさないため。
-- vim.system/vim.fnを呼ばない純関数(vim.json.decodeの例外はpcallで吸収する)。
---@param text string
---@return table[]|nil sessions, string|nil err
function M.parse_sessions(text)
  if type(text) ~= "string" or text == "" then
    return nil, "hunk session list の応答が空です"
  end

  local ok, decoded = pcall(vim.json.decode, text)
  if not ok then
    return nil, "hunk session list の応答をJSONとして読めません(hunk の版上げの可能性): " .. tostring(decoded)
  end
  if type(decoded) ~= "table" or not vim.islist(decoded.sessions) then
    return nil, "応答の sessions が配列ではありません(field欠落または型不正。hunk の版上げの可能性)"
  end

  for i, raw in ipairs(decoded.sessions) do
    if type(raw) ~= "table" then
      return nil, string.format("sessions[%d] がtableではありません(型不正)", i - 1)
    end
    if type(raw.pid) ~= "number" then
      return nil, string.format("sessions[%d].pid が number ではありません(型不正)", i - 1)
    end
  end
  return decoded.sessions
end

-- pidに一致するセッションをちょうど1件だけ選ぶ。0件・2件以上は失敗として
-- 扱い、他セッションの要素は戻り値に一切混ぜない。
---@param sessions table[]
---@param pid number
---@return table|nil session, string|nil err
function M.match_session(sessions, pid)
  local matched, count = nil, 0
  for _, session in ipairs(sessions or {}) do
    if session.pid == pid then
      matched = session
      count = count + 1
    end
  end
  if count == 0 then
    return nil, string.format("pid %s に一致する hunk セッションがありません", tostring(pid))
  end
  if count > 1 then
    return nil, string.format("pid %s に一致する hunk セッションが複数あります", tostring(pid))
  end
  return matched
end

-- 通知に載せる表示名。titleはhunkが生成する自由文字列で改行等を含みうるため、
-- 制御文字を潰し長さを切り詰める(通知本文が複数行に伸びるのを防ぐ)。
local function sanitize_label(text)
  local cleaned = text:gsub("%c", " ")
  if vim.fn.strcharlen(cleaned) > LABEL_MAX_CHARS then
    cleaned = vim.fn.strcharpart(cleaned, 0, LABEL_MAX_CHARS - 1) .. "…"
  end
  return cleaned
end

local function longest_backtick_run(text)
  local max_run = 0
  for run in text:gmatch("`+") do
    max_run = math.max(max_run, #run)
  end
  return max_run
end

-- 固定のMarkdown。可変値は一切含めない(ノート本文・title・repoRoot・パスは
-- すべてこの後ろのJSONフェンス側に置く)。
local HEADER_LINES = {
  "# hunk review context",
  "",
  "## how to read",
  "",
  "- この packet はユーザーの hunk 画面を読むためのものです。変更しないでください。",
  "- 使ってよいコマンド: `hunk session review <sessionId> --json`(全体の構造把握)、"
    .. "`hunk session review <sessionId> --include-patch --json`(全 file の raw diff。大きい)",
  "- 使ってはいけないコマンド: `hunk session comment add` / `comment apply` / `comment rm` / "
    .. "`comment clear` / `navigate` / `reload`。hunk へは書き戻さないでください。回答はこの"
    .. "セッションで返してください。",
  "- 対象 file の diff を取る推奨経路: `repoRoot` で `git diff -- <filePath>`。`command` が "
    .. "`hunk diff --staged` なら `git diff --cached -- <filePath>`。untracked file は git diff "
    .. "に出ないので file を直接読んでください。取れた範囲が packet の range と合わなければ、"
    .. "`hunk session review <sessionId> --include-patch --json` にフォールバックしてください"
    .. "(ユーザーが hunk 内で表示内容を差し替えている場合があります)。",
  "- `hunkIndex` は0始まりで、`review --json` の `hunks[].index` と同じ基数です。CLI 引数の "
    .. "`--hunk` は1始まりなので、渡すときは+1してください。",
  "- ノートの最新が要る場合は `hunk session list --json` の同じ `sessionId` の `reviewNotes` を"
    .. "読み直してください。",
  "- 本文が無いときの既定:",
  "  - ユーザーが入力欄に本文を書いていればそれに従ってください。",
  "  - 本文が無く、ノートがあれば各ノートに答えてください。`focus` がノートの対象と異なる hunk "
    .. "なら最後に一言触れてください。",
  "  - 本文が無く、ノートも無ければ、`focus` の hunk が何をしている変更かを説明し、気づいた点"
    .. "(正しさ・抜け・読みにくさ)があれば挙げてください。気づいた点が無ければ無いと言ってくだ"
    .. "さい。要点を先に。",
  "",
  "## if session is gone",
  "",
  "- この `sessionId` が `hunk session list` に無ければ何もしないでください。稼働中のセッション"
    .. "が他に1件しか無くても選ばないでください。「Neovim の hunk 画面で `<C-a>` をもう一度押し"
    .. "てください」とユーザーに伝えて終わってください。",
  "",
  "## data",
  "",
}

-- packet本文を生成する純関数。可変部はJSON側の1オブジェクト・1フェンスに閉じ込める。
---@param session table normalize_sessionが返す形
---@param cmd string "hunk diff" | "hunk diff --staged"
---@return string[]|nil lines, string|nil err
function M.build_packet(session, cmd)
  if not VALID_COMMANDS[cmd] then
    return nil, "command が不明です: " .. tostring(cmd)
  end
  if not session.focus and #session.notes == 0 then
    return nil, "focus も notes もありません"
  end

  local data = {
    sessionId = session.sessionId,
    repoRoot = session.repoRoot,
    title = session.title,
    command = cmd,
    focus = session.focus,
    notes = session.notes,
  }
  local json = vim.json.encode(data)
  -- vim.json.encodeはバッククォートをエスケープしないため、本文中の連続数を
  -- 上回る長さのフェンスでないと ``` を含むノートでフェンスが破れる。
  local fence = string.rep("`", math.max(3, longest_backtick_run(json) + 1))

  local lines = vim.deepcopy(HEADER_LINES)
  table.insert(lines, fence .. "json")
  table.insert(lines, json)
  table.insert(lines, fence)
  return lines
end

-- a〜hのI/Oパイプライン。既知の失敗は(false, err)を返す。
-- vim.systemのspawn例外などはここで捕まえず、呼び出し元のpcallに委ねる。
local function pipeline(term, cmd)
  local pid = vim.fn.jobpid(term.job_id)

  local result = vim
    .system({ "hunk", "session", "list", "--json" }, { text = true, timeout = SESSION_LIST_TIMEOUT_MS })
    :wait()
  if result.code ~= 0 then
    return false, string.format("hunk session list が失敗しました(exit %d): %s", result.code, result.stderr or "")
  end

  local raw_sessions, sessions_err = M.parse_sessions(result.stdout)
  if not raw_sessions then
    return false, sessions_err
  end

  local raw_session, match_err = M.match_session(raw_sessions, pid)
  if not raw_session then
    return false, match_err
  end

  -- 型の深い検査はmatch確定後、対象の1件にだけ行う(他セッションの型崩れに引きずられない)
  local session, normalize_err = M.normalize_session(raw_session)
  if not session then
    return false, normalize_err
  end

  if vim.fn.jobwait({ term.job_id }, 0)[1] ~= -1 then
    return false, "hunk セッションのジョブが終了しています"
  end

  local packet, packet_err = M.build_packet(session, cmd)
  if not packet then
    return false, packet_err
  end

  -- tmpfile作成からsendのpcallまでの間に例外を投げうる式を残さないため、
  -- 送信関数の解決とlabelの組み立てをtmpfile作成より前に済ませておく
  local send_fn = require("utils.claudecode_send").send
  local send_opts = {
    context = "hunk-notes",
    label = sanitize_label(string.format("hunk レビュー: %s", session.title or session.repoRoot)),
  }

  local path, tmp_err = require("utils.claudecode_tmpfile").create(packet, TMPFILE_OPTS)
  if not path then
    return false, tmp_err
  end

  -- send自体の例外もdeleteに落とす(戻り値失敗・例外のどちらでも一時ファイルを残さない)
  local no_send_exception, send_ret1, send_ret2 = pcall(send_fn, path, send_opts)
  local send_ok, send_err
  if no_send_exception then
    send_ok, send_err = send_ret1, send_ret2
  else
    send_ok, send_err = false, send_ret1
  end
  if not send_ok then
    pcall(vim.fn.delete, path)
    return false, send_err
  end

  return true
end

-- キーマップから直接呼ばれる公開エントリ。パイプライン全体を1回pcallし、
-- 想定外の例外も含めて通知は1回だけ出す。
---@param term table toggletermのTerminal
---@param cmd string "hunk diff" | "hunk diff --staged"
function M.send(term, cmd)
  local no_exception, ret1, ret2 = pcall(pipeline, term, cmd)
  local ok, err
  if no_exception then
    ok, err = ret1, ret2
  else
    ok, err = false, ret1
  end
  if not ok then
    vim.notify("hunk レビューノートの送信に失敗しました: " .. tostring(err), vim.log.levels.WARN)
  end
  return ok, err
end

return M
