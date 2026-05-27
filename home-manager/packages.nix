{ pkgs, gws, ... }:
let
  sheets = pkgs.buildGoModule {
    pname = "sheets";
    version = "unstable-a5af8b3";
    src = pkgs.fetchFromGitHub {
      owner = "maaslalani";
      repo = "sheets";
      rev = "a5af8b38bb68003d041d9827d12914a5ae5ace7e";
      hash = "sha256-LxAlttxefsi+bzS8bSErcZwK+rkMFTzhrPBzqvyi1Dc=";
    };
    vendorHash = "sha256-WWtAt0+W/ewLNuNgrqrgho5emntw3rZL9JTTbNo4GsI=";
  };
in
{
  home.packages = with pkgs; [
    nixfmt
    nil # Nix LSP

    # ファイル操作
    fd tree unar

    # テキスト処理
    jq wget pv gron nkf

    # セキュリティ
    gnupg pinentry_mac git-crypt bitwarden-cli

    # 開発ツール
    gh lazygit neovim deno bun luarocks hadolint markdownlint-cli
    grpcurl gitui imagemagick gifski pwgen tmux ansifilter
    ripgrep oxlint unzip

    # インフラ
    colima docker qemu

    # その他
    terminal-notifier zoxide

    # Google Workspace CLI
    gws

    # TUI スプレッドシート
    sheets
  ];

  # direnv: シェル統合を HM に任せる
  programs.direnv = {
    enable = true;
    nix-direnv.enable = true;
  };
}
