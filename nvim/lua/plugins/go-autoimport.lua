-- Go専用の自動インポート設定
return {
  {
    "neovim/nvim-lspconfig",
    opts = function(_, opts)
      -- Goファイルでの自動インポート用autocmdを設定
      vim.api.nvim_create_autocmd("BufWritePre", {
        pattern = "*.go",
        callback = function(args)
          -- LSP経由で自動インポート整理
          local client = vim.lsp.get_active_clients({ bufnr = args.buf, name = "gopls" })[1]
          if client then
            local params = vim.lsp.util.make_range_params()
            params.context = { only = { "source.organizeImports" } }
            
            -- 同期的にcode_actionを実行
            local result = vim.lsp.buf_request_sync(args.buf, "textDocument/codeAction", params, 3000)
            if result and result[client.id] then
              local actions = result[client.id].result
              if actions then
                for _, action in ipairs(actions) do
                  if action.edit then
                    vim.lsp.util.apply_workspace_edit(action.edit, client.offset_encoding)
                  elseif action.command then
                    vim.lsp.buf.execute_command(action.command)
                  end
                end
              end
            end
          end
        end,
      })
      
      return opts
    end,
  },
}
