# herdr(ターミナルマルチプレクサ)へのローカルパッチ適用と、パッチ検証用の派生一式。
# `herdr` は flake input から渡される派生そのもの（`herdr.packages.aarch64-darwin.default`）。
{ herdr }:
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
  # sidebarトークンの区切りを" · "から" › "へ変更（spaces/agentsパネル共通）。
  # copy modeでy/Enterによるyank後もcopy modeに留まる（上流はyank後に必ず抜ける実装で
  # 設定キーなし）。選択ハイライトは解除しカーソル・スクロール位置は維持、q/Escでの
  # 退出時のみ進入時スクロール位置へ復元する。
  # pane削除の確認判定を、tty前景プロセスグループがpane一次プロセス自身のグループか
  # どうかで行う。グループ全メンバーの走査だと、reparent後もshellのpgidを保つ常駐
  # プロセス(zeno.zshのdenoサーバ)に恒久的にヒットし、アイドルなpaneで確認が出続ける。
  # 非focus paneの端末内容をTerminalモード中もdim表示する(上流はNavigateモード限定)。
  # TerminalDirtyPatchの部分再描画パスにも同じdim判定を適用し、dimが抜けて
  # 高頻度更新paneがちらつくのを防ぐ。
  # combined-frame-digestは個別機能ではなく、上記UIパッチ全部の合成描画を固定する
  # integration fixture。必ず最後に適用し、UIパッチを変えたらdigestを再生成する。
  # 並び順は開発branch v082-upgrade(専用worktreeで保持)のcommit順と一致させる。
  # 各パッチは親commit時点のツリーに対するdiffなので、順を崩すとoffset依存になる。
  patches = [
    ./patches/herdr-active-pane-border-white.patch
    ./patches/herdr-confirm-close-running-process.patch
    ./patches/herdr-sidebar-token-separator.patch
    ./patches/herdr-copy-mode-yank-stay.patch
    ./patches/herdr-expanded-sidebar-space-numbers.patch
    ./patches/herdr-confirm-close-leader-only.patch
    ./patches/herdr-dim-inactive-panes.patch
    ./patches/herdr-combined-frame-digest.patch
  ];
in
{
  # パッチ適用後のherdr本体。
  patched = herdr.overrideAttrs (old: {
    patches = (old.patches or [ ]) ++ patches;
  });

  # パッチ適用前のsrc(fileset適用済み。herdr本体のビルド入力そのもの)。
  srcUnpatched = herdr.src;

  # パッチ適用後のsrc。コンパイルは走らせず、unpack+patchのみ実行してツリーを$outへ出す。
  # herdr本体(rustPlatform.buildRustPackage)のnativeBuildInputsは継承しない。
  # cargoSetupHookのpostUnpack/postPatchフックが.cargo/config.tomlやvendorコピーを
  # ツリーへ生成し、パッチとは無関係な差分を混入させるため。
  srcPatched = herdr.stdenv.mkDerivation {
    pname = "${herdr.pname}-patched-src";
    inherit (herdr) version;
    src = herdr.src;
    inherit patches;
    phases = [ "unpackPhase" "patchPhase" "installPhase" ];
    installPhase = ''
      mkdir -p "$out"
      cp -a . "$out"/
    '';
  };
}
