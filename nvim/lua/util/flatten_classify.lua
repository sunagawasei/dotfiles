local NIX_CMD_PAYLOAD =
  "lua vim.g.loaded_node_provider=0;vim.g.loaded_perl_provider=0;"
  .. "vim.g.loaded_ruby_provider=0;vim.g.loaded_python3_provider=0"

local M = {}
M.NIX_CMD_PAYLOAD = NIX_CMD_PAYLOAD

--- vim.v.argvと同じ形式のargv配列から、ローカルnestすべきかを判定する純粋関数。
--- vim.v.argvはread-onlyでテストから差し替えられないため、判定ロジックを
--- ここに分離しargvを直接引数で渡せるようにしている。
---@param argv string[]
---@return boolean should_nest true=ローカルnest, false=host委譲
function M.classify(argv)
  local has_file = false
  local skip_next = false
  for i = 2, #argv do
    local a = argv[i]
    if skip_next then
      skip_next = false
    elseif a == "--cmd" then
      -- payload内容がNixラッパーの固定注入文と厳密一致する場合のみスキップ。
      -- 出現順には依存しない。不一致(ユーザー/外部ツール指定の--cmd)は即nest
      -- 対象にし、host側でのpre_cmds実行による親neovim終了を防ぐ。
      if argv[i + 1] == NIX_CMD_PAYLOAD then
        skip_next = true
      else
        return true
      end
    elseif a == "--embed" then
      -- skip: 対話的TUI起動時、Neovim 0.12はTUI(親)+embed core(子)の2プロセス
      -- 構成になり、coreプロセスのargvには常に--embedが付く。ユーザーが明示的に
      -- `nvim --embed file`を打つケースも同様に委譲対象になる既知の制限として
      -- 許容する(argvだけからは出自を判別できないため)。
    elseif a == "--" then
      return true
    elseif a:sub(1, 1) == "-" then
      return true
    elseif a:sub(1, 1) == "+" then
      return true -- +<数字>を含む全ての+開始トークンをnest対象にする
    else
      has_file = true
    end
  end
  return not has_file
end

return M
