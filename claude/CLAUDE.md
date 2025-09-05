# YOU MUST:

- 回答は日本語で行ってください

# Golang Guidelines

## Version

- Always use the latest stable version of Go , best practices, and Go idioms.
- go の standard library を優先して使用してください

## Code Design

- テスト容易性を高めるために Dependency Injection (DI) を用いた実装を行う
- サードパーティライブラリは使用しない

## Testing

- テストは明示的に指示された場合のみ生成する
- `t.Run()`を用いたサブテストの形でテストケースを列挙する
- テーブルテストは基本的には用いず

# TDD (Test-Driven Development) 基本方針

## Red-Green-Refactorサイクル

- **[RED]** → 失敗するテストを先に書く
- **[GRN]** → テストを通す最小限の実装
- **[REF]** → コードの品質を改善

## TDD表示規則

- 進捗状態を `[RED]` `[GRN]` `[REF]` で明示
- 各フェーズの目的を明確に説明
- テストファースト原則を厳守

## 開発プラクティス

### TDD（テスト駆動開発）の適用
- **RED-GREEN-REFACTORサイクル**を厳守
  - [RED]: 失敗するテストを先に書く
  - [GRN]: テストを通す最小限の実装
  - [REF]: コードの品質を改善
- **統合テストは実装より先に修正**して、テストファーストを維持
- リファクタリング時はテストから修正を開始

### ドキュメント管理
- **進捗の随時更新**: TODO-*.mdファイルは作業完了時に即座に更新
- **実装とドキュメントの同期**: コード変更とドキュメントを同時に更新
- **完了タスクの明示**: 完了したタスクは「完了済み」セクションへ移動

### テストカバレッジ管理

#### テスト実行コマンド
```bash
# 全テスト実行（カバレッジ付き）
go test ./... -cover

# 詳細なカバレッジレポート生成
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 特定パッケージのテスト
go test -v ./internal/package_name/

# PRD環境テストの実行（要設定ファイル）
go test -v -run "_PRD" ./...
```
- output-styleに沿っているか確認して。