{ ... }:
let
  colors = import ./colors.nix;
in
{
  programs.hunk = {
    enable = true;
    enableGitIntegration = true;
    enableClaudeIntegration = true;

    settings = {
      theme = "custom";
      mode = "auto";
      line_numbers = true;
      wrap_lines = true;

      custom_theme = {
        label = colors.metadata.name;

        background = colors.core.background;
        panel = colors.core.panel_bg;
        panelAlt = colors.core.ui_shadow;
        border = colors.teals.border;
        accent = colors.teals.bright;
        accentMuted = colors.teals.standard;
        text = colors.foregrounds.main;
        muted = colors.foregrounds.dim;

        addedBg = colors.nvim.diff_add_bg;
        removedBg = colors.nvim.diff_delete_bg;
        movedAddedBg = colors.nvim.diff_add_inline_bg;
        movedRemovedBg = colors.nvim.diff_delete_inline_bg;
        contextBg = colors.core.background;
        addedContentBg = colors.nvim.diff_add_inline_bg;
        removedContentBg = colors.nvim.diff_delete_inline_bg;
        contextContentBg = colors.core.background;
        addedSignColor = colors.git.added;
        removedSignColor = colors.git.deleted;

        lineNumberBg = colors.core.background;
        lineNumberFg = colors.foregrounds.subdued;
        selectedHunk = colors.core.selection_bg;

        badgeAdded = colors.git.added;
        badgeRemoved = colors.git.deleted;
        badgeNeutral = colors.blues_slates.punctuation_gray;

        fileNew = colors.git.added;
        fileDeleted = colors.git.deleted;
        fileRenamed = colors.purples.muted_purple;
        fileModified = colors.git.changed;
        fileUntracked = colors.foregrounds.dim;

        noteBorder = colors.teals.border;
        noteBackground = colors.core.panel_bg;
        noteTitleBackground = colors.teals.dark_accent;
        noteTitleText = colors.foregrounds.main;

        syntax = {
          default = colors.semantic.variable;
          keyword = colors.semantic.keyword;
          string = colors.semantic.string;
          comment = colors.semantic.comment;
          number = colors.semantic.number;
          function = colors.semantic.function;
          property = colors.semantic.constant;
          type = colors.semantic.type;
          variable = colors.semantic.variable;
          operator = colors.semantic.operator;
          punctuation = colors.semantic.punctuation;
        };
      };
    };
  };
}
