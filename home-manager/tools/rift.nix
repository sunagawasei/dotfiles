{ config, rift, ... }:
let
  user = config.home.username;
in
{
  home.packages = [ rift ];

  # 全操作を Ctrl+Option+キーにする(roBa は左下段の LCTRL と LALT で押せる)。wezterm/herdr/nvim/Claude Code はどれも使わない組み合わせ。
  # 名前が hyper なのは [modifier_combinations] の定義名で、Cmd は含まない。
  # roBa は数字キーが NUM レイヤーにあるので、ワークスペースは右手の文字キー(U/I/O/P)に割り当てる。
  # [keys] は標準のキー割り当てを丸ごと置き換える。0.5.5 は [settings] が無いと起動時に panic する
  xdg.configFile."rift/config.toml".text = ''
    [settings]
    default_disable = false

    [settings.ui.menu_bar]
    enabled = true

    [settings.ui.mission_control]
    enabled = true

    # フォーカス枠(JankyBorders, 太さ6)は窓の外側に描かれるので、隙間を枠と同じ6にして4辺の太さを揃える
    [settings.layout.gaps.outer]
    top = 6
    left = 6
    bottom = 6
    right = 6

    [settings.layout.gaps.inner]
    horizontal = 6
    vertical = 6

    # 既定の true だと traditional から scrolling へ切り替えても各窓の幅を保ち、見た目が変わらない
    [settings.layout.scrolling]
    preserve_window_sizes = false

    [modifier_combinations]
    hyper = "Ctrl + Alt"
    hyper_shift = "Ctrl + Alt + Shift"

    # Main=ブラウザ・読み物・AIチャット / Development=WezTermだけ / Communication=Slack・Zoom・カレンダー / Utilities=設定・VPN・ツール類
    [[virtual_workspaces.app_rules]]
    app_id = "com.google.Chrome"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.Safari"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "notion.id"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.linear"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.anthropic.claudefordesktop"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.github.wez.wezterm"
    workspace = "Development"

    [[virtual_workspaces.app_rules]]
    app_id = "com.tinyspeck.slackmacgap"
    workspace = "Communication"

    [[virtual_workspaces.app_rules]]
    app_id = "us.zoom.xos"
    workspace = "Communication"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.systempreferences"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "cc.ffitch.shottr"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "org.keepassxc.keepassxc"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.ActivityMonitor"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.microsoft.rdc.macos"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "org.pqrs.Karabiner-EventViewer"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "org.pqrs.Karabiner-Elements.Settings"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.Passwords"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.AppStore"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.cisco.secureclient.gui"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.fortinet.FortiClient"
    workspace = "Utilities"

    [[virtual_workspaces.app_rules]]
    app_id = "com.openai.codex"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.amazon.Lassen"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.iBooksX"
    workspace = "Main"

    [[virtual_workspaces.app_rules]]
    app_id = "com.apple.iCal"
    workspace = "Communication"

    # Prompt Line は全面に出すオーバーレイなので、Rift の配置対象から外す
    [[virtual_workspaces.app_rules]]
    app_id = "com.electron.prompt-line"
    manage = false

    [keys]
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

    "hyper + Period" = "next_window"
    "hyper_shift + Period" = "prev_window"

    "hyper + Y" = "show_mission_control_all"

    "hyper + M" = "toggle_fullscreen"
    "hyper + Space" = "toggle_window_floating"
    "hyper + T" = { set_workspace_layout = { mode = "traditional" } }
    "hyper + S" = { set_workspace_layout = { mode = "stack" } }
    "hyper + R" = { set_workspace_layout = { mode = "scrolling" } }

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
