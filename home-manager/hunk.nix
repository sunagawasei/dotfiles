{ ... }:
{
  programs.hunk = {
    enable = true;
    enableGitIntegration = true;
    enableClaudeIntegration = true;

    settings = {
      theme = "custom";
      mode = "auto";
      line_numbers = true;

      custom_theme = {
        label = "Abyssal Teal";

        background = "#0b0c0c";
        panel = "#152a2b";
        panelAlt = "#1e1e24";
        border = "#4d8f9e";
        accent = "#6cd8d3";
        accentMuted = "#659d9e";
        text = "#cef5f2";
        muted = "#92a2ab";

        addedBg = "#0d1f1f";
        removedBg = "#151515";
        movedAddedBg = "#1e4a40";
        movedRemovedBg = "#4a2328";
        contextBg = "#0b0c0c";
        addedContentBg = "#1e4a40";
        removedContentBg = "#4a2328";
        contextContentBg = "#0b0c0c";
        addedSignColor = "#6ab9a8";
        removedSignColor = "#a37aa7";

        lineNumberBg = "#0b0c0c";
        lineNumberFg = "#525b65";
        selectedHunk = "#64bbbe";

        badgeAdded = "#6ab9a8";
        badgeRemoved = "#a37aa7";
        badgeNeutral = "#7a8599";

        fileNew = "#6ab9a8";
        fileDeleted = "#a37aa7";
        fileRenamed = "#8a99bd";
        fileModified = "#6cd8d3";
        fileUntracked = "#92a2ab";

        noteBorder = "#4d8f9e";
        noteBackground = "#152a2b";
        noteTitleBackground = "#304d4f";
        noteTitleText = "#cef5f2";

        syntax = {
          default = "#cef5f2";
          keyword = "#8a99bd";
          string = "#659d9e";
          comment = "#7a869a";
          number = "#b1f4ed";
          function = "#6cd8d3";
          property = "#ced5e9";
          type = "#a4abcb";
          variable = "#cef5f2";
          operator = "#64bbbe";
          punctuation = "#7a8599";
        };
      };
    };
  };
}
