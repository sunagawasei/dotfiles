return {
  {
    "mason-org/mason.nvim",
    opts = {
      ensure_installed = {
        -- Go
        "gopls",
        -- 既存のLSPサーバーも含める
        "lua-language-server",
        "stylua",
        "typescript-language-server",
        "prettier",
        "black",
        "isort",
        "shfmt",
      },
    },
  },
}