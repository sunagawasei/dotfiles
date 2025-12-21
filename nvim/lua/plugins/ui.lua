return {
  {
    "akinsho/bufferline.nvim",
    opts = {
      options = {
        always_show_bufferline = true,
        custom_filter = function(buf)
          -- [No Name]バッファを非表示（claudecode.nvim新タブdiff対策）
          return vim.fn.bufname(buf) ~= ""
        end,
      },
    },
  },
  {
    "folke/noice.nvim",
    opts = {
      presets = {
        -- LSPドキュメントにボーダーを追加
        lsp_doc_border = true,
      },
    },
  },
}
