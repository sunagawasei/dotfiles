return {
  {
    "mason-org/mason.nvim",
    opts = {
      ensure_installed = {
        -- Go
        "gopls",
        -- Markdown (<leader>mo アウトラインピッカーに必要)
        "marksman",
        -- 既存のLSPサーバーも含める
        "lua-language-server",
        "stylua",
        "typescript-language-server",
        "vue-language-server",
        "black",
        "isort",
        "shfmt",
      },
    },
  },
}
