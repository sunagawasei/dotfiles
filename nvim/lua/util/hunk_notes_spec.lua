local hn = require("utils.hunk_notes")

local total = 0

-- 2026-09-11、hunkdiff 0.16.0で実機採取した`hunk session list --json`のfixture。
-- フォーカスは2つ目のhunk(index 1)、ノートは1件でバッククォート3連と##を含む本文を持つ。
local FIXTURE_JSON = [[
{
  "sessions": [
    {
      "sessionId": "5ed3fe6f-4d8c-45e4-8b7b-d678c2d690b1",
      "pid": 31091,
      "cwd": "/tmp/fx",
      "launchedAt": "2026-09-11T02:56:30.778Z",
      "inputKind": "vcs",
      "title": "fx working tree",
      "sourceLabel": "/tmp/fx",
      "repoRoot": "/tmp/fx",
      "fileCount": 1,
      "files": [
        { "id": "/tmp/fx:0:a.txt", "path": "a.txt", "additions": 2, "deletions": 2, "hunkCount": 2 }
      ],
      "snapshot": {
        "updatedAt": "2026-09-11T02:56:35.953Z",
        "state": {
          "selectedFileId": "/tmp/fx:0:a.txt",
          "selectedFilePath": "a.txt",
          "selectedHunkIndex": 1,
          "selectedHunkOldRange": [28, 34],
          "selectedHunkNewRange": [28, 34],
          "showAgentNotes": false,
          "liveCommentCount": 0,
          "liveComments": [],
          "reviewNoteCount": 1,
          "reviewNotes": [
            {
              "noteId": "user:1789095394549",
              "source": "user",
              "filePath": "a.txt",
              "hunkIndex": 0,
              "newRange": [2, 2],
              "body": "ここ ``` と ## を含む本文",
              "author": "user",
              "createdAt": "2026-09-11T02:56:34.549Z",
              "editable": true
            }
          ]
        }
      }
    }
  ]
}
]]

local fixture = vim.json.decode(FIXTURE_JSON)
local fixture_state = fixture.sessions[1].snapshot.state
local fixture_note = fixture_state.reviewNotes[1]

-- M.parse_sessions(第1段)向け: raw JSON形の生セッション要素。
-- 呼び出しごとに新規tableを返し、テスト間でmutationが漏れないようにする。
local function make_valid_session()
  return {
    sessionId = "sess-1",
    pid = 100,
    repoRoot = "/repo",
    title = "t",
    snapshot = {
      state = {
        selectedFilePath = "a.txt",
        selectedHunkOldRange = { 1, 2 },
        selectedHunkNewRange = { 1, 2 },
        selectedHunkIndex = 0,
        reviewNotes = {},
      },
    },
  }
end

local function sessions_json(list)
  return vim.json.encode({ sessions = list })
end

-- build_packet向け: M.normalize_sessionが返す形をした最小の有効セッション。
-- focusはno_focus=trueを渡さない限り既定値が入る(未指定時はnilと区別できないため)。
local function make_normalized_session(overrides)
  overrides = overrides or {}
  local focus = overrides.focus
  if focus == nil and not overrides.no_focus then
    focus = { filePath = "a.txt", oldRange = { 1, 2 }, newRange = { 1, 2 }, hunkIndex = 0 }
  end
  return {
    sessionId = overrides.sessionId or "sess-1",
    pid = overrides.pid or 100,
    repoRoot = overrides.repoRoot or "/repo",
    title = overrides.title,
    focus = focus,
    notes = overrides.notes or {},
  }
end

-- 生成linesの末尾の```json ... ```フェンスを取り出す。戻り値2つ目はフェンスの
-- バッククォート本数(長さ検査に使う)。
local function extract_fenced_json(lines)
  local n = #lines
  local close_ticks = lines[n]:match("^(`+)$")
  assert(close_ticks, "closing fence not found on last line: " .. tostring(lines[n]))
  local json_line = lines[n - 1]
  local open_ticks = lines[n - 2]:match("^(`+)json$")
  assert(open_ticks, "opening fence not found: " .. tostring(lines[n - 2]))
  assert(#open_ticks == #close_ticks, "fence length mismatch (open vs close)")
  return vim.json.decode(json_line), #open_ticks
end

-- fixture丸ごとが3つの純関数(parse_sessions → match_session → normalize_session)を
-- 素通りすること(型検査の実データ確認)
do
  local raw_sessions, err = hn.parse_sessions(FIXTURE_JSON)
  assert(raw_sessions ~= nil, "fixture: parse_sessions failed: " .. tostring(err))
  assert(#raw_sessions == 1 and raw_sessions[1].pid == 31091, "fixture: unexpected sessions shape")

  local matched, match_err = hn.match_session(raw_sessions, 31091)
  assert(matched ~= nil, "fixture: match_session failed: " .. tostring(match_err))

  local session, norm_err = hn.normalize_session(matched)
  assert(session ~= nil, "fixture: normalize_session failed: " .. tostring(norm_err))
  assert(session.focus ~= nil and session.focus.hunkIndex == 1, "fixture: focus.hunkIndex should be 1")
  assert(#session.notes == 1, "fixture: notes count mismatch")
  total = total + 1
end

-- ==== セッション照合 ====

-- 1. pid一致0件 → 失敗
do
  local sessions = { { pid = 1 }, { pid = 2 } }
  local matched, err = hn.match_session(sessions, 999)
  assert(matched == nil and err ~= nil, "match: no match should fail")
  total = total + 1
end

-- 2. pid一致1件 → その要素を返す
do
  local target = { pid = 2, sessionId = "s2" }
  local sessions = { { pid = 1 }, target, { pid = 3 } }
  local matched, err = hn.match_session(sessions, 2)
  assert(matched == target and err == nil, "match: single match should return that element")
  total = total + 1
end

-- 3. pid一致2件以上 → 失敗
do
  local sessions = { { pid = 5 }, { pid = 5 } }
  local matched, err = hn.match_session(sessions, 5)
  assert(matched == nil and err ~= nil, "match: duplicate pid should fail")
  total = total + 1
end

-- 4. 複数セッションを含む入力で、pid一致した1件のノートだけがpacketに載る
do
  local session_a = make_valid_session()
  session_a.sessionId = "sess-a"
  session_a.pid = 111
  session_a.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "note-A", newRange = { 1, 2 } } }

  local session_b = make_valid_session()
  session_b.sessionId = "sess-b"
  session_b.pid = 222
  session_b.snapshot.state.reviewNotes = { { filePath = "b.txt", body = "note-B", newRange = { 1, 2 } } }

  local raw_sessions, err = hn.parse_sessions(sessions_json({ session_a, session_b }))
  assert(raw_sessions ~= nil, "match/multi: parse_sessions failed: " .. tostring(err))

  local matched = hn.match_session(raw_sessions, 111)
  assert(matched ~= nil and matched.sessionId == "sess-a", "match/multi: expected sess-a")

  local session, norm_err = hn.normalize_session(matched)
  assert(session ~= nil, "match/multi: normalize_session failed: " .. tostring(norm_err))

  local lines = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "match/multi: build_packet failed")
  local data = extract_fenced_json(lines)
  assert(#data.notes == 1 and data.notes[1].body == "note-A", "match/multi: only matched session's note should appear")
  total = total + 1
end

-- K2回帰: snapshot.state欠落/reviewNotes不正の対象外セッションが配列の前・後どちらに
-- あっても、pid一致した対象セッションだけでmatch→normalize→build_packetまで成功する
for _, order in ipairs({ "before", "after" }) do
  local bad = { pid = 999, snapshot = { state = { reviewNotes = "garbage-not-array" } } }
  local good = make_valid_session()
  good.pid = 12345

  local list = order == "before" and { bad, good } or { good, bad }
  local raw_sessions, err1 = hn.parse_sessions(sessions_json(list))
  assert(raw_sessions ~= nil, "K2(" .. order .. "): minimal parse should tolerate malformed other session: " .. tostring(err1))

  local matched, err2 = hn.match_session(raw_sessions, good.pid)
  assert(matched ~= nil, "K2(" .. order .. "): match should find target: " .. tostring(err2))

  local normalized, err3 = hn.normalize_session(matched)
  assert(normalized ~= nil, "K2(" .. order .. "): normalize should succeed for target: " .. tostring(err3))

  local lines = hn.build_packet(normalized, "hunk diff")
  assert(lines ~= nil, "K2(" .. order .. "): build_packet should succeed")
  total = total + 1
end

-- ==== JSON型検査: M.parse_sessions(第1段。root/配列/pidの型だけ) ====

-- 5. 空文字列 / malformed JSON → 失敗
do
  local sessions, err = hn.parse_sessions("")
  assert(sessions == nil and err ~= nil, "parse: empty string should fail")
  total = total + 1
end
do
  local sessions, err = hn.parse_sessions("{not json")
  assert(sessions == nil and err ~= nil, "parse: malformed json should fail")
  total = total + 1
end

-- 6. sessionsが配列でない → 失敗
do
  local sessions, err = hn.parse_sessions(vim.json.encode({ sessions = { foo = "bar" } }))
  assert(sessions == nil and err ~= nil, "parse: sessions as object should fail")
  total = total + 1
end
do
  local sessions, err = hn.parse_sessions(vim.json.encode({ sessions = "not-an-array" }))
  assert(sessions == nil and err ~= nil, "parse: sessions as string should fail")
  total = total + 1
end
do
  local sessions, err = hn.parse_sessions(vim.json.encode({ foo = "bar" }))
  assert(sessions == nil and err ~= nil, "parse: sessions missing entirely should fail")
  total = total + 1
end

-- 7a. pidがnumberでない(第1段の担当) → 失敗
do
  local sessions, err = hn.parse_sessions(sessions_json({ { pid = "100" } }))
  assert(sessions == nil and err ~= nil, "parse: pid as string should fail")
  total = total + 1
end

-- ==== JSON型検査: M.normalize_session(match後の対象1件だけに掛ける第2段) ====

-- 7b. sessionIdが非string → 失敗
do
  local s = make_valid_session()
  s.sessionId = 12345
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: sessionId as number should fail")
  total = total + 1
end

-- 8. snapshot.state 欠落 → 失敗
do
  local s = make_valid_session()
  s.snapshot.state = nil
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: missing snapshot.state should fail")
  total = total + 1
end

-- 9. focus4フィールド全部揃う → 受理
do
  local s = make_valid_session()
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: valid focus should be accepted: " .. tostring(err))
  assert(session.focus ~= nil and session.focus.filePath == "a.txt", "normalize: focus shape mismatch")
  total = total + 1
end

-- 10. selectedFilePathと2つのrangeが3つとも欠落(index=0のみ残る) → 受理(no-focus)
do
  local s = make_valid_session()
  s.snapshot.state.selectedFilePath = nil
  s.snapshot.state.selectedHunkOldRange = nil
  s.snapshot.state.selectedHunkNewRange = nil
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: no-focus state should be accepted: " .. tostring(err))
  assert(session.focus == nil, "normalize: focus should be nil when path/ranges all absent")
  total = total + 1
end

-- 11. focusの部分欠落(pathだけ有る/rangeが片方だけ/indexだけ無い) → すべて失敗
local partial_focus_cases = {
  {
    name = "path only",
    mutate = function(state)
      state.selectedHunkOldRange = nil
      state.selectedHunkNewRange = nil
    end,
  },
  {
    name = "range only one side",
    mutate = function(state)
      state.selectedFilePath = nil
      state.selectedHunkNewRange = nil
    end,
  },
  {
    name = "index missing only",
    mutate = function(state)
      state.selectedHunkIndex = nil
    end,
  },
}
for _, c in ipairs(partial_focus_cases) do
  local s = make_valid_session()
  c.mutate(s.snapshot.state)
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: partial focus (" .. c.name .. ") should fail")
  total = total + 1
end

-- 12. rangeの要素数不足・過剰・非整数・負数 → 失敗
local bad_range_cases = {
  { name = "too few elements", range = { 1 } },
  { name = "too many elements", range = { 1, 2, 3 } },
  { name = "non-integer", range = { 1.5, 2 } },
  { name = "negative", range = { -1, 2 } },
}
for _, c in ipairs(bad_range_cases) do
  local s = make_valid_session()
  s.snapshot.state.selectedHunkOldRange = c.range
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: bad range (" .. c.name .. ") should fail")
  total = total + 1
end

-- 13. selectedHunkIndexが小数・負数 → 失敗
local bad_index_cases = {
  { name = "float index", index = 1.5 },
  { name = "negative index", index = -1 },
}
for _, c in ipairs(bad_index_cases) do
  local s = make_valid_session()
  s.snapshot.state.selectedHunkIndex = c.index
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: bad hunkIndex (" .. c.name .. ") should fail")
  total = total + 1
end

-- 14. reviewNoteCount欠落 → 受理(配列長を正とする)
do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n", newRange = { 1, 2 } } }
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: missing reviewNoteCount should be accepted: " .. tostring(err))
  assert(#session.notes == 1, "normalize: notes count mismatch")
  total = total + 1
end

-- 15. reviewNoteCountが配列長と不一致/負数/小数 → 失敗
local bad_count_cases = {
  { name = "mismatch", count = 5 },
  { name = "negative", count = -1 },
  { name = "decimal", count = 0.5 },
}
for _, c in ipairs(bad_count_cases) do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n", newRange = { 1, 2 } } }
  s.snapshot.state.reviewNoteCount = c.count
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: bad reviewNoteCount (" .. c.name .. ") should fail")
  total = total + 1
end

-- 16. noteのfilePath欠落・空文字 / body欠落・非string → 失敗
local bad_note_field_cases = {
  { name = "filePath missing", note = { body = "n", newRange = { 1, 2 } } },
  { name = "filePath empty", note = { filePath = "", body = "n", newRange = { 1, 2 } } },
  { name = "body missing", note = { filePath = "a.txt", newRange = { 1, 2 } } },
  { name = "body non-string", note = { filePath = "a.txt", body = 123, newRange = { 1, 2 } } },
}
for _, c in ipairs(bad_note_field_cases) do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { c.note }
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: bad note (" .. c.name .. ") should fail")
  total = total + 1
end

-- 17. noteの位置はnewRangeのみ/oldRangeのみ/両方 → すべて受理。両方欠落 → 失敗
do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n", newRange = { 1, 2 } } }
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: note with newRange only should be accepted: " .. tostring(err))
  assert(session.notes[1].newRange ~= nil and session.notes[1].oldRange == nil, "normalize: newRange-only shape")
  total = total + 1
end
do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n", oldRange = { 1, 2 } } }
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: note with oldRange only should be accepted: " .. tostring(err))
  assert(session.notes[1].oldRange ~= nil and session.notes[1].newRange == nil, "normalize: oldRange-only shape")
  total = total + 1
end
do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n", oldRange = { 1, 2 }, newRange = { 3, 4 } } }
  local session, err = hn.normalize_session(s)
  assert(session ~= nil, "normalize: note with both ranges should be accepted: " .. tostring(err))
  assert(session.notes[1].oldRange ~= nil and session.notes[1].newRange ~= nil, "normalize: both-ranges shape")
  total = total + 1
end
do
  local s = make_valid_session()
  s.snapshot.state.reviewNotes = { { filePath = "a.txt", body = "n" } }
  local session, err = hn.normalize_session(s)
  assert(session == nil and err ~= nil, "normalize: note with both ranges missing should fail")
  total = total + 1
end

-- ==== packet本文 ====

-- 18. フェンスの中身を構造比較する(文字列比較にしない)
do
  local session = make_normalized_session({})
  local lines = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "packet: build_packet failed")
  local data = extract_fenced_json(lines)
  assert(data.sessionId == session.sessionId, "packet: sessionId mismatch")
  assert(data.repoRoot == session.repoRoot, "packet: repoRoot mismatch")
  assert(data.command == "hunk diff", "packet: command mismatch")
  assert(data.focus.filePath == session.focus.filePath, "packet: focus.filePath mismatch")
  assert(vim.deep_equal(data.focus.oldRange, session.focus.oldRange), "packet: focus.oldRange mismatch")
  assert(vim.deep_equal(data.focus.newRange, session.focus.newRange), "packet: focus.newRange mismatch")
  total = total + 1
end

-- fixtureの実データ(focus index1、バッククォート3連+##を含むノート)がpacketを壊さない
do
  local session = make_normalized_session({
    sessionId = fixture.sessions[1].sessionId,
    repoRoot = fixture.sessions[1].repoRoot,
    focus = {
      filePath = fixture_state.selectedFilePath,
      oldRange = fixture_state.selectedHunkOldRange,
      newRange = fixture_state.selectedHunkNewRange,
      hunkIndex = fixture_state.selectedHunkIndex,
    },
    notes = { { filePath = fixture_note.filePath, body = fixture_note.body, newRange = fixture_note.newRange } },
  })
  local lines = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "packet/fixture: build_packet failed")
  local data, fence_len = extract_fenced_json(lines)
  assert(fence_len == 4, "packet/fixture: fence length should be 4, got " .. fence_len)
  assert(data.notes[1].body == fixture_note.body, "packet/fixture: note body mismatch")
  assert(data.focus.hunkIndex == 1, "packet/fixture: focus.hunkIndex should be 1")
  total = total + 1
end

-- 19. ノート0件のとき notes が [] になる({}ではない)
do
  local session = make_normalized_session({ notes = {} })
  local lines = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "packet: build_packet failed for empty notes")
  local data = extract_fenced_json(lines)
  assert(vim.islist(data.notes) and #data.notes == 0, "packet: notes should be an empty array")
  total = total + 1
end

-- 20. フェンス長: バッククォート無し→3。3連あり→4(本文の先頭・途中の両方)
do
  local session = make_normalized_session({ notes = { { filePath = "a.txt", body = "plain body" } } })
  local lines = hn.build_packet(session, "hunk diff")
  local _, fence_len = extract_fenced_json(lines)
  assert(fence_len == 3, "packet: fence length should be 3 without backticks, got " .. fence_len)
  total = total + 1
end
do
  -- バッククォート3連が本文の先頭にある場合
  local session = make_normalized_session({ notes = { { filePath = "a.txt", body = "```code``` trailing" } } })
  local lines = hn.build_packet(session, "hunk diff")
  local _, fence_len = extract_fenced_json(lines)
  assert(fence_len == 4, "packet: fence length should be 4 (leading backticks), got " .. fence_len)
  total = total + 1
end
do
  -- バッククォート3連が本文の途中にある場合
  local session = make_normalized_session({ notes = { { filePath = "a.txt", body = "leading ```code``` trailing" } } })
  local lines = hn.build_packet(session, "hunk diff")
  local _, fence_len = extract_fenced_json(lines)
  assert(fence_len == 4, "packet: fence length should be 4 (mid-body backticks), got " .. fence_len)
  total = total + 1
end

-- 21. ノート本文にMarkdown見出し・連続バッククォート・"## if session is gone"・改行を含めても
--     固定セクションの行構造が壊れない(本文はJSON側1行に閉じ込められる)
do
  local plain_body = "note"
  local tricky_body = "# heading\n``` inner ```\n## if session is gone\nmore text"

  local plain_lines = hn.build_packet(make_normalized_session({ notes = { { filePath = "a.txt", body = plain_body } } }), "hunk diff")
  local tricky_lines =
    hn.build_packet(make_normalized_session({ notes = { { filePath = "a.txt", body = tricky_body } } }), "hunk diff")

  assert(#plain_lines == #tricky_lines, "packet: header line count should be unaffected by note body content")

  local data = extract_fenced_json(tricky_lines)
  assert(data.notes[1].body == tricky_body, "packet: tricky note body should round-trip unchanged")
  total = total + 1
end

-- 22. titleに見出し記号・バッククォート・改行が入ってもMarkdown側の構造が壊れない
do
  local plain_lines = hn.build_packet(make_normalized_session({ title = "plain" }), "hunk diff")
  local tricky_title = "# Title\n```weird```\nmore"
  local tricky_lines = hn.build_packet(make_normalized_session({ title = tricky_title }), "hunk diff")

  assert(#plain_lines == #tricky_lines, "packet: header line count should be unaffected by title content")

  local data = extract_fenced_json(tricky_lines)
  assert(data.title == tricky_title, "packet: tricky title should round-trip unchanged")
  total = total + 1
end

-- 23. 特殊文字を含むfile pathでも壊れない
do
  local weird_path = 'a "quoted" \\path\nwith newline and `backtick`.txt'
  local session = make_normalized_session({
    focus = { filePath = weird_path, oldRange = { 1, 2 }, newRange = { 1, 2 }, hunkIndex = 0 },
  })
  local lines = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "packet: build_packet failed for weird path")
  local data = extract_fenced_json(lines)
  assert(data.focus.filePath == weird_path, "packet: filePath should round-trip unchanged")
  total = total + 1
end

-- 24. commandが両方の既知値で正しくJSONに入る。欠落・未知値は失敗
do
  local session = make_normalized_session({})
  for _, cmd in ipairs({ "hunk diff", "hunk diff --staged" }) do
    local lines = hn.build_packet(session, cmd)
    assert(lines ~= nil, "packet: build_packet failed for cmd=" .. cmd)
    local data = extract_fenced_json(lines)
    assert(data.command == cmd, "packet: command mismatch for " .. cmd)
    total = total + 1
  end

  local lines_bad, err_bad = hn.build_packet(session, "hunk diff --unknown")
  assert(lines_bad == nil and err_bad ~= nil, "packet: unknown command should fail")
  total = total + 1

  local lines_nil, err_nil = hn.build_packet(session, nil)
  assert(lines_nil == nil and err_nil ~= nil, "packet: missing command should fail")
  total = total + 1
end

-- 25. focus無し・ノート0件 → packet生成が失敗値を返す
do
  local session = make_normalized_session({ no_focus = true, notes = {} })
  local lines, err = hn.build_packet(session, "hunk diff")
  assert(lines == nil and err ~= nil, "packet: no focus and no notes should fail")
  total = total + 1
end

-- 26. focus無し・ノート有り → 成功
do
  local session =
    make_normalized_session({ no_focus = true, notes = { { filePath = "a.txt", body = "n", newRange = { 1, 2 } } } })
  local lines, err = hn.build_packet(session, "hunk diff")
  assert(lines ~= nil, "packet: no focus but notes present should succeed: " .. tostring(err))
  total = total + 1
end

-- 27. 型検査失敗のerr文字列に原因種別(field欠落/型不正/版上げの可能性)が読み取れる
do
  local _, err = hn.parse_sessions(vim.json.encode({ foo = "bar" }))
  assert(err:find("版上げ") ~= nil, "err should mention version-drift possibility: " .. tostring(err))
  total = total + 1
end
do
  local s = make_valid_session()
  s.snapshot.state.selectedHunkIndex = nil
  local _, err = hn.normalize_session(s)
  assert(err:find("field欠落") ~= nil, "err should mention missing field: " .. tostring(err))
  total = total + 1
end
do
  local sessions, err = hn.parse_sessions(sessions_json({ { pid = "not-a-number" } }))
  assert(sessions == nil, "err should mention wrong type precondition: " .. tostring(err))
  assert(err:find("型不正") ~= nil, "err should mention wrong type: " .. tostring(err))
  total = total + 1
end

-- ==== I/O境界(M.send): vim.system / vim.fn.jobpid / vim.fn.jobwait / vim.notify /
--      claudecode_tmpfile / claudecode_send を差し替えて検証する。各テストの最後で必ず元に戻す。

local IO_PID = 42

local function stub_system_success(json_text)
  return function()
    return {
      wait = function()
        return { code = 0, stdout = json_text, stderr = "" }
      end,
    }
  end
end

local function stub_system_exit_code(code, stderr)
  return function()
    return {
      wait = function()
        return { code = code, stdout = "", stderr = stderr or "" }
      end,
    }
  end
end

local function stub_system_throws(msg)
  return function()
    error(msg or "ENOENT: no such file or directory (cmd): 'hunk'")
  end
end

local function make_send_stub(mode, sent_calls)
  return {
    send = function(path, opts)
      table.insert(sent_calls, { path = path, opts = opts })
      if mode == "false" then
        return false, "send failed (stub)"
      elseif mode == "throw" then
        error("send exploded (stub)")
      end
      return true
    end,
  }
end

-- vim.system/vim.fn.jobpid/vim.fn.jobwait/vim.fn.delete/vim.notify/claudecode_tmpfile/
-- claudecode_send を差し替えてM.send(term, cmd)を1回実行し、必ず元に戻す。
-- opts.check(result)で検証する。opts.send_loader_throwsを立てると、package.loadedを
-- 差し替える代わりにpackage.preloadへ例外を投げるローダーを仕込む(requireの失敗を再現する)。
-- opts.deleteを渡すとvim.fn.deleteだけ差し替える(片付けは常に元のdeleteで行う)。
local function run_io_case(opts)
  local original_system = vim.system
  local original_jobpid = vim.fn.jobpid
  local original_jobwait = vim.fn.jobwait
  local original_delete = vim.fn.delete
  local original_notify = vim.notify
  local original_tmpfile_mod = package.loaded["utils.claudecode_tmpfile"]
  local original_send_mod = package.loaded["utils.claudecode_send"]
  local original_send_preload = package.preload["utils.claudecode_send"]

  local created = {}
  local sent_calls = {}
  local notify_calls = {}

  vim.system = opts.system
  vim.fn.jobpid = function()
    return IO_PID
  end
  vim.fn.jobwait = function()
    return { opts.alive == false and 0 or -1 }
  end
  vim.notify = function(msg, level)
    table.insert(notify_calls, { msg = msg, level = level })
  end
  if opts.delete then
    vim.fn.delete = opts.delete
  end
  package.loaded["utils.claudecode_tmpfile"] = {
    create = function(lines, tmp_opts)
      local path = vim.fn.tempname() .. "." .. (tmp_opts and tmp_opts.ext or "md")
      vim.fn.writefile(lines, path)
      table.insert(created, path)
      return path
    end,
  }
  if opts.send_loader_throws then
    package.loaded["utils.claudecode_send"] = nil
    package.preload["utils.claudecode_send"] = function()
      error("claudecode_send loader boom (stub)")
    end
  else
    package.loaded["utils.claudecode_send"] = make_send_stub(opts.send_mode or "ok", sent_calls)
  end

  local run_ok, ret_ok, ret_err = pcall(hn.send, { job_id = 1 }, opts.cmd or "hunk diff")

  vim.system = original_system
  vim.fn.jobpid = original_jobpid
  vim.fn.jobwait = original_jobwait
  vim.fn.delete = original_delete
  vim.notify = original_notify
  package.loaded["utils.claudecode_tmpfile"] = original_tmpfile_mod
  package.loaded["utils.claudecode_send"] = original_send_mod
  package.preload["utils.claudecode_send"] = original_send_preload

  assert(run_ok, "M.send itself should never throw past its own pcall: " .. tostring(ret_ok))

  local check_ok, check_err = pcall(
    opts.check,
    { ok = ret_ok, err = ret_err, created = created, sent_calls = sent_calls, notify_calls = notify_calls }
  )

  -- 検証後に片付ける(成功ケースはpipelineが削除しないので残る。ここで確実に掃除する。
  -- opts.deleteが壊れていてもここは常に元のdeleteを使う)
  for _, p in ipairs(created) do
    pcall(original_delete, p)
  end

  if not check_ok then
    error(check_err, 0)
  end
end

local function valid_raw_session(overrides)
  local s = make_valid_session()
  s.pid = IO_PID
  for k, v in pairs(overrides or {}) do
    s[k] = v
  end
  return s
end

-- K1回帰 + 追加テスト: claudecode_send.sendが例外を投げる場合、falseを返し、
-- notifyがちょうど1回、作成済み一時ファイルが削除されている
do
  local json_text = sessions_json({ valid_raw_session() })
  run_io_case({
    system = stub_system_success(json_text),
    send_mode = "throw",
    alive = true,
    check = function(r)
      assert(r.ok == false, "send throw: expected ok=false")
      assert(#r.notify_calls == 1, "send throw: expected exactly 1 notify, got " .. #r.notify_calls)
      assert(#r.created == 1, "send throw: expected tmpfile to be created")
      assert(vim.fn.filereadable(r.created[1]) == 0, "send throw: tmpfile should be deleted")
    end,
  })
  total = total + 1
end

-- 追加テスト: claudecode_send.sendがfalseを返す場合も同様
do
  local json_text = sessions_json({ valid_raw_session() })
  run_io_case({
    system = stub_system_success(json_text),
    send_mode = "false",
    alive = true,
    check = function(r)
      assert(r.ok == false, "send false: expected ok=false")
      assert(#r.notify_calls == 1, "send false: expected exactly 1 notify, got " .. #r.notify_calls)
      assert(#r.created == 1, "send false: expected tmpfile to be created")
      assert(vim.fn.filereadable(r.created[1]) == 0, "send false: tmpfile should be deleted")
    end,
  })
  total = total + 1
end

-- K4回帰: 送信成功時、M.sendの戻り値は(true, nil)。notifyは呼ばれない
do
  local json_text = sessions_json({ valid_raw_session() })
  run_io_case({
    system = stub_system_success(json_text),
    send_mode = "ok",
    alive = true,
    check = function(r)
      assert(r.ok == true and r.err == nil, "K4: success should be (true, nil), got ok=" .. tostring(r.ok) .. " err=" .. tostring(r.err))
      assert(#r.notify_calls == 0, "K4: success should not notify")
      assert(#r.sent_calls == 1, "K4: send should be called once")
      assert(#r.created == 1 and vim.fn.filereadable(r.created[1]) == 1, "K4: tmpfile should remain on success")
    end,
  })
  total = total + 1
end

-- 追加テスト: vim.systemがspawn例外を投げる場合、falseを返し、notifyが1回、
-- 一時ファイルは作られない
do
  run_io_case({
    system = stub_system_throws(),
    send_mode = "ok",
    alive = true,
    check = function(r)
      assert(r.ok == false, "spawn exception: expected ok=false")
      assert(#r.notify_calls == 1, "spawn exception: expected exactly 1 notify")
      assert(#r.sent_calls == 0, "spawn exception: send should not be reached")
      assert(#r.created == 0, "spawn exception: tmpfile should not be created")
    end,
  })
  total = total + 1
end

-- 追加テスト: vim.systemが非ゼロcodeを返す場合(timeoutを含む)も同様
do
  run_io_case({
    system = stub_system_exit_code(124, "context deadline exceeded"),
    send_mode = "ok",
    alive = true,
    check = function(r)
      assert(r.ok == false, "non-zero exit: expected ok=false")
      assert(#r.notify_calls == 1, "non-zero exit: expected exactly 1 notify")
      assert(#r.sent_calls == 0, "non-zero exit: send should not be reached")
      assert(#r.created == 0, "non-zero exit: tmpfile should not be created")
    end,
  })
  total = total + 1
end

-- 追加テスト: jobwaitが終了値を返す場合(ジョブが死んでいる)も同様
do
  local json_text = sessions_json({ valid_raw_session() })
  run_io_case({
    system = stub_system_success(json_text),
    send_mode = "ok",
    alive = false,
    check = function(r)
      assert(r.ok == false, "jobwait dead: expected ok=false")
      assert(#r.notify_calls == 1, "jobwait dead: expected exactly 1 notify")
      assert(#r.sent_calls == 0, "jobwait dead: send should not be reached")
      assert(#r.created == 0, "jobwait dead: tmpfile should not be created")
    end,
  })
  total = total + 1
end

-- K3回帰: 改行・タブを含む長いtitleでも、送信labelが1行に収まり所定の長さで切り詰められる
do
  local tricky_title = "line1\nline2\ttab-" .. string.rep("a", 80)
  local json_text = sessions_json({ valid_raw_session({ title = tricky_title }) })
  run_io_case({
    system = stub_system_success(json_text),
    send_mode = "ok",
    alive = true,
    check = function(r)
      assert(r.ok == true, "K3: pipeline should succeed: " .. tostring(r.err))
      assert(#r.sent_calls == 1, "K3: send should be called once")
      local label = r.sent_calls[1].opts.label
      assert(label:find("[\r\n\t]") == nil, "K3: label must not contain raw control characters, got " .. vim.inspect(label))
      assert(
        vim.fn.strcharlen(label) <= 60,
        "K3: label should be truncated to <=60 chars, got " .. vim.fn.strcharlen(label)
      )
    end,
  })
  total = total + 1
end

-- L1回帰: claudecode_sendのローダー自体が例外を投げる場合(requireの失敗)。
-- tmpfile作成前にsend関数を解決しているため、tmpfileはそもそも作られない
do
  run_io_case({
    system = stub_system_success(sessions_json({ valid_raw_session() })),
    send_loader_throws = true,
    alive = true,
    check = function(r)
      assert(r.ok == false, "send loader throws: expected ok=false")
      assert(#r.notify_calls == 1, "send loader throws: expected exactly 1 notify, got " .. #r.notify_calls)
      assert(#r.created == 0, "send loader throws: tmpfile should not have been created")
    end,
  })
  total = total + 1
end

-- L1回帰: vim.fn.deleteが例外を投げても、元の送信エラーがそのまま通知される
-- (deleteの失敗で通知が差し替わらない)
do
  run_io_case({
    system = stub_system_success(sessions_json({ valid_raw_session() })),
    send_mode = "false",
    alive = true,
    delete = function()
      error("delete exploded (stub)")
    end,
    check = function(r)
      assert(r.ok == false, "delete throws: expected ok=false")
      assert(r.err == "send failed (stub)", "delete throws: original send error should be preserved, got " .. tostring(r.err))
      assert(#r.notify_calls == 1, "delete throws: expected exactly 1 notify, got " .. #r.notify_calls)
    end,
  })
  total = total + 1
end

-- L1回帰: vim.fn.deleteが-1(失敗)を返す場合も同様
do
  run_io_case({
    system = stub_system_success(sessions_json({ valid_raw_session() })),
    send_mode = "false",
    alive = true,
    delete = function()
      return -1
    end,
    check = function(r)
      assert(r.ok == false, "delete returns -1: expected ok=false")
      assert(
        r.err == "send failed (stub)",
        "delete returns -1: original send error should be preserved, got " .. tostring(r.err)
      )
      assert(#r.notify_calls == 1, "delete returns -1: expected exactly 1 notify, got " .. #r.notify_calls)
    end,
  })
  total = total + 1
end

print(string.format("hunk_notes_spec: %d/%d passed", total, total))
