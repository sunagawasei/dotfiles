# Codex CLI グローバル設定

このファイルは Codex CLI の全プロジェクト共通の指示を定義します。

- dotfiles 内の `codex/config.toml` はローカル専用。変更時は `~/.config/scripts` で `go run ./cmd/publish-codex-settings` を実行し、公開用の `codex/config.public.toml` を更新して commit する。Git 管理の範囲と初回セットアップは `~/.config/codex/README.md` を参照する。

## Claude Code と共有する規約

- 毎タスク開始時に `$CODEX_HOME/claude-rules.md` を読む。この symlink の正本は `~/.config/claude/CLAUDE.md` であり、コピーや同期用の一覧を作らない。
- 出力規約、対話・確認の規約、操作の安全規約、スクリプト言語の規約を適用する。競合時はこの AGENTS.md とプロジェクトの明示指示を優先する。
- Claude Code 固有のツール名は Codex の利用可能な同等ツールへ読み替える。同等機能が無ければ、その機能が必要な手順は実行できないことを明示する。
- 「エージェント役割分担」とその配下の委譲フローは Claude Code 用であり、自動適用しない。「codex に裏取り・査読させる」という指示も自己査読の再帰呼び出しには読み替えない。独立した査読が必要な場合は利用可能な別の査読者を使う。
- グローバルの `~/.config/claude/rules/` とプロジェクトの `.claude/rules/` は、機密ファイルの禁止規約を守り、frontmatter の `paths` が作業対象に一致する規約と、`paths` の無い共通規約を読む。
- `$CODEX_HOME/skills` は `~/.config/claude/skills` 全体への symlink。スキルの追加・変更・削除は正本へ行い、個別リンクの一覧は管理しない。Claude Code 固有の実行環境を前提とするスキルは、必要な機能が利用可能なときだけ実行する。

## プロジェクト単位の共有

- 毎タスク開始時に、祖先ディレクトリから作業ディレクトリまでの `CLAUDE.md`、`.claude/CLAUDE.md`、`CLAUDE.local.md` を読み、対象ファイルがさらに下の階層ならその階層までの指示も読む。同じ階層に `AGENTS.md` があっても省略しない。原本中の `@path` による参照も機密ファイルの禁止規約を守って読む。symlink などで同じ原本を参照する場合は一度だけ読む。
- `.claude/rules/` も各祖先と対象ディレクトリで確認する。`paths` の無い規約は共通、`paths` のある規約は置かれたプロジェクトからの相対パスで対象を判断する。グローバル、祖先、下位ディレクトリの順に適用し、明示された Codex 固有の制約を優先する。機密ファイルの禁止規約は共有時にも守る。
- プロジェクトの `.claude/skills/` がある場合は `go -C ~/.config/scripts run ./cmd/share-project-skills <ディレクトリ>` を実行し、`.agents/skills` から原本を共有する。Git ルートから作業ディレクトリまでと、今回扱う下位ディレクトリが対象。既存リンクは再作成せず、衝突した設定は上書きしない。新しいスキルが候補に出なければ Codex を再起動する。
- スキルと指示の原本は Claude 側で編集する。`settings.json`、hooks、MCP の接続設定、エージェント定義を Codex の設定形式として流用しない。専用ツールが必要な手順は利用可能な同等機能がある範囲で実行する。

## 言語設定

**重要**: すべての応答は日本語で行ってください。技術用語やコード識別子は原文のままで構いません。

## 日本語文章規範

- 実質的な日本語の文章（返信・レポート・ドラフト・ドキュメント）を書く/推敲するときは、`japanese-tech-writing` skill（`$CODEX_HOME/skills/japanese-tech-writing/SKILL.md`）の規範に従ってください。ただし「整形」節（一文一行改行・脚注記法等）は書籍原稿・記事向けの規則であり、チャット返信には適用しません
- 記事・エッセイ・解説文など読み物を書くときは `cognitive-rhythm-writing` skill（`$CODEX_HOME/skills/cognitive-rhythm-writing/SKILL.md`）も併用してください（日常の返信には併用しない。未回収の緊張を残す読み物向け設計であり、結論先行の返信とは目的が逆のため）

## コーディング規約

- コード生成時は、明示的に指定がない限り **Go言語** を使用してください（Pythonは使用しない）
- 簡潔で読みやすいコードを心がけてください
- 適切なエラーハンドリングを含めてください

## セキュリティガイドライン

### 通常作業の自動承認

- 依頼の範囲内の編集、調査、テスト、ビルド、許可済みのコミットは追加の確認を挟まず進める。共有した Claude Code の規約にある一般的な承認待ちの手順も、同じ操作が既に許可されていれば重ねて要求しない。
- サンドボックス外へのアクセスが必要な通常作業は、チャットで先に許可を聞かず、`require_escalated` で auto-review に回す。レビュー対象の表示と、人間への承認要求を区別する。
- 自動レビューが拒否した場合は理由を確認し、より安全な方法を使う。続行に人間の判断が必要な場合だけ、拒否された操作と理由を説明して確認する。
- 次節の危険な操作や外部への送信・公開について明示されている確認要件は引き続き守る。

### 危険なコマンドの実行前確認

以下のコマンドを実行する前に、必ずユーザーに確認を求めてください：

- `sudo` を含むすべてのコマンド
- `rm -rf` または `rm -r` を含むコマンド
- `git push --force` または `git push -f`
- `git reset --hard`
- `chmod 777` または過度に緩い権限設定
- `curl | sh` や `wget | bash` 等のパイプ実行
- データベースの DROP/TRUNCATE 操作

### アクセス禁止ファイル

以下のファイルやパターンにマッチするファイルの内容を読み取ったり、表示したりしないでください：

- `.env` ファイル（`.env.example` は除く）
- `secrets` を含むファイル名
- `*token*` パターンにマッチするファイル
- `*credential*` パターンにマッチするファイル
- `*private*` パターンにマッチするファイル（秘密鍵など）
- SSH秘密鍵（`id_rsa`, `id_ed25519` 等）
- AWS認証情報ファイル（`~/.aws/credentials`）

### クレデンシャル平文表示禁止

- `cat secrets`, `echo $API_KEY` などの平文表示は禁止
- Parameter Store/Secrets Manager から取得時は変数に直接代入

```bash
# ✅ 良い例
API_KEY=$(aws ssm get-parameter --name "/path/API_KEY" --with-decryption --query "Parameter.Value" --output text) && \
curl -H "Authorization: Bearer ${API_KEY}" https://api.example.com

# ❌ 悪い例
cat secrets
echo $API_KEY
aws ssm get-parameter --with-decryption ... # 結果を直接表示
```

## 応答スタイル

- 簡潔かつ的確な応答を心がけてください
- 技術的な正確性を優先し、過度な装飾は避けてください
- 絵文字は明示的に要求されない限り使用しないでください
- コードリファレンスには `file_path:line_number` 形式を使用してください

## 注意事項

このファイルでの指示は技術的な強制力を持ちません（Claude Code の `permissions.deny` とは異なり、ソフトウェア的にブロックすることはできません）。これらは「ベストプラクティス」としての指示です。
