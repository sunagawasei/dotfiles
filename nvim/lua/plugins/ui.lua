return {
  {
    "akinsho/bufferline.nvim",
    opts = function(_, opts)
      local p = require("config.palette").colors

      opts.options = vim.tbl_deep_extend("force", opts.options or {}, {
        always_show_bufferline = true,
        custom_filter = function(buf)
          -- [No Name]バッファを非表示（claudecode.nvim新タブdiff対策）
          return vim.fn.bufname(buf) ~= ""
        end,
      })

      -- bufferline の前景設定
      opts.highlights = {
        fill = { bg = "none" },
        background = { bg = "none", fg = p.light_gray },
        tab = { bg = "none", fg = p.light_gray },
        tab_selected = { bg = "none", fg = p.highlight_white, bold = true },
        tab_separator = { bg = "none", fg = p.border },
        tab_separator_selected = { bg = "none", fg = p.cyan },
        buffer_visible = { bg = "none", fg = p.fg },
        buffer_selected = { bg = "none", fg = p.highlight_white, bold = true, italic = false },
        separator = { bg = "none", fg = p.border },
        separator_visible = { bg = "none", fg = p.border },
        separator_selected = { bg = "none", fg = p.border },
        indicator_selected = { bg = "none", fg = p.cyan },
        modified = { bg = "none", fg = p.light_gray },
        modified_visible = { bg = "none", fg = p.lavender },
        modified_selected = { bg = "none", fg = p.cyan },
        close_button = { bg = "none", fg = p.light_gray },
        close_button_visible = { bg = "none", fg = p.fg },
        close_button_selected = { bg = "none", fg = p.magenta },
        numbers = { bg = "none", fg = p.light_gray },
        numbers_visible = { bg = "none", fg = p.light_gray },
        numbers_selected = { bg = "none", fg = p.highlight_white, bold = true },
        error = { bg = "none", fg = p.magenta },
        error_visible = { bg = "none", fg = p.magenta },
        error_selected = { bg = "none", fg = p.magenta },
        warning = { bg = "none", fg = p.bright_magenta },
        warning_visible = { bg = "none", fg = p.bright_magenta },
        warning_selected = { bg = "none", fg = p.bright_magenta },
        hint = { bg = "none", fg = p.cyan },
        hint_visible = { bg = "none", fg = p.cyan },
        hint_selected = { bg = "none", fg = p.cyan },
        info = { bg = "none", fg = p.bright_cyan },
        info_visible = { bg = "none", fg = p.bright_cyan },
        info_selected = { bg = "none", fg = p.bright_cyan },
        pick = { bg = "none", fg = p.magenta, bold = true },
        pick_visible = { bg = "none", fg = p.magenta, bold = true },
        pick_selected = { bg = "none", fg = p.magenta, bold = true },
        offset_separator = { bg = "none", fg = p.border },
        trunc_marker = { bg = "none", fg = p.light_gray },
      }

      -- 全グループを透過に揃える。デフォルトキーは実行時取得なので、
      -- プラグイン更新でグループが増えても不透明な帯が混ざらない
      local cfgmod = require("bufferline.config")
      cfgmod.setup({})
      cfgmod.apply(true)
      local highlight_names = vim.tbl_keys(cfgmod.highlights)
      if #highlight_names == 0 then
        error("bufferline default highlight discovery returned no groups")
      end

      for _, name in ipairs(highlight_names) do
        local highlight = opts.highlights[name] or {}
        highlight.bg = "none"
        opts.highlights[name] = highlight
      end
    end,
  },
  {
    "folke/noice.nvim",
    opts = {
      presets = {
        -- LSPドキュメントにボーダーを追加
        lsp_doc_border = true,
      },
    },
  },
}
