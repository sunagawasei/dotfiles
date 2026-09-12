local M = {}

--- gopls は DocumentSymbol の name を `(*Receiver).Method` 形式で返すため、telescope の
--- symbol 列(末尾から切り詰める固定幅)ではレシーバだけで埋まりメソッド名が消える。
--- 種別列に method と出る時点でレシーバの存在は分かり、型はプレビューで読めるので落とす。
---
--- 受け取るのは `vim.lsp.util.symbols_to_items` が作る `"[Kind] name"` 形式の文字列。
--- kind が Method で name が `(recv).name` の形のときだけレシーバを外し、他はそのまま返す。
---@param text string?
---@return string?
function M.strip_receiver(text)
  if type(text) ~= "string" then
    return text
  end
  -- レシーバ型に `)` は現れないので、最短一致で最初の閉じ括弧まで取れる
  local kind, name = text:match("^(%[Method%]%s+)%(.-%)%.(.+)$")
  if not kind then
    return text
  end
  return kind .. name
end

return M
