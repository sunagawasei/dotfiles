{ gws-cli, herdr, hunk, ... }:
{
  home-manager.useGlobalPkgs = true;
  home-manager.useUserPackages = true;
  home-manager.extraSpecialArgs = {
    gws = gws-cli.packages.aarch64-darwin.gws;
    herdr = herdr.packages.aarch64-darwin.default;
  };
  home-manager.users."s23159" = {
    imports = [
      ../home-manager/home.nix
      hunk.homeManagerModules.default
    ];
  };
}
