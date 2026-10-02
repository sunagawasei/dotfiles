{ pkgs, ... }:
{
  home.packages = [ pkgs.gomi ];
  # AI も rm を打つ。rm 互換オプションを受けるので、誤削除をゴミ箱から戻せる(gomi -b)ようにする
  my.human.aliases.rm = "gomi";
  my.ai.aliases.rm = "gomi";
}
