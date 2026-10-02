{ config, lib, ... }:
# 設定を「誰に届けるか」で書き分けるオプション。各ツールは tools/<name>.nix にこれを書く。
# 値は単一引用符で書き出すため、$HOME などは展開されない(必要なら config.home.homeDirectory を使う)。
let
  inherit (lib) mkOption types;
  cfg = config.my;
  sh = lib.escapeShellArg;

  strs = mkOption {
    type = types.attrsOf types.str;
    default = { };
  };
  render =
    { env ? { }, aliases ? { }, init ? "" }:
    lib.concatLines (
      lib.mapAttrsToList (k: v: "export ${k}=${sh v}") env
      ++ lib.mapAttrsToList (k: v: "alias ${k}=${sh v}") aliases
      ++ lib.optional (init != "") init
    );
in
{
  options.my = {
    # .zshrc の is_human ガードより後で読む。人間だけに届く
    human = {
      aliases = strs;
      env = strs;
      init = mkOption {
        type = types.lines;
        default = "";
      };
    };
    # .zshenv の AI 側分岐で読む。AI だけに届く
    ai = {
      aliases = strs;
      env = strs;
    };
    # .zshenv の分岐より前で読む。人間にも AI にも届く
    env = strs;
    # ツール同梱の SKILL.md を Claude Code と Codex 等の両方へリンクする
    skills = mkOption {
      type = types.attrsOf types.path;
      default = { };
    };
  };

  config = {
    xdg.configFile."zsh/human.zsh".text = render cfg.human;
    xdg.configFile."zsh/ai.zsh".text = render cfg.ai;
    xdg.configFile."zsh/env.zsh".text = render { env = cfg.env; };

    home.file = lib.concatMapAttrs (name: path: {
      ".claude/skills/${name}".source = path;
      ".agents/skills/${name}".source = path;
    }) cfg.skills;
  };
}
