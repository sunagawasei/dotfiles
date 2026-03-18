{
  self,
  lib,
  ...
}:
{
  nixpkgs.hostPlatform = "aarch64-darwin";
  system.stateVersion = 6;
  nix.enable = false;
  system.configurationRevision = self.rev or self.dirtyRev or null;
  security.pam.services.sudo_local.touchIdAuth = true;

  programs.zsh.shellInit = ''
    export ZDOTDIR="$HOME/.config/zsh"
  '';
  system.defaults.NSGlobalDomain.AppleShowAllExtensions = true; # ファイル拡張子を常に表示
  system.defaults.finder.AppleShowAllFiles = true; # 隠しファイルを表示
  system.defaults.finder.ShowPathbar = true; # パスバーを表示
  programs.zsh.enableGlobalCompInit = false;
  users.users."s23159".home = "/Users/s23159";
  system.primaryUser = "s23159";
  nixpkgs.config.allowUnfreePredicate =
    pkg:
    builtins.elem (lib.getName pkg) [
      "terraform"
    ];

  imports = [
    ./home_manager.nix
    ./homebrew.nix
  ];
}
