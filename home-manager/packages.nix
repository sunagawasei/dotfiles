{ pkgs, gws, herdr, ... }:
let
  # アクティブpane枠=白 / 非アクティブ=青（デフォルトは逆）にするための上流パッチ。
  # レガシー制御バイト31(US)の逆デコードがCtrl+-になっておりCtrl+_/Ctrl+/が
  # kittyキーボードプロトコル配下（例: Neovim）で別キーに化けるバグの修正。
  # pane削除(prefix+x)時、フォーカスpane内でshell以外のプロセス（neovim等）が
  # 実行中なら閉じる前に確認ダイアログを出す機能追加（wezterm skip_close_confirmation
  # _for_processes_named の逆相当）。config `confirm_close_running_process`(既定true)。
  # 非フォーカスpaneをターミナルモード中も暗く(SGR DIM)する。上流は should_dim を
  # !terminal_active でナビゲートモード限定にしていたのを外し、作業中もフォーカスを明示。
  herdrPatched = herdr.overrideAttrs (old: {
    patches = (old.patches or [ ]) ++ [
      ./patches/herdr-active-pane-border-white.patch
      ./patches/herdr-ctrl-underscore-decode.patch
      ./patches/herdr-confirm-close-running-process.patch
      ./patches/herdr-dim-unfocused-in-terminal.patch
    ];
  });

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
    gh lazygit neovim deno bun luarocks lua-language-server hadolint markdownlint-cli
    grpcurl gitui imagemagick gifski pwgen tmux ansifilter
    ripgrep oxlint unzip

    # その他
    zoxide

    # Google Workspace CLI
    gws

    # TUI スプレッドシート
    sheets

    # AIエージェント用ターミナルマルチプレクサ
    herdrPatched
  ];

  # direnv: シェル統合を HM に任せる
  programs.direnv = {
    enable = true;
    nix-direnv.enable = true;
    # loading/export ログを抑制（direnv>=2.36 では direnv.toml に log_filter="^$" を生成）。
    # env var DIRENV_LOG_FORMAT は 2.37.1 の export 経路で無視されるため silent を使う。
    silent = true;
  };
}
