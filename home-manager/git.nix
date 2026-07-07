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
        wdiff = "diff --color-words";
      };
      core = {
        editor = "vim";
        quotepath = false;
      };
      init.defaultBranch = "main";
      diff.wordRegex = "[a-zA-Z0-9_]+|[ぁ-ん]+|[ァ-ヶー]+|[一-龥々〇〆]+|.";
      diff.tool = "hunk";
      difftool.hunk.cmd = ''hunk difftool "$LOCAL" "$REMOTE" "$MERGED"'';
      difftool.prompt = false;
      merge.conflictstyle = "zdiff3";
    };
  };

  # lazygitのpager用（core.pagerはhunkのまま。deltaはlazygit/config.ymlのpagersからのみ呼ばれる）
  programs.delta = {
    enable = true;
    enableGitIntegration = false;
  };
}
