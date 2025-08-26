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
      { "<leader>t", desc = "Test" },
      { "<leader>tr", function() require("neotest").run.run() end, desc = "Run Nearest" },
      { "<leader>tt", function() require("neotest").run.run(vim.fn.expand("%")) end, desc = "Run File" },
      { "<leader>tT", function() require("neotest").run.run(vim.uv.cwd()) end, desc = "Run All Test Files" },
      { "<leader>tl", function() require("neotest").run.run_last() end, desc = "Run Last" },
      { "<leader>ts", function() require("neotest").summary.toggle() end, desc = "Toggle Summary" },
      { "<leader>to", function() require("neotest").output.open({ enter = true, auto_close = true }) end, desc = "Show Output" },
      { "<leader>tO", function() require("neotest").output_panel.toggle() end, desc = "Toggle Output Panel" },
      { "<leader>tS", function() require("neotest").run.stop() end, desc = "Stop" },
      { "<leader>tw", function() require("neotest").watch.toggle(vim.fn.expand("%")) end, desc = "Toggle Watch" },
      { "<leader>td", function() require("neotest").run.run({ strategy = "dap" }) end, desc = "Debug Nearest" },
      -- 拡張キーマップ
      { "<leader>tp", function() require("neotest").run.run(vim.fn.expand("%:h")) end, desc = "Run Package/Directory" },
      { "<leader>tc", function() require("neotest").run.run({ extra_args = { "-cover", "-v" } }) end, desc = "Run with Coverage" },
      { "<leader>tC", function() require("neotest").run.run(vim.fn.expand("%"), { extra_args = { "-cover", "-v" } }) end, desc = "Run File with Coverage" },
      { "<leader>tA", function() require("neotest").run.run({ suite = true, extra_args = { "-v", "-race" } }) end, desc = "Run All with Race Detection" },
    },
    config = function()
      local neotest = require("neotest")
      
      -- Go言語アダプタの直接初期化（これが核心的な解決策）
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
      })
    end,
  },
}
