{ config, pkgs, ... }:
let
  home = config.home.homeDirectory;
in
{
  home.packages = [ pkgs.gomi ];
  # AI も rm を打つ。rm 互換オプションを受けるので、誤削除をゴミ箱から戻せる(gomi -b)ようにする
  my.human.aliases.rm = "gomi";
  my.ai.aliases.rm = "gomi";

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
