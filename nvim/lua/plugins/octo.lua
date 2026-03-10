return {
  "pwntester/octo.nvim",
  cmd = "Octo",
  dependencies = {
    "nvim-lua/plenary.nvim",
    "nvim-tree/nvim-web-devicons",
  },
  keys = {
    { "<leader>gi", "<cmd>Octo issue list<cr>", desc = "Issues (Octo)" },
    { "<leader>gp", "<cmd>Octo pr list<cr>", desc = "PRs (Octo)" },
    { "<leader>gs", "<cmd>Octo search<cr>", desc = "Search (Octo)" },
  },
  opts = {
    picker = "snacks",
  },
}
