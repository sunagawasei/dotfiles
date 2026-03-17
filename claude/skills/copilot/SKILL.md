---
name: copilot
context: fork
allowed-tools: Bash, Read, Glob, Grep
description: |
  Execute GitHub Copilot CLI for code review, error analysis, test suggestions, and documentation recommendations.
  Copilot provides analysis and recommendations ONLY - no implementation.
  Use /copilot to trigger. Implementation is handled by Claude Code.
  Supports session continuity and autopilot mode.
  Explicit triggers: "copilot review", "copilot analyze", "copilot suggest", "test generation", "error diagnosis", "documentation review".
---

# Copilot CLI — Review & Suggestion Specialist

## Role Definition (CRITICAL)

**Copilot is a REVIEWER ONLY** - provides analysis and suggestions but does NOT implement changes. Do NOT use for deep design decisions or research; use /codex or /gemini instead.

| Tool | Role | Permission | Responsibility |
|------|------|------------|----------------|
| **Copilot CLI** | Review Specialist | Analysis only | Code review, error diagnosis, test/doc suggestions |
| **Claude Code** | Implementation | Full access | Actual code editing, file creation, test writing |

## Configuration

- **CLI**: `copilot` (must be in PATH)
- **CLI version**: `v0.0.421+`
- **Default model**: config.json で設定済み。`--model <model>` で上書き可能
- **Config**: `~/.config/.copilot/config.json`
- **Auth**: Authenticated GitHub account (via `gh auth`)
- **User instructions**: `~/.config/claude/skills/copilot/instructions/*.instructions.md`

## Security Notice

Copilot sends code snippets to external APIs for analysis. Do NOT pass sensitive data such as credentials, API keys, or proprietary secrets in prompts.

## How to Execute

### Command Pattern

```bash
copilot -p "<autonomous prompt>" --no-ask-user -s
```

**Key Options:**
- `-p` / `--prompt`: The task prompt
- `--no-ask-user`: 質問を無効化、自律的動作を強制
- `--autopilot`: 複数ステップタスクで自動継続
- `--continue`: 直前のセッション再開
- `--resume [id]`: 特定セッション再開
- `--excluded-tools`: ツール制限
- `-s` / `--silent`: Suppress interactive mode, output results only

### Autonomous Prompt Design

Prompts must be self-contained and detailed. Avoid vague instructions.

**Bad:** `"Review this code."`

**Good:**
```
"Review this code for:
1. Bugs and potential issues
2. Best practice violations
3. Performance concerns
4. Security vulnerabilities

Provide specific fixes with code examples."
```

## Error Handling & Fallback

If Copilot CLI fails (authentication error, model unavailable, network issue):

1. Show the error message to the user as-is
2. Streaming errors are auto-retried (v0.0.407+)
3. Do NOT retry manually for other errors
4. Suggest alternatives:
   - Authentication: `gh auth login`
   - Model unavailable: Try `--model <model>` (see config.json for available models)
   - Network: Check connectivity and retry later
   - MCP issues: `/mcp reload`

## Language Protocol

Copilot CLIは英語で実行し、分析結果を日本語で要約してユーザーに提示する。

## Session & Autopilot

```bash
# 直前のセッションを再開
copilot --continue -s

# 特定のセッションを再開
copilot --resume <session-id> -s

# 複数ステップの分析タスク
copilot -p "<complex analysis prompt>" --autopilot --no-ask-user -s
```

## Limitations

**What Copilot CANNOT do:** Edit files, create new files, run tests/builds, or commit changes.

**Interactive-only features (not available in `-p` mode):** `/fleet`, `/ide`, `/mcp reload`

## Reference Templates

- [Code Review](references/code-review.md)
- [Error Analysis](references/error-analysis.md)
- [Test Generation](references/test-generation.md)
- [Documentation](references/documentation.md)
