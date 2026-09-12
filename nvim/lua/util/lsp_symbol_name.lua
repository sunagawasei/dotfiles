local M = {}

--- gopls は DocumentSymbol の name を `(*Receiver).Method` 形式で返すため、telescope の
--- symbol 列(末尾から切り詰める固定幅)ではレシーバだけで埋まりメソッド名が消える。
--- メソッド名を先頭へ回して優先的に残す。
---
--- 受け取るのは `vim.lsp.util.symbols_to_items` が作る `"[Kind] name"` 形式の文字列。
--- kind が Method で name が `(recv).name` の形のときだけ並べ替え、他はそのまま返す。
---@param text string?
---@return string?
function M.method_first(text)
  if type(text) ~= "string" then
    return text
  end
  -- レシーバ型に `)` は現れないので、最短一致で最初の閉じ括弧まで取れる
  local kind, recv, name = text:match("^(%[Method%]%s+)%((.-)%)%.(.+)$")
  if not kind then
    return text
  end
  return string.format("%s%s (%s)", kind, name, recv)
end

return M
