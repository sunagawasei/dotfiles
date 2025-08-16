return {
  "echasnovski/mini.icons",
  opts = {
    -- すべてのアイコンから色を削除
    style = "glyph",
    
    -- デフォルトハイライトを無効化
    default = {
      file = { hl = "Normal" },
      extension = { hl = "Normal" },
      directory = { hl = "Normal" },
      lsp = { hl = "Normal" },
      os = { hl = "Normal" },
    },
    
    -- アイコンのハイライトを統一
    filetype = {
      lua = { hl = "Normal" },
      vim = { hl = "Normal" },
      javascript = { hl = "Normal" },
      typescript = { hl = "Normal" },
      python = { hl = "Normal" },
      json = { hl = "Normal" },
      markdown = { hl = "Normal" },
      yaml = { hl = "Normal" },
      toml = { hl = "Normal" },
      txt = { hl = "Normal" },
    },
  },
}
