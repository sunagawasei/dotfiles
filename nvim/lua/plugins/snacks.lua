return {
  "folke/snacks.nvim",
  opts = {
    explorer = {
      -- 隠しファイルを表示
      hidden = true,
      -- gitignoreされたファイルも表示
      show_hidden = true,
      -- デフォルトのフィルター設定
      filters = {
        dotfiles = false, -- dotfilesを隠さない
        git_ignored = false, -- gitignoreされたファイルを隠さない
      },
    },
    picker = {
      -- ピッカーでも隠しファイルを表示
      hidden = true,
      show_hidden = true,
      sources = {
        explorer = {
          hidden = true,
          ignored = true, -- git除外ファイルを表示（.git/info/excludeを含む）
        },
      },
    },
  },
}