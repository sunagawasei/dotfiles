---
name: codex
context: fork
allowed-tools: Bash, Read, Glob, Grep
description: |
  Execute Codex CLI directly for expert code analysis. Codex is a read-only analyst —
  Claude Code handles all implementation. Use when the user requests:
  - Code review or PR review ("review", "レビュー")
  - Bug / root cause analysis ("debug", "analyze", "デバッグ", "原因調査")
  - Architecture review ("設計", "アーキテクチャ", "design advice")
  - Security audit ("セキュリティ監査", "脆弱性チェック")
  - Performance optimization analysis ("パフォーマンス", "最適化")
  - Refactoring proposal ("リファクタ", "refactor")
  - Explicit triggers: "consult codex", "/codex"
  Do NOT use for: typo fixes, single-line edits, git operations, or tasks with one obvious solution.
---

# Codex System — Code Review & Analysis Expert

**Codex CLI provides direct, real-time code analysis in read-only mode.**

## Security Notice

Codex sends code to external OpenAI API. Do NOT use with secrets, credentials, or proprietary code unless explicitly authorized.

## How to Execute

```bash
codex exec --sandbox read-only "
[Autonomous prompt with clear task and context]
"
```

**Notes:**
- Use `--sandbox read-only` — omit `--full-auto` (it implies workspace-write)
- Omit `-m <model>` to use CLI default (latest). Override with `-m <model>` if needed
- Execute prompts in **English**, translate results to **Japanese** for user

### Autonomous Prompt Design

**❌ AVOID:**
```
"Review this code. Any questions?"
```

**✅ RECOMMENDED:**
```
"Review this code and provide:
1. Specific issues (with line references)
2. Concrete fixes (with code examples)
3. Security/performance concerns

DO NOT ask questions. State assumptions clearly.
CRITICAL: You are in read-only mode. Provide suggestions only.
Claude Code will implement the actual changes."
```

## Session Management

```bash
# Resume a previous session
codex exec resume <SESSION_ID>

# Fork from a session
codex exec fork <SESSION_ID>

# List recent sessions
codex sessions list
```

## Language Protocol

1. Execute Codex in **English**
2. Receive analysis in **English**
3. Translate findings to **Japanese** for user
4. Claude Code implements changes based on recommendations

## Error Handling & Fallback

| Error | Resolution |
|-------|------------|
| Auth failure | Run `codex login`, then retry |
| Rate limit | Wait 60s, retry (use `-m <model>` to switch model) |
| Sandbox permission denied | Expected in read-only — do NOT switch to workspace-write |
| Network timeout | Check VPN/proxy, retry |
| Git repo required | Add `--skip-git-repo-check` flag |

If Codex is unavailable: use `/copilot` for quick review or Claude Code directly.

## Limitations

- **Read-only**: Cannot write files, run tests, or execute code
- **Context limit**: Large codebases may need scoped file paths
- **External API**: Requires internet access and valid auth

## Reference Templates

Read as needed:

- `references/code-review-task.md` — Review prompt templates and `codex exec review` usage
- `references/refactoring-task.md` — Refactoring proposal workflow
- `references/delegation-patterns.md` — Decision flowchart: when to delegate, patterns by use case
- `references/agent-prompts.md` — Specialized templates (Architect / Analyzer / Optimizer / Security) — read when user requests a specific analysis type
- `references/troubleshooting.md` — Install, auth, config, and error fixes
