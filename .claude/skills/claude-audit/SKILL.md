---
name: claude-audit
description: skill/rule/CLAUDE.md/メモリの棚卸しと公式ベストプラクティスへの準拠化。グローバルと全プロジェクトの重複・矛盾・陳腐化をsonnet並列で検出・整理し、公式一次情報に基づき全CLAUDE.mdの構成・サイズを是正、メイン(Fable/Opus)が検証・適用する定期メンテナンス
---

# Claude資産の棚卸し

グローバル設定と全プロジェクトのClaude資産(CLAUDE.md・rules・skills・メモリ)から、重複ルール・矛盾する指示・陳腐化した記述を洗い出して整理する。あわせて公式ベストプラクティスに照らして全CLAUDE.mdの構成・サイズを是正する。目安: 月1回、モデル移行期、または「設定が増えて出力の質が落ちた」と感じた時。

## 対象の列挙

1. グローバル: `~/.config/claude/{CLAUDE.md,rules/,skills/,agents/,commands/}` + グローバルメモリ(`~/.config/claude/projects/-Users-s23159--config/memory/`)
2. プロジェクト: `ls ~/.config/claude/projects/` でメモリ持ちを列挙し、対応する実ディレクトリの存在を確認。**消滅済みプロジェクトのメモリは削除候補として報告**(勝手に消さない)。注意: プロジェクトスラッグはパス中の `_`(アンダースコア)や `/`(ネスト)も `-` に変換されるため、スラッグから機械的に復元したパスが不在でも即「消滅」と断定しない。`ls ~/work ~/poc` の実名と突き合わせる(例: `-work-value-domain-management` の実体は `work/value_domain_management`、`-work-tky02-cn-down-backlog-cn-list-updater` の実体は `work/tky02-cn-down-backlog/cn-list-updater`)
3. 各実在プロジェクトの `CLAUDE.md` / `.claude/`(rules・skills・docs。settings系は読むだけ)

## 検出観点

- **重複**: メモリ同士 / メモリとrule・CLAUDE.md / プロジェクトとグローバル(例: datadog.md)の二重記録 → 1箇所に統合し、他方は削除または参照化
- **矛盾**: 新旧方針の食い違い(同一ファイル内の残骸も)、実構成との乖離(参照先ファイル・スキル・コマンドの存在を`ls`/`glob`で確認)
- **陳腐化**: 「未コミット」「〜予定」「作業中」等の状態記述を`git log`/`status`と突き合わせ。削除済みツールへの言及
- **整合性**: MEMORY.mdインデックス↔ファイルの同期、`[[リンク]]`の解決、frontmatter nameとファイル名の一致(snake_case統一)
- **公式準拠**: CLAUDE.mdの行数超過、rule/skillの非公式frontmatterフィールド、公式の使い分け基準(下記フェーズ0)からの逸脱

## 実行手順

### 0. 公式最新推奨の確認

- `claude-code-guide`エージェント(**`model: sonnet`**)に、CLAUDE.md/rules/skills関連の**公式一次情報**を調査させる: code.claude.com/docsの`memory` / `features-overview` / `large-codebases`ページ、Anthropic公式ブログ(Steering Claude Code等)。**記憶で答えさせずWebFetch/WebSearchで現物確認**させる
- 得るもの: 推奨サイズ(2026-07時点は「CLAUDE.mdは200行未満・毎セッション必要な事実のみ」)、使い分け基準(30行超の手順→skill、path固有→rules、確実に実行させたいもの→hook)、公式frontmatterフィールド(ruleは`paths:`のみ)、アンチパターン
- 以降のフェーズはこの調査結果を基準に判断する(数値は時期により更新されうるため、記憶で決め打ちせず毎回この確認をやり直す)

### 1. sonnet班の並列起動

- プロジェクトを作業量バランスでグループ化(メモリ50件超は単独班)し、general-purposeサブエージェント(**`model: sonnet`**)を並列起動
- プロンプトに必ず含める: 担当範囲の絶対パス(メモリdirのslugも) / 上記の検出観点 / 適用範囲(メモリ更新・統合・MEMORY.md同期・明確な誤りの修正) / **安全ルール**(git commit/push禁止・ソースコード本体変更禁止・削除は「repoに記録済み/統合済み/明確に陳腐化」のみ・迷ったら残して報告) / 報告フォーマット(修正した点・削除統合した点・未対応の提案)

### 2. 全CLAUDE.mdの洗い出しと是正

フェーズ0で確認した基準をもとに、CLAUDE.md単体にフォーカスした是正パスをsonnet班(**`model: sonnet`**)に実施させる。

- 網羅列挙(サブディレクトリ・`.claude/CLAUDE.md`含む):
  ```bash
  find ~/work ~/poc ~/.config -maxdepth 4 -name 'CLAUDE.md' -not -path '*/node_modules/*' -not -path '*/.git/*' -not -path '*/attic/*'
  ```
  各ファイルの行数を測定。upstream管理ファイル(symlink等。例: herdrの`CLAUDE.md`→`AGENTS.md`)は対象外
- **基準超過(目安200行超)**: スリム化。原則は「情報は削除でなく移設」— 詳細・手順・リファレンスを`.claude/docs/`や既存参照ドキュメント(KEYMAPS.md等)へロスレス移設して1行参照化。ルート/サブ/`.claude`間の二重ロード重複は一本化
- **基準内(目安200行以内)**: 軽量点検のみ — (a)30行超の手順ブロック→移設 (b)明白な重複 (c)実構成との乖離、の3基準。**問題なければ触らない**(無理に変更を作らない)
- 非公式frontmatterフィールド(過去の例: ruleの`trigger:`)が見つかったら、対象ファイル本体だけでなく**生成テンプレート(add-rule references等)からも**除去する

### 3. メイン(Fable/Opus)の検証 — 必須・省略禁止

- 班の「編集した」報告は鵜呑みにしない。**全報告をgrep/存在確認で実物スポットチェック**する(2026-07-04に「修正したと報告したが未適用」の実例)
- 削除の根拠(「グローバルruleと重複」等)は、引用元の記述が実在するか自分で確認する
- ステップ2の追加検証: 行数再測定、移設先ファイルの存在確認、移設元への参照残存・内容欠落がないか確認
- メモリ整合性スイープを機械実行:

```bash
cd ~/.config/claude/projects && for d in ./*/memory; do cd "$d" 2>/dev/null || continue
  for f in *.md; do [ "$f" = MEMORY.md ] && continue; grep -q "$f" MEMORY.md || echo "$d unindexed:$f"; done
  for n in $(grep -ho '\[\[[a-zA-Z0-9_-]*\]\]' *.md 2>/dev/null | tr -d '[]' | sort -u); do [ -f "$n.md" ] || echo "$d unresolved:[[$n]]"; done
  cd ~/.config/claude/projects; done
```

### 4. ユーザー判断と適用

- 判断が分かれるもの(削除可否・公開可否・方針矛盾の解消方法)は**AskUserQuestionでまとめて確認**
- git非追跡ファイルの整理は削除でなく`attic/`退避(可逆)を選ぶ
- dotfiles(`~/.config`)の変更はユーザー承認後、commitエージェント(`model: sonnet`)でコミット。他リポジトリはcommitしない(working treeに残して報告)

## 定期トリガー

- 手動: `/claude-audit`
- 定期化したい場合はCronCreateで月次スケジュールに載せる(例: 毎月1日に`/claude-audit`)
