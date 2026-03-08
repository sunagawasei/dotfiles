---
name: Effective Go
description: "Apply Go best practices, idioms, and conventions from golang.org/doc/effective_go. Use when writing, reviewing, or refactoring Go code to ensure idiomatic, clean, and efficient implementations."
---

# Effective Go

Apply best practices and conventions from the official [Effective Go guide](https://go.dev/doc/effective_go) to write clean, idiomatic Go code.

## When to Apply

Use this skill automatically when:
- Writing new Go code
- Reviewing Go code
- Refactoring existing Go implementations

## Key Reminders

Follow the conventions and patterns documented at https://go.dev/doc/effective_go, with particular attention to:

- **Formatting**: Always use `gofmt` - this is non-negotiable
- **Naming**: No underscores, use MixedCaps for exported names, mixedCaps for unexported
- **Error handling**: Always check errors; return them, don't panic
- **Concurrency**: Share memory by communicating (use channels)
- **Interfaces**: Keep small (1-3 methods ideal); accept interfaces, return concrete types
- **Documentation**: Document all exported symbols, starting with the symbol name

## Project-Specific Rules

以下はこのプロジェクト固有のGo開発ルール：

- **Version**: 常に最新安定版のGoを使用
- **Standard Library優先**: 標準ライブラリを優先して使用
- **Dependency Injection**: テスト容易性を高めるためにDIパターンを採用
- **サードパーティライブラリ**: 基本的に使用しない
- **テスト生成**: 明示的に指示された場合のみテストを生成
- **テスト形式**: `t.Run()` サブテスト形式を使用（テーブルテストは非推奨）

## References

- Official Guide: https://go.dev/doc/effective_go
- Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
- Standard Library: Use as reference for idiomatic patterns
