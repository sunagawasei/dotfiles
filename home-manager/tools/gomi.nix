{ config, pkgs, ... }:
let
  home = config.home.homeDirectory;
in
{
  # 誤削除をゴミ箱から戻せる(gomi -b)ようにする。
  # Claude Code の Bash ツールはコマンドごとに全 alias を外す(snapshot 先頭の unalias -a)ので、
  # alias でなく PATH 上の rm wrapper で、人間にも AI にも届ける
  home.packages = [
    pkgs.gomi
    (pkgs.writeShellScriptBin "rm" ''exec ${pkgs.gomi}/bin/gomi "$@"'')
  ];

  # 既定は /var を削除禁止にするが、macOS の $TMPDIR(/var/folders/...)が削除できなくなる。
  # JSON は YAML のサブセットなので toJSON で書く
  xdg.configFile."gomi/config.yaml".text = builtins.toJSON {
    core = {
      trash = {
        strategy = "auto";
        home_fallback = true;
        gomi_dir = "${home}/.gomi";
        forbidden_paths = [
          "${home}/.local/share/Trash"
          "${home}/.trash"
          "/tmp/Trash"
          "/var/tmp/Trash"
          "${home}/.gomi"
          "/"
          "/etc"
          "/usr"
          "/bin"
          "/sbin"
          "/lib"
          "/lib64"
        ];
      };
      restore = {
        confirm = true;
        verbose = true;
      };
      permanent_delete.enable = false;
    };
  };
}
