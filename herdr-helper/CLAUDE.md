# herdr外部・英語返信ヘルパー専用セッション(Claude Code用)

このディレクトリをcwdとするClaude Codeセッションは、herdrの外(WezTermの右pane)に常駐する英語返信ヘルパー専用。主用途は `/herdr-english-reply`(外部ヘルパーmodeで動作する)。

常用するヘルパーはcursor-agent(quota節約のため)で、その規則は `AGENTS.md`。本ファイルはClaude Codeで起動した場合用。

## herdr CLIの利用範囲(session全体の常設規則)

使用してよいherdr CLI操作は次の4つだけ(herdr CLI操作に関する読み取り専用allowlist):

- `herdr workspace list`
- `herdr pane list`
- `herdr pane read`
- `herdr api snapshot`

この4つは、`herdr` skillの「外側からfocused paneをinspect/controlしない」という安全既定(herdr upstream同梱の汎用ガード)に対する**ユーザー承認済みの例外**。

上記以外のherdrコマンド(`pane send-text` / `send-keys` / `run` / `focus` / `close` / `split`、workspace・tab・server・config等の変更系すべて)は、ユーザーに頼まれても自発的に実行せず、「元のpane(herdr側)で送信・操作してください」と案内する。`/herdr` skillをロードした場合も本規則が優先される。

一般のshellコマンド・ファイル読み書きはこのallowlistの対象外(通常どおりpermissions設定に従う)。

## 起動

`Ctrl+Shift+U`でヘルパーpaneの起動/表示/非表示をトグルする(起動コマンドの実体は`wezterm/keybinds.lua`の`helper_spawn`。既定はcursor-agent)。Claude Codeで立てたい場合は同じ形のコマンドを`/opt/homebrew/bin/claude`に差し替えて手動実行する。`env -u HERDR_PANE_ID`等は内部mode誤判定を防ぐため必須。往復は`Ctrl+Shift+E`。

## 技術的な強制について

主要な変更系コマンドは `.claude/settings.json` のpermissions denyで技術的にも遮断される。ただしBashルールのマッチングには回避可能な形が存在する(公式既知の制限)ため、本ファイルの規則が一次の行動規範であり、permissionsは安全網という関係にある。
