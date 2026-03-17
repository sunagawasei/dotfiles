{ ... }:

{
  programs.git = {
    enable = true;

    lfs.enable = true;

    ignores = [
      ".DS_Store"
      ".perman-aws-vault"
      "**/.claude/settings.local.json"
    ];

    settings = {
      user = {
        name = "sunagawasei";
        email = "132873106+sunagawasei@users.noreply.github.com";
      };
      alias = {
        dw = "diff --ignore-all-space";
        sw = "show --ignore-all-space";
        dcw = "diff --cached --ignore-all-space";
      };
      core = {
        editor = "vim";
        quotepath = false;
      };
      init.defaultBranch = "main";
      merge.conflictstyle = "zdiff3";
    };
  };

  programs.delta = {
    enable = true;
    enableGitIntegration = true;
  };
}
