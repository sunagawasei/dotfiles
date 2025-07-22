return {
  {
    "nvim-neo-tree/neo-tree.nvim",
    opts = {
      filesystem = {
        filtered_items = {
          hide_dotfiles = false,  -- dotfileを常に表示
          hide_gitignored = false, -- gitignoreされたファイルも表示
        },
      },
    },
  },
}