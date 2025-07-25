-- GitGraph highlight groups with Vercel Geist colors
local M = {}

M.setup = function()
  -- Commit information highlights
  vim.api.nvim_set_hl(0, "GitGraphHash", { fg = "#0070F3" }) -- Geist Blue 8
  vim.api.nvim_set_hl(0, "GitGraphTimestamp", { fg = "#A3A3A3" }) -- Geist Gray 6
  vim.api.nvim_set_hl(0, "GitGraphAuthor", { fg = "#16A085" }) -- Geist Green 8
  vim.api.nvim_set_hl(0, "GitGraphBranchName", { fg = "#7928CA" }) -- Geist Purple 7
  vim.api.nvim_set_hl(0, "GitGraphBranchTag", { fg = "#F5A623" }) -- Geist Amber 7
  vim.api.nvim_set_hl(0, "GitGraphBranchMsg", { fg = "#FFFFFF" }) -- Geist Gray 10 (Dark)

  -- Branch color highlights
  vim.api.nvim_set_hl(0, "GitGraphBranch1", { fg = "#3291FF" }) -- Geist Blue 7
  vim.api.nvim_set_hl(0, "GitGraphBranch2", { fg = "#29BC9B" }) -- Geist Green 7
  vim.api.nvim_set_hl(0, "GitGraphBranch3", { fg = "#A855F7" }) -- Geist Purple 6
  vim.api.nvim_set_hl(0, "GitGraphBranch4", { fg = "#FFCC33" }) -- Geist Amber 6
  vim.api.nvim_set_hl(0, "GitGraphBranch5", { fg = "#14B8A6" }) -- Geist Teal 6
end

-- Set up autocmd to reapply highlights on colorscheme change
vim.api.nvim_create_autocmd("ColorScheme", {
  pattern = "*",
  callback = function()
    M.setup()
  end,
  group = vim.api.nvim_create_augroup("GitGraphHighlights", { clear = true }),
})

return M