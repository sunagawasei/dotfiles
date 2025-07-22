return {
  "cappyzawa/trim.nvim",
  event = { "BufWritePre" },
  opts = {
    trim_last_line = false,
    patterns = {
      [[%s/\n*\%$/\r/]],
    },
  },
}