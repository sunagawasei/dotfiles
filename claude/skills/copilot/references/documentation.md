# Copilot CLI - Documentation Template

## Purpose

Use Copilot CLI to suggest documentation improvements and content.

> **Note**: The reviewer-only role constraint is applied automatically via
> `~/.config/claude/skills/copilot/instructions/reviewer-only.instructions.md`.

## Usage

```bash
copilot -p "
Documentation suggestion task:

Suggest documentation improvements for this code:
1. Missing docstrings/comments
2. API documentation structure
3. Usage examples needed
4. Parameter descriptions
5. Return value documentation
6. Example documentation content

Provide comprehensive documentation recommendations.

Code:
{code}

Documentation format: {format}
Target audience: {audience}
" --no-ask-user -s
```

## Example 1: Function Documentation

```bash
copilot -p "
Documentation task:

Suggest docstring for this Lua function:

\`\`\`lua
local function setup_colorscheme(config)
  local colors = load_colors(config.palette)
  apply_highlights(colors)
  if config.transparent then
    make_transparent()
  end
  return colors
end
\`\`\`

Provide:
1. Function description
2. Parameter documentation
3. Return value description
4. Usage example
5. Complete docstring in LuaDoc format
" --no-ask-user -s
```

## Example 2: Module Documentation

```bash
copilot -p "
Module documentation task:

Suggest documentation for this module:

File: nvim/lua/utils/color-validator.lua

Exported functions:
- validate_hex(color)
- convert_to_rgb(hex)
- calculate_contrast(color1, color2)
- is_accessible(fg, bg)

Suggest:
1. Module overview and purpose
2. Installation/usage instructions
3. Function documentation for each export
4. Code examples
5. Common use cases
6. Complete module doc structure

Format: Markdown with LuaDoc comments

Provide comprehensive module documentation.
" --no-ask-user -s
```

## Tips

- Specify the documentation format (LuaDoc, JSDoc, etc.)
- Mention the target audience (users, contributors, etc.)
- Provide context about the code's purpose
- Ask for examples to be included
- 大規模ドキュメントタスクには /gemini スキルでベストプラクティスをリサーチ
- ドキュメント構造レビューには /codex スキルでアーキテクチャ観点を確認
