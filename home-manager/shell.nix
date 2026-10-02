{ pkgs, ... }:
{
  home.sessionVariables = {
    XDG_CONFIG_HOME = "$HOME/.config";
    XDG_CACHE_HOME = "$HOME/.cache";
    XDG_DATA_HOME = "$HOME/.local/share";
    XDG_STATE_HOME = "$HOME/.local/state";
    EDITOR = "vim";
    CVSEDITOR = "vim";
    SVN_EDITOR = "vim";
    GIT_EDITOR = "vim";
    VISUAL = "vim";
    LANGUAGE = "en_US.UTF-8";
    LANG = "en_US.UTF-8";
    LC_ALL = "en_US.UTF-8";
    LC_CTYPE = "en_US.UTF-8";
    LISTMAX = "50";
    ZENO_HOME = "$HOME/.config/zeno";
    ZENO_HISTORY_LIMIT = "5000";
    CLAUDE_CONFIG_DIR = "$HOME/.config/claude";
    CODEX_HOME = "$HOME/.config/codex";
    BASH_SILENCE_DEPRECATION_WARNING = "1";
  };

  home.sessionPath = [
    "$HOME/.local/bin"
    "/usr/local/sbin"
  ];
}
