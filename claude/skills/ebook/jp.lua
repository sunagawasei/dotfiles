-- note.com HTML -> 日本語書籍向けAST整形
local function is_cjk(c)
  if not c then return false end
  local cp = utf8.codepoint(c)
  return (cp >= 0x3000 and cp <= 0x30FF) or (cp >= 0x3400 and cp <= 0x9FFF)
      or (cp >= 0xFF00 and cp <= 0xFFEF) or (cp >= 0x2010 and cp <= 0x203A)
end
local function last_char(s) return s:match("[%z\1-\127\194-\244][\128-\191]*$") end
local function first_char(s) return s:match("^[%z\1-\127\194-\244][\128-\191]*") end

local function edge(inl, last)
  if not inl then return nil end
  if inl.t == "Str" then return last and last_char(inl.text) or first_char(inl.text) end
  if inl.content and #inl.content > 0 then
    return edge(last and inl.content[#inl.content] or inl.content[1], last)
  end
  return nil
end

-- 和文どうしの改行位置に入る空白を除く
function Inlines(inls)
  local out = pandoc.List()
  for i, el in ipairs(inls) do
    if el.t == "SoftBreak" then
      local p, n = edge(inls[i-1], true), edge(inls[i+1], false)
      if not (is_cjk(p) or is_cjk(n)) then out:insert(pandoc.Space()) end
    else
      out:insert(el)
    end
  end
  return out
end

local function strip_edges(inls)
  while #inls > 0 and (inls[1].t == "Space" or inls[1].t == "LineBreak" or inls[1].t == "SoftBreak") do inls:remove(1) end
  while #inls > 0 and (inls[#inls].t == "Space" or inls[#inls].t == "LineBreak" or inls[#inls].t == "SoftBreak") do inls:remove(#inls) end
  if #inls > 0 and inls[1].t == "Str" then
    local t = inls[1].text
    while true do
      local n = t:gsub("^%s+", "")
      if t:sub(1, 3) == "\u{3000}" then n = t:sub(4) end
      if n == t then break end
      t = n
    end
    if t == "" then inls:remove(1) else inls[1].text = t end
  end
end

local openers = { ["「"]=1, ["『"]=1, ["（"]=1, ["【"]=1, ["〈"]=1, ["《"]=1, ["〔"]=1, ["“"]=1, ["‘"]=1, ["〝"]=1, ["［"]=1, ["｛"]=1, ["〖"]=1, ["〘"]=1, ["〚"]=1 }

function Para(el)
  strip_edges(el.content)
  if #el.content == 0 then return {} end
  local c = edge(el.content[1], false)
  if c and openers[c] then
    return pandoc.Div({ el }, pandoc.Attr("", { "noindent" }))
  end
  return el
end

local function find_first_link(block)
  local found
  pandoc.walk_block(block, { Link = function(l) if not found then found = l end end })
  return found
end

local function card_title(link)
  local title
  pandoc.walk_inline(pandoc.Span(link.content), { Strong = function(s) if not title then title = s.content end end })
  return title or link.content
end

function Figure(el)
  local svc = el.attributes["embedded-service"]
  if svc == "note" then return {} end
  if svc == "external-article" then
    -- 外部リンクカードは見出し(太字タイトル)のリンクだけ残す
    local link = find_first_link(el)
    if not link then return {} end
    return pandoc.Div({ pandoc.Para({ pandoc.Link(card_title(link), link.target) }) },
                      pandoc.Attr("", { "noindent", "card" }))
  end
  return nil
end
