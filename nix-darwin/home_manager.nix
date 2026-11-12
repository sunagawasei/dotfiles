{ nixpkgs-cursor, nixpkgs-gh, nixpkgs-git, nixpkgs-rift, nixpkgs-stable, gws-cli, herdr, hunk, babarot, ... }:
let
  # cursor-cli だけ nixpkgs-unstable 追従。root の nixpkgs pin は動かさない
  cursorPkgs = import nixpkgs-cursor {
    system = "aarch64-darwin";
    config.allowUnfreePredicate = pkg: (pkg.pname or "") == "cursor-cli";
  };
  # gh だけ nixpkgs-unstable 追従。root の nixpkgs pin は動かさない
  ghPkgs = import nixpkgs-gh { system = "aarch64-darwin"; };
  # git だけ nixpkgs-unstable 追従(2.56.1のセキュリティ修正をoverrideで入れる土台)。root の nixpkgs pin は動かさない
  gitPkgs = import nixpkgs-git { system = "aarch64-darwin"; };
  # rift-wm だけ nixpkgs-unstable 追従。root の nixpkgs pin(3/28)には rift-wm が無い
  riftPkgs = import nixpkgs-rift { system = "aarch64-darwin"; };
  # cdk・oxfmt は nixpkgs の stable 版(26.05)を使う。root の nixpkgs pin は動かさない
  stablePkgs = import nixpkgs-stable { system = "aarch64-darwin"; };
  # CLI に設定が無いためバンドルを書き換える。
  # Esc/Ctrl+G では実行中ターンを止めない。Alt+J/K/U/D は Claude Code と同じスクロール。
  cursorCli = cursorPkgs.cursor-cli.overrideAttrs (old: {
    nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ cursorPkgs.perl ];
    postPatch = (old.postPatch or "") + ''
      perl ${./cursor-agent-no-esc-abort.pl} *.js
      perl ${./cursor-agent-alt-scroll.pl} *.js
    '';
  });
in
{
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.extraSpecialArgs = {
    gws = gws-cli.packages.aarch64-darwin.gws;
    herdr = herdr.packages.aarch64-darwin.default;
    cursor-cli = cursorCli;
    gh-cli = ghPkgs.gh;
    git-base = gitPkgs.git;
    rift = riftPkgs.rift-wm;
    aws-cdk-cli = stablePkgs.aws-cdk-cli;
    oxfmt = stablePkgs.oxfmt;
    claude-recall = babarot.packages.aarch64-darwin.claude-recall;
  };
  home-manager.users."s23159" = {
    imports = [
      ../home-manager/home.nix
      hunk.homeManagerModules.default
    ];
  };
}
