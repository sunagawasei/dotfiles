return {
  {
    "hat0uma/csvview.nvim",
    ft = { "csv", "tsv" }, -- CSV/TSVファイルで自動ロード
    opts = {
      parser = {
        -- コメント行のプレフィックス設定
        comments = { "#", "//" },
      },
      view = {
        -- デフォルトの表示モード（"highlight" または "border"）
        display_mode = "border",
      },
      keymaps = {
        -- テキストオブジェクト設定（CSVフィールドの選択）
        textobject_field_inner = { "if", mode = { "o", "x" } },
        textobject_field_outer = { "af", mode = { "o", "x" } },
      },
    },
    keys = {
      -- トグルキーの設定
      { "<leader>cv", "<cmd>CsvViewToggle<cr>", desc = "Toggle CSV View" },
    },
  },
}
