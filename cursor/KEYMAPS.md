# Cursor CLI キーマップ

Cursor CLIのキーバインド・ショートカットのリファレンスです。

## Vimモード

Cursor CLIはVimモードが有効なため、標準的なVimキーバインドが使用できます。

### 基本操作

| キー | モード | 動作 |
|------|--------|------|
| `i` | Normal | Insert モードに切り替え |
| `a` | Normal | カーソルの次の文字から Insert モード |
| `A` | Normal | 行末から Insert モード |
| `o` | Normal | 下に新しい行を作成して Insert モード |
| `O` | Normal | 上に新しい行を作成して Insert モード |
| `ESC` | Insert/Visual | Normal モードに戻る |
| `Ctrl+[` | Insert/Visual | Normal モードに戻る（ESCの代替） |

### カーソル移動

| キー | モード | 動作 |
|------|--------|------|
| `h` | Normal | 左に移動 |
| `j` | Normal | 下に移動 |
| `k` | Normal | 上に移動 |
| `l` | Normal | 右に移動 |
| `w` | Normal | 次の単語の先頭に移動 |
| `b` | Normal | 前の単語の先頭に移動 |
| `e` | Normal | 単語の末尾に移動 |
| `0` | Normal | 行頭に移動 |
| `^` | Normal | 行の最初の非空白文字に移動 |
| `$` | Normal | 行末に移動 |
| `gg` | Normal | ファイルの先頭に移動 |
| `G` | Normal | ファイルの末尾に移動 |
| `{` | Normal | 前の段落に移動 |
| `}` | Normal | 次の段落に移動 |
| `Ctrl+d` | Normal | 半画面下にスクロール |
| `Ctrl+u` | Normal | 半画面上にスクロール |
| `Ctrl+f` | Normal | 1画面下にスクロール |
| `Ctrl+b` | Normal | 1画面上にスクロール |

### 編集操作

| キー | モード | 動作 |
|------|--------|------|
| `x` | Normal | カーソル下の文字を削除 |
| `dd` | Normal | 行を削除 |
| `dw` | Normal | 単語を削除 |
| `d$` | Normal | カーソル位置から行末まで削除 |
| `D` | Normal | カーソル位置から行末まで削除（`d$`と同じ） |
| `yy` | Normal | 行をヤンク（コピー） |
| `yw` | Normal | 単語をヤンク |
| `p` | Normal | カーソルの後にペースト |
| `P` | Normal | カーソルの前にペースト |
| `u` | Normal | アンドゥ |
| `Ctrl+r` | Normal | リドゥ |
| `.` | Normal | 直前の変更を繰り返す |
| `>>` | Normal | 行をインデント |
| `<<` | Normal | 行のインデントを解除 |

### Visual モード

| キー | モード | 動作 |
|------|--------|------|
| `v` | Normal | Visual モードに切り替え（文字選択） |
| `V` | Normal | Visual Line モードに切り替え（行選択） |
| `Ctrl+v` | Normal | Visual Block モードに切り替え（矩形選択） |
| `d` | Visual | 選択範囲を削除 |
| `y` | Visual | 選択範囲をヤンク（コピー） |
| `>` | Visual | 選択範囲をインデント |
| `<` | Visual | 選択範囲のインデントを解除 |

### 検索・置換

| キー | モード | 動作 |
|------|--------|------|
| `/pattern` | Normal | 前方検索 |
| `?pattern` | Normal | 後方検索 |
| `n` | Normal | 次の検索結果に移動 |
| `N` | Normal | 前の検索結果に移動 |
| `*` | Normal | カーソル下の単語を前方検索 |
| `#` | Normal | カーソル下の単語を後方検索 |
| `:s/old/new/` | Normal | 現在行で最初に出現するoldをnewに置換 |
| `:s/old/new/g` | Normal | 現在行のすべてのoldをnewに置換 |
| `:%s/old/new/g` | Normal | ファイル全体のすべてのoldをnewに置換 |

### ファイル操作

| キー | モード | 動作 |
|------|--------|------|
| `:w` | Normal | 保存 |
| `:q` | Normal | 終了 |
| `:wq` | Normal | 保存して終了 |
| `:q!` | Normal | 保存せずに強制終了 |
| `ZZ` | Normal | 保存して終了（`:wq`と同じ） |
| `ZQ` | Normal | 保存せずに終了（`:q!`と同じ） |

## Cursor CLI固有の機能

Cursor CLIはAI統合コードエディタとして、以下の独自機能を提供します。

### AIアシスタント

| キー | モード | 動作 |
|------|--------|------|
| `Ctrl+K` | Normal/Insert | AIコマンドパレットを開く |
| `Ctrl+L` | Normal/Insert | チャットパネルを開く |
| `Tab` | Insert | AI補完を受け入れる |
| `Esc` | Insert | AI補完を拒否 |

### コード操作

| キー | モード | 動作 |
|------|--------|------|
| `Ctrl+/` | Normal/Insert | 行コメントのトグル |
| `Ctrl+Space` | Insert | 手動で補完をトリガー |

## WezTermとの統合

Cursor CLIはWezTerm上で動作するため、WezTermのキーバインドも使用できます。

### ペイン操作（WezTerm）

| キー | 動作 |
|------|------|
| `Ctrl+Q H` | 左のペインに移動 |
| `Ctrl+Q J` | 下のペインに移動 |
| `Ctrl+Q K` | 上のペインに移動 |
| `Ctrl+Q L` | 右のペインに移動 |
| `Ctrl+Q X` | 現在のペインを閉じる |

詳細なWezTermキーバインドは `/Users/s23159/.config/wezterm/KEYMAPS.md` を参照してください。

## カスタマイズ

Cursor CLIの設定は `cli-config.json` で管理されています。
現在の設定では以下が有効化されています：

- **Vimモード**: 有効
- **行番号表示**: 有効
- **AIモデル**: GPT-5.2 Codex Extra High
- **maxMode**: 有効（最大コンテキスト長）

## 関連ドキュメント

- **README**: `/Users/s23159/.config/cursor/README.md`
- **WezTerm KEYMAPS**: `/Users/s23159/.config/wezterm/KEYMAPS.md`
- **Neovim KEYMAPS**: `/Users/s23159/.config/nvim/KEYMAPS.md`
