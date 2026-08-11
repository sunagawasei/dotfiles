vim.wait(8000, function() return vim.g.did_very_lazy == true end, 100)
if not vim.g.did_very_lazy then
  pcall(vim.api.nvim_exec_autocmds, "User", { pattern = "VeryLazy", modeline = false })
  vim.wait(1500, function() return false end, 100)
end
local hex = function(s) return (s:gsub(".", function(c) return string.format("%02x", c:byte()) end)) end
local norm = function(m) return (m=="x" or m=="s" or m=="v") and "v" or m end
-- spec由来のIDを集める
local Config = require("lazy.core.config")
local specid = {}
for _, p in pairs(Config.plugins) do
  local h = p._ and p._.handlers and p._.handlers.keys
  if h then for _, k in pairs(h) do
    if k.lhs and k.lhs ~= "" then
      specid[norm(k.mode or "n") .. ":" .. hex(vim.api.nvim_replace_termcodes(k.lhs, true, true, true))] = true
    end
  end end
end
-- 実マッピング全部を走査し、spec由来でないものを出す
local out = {}
for _, m in ipairs({ "n","v","x","s","o","i","t","c" }) do
  for _, mp in ipairs(vim.api.nvim_get_keymap(m)) do
    local lhs = mp.lhs
    if lhs and lhs ~= "" and lhs:sub(1,6) ~= "<Plug>" and not lhs:find("<SNR>",1,true) then
      local raw = mp.lhsraw
      if not raw or raw == "" then raw = vim.api.nvim_replace_termcodes(lhs, true, true, true) end
      local id = norm(m) .. ":" .. hex(raw)
      if not specid[id] then
        local src = ""
        if mp.callback then
          local ok,i = pcall(debug.getinfo, mp.callback, "S")
          if ok and i then src = (i.source or ""):gsub("^@","") end
        end
        out[#out+1] = { id = id, mode = m, lhs = lhs, desc = mp.desc or "", src = src, rhs = (mp.rhs or ""):sub(1,40) }
      end
    end
  end
end
local f = io.open(vim.env.DUMP_OUT,"w"); f:write(vim.json.encode(out)); f:close()
