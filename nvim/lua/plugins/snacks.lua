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
          exclude = { ".git", ".DS_Store" }, -- .gitディレクトリ自体を除外
        },
        files = {
          hidden = true, -- 隠しファイルを表示
          ignored = true, -- gitignoreされたファイルも表示
          exclude = {}, -- 除外パターンを空にして.gitも検索対象にする
        },
      },
      -- スマートピッカーの設定
      smart = {
        sources = {
          files = {
            hidden = true,
            ignored = true,
            exclude = {}, -- .gitディレクトリも検索対象にする
          },
        },
      },
    },
  },
}

