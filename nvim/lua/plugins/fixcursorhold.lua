-- FixCursorHold.nvim - CursorHoldイベントのパフォーマンス改善
-- Neovimの既知問題（neovim/neovim#12587）への対策
-- CursorHoldをupdatetimeから分離し、イベントキューのブロッキングを防止
return {
  "antoinemadec/FixCursorHold.nvim",
  event = { "BufReadPre", "BufNewFile" },
  config = function()
    -- CursorHoldの更新間隔を1000msに設定
    vim.g.cursorhold_updatetime = 1000
  end,
}
