# Troubleshooting

## バージョン・状態確認

```bash
agent --version          # バージョン確認
agent about              # バージョン、OS、モデル、認証情報
agent status             # 認証状態
agent models             # 利用可能なモデル一覧
```

## 認証エラー

```bash
agent login              # ブラウザで認証
agent status             # 認証確認
```

## モデル関連

```bash
# composer-2 が利用不可の場合
agent -p --model composer-2-fast --mode plan --trust "<prompt>"

# 利用可能モデルを確認
agent models
```

## バイナリ関連

```bash
which agent              # バイナリ確認
ls -la ~/.local/bin/agent  # シンボリックリンク確認
# 期待値: ~/.local/bin/agent -> ~/.local/share/cursor-agent/versions/.../cursor-agent
```

## `--trust` フラグが必要な理由

`-p`（非対話モード）でサブエージェントとして使う場合、Cursor はワークスペース信頼確認を対話的に求める。
`--trust` フラグで確認をスキップする。必ず `-p` と組み合わせて使うこと。

```bash
# 正しい
agent -p --model composer-2 --mode plan --trust "<prompt>"

# NG: --trust なしでは対話プロンプトで止まる可能性
agent -p --model composer-2 --mode plan "<prompt>"
```

## `--mode` フラグの省略禁止

`--mode` を省略するとフルエージェントモード（write権限あり）で起動する。
常に `--mode plan` または `--mode ask` を明示すること。

```bash
# NG: write権限が有効になる
agent -p --model composer-2 --trust "<prompt>"

# OK
agent -p --model composer-2 --mode plan --trust "<prompt>"
```

## タイムアウト

大きなコードベースでは `--mode plan` が時間を要することがある。
分析対象を特定のファイル・ディレクトリに絞ること：

```bash
# スコープを絞る
agent -p --model composer-2 --mode plan --trust "
Analyze only src/auth/ directory. Focus on ...
"
```

## セッション再開

```bash
# 直前のセッションを再開
agent --continue -p --model composer-2 --mode plan --trust "<prompt>"

# 特定セッションを再開
agent --resume <chatId> -p --model composer-2 --mode plan --trust "<prompt>"
```

## Cursor IDE との混同について

このサブエージェントは **Cursor Agent CLI**（ヘッドレスCLIツール）を使用。
Cursor IDE（GUIアプリケーション）とは別物。バイナリは `~/.local/bin/agent`。
