{
  self,
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

  programs.zsh.enableGlobalCompInit = false;
}
