# Copilot CLI - Test Generation Template

## Purpose

Use Copilot CLI to suggest comprehensive test specifications and examples.

> **Note**: The reviewer-only role constraint is applied automatically via
> `~/.config/claude/skills/copilot/instructions/reviewer-only.instructions.md`.

## Usage

```bash
copilot -p "
Test generation suggestion task:

Suggest comprehensive tests for this code:
1. Unit test structure and test cases
2. Edge cases to cover
3. Mock/stub requirements
4. Test utilities needed
5. Example test code

Provide detailed test specifications with code examples.

Code to test:
{code}

Testing framework: {framework}
Language: {language}
" --no-ask-user -s
```

## Example 1: Function Test Suggestions

```bash
copilot -p "
Test suggestion task:

Suggest tests for this Lua function:

\`\`\`lua
local function validate_color(hex)
  if not hex then return false end
  if type(hex) ~= 'string' then return false end
  return hex:match('^#[0-9A-Fa-f]{6}$') ~= nil
end
\`\`\`

Provide:
1. Test case list (happy path, edge cases, errors)
2. Expected inputs and outputs
3. Example test code using plenary.nvim
4. Additional test scenarios to consider
" --no-ask-user -s
```

## Example 2: Module Test Plan

```bash
copilot -p "
Test plan task:

Create a test plan for this module:

File: nvim/lua/utils/color-validator.lua
Functions:
- validate_hex(color)
- convert_to_rgb(hex)
- calculate_contrast(color1, color2)

Suggest:
1. Test file structure
2. Setup/teardown requirements
3. Test cases for each function
4. Integration test scenarios
5. Example test implementations

Test framework: plenary.nvim (Neovim)

Provide comprehensive test specifications.
" --no-ask-user -s
```

## Example 3: Integration Test Suggestions

```bash
copilot -p "
Integration test suggestions:

Suggest integration tests for:

Feature: Colorscheme loading and application
Components:
- Color config loader
- Theme applier
- Highlight group setter

Suggest:
1. Integration test scenarios
2. Mock requirements (if any)
3. Assertions to verify
4. Test data needed
5. Example integration test code

Framework: plenary.nvim

Provide test design with examples.
" --no-ask-user -s
```

## Tips

- Be specific about the testing framework
- Provide context about the code's purpose
- Mention any existing test patterns to follow
- Ask for prioritized test cases
- 複雑なロジックには /codex スキルでテスト設計レビューを実施
