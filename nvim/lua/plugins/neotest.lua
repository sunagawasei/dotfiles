return {
  {
    "nvim-neotest/neotest",
    dependencies = {
      "nvim-neotest/nvim-nio",
      "nvim-lua/plenary.nvim",
      "antoinemadec/FixCursorHold.nvim",
      "nvim-treesitter/nvim-treesitter",
      "fredrikaverpil/neotest-golang",
    },
    keys = {
      { "<leader>t", "", desc = "+test" },
    },
    config = function()
      local neotest = require("neotest")

      -- Go言語アダプタの直接初期化
      local success, neotest_golang = pcall(require, "neotest-golang")
      if not success then
        vim.notify("Failed to load neotest-golang", vim.log.levels.ERROR)
        return
      end

      neotest.setup({
        adapters = {
          neotest_golang({
            go_test_args = { "-v", "-count=1", "-timeout=60s" },
            dap_go_enabled = true,
          }),
        },
        output = {
          enabled = true,
          open_on_run = true, -- nearest testの結果を自動表示
        },
        output_panel = {
          enabled = true,
          open = "botright split | resize 15", -- ボトムスプリットで15行表示
        },
      })

      -- キーマッピング（setup後に設定）
      local map = vim.keymap.set
      local opts = { noremap = true, silent = true }

      -- Test execution mappings (LazyVim-style)
      map("n", "<leader>tr", function() neotest.run.run() end, vim.tbl_extend("force", opts, { desc = "Run Nearest" }))
      map("n", "<leader>tt", function()
        neotest.run.run(vim.fn.expand("%"))
        vim.defer_fn(function()
          -- Output panelが閉じている場合のみ開く
          local output_panel_open = false
          for _, win in ipairs(vim.api.nvim_list_wins()) do
            local buf = vim.api.nvim_win_get_buf(win)
            local buf_name = vim.api.nvim_buf_get_name(buf)
            if buf_name:match("Neotest Output Panel") then
              output_panel_open = true
              break
            end
          end
          if not output_panel_open then
            neotest.output_panel.toggle()
          end
        end, 500)
      end, vim.tbl_extend("force", opts, { desc = "Run File" }))
      map("n", "<leader>tT", function() neotest.run.run(vim.uv.cwd()) end, vim.tbl_extend("force", opts, { desc = "Run All Test Files" }))
      map("n", "<leader>tl", function() neotest.run.run_last() end, vim.tbl_extend("force", opts, { desc = "Run Last" }))

      -- Output and summary toggles
      map("n", "<leader>ts", function() neotest.summary.toggle() end, vim.tbl_extend("force", opts, { desc = "Toggle Summary" }))
      map("n", "<leader>to", function() neotest.output.open({ enter = true, auto_close = true }) end, vim.tbl_extend("force", opts, { desc = "Show Output" }))
      map("n", "<leader>tO", function() neotest.output_panel.toggle() end, vim.tbl_extend("force", opts, { desc = "Toggle Output Panel" }))

      -- Control mappings
      map("n", "<leader>tS", function() neotest.run.stop() end, vim.tbl_extend("force", opts, { desc = "Stop" }))
      map("n", "<leader>tw", function() neotest.watch.toggle(vim.fn.expand("%")) end, vim.tbl_extend("force", opts, { desc = "Toggle Watch" }))
      map("n", "<leader>td", function() neotest.run.run({ strategy = "dap" }) end, vim.tbl_extend("force", opts, { desc = "Debug Nearest" }))

      -- Extended mappings with custom args
      map("n", "<leader>tp", function() neotest.run.run(vim.fn.expand("%:h")) end, vim.tbl_extend("force", opts, { desc = "Run Package/Directory" }))
      map("n", "<leader>tc", function() neotest.run.run({ extra_args = { "-cover", "-v" } }) end, vim.tbl_extend("force", opts, { desc = "Run with Coverage" }))
      map("n", "<leader>tC", function() neotest.run.run(vim.fn.expand("%"), { extra_args = { "-cover", "-v" } }) end, vim.tbl_extend("force", opts, { desc = "Run File with Coverage" }))
      map("n", "<leader>tA", function() neotest.run.run({ suite = true, extra_args = { "-v", "-race" } }) end, vim.tbl_extend("force", opts, { desc = "Run All with Race Detection" }))
    end,
  },
}
