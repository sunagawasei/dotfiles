{ pkgs, git-base, ... }:
let
  colors = import ./colors.nix;
in
{
  programs.git = {
    enable = true;

    # nixpkgs-unstableのgitは2.56.0。2026-11-10公開のセキュリティ修正(2.56.1)がnixpkgsに入るまでの暫定。
    # nixpkgsが2.56.1以上を配るようになったらこのoverrideを外す。
    package = git-base.overrideAttrs (old: rec {
      version = "2.56.1";
      src = pkgs.fetchurl {
        url = "https://www.kernel.org/pub/software/scm/git/git-${version}.tar.xz";
        hash = "sha256-8LC/iSZLm9LhxkA2PnUpZDmJ2nYoFgBHPFB0XNNKYH8=";
      };
    });

    lfs.enable = true;

    ignores = [
      ".DS_Store"
      ".perman-aws-vault"
      "**/.claude/settings.local.json"
      ".ignore"
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
      plus-emph-style = "${colors.ansi.bright_white} ${colors.nvim.diff_add_inline_bg}";
      minus-emph-style = "${colors.ansi.bright_white} ${colors.nvim.diff_delete_inline_bg}";
      syntax-theme = "ghost-visor";
      line-numbers-plus-style = "${colors.git.added}";
      line-numbers-minus-style = "${colors.git.deleted}";
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
