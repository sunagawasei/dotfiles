# YOU MUST:

- 回答は日本語で行ってください
- Answer in Japanese even if the reply is in English.
- コード生成時は、明示的に指定がない限りGolangを使用してください（Pythonは使用しない）

# General Guidelines

## ドキュメントシンプル化方針

- 実例・コード例は不要
- 本質的な情報のみ記載

# TDD (Test-Driven Development)

## 基本原則：RED-GREEN-REFACTORサイクル

- **[RED]** → 失敗するテストを先に書く
- **[GRN]** → テストを通す最小限の実装
- **[REF]** → コードの品質を改善

## 表示規則

- 進捗状態を `[RED]` `[GRN]` `[REF]` で明示
- 各フェーズの目的を明確に説明
- テストファースト原則を厳守
- test caseは1つずつ, red->green->blueで回す

@guidelines/dockerfile.md
