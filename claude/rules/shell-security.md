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
