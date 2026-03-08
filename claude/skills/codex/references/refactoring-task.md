# Codex System - Refactoring Task Template

## Purpose

Use Codex CLI for refactoring analysis and recommendations.

## Usage

```bash
codex exec --sandbox read-only "
Refactoring proposal task:

Analyze this code and propose refactoring:
1. Identify code smells and complexity issues
2. Provide specific refactoring steps
3. Include before/after code examples
4. Explain benefits and risks

DO NOT ask for approval. Present concrete refactoring plan with rationale.

IMPORTANT: You are in read-only mode. Propose refactoring only.
Claude Code will implement the refactoring.

Code to refactor:
{code}

Goals:
{refactoring_goals}
"
```

## Example 1: Simplification Refactoring

```bash
codex exec --sandbox read-only "
Simplification refactoring task:

Refactor this complex function:

$(cat nvim/lua/config/autocmds.lua | grep -A 50 'function complex_handler')

Goals:
- Reduce complexity
- Improve readability
- Extract reusable parts

Provide:
1. Complexity analysis
2. Refactoring steps
3. Before/after examples
4. Benefits explanation

You are proposing only - Claude Code will refactor.
"
```

## Example 2: Pattern Application

```bash
codex exec --sandbox read-only "
Pattern refactoring task:

Apply design pattern to this code:

Current implementation:
{current_code}

Suggested pattern: Strategy Pattern

Provide:
1. Why this pattern fits
2. Refactored structure
3. Implementation steps
4. Code examples

Proposal only - Claude Code will implement.
"
```

## Integration Workflow

1. Codex refactoring analysis → Detailed plan
2. Gemini pattern research → Best practices (optional)
3. Claude Code → Implementation

## Tips

- Be specific about refactoring goals
- Provide context about constraints
- Ask for risk assessment
- Request step-by-step plan
