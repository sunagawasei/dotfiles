# 実装計画テンプレート（Plan Mode）

cursor の `--mode plan` を使って実装計画を生成するテンプレート。

## 基本コマンド

```bash
agent -p --model composer-2 --mode plan --trust "
Implementation planning task:

Feature/change to implement: {feature_description}

Codebase context:
- Relevant files: {file_paths}
- Current behavior: {current_behavior}
- Desired behavior: {desired_behavior}

Provide a detailed implementation plan including:
1. High-level approach and key design decisions
2. File-by-file change plan (files to create / modify / delete)
3. Step-by-step implementation sequence with dependency ordering
4. Risk assessment and mitigation strategies
5. Testing strategy (unit, integration, e2e)

Output format: Numbered steps with specific file paths and code snippets.
DO NOT ask questions. State assumptions clearly.
IMPORTANT: You are in plan mode (read-only). Provide the plan only.
Claude Code will implement the actual changes.
"
```

## 新機能実装計画

```bash
agent -p --model composer-2 --mode plan --trust "
New feature implementation plan:

Feature: {feature_name}
Description: {description}

Analyze the existing codebase structure at {path} and provide:
1. Where the feature fits in the existing architecture
2. New files to create (with proposed structure)
3. Existing files to modify (with specific changes)
4. Implementation order (what must be built first)
5. Interface/API design if applicable
6. Potential breaking changes

Constraints: {constraints}
DO NOT ask questions. State assumptions clearly.
IMPORTANT: Read-only mode — provide plan only.
"
```

## リファクタリング計画

```bash
agent -p --model composer-2 --mode plan --trust "
Refactoring plan:

Target: {component_or_file_path}
Goal: {refactoring_goal}

Provide:
1. Current structure analysis (what exists and why it's problematic)
2. Target structure (what it should look like)
3. Migration steps in safe, incremental order
4. Files affected and specific changes per file
5. How to verify correctness at each step (tests/checks)
6. Rollback strategy if something goes wrong

IMPORTANT: Read-only mode — provide plan only.
"
```
