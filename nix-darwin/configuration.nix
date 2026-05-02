{
  self,
  lib,
  ...
}:
{
  nixpkgs.hostPlatform = "aarch64-darwin";
  system.stateVersion = 6;
  nix.enable = false; # Nix daemon は外部インストーラー（Determinate Systems）で管理
  system.configurationRevision = self.rev or self.dirtyRev or null;
  security.pam.services.sudo_local.touchIdAuth = true;

  programs.zsh.shellInit = ''
    export ZDOTDIR="$HOME/.config/zsh"
  '';
  programs.zsh.enableGlobalCompInit = false;
  users.users."s23159".home = "/Users/s23159";
  system.primaryUser = "s23159";
  nixpkgs.config.allowUnfreePredicate =
    pkg:
    builtins.elem (lib.getName pkg) [
      "terraform"
    ];

  # ── NSGlobalDomain（外観・入力・単位） ─────────────────────────────
  system.defaults.NSGlobalDomain.AppleShowAllExtensions = true;
  system.defaults.NSGlobalDomain.AppleInterfaceStyleSwitchesAutomatically = true;
  system.defaults.NSGlobalDomain.AppleKeyboardUIMode = 2;
  system.defaults.NSGlobalDomain.AppleMeasurementUnits = "Centimeters";
  system.defaults.NSGlobalDomain.AppleMetricUnits = 1;
  system.defaults.NSGlobalDomain.AppleTemperatureUnit = "Celsius";
  system.defaults.NSGlobalDomain.AppleSpacesSwitchOnActivate = false;
  system.defaults.NSGlobalDomain.NSAutomaticCapitalizationEnabled = false;
  system.defaults.NSGlobalDomain.NSAutomaticPeriodSubstitutionEnabled = false;
  system.defaults.NSGlobalDomain.NSAutomaticSpellingCorrectionEnabled = false;
  system.defaults.NSGlobalDomain.NSTableViewDefaultSizeMode = 2;
  system.defaults.NSGlobalDomain."com.apple.sound.beep.feedback" = 1;
  system.defaults.NSGlobalDomain."com.apple.trackpad.scaling" = 2.0;
  system.defaults.NSGlobalDomain."com.apple.trackpad.forceClick" = true;
  system.defaults.NSGlobalDomain."com.apple.springing.enabled" = true;
  system.defaults.NSGlobalDomain._HIHideMenuBar = true;
  system.defaults.CustomUserPreferences.NSGlobalDomain.AppleActionOnDoubleClick = "None";

  # ── Trackpad ────────────────────────────────────────────────────────
  system.defaults.trackpad.Clicking = true;
  system.defaults.trackpad.TrackpadRightClick = true;
  system.defaults.trackpad.FirstClickThreshold = 0;
  system.defaults.trackpad.SecondClickThreshold = 0;
  system.defaults.NSGlobalDomain."com.apple.mouse.tapBehavior" = 1;
  system.defaults.NSGlobalDomain.KeyRepeat = 2;
  system.defaults.NSGlobalDomain.InitialKeyRepeat = 15;
  system.defaults.NSGlobalDomain.ApplePressAndHoldEnabled = false;

  # ── Dock ────────────────────────────────────────────────────────────
  system.defaults.dock.autohide = true;
  system.defaults.dock.orientation = "left";
  system.defaults.dock.tilesize = 27;
  system.defaults.dock.magnification = false;
  system.defaults.dock.minimize-to-application = false;
  system.defaults.dock.mru-spaces = false;
  system.defaults.dock.show-recents = false;
  system.defaults.dock.expose-group-apps = false;
  system.defaults.dock.wvous-br-corner = 1;
  system.defaults.dock.persistent-apps = [
    "/System/Applications/Apps.app"
    "/System/Applications/Calendar.app"
    "/System/Applications/System Settings.app"
    "/Applications/FortiClient.app"
    "/Applications/Slack.app"
  ];

  # ── Finder ──────────────────────────────────────────────────────────
  system.defaults.finder.AppleShowAllFiles = true;
  system.defaults.finder.ShowPathbar = true;
  system.defaults.finder.FXPreferredViewStyle = "Nlsv";
  system.defaults.finder.FXRemoveOldTrashItems = true;
  system.defaults.finder.NewWindowTarget = "Recents";
  system.defaults.finder.ShowHardDrivesOnDesktop = false;
  system.defaults.finder.ShowExternalHardDrivesOnDesktop = true;
  system.defaults.finder.ShowRemovableMediaOnDesktop = true;

  # ── WindowManager（Stage Manager） ──────────────────────────────────
  system.defaults.WindowManager.GloballyEnabled = false;
  system.defaults.WindowManager.AutoHide = false;
  system.defaults.WindowManager.EnableStandardClickToShowDesktop = false;
  system.defaults.WindowManager.EnableTiledWindowMargins = false;
  system.defaults.WindowManager.EnableTopTilingByEdgeDrag = true;
  system.defaults.WindowManager.AppWindowGroupingBehavior = true;
  system.defaults.WindowManager.HideDesktop = false;
  system.defaults.WindowManager.StageManagerHideWidgets = true;
  system.defaults.WindowManager.StandardHideDesktopIcons = true;
  system.defaults.WindowManager.StandardHideWidgets = true;

  # ── Control Center（メニューバー表示） ──────────────────────────────
  system.defaults.controlcenter.BatteryShowPercentage = true;
  system.defaults.controlcenter.Bluetooth = true;
  system.defaults.controlcenter.FocusModes = false;
  system.defaults.controlcenter.Sound = true;

  # ── メニューバー時計 ─────────────────────────────────────────────────
  system.defaults.menuExtraClock.Show24Hour = true;
  system.defaults.menuExtraClock.ShowDayOfWeek = true;
  system.defaults.menuExtraClock.ShowDate = 0;

  # ── カスタムキーボードショートカット ────────────────────────────────
  system.defaults.CustomUserPreferences.NSGlobalDomain.NSUserKeyEquivalents."タブを複製" = "^t";

  # ── 日本語IME（Kotoeri）────────────────────────────────────────────
  system.defaults.CustomUserPreferences."com.apple.inputmethod.Kotoeri" = {
    JIMPrefLiveConversionKey = 0;
    JIMPrefPredictiveCandidateKey = 0;
    JIMPrefAutocorrectionKey = 0;
    JIMPrefCapsLockActionKey = 0;
    JIMPrefCharacterForSlashKey = 1;
    JIMPrefCharacterForYenKey = 1;
    JIMPrefShiftKeyActionKey = 1;
  };

  # ── Safari ──────────────────────────────────────────────────────────
  system.defaults.CustomUserPreferences."com.apple.Safari".ShowFullURLInSmartSearchField = 1;

  # ── TextEdit ─────────────────────────────────────────────────────────
  system.defaults.CustomUserPreferences."com.apple.TextEdit".RichText = 0;

  # ── iCal ────────────────────────────────────────────────────────────
  system.defaults.iCal."first day of week" = "Sunday";
  system.defaults.iCal.CalendarSidebarShown = false;

  # ── スクリーンショット ───────────────────────────────────────────────
  system.defaults.screencapture.location = "/Users/s23159/.Trash";

  # ── 電源管理（ディスプレイスリープ） ─────────────────────────────────
  system.activationScripts.pmset.text = ''
    /usr/bin/pmset -c displaysleep 30
    /usr/bin/pmset -c sleep 0
    /usr/bin/pmset -c ttyskeepawake 1
    /usr/bin/pmset -b displaysleep 30
    /usr/bin/pmset -b lessbright 0
    /usr/bin/pmset -b ttyskeepawake 1
  '';

  # ── キーボードショートカット（Spotlight無効化） ───────────────────────
  system.defaults.CustomUserPreferences."com.apple.symbolichotkeys".AppleSymbolicHotKeys = {
    "64" = {
      enabled = false;
      value = {
        parameters = [ 262144 4294705151 ];
        type = "modifier";
      };
    };
    "65" = {
      enabled = false;
      value = {
        parameters = [ 65535 49 1572864 ];
        type = "standard";
      };
    };
  };

  system.activationScripts.activateHotkeys.text = ''
    /System/Library/PrivateFrameworks/SystemAdministration.framework/Resources/activateSettings -u
  '';

  imports = [
    ./home_manager.nix
    ./homebrew.nix
  ];
}
