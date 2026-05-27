return {
  {
    "delphinus/md-render.nvim",
    version = "*",
    ft = { "markdown" },
    dependencies = {
      { "nvim-tree/nvim-web-devicons", version = "*" },
      { "delphinus/budoux.lua", version = "*" },
    },
    keys = {
      { "<leader>mr", "<Plug>(md-render-toggle)",      desc = "Markdown Render Toggle (in-place)" },
      { "<leader>mR", "<Plug>(md-render-preview)",     desc = "Markdown Render Float Preview" },
      { "<leader>mS", "<Plug>(md-render-split)",       desc = "Markdown Render Split (source+render)" },
      { "<leader>mT", "<Plug>(md-render-preview-tab)", desc = "Markdown Render Tab Preview" },
      { "<leader>md", "<Plug>(md-render-demo)",        desc = "Markdown Render Demo" },
      { "<leader>ma", "<Plug>(md-render-auto)",        desc = "Markdown Render Auto-toggle" },
    },
  },
}
