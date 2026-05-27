{ pkgs, ... }:
{
  home.packages = with pkgs; [
    go
    nodejs_22
    pnpm
    python3
    uv # Python バージョン管理・パッケージ管理（pyenv 代替）
    jdk
    cursor-cli # provides `cursor-agent` binary
  ];
}
