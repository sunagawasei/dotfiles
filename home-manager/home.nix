{ config, pkgs, ... }:

{
  imports = [
    ./git.nix
    ./packages.nix
    ./fzf.nix
    ./cloud.nix
    ./dev.nix
    ./shell.nix
  ];

  home.username = "s23159";
  home.homeDirectory = "/Users/s23159";
  home.stateVersion = "25.11";

  programs.home-manager.enable = true;
}
