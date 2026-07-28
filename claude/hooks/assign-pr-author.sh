#!/usr/bin/env bash
# PostToolUse(Bash) hook — assign the PR author (@me) on `gh pr create`.
#
# Fires after every Bash tool call but exits immediately unless the command was
# a `gh pr create`. On a create, it pulls the new PR URL out of the tool result
# (anywhere in the JSON payload) and runs `gh pr edit <url> --add-assignee @me`.
# Only covers `gh pr create` run through Claude Code's Bash tool — a PR created
# in a separate terminal does not trigger this hook.
input=$(cat)

# Cheap bail: skip the whole thing unless "gh pr create" appears at all.
case "$input" in
  *'gh pr create'*) ;;
  *) exit 0 ;;
esac

# Precise gate: only act when the *command* itself is a create (not, say, a
# `gh pr view` whose output happens to mention the phrase).
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // ""' 2>/dev/null)
case "$cmd" in
  *'gh pr create'*) ;;
  *) exit 0 ;;
esac

# Extract the created PR URL from anywhere in the payload (output field name
# varies; recursing over all strings is robust). Take the first match.
url=$(printf '%s' "$input" | jq -r '.. | strings' 2>/dev/null \
  | grep -oE 'https://github\.com/[^ "]+/pull/[0-9]+' | head -n1)
[ -n "$url" ] || exit 0

if gh pr edit "$url" --add-assignee @me >/dev/null 2>&1; then
  printf '{"systemMessage":"Assigned %s to you (@me)"}\n' "$url"
fi
exit 0
