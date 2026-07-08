---
paths:
  - "**/*.sh"
  - "**/*.zsh"
  - "**/*.bash"
  - "**/zsh/**"
  - "**/fish/**"
---

# シェルスクリプトセキュリティガイドライン

## クレデンシャル（認証情報）を平文表示しない

**重要**: `cat secrets`, `echo $API_KEY` などの平文表示は禁止

### 基本原則

- Parameter Store/Secrets Managerから取得時は変数に直接代入
- 認証情報をターミナルに表示しない
- ログファイルに認証情報を記録しない

### 良い例 ✅

```bash
# AWS Systems Manager Parameter Storeから取得
API_KEY=$(aws ssm get-parameter --name "/path/API_KEY" --with-decryption --query "Parameter.Value" --output text) && \
curl -H "Authorization: Bearer ${API_KEY}" https://api.example.com

# Secrets Managerから取得
DB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id "db-password" --query "SecretString" --output text) && \
psql -h localhost -U user -d mydb
```

### 悪い例 ❌

```bash
# 平文表示（禁止）
cat secrets
echo $API_KEY
aws ssm get-parameter --name "/path/API_KEY" --with-decryption

# ログに残る（禁止）
echo "Using API key: $API_KEY"
```

## ベストプラクティス

1. 環境変数は使用直前に取得
2. 使用後は`unset`で削除（必要に応じて）
3. スクリプト終了時にクリーンアップ
4. デバッグ出力に認証情報を含めない

## 対話的なコマンド提案・実行にも適用する

このルールはスクリプトファイル編集時だけでなく、**ユーザーに提示するコマンド・Claudeが実行するコマンド全般**に適用する。

- 認証情報が標準出力に出るコマンド（`cat credentials`、`terraform output`の生表示、IAMキー発行結果の表示等）を提案・実行しない
- 秘密を扱う手順を提示するときは、変数代入・ファイルリダイレクト・`--with-decryption`の直接パイプなど、**画面に出ない形**で提示する
- ユーザーが誤って秘密を貼り付けそうな手順（発行結果のコピペ等）には「結果は貼り付けないでください」と一言添える

## コマンド置換を使った入力検証の落とし穴

`$(...)` は**末尾の改行をすべて剥がす**ため、「許可文字を削除して残余が空なら合格」型の検証は、違反が末尾改行のみの値を誤合格させる:

```sh
# ❌ false-passの例: val="value<改行>" は残余が "<改行>" だが $() が剥がして空になり合格してしまう
[ -z "$(printf '%s' "$val" | tr -d 'A-Za-z0-9._-')" ]
```

対策は**センチネル方式**: 同一コマンド置換内の最後に非改行文字とコマンドの終了コードを付加し、期待値と厳密比較する:

```sh
# ✅ 残余なし かつ tr成功 の場合のみ合格（trの実行失敗によるfail-openも同時に防ぐ）
rest="$(printf '%s' "$val" | LC_ALL=C tr -d 'A-Za-z0-9._-'; printf 'X%s' "$?")"
[ "$rest" = "X0" ]
```

補足:
- 文字範囲の解釈はロケール依存（shellのcaseグロブ/EREの `[A-Z]` 等）なので、検証は `LC_ALL=C` で固定したbyte単位判定にする
- 検証で弾いた未信頼値をそのまま警告・ログに出力しない（改行/ANSI制御文字によるログ行偽装が可能）。制御文字を除去し長さを切り詰めてから出す
- 実例: 2026-07-08 agmsg headlessモデル配線の `agmsg_codex_safe_token`（提案式のfalse-passバグを実機テストで検出し、この方式に修正した）
