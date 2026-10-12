{ rift, ... }:
{
  # 常駐は `rift service install` が launchd へ登録する
  home.packages = [ rift ];
}
