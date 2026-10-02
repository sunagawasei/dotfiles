{ config, lib, pkgs, ... }:

{
  imports = [
    ./git.nix
    ./packages.nix
    ./fzf.nix
    ./cloud.nix
    ./dev.nix
    ./shell.nix
    ./zsh.nix
    ./my.nix
  ]
  # tools/ のファイルは置くだけで読み込まれる(消せば設定も一緒に消える)
  ++ lib.mapAttrsToList (name: _: ./tools + "/${name}") (
    lib.filterAttrs (name: type: type == "regular" && lib.hasSuffix ".nix" name) (builtins.readDir ./tools)
  );

  home.username = "s23159";
  home.homeDirectory = "/Users/s23159";
  home.stateVersion = "25.11";

  programs.home-manager.enable = true;
}
