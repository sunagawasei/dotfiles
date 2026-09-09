return {
  "dmtrKovalenko/fff.nvim",
  version = "^0.10.6",
  lazy = true,
  build = function()
    require("fff.download").download_or_build_binary()
  end,
  keys = {
    {
      "<leader>fP",
      function()
        require("fff").find_files({ cwd = LazyVim.root() })
      end,
      desc = "Find Files (fff)",
    },
    {
      "<leader>sF",
      function()
        require("fff").live_grep({ cwd = LazyVim.root() })
      end,
      desc = "Grep (fff)",
    },
  },
}
