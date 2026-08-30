# Neovim

## LSP定義ジャンプ

- gd: 定義へジャンプ（カレントバッファ）

## 診断全般のジャンプ（warning/error/info含む）

- ]d: 次の診断へ移動
- [d: 前の診断へ移動

## コメントアウト

`gcc`

## 折り畳み開閉

- za: 折り畳みの開閉（トグル）
- zM: すべての折り畳みを閉じる
- zR: すべての折り畳みを開く

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

- <leader>t1 / <leader>t2 / <leader>t3: 番号付きターミナルをトグル
- <leader>tn: 新しいターミナルを開く（1〜9のうち未使用の最小番号。満杯なら警告）
- <C-/>: 最後のターミナルをトグル
  - 端末はCtrl+/を`<C-_>`として送るため、設定側は両方を同じ動作に割り当てている

ターミナル間の切り替え:

- ]t / [t: 次 / 前のターミナルへ（ノーマルモードのみ。`]` `[` はターミナルモードではシェルへ渡す）
  - 存在するターミナルだけを巡回するので、番号が飛んでいても空のターミナルは作られない
  - lazygit（99）とHunk（98）は巡回対象外。ただしフロート表示中に切り替えるとウィンドウは閉じる（プロセスは生き続けるので`<leader>gg`で同じセッションに戻れる）

サイズ変更:

- <M-k> / <M-j>: 高さを1行ずつ増減（Alt + k/j）

ターミナル内操作:

- <leader>tc: ターミナル画面とスクロールバック（履歴）を両方クリア
    - 使い方: ターミナルモード（t）で実行。`reset`コマンドの代わりに使う
- <Esc><Esc>: ノーマルモードへ
- <C-h/j/k/l>: ウィンドウ移動（上記「ウィンドウ間移動」と同様）

Git統合:

- <leader>gg: LazyGit（フローティング）
- <leader>gr: Hunk（差分レビュー）

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
- gl: 行の診断をフローティングで表示
- <leader>dy: 行の診断をクリップボードにコピー
- [e / ]e: 前/次のエラーへ移動
- <leader>ca: コードアクション
- <leader>cr: リネーム
- <leader>cl: LSP情報表示
- <leader>cf: 現在のファイルをフォーマット（LazyVimデフォルト、Conform.nvim経由）

※ VSCodeの`gh`（定義をフローティング表示）相当は、Neovimでは`K`が同等の機能を提供

## ファイル検索・エクスプローラー

- <leader>e / -: ファイルエクスプローラー（oil.nvim。`-`は親ディレクトリ）
- <leader><Space>: ファイル検索（Snacks picker）
- <leader>sg: プロジェクト内grep（Snacks picker）

Snacksの`<leader>f*`・`<leader>s*`の大半は1ヶ月の実測で未使用だったため無効化してある。
無効化したキーの一覧は`nvim/lua/plugins/zz-disabled-keys.lua`にあり、該当行を消せば復活する。

## claudecode.nvim（Neovim統合、Claude Code CLIとの連携）

- <leader>ab: 現在のバッファをClaudeに追加
- <leader>as: 選択範囲をClaudeへ送信（ビジュアルモード）
- <leader>at: カーソル位置のファイルをClaudeに追加（oilバッファ専用）
- <leader>aa: 差分の変更を受け入れる

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

## ペイン移動

- `Ctrl+Shift+H` / `Ctrl+Shift+L` / `Ctrl+Shift+K` / `Ctrl+Shift+J`: 左/右/上/下のペインへ（`ActivatePaneDirection`）
- `Leader+z`: ペインのズーム切り替え（`TogglePaneZoomState`）

## 英語返信ヘルパー（herdr外の常駐pane）

- `Ctrl+Shift+E`: フォーカス入れ替え（`ActivatePaneDirection("Next")`。2ペイン構成なら1キーで往復する）
- `Ctrl+Shift+U`: ヘルパーの起動/表示/非表示トグル
  - ヘルパーペインが無ければ右27%幅でcursor-agentを新規起動する（初回起動もこのキー1つ。手動コマンド不要）
  - 起動時に`--force`（TUIの`Run Everything`）を付けるため、allowlist外のコマンドで承認プロンプトが出ない。`.cursor/cli.json`の`deny`は`--force`より優先されるので、herdrの変更系（`pane send-text`等）は引き続き拒否される
  - 表示中なら隠す（weztermにペイン単位のhide/showが無いため、herdrペインのズームで代替。非表示中もヘルパーのプロセスと会話は生存）。非表示中なら表示に戻してヘルパーへフォーカス
- `Ctrl+Shift+T`: herdrペインからヘルパーへ `translate` を**下書き**してフォーカス移動（Enterは自分で押す）
  - 確定を送らないのは、実Enterとの等価性が未実測で、ヘルパー側に下書きが残っていた場合の連結やモーダルの誤確定を避けるため
  - herdrペイン以外で押しても何もしない。ヘルパー不在時も何もしない
- ヘルパーペインの識別は起動時に立てるuser var `herdr_helper=1`（`keybinds.lua`の`is_helper_pane`）。プロセス名や「herdrでない」判定では別TUIのペインを誤爆しうるため使わない。herdrペイン自体の判定は`is_herdr_pane`（フォアグラウンドプロセス名）
- copy modeやペインナビゲーションモード（`Leader+q`）がactiveな間はそのkey tableが優先されるため、往復前にmodeを抜ける

# herdr

## copy mode

- `Ctrl+Shift+Y` / `prefix+[`: copy modeに入る（`herdr/config.toml`の`keys.copy_mode`）
  - `Ctrl+Shift+Y`はweztermに同名バインドがあると奪われる（translateトリガーを`Ctrl+Shift+T`へ移した経緯あり）。weztermのCTRL|SHIFT系を増やすときはここと衝突しないか確認する

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
