---
name: codex-copilot
user-invocable: true
description: |
  codexとcopilotを並列起動してコード分析を実行する。
  Usage: /codex-copilot "プロンプト"
---

# /codex-copilot

ユーザーの引数をタスクプロンプトとして、codexとcopilotの2エージェントを並列起動する。

## オーケストレータ委譲原則（CRITICAL）

オーケストレータ（このスキルを実行する claude 自身）は、**ユーザーの論点・スコープ・指定ファイルパスのみ**を自己完結プロンプトとして各エージェントに渡す。
オーケストレータ自身が Bash/Read/Grep/Glob でリポジトリを事前調査してファイル内容を埋め込んではならない。
実際の調査は各 CLI（codex/copilot）が CWD で自律実行する。

## 重要な制約

- `run_in_background: true` は絶対に使用しない（セッション終了時やコンテキスト進行時にkillされる原因になる）
- 2エージェントがフォアグラウンドで完了するまで待機し、結果を統合すること

> 各CLIの canonical form・フラグ・timeout は各エージェント定義（`agents/{codex,copilot}.md`）が管理する。オーケストレータは subagent を起動するだけで CLI コマンドを直接組み立てない。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 2つのエージェントを**同一メッセージ内で並列に**起動:
   - codexエージェント（subagent_type: codex）
   - copilotエージェント（subagent_type: copilot）
3. 両方の結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
