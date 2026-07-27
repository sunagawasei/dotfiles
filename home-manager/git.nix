{ ... }:
let
  colors = import ./colors.nix;
in
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
    options = {
      plus-style = "syntax ${colors.nvim.diff_add_bg}";
      minus-style = "syntax ${colors.nvim.diff_delete_bg}";
      plus-emph-style = "syntax ${colors.nvim.diff_add_inline_bg}";
      minus-emph-style = "syntax ${colors.nvim.diff_delete_inline_bg}";
      line-numbers-plus-style = "${colors.semantic.success}";
      line-numbers-minus-style = "${colors.semantic.error}";
      line-numbers-zero-style = "${colors.foregrounds.subdued}";
      line-numbers-left-style = "${colors.foregrounds.subdued}";
      line-numbers-right-style = "${colors.foregrounds.subdued}";
      hunk-header-line-number-style = "${colors.foregrounds.subdued}";
      hunk-header-decoration-style = "${colors.teals.border} box";
      file-style = "${colors.foregrounds.heading}";
      whitespace-error-style = "${colors.semantic.error}";
    };
  };
}
