# Codex の設定管理

`CODEX_HOME` は `~/.config/codex` に設定する(`home-manager/shell.nix`)。
Claude Code と同様に、実設定はローカルに置き、公開用の設定を Git で管理する。

| ファイル | 用途 |
| --- | --- |
| `AGENTS.md` | Codex のグローバル指示 |
| `config.public.toml` | モデル、承認方式、Vim、キーバインドなどの公開用設定 |
| `claude-rules.md` | `../claude/CLAUDE.md` への symlink |
| `skills` | `../claude/skills` 全体への symlink |
| `config.toml` | 実際に読み込むローカル設定(Git 管理対象外) |

認証情報、セッション、キャッシュ、プロジェクトの信頼設定、通知コマンド、MCP 接続設定は公開しない。
共有スキルは Claude 側で公開対象にしている原本を追跡する。
Codex が配置する `.system` スキルは Git 管理対象外。

## 初回セットアップ

既存の `config.toml` が無い環境では、公開用設定をコピーする。
既存のローカル設定には上書きしない。

```sh
cd ~/.config/codex
cp -n config.public.toml config.toml
```

必要なプロジェクトの信頼設定やアプリの接続設定は、その環境で追加する。
`config.public.toml` は Codex が直接読み込むファイルではない。

## 設定変更時

`config.toml` を編集したら、公開用設定を再生成する。

```sh
cd ~/.config/scripts
go run ./cmd/publish-codex-settings
```

出力対象はコードで指定したキーだけで、新しい MCP や環境変数の設定は自動的には公開されない。
生成結果に個人パスなどが残る場合は、出力を更新せずエラーになる。
