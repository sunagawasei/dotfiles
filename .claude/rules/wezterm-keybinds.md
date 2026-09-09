---
paths:
  - "**/wezterm/keybinds.lua"
---

# WezTermキーバインド規約

## 追加前に衝突を確認する

WezTermのキーは、配下で動くTUI(herdr等)が使うキーを奪う。特に`CTRL|SHIFT`系は
herdrのcopy modeやタブ操作と重なりやすい。追加前に`herdr/config.toml`の`keys.*`と
`wezterm/keybinds.lua`の既存エントリを両方grepして、同じキーが無いことを確認する。

奪ってはいけないキーは、その場に理由を1行書いて空けておく(実例: `Ctrl+Shift+Y`は
herdrの`keys.copy_mode`が使うため、keybinds.lua側で意図的に未割り当てにしている)。
