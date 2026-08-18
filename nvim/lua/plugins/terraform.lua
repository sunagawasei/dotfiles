-- lang.terraform extra同梱のlint診断(terraform_validate)を無効化。
-- treesitter / terraform-ls / terraform_fmt は有効のまま(lintノイズのみ回避)。
-- 注: tflintはmasonが導入するが、linter未配線のため実行されない(inert)。
return {
  {
    -- terraform-lsはheredoc内のテンプレート補間で負のdeltaStartを返し、
    -- nvimのsemantic tokens行跨ぎループが約43億回空回りしてUIが固まる。
    "neovim/nvim-lspconfig",
    optional = true,
    opts = {
      servers = {
        terraformls = {
          -- Wasabi backendのrepoはbackend.tfにprofile未指定で、AWSのdefault profile
          -- (PERMAN credential_process)を引きSlackに認証通知を飛ばすため塞ぐ。
          cmd_env = { AWS_PROFILE = "wasabi-tfstate" },
        },
      },
    },
    init = function()
      vim.api.nvim_create_autocmd("LspAttach", {
        callback = function(args)
          local client = vim.lsp.get_client_by_id(args.data.client_id)
          if client and client.name == "terraformls" then
            client.server_capabilities.semanticTokensProvider = nil
          end
        end,
      })
    end,
  },
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
