return {
  "nvim-treesitter/nvim-treesitter-context",
  event = { "BufReadPost", "BufNewFile" },
  opts = {
    max_lines = 3,
    mode = "topline",  -- ビューポートスクロール時のみ更新（カーソル移動では更新しない、パフォーマンス最適化）
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
