{ config, lib, pkgs, gws, herdr, gh-cli, ... }:
let
  # アクティブpane枠=白 / 非アクティブ=青（デフォルトは逆）にするための上流パッチ。
  # pane削除(prefix+x)時、フォーカスpane内でshell以外のプロセス（neovim等）が
  # 実行中なら閉じる前に確認ダイアログを出す機能追加（wezterm skip_close_confirmation
  # _for_processes_named の逆相当）。config `confirm_close_running_process`(既定true)。
  # accent等の濃色バッジ上の文字色panel_contrast_fg（theme=terminalではDarkGray=ANSI 8で、
  # ghost-visorだとbg(ANSI 4)と同化し不可視）をoverlay1（同White）へ変更。active tab・
  # navigator選択行・PREFIX等20箇所が改善。bgが明色になるRESIZEバッジ(mauve=ANSI 7)と
  # config警告(yellow=ANSI 3)の2箇所のみ旧ロジック(_dim)据え置き。
  # [theme.custom]のsurface_dim overrideはsidebar選択行のbgにも波及するため不採用。
  # expanded sidebarのspaces一覧にもcollapsed railと同じ表示順番号を常時表示
  # （上流は番号なしを明示テストで固定した意図的設計・configキーなし）。
  # 同一覧の2行目のdirラベルは0.7.4のrows設定+$dirメタデータで表示
  # （herdr/config.tomlのrows設定とherdr-task-label hookの分業。パッチ不使用）。
  # sidebarトークンの区切りを" · "から" › "へ変更（spaces/agents両パネル共通）。
  # copy modeでy/Enterによるyank後もcopy modeに留まる（上流はyank後に必ず抜ける実装で
  # 設定キーなし）。選択ハイライトは解除しカーソル・スクロール位置は維持、q/Escでの
  # 退出時のみ進入時スクロール位置へ復元する。
  # pane削除の確認判定を、tty前景プロセスグループがpane一次プロセス自身のグループか
  # どうかで行う。グループ全メンバーの走査だと、reparent後もshellのpgidを保つ常駐
  # プロセス(zeno.zshのdenoサーバ)に恒久的にヒットし、アイドルなpaneで確認が出続ける。
  # combined-frame-digestは個別機能ではなく、上記UIパッチ全部の合成描画を固定する
  # integration fixture。必ず最後に適用し、UIパッチを変えたらdigestを再生成する。
  # 並び順は開発branch v080-upgrade(専用worktreeで保持)のcommit順と一致させる。
  # 各パッチは親commit時点のツリーに対するdiffなので、順を崩すとoffset依存になる。
  herdrPatched = herdr.overrideAttrs (old: {
    patches = (old.patches or [ ]) ++ [
      ./patches/herdr-active-pane-border-white.patch
      ./patches/herdr-confirm-close-running-process.patch
      ./patches/herdr-sidebar-token-separator.patch
      ./patches/herdr-copy-mode-yank-stay.patch
      ./patches/herdr-expanded-sidebar-space-numbers.patch
      ./patches/herdr-confirm-close-leader-only.patch
      ./patches/herdr-combined-frame-digest.patch
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
      exec sudo /run/current-system/sw/bin/darwin-rebuild switch --flake "$HOME/.config"
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
