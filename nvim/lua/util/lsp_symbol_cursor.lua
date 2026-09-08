local M = {}

--- LSP range は半開区間。行 row が含まれる ⇔ ∃c: start <= (row, c) < end
---@param row integer 0-based
---@param r table? LSP Range(0-based)
---@return boolean
function M.contains(row, r)
  if not (r and r.start and r["end"]) then
    return false
  end
  if row < r.start.line or row > r["end"].line then
    return false
  end
  if row < r["end"].line then
    return true
  end
  -- row == end.line。この行で許される最小の column は start と同じ行かで変わる
  local min_c = row == r.start.line and r.start.character or 0
  return r["end"].character > min_c
end

--- row に対応する item の index を返す。無ければ nil
---@param items table[] pos 昇順の symbol item(`range` は LSP 生データ、`pos` は 1-based)
---@param row integer 0-based
---@return integer?
function M.select_index(items, row)
  local best, best_dl, best_dc, best_r
  local fallback
  for i = 1, #items do
    local item = items[i]
    local r = item and item.range
    if M.contains(row, r) then
      local dl = r["end"].line - r.start.line
      local dc = r["end"].character - r.start.character
      local better
      if not best then
        better = true
      elseif dl ~= best_dl then
        better = dl < best_dl
      elseif dc ~= best_dc then
        better = dc < best_dc
      elseif r.start.line ~= best_r.start.line then
        better = r.start.line > best_r.start.line
      elseif r.start.character ~= best_r.start.character then
        better = r.start.character > best_r.start.character
      else
        better = true
      end
      if better then
        best, best_dl, best_dc, best_r = i, dl, dc, r
      end
    elseif item and item.pos and item.pos[1] - 1 <= row then
      fallback = i
    end
  end
  return best or fallback
end

return M
