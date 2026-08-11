---
name: nvim-usage-audit
description: usage-tracker.jsonの実測でnvimの未使用キー・プラグインを棚卸しする手順
---

# nvim使用状況の棚卸し

`nvim/lua/config/usage-tracker.lua` が貯めた実測データ（`~/.local/share/nvim/usage-tracker.json`）から、未使用のキーバインド・プラグインを判定して撤去する。

## 鉄則

**0回の原因を「使っていない」と「そもそも到達できない」に必ず分ける。** 1回目（2026-08-11）はこれで配線バグを2件発見した。neotestは1ヶ月押しても起動できない状態で、octoの3キーは起動ごとに勝者が変わる衝突だった。数字だけで撤去すると原因を取り違える。

## ワークフロー

### 1. 解決済みowner表を取る

`references/audit.lua` を使う。**`Config.plugins[*].keys` を直接読んではいけない** — これはマージ前のfragmentで、snacks.nvimはraw 2件・resolved 58件になる。`plugin._.handlers.keys` を読むと、継承マージと `{ lhs, false }` 削除まで適用済みの表が得られる。

```bash
DUMP_OUT=/tmp/audit.json TRACKER=~/.local/share/nvim/usage-tracker.json \
  nvim --headless --cmd 'let g:usage_tracker_disable = 1' \
  <実ファイル> -c 'luafile <path>/references/audit.lua' -c 'qa!'
```

`references/nonspec.lua` は逆に「spec由来でない実マッピング」を洗い出す（LazyVim `config/keymaps.lua` 由来やプラグインが直接 `vim.keymap.set` するもの）。手順4の後半で使う。

**実ファイルを1つ開くこと。** 未オープンだと `LazyFile` イベントのプラグイン（mini.indentscope, gitsigns等）が載らず、実使用のあるキーを「到達不能」に誤分類する。

### 2. lhsraw基準で突き合わせる

トラッカーは `mode:display` で数え、owner表は lazy のIDを持つ。**表記文字列で突き合わせると誤判定する** — `<S-h>` と `H` は同一キー（`nvim_replace_termcodes` で確認できる）。必ず `lhsraw` をhex化して結合キーにする。

同様に `<C-/>` は0回に見えるが、端末はCtrl+/を `<C-_>` として送るため実体は使われている。**別名ペアは0回でも撤去してはいけない。**

### 3. 曖昧なキーは自動判定から外す

同一 `(mode, lhsraw)` に複数ownerがある、またはft限定のキーは、どの定義が実行されたか帰属できない。未使用リストから外して別枠で人間が判断する。

### 4. 撤去する

- **spec宣言キー** → `nvim/lua/plugins/zz-disabled-keys.lua` に `{ lhs, false }` を足す
- **spec由来でないキー**（LazyVim `config/keymaps.lua` 由来、プラグインが直接 `vim.keymap.set` するもの）→ `vim.keymap.del` が要る。specの `false` では消せない
- **自前定義** → 該当ファイルを直接編集

### 5. 検証は「消えたか」ではなく「遷移したか」で見る

自前定義を消すとLazyVim版が同じキーで復活する（`[e`/`]e`/`<leader>uh` が実例）。spec の false を足しても `config/keymaps.lua` 由来の定義は残る（`[b`/`]b` が実例）。**キーの有無ではなく、期待するowner・rhs・descへ遷移したかを合格条件にする。**

適用前後で owner表を取り、期待した削除件数とちょうど一致すること・想定外の増減がないことを確認する。

### 6. 削除台帳をコミットメッセージに残す

削除したキーは押しても記録されないため、次回の棚卸しでは検出できない。キー・最終使用日・削除理由・代替手段をコミットメッセージに残すのが唯一の記録になる。

## 落とし穴

- **headlessではVeryLazyが自然発火しない**。lazy.nvimの `very_lazy()` は `UIEnter` 待ちのため。`script -q /dev/null` でptyを与えれば自然発火する。手動で `nvim_exec_autocmds("User", {pattern="VeryLazy"})` しても実測ではキー482件が完全一致したので代替可
- **`--startuptime` は `--headless` より前に置く**。後ろだとログが書かれない
- **検証で起動するnvimには必ず `--cmd 'let g:usage_tracker_disable = 1'`**。付け忘れると計測データを汚染する
- **`vim.keymap.set` のラップで出自を取ろうとしない**。lazy.nvimの `core/handler/keys.lua` が `vim.keymap.set` を呼ぶ側なので、得られるsourceはspecファイルではなくハンドラになる
- 現行トラッカーの `build_registry` は全modeを1エントリに併合しdesc/scopeを持たない。レジストリを永続化する拡張をするなら、この構造から作り直しが要る

## 関連

- 経緯と残課題: メモリ `project_nvim_usage_tracker.md`
- lazy.nvimの罠: `nvim/CLAUDE.md` の「踏んだ罠」
- 無効化済みキー一覧: `nvim/lua/plugins/zz-disabled-keys.lua`
