return {
  "nvim-mini/mini.surround",
  opts = {
    mappings = {
      add = "gsa",
      delete = "gsd",
      replace = "gsr",
      find = "gsf",
      find_left = "gsF",
      highlight = "gsh",
      update_n_lines = "gsn",
    },
    custom_surroundings = {
      c = {
        input = { "/%*().-()%*/" },
        output = { left = "/* ", right = " */" },
      },
    },
  },
  keys = {
    { "gsa", desc = "Surround Add", mode = { "n", "v" } },
    { "gsd", desc = "Surround Delete" },
    { "gsr", desc = "Surround Replace" },
  },
}
