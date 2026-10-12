{ config, rift, ... }:
let
  user = config.home.username;
in
{
  home.packages = [ rift ];

  # `rift service install` が書く plist と同じ内容。Nice=-20 は root でないと効かないので除く。
  # 二重登録になるので `rift service install` は実行しない
  launchd.agents.rift = {
    enable = true;
    config = {
      Label = "git.acsandmann.rift";
      ProgramArguments = [ "${rift}/bin/rift" ];
      EnvironmentVariables = {
        PATH = "${config.home.profileDirectory}/bin:/run/current-system/sw/bin:/usr/bin:/bin";
        RUST_LOG = "error,warn,info";
      };
      RunAtLoad = true;
      KeepAlive = {
        SuccessfulExit = false;
        Crashed = true;
      };
      StandardOutPath = "/tmp/rift_${user}.out.log";
      StandardErrorPath = "/tmp/rift_${user}.err.log";
      ProcessType = "Interactive";
      LimitLoadToSessionType = "Aqua";
    };
  };
}
