---
name: report-note
description: 次回の/reportでNotionレポートへ挿入する内容をwork-managementのreport-notesへ保存する
---

# report-note スキル

work-management repoの`/report` skill（Step 2.5）が読み取るメモを`docs/report-notes/`に保存する。次回`/report`実行時にNotionページの該当プロジェクト配下へ挿入し、反映後は消化済み記録へ更新する。

`/report`はwork-management repoのプロジェクト固有スキルである。このスキルはどの作業ディレクトリから呼び出しても`/Users/s23159/work/work-management/docs/report-notes/`へ保存する。

## 概要

「これも次のMTGで伝えたい」「次のレポートに入れといて」と言われた内容を、`/report` skillが読み取れる形式で`docs/report-notes/`に保存する。**保存するだけ**で、Notionへの実際の挿入と消化済み状態への更新は`/report`側が行う。

## ワークフロー

### 1. 内容の確認

引数があればそれを使う。無ければユーザーに「何を残しておきたいか」を確認する。

### 2. report-noteファイルの作成

保存先は`/Users/s23159/work/work-management/docs/report-notes/`。現在の作業ディレクトリに関わらず、このディレクトリに`report_note_<slug>.md`を作成する。ファイル名は内容からkebab-caseで生成する。

```markdown
---
name: report-note-<slug>
description: <一行要約>
status: pending
---

<伝えたい内容そのもの。関連リンクがあれば含める>

**Why:** 次回の/report実行時にNotionレポートへ挿入するため（記録日: <今日の日付>）。
**How to apply:** /report Step 2.5で読み取り、該当プロジェクト配下へ箇条書きで挿入する。挿入後はstatusをconsumedにして内容を短い消化済み記録へ置き換える。
```

### 3. READMEへのポインタ追加

`/Users/s23159/work/work-management/docs/report-notes/README.md`に、作成したファイルへの相対リンクを追加する。既存の項目は消さない。

```markdown
- [<一行要約>](report_note_<slug>.md)
```

### 4. 完了報告

何をどの`report_note_<slug>.md`へ保存したか、次回`/report`で消費されstatusが更新されることを報告する。

## 使用例

```
/report-note cos-haproxyのstate wasabi化とCI/CD整備が完了した(PR: https://github.com/cycloud-io/cos-haproxy-monitoring/pull/15)
```

→ `/Users/s23159/work/work-management/docs/report-notes/report_note_cos-haproxy-wasabi.md`を作成し、同じディレクトリのREADMEに相対リンクを追加する。呼び出し元の作業ディレクトリは問わない。

## 注意事項

- 1メモファイル1トピック。複数の伝達事項があるときはファイルを分ける。
- ファイルの`status: pending`を維持し、/reportが消費した後に`status: consumed`へ更新する。
- ここではメモを保存する。Notionへの挿入と消化済み更新は`/report` 側が行う。
- 保存先はwork-management repo内の`docs/report-notes/`に固定する。auto memoryや現在の作業ディレクトリには保存しない。

## 関連スキル

- work-management repoの`/report` - 稼働時間集計とNotionレポート作成（このスキルで登録した内容をStep 2.5で読み取り消費する）
