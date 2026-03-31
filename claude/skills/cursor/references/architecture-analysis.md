# アーキテクチャ分析テンプレート（Plan Mode）

cursor の `--mode plan` を使ってアーキテクチャを分析するテンプレート。

## 基本コマンド

```bash
agent -p --model composer-2 --mode plan --trust "
Architecture analysis task:

Analyze the architecture of {component_path} and provide:
1. Component structure and responsibility mapping
2. Dependency direction analysis (do dependencies point inward toward the domain?)
3. Interface design evaluation (are interfaces well-defined and minimal?)
4. Coupling and cohesion assessment
5. Specific improvement recommendations with file paths and concrete steps

DO NOT ask questions. State assumptions clearly.
IMPORTANT: You are in plan mode (read-only). Provide analysis and recommendations only.
Claude Code will implement the actual changes.
"
```

## 依存関係分析

```bash
agent -p --model composer-2 --mode plan --trust "
Dependency analysis:

Analyze module dependencies in {directory_path}:
1. Map all import/dependency relationships between modules
2. Identify circular dependencies (list them explicitly)
3. Identify tight coupling (modules that change together frequently)
4. Identify low cohesion (modules with unrelated responsibilities)
5. Propose dependency restructuring with specific file-by-file steps

Output: Dependency graph description + prioritized refactoring plan.
IMPORTANT: Read-only mode — analysis only.
"
```

## モジュール間インターフェース評価

```bash
agent -p --model composer-2 --mode plan --trust "
Interface evaluation:

Evaluate the interfaces between modules in {path}:
1. List all public interfaces (functions, types, methods)
2. Identify which interfaces are used only internally vs externally
3. Identify unnecessary public exposure
4. Suggest encapsulation improvements
5. Identify missing abstractions

IMPORTANT: Read-only mode — analysis only.
"
```
