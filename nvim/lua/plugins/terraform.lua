-- lang.terraform extra同梱のlint診断(terraform_validate)を無効化。
-- treesitter / terraform-ls / terraform_fmt は有効のまま(lintノイズのみ回避)。
-- 注: tflintはmasonが導入するが、linter未配線のため実行されない(inert)。
return {
  {
    "mfussenegger/nvim-lint",
    optional = true,
    opts = function(_, opts)
      opts.linters_by_ft = opts.linters_by_ft or {}
      opts.linters_by_ft.terraform = {}
      opts.linters_by_ft.tf = {}
    end,
  },
}
