{ pkgs, cursor-cli, ... }:
{
  home.packages =
    (with pkgs; [
      go
      nodejs_22
      pnpm
      python3
      uv # Python バージョン管理・パッケージ管理（pyenv 代替）
      jdk
    ])
    # nixpkgs 全体の pin を動かさずに cursor-agent だけ追従させるため pkgs 経由にしない
    ++ [ cursor-cli ];
}
