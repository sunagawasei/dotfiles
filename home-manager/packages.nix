{ pkgs, ... }:
{
  home.packages = with pkgs; [
    nixfmt

    # ファイル操作
    fd tree unar

    # テキスト処理
    jq wget

    # セキュリティ
    gnupg pinentry_mac git-crypt bitwarden-cli

    # 開発ツール
    gh lazygit neovim deno bun luarocks hadolint markdownlint-cli

    # インフラ
    colima docker qemu

    # その他
    terminal-notifier zoxide
  ];

  # direnv: シェル統合を HM に任せる
  programs.direnv = {
    enable = true;
    nix-direnv.enable = true;
  };
}
