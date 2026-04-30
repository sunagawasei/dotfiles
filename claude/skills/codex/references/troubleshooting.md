# トラブルシューティング

## インストール・設定

```bash
npm install -g @openai/codex   # インストール
codex --version                # バージョン確認
codex login                    # 認証
codex login status             # 認証状態確認
# Config: ~/.config/codex/config.toml
```

## Codex CLI が見つからない

```bash
# 確認
which codex
codex --version

# インストール
npm install -g @openai/codex
```

## 認証エラー

```bash
# 再認証
codex login

# ステータス確認
codex login status
```

## タイムアウト

| reasoning_effort | 推奨タイムアウト |
|-----------------|-----------------|
| low             | 60s             |
| medium          | 180s            |
| high            | 600s            |
| xhigh           | 900s            |

config.toml で設定:
```toml
[mcp_servers.codex]
tool_timeout_sec = 600
```

## Git リポジトリエラー

```bash
# Git 管理外で実行する場合
codex exec --skip-git-repo-check ...
```

## reasoning 出力が多すぎる

config.toml で設定:
```toml
hide_agent_reasoning = true
```

## stdinハング / プロセスがkillされる

**症状**: `codex exec` がそのまま止まる、または `< stdin >` と表示されて入力待ちになる

**原因と修正**:

| 原因 | 修正 |
|------|------|
| `< /dev/null` なし → EOF待ちでハング | 常に末尾に `< /dev/null` を追加 |
| `codex "..."` (`exec` なし) → TUI起動 | 必ず `codex exec` を使う |
| approval_policy が対話承認要求 | `-c approval_policy="never"` を追加 |
| 認証未完了 → パスワード入力待ち | `codex login` を先に実行 |
| `resume` でセッションID省略 | `--last` フラグを使う |

**canonical form（これ以外を使わない）**:

```bash
codex exec \
  --sandbox read-only \
  -c approval_policy="never" \
  --skip-git-repo-check \
  --color never \
  --cd "$(pwd)" \
  "<prompt>" \
  < /dev/null
```

## セッション継続できない

`codex sessions list` は codex-cli 0.122 には存在しない。

```bash
# 直前のセッションを再開
codex exec resume --last -c approval_policy="never" --color never < /dev/null

# 特定セッションIDで再開
codex exec resume {SESSION_ID} -c approval_policy="never" --color never < /dev/null
```

## sandbox 権限エラー

| エラー | 原因 | 解決策 |
|--------|------|--------|
| Permission denied | read-only で書き込み | workspace-write に変更 |
| Network blocked | sandbox 制限 | danger-full-access（慎重に） |

## メモリ不足

大きなコードベースを分析する場合:
1. 対象ファイルを絞る
2. 段階的に分析
3. `--config context_limit=...` で調整
