{
  description = "macOS system configuration (nix-darwin + home-manager + nix-homebrew)";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nixpkgs-cursor.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nixpkgs-gh.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nixpkgs-rift.url = "github:nixos/nixpkgs/master";
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
    herdr.url = "github:herdrdev/herdr/v0.8.2";
    hunk = {
      url = "github:modem-dev/hunk";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    # recall(claude-recall)など babarot のツールを、リリース済みバイナリのまま取り込む
    babarot = {
      url = "github:babarot/nur-packages";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-cursor,
      nixpkgs-gh,
      nixpkgs-rift,
      home-manager,
      nix-darwin,
      nix-homebrew,
      gws-cli,
      herdr,
      hunk,
      babarot,
      ...
    }:
    let
      herdrPkgs = import ./home-manager/herdr.nix { herdr = herdr.packages.aarch64-darwin.default; };
    in
    {
      # attr名はhostname。darwin-rebuildは--flakeでattrを省略すると
      # scutil --get LocalHostName で解決するため、PC交換時はここを改名する。
      darwinConfigurations."CA-20038442" = nix-darwin.lib.darwinSystem {
        specialArgs = { inherit self nix-homebrew gws-cli herdr hunk babarot nixpkgs-cursor nixpkgs-gh nixpkgs-rift; };
        modules = [
          ./nix-darwin/configuration.nix
          home-manager.darwinModules.home-manager
          nix-homebrew.darwinModules.nix-homebrew
        ];
      };

      # herdrパッチ検査CLI(反映確認)向けに、パッチ適用前後のsrcと本体を公開する。
      packages.aarch64-darwin = {
        herdr-src-unpatched = herdrPkgs.srcUnpatched;
        herdr-src-patched = herdrPkgs.srcPatched;
        herdr-patched = herdrPkgs.patched;
      };
    };
}
