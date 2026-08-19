{ nixpkgs-cursor, gws-cli, herdr, hunk, ... }:
let
  # cursor-cli だけ nixpkgs-unstable 追従。root の nixpkgs pin は動かさない
  cursorPkgs = import nixpkgs-cursor {
    system = "aarch64-darwin";
    config.allowUnfreePredicate = pkg: (pkg.pname or "") == "cursor-cli";
  };
  # Esc/Ctrl+G で実行中ターンが止まらないようにする。CLI に設定が無いためバンドルを書き換える
  cursorCli = cursorPkgs.cursor-cli.overrideAttrs (old: {
    nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ cursorPkgs.perl ];
    postPatch = (old.postPatch or "") + ''
      perl ${./cursor-agent-no-esc-abort.pl} *.js
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
  };
  home-manager.users."s23159" = {
    imports = [
      ../home-manager/home.nix
      hunk.homeManagerModules.default
    ];
  };
}
