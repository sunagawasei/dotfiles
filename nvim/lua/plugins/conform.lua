-- .github 配下は GitHub Actions の慣習（dash をキー配下に字下げ）、それ以外は
-- Kubernetes の正規形（dash を親キーと同じ column = indentless sequence）に揃える。
local indented_yaml_patterns = {
  "/%.github/",
}

local function yamlfmt_args(_, ctx)
  local indentless = true
  for _, pattern in ipairs(indented_yaml_patterns) do
    if ctx.filename:match(pattern) then
      indentless = false
      break
    end
  end
  return {
    "-formatter",
    "indentless_arrays=" .. tostring(indentless),
    -- 未指定だと空行が畳まれ、保存ごとに差分が出る
    "-formatter",
    "retain_line_breaks=true",
    "-in",
  }
end

return {
  "stevearc/conform.nvim",
  opts = {
    formatters = {
      yamlfmt = {
        args = yamlfmt_args,
        -- 同梱定義は stdin を立てないため、yamlfmt に入力が渡らず無言で何もしない
        stdin = true,
        -- ここで指定していないキーを repo-local 設定に委ねる。yamlfmt 自身が認識する
        -- 5つの名前すべてを挙げ、monorepo のサブディレクトリ設定も拾えるようにする。
        cwd = function(self, ctx)
          return require("conform.util").root_file({
            ".yamlfmt",
            ".yamlfmt.yml",
            ".yamlfmt.yaml",
            "yamlfmt.yml",
            "yamlfmt.yaml",
            ".git",
          })(self, ctx)
        end,
      },
    },
    formatters_by_ft = {
      lua = { "stylua" },
      python = { "black", "isort" },
      javascript = { "oxfmt" },
      typescript = { "oxfmt" },
      javascriptreact = { "oxfmt" },
      typescriptreact = { "oxfmt" },
      json = { "oxfmt" },
      jsonc = { "oxfmt" },
      yaml = { "yamlfmt" },
      markdown = { "oxfmt" },
      html = { "oxfmt" },
      css = { "oxfmt" },
      scss = { "oxfmt" },
      vue = { "oxfmt" },
      rust = { "rustfmt" },
      go = { "goimports" },
      sh = { "shfmt" },
      fish = { "fish_indent" },
      nix = { "nixfmt" },
      ["_"] = { "trim_whitespace" },
    },
  },
}
