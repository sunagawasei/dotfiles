{ config, lib, pkgs, gws, herdr, gh-cli, ... }:
let
  # herdrへのローカルパッチ適用・検証用派生一式(パッチの「なぜ」はherdr.nix側に記載)。
  herdrPatched = (import ./herdr.nix { inherit herdr; }).patched;

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

  # mdrollのラスタライズ見出し(herdrペインで使う経路)はfc-matchでしかCJKフォントを
  # 引かないため、fc-matchが無いと日本語が豆腐になる。macOSのフォント置き場を教える。
  # familyはweztermのfont_with_fallbackに揃える。AssetsV2はOsakaの実体(on-demand asset)。
  # 日本語ruleを先に置くのは、後続ruleの置換先family名に依存させないため。
  mdrollFontsConf = pkgs.writeText "mdroll-fonts.conf" ''
    <?xml version="1.0"?>
    <!DOCTYPE fontconfig SYSTEM "fonts.dtd">
    <fontconfig>
      <dir>/System/Library/Fonts</dir>
      <dir>/System/Library/Fonts/Supplemental</dir>
      <dir>/System/Library/AssetsV2/com_apple_MobileAsset_Font8</dir>
      <dir>/Library/Fonts</dir>
      <dir>~/Library/Fonts</dir>
      <cachedir>~/.cache/fontconfig</cachedir>
      <match target="pattern">
        <test name="family"><string>sans</string></test>
        <test name="lang" compare="contains"><string>ja</string></test>
        <edit name="family" mode="assign" binding="strong">
          <string>Osaka</string>
          <string>Hiragino Kaku Gothic ProN</string>
        </edit>
      </match>
      <match target="pattern">
        <test name="family"><string>sans</string></test>
        <edit name="family" mode="assign" binding="strong"><string>GeistMono NF</string></edit>
      </match>
    </fontconfig>
  '';

  # cargoHash経路のvendorスクリプト(python-requests)はcrates.ioのAPIポリシーで403に
  # なるため、nixのfetchurlで各crateを取るcargoLock経路を使う。
  # lockは上流からコピー（`"${src}/Cargo.lock"`はsystem evalごとにsrc取得を強制する）。
  mdroll = pkgs.rustPlatform.buildRustPackage rec {
    pname = "mdroll";
    version = "0.4.2";
    src = pkgs.fetchFromGitHub {
      owner = "tokuhirom";
      repo = "mdroll";
      rev = "v${version}";
      hash = "sha256-f3rXbLi9WFRid/BG2PcNF0JPWOE3scWhz3Smmohzy5w=";
    };
    cargoLock.lockFile = ./mdroll-Cargo.lock;
    nativeBuildInputs = [ pkgs.makeWrapper ];
    # --setにするのは、将来グローバルなfontconfig設定が入っても同じ不具合を再発させないため。
    postInstall = ''
      wrapProgram $out/bin/mdroll \
        --set FONTCONFIG_FILE ${mdrollFontsConf} \
        --prefix PATH : ${pkgs.fontconfig.bin}/bin
    '';
  };

  # gh-boardはnixpkgs未収録のためソースビルドする。crates.ioがpython-requestsのUAを403で
  # 弾くためcargoHash(fetch-cargo-vendor)経路は使えず、fetchurlで取るimportCargoLockを使う。
  ghBoardSchemaCommit = "baf144f319c7705e822de9a26f05d12e1c7c9df4";
  ghBoardSchema = pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/octokit/graphql-schema/${ghBoardSchemaCommit}/schema.graphql";
    hash = "sha256-PGLQUm0TPO5TIhyJ3ptFWt4k23i5561W1kLEwVvOJlQ=";
  };

  ghBoard = pkgs.rustPlatform.buildRustPackage rec {
    pname = "gh-board";
    version = "1.5.0";
    src = pkgs.fetchFromGitHub {
      owner = "uzimaru0000";
      repo = "gh-board";
      rev = "v${version}";
      hash = "sha256-gYoNRBQiSAim3/PAo6DSAbDvlaWQnGDTZyXvfnu4Qsc=";
    };
    cargoDeps = pkgs.rustPlatform.importCargoLock { lockFile = "${src}/Cargo.lock"; };
    nativeBuildInputs = [ pkgs.makeWrapper ];
    # build.rsはschema.graphqlが無いとgh api経由で取りに行くため事前配置して短絡させる。
    # version更新時に古いschemaが残るのを防ぐためCargo.tomlのpinと一致を検査する。
    postPatch = ''
      grep -q '${ghBoardSchemaCommit}' Cargo.toml \
        || (echo "gh-board: schema commit mismatch, update ghBoardSchemaCommit" >&2; exit 1)
      cp ${ghBoardSchema} schema.graphql
    '';
    # 認証をgh auth tokenへshell outするためPATHにghを埋める。
    postInstall = ''
      wrapProgram $out/bin/gh-board --prefix PATH : ${pkgs.lib.makeBinPath [ gh-cli ]}
    '';
    doCheck = false;
  };

  # Claudeへ`sudo darwin-rebuild switch`だけを開放するための引数なしラッパー。
  # settings.jsonの`Bash(sudo:*)` denyは維持したまま`Bash(darwin-apply)`のみallowする
  # (denyはallowより先に評価されるため、sudoを含む形では例外を作れない)。
  darwinApply = pkgs.writeShellApplication {
    name = "darwin-apply";
    text = ''
      if [ "$#" -ne 0 ]; then
        echo "darwin-apply: takes no arguments (flake target is fixed to this host)" >&2
        exit 2
      fi
      # attrを省略するとdarwin-rebuildがscutil --get LocalHostNameで解決するため、
      # PC交換でhostnameが変わってもここは触らなくてよい。
      # rebuild失敗時はチェックへ進まずその終了コードで抜けるため、execにはしない。
      status=0
      sudo /run/current-system/sw/bin/darwin-rebuild switch --flake "$HOME/.config" || status=$?
      if [ "$status" -ne 0 ]; then
        exit "$status"
      fi

      # herdr serverが旧バイナリのまま稼働していないかを1点だけ確認する。
      # ps の comm/args は symlink 経由のパスを返し実体判定には使えないため、
      # 稼働中バイナリの store path は lsof の txt 行(実行中コードの参照先)で見る。
      server_pid=""
      while IFS=' ' read -r pid args; do
        case "$args" in
          */bin/herdr\ server | */bin/herdr\ server\ *)
            server_pid="$pid"
            break
            ;;
        esac
      done < <(ps -axo pid=,args=)

      if [ -n "$server_pid" ]; then
        running_herdr="$(lsof -p "$server_pid" 2>/dev/null \
          | awk '$4 == "txt" && $9 ~ /^\/nix\/store\// && $9 ~ /\/bin\/herdr$/ { print $9; exit }')" || true
        current_user="$(id -un)"
        current_herdr="$(realpath "/etc/profiles/per-user/$current_user/bin/herdr" 2>/dev/null || true)"
        if [ -n "$running_herdr" ] && [ -n "$current_herdr" ] && [ "$running_herdr" != "$current_herdr" ]; then
          echo "darwin-apply: herdr server は旧バイナリで稼働中。再起動が要る" >&2
        fi
      fi

      exit "$status"
    '';
  };

  # 尊師スタイル(内蔵キーボード上にroBaを載せる運用)用にKarabiner-Elements profileを
  # 切り替える。--select-profileは存在しないprofile名でも終了コード0を返すため、
  # 切り替え後に--show-current-profile-nameで読み戻して検証する。
  sonshi = pkgs.writeShellApplication {
    name = "sonshi";
    text = ''
      karabiner_cli="/Library/Application Support/org.pqrs/Karabiner-Elements/bin/karabiner_cli"

      usage() {
        echo "usage: sonshi {on|off|status}" >&2
      }

      switch_to() {
        local target="$1"
        "$karabiner_cli" --select-profile "$target"
        local current
        current="$("$karabiner_cli" --show-current-profile-name)"
        if [ "$current" != "$target" ]; then
          echo "sonshi: profile '$target' への切り替えに失敗した(現在の profile: '$current')" >&2
          exit 1
        fi
        echo "sonshi: profile を '$target' へ切り替えた"
      }

      show_status() {
        local current
        current="$("$karabiner_cli" --show-current-profile-name)"
        case "$current" in
          sonshi)
            echo "sonshi: 選択中の profile は 'sonshi'(内蔵キーボードのキーを常に無効にする設定)。profile名の確認のみで、ルール内容の健全性は保証しない。"
            ;;
          normal)
            echo "sonshi: 選択中の profile は 'normal'(内蔵キーボードを使う設定)。profile名の確認のみで、ルール内容の健全性は保証しない。"
            ;;
          *)
            echo "sonshi: 選択中の profile が 'sonshi'/'normal' のどちらでもない('$current')" >&2
            exit 1
            ;;
        esac
      }

      if [ "$#" -gt 1 ]; then
        usage
        exit 2
      fi

      case "''${1:-status}" in
        on)
          switch_to sonshi
          ;;
        off)
          switch_to normal
          ;;
        status)
          show_status
          ;;
        *)
          usage
          exit 2
          ;;
      esac
    '';
  };
in
{
  home.packages = with pkgs; [
    nixfmt
    nil # Nix LSP

    # ファイル操作
    fd tree unar eza

    # テキスト処理
    bat jq yq-go wget pv gron nkf

    # セキュリティ
    gnupg pinentry_mac git-crypt bitwarden-cli

    # 開発ツール
    gh-cli
    lazygit neovim deno bun luarocks lua-language-server hadolint markdownlint-cli
    grpcurl buf gitui imagemagick gifski pwgen tmux ansifilter
    ripgrep oxlint unzip yamlfmt

    # GitHub Projects v2 TUI
    ghBoard

    # GitHub issue/PR トリアージ TUI
    gh-dash

    # その他
    zoxide

    # Google Workspace CLI
    gws

    # TUI スプレッドシート
    sheets

    # ターミナル Markdown ビューア
    mdroll

    # AIエージェント用ターミナルマルチプレクサ
    herdrPatched

    # nix-darwin適用ラッパー(Claudeへの限定開放用)
    darwinApply

    # 尊師スタイル用Karabiner-Elements profile切り替え
    sonshi
  ];

  # eza/bat のテーマ配置先を用意し、bat のテーマキャッシュを毎回再構築する
  home.activation.buildBatCache = lib.hm.dag.entryAfter [ "linkGeneration" ] ''
    $DRY_RUN_CMD mkdir -p \
      "${config.xdg.configHome}/eza" \
      "${config.xdg.configHome}/bat/themes"
    $DRY_RUN_CMD env \
      XDG_CONFIG_HOME="${config.xdg.configHome}" \
      XDG_CACHE_HOME="${config.xdg.cacheHome}" \
      ${pkgs.bat}/bin/bat cache --build
  '';

  # direnv: シェル統合を HM に任せる
  programs.direnv = {
    enable = true;
    nix-direnv.enable = true;
    # loading/export ログを抑制（direnv>=2.36 では direnv.toml に log_filter="^$" を生成）。
    # env var DIRENV_LOG_FORMAT は 2.37.1 の export 経路で無視されるため silent を使う。
    silent = true;
  };
}
