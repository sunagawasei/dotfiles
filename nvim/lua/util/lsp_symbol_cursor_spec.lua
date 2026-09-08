local M = require("util.lsp_symbol_cursor")

local row = 10 -- 全ケース共通の基準行(0-based)

local contains_cases = {
  -- P1: zero-length range
  { name = "zero-length range", r = { start = { line = row, character = 5 }, ["end"] = { line = row, character = 5 } }, expected = false },
  -- P2: 同一行の空range・character 0
  { name = "empty range at character 0", r = { start = { line = row, character = 0 }, ["end"] = { line = row, character = 0 } }, expected = false },
  -- P3: 同一行の非空range
  { name = "non-empty range on same line", r = { start = { line = row, character = 3 }, ["end"] = { line = row, character = 8 } }, expected = true },
  -- P4: 多行rangeでendが(row,0)
  { name = "multi-line range ending at character 0", r = { start = { line = row - 2, character = 0 }, ["end"] = { line = row, character = 0 } }, expected = false },
  -- P5: start.line == row < end.line
  { name = "start on row, end after", r = { start = { line = row, character = 3 }, ["end"] = { line = row + 2, character = 0 } }, expected = true },
  -- P6: rowがrangeより前
  { name = "row before range", r = { start = { line = row + 1, character = 0 }, ["end"] = { line = row + 3, character = 0 } }, expected = false },
  -- P7: rowがrangeより後
  { name = "row after range", r = { start = { line = row - 3, character = 0 }, ["end"] = { line = row - 1, character = 0 } }, expected = false },
  -- P8: rがnil / startが欠落
  { name = "nil range", r = nil, expected = false },
  { name = "range missing start", r = { ["end"] = { line = row, character = 5 } }, expected = false },
  -- P14: 多行rangeの中間行
  { name = "row strictly between start and end lines", r = { start = { line = row - 2, character = 0 }, ["end"] = { line = row + 2, character = 0 } }, expected = true },
  -- P15: startが前の行、endが(row, 1以上)
  { name = "start on previous line, end after character 0", r = { start = { line = row - 1, character = 0 }, ["end"] = { line = row, character = 1 } }, expected = true },
}

---@param case { name: string, r: table?, expected: boolean }
local function assert_contains(case)
  local actual = M.contains(row, case.r)
  assert(
    actual == case.expected,
    string.format("contains(%s): expected %s, got %s", case.name, case.expected, actual)
  )
end

for _, case in ipairs(contains_cases) do
  assert_contains(case)
end

-- select_index: 各ケースは items(1-based配列)と row を渡し、返る index を検証する

-- P9: 入れ子(親がrowを含み、子は含まない) → 親のindex
local parent_only = {
  { pos = { 1, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 5, character = 0 } } },
  { pos = { 2, 1 }, range = { start = { line = row + 10, character = 0 }, ["end"] = { line = row + 12, character = 0 } } },
}

-- P10: 入れ子(親も子もrowを含む) → 子(最狭)のindex
local parent_and_child = {
  { pos = { 1, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 5, character = 0 } } },
  { pos = { 2, 1 }, range = { start = { line = row, character = 2 }, ["end"] = { line = row, character = 8 } } },
}

-- P11: 包含itemなし、fallback候補が複数 → 最後のindex
local fallback_only = {
  { pos = { 1, 1 } },
  { pos = { 2, 1 } },
  { pos = { 3, 1 } },
}

-- P12: 包含itemもfallbackもなし → nil
local none_match = {
  { pos = { row + 5, 1 } },
}

-- P13: 完全同一rangeのitemが2つ → 後のindex
local identical_ranges = {
  { pos = { 1, 1 }, range = { start = { line = row, character = 2 }, ["end"] = { line = row, character = 8 } } },
  { pos = { 2, 1 }, range = { start = { line = row, character = 2 }, ["end"] = { line = row, character = 8 } } },
}

-- P16: dlが同じでdcだけ異なる2つ → dcが小さい方のindex(1)
local same_dl_diff_dc = {
  { pos = { 1, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 1, character = 2 } } },
  { pos = { 2, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 1, character = 5 } } },
}

-- P17: dl/dcが同じでstart.lineだけ異なる2つ → start.lineが後の方のindex(1)
local same_dl_dc_diff_start_line = {
  { pos = { 1, 1 }, range = { start = { line = row, character = 0 }, ["end"] = { line = row + 2, character = 0 } } },
  { pos = { 2, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 1, character = 0 } } },
}

-- P18: dl/dcが同じでstart.characterだけ異なる2つ → start.characterが後の方のindex(1)
local same_dl_dc_diff_start_char = {
  { pos = { 1, 1 }, range = { start = { line = row - 1, character = 6 }, ["end"] = { line = row + 1, character = 9 } } },
  { pos = { 2, 1 }, range = { start = { line = row - 1, character = 2 }, ["end"] = { line = row + 1, character = 5 } } },
}

-- P19: 包含候補があり、それより後ろにfallback候補がある → 包含候補のindex(1)
local contains_before_fallback = {
  { pos = { row, 1 }, range = { start = { line = row - 1, character = 0 }, ["end"] = { line = row + 1, character = 0 } } },
  { pos = { row - 2, 1 }, range = { start = { line = row + 5, character = 0 }, ["end"] = { line = row + 7, character = 0 } } },
}

local select_cases = {
  { name = "parent contains, child does not", items = parent_only, expected = 1 },
  { name = "parent and child both contain, child wins", items = parent_and_child, expected = 2 },
  { name = "no containing item, multiple fallbacks, last wins", items = fallback_only, expected = 3 },
  { name = "no containing item, no fallback", items = none_match, expected = nil },
  { name = "identical ranges, later wins", items = identical_ranges, expected = 2 },
  { name = "same dl, differing dc, smaller dc wins", items = same_dl_diff_dc, expected = 1 },
  { name = "same dl/dc, differing start.line, later start.line wins", items = same_dl_dc_diff_start_line, expected = 1 },
  { name = "same dl/dc, differing start.character, later start.character wins", items = same_dl_dc_diff_start_char, expected = 1 },
  { name = "containing candidate before trailing fallback", items = contains_before_fallback, expected = 1 },
}

---@param case { name: string, items: table[], expected: integer? }
local function assert_select(case)
  local actual = M.select_index(case.items, row)
  assert(
    actual == case.expected,
    string.format("select_index(%s): expected %s, got %s", case.name, tostring(case.expected), tostring(actual))
  )
end

for _, case in ipairs(select_cases) do
  assert_select(case)
end

local total = #contains_cases + #select_cases
print(string.format("lsp_symbol_cursor_spec: %d/%d passed", total, total))
