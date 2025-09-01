# Neovim キーマップ一覧

このドキュメントは、Neovim設定で定義されているカスタムキーマップをまとめたものです。

## 基本操作

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<C-s>` | ファイルを保存 | n, i, v |
| `<C-c>` | ファイル全体をクリップボードにコピー | n |

## ファイルエクスプローラー

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>e` | エクスプローラーを開く/フォーカス/閉じる（ルートディレクトリ） | n |
| `<leader>E` | エクスプローラーを開く/フォーカス/閉じる（現在のディレクトリ） | n |
| `<leader>fe` | ファイルエクスプローラー（LazyVimデフォルト） | n |

## コードフォーマット

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>fm` | ファイルまたは選択範囲をフォーマット | n, v |
| `<leader>cf` | コードフォーマット（LazyVimデフォルト） | n, v |

### 自動フォーマットの一時的な無効化

YAMLファイルなどで保存時の自動フォーマットを一時的に無効化：

```vim
" 現在のバッファのみ無効化（YAMLファイルで特に便利）
:lua vim.b.autoformat = false

" 現在のバッファのみ再度有効化
:lua vim.b.autoformat = true

" グローバルに無効化（すべてのファイルで無効）
:lua vim.g.autoformat = false

" グローバルに再度有効化
:lua vim.g.autoformat = true

" コマンドで切り替え（トグル）
:FormatToggle   " 現在のバッファの設定を切り替え
:FormatToggle!  " グローバル設定を切り替え
```

**注意**: 
- Neovimを再起動すると設定はリセットされます（デフォルト: `vim.g.autoformat = true`）
- バッファローカル設定（`vim.b.autoformat`）はファイルを開き直すとリセットされます
- グローバル設定が`false`の場合、バッファローカル設定は無視されます

## 診断（Diagnostics）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `gl` または `<leader>ld` | 行の診断を表示 | n |
| `<leader>cd` | 診断を表示（LazyVimデフォルト） | n |
| `[d` | 前の診断へ移動 | n |
| `]d` | 次の診断へ移動 | n |
| `[e` | 前のエラーへ移動 | n |
| `]e` | 次のエラーへ移動 | n |
| `<leader>dl` | 診断をロケーションリストに追加 | n |
| `<leader>dq` | 診断をQuickfixリストに追加 | n |

### 診断フローティングウィンドウの操作

診断を表示した後のフローティングウィンドウ操作：

| キーマップ | 説明 | 使用場面 |
|-----------|------|---------|
| `<C-w>w` または `<C-w><C-w>` | フローティングウィンドウにフォーカス | ウィンドウが開いている時 |
| `q` または `<Esc>` | フローティングウィンドウを閉じる | フォーカス後 |
| `<C-w>q` | フローティングウィンドウを閉じる | フォーカス後 |
| スクロールキー | 長い診断メッセージをスクロール | フォーカス後 |

## LSP ナビゲーション（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `gd` | 定義へジャンプ | n |
| `gr` | 参照箇所を検索 | n |
| `gI` | 実装へジャンプ | n |
| `gy` | 型定義へジャンプ | n |
| `gD` | 宣言へジャンプ | n |
| `K` | ホバー情報表示（定義・ドキュメント） | n |
| `gK` | シグネチャヘルプ | n |
| `<leader>ca` | コードアクション | n, v |
| `<leader>cr` | リネーム | n |
| `<leader>cl` | LSP情報表示 | n |

## テスト実行（Neotest）

### 基本操作
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>tr` | カーソル位置のテストを実行（最も近いテスト） | n |
| `<leader>tt` | 現在のファイルのすべてのテストを実行 | n |
| `<leader>tT` | プロジェクト全体のテストを実行 | n |
| `<leader>tl` | 最後に実行したテストを再実行 | n |
| `<leader>td` | カーソル位置のテストをデバッグ実行 | n |

### 拡張操作
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>tp` | パッケージ/ディレクトリ単位でテスト実行 | n |
| `<leader>tc` | 最寄りのテストをカバレッジ測定付きで実行 | n |
| `<leader>tC` | 現在ファイルをカバレッジ測定付きで実行 | n |
| `<leader>tA` | プロジェクト全体のテストをレース検出付きで実行 | n |

### 結果表示
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>ts` | テスト結果サマリーを表示/非表示 | n |
| `<leader>to` | テスト出力を表示 | n |
| `<leader>tO` | テスト出力パネルを切り替え | n |
| `<leader>tS` | 実行中のテストを停止 | n |
| `<leader>tw` | テストウォッチモード切り替え | n |

**サポート言語**: 現在Go言語のテストに対応（サブテスト、テーブルテスト、カバレッジ測定、レース検出）

## ターミナル

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<C-\>` | ターミナルをトグル | n, t |
| `<leader>tf` | フローティングターミナル | n |
| `<leader>th` | 水平分割ターミナル | n |
| `<leader>tv` | 垂直分割ターミナル | n |
| `<leader>tg` | LazyGitを開く（要: `brew install lazygit`） | n |
| `<leader>tu` | GitUIを開く（要: `brew install gitui`） | n |
| `<leader>tb` | Btopを開く（要: `brew install btop`） | n |

## Git

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>gl` | Gitグラフを表示 | n |
| `<leader>gg` | LazyGit（LazyVimデフォルト） | n |
| `<leader>gG` | LazyGit（現在のファイル履歴） | n |
| `<leader>gb` | Git blame | n |
| `<leader>gB` | Git browse | n |

## AI アシスタント

### GitHub Copilot
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<Tab>` | Copilotの提案を受け入れ | i |
| `<leader>cc` | CopilotChatを開く | n |
| `<leader>ct` | CopilotChatをトグル | n |
| `<leader>cs` | CopilotChatを停止 | n |
| `<leader>cr` | CopilotChatをリセット | n |
| `<leader>cE` | コードを説明（選択範囲） | v |
| `<leader>cR` | コードレビュー（選択範囲） | v |
| `<leader>cF` | 修正提案（選択範囲） | v |
| `<leader>cO` | 最適化提案（選択範囲） | v |
| `<leader>cD` | ドキュメント生成（選択範囲） | v |
| `<leader>cT` | テストケース生成（選択範囲） | v |

### Claude Code
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>ac` | Claude Codeをトグル | n |
| `<leader>af` | Claude Codeにフォーカス | n |
| `<leader>ar` | 前回のセッションを再開 | n |
| `<leader>aC` | セッションを継続 | n |

## Markdown

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>mp` | Markdownプレビューをトグル | n |
| `<leader>ms` | Markdownプレビューを停止 | n |

## ウィンドウ管理

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>wp` | ウィンドウピッカーで移動 | n |
| `<leader>nn` | No Neck Pain（中央寄せ表示）をトグル | n |
| `<C-h>` | 左のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-j>` | 下のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-k>` | 上のウィンドウへ移動（LazyVimデフォルト） | n |
| `<C-l>` | 右のウィンドウへ移動（LazyVimデフォルト） | n |
| `<leader>w-` | 水平分割（LazyVimデフォルト） | n |
| `<leader>w\|` | 垂直分割（LazyVimデフォルト） | n |
| `<leader>wd` | ウィンドウを削除（LazyVimデフォルト） | n |

## 検索・置換（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>ff` | ファイル検索 | n |
| `<leader>fg` | Grep検索 | n |
| `<leader>fb` | バッファ検索 | n |
| `<leader>fh` | ヘルプタグ検索 | n |
| `<leader>fc` | 設定ファイル検索 | n |
| `<leader>fk` | キーマップ検索 | n |
| `<leader>fr` | 最近開いたファイル | n |
| `<leader>sr` | 検索と置換（Grug-far） | n, v |

## バッファ操作（LazyVimデフォルト）

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<S-h>` | 前のバッファ | n |
| `<S-l>` | 次のバッファ | n |
| `[b` | 前のバッファ | n |
| `]b` | 次のバッファ | n |
| `<leader>bb` | バッファを切り替え | n |
| `<leader>bd` | バッファを削除 | n |
| `<leader>bD` | バッファを強制削除 | n |

## その他のユーティリティ

| キーマップ | 説明 | モード |
|-----------|------|--------|
| `<leader>l` | Lazy（プラグインマネージャー） | n |
| `<leader>L` | LazyVim Changelog | n |
| `<leader>qq` | 終了 | n |
| `<leader>ui` | UIのインデントガイドをトグル | n |
| `<leader>ul` | 行番号をトグル | n |
| `<leader>uL` | 相対行番号をトグル | n |
| `<leader>uw` | 折り返しをトグル | n |
| `<leader>uh` | インラインヒント（型情報など）をトグル | n |

## ファイルタイプの変更

```vim
" 現在のファイルタイプを確認
:set filetype?

" ファイルタイプを設定（例：Markdownに変更）
:set filetype=markdown

" ファイルタイプを設定（短縮形）
:set ft=javascript

" よく使うファイルタイプの例
:set ft=python       " Python
:set ft=javascript   " JavaScript
:set ft=typescript   " TypeScript
:set ft=json        " JSON
:set ft=yaml        " YAML
:set ft=html        " HTML
:set ft=css         " CSS
:set ft=sh          " Shell Script
:set ft=vim         " Vim Script
:set ft=lua         " Lua
:set ft=rust        " Rust
:set ft=go          " Go
:set ft=java        " Java
:set ft=cpp         " C++
:set ft=c           " C
:set ft=ruby        " Ruby
:set ft=php         " PHP
:set ft=sql         " SQL
:set ft=toml        " TOML
:set ft=xml         " XML
:set ft=dockerfile  " Dockerfile
:set ft=gitcommit   " Git Commit
:set ft=conf        " 設定ファイル
:set ft=text        " プレーンテキスト
```

**注意**: 
- ファイルタイプを変更すると、シンタックスハイライト、インデント、LSP、フォーマッターなどが切り替わります
- 一時的な変更で、ファイルを開き直すと元に戻ります

## Zsh キーバインド

| キーマップ | 説明 |
|-----------|------|
| `Ctrl+E` | 現在のコマンドラインをエディタ（nvim）で編集 |
| `Ctrl+P` | 履歴を前方検索（入力済み文字列にマッチ） |
| `Ctrl+N` | 履歴を後方検索（入力済み文字列にマッチ） |
| `Ctrl+Y` | 現在の単語だけを適用して次の補完位置へ |
| `Ctrl+]` | 現在の補完を確定して続けて入力可能に |

## WezTerm キーバインド

### ペイン操作
| キーマップ | 説明 |
|-----------|------|
| `Ctrl+q → h/l/k/j` | ペインフォーカスを移動（左/右/上/下） |
| `Ctrl+q → d` | ペインを垂直分割 |
| `Ctrl+q → r` | ペインを水平分割 |
| `Ctrl+q → x` | ペインを閉じる |
| `Ctrl+q → z` | ペインをズーム（最大化/元に戻す） |
| `Ctrl+q → R` | ペインを時計回りに回転（入れ替え） |
| `Ctrl+q → Shift+R` | ペインを反時計回りに回転（入れ替え） |
| `Ctrl+Shift+H/L/K/J` | ペインフォーカスを即座に移動（左/右/上/下） |
| `Ctrl+Shift+[` | ペイン選択モード |

### ペインナビゲーションモード（`Ctrl+q → q`）
| キーマップ | 説明 |
|-----------|------|
| `h/l/k/j` | ペインフォーカスを移動（左/右/上/下） |
| `1-9` | ペインを番号で直接選択 |
| `Shift+H/L/K/J` | ペインサイズを調整（5単位） |
| `x` | ペインを閉じる |
| `z` | ペインをズーム |
| `r` | ペインを時計回りに回転 |
| `Shift+R` | ペインを反時計回りに回転 |
| `Enter/Esc/q` | モードを終了 |

### タブ操作
| キーマップ | 説明 |
|-----------|------|
| `Cmd+t` | 新しいタブを開く |
| `Cmd+w` | タブを閉じる |
| `Cmd+Shift+]` | 次のタブへ移動 |
| `Cmd+Shift+[` | 前のタブへ移動 |
| `Cmd+1-9` | タブを番号で選択 |
| `Ctrl+q → {` | タブを左に移動 |
| `Ctrl+q → }` | タブを右に移動 |

### その他
| キーマップ | 説明 |
|-----------|------|
| `Cmd+c` | コピー |
| `Cmd+v` | ペースト |
| `Ctrl+q → [` | コピーモード（viキーバインド） |
| `Ctrl+Shift+Space` | Quick Selectモード |
| `Cmd+p` | コマンドパレット |
| `Ctrl+Shift+R` | 設定再読み込み |

### フォントサイズ
| キーマップ | 説明 |
|-----------|------|
| `Ctrl++` | フォントサイズを大きくする |
| `Ctrl+-` | フォントサイズを小さくする |
| `Ctrl+0` | フォントサイズをリセット |

**注**: 
- リーダーキー（`Ctrl+q`）のタイムアウトは1000ms
- ペインの回転は、2つのペインなら左右入れ替え、3つ以上なら順序の回転

## Lazygit キーバインド

### ブランチビュー
| キーマップ | 説明 |
|-----------|------|
| `]` | 次のタブへ（ローカル→リモート→タグ） |
| `[` | 前のタブへ（タグ→リモート→ローカル） |

リモートブランチを表示してからチェックアウトすることで、ローカルにないブランチへの切り替えが可能。

### 全ブランチグラフの表示
| キーマップ | 説明 |
|-----------|------|
| `a` | Statusパネル（1番）で全ブランチのログをグラフ表示 |
| `+` を2回 | Commitsパネル（4番）でパネルを最大化してグラフ表示 |
| `Ctrl+O` | Commitsパネルでビューログオプションメニューを開く（Ctrl+Lから変更） |

**注**: `showWholeGraph: true`の設定により、Commitsパネルはデフォルトで全ブランチを表示します（VSCodeのGit Graphのような表示）。

## Serena (コーディングエージェント)

### Claude Codeへの追加
```bash
# プロジェクトディレクトリで実行
claude mcp add serena -s project -- uvx --from git+https://github.com/oraios/serena serena start-mcp-server --context ide-assistant --project $(pwd)
```

### 使用方法

#### オンボーディング
1. Claude Codeを起動
2. `/mcp`コマンドでSerenaサーバーの確認
3. 以下のプロンプトでオンボーディングを開始：
   ```
   Serena のオンボーディングを開始してください。
   ```

### 参考リンク
- [Serena GitHub](https://github.com/oraios/serena)
- [Serena 解説記事](https://azukiazusa.dev/blog/serena-coding-agent/)

## 文字ジャンプ（f/F/t/T コマンド）

### 基本操作
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `f{char}` | 行内で指定文字を前方検索してジャンプ | n |
| `F{char}` | 行内で指定文字を後方検索してジャンプ | n |
| `t{char}` | 行内で指定文字の直前にジャンプ | n |
| `T{char}` | 行内で指定文字の直後にジャンプ | n |

### 検索の繰り返し
| キーマップ | 説明 | モード |
|-----------|------|--------|
| `;` | 最後の文字検索を前方向に繰り返し | n |
| `,` | 最後の文字検索を逆方向に繰り返し | n |

**重要**: Escapeで中断しても`;`と`,`は最後の検索を記憶しているため、引き続き同じ文字へのジャンプが可能。

## ノート

- `<leader>` キーはデフォルトでスペースキーに設定されています
- `n` = Normal mode, `i` = Insert mode, `v` = Visual mode, `t` = Terminal mode
- LazyVimのデフォルトキーマップの詳細は[公式ドキュメント](https://www.lazyvim.org/keymaps)を参照してください
