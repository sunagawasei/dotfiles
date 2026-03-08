# Codex System - Code Review Task Template

## Purpose

Use Codex CLI for deep code review with architectural and design insights.

## Usage

```bash
codex exec --sandbox read-only "
Code review task:

Review the following code and provide:
1. Specific issues found (with line references)
2. Concrete improvement suggestions (with code examples)
3. Security/performance concerns
4. Best practice violations
5. Architectural considerations

DO NOT ask questions. Make reasonable assumptions and state them clearly.
Provide actionable recommendations immediately.

IMPORTANT: You are in read-only mode. Provide suggestions only.
Claude Code will implement the actual changes.

Context:
{relevant_code_or_file_path}

Code:
{code_to_review}
"
```

## Quick Review (git-integrated)

Use `codex exec review` for git-based diffing:

```bash
# Review uncommitted changes
codex exec review --uncommitted

# Review changes against main branch
codex exec review --base main

# Review a specific commit
codex exec review --commit <SHA>
```

## Example: Full File Review

```bash
codex exec --sandbox read-only "
Code review task:

Review this complete Neovim plugin configuration:

File: nvim/lua/plugins/lualine.lua

$(cat nvim/lua/plugins/lualine.lua)

Analyze:
1. Configuration structure and organization
2. Performance implications
3. Lua best practices adherence
4. Plugin integration patterns
5. Potential runtime issues

Provide detailed review with specific improvements.
Include code examples for recommended changes.

Remember: Analysis only, Claude Code will implement changes.
"
```

## Integration with Other Tools

- **Copilot CLI**: Quick, tactical reviews
- **Gemini CLI**: Research best practices for domain
- **Claude Code**: Implementation of all changes
