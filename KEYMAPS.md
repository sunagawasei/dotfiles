# Neovim

## LSP定義ジャンプ

- gd: 定義へジャンプ（カレントバッファ）
- gvd: 定義へジャンプ（垂直split）
- ghd: 定義へジャンプ（水平split）

## 診断全般のジャンプ（warning/error/info含む）

- ]d: 次の診断へ移動
- [d: 前の診断へ移動

## コメントアウト

`gcc`

## 折り畳み開閉

- za: 折り畳みの開閉（トグル）
- zM: すべての折り畳みを閉じる
- zR: すべての折り畳みを開く

## 囲み文字

1. 囲み文字を追加 (gsa)

- gsaiw" - カーソル位置の単語を""で囲む
- v (ビジュアルモード) → 選択 → gsa"

2. 囲み文字を削除 (gsd)
   操作: gsd"
3. 囲み文字を置換 (gsr)
   操作: gsr"'

## GitHubのリンクをコピーするキーバインド

- <leader>yg：ファイルのGitHub URLをコピー（ノーマルモード, oil）
- <leader>yl：行リンクをコピー（ノーマル/ビジュアル。選択範囲があると範囲リンク）

## gitのコミットや編集ユーザーなどの詳細を表示する

- <leader>hp: Hunkプレビュー（差分確認）
- <leader>hb: Blame詳細表示（コミット情報のポップアップ）

## page up/down

ctrlを押しながらfで1ページ進みます。forwardのfです。ctrlを押しながらbで 1ページ戻ります。

## 括弧

- ib/ab ()括弧
- iB/aB {}括弧

## ターミナル操作

### ウィンドウ間移動

- Ctrl+h: 左のウィンドウへ移動
- Ctrl+j: 下のウィンドウへ移動
- Ctrl+k: 上のウィンドウへ移動
- Ctrl+l: 右のウィンドウへ移動

※ ノーマルモード・ターミナルモード両方で使用可能

### toggletrm

基本操作:

- <leader>t1 / <leader>t2 / <leader>t3: 番号付きターミナル
- <C-/>: 最後のターミナルをトグル
- <leader>ta: 全ターミナル一括トグル

モード・方向切り替え:

- <leader>tm: モード切り替え（Single/Side-by-Side）
- <leader>th: Horizontal（下部横分割）
- <leader>tv: Vertical（右側縦分割）
- <leader>tf: Float（フローティング）
- <leader>tD: 方向をサイクル切り替え（H→V→F）

サイズ変更:

- <M-k> / <M-j>: 高さを1行ずつ増減（Alt + k/j）
- <M-K> / <M-J>: 高さを5行ずつ増減（Alt + Shift + k/j）
- <leader>t+: 最大化
- <leader>t-: 小サイズ（10行）

REPL機能（選択範囲送信）:

- <leader>ts: 選択範囲をターミナルに送信
    - コマンド: ToggleTermSendVisualSelection
    - 使い方: ビジュアルモード（v）で範囲を選択してから実行
- <leader>tl: 選択行をターミナルに送信
    - コマンド: ToggleTermSendVisualLines
    - 使い方: ビジュアルラインモード（V）で行を選択してから実行
- <leader>tc: ターミナル画面とスクロールバック（履歴）を両方クリア
    - 使い方: ターミナルモード（t）で実行。`reset`コマンドの代わりに使う

管理機能:

- <leader>tn: ターミナルに名前を付ける
- <leader>tS: ターミナル選択UI

Git統合:

- <leader>gg: LazyGit（フローティング）

ターミナル内操作:

- <Esc><Esc> / <C-q> / jk: ノーマルモードへ
- <C-h/j/k/l>: ウィンドウ移動（上記「ウィンドウ間移動」と同様）

## scratch buffer

- <leader>+.

## LSPナビゲーション・診断・コード操作

- gr: 参照箇所を検索（関数が使われている場所を表示）
- gI: 実装へジャンプ
- gy: 型定義へジャンプ
- gD: 宣言へジャンプ
- K: ホバー情報表示（定義・ドキュメントをフローティングウィンドウで表示）
- gK: シグネチャヘルプ
- <leader>cd: 診断を表示
- gl / <leader>ld: 行の診断をフローティングで表示
- [e / ]e: 前/次のエラーへ移動
- <leader>ca: コードアクション
- <leader>cr: リネーム
- <leader>cl: LSP情報表示
- <leader>cf: 現在のファイルをフォーマット（LazyVimデフォルト、Conform.nvim経由）

※ VSCodeの`gh`（定義をフローティング表示）相当は、Neovimでは`K`が同等の機能を提供

## ファイル検索・エクスプローラー（Snacks.nvim）

- <leader>fe: ファイルエクスプローラーを開く
- <leader>ff: ファイル検索（隠しファイル・gitignoreファイルもデフォルトで表示）

## CopilotChat

- <leader>cc: チャットを開く
- <leader>cE: コード説明（ビジュアルモードで選択範囲、未選択時はバッファ全体が対象）
- <leader>cR: コードレビュー（同上）
- コンテキスト指定（チャット内で使用）:
    - `#file:path/to/file` - 特定ファイルを追加
    - `#buffer:current` - 現在のバッファ
    - `#buffers:visible` - 表示中の全バッファ
    - `#gitdiff` - Git差分
    - `#diagnostics:current` - 診断情報
    - `#glob:*.lua` - パターンマッチするファイル一覧
    - `#grep:TODO` - ワークスペース内検索

## claudecode.nvim（Neovim統合、Claude Code CLIとの連携）

- <leader>aI: Claude Code起動
- <leader>aS: Claude Code停止
- <leader>ai: Claude Codeステータス確認

## no-neck-pain.nvim

- <leader>nn: 中央寄せ表示モードの切り替え

## flash.nvim（高速移動・検索）

- s: Flash jump（ラベル付きジャンプ）
- S: Flash Treesitter（構文要素ジャンプ）
- f / F: 前方/後方文字検索
- ; / ,: 文字検索の繰り返し/逆方向
- ※ `t`/`T`キーは無効化済み（Neotestのキーバインドとの競合回避のため。設定ファイル: `lua/plugins/flash-config.lua`）

## テスト実行（Neotest）

neotest + neotest-golang（Go言語テストアダプタ）。サブテスト/テーブルテスト実行、レース検出（`-race`）、カバレッジ測定（`-cover`）、Testifyサポート、DAP（Debug Adapter Protocol）連携に対応。

### 基本

- <leader>tr: カーソル位置のテストを実行（最も近いテスト）
- <leader>tt: 現在のファイルのすべてのテストを実行
- <leader>tT: プロジェクト全体のテストを実行
- <leader>tl: 最後に実行したテストを再実行
- <leader>td: カーソル位置のテストをデバッグ実行

### 拡張

- <leader>tp: パッケージ/ディレクトリ単位でテスト実行
- <leader>tf: 失敗したテストのみ再実行（詳細出力付き）
- <leader>ta: プロジェクト全体のテスト実行（詳細出力付き）
- <leader>tA: プロジェクト全体のテスト実行（レース検出付き）
- <leader>tc: 最寄りのテストをカバレッジ測定付きで実行
- <leader>tC: 現在ファイルをカバレッジ測定付きで実行

### 結果・出力操作

- <leader>ts: テスト結果サマリーを表示/非表示
- <leader>to: テスト出力を表示
- <leader>tO: テスト出力パネルを切り替え
- <leader>tx: 実行中のテストを停止
- <leader>tq: テストquickfixウィンドウを閉じる
- <leader>t?: テストステータスを開く
- <leader>tw: テストウォッチモード切り替え

視覚表示: テスト関数の横にインラインで成功(✓)/失敗(✗)/実行中(◐)/スキップ(○)アイコンを表示。サマリーパネルはツリー形式で構造・結果を表示。

トラブルシューティング（"No tests found"等）: `.claude/docs/nvim-troubleshooting.md` 参照

# Claude Code

## モデル切り替え

`cmd+option+p`

## スクロール（Scrollコンテキスト／フルスクリーンレンダリング時のみ）

MacBook内蔵キーボードとroBa（ZMK自作キーボード）の両方で押せるよう追加した割り当て。
既定の `PageUp`/`PageDown`/`Ctrl+Home`/`Ctrl+End` も併存。

- `Option+u`（⌥U）: 半ページ上スクロール
  - アクション: `scroll:pageUp`
- `Option+d`（⌥D）: 半ページ下スクロール
  - アクション: `scroll:pageDown`
- `Option+g`（⌥G）: 会話の先頭へジャンプ
  - アクション: `scroll:top`
- `Option+Shift+g`（⌥⇧G）: 最新メッセージへジャンプ（オートフォロー復帰）
  - アクション: `scroll:bottom`
- `Ctrl+PageUp` / `Ctrl+PageDown`: roBaのARROWレイヤーでトラックボール回転時に送出されるキーを受ける
  - アクション: `scroll:pageUp` / `scroll:pageDown`
  - 使い方: roBaのARROWレイヤーを保持しながらトラックボールを回す

# WezTerm

## タブ移動

- `Cmd+Shift+]` / `Cmd+Shift+[`: 次/前のタブへ（wezterm自身のタブ）
- `Ctrl+Tab` / `Ctrl+Shift+Tab`: 次/前のタブへ（外部キーボード互換用）
  - フォアグラウンドプロセスが`herdr`の時は、この2つはwezterm自身のタブ切り替えではなくherdr側へ素通し（`SendKey`）される
  - roBaのARROWレイヤー（かな/LANG1ホールド）で `R`/`W` を押すとこのCtrl+Tab系が送出される

# herdr

## タブ移動

- `Ctrl+Tab` / `Ctrl+Shift+Tab`: 次/前のタブへ（wezterm経由でroBaのかな+R/Wから届く）
- `Cmd+Shift+]` / `Cmd+Shift+[`: 次/前のタブへ（wezterm配下で使う場合、wezterm自身のタブ切り替えに奪われるため実質無効。単体ターミナルやSSH越しでの利用時のみ有効）
- `Cmd+1`〜`Cmd+9`: タブ番号を直接指定（同上の制約あり）
- 既存の`prefix+p`/`prefix+n`/`prefix+1..9`（tmux風）は維持

# zsh

## suspenndを戻すコマンド

`fg`

## autosuggestions（サジェスト受諾）

- `→`（右矢印）: サジェストを全て受諾
- `Option+f`: 1単語ずつ受諾（forward-word）
- `Ctrl+F`: 1文字ずつ受諾（partial-accept）
- `Ctrl+O`: 非英数字（`/`・空白・`.`・`-` 等）を区切りに受諾（partial-accept）
