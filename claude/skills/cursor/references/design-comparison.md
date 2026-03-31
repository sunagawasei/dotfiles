# 設計比較テンプレート（Plan / Ask Mode）

cursor の `--mode plan` または `--mode ask` を使って設計を比較するテンプレート。

## 2アプローチ比較（Plan Mode）

```bash
agent -p --model composer-2 --mode plan --trust "
Design comparison:

Task: {task_description}

Approach A: {approach_a_description}
Approach B: {approach_b_description}

Evaluate each approach on:
1. Correctness and completeness
2. Performance characteristics (time/space complexity, I/O patterns)
3. Maintainability and readability
4. Extensibility for future changes
5. Testing complexity
6. Risk and failure modes

Codebase context:
{relevant_code_snippets_or_file_paths}

Provide:
- Structured comparison table
- Final recommendation with rationale
- Implementation notes for the recommended approach

DO NOT ask questions. State assumptions clearly.
IMPORTANT: Read-only mode — analysis only.
"
```

## ライブラリ・フレームワーク選定（Ask Mode）

```bash
agent -p --model composer-2 --mode ask --trust "
Library evaluation for {use_case}:

Candidates:
- {library_a}: {brief_description}
- {library_b}: {brief_description}

Evaluate on:
1. API ergonomics and learning curve
2. Performance characteristics
3. Maintenance status and community
4. License compatibility
5. Integration complexity with existing stack: {current_stack}

Recommend the best option with specific rationale.
"
```

## アルゴリズム選択（Ask Mode）

```bash
agent -p --model composer-2 --mode ask --trust "
Algorithm selection for {problem_description}:

Current approach: {current_algorithm}
Alternative: {alternative_algorithm}

Compare on:
1. Time complexity for typical input size ({expected_size})
2. Space complexity
3. Implementation complexity
4. Edge cases and failure modes

Given the constraints: {constraints}
Which should we use and why?
"
```
