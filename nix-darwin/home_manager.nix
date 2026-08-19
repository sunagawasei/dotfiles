{ nixpkgs-cursor, gws-cli, herdr, hunk, ... }:
let
  # cursor-cli だけ nixpkgs-unstable 追従。root の nixpkgs pin は動かさない
  cursorPkgs = import nixpkgs-cursor {
    system = "aarch64-darwin";
    config.allowUnfreePredicate = pkg: (pkg.pname or "") == "cursor-cli";
  };
in
{
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.extraSpecialArgs = {
    gws = gws-cli.packages.aarch64-darwin.gws;
    herdr = herdr.packages.aarch64-darwin.default;
    cursor-cli = cursorPkgs.cursor-cli;
  };
  home-manager.users."s23159" = {
    imports = [
      ../home-manager/home.nix
      hunk.homeManagerModules.default
    ];
  };
}
