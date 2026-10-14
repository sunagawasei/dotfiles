{ pkgs, ... }:
let
  borders = pkgs.jankyborders;
in
{
  home.packages = [ borders ];

  # 枠は窓の外側に描かれるので、太さ(width)は rift.nix の gaps と同じ値にする。
  # active_color は colors/ghost-visor.toml の [semantic] keyword (#C69BF6)。非フォーカスは透明
  # order=above だと枠のウィンドウが窓の上に重なり、Shottr のウィンドウ撮影が窓でなく枠を選んで枠線しか撮れない
  launchd.agents.jankyborders = {
    enable = true;
    config = {
      Label = "git.felixkratz.jankyborders";
      ProgramArguments = [
        "${borders}/bin/borders"
        "style=round"
        "order=below"
        "width=6.0"
        "hidpi=on"
        "active_color=0xffC69BF6"
        "inactive_color=0x00000000"
      ];
      RunAtLoad = true;
      KeepAlive = true;
      ProcessType = "Interactive";
      LimitLoadToSessionType = "Aqua";
    };
  };
}
