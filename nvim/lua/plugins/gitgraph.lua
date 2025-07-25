return {
  {
    "isakbm/gitgraph.nvim",
    dependencies = { "sindrets/diffview.nvim" },
    config = function(_, opts)
      -- Set up GitGraph highlights with Vercel Geist colors
      require("utils.gitgraph-highlights").setup()
      
      -- Set up GitGraph with options
      require("gitgraph").setup(opts)
    end,
    opts = {
      symbols = {
        merge_commit = "◉",
        commit = "●",
      },
      format = {
        timestamp = "%H:%M:%S %d-%m-%Y",
        fields = { "hash", "timestamp", "author", "branch_name", "tag" },
      },
      hooks = {
        on_select_commit = function(commit)
          vim.cmd(":DiffviewOpen " .. commit.hash .. "^!")
        end,
        on_select_range_commit = function(from, to)
          vim.cmd(":DiffviewOpen " .. from.hash .. "~1.." .. to.hash)
        end,
      },
    },
    keys = {
      {
        "<leader>gl",
        function()
          require("gitgraph").draw({}, { all = true, max_count = 5000 })
        end,
        desc = "GitGraph - Draw",
      },
    },
  },
}