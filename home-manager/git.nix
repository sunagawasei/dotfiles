{ ... }:

{
  programs.git = {
    enable = true;
    userName = "sunagawasei";
    userEmail = "132873106+sunagawasei@users.noreply.github.com";

    lfs.enable = true;

    delta.enable = true;

    aliases = {
      dw = "diff --ignore-all-space";
      sw = "show --ignore-all-space";
      dcw = "diff --cached --ignore-all-space";
    };

    ignores = [
      ".DS_Store"
      ".perman-aws-vault"
      "**/.claude/settings.local.json"
    ];

    extraConfig = {
      core = {
        editor = "vim";
        quotepath = false;
      };
      init.defaultBranch = "main";
      merge.conflictstyle = "zdiff3";
    };
  };
}
