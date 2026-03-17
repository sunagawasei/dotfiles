{ pkgs, ... }:
{
  programs.fzf = {
    enable = true;
    enableZshIntegration = false; # zinit/zeno との順序制御のため手動
  };
}
