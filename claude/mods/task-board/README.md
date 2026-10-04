# task-board

プロジェクトの `.claude/tasks/*.json` を読み、タスクの一覧をプロンプト上の帯に表示する Claude Code の mod。
表示するだけで、Task ツールの登録やプロンプトの変更はしない。エージェントの動作には影響しない。

複数セッションに作業を割り振る統括セッションで、どのタスクが進行中・待機中・完了かを人が確認する用途を想定している。

## 表示

- 帯: 見出し(`N tasks (… done, … in progress, … open)`)と、未完了の上位5件。残りは `… +N pending`。
- `/board`: 完了を含む全件を Pane に開く。もう一度 `/board` か Esc で閉じる。
- 色と記号は標準のタスク一覧に合わせてある(`◼` 進行中、`◻` 未着手、`✔` 完了)。

## データ

セッションの作業ディレクトリの `.claude/tasks/` に、1リスト=1ファイルで置く。3秒ごとに読み直す。ファイルが複数あるときは、Pane でファイル名ごとにまとめる。

```json
{
  "updated_at": "2026-10-04T22:00+09:00",
  "tasks": [
    { "id": "T1", "title": "短いタイトル", "status": "implementing", "ws": "w2" },
    { "id": "T2", "title": "別のタイトル", "status": "waiting (T1後)", "ws": null }
  ]
}
```

`status` の先頭の語で分類する。

| 先頭の語 | 分類 |
| :- | :- |
| `implementing` / `fixing` / `running` | 進行中 |
| `merged` / `done` | 完了 |
| それ以外(`waiting` / `queued` など) | 未着手 |

先頭の語より後ろ(例: `(T1後)`)と `ws` は、行の右に薄字で添える。

ファイルは Claude に書かせる。CLAUDE.md には、たとえば次のように書く。

```markdown
作業の割り振りが変わったら `.claude/tasks/<名前>.json` を更新する(形式は task-board の README)。
```

## 使い方

```sh
claude --plugin-dir <このフォルダ>
```

## 制約

- 帯の高さにはエンジン側の上限がある。畳んだ表示は未完了の上位5件に限り、全件は `/board` の Pane で見る。
- `ctrl+t` による表示・非表示の切り替えには対応していない。帯を隠すにはエンジン標準の `[-]` を使う。
