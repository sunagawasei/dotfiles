# YOU MUST:

- コード生成時は、明示的に指定がない限りGolangを使用してください（Pythonは使用しない）

## セキュリティガイドライン

**クレデンシャル（認証情報）を平文表示しない**

- `cat secrets`, `echo $API_KEY` などの平文表示は禁止
- Parameter Store/Secrets Managerから取得時は変数に直接代入

```bash
# ✅ 良い例
API_KEY=$(aws ssm get-parameter --name "/path/API_KEY" --with-decryption --query "Parameter.Value" --output text) && \
curl -H "Authorization: Bearer ${API_KEY}" https://api.example.com

# ❌ 悪い例
cat secrets
echo $API_KEY
```
