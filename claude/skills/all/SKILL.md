---
name: all
user-invocable: true
description: |
  codex・copilot・cursorの3エージェントを並列起動してコード分析を実行する。
  Usage: /all "プロンプト"
---

# /all

**read-only 分析専用。ファイル編集・コミット・プッシュを伴うタスクには使わないこと。**

ユーザーの引数をタスクプロンプトとして、3エージェントを並列起動する。

## オーケストレータ委譲原則（CRITICAL）

オーケストレータ（このスキルを実行する claude 自身）は、**ユーザーの論点・スコープ・指定ファイルパスのみ**を自己完結プロンプトとして各エージェントに渡す。
オーケストレータ自身が Bash/Read/Grep/Glob でリポジトリを事前調査してファイル内容を埋め込んではならない。
実際の調査（ファイル読み取り・git 探索）は各 CLI（codex/copilot/cursor）が CWD で自律実行する。

## 重要な制約

- `run_in_background: true` は絶対に使用しない（セッション終了時やコンテキスト進行時にkillされる原因になる）
- 3エージェントすべてがフォアグラウンドで完了するまで待機し、結果を統合すること

> 各CLIの canonical form・フラグ・timeout は各エージェント定義（`agents/{codex,copilot,cursor}.md`）が管理する。オーケストレータは subagent を起動するだけで CLI コマンドを直接組み立てない。

## 手順

1. 引数をタスクプロンプトとして受け取る
2. 3つのエージェントを**同一メッセージ内で並列に**起動:
   - codexエージェント（subagent_type: codex）
   - copilotエージェント（subagent_type: copilot）
   - cursorエージェント（subagent_type: cursor）
3. 3つの結果を統合して日本語で提示
   - 共通して指摘された点を強調
   - 各エージェント固有の指摘も明記
