{ config, rift, ... }:
let
  user = config.home.username;
in
{
  home.packages = [ rift ];

  # 全操作を Hyper(Ctrl+Option+Cmd)+キーにする。wezterm/herdr/nvim/Claude Code はどれも使わない組み合わせ。
  # roBa は数字キーが NUM レイヤーにあるので、ワークスペースは右手の文字キー(U/I/O/P)に割り当てる。
  # [keys] は標準のキー割り当てを丸ごと置き換える。0.5.5 は [settings] が無いと起動時に panic する
  xdg.configFile."rift/config.toml".text = ''
    [settings]
    default_disable = false

    [settings.layout.stack]
    stack_offset = 0.0

    [modifier_combinations]
    hyper = "Ctrl + Alt + Meta"
    hyper_shift = "Ctrl + Alt + Meta + Shift"

    [keys]
    "hyper + T" = "toggle_space_activated"

    "hyper + H" = { move_focus = "left" }
    "hyper + J" = { move_focus = "down" }
    "hyper + K" = { move_focus = "up" }
    "hyper + L" = { move_focus = "right" }

    "hyper_shift + H" = { move_node = "left" }
    "hyper_shift + J" = { move_node = "down" }
    "hyper_shift + K" = { move_node = "up" }
    "hyper_shift + L" = { move_node = "right" }

    "hyper + U" = { switch_to_workspace = 0 }
    "hyper + I" = { switch_to_workspace = 1 }
    "hyper + O" = { switch_to_workspace = 2 }
    "hyper + P" = { switch_to_workspace = 3 }
    "hyper + N" = "switch_to_last_workspace"

    "hyper_shift + U" = { move_window_to_workspace = 0 }
    "hyper_shift + I" = { move_window_to_workspace = 1 }
    "hyper_shift + O" = { move_window_to_workspace = 2 }
    "hyper_shift + P" = { move_window_to_workspace = 3 }

    "hyper + F" = "toggle_fullscreen"
    "hyper + Space" = "toggle_window_floating"
    "hyper + Comma" = "toggle_stack"
    "hyper + Slash" = "toggle_orientation"
    "hyper + Equal" = "resize_window_grow"
    "hyper + Minus" = "resize_window_shrink"
  '';

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
