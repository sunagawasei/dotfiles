-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here

-- avante.nvim keymaps
vim.keymap.set("n", "<leader>aa", function() require("avante.api").ask() end, { desc = "avante: ask" })
vim.keymap.set("v", "<leader>ar", function() require("avante.api").refresh() end, { desc = "avante: refresh" })
vim.keymap.set("v", "<leader>ae", function() require("avante.api").edit() end, { desc = "avante: edit" })
vim.keymap.set("n", "<leader>af", function() require("avante.api").focus() end, { desc = "avante: focus" })
