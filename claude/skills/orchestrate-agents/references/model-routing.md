# codexワーカーのモデル/effort振り分け（2026-07-08〜）

agmsg configのper-workerキー（`spawn.codex_model.<name>` / `spawn.codex_effort.<name>`、codex headless限定）により、ワーカー名がモデル+effortのプリセットになっている。タスク難易度の判定はコードで自動化せず、依頼側（Claude）が適切な名前のワーカーへ送ることで実現する。

| タスク | 宛先ワーカー | モデル/effort |
|---|---|---|
| 実質的な機能実装の自走 | codex-impl | gpt-5.6-sol / xhigh（per-workerキー明示）・turn_timeout 3600s・implementer layout（cwd=対象repo・permission profileでrepo書き込み可） |
| 実装後レビュー（オンデマンド） | codex | gpt-5.6-sol / xhigh（per-workerキー明示） |
| 調査・軽微な確認・大量列挙 | codex-research | gpt-5.6-sol / xhigh（per-workerキー明示。遅いと感じたら `spawn.codex_effort.codex-research: high` へ下げる） |
| 大規模・設計横断の節目レビュー | codex-deep（一時spawn→使い捨て） | gpt-5.6-sol / max（configキー設定済み） |

codex-deepは常駐させず、必要時にspawnし終わったらdespawnする:

```bash
~/.agents/skills/agmsg/scripts/spawn.sh codex codex-deep --team <team> --project <project> \
  --role-file ~/.agents/skills/agmsg/db/spawn-roles/codex.codex.md   # review役を流用
# 使用後: ~/.agents/skills/agmsg/scripts/despawn.sh <team> claude codex-deep
```

注意:
- 重いモード（max/ultra）への自動エスカレーションはしない。gpt-5.6系のクレジット消費倍率が未公開のため、明示的に選んだ時だけ使う
- gpt-5.6系はプレビュー段階。モデルIDが無効化・改名されたら該当configキー（実効config: `~/.agents/skills/agmsg/db/config.yaml`）を更新して戻す（振り分けの仕組み自体はモデル非依存）
- **gpt-5.6-sol が bridge 経由で `400 "requires a newer version of Codex"` になる場合**: 原因はCLI版ではなく、app-server `initialize` の `clientInfo.name` に対する server-side gate（bridgeの既定名 `agmsg-codex-bridge` が sol の first-party allowlist に弾かれる）。per-worker キー `spawn.codex_client_name.<name>: codex_cli` を設定して first-party 名を名乗らせると解消する（既定は従来名のまま。sol を使う worker にのみ設定）。CLI を最新stableに上げても・再認証しても直らない（2026-07-10確認）。純正 `codex exec` は別API面のため通るので、exec成功をbridge成功の証拠にしないこと。
