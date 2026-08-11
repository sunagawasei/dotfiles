vim.wait(8000, function() return vim.g.did_very_lazy == true end, 100)
if not vim.g.did_very_lazy then
  pcall(vim.api.nvim_exec_autocmds, "User", { pattern = "VeryLazy", modeline = false })
  vim.wait(1500, function() return false end, 100)
end
local hex = function(s) return (s:gsub(".", function(c) return string.format("%02x", c:byte()) end)) end
local norm = function(m) return (m == "x" or m == "s" or m == "v") and "v" or m end

-- 1) spec由来のowner表（解決済み）
local Config = require("lazy.core.config")
local owners = {}
for name, p in pairs(Config.plugins) do
  local h = p._ and p._.handlers and p._.handlers.keys
  if h then
    for _, k in pairs(h) do
      local lhs = k.lhs or ""
      if lhs ~= "" then
        local key = norm(k.mode or "n") .. ":" .. hex(vim.api.nvim_replace_termcodes(lhs, true, true, true))
        owners[key] = owners[key] or { lhs = lhs, entries = {} }
        table.insert(owners[key].entries, {
          plugin = name, mode = k.mode or "n", desc = k.desc or "",
          ft = type(k.ft) == "table" and table.concat(k.ft, ",") or (k.ft or ""),
        })
      end
    end
  end
end

-- 2) トラッカーのカウントをlhsrawへ変換
local tracker = vim.json.decode(io.open(vim.env.TRACKER, "r"):read("*a"))
local usage = {}
for k, v in pairs(tracker.keys) do
  local mode, disp = k:match("^(.-):(.*)$")
  local ok, raw = pcall(vim.api.nvim_replace_termcodes, disp, true, true, true)
  if ok then
    local key = norm(mode) .. ":" .. hex(raw)
    usage[key] = (usage[key] or 0) + v.count
  end
end

local out = { owners = {}, usage = usage }
for key, o in pairs(owners) do out.owners[key] = o end
local f = io.open(vim.env.DUMP_OUT, "w"); f:write(vim.json.encode(out)); f:close()
