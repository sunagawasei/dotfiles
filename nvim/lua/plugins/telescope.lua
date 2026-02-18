return {
  "nvim-telescope/telescope.nvim",
  opts = function(_, opts)
    -- LazyVim形式でoptsをマージ
    opts.defaults = vim.tbl_deep_extend("force", opts.defaults or {}, {
      vimgrep_arguments = {
        "rg",
        "--color=never",
        "--no-heading",
        "--with-filename",
        "--line-number",
        "--column",
        "--smart-case",
        "--no-ignore", -- .gitignore, .ignore, .rgignoreを無視
        "--no-ignore-vcs", -- .git/info/excludeなどのVCS除外ルールを無視
        "--hidden", -- 隠しファイルも検索
        "--glob", -- .gitディレクトリは除外
        "!**/.git/*",
      },
    })
    return opts
  end,
}
