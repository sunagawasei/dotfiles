{ gws-cli, ... }:
{
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.extraSpecialArgs = {
    gws = gws-cli.packages.aarch64-darwin.gws;
  };
  home-manager.users."s23159" = ../home-manager/home.nix;
}
