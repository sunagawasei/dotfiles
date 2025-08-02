return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      -- LazyVimの診断設定
      diagnostics = {
        underline = false, -- 赤い波線を完全に無効化
        virtual_text = false, -- インライン診断テキストも無効化
        signs = true, -- 左側のサインカラムは表示
        float = {
          border = "rounded",
          source = "always",
        },
        severity_sort = true,
        update_in_insert = false,
      },
      -- LSPサーバーの設定
      servers = {
        -- Go言語用のLSPサーバー
        gopls = {
          settings = {
            gopls = {
              gofumpt = true,
              codelenses = {
                gc_details = false,
                generate = true,
                regenerate_cgo = true,
                run_govulncheck = true,
                test = true,
                tidy = true,
                upgrade_dependency = true,
                vendor = true,
              },
              hints = {
                assignVariableTypes = true,
                compositeLiteralFields = true,
                compositeLiteralTypes = true,
                constantValues = true,
                functionTypeParameters = true,
                parameterNames = true,
                rangeVariableTypes = true,
              },
              analyses = {
                fieldalignment = true,
                nilness = true,
                unusedparams = true,
                unusedwrite = true,
                useany = true,
              },
              usePlaceholders = true,
              completeUnimported = true,
              staticcheck = true,
              directoryFilters = { "-.git", "-.vscode", "-.idea", "-.vscode-test", "-node_modules" },
              semanticTokens = true,
            },
          },
        },
      },
    },
  },
}