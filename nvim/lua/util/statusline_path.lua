local M = {}

local ELL = "…"
local ew = vim.fn.strdisplaywidth(ELL)
local MIN_INNER_W = 16 -- ファイル名のhead+tailに最低限使う表示幅
local RESERVE = 40 -- 他コンポーネント用に確保する幅(lualine組み込みのshorting_target既定値と同値)

--- name の先頭から表示幅 width に収まる最大文字数を切り出す。skipcc=true で結合文字を分断しない。
---@param name string
---@param width integer
---@return string
local function head_by_width(name, width)
  if width <= 0 then
    return ""
  end
  local total = vim.fn.strchars(name)
  local result = ""
  for n = 1, total do
    local candidate = vim.fn.strcharpart(name, 0, n, true)
    if vim.fn.strdisplaywidth(candidate) > width then
      break
    end
    result = candidate
  end
  return result
end

--- name の末尾から表示幅 width に収まる最大文字数を切り出す。skipcc=true で結合文字を分断しない。
---@param name string
---@param width integer
---@return string
local function tail_by_width(name, width)
  if width <= 0 then
    return ""
  end
  local total = vim.fn.strchars(name)
  local result = ""
  for n = 1, total do
    local candidate = vim.fn.strcharpart(name, total - n, n, true)
    if vim.fn.strdisplaywidth(candidate) > width then
      break
    end
    result = candidate
  end
  return result
end

--- name を budget 幅に収まるよう中間を省略する。
---@param name string
---@param budget integer
---@return string
local function elide(name, budget)
  if vim.fn.strdisplaywidth(name) <= budget then
    return name
  end
  local inner = budget - ew
  if inner <= 0 then
    return head_by_width(name, budget)
  end
  local tail_w = math.floor(inner / 2)
  local head_w = inner - tail_w
  return head_by_width(name, head_w) .. ELL .. tail_by_width(name, tail_w)
end

--- full を avail 幅に収まるよう短縮する。sym の幅は呼び出し側で avail から差し引き済みとする。
---@param full string
---@param avail integer
---@return string
function M.shorten(full, avail)
  if avail <= 0 then
    return ""
  end
  if vim.fn.strdisplaywidth(full) <= avail then
    return full
  end

  local slash = full:find("/[^/]*$")
  local dir, name
  if slash then
    dir = full:sub(1, slash)
    name = full:sub(slash + 1)
  else
    dir = ""
    name = full
  end

  local min_w = math.min(vim.fn.strdisplaywidth(name), MIN_INNER_W + ew)

  local budget = avail - vim.fn.strdisplaywidth(dir)
  if budget >= min_w then
    return dir .. elide(name, budget)
  end

  -- 先頭セグメントはディレクトリを1階層ずつ落とすループの間だけ保持する。
  -- 全部落としても収まらない場合はディレクトリ表示自体を諦め、幅をファイル名へ回す。
  if slash then
    local first_slash = full:find("/", 1, true)
    local seg1 = full:sub(1, first_slash - 1)
    local middle = full:sub(first_slash + 1, slash - 1)
    local segs = {}
    if middle ~= "" then
      for seg in (middle .. "/"):gmatch("([^/]*)/") do
        table.insert(segs, seg)
      end
    end
    for drop = 1, #segs do
      local remaining = {}
      for i = drop + 1, #segs do
        table.insert(remaining, segs[i])
      end
      local dir2
      if #remaining > 0 then
        dir2 = seg1 .. "/" .. ELL .. "/" .. table.concat(remaining, "/") .. "/"
      else
        dir2 = seg1 .. "/" .. ELL .. "/"
      end
      budget = avail - vim.fn.strdisplaywidth(dir2)
      if budget >= min_w then
        return dir2 .. elide(name, budget)
      end
    end
  end

  return elide(name, avail)
end

--- lualine の filename コンポーネント代替。現在バッファのパスを幅に収まるよう短縮して返す。
---@return string
function M.current()
  local full = vim.fn.expand("%:p:~")
  if full == "" then
    return "[No Name]"
  end

  local sym = ""
  if vim.bo.modified then
    sym = sym .. "[+]"
  end
  if vim.bo.modifiable == false or vim.bo.readonly then
    sym = sym .. "[-]"
  end
  local suffix = (sym ~= "") and (" " .. sym) or ""

  -- lualineのglobalstatus既定はlaststatus==3判定(lualine.nvim/lua/lualine/config.lua:22)で、
  -- config/options.luaがlaststatus=3を明示しているため一致する。globalstatus明示指定時は乖離する。
  local win_w = (vim.go.laststatus == 3) and vim.o.columns or vim.fn.winwidth(0)
  local avail = win_w - RESERVE - vim.fn.strdisplaywidth(suffix)

  local path = M.shorten(full, avail)
  if path == "" then
    return sym
  end
  return path .. suffix
end

return M
