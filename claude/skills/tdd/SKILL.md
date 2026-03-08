---
name: tdd
description: "Use PROACTIVELY when implementing new features, designing code architecture, or writing tests. Apply RED-GREEN-REFACTOR cycle to build quality code incrementally."
---

# TDD (Test-Driven Development)

テスト駆動開発のベストプラクティスに従い、RED-GREEN-REFACTORサイクルで開発を進める。

## 基本原則：RED-GREEN-REFACTORサイクル

- **[RED]** → 失敗するテストを先に書く
- **[GRN]** → テストを通す最小限の実装
- **[REF]** → コードの品質を改善

## 表示規則

- 進捗状態を `[RED]` `[GRN]` `[REF]` で明示
- 各フェーズの目的を明確に説明
- テストファースト原則を厳守
- test caseは1つずつ、red->green->blueで回す

## ワークフロー

1. **RED**: テストを書いて失敗させる
2. **GREEN**: 最小限の実装でテストを通す
3. **REFACTOR**: コードを改善（テストは通ったまま）
4. 次のテストケースへ
