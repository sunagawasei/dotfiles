return {
  "choplin/code-review.nvim",
  event = "VeryLazy", -- デフォルトキーマップ(<leader>rc等)を早期に有効化
  config = function()
    require("code-review").setup({
      output = {
        format = "minimal", -- AI向け簡潔形式
      },
    })
  end,
  keys = {
    {
      "<leader>ra",
      function()
        local ok_state, state = pcall(require, "code-review.state")
        local ok_fmt, formatter = pcall(require, "code-review.formatter")
        if not ok_state or not ok_fmt then
          vim.notify("code-review.nvim is not loaded", vim.log.levels.ERROR)
          return
        end

        local comments = state.get_comments()
        if not comments or #comments == 0 then
          vim.notify("No review comments", vim.log.levels.WARN)
          return
        end

        local content = formatter.format(comments)
        local cwd_hash = vim.fn.sha256(vim.fn.getcwd()):sub(1, 8)
        local tmpfile = "/tmp/claude-review-" .. cwd_hash .. ".md"
        vim.fn.writefile(vim.split(content, "\n", { plain = true }), tmpfile)

        local ok_cc, cc = pcall(require, "claudecode")
        if not ok_cc then
          vim.notify("claudecode.nvim is not loaded", vim.log.levels.ERROR)
          return
        end

        local success, err = cc.send_at_mention(tmpfile)
        if success then
          vim.notify("Review sent to Claude Code (" .. #comments .. " comments)")
        else
          vim.notify("Failed: " .. (err or "unknown"), vim.log.levels.ERROR)
        end
      end,
      desc = "Send review to Claude Code",
    },
  },
}
