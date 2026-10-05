# task-board

プロジェクトの `.claude/tasks/*.json` を読み、タスクの一覧をプロンプト上の帯に表示する Claude Code の mod。
表示するだけで、Task ツールの登録やプロンプトの変更はしない。エージェントの動作には影響しない。

複数セッションに作業を割り振る統括セッションで、どのタスクが進行中・待機中・完了かを人が確認する用途を想定している。

## 表示

- 帯: 見出し(`N tasks (… done, … in progress, … open)`)と、未完了の上位5件。残りは `… +N pending`。
- `/board`: 完了を含む全件を Pane に開く。もう一度 `/board` か Esc で閉じる。
- `ctrl+x t`: `/board` と同じ開閉。mod 側にキー登録 API が無いため、帯と Pane の Button の `action` に未使用のエンジンアクション `app:toggleDiffNoiseFilter` を結び、`~/.config/claude/keybindings.json` の Global にそのキーを割り当てている。
- 色と記号は標準のタスク一覧に合わせてある(`◼` 進行中、`◻` 未着手、`✔` 完了)。

## データ

セッションの作業ディレクトリの `.claude/tasks/` に、1リスト=1ファイルで置く。3秒ごとに読み直す。ファイルが複数あるときは、Pane でファイル名ごとにまとめる。

`session` には、その一覧を持つセッションの ID の先頭8桁を含む文字列を入れる(例: `session-0f692c4e`)。帯と `/board` には、現在のセッションの ID の先頭8桁を `session` に含む一覧だけを出す。

- 現在のセッションに一致する一覧が1つでもあれば、一致しない一覧と `session` の無い一覧は表示しない。`/board` の見出しに「このセッションの一覧だけ表示中」と出る。
- 一致する一覧が1つも無ければ、絞り込まずに全件を合算して出す。
- 現在のセッションの ID は `$.session.id()` を3秒ごとの読み直しのたびに取る。`/clear` で ID が変わると、次の読み直しで絞り込み先も変わる。

```json
{
  "updated_at": "2026-10-04T22:00+09:00",
  "session": "session-0f692c4e",
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

## サイドカー(自動更新)

`.claude/tasks/` があるプロジェクトでは、メインの回答が終わるたびに(30秒に1回まで)`$.model.fork` で会話を読み直し、`.claude/tasks/auto.json` を更新する。メインと同じモデル・同じ履歴を使い、メインのターンの終了は待たせない。

- 書くのは `auto.json` だけ。手書きのファイルには触れない。
- 変更が無い・返答が読めない・ファイルが壊れているときは何も書かない。
- `.claude/tasks/` が無いプロジェクトでは動かない(全プロジェクトにファイルを作らないため)。
- ターンごとに fork を1回呼ぶので、その分のトークンを使う。止めるには `.claude/tasks/` を消す。

## 使い方

```sh
claude --plugin-dir <このフォルダ>
```

## 検証

```sh
npx -y -p typescript tsc -p .   # 型チェック
claude plugin validate .        # 構文
claude plugin test .            # tests/*.test.ts(status の分類と補足)
```

## 制約

- 帯の高さにはエンジン側の上限がある。畳んだ表示は未完了の上位5件に限り、全件は `/board` の Pane で見る。
- `ctrl+t` による表示・非表示の切り替えには対応していない。帯を隠すにはエンジン標準の `[-]` を使う。
