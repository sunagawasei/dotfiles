# Copilot CLI - Code Review Template

## Purpose

Use Copilot CLI to perform comprehensive code reviews.

> **Note**: The reviewer-only role constraint is applied automatically via
> `~/.config/claude/skills/copilot/instructions/reviewer-only.instructions.md`.

## Usage

```bash
copilot -p "
Code review task:

Review the following code for:
1. Bugs and potential issues (with line references)
2. Best practice violations
3. Performance concerns
4. Security vulnerabilities
5. Code style and readability issues

Provide specific fixes with code examples for each issue found.

File: {file_path}

Code:
{code_to_review}

Context:
{relevant_context}
" --no-ask-user -s
```

## Example 1: Full File Review

```bash
copilot -p "
Code review task:

Review this entire file for quality issues:

File: nvim/lua/plugins/lualine.lua

$(cat nvim/lua/plugins/lualine.lua)

Focus on:
1. Configuration correctness
2. Performance optimizations
3. Lua best practices
4. Plugin integration issues

Provide specific recommendations with code examples.
" --no-ask-user -s
```

## Example 2: Function Review

```bash
copilot -p "
Code review task:

Review this specific function:

\`\`\`lua
local function setup_colorscheme()
  -- Configuration here
end
\`\`\`

Check for:
1. Logic errors
2. Edge cases not handled
3. Performance issues
4. Better implementation patterns

Provide concrete improvements with code examples.
" --no-ask-user -s
```

## Example 3: Security-Focused Review

```bash
# --effort high 推奨（extended thinking対応モデルのみ。非対応なら省略）
copilot -p "
Security review task:

Review this code for security vulnerabilities:

{code}

Specifically check for:
1. Input validation issues
2. Injection vulnerabilities
3. Insecure defaults
4. Sensitive data exposure
5. Authentication/authorization flaws

Provide security recommendations with secure code examples.
" --no-ask-user -s --effort high
```

## Example 4: Pre-Commit Review

```bash
# --effort low 推奨（速度重視。extended thinking対応モデルのみ。非対応なら省略）
copilot -p "
Pre-commit review task:

Review these staged changes before commit:

$(git diff --cached)

Identify:
1. Obvious bugs or issues
2. Missing error handling
3. Breaking changes
4. Code quality concerns

Provide quick feedback with specific line references.
" --no-ask-user -s --effort low
```

## Integration with Claude Code

1. Run Copilot review → Get recommendations
2. Claude Code implements the suggested changes
3. Optional: Run Copilot again to verify fixes

## Tips

- Be specific about what to review (security, performance, etc.)
- Provide sufficient context for better analysis
- Use for both full files and specific functions
- /codex スキルと組み合わせてより深いアーキテクチャレビューを実施
