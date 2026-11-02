---
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
---

# Go言語規約

公式 [Effective Go](https://go.dev/doc/effective_go) と [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) を基準に、慣用的でクリーンなGoを書く。

## プロジェクト方針

- **Goバージョン**: 常に最新安定版を使用
- **標準ライブラリ優先**: stdlibを優先し、サードパーティライブラリは基本的に使わない
- **依存性注入**: テスト容易性を高めるためDIパターンを採用
- **テスト生成**: 明示的に指示された場合のみテストを生成する

## プロジェクト構造

- `go.mod`: モジュール定義 / `go.sum`: 依存関係のチェックサム
- テストファイルは `*_test.go` パターンに従う

## テスト規約

- テスト関数は `Test*`、ベンチマークは `Benchmark*` で始める
- サブテストは `t.Run()` 形式を使用する（テーブル駆動テストは必須ではない）
- 確認は `go test` のみでよい。`go test` は内部でビルドを行うため、`go build` を別途実行するのは冗長で不要
- repoを変えずに一時的なテストやログ出力を足すときは、`go test -overlay`/`go build -overlay`でscratchpadのファイルを差し込む。このとき`-count=1`を付ける。2026-11-02、scratchpadの入力ファイルを差し替えても前回のPASSがキャッシュから返り、原因を一度見誤った
