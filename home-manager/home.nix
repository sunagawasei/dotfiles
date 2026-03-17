{ config, pkgs, ... }:

{
  imports = [
    ./git.nix
  ];

  home.username = "s23159";
  home.homeDirectory = "/Users/s23159";
  home.stateVersion = "25.11";

  home.packages = with pkgs; [
    nixfmt
  ];

  home.sessionVariables = {};

  programs.home-manager.enable = true;
}
