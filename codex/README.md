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

## プロジェクト単位の共有

指示の原本は各プロジェクトの `CLAUDE.md`、`.claude/CLAUDE.md`、`CLAUDE.local.md`、`.claude/rules/` に置く。
グローバルの `AGENTS.md` で、祖先と作業対象の階層にある原本を読むよう指示している。
`project_doc_fallback_filenames` だけでは、同じ階層に `AGENTS.md` がある場合の `CLAUDE.md` や、Git ルートより上の共通指示は補えない。
ルールの `paths` の判定と追加ファイルの読み込みはエージェントが行う。

スキルは `.agents/skills -> ../.claude/skills` の相対リンクで共有する。
新しいプロジェクトも Codex のタスク開始時に設定する。初回の入力前から候補に出したい場合は、先に次を実行する。

```sh
go -C ~/.config/scripts run ./cmd/share-project-skills /path/to/project
```

既存の `.agents/skills` が実ディレクトリなら、Claude 側の各スキルへのリンクだけを追加する。
同名の別設定や別のディレクトリへのリンクは上書きせず、エラーとして報告する。
原本の更新はそのまま反映される。実ディレクトリへ新しいスキルを追加した場合はコマンドを再実行する。
リンクの削除や既存ファイルの変更は行わない。既存の `.agents/skills` が実ディレクトリなら、Codex 専用スキルも保持する。

```sh
# 配下のプロジェクトをまとめて設定する
go -C ~/.config/scripts run ./cmd/share-project-skills --recursive /path/to/projects
# 変更せずに検査する
go -C ~/.config/scripts run ./cmd/share-project-skills --recursive --check /path/to/projects
```

`settings.json` の権限、hooks、MCP 接続設定、エージェント定義はツールごとに管理する。
この共有は、Claude Code 専用の実行機能を Codex に追加するものではない。

## 初回セットアップ

既存の `config.toml` が無い環境では、公開用設定をコピーする。
既存のローカル設定には上書きしない。

```sh
cd ~/.config/codex
cp -n config.public.toml config.toml
```

必要なプロジェクトの信頼設定やアプリの接続設定は、その環境で追加する。
`config.public.toml` は Codex が直接読み込むファイルではない。

## 自動承認と開発用キャッシュ

既定は `approval_policy = "on-request"`、`approvals_reviewer = "auto_review"`、`sandbox_mode = "workspace-write"`。
設定を保存しても実行中のセッションの権限は自動では変わらないため、`/permissions` で Approve for me を選ぶか、Codex を起動し直す。

Go の通常作業で毎回キャッシュへの書き込みがレビュー対象にならないよう、実設定の `[sandbox_workspace_write].writable_roots` に、そのマシンの `~/Library/Caches/go-build`、`~/Library/Caches/golangci-lint`、`~/go` を絶対パスで登録する。
これらの個人パスは公開用設定へ出力しない。
`.git` などの保護対象や Computer Use のアプリ承認は別の制約に従う。

## 設定変更時

`config.toml` を編集したら、公開用設定を再生成する。

```sh
cd ~/.config/scripts
go run ./cmd/publish-codex-settings
```

出力対象はコードで指定したキーだけで、新しい MCP や環境変数の設定は自動的には公開されない。
生成結果に個人パスなどが残る場合は、出力を更新せずエラーになる。
