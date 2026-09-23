---
name: report-note
description: 次回の/report実行時にNotionレポートへ挿入したい内容をメモリに登録する
---

# report-note スキル

`worklog`リポジトリの`/report` skill（Step 2.5）が読み取る「MTGで伝えたいこと」をメモリに書き込むスキル。ここで登録した内容は、次回`/report`実行時にNotionページの該当プロジェクト配下へ自動挿入され、挿入後はメモリから削除される。

`/report`は`worklog`のプロジェクト固有スキルのため、このスキルはどのディレクトリで呼び出しても、保存先は常に`worklog`のauto memoryディレクトリに固定する。

## 概要

「これも次のMTGで伝えたい」「次のレポートに入れといて」と言われた内容を、`/report` skillが読み取れる形式でauto memory（`MEMORY.md` とメモリファイル）に登録する。**登録するだけ**で、Notionへの実際の挿入とメモリの削除は `/report` 側が行う。

## ワークフロー

### 1. 内容の確認

引数があればそれを使う。無ければユーザーに「何を残しておきたいか」を確認する。

### 2. メモリファイルの作成

保存先は`worklog`のauto memoryディレクトリ（`~/.config/claude/projects/-Users-s23159-poc-worklog/memory/`）に固定する。現在の作業ディレクトリに関わらずここへ書く（導出規則: `worklog`リポジトリの絶対パス`/Users/s23159/poc/worklog`の`/`を`-`に置換したディレクトリ名）。

そこに、type: project のメモリファイルを新規作成する。ファイル名は `report_note_<slug>.md`（`<slug>`は内容から生成するkebab-case、英数字とハイフンのみ）。

```yaml
---
name: report-note-<slug>
description: <一行要約>
metadata:
  type: project
---
```

本文構成:

```markdown
<伝えたい内容そのもの。関連リンクがあれば含める>

**Why:** 次回の/report実行時にNotionレポートへ挿入するため（記録日: <今日の日付>）。
**How to apply:** /report Step 2.5で読み取り、該当プロジェクト配下へ箇条書きで挿入する。挿入後はこのメモリファイルとMEMORY.mdのエントリを削除する。
```

### 3. MEMORY.mdへのポインタ追加

`worklog`のauto memoryディレクトリの `MEMORY.md`（`~/.config/claude/projects/-Users-s23159-poc-worklog/memory/MEMORY.md`）を**絶対パスでRead**してから編集する。他プロジェクトのセッション中はこのファイルがシステムプロンプトに載らないため、読まずに見出し判定をしない。

- 見出し `## MTGで伝えたいこと` / `## 次回レポート挿入事項` / `## 次回report時の挿入メッセージ` のいずれかが既に存在すれば、その見出しをそのまま使う（表記を混在させない）。
- どれも存在しなければ `## MTGで伝えたいこと` を新規作成する。

その見出しの下に1行追加する（既存の他のエントリは消さない）:

```markdown
- [<一行要約>](report_note_<slug>.md) — <補足>
```

### 4. 完了報告

何を・どのファイルに保存したか、次回`/report`実行時に消費されて自動削除される旨を一言で報告する。

## 使用例

```
/report-note cos-haproxyのstate wasabi化とCI/CD整備が完了した(PR: https://github.com/cycloud-io/cos-haproxy-monitoring/pull/15)
```

→ `worklog`のauto memoryディレクトリに `report_note_cos-haproxy-wasabi.md` を作成し、`MEMORY.md`の「## MTGで伝えたいこと」に1行追加する。呼び出し元の作業ディレクトリ（`kot`など）は問わない。

## 注意事項

- 1メモリファイル1トピック。複数の伝達事項があるときはファイルを分ける。
- `/report` skillが検索する見出し表記（「MTGで伝えたいこと」「次回レポート挿入事項」「次回report時の挿入メッセージ」）以外の見出しを使わない。
- ここで書くのは「登録」のみ。Notionへの実際の挿入・メモリの削除は `/report` 側が行う。
- 保存先は`worklog`のauto memoryディレクトリに固定。現在の作業ディレクトリのauto memoryへは書かない。

## 関連スキル

- `worklog`リポジトリの `/report` - 稼働時間集計とNotionレポート作成（このスキルで登録した内容をStep 2.5で読み取り消費する）
