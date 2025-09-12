#!/usr/bin/env bash

# Read JSON input from stdin
input=$(cat)

MODEL_DISPLAY=$(echo "$input" | jq -r '.model.display_name')
CURRENT_DIR=$(echo "$input" | jq -r '.workspace.current_dir')
TRANSCRIPT_PATH=$(echo "$input" | jq -r '.transcript_path')

# Get git branch information
GIT_BRANCH=""
if git rev-parse &>/dev/null; then
  BRANCH=$(git branch --show-current)
  if [ -n "$BRANCH" ]; then
    GIT_BRANCH=" |  $BRANCH"
  else
    COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
    if [ -n "$COMMIT_HASH" ]; then
      GIT_BRANCH=" |  HEAD ($COMMIT_HASH)"
    fi
  fi
fi

# Get token summary
if [ -z "$TRANSCRIPT_PATH" ] || [ ! -f "$TRANSCRIPT_PATH" ]; then
  TOKEN_COUNT="_%"
else
  # Get last assistant message with usage data using jq
  total_tokens=$(tail -n 100 "$TRANSCRIPT_PATH" 2>/dev/null |
    jq -s 'map(select(.type == "assistant" and .message.usage)) |
    last |
    .message.usage |
    (.input_tokens // 0) +
    (.output_tokens // 0) +
    (.cache_creation_input_tokens // 0) +
  (.cache_read_input_tokens // 0)' 2>/dev/null)

  # Default to 0 if no valid result
  total_tokens=${total_tokens:-0}

  # max token count: 200k
  # compaction threshold: 80% (160k)
  COMPACTION_THRESHOLD=160000
  # Calculate percentage
  percentage=$((total_tokens * 100 / COMPACTION_THRESHOLD))


  # Color coding for percentage
  if [ "$percentage" -ge 90 ]; then
    color="\033[38;2;239;68;68m" # Geist Red 6
  elif [ "$percentage" -ge 70 ]; then
    color="\033[38;2;245;158;11m" # Geist Amber 6
  else
    color="\033[38;2;34;197;94m" # Geist Green 6
  fi

  # Format: "10%"
  TOKEN_COUNT=$(echo -e "${color}${percentage}%\033[0m")
fi

echo "󰚩 ${MODEL_DISPLAY} |  ${CURRENT_DIR##*/}${GIT_BRANCH} |  ${TOKEN_COUNT}"

