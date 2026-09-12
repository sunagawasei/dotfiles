local M = require("util.lsp_symbol_name")

local cases = {
  -- P1: ポインタレシーバ
  { name = "pointer receiver", input = "[Method] (*T).M", expected = "[Method] M" },
  -- P2: 値レシーバ
  { name = "value receiver", input = "[Method] (T).M", expected = "[Method] M" },
  -- P3: ジェネリクスレシーバ。型引数の `]` で止まらない
  { name = "generic receiver", input = "[Method] (*Cache[K, V]).Get", expected = "[Method] Get" },
  -- P4: 実際に問題になった長い名前
  {
    name = "long receiver",
    input = "[Method] (*ContainerRegistryServer).robotSecret",
    expected = "[Method] robotSecret",
  },
  -- P5: レシーバの無い関数
  { name = "function", input = "[Function] main", expected = "[Function] main" },
  -- P6: kind が Method でもレシーバ形でない
  { name = "method without receiver", input = "[Method] plainName", expected = "[Method] plainName" },
  -- P7: レシーバ形でも kind が Method でない
  { name = "non-method kind", input = "[Field] (x).y", expected = "[Field] (x).y" },
  -- P8: 変換後の文字列を再入力しても変わらない
  { name = "idempotent", input = "[Method] M", expected = "[Method] M" },
  -- P9: 空文字列
  { name = "empty", input = "", expected = "" },
  -- P10: nil はそのまま返す
  { name = "nil", input = nil, expected = nil },
}

for _, case in ipairs(cases) do
  local actual = M.strip_receiver(case.input)
  assert(
    actual == case.expected,
    string.format("strip_receiver(%s): expected %s, got %s", case.name, tostring(case.expected), tostring(actual))
  )
end

print(string.format("lsp_symbol_name_spec: %d/%d passed", #cases, #cases))
