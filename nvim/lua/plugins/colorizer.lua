return {
  {
    "catgoose/nvim-colorizer.lua",
    event = { "BufReadPre", "BufNewFile" },
    config = function()
      require("colorizer").setup({
        "*", -- すべてのファイルタイプでハイライトを有効化
        css = {
          RGB = true, -- #RGB hex codes
          RRGGBB = true, -- #RRGGBB hex codes
          names = true, -- "Blue" や "Red" などの色名
          RRGGBBAA = true, -- #RRGGBBAA hex codes
          rgb_fn = true, -- CSS rgb() and rgba() functions
          hsl_fn = true, -- CSS hsl() and hsla() functions
          mode = "background", -- 背景色でハイライト
        },
        html = {
          mode = "background",
        },
        lua = {
          mode = "background",
        },
        yaml = {
          mode = "background",
        },
        json = {
          mode = "background",
        },
        markdown = {
          mode = "background",
          names = false, -- マークダウンでは色名を無効化
        },
      }, {
        -- デフォルトオプション
        RGB = true, -- #RGB hex codes
        RRGGBB = true, -- #RRGGBB hex codes  
        names = true, -- "Blue" や "Blue" などの色名
        RRGGBBAA = false, -- #RRGGBBAA hex codes
        rgb_fn = false, -- CSS rgb() and rgba() functions
        hsl_fn = false, -- CSS hsl() and hsla() functions
        css = false, -- CSSキーワードを有効化
        css_fn = false, -- CSS *functions* を有効化
        mode = "background", -- 背景色でハイライト (background|foreground|virtualtext)
        tailwind = false, -- Tailwind CSSカラー ("bg-blue-200", "text-blue-600", etc.)を有効化
        sass = { enable = false }, -- sass色の解析とハイライト
        virtualtext = "■", -- virtualtext mode用の文字
      })
    end,
  },
}
