# Copilot CLI - Error Analysis Template

## Purpose

Use Copilot CLI to diagnose errors and provide fix recommendations.

> **Note**: The reviewer-only role constraint is applied automatically via
> `~/.config/claude/skills/copilot/instructions/reviewer-only.instructions.md`.

## Usage

```bash
copilot -p "
Error analysis task:

Analyze this error and provide:
1. Root cause identification
2. Step-by-step fix instructions
3. Code examples for the fix
4. Prevention strategies for the future

Error message:
{error_output}

Stack trace:
{stack_trace}

Relevant code:
{code_context}

Recent changes:
{git_diff}
" --no-ask-user -s
```

## Example 1: Runtime Error

```bash
# --effort high 推奨（extended thinking対応モデルのみ。非対応なら省略）
copilot -p "
Error analysis task:

Analyze this runtime error:

Error: attempt to index a nil value (local 'config')
Stack traceback:
  nvim/lua/plugins/colorscheme.lua:15: in function 'setup'
  nvim/init.lua:42: in main chunk

Code context:
\`\`\`lua
local function setup()
  local colors = config.colors  -- Line 15
  vim.cmd('colorscheme ' .. colors.theme)
end
\`\`\`

Diagnose:
1. Why is config nil?
2. How to fix it properly?
3. How to prevent this in the future?

Provide fix with code examples.
" --no-ask-user -s --effort high
```

## Example 2: Build/Compile Error

```bash
copilot -p "
Build error analysis:

Build failed with:
{build_output}

Relevant files:
{file_list}

Recent changes:
$(git diff HEAD~1)

Identify:
1. What's causing the build failure?
2. Which file/line is the problem?
3. Specific fix needed
4. Related issues to check

Provide diagnostic with fix examples.
" --no-ask-user -s
```

## Example 3: Plugin Error

```bash
copilot -p "
Plugin error analysis:

Plugin loading error:
{error_message}

Plugin config:
{plugin_config}

Neovim version: {nvim_version}

Diagnose:
1. Is this a config issue or plugin bug?
2. Version compatibility problems?
3. Missing dependencies?
4. Correct configuration

Provide solution with corrected config example.
" --no-ask-user -s
```

## Example 4: Performance Issue

```bash
copilot -p "
Performance issue analysis:

Symptom: Neovim startup is slow (3+ seconds)

Profiling data:
{profile_output}

Loaded plugins:
$(nvim --startuptime startup.log && cat startup.log)

Analyze:
1. Which plugins are slow?
2. Configuration causing delays?
3. Optimization opportunities
4. Recommended changes

Provide performance fixes with examples.
" --no-ask-user -s
```

## Integration Workflow

```
1. Error occurs → Copy error message
2. Run Copilot analysis → Get diagnosis
3. Claude Code implements fix
4. Verify error is resolved
5. Optional: Add prevention measures
```

## Tips

- Include full error messages and stack traces
- Provide recent git changes (git diff)
- Add relevant configuration context
- Mention environment details (versions, OS)
- 複雑なエラーには /codex スキルでより深い原因分析を実施
