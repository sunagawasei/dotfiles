return {
  "nvim-treesitter/nvim-treesitter-context",
  event = { "BufReadPost", "BufNewFile" },
  opts = {
    max_lines = 3,
    mode = "cursor",
  },
  keys = {
    {
      "[c",
      function()
        require("treesitter-context").go_to_context()
      end,
      desc = "Go to context",
    },
  },
}
