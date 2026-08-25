local classifier = require("util.flatten_classify")
local payload = classifier.NIX_CMD_PAYLOAD
local one_char_diff = payload:sub(1, -2) .. "1"

-- 2026-08-25実機採取(toggleterm内`nvim`引数なし起動)を固定したsnapshot。
-- 検出できるのはNIX_CMD_PAYLOADをこのsnapshotの更新なしに変更した場合のみ。
-- 実際のNixラッパー更新によるprovider列挙のdriftは、実wrapped nvimからargvを
-- 採取する統合検査でしか検知できない(この行は対象外)。
local measured_payload =
  "lua vim.g.loaded_node_provider=0;vim.g.loaded_perl_provider=0;"
  .. "vim.g.loaded_ruby_provider=0;vim.g.loaded_python3_provider=0"
assert(
  measured_payload == classifier.NIX_CMD_PAYLOAD,
  "measured payload no longer matches NIX_CMD_PAYLOAD constant"
)
local provider_order_diff =
  "lua vim.g.loaded_perl_provider=0;vim.g.loaded_node_provider=0;"
  .. "vim.g.loaded_ruby_provider=0;vim.g.loaded_python3_provider=0"

local cases = {
  {
    name = "cmd before embed delegates",
    argv = { vim.v.progpath, "--cmd", payload, "--embed", "file.txt" },
    expected = false,
  },
  {
    name = "fixed Nix payload delegates",
    argv = { vim.v.progpath, "--cmd", payload, "file.txt" },
    expected = false,
  },
  {
    name = "bad cmd nests",
    argv = { vim.v.progpath, "--cmd", "qa!", "file.txt" },
    expected = true,
  },
  {
    name = "cmd without payload nests",
    argv = { vim.v.progpath, "--cmd" },
    expected = true,
  },
  {
    name = "fixed then bad cmd nests",
    argv = { vim.v.progpath, "--cmd", payload, "--cmd", "qa!", "file.txt" },
    expected = true,
  },
  {
    name = "multiple fixed payloads delegate",
    argv = { vim.v.progpath, "--cmd", payload, "--cmd", payload, "file.txt" },
    expected = false,
  },
  {
    name = "one character payload difference nests",
    argv = { vim.v.progpath, "--cmd", one_char_diff, "file.txt" },
    expected = true,
  },
  {
    name = "payload whitespace difference nests",
    argv = { vim.v.progpath, "--cmd", payload .. " ", "file.txt" },
    expected = true,
  },
  {
    name = "provider order difference nests",
    argv = { vim.v.progpath, "--cmd", provider_order_diff, "file.txt" },
    expected = true,
  },
  {
    name = "embed without file nests",
    argv = { vim.v.progpath, "--embed" },
    expected = true,
  },
  {
    name = "embed with file delegates",
    argv = { vim.v.progpath, "--embed", "file.txt" },
    expected = false,
  },
  {
    name = "embed with fixed cmd and file delegates",
    argv = { vim.v.progpath, "--embed", "--cmd", payload, "file.txt" },
    expected = false,
  },
  {
    name = "embed with fixed cmd and no file nests",
    argv = { vim.v.progpath, "--embed", "--cmd", payload },
    expected = true,
  },
  {
    name = "embed with fixed cmd and line command nests",
    argv = { vim.v.progpath, "--embed", "--cmd", payload, "+42", "file.txt" },
    expected = true,
  },
  {
    name = "line command nests",
    argv = { vim.v.progpath, "+42", "file.txt" },
    expected = true,
  },
  {
    name = "plus command nests",
    argv = { vim.v.progpath, "+qa", "file.txt" },
    expected = true,
  },
  {
    name = "read-only flag nests",
    argv = { vim.v.progpath, "-R", "file.txt" },
    expected = true,
  },
  {
    name = "stdin marker nests",
    argv = { vim.v.progpath, "-" },
    expected = true,
  },
  {
    name = "option separator nests",
    argv = { vim.v.progpath, "--" },
    expected = true,
  },
  {
    name = "no arguments nests",
    argv = { vim.v.progpath },
    expected = true,
  },
  {
    name = "plain file delegates",
    argv = { vim.v.progpath, "file.txt" },
    expected = false,
  },
  {
    name = "multiple files delegate",
    argv = { vim.v.progpath, "first.txt", "second.txt" },
    expected = false,
  },
}

---@param case { name: string, argv: string[], expected: boolean }
local function assert_case(case)
  local actual = classifier.classify(case.argv)
  assert(
    actual == case.expected,
    string.format("%s: expected %s, got %s", case.name, case.expected, actual)
  )
end

for _, case in ipairs(cases) do
  assert_case(case)
end

print(string.format("flatten_classify_spec: %d/%d passed", #cases, #cases))
