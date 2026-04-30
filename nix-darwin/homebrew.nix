{
  nix-homebrew,
  lib,
  pkgs,
  config,
  ...
}:
let
  cfg = config.homebrew;
in
{
  nix-homebrew = {
    enable = true;
    user = "s23159";
    enableRosetta = false;
    autoMigrate = true;
  };

  homebrew = {
    enable = true;

    onActivation = {
      autoUpdate = true;
      upgrade = true;
      cleanup = "uninstall";
    };

    taps = [
      "cycloud-io/tap"
      "perman/tap"
    ];

    brews = [
      # カスタム tap
      "cycloud-cli" # cycloud-io/tap
      "perman-aws-vault" # perman/tap
      # nixpkgs で問題があるもの
      "aws-cdk" # nodePackages 版が不安定
      "mysql-client" # nixpkgs にスタンドアロン版なし
      "oauth2l" # macOS 認証統合
      "oxfmt" # brew 版が大幅に新しい (0.41 vs 0.27)
      "zsh" # /etc/shells 統合が必要なシステムシェル
    ];

    casks = [
      # ユーティリティ
      "alt-tab"
      "azookey"
      "homerow"
      # アプリケーション
      "claude-code@latest"
      "codex"
      "copilot-cli"
      "kindle"
      "linear-linear"
      "microsoft-remote-desktop"
      "notion"
      "raycast"
      "spotify"
      # ターミナル
      "wezterm@nightly"
      # フォント
      "font-geist"
      "font-geist-mono"
      "font-geist-mono-nerd-font"
      "font-hackgen-nerd"
      "font-ibm-plex"
      "font-plemol-jp-nf"
    ];
  };

  # nix-darwin の activation は #!/usr/bin/env -i bash で全環境変数をワイプするため
  # HOMEBREW_GITHUB_API_TOKEN を外部から渡せない。
  # gh auth token で動的に取得して env コマンドに注入する。
  # 上書き元: nix-darwin modules/homebrew.nix lines 941-955
  # 注意: nix-darwin バージョンアップ時はアップストリームの変更を確認してこのスクリプトを更新すること
  system.activationScripts.homebrew.text = lib.mkForce ''
    echo >&2 "Homebrew bundle..."
    if [ -f "${cfg.prefix}/bin/brew" ]; then
      _BREW_TOKEN=""
      _GH_BIN=""
      for _p in "${cfg.prefix}/bin/gh" "/Users/${cfg.user}/.nix-profile/bin/gh" "/run/current-system/sw/bin/gh"; do
        [ -x "$_p" ] && { _GH_BIN="$_p"; break; }
      done
      if [ -n "$_GH_BIN" ]; then
        _UID=$(id -u ${lib.escapeShellArg cfg.user} 2>/dev/null)
        if [ -n "$_UID" ]; then
          # launchctl asuser でユーザーセッション（キーチェーン含む）のコンテキストで gh を実行
          _BREW_TOKEN=$(launchctl asuser "$_UID" /usr/bin/sudo -u ${lib.escapeShellArg cfg.user} -H "$_GH_BIN" auth token 2>/dev/null)
        fi
      fi
      _TOKEN_ENV=()
      if [ -n "$_BREW_TOKEN" ]; then
        _TOKEN_ENV=("HOMEBREW_GITHUB_API_TOKEN=$_BREW_TOKEN")
      fi
      PATH="${cfg.prefix}/bin:${lib.makeBinPath [ pkgs.mas ]}:$PATH" \
      sudo \
        --preserve-env=PATH \
        --user=${lib.escapeShellArg cfg.user} \
        --set-home \
        env \
        "''${_TOKEN_ENV[@]}" \
        ${cfg.onActivation.brewBundleCmd}
    else
      echo -e "\e[1;31merror: Homebrew is not installed, skipping...\e[0m" >&2
    fi
  '';
}
