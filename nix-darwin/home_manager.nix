{ gws-cli, herdr, ... }:
{
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.extraSpecialArgs = {
    gws = gws-cli.packages.aarch64-darwin.gws;
    herdr = herdr.packages.aarch64-darwin.default;
  };
  home-manager.users."s23159" = ../home-manager/home.nix;
}
