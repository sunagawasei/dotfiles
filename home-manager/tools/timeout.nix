{ pkgs, ... }:
{
  # macOS に無い timeout だけを公開する。coreutils 全体を PATH に入れると BSD 前提の書き方が壊れる
  home.packages = [ (pkgs.writeShellScriptBin "timeout" "exec ${pkgs.coreutils}/bin/timeout \"$@\"") ];
}
