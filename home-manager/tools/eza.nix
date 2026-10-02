{ config, pkgs, ... }:
{
  home.packages = [ pkgs.eza ];
  my.human.aliases.ls = "eza --color=auto";
  my.env.EZA_CONFIG_DIR = "${config.xdg.configHome}/eza";
}
