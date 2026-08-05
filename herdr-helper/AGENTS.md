# herdr外部・英語返信ヘルパー専用セッション

このディレクトリをcwdとするエージェントセッションは、herdrの外(WezTermの右pane)に常駐する英語返信ヘルパー専用。主用途は `/herdr-english-reply`。

## herdr CLIの利用範囲

読み取りのみ行う:

- `herdr workspace list`
- `herdr pane list`
- `herdr pane read`
- `herdr api snapshot`

この読み取りは、`herdr` skillの「外側からfocused paneをinspect/controlしない」という安全既定(herdr upstream同梱の汎用ガード)に対するユーザー承認済みの例外。

`pane send-text` / `send-keys` / `run` / `focus` / `close` / `split`、workspace・tab・server・configの変更系は実行しない。送信を頼まれたら「元のpane(herdr側)で送信してください」と案内する。読み取り専用は技術的に強制されていないので、この規則自体が行動の根拠になる。

## キー

- `Ctrl+Shift+U`: ヘルパーペインの起動/表示/非表示トグル(初回起動もこのキー。非表示中もプロセスは生存)
- `Ctrl+Shift+E`: herdrペインとヘルパーペインのフォーカス入れ替え

起動コマンドの実体は `wezterm/keybinds.lua` の `helper_spawn`(右27%幅・cwd=このディレクトリ・`env -u HERDR_PANE_ID`等で内部mode誤判定を防いだ上で `HERDR_ENV=1` を注入)。手動で作り直す場合も同じ形を使う。
