return {
  "folke/snacks.nvim",
  opts = {
    lazygit = {
      enabled = true,
      configure = true,  -- 自動テーマ設定
      config = {
        os = { editPreset = "nvim-remote" },
        gui = { nerdFontsVersion = "3" },
      },
      win = {
        style = "lazygit",
        wo = {
          scrolloff = 0,  -- スクロールオフセットを無効化してパフォーマンス改善
        },
      },
    },
    picker = {
      sources = {
        explorer = {
          hidden = true,
          ignored = true,
          exclude = { ".git", ".DS_Store" },
        },
        files = {
          hidden = true,
          ignored = true,
        },
      },
      -- カスタムアクションの定義
      actions = {
        copy_diagnostic = function(picker)
          local item = picker:current()
          if not item then
            vim.notify("No item selected", vim.log.levels.INFO)
            return
          end

          local text_to_copy = nil

          -- 診断情報がある場合
          if item.diagnostics and #item.diagnostics > 0 then
            local messages = {}
            for _, diag in ipairs(item.diagnostics) do
              table.insert(messages, diag.message)
            end
            text_to_copy = table.concat(messages, "\n")
          -- LSP結果やファイル内容がある場合
          elseif item.text then
            text_to_copy = item.text
          elseif item.value then
            text_to_copy = item.value
          -- ファイルパスがある場合は行内容を取得
          elseif item.file and item.lnum then
            local ok, lines = pcall(vim.fn.readfile, item.file, "", item.lnum)
            if ok and lines and lines[item.lnum] then
              text_to_copy = lines[item.lnum]
            end
          end

          if text_to_copy then
            vim.fn.setreg('+', text_to_copy)
            vim.notify("Copied to clipboard: " .. text_to_copy:sub(1, 50) .. (text_to_copy:len() > 50 and "..." or ""), vim.log.levels.INFO)
          else
            vim.notify("Nothing to copy", vim.log.levels.WARN)
          end
        end,
      },
      -- picker内のキーマップ設定
      win = {
        input = {
          keys = {
            -- 診断コピー用キーマップ
            ["<C-y>"] = { "copy_diagnostic", mode = { "n", "i" } },
            ["gy"] = { "copy_diagnostic", mode = { "n" } },
          },
        },
      },
    },
    bigfile = {
      enabled = false, -- bigfile機能を完全に無効化
    },
    indent = {
      enabled = false, -- インデントガイドを無効化（全階層表示を防ぐ）
    },
    zen = {
      enabled = true, -- Zoom機能を有効化
    },
  },
  keys = {
    { "<leader>gg", function() Snacks.lazygit() end, desc = "Lazygit" },
  },
}

