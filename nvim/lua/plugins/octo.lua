return {
  "pwntester/octo.nvim",
  cmd = "Octo",
  dependencies = {
    "nvim-lua/plenary.nvim",
    "nvim-tree/nvim-web-devicons",
  },
  keys = {
    { "<leader>gi", "<cmd>Octo issue list<cr>", desc = "Issues (Octo)" },
    { "<leader>gp", "<cmd>Octo pr list<cr>", desc = "PRs (Octo)" },
    { "<leader>gs", "<cmd>Octo search<cr>", desc = "Search (Octo)" },
  },
  opts = {
    picker = "snacks",
  },
  config = function(_, opts)
    require("octo").setup(opts)

    -- Guard against async callbacks firing after the user already closed the buffer.
    -- octo/init.lua:111 calls nvim_buf_call(bufnr, ...) inside a vim.schedule callback;
    -- if bufnr is destroyed before the gh API responds, Neovim raises "Invalid buffer id".
    -- Remove this patch once the fix is merged upstream (pwntester/octo.nvim).
    local octo = require("octo")
    local utils = require("octo.utils")
    local uri = require("octo.uri")
    octo.load_buffer = function(lopts)
      lopts = lopts or {}
      local bufnr = lopts.bufnr or vim.api.nvim_get_current_buf()
      local cursor_pos = vim.api.nvim_win_get_cursor(0)
      local bufname = vim.fn.bufname(bufnr)
      local buffer_info = uri.parse(bufname)
      if buffer_info == nil then
        utils.print_err("Cannot parse buffer name: " .. bufname)
        return
      end
      local repo, kind, id, hostname =
        buffer_info.repo, buffer_info.kind, buffer_info.id, buffer_info.hostname

      octo.load(repo, kind, id, hostname, function(obj)
        if not vim.api.nvim_buf_is_valid(bufnr) then
          return
        end
        vim.api.nvim_buf_call(bufnr, function()
          octo.create_buffer(kind, obj, repo, false, hostname)
          local lines = vim.api.nvim_buf_line_count(bufnr)
          local new_cursor_pos = {
            math.min(cursor_pos[1], lines),
            math.max(0, cursor_pos[2] - 1),
          }
          vim.api.nvim_win_set_cursor(0, new_cursor_pos)
          if lopts.verbose then
            utils.info(string.format("Loaded %s/%s/%d", repo, kind, id))
          end
        end)
      end)
    end
  end,
}
