{
  description = "macOS system configuration (nix-darwin + home-manager + nix-homebrew)";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nixpkgs-cursor.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nix-darwin = {
      url = "github:nix-darwin/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nix-homebrew.url = "github:zhaofengli/nix-homebrew";
    gws-cli.url = "github:googleworkspace/cli";
    herdr.url = "github:herdrdev/herdr/v0.8.0";
    hunk = {
      url = "github:modem-dev/hunk";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-cursor,
      home-manager,
      nix-darwin,
      nix-homebrew,
      gws-cli,
      herdr,
      hunk,
      ...
    }:
    {
      # attr名はhostname。darwin-rebuildは--flakeでattrを省略すると
      # scutil --get LocalHostName で解決するため、PC交換時はここを改名する。
      darwinConfigurations."CA-20038442" = nix-darwin.lib.darwinSystem {
        specialArgs = { inherit self nix-homebrew gws-cli herdr hunk nixpkgs-cursor; };
        modules = [
          ./nix-darwin/configuration.nix
          home-manager.darwinModules.home-manager
          nix-homebrew.darwinModules.nix-homebrew
        ];
      };
    };
}
