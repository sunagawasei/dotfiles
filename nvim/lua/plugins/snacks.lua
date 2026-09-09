-- lsp_symbols の on_show が登録する vim.on_key handler。前回の解除漏れを次回に外すため保持する
local follow_ns ---@type integer?

-- .gitignore を無視して検索する。代わりに rgignore でリポジトリ横断のノイズを落とす
local rg_args = { "--no-ignore-vcs", "--ignore-file", vim.fn.stdpath("config") .. "/rgignore" }

return {
  "folke/snacks.nvim",
  dependencies = { "delphinus/md-render.nvim" },
  opts = {
    dashboard = {
      enabled = false,  -- oil.nvim起動のためdashboardを無効化
    },
    terminal = {
      enabled = false,  -- toggleterm使用のため無効化
    },
    lazygit = {
      enabled = false,  -- toggletermで実装
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
          preview = function(ctx)
            require("md-render.snacks").preview()(ctx)
          end,
        },
        grep = {
          args = rg_args,
          preview = function(ctx)
            require("md-render.snacks").preview()(ctx)
          end,
        },
        lsp_symbols = {
          -- カーソル行を含むシンボルを初期選択にする。symbol は LSP から非同期で
          -- 届くので、finder/matcher の完了を待ってから選択を移す
          on_show = function(picker)
            if follow_ns then
              vim.on_key(nil, follow_ns)
              follow_ns = nil
            end

            local win = picker.main
            if not (win and vim.api.nvim_win_is_valid(win)) then
              return
            end
            local row = vim.api.nvim_win_get_cursor(win)[1] - 1 -- LSP と同じ 0-based
            local deadline = vim.uv.now() + 10000

            -- 待機中にユーザーが打鍵したら選択を動かさない。typed が空のものは
            -- 非typed key と、同じ打鍵から展開された2つ目以降の key なので除く
            local interacted = false
            follow_ns = vim.on_key(function(_, typed)
              if typed ~= "" then
                interacted = true
              end
            end)
            local ns = follow_ns

            local function finish(idx)
              if follow_ns == ns then
                vim.on_key(nil, ns)
                follow_ns = nil
              end
              if idx and not interacted and not picker.closed then
                picker.list:view(idx)
              end
            end

            local step

            -- pcall で包んだ入口。defer_fn からの再入もここを経由させ、
            -- runtime error で on_key handler が残らないようにする
            local function guarded()
              local ok, err = pcall(step)
              if not ok then
                finish(nil)
                vim.notify("lsp_symbols on_show: " .. tostring(err), vim.log.levels.ERROR)
              end
            end

            step = function()
              if picker.closed or interacted then
                return finish(nil)
              end

              if picker.finder:running() or picker.matcher.task:running() then
                if vim.uv.now() < deadline then
                  return vim.defer_fn(guarded, 30)
                end
                return finish(nil)
              end

              local items = {}
              for i = 1, picker.list:count() do
                items[i] = picker.list:get(i)
              end
              finish(require("util.lsp_symbol_cursor").select_index(items, row))
            end

            guarded()
          end,
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
    -- LazyVimデフォルトのgit diff pickerを無効化（diffview.nvimを使用）
    { "<leader>gd", false },
  },
}
