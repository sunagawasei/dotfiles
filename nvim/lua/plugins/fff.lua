return {
  {
    "dmtrKovalenko/fff.nvim",
    version = "^0.10.6",
    lazy = true,
    build = function()
      require("fff.download").download_or_build_binary()
    end,
    keys = {
      {
        "<leader><space>",
        function()
          require("fff").find_files({ cwd = LazyVim.root() })
        end,
        desc = "Find Files (Root Dir, fff)",
      },
      {
        "<leader>sg",
        function()
          require("fff").live_grep({ cwd = LazyVim.root() })
        end,
        desc = "Grep (Root Dir, fff)",
      },
      {
        "<leader>fP",
        function()
          -- LazyVim.root()はLSPルートを先に見るため、nvim配下を開いていると~/.config/nvimに絞られる
          require("fff").find_files({ cwd = LazyVim.root.git() })
        end,
        desc = "Find Files (Git Root, fff)",
      },
    },
  },
  {
    -- lazy.nvimは同一lhsを複数specが登録すると最後に張った側が勝ち、順序は決まらない。
    -- fffを確実に効かせるにはsnacks側をfalseで外すしかない
    "folke/snacks.nvim",
    keys = {
      { "<leader><space>", false },
      { "<leader>sg", false },
    },
  },
}
