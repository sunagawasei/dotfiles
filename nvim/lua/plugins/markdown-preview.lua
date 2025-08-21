return {
  {
    "iamcco/markdown-preview.nvim",
    cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
    build = "cd app && npm install",
    ft = { "markdown" },
    init = function()
      -- システムテーマを検出してMarkdownプレビューのテーマを設定
      local function set_markdown_preview_theme()
        if vim.fn.has("mac") == 1 then
          -- macOSのシステムテーマを検出
          local handle = io.popen("defaults read -g AppleInterfaceStyle 2>/dev/null")
          if handle then
            local result = handle:read("*a")
            handle:close()
            if result and result:match("Dark") then
              vim.g.mkdp_theme = "dark"
            else
              vim.g.mkdp_theme = "light"
            end
          else
            vim.g.mkdp_theme = "light" -- デフォルトはライト
          end
        else
          -- macOS以外はダークテーマをデフォルトに
          vim.g.mkdp_theme = "dark"
        end
      end

      -- 初回設定
      set_markdown_preview_theme()

      -- システムテーマ変更時に自動更新
      vim.api.nvim_create_autocmd({ "VimEnter", "FocusGained" }, {
        callback = set_markdown_preview_theme,
        desc = "Update markdown preview theme based on system theme",
      })

      -- 基本設定
      vim.g.mkdp_auto_start = 0 -- 自動開始を無効
      vim.g.mkdp_auto_close = 1 -- バッファ切り替え時に自動クローズ
      vim.g.mkdp_refresh_slow = 0 -- リアルタイム更新
      vim.g.mkdp_command_for_global = 0 -- Markdownファイルのみで有効
      vim.g.mkdp_open_to_the_world = 0 -- ローカルのみでプレビュー
      vim.g.mkdp_open_ip = "" -- デフォルトIPを使用
      vim.g.mkdp_browser = "" -- デフォルトブラウザを使用
      vim.g.mkdp_echo_preview_url = 0 -- プレビューURLをエコーしない
      vim.g.mkdp_preview_options = {
        mkit = {}, -- markdown-itオプション
        katex = {}, -- KaTeXオプション
        uml = {}, -- markdown-it-plantumlオプション
        maid = {}, -- mermaidオプション
        disable_sync_scroll = 0, -- 同期スクロールを有効
        sync_scroll_type = "middle", -- スクロール位置を中央に
        hide_yaml_meta = 1, -- YAMLメタデータを隠す
        sequence_diagrams = {}, -- js-sequence-diagramsオプション
        flowchart_diagrams = {}, -- flowchart.jsオプション
        content_editable = false, -- コンテンツ編集を無効
        disable_filename = 0, -- ファイル名表示を有効
        toc = {}, -- Table of Contentsオプション
      }
      vim.g.mkdp_markdown_css = "" -- カスタムMarkdown CSSなし
      vim.g.mkdp_highlight_css = "" -- カスタムハイライトCSSなし
      vim.g.mkdp_port = "" -- ランダムポートを使用
      vim.g.mkdp_page_title = "「${name}」" -- ページタイトルのフォーマット
      vim.g.mkdp_filetypes = { "markdown" } -- 対応ファイルタイプ
    end,
    keys = {
      { "<leader>mp", "<cmd>MarkdownPreviewToggle<cr>", desc = "Markdown Preview Toggle" },
      { "<leader>ms", "<cmd>MarkdownPreviewStop<cr>", desc = "Markdown Preview Stop" },
    },
  },
}
