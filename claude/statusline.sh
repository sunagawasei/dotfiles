#!/usr/bin/env bash

input=$(cat)

# jq失敗時も既存バーを描画できるようデフォルト値を設定
MODEL="?"
DIR=""
USED_PCT=0
ADDED=0
REMOVED=0
RATE_USED=""
SESSION_ID=""

# jqで一括抽出
eval "$(printf '%s' "$input" | jq -r '
  @sh "MODEL=\(.model.display_name // "?")",
  @sh "DIR=\(.workspace.current_dir // "")",
  @sh "USED_PCT=\(.context_window.used_percentage // 0)",
  @sh "ADDED=\(.cost.total_lines_added // 0)",
  @sh "REMOVED=\(.cost.total_lines_removed // 0)",
  @sh "RATE_USED=\(.rate_limits.five_hour.used_percentage // "")",
  @sh "SESSION_ID=\(.session_id // "")"
' 2>/dev/null)" 2>/dev/null

# モデル名を短縮（例: "claude-sonnet-4-6" → "sonnet", "Claude Sonnet 4.6" → "sonnet"）
MODEL_SHORT=$(echo "$MODEL" | sed -E 's/[Cc]laude[- ]+//g; s/[- ]*[0-9]+(\.[0-9]+)*//g; s/[- ]+$//; s/^ +//; s/ +$//' | tr '[:upper:]' '[:lower:]')
[ -n "$MODEL_SHORT" ] && MODEL="$MODEL_SHORT"

# ディレクトリ名
DIR_NAME="${DIR##*/}"
[ -z "$DIR_NAME" ] && DIR_NAME="~"

# Gitブランチ
GIT_BRANCH=""
if cd "$DIR" 2>/dev/null && git rev-parse --is-inside-work-tree &>/dev/null; then
  GIT_BRANCH=$(git branch --show-current 2>/dev/null)
  [ -z "$GIT_BRANCH" ] && GIT_BRANCH="HEAD:$(git rev-parse --short HEAD 2>/dev/null)"
fi

# herdr用busyマーカーの状態判定
AGMSG_TEAM=""
CODEX_PENDING=0
if [[ -n "$SESSION_ID" && "$SESSION_ID" =~ ^[0-9a-fA-F-]+$ ]]; then
  AGMSG_TEAM="s-${SESSION_ID}"
  AGMSG_DB=~/.agents/skills/agmsg/db/messages.db
  if [ -r "$AGMSG_DB" ] && command -v sqlite3 >/dev/null 2>&1; then
    CODEX_PENDING=$(sqlite3 -readonly "$AGMSG_DB" "
      SELECT EXISTS(
        SELECT 1 FROM messages m
        WHERE m.team='${AGMSG_TEAM}' AND m.from_agent='claude' AND m.to_agent LIKE 'codex%'
          AND m.id > COALESCE((SELECT MAX(r.id) FROM messages r
            WHERE r.team=m.team AND r.from_agent=m.to_agent AND r.to_agent='claude'), 0)
      );
    " 2>/dev/null) || CODEX_PENDING=0
    [ "$CODEX_PENDING" = "1" ] || CODEX_PENDING=0
  fi
fi

# psは1回だけ走査し、bridge生存とClaude配下のバックグラウンドタスクを導出
BRIDGE_ALIVE=0
BG_BUSY=0
PROCESS_STATE=$(ps -ax -o pid=,ppid=,command= 2>/dev/null | awk \
  -v self_pid="$$" -v team="$AGMSG_TEAM" '
  {
    pid = $1 + 0
    ppid = $2 + 0
    command = $0
    sub(/^[[:space:]]*[0-9]+[[:space:]]+[0-9]+[[:space:]]*/, "", command)
    parents[pid] = ppid
    commands[pid] = command
    pids[pid] = 1
  }
  END {
    bridge_signature = "codex" "-bridge"
    snapshot_signature = "shell-snapshots/" "snapshot-"
    watch_signature = "agmsg/scripts/" "watch.sh"
    bridge_alive = 0
    bg_busy = 0

    if (team != "") {
      for (pid in pids) {
        if (index(commands[pid], bridge_signature) && index(commands[pid], team)) {
          bridge_alive = 1
          break
        }
      }
    }

    current = self_pid + 0
    while (current > 0 && !visited[current]) {
      visited[current] = 1
      ancestors[current] = 1
      first_token = commands[current]
      sub(/[[:space:]].*$/, "", first_token)
      sub(/^.*\//, "", first_token)
      if (first_token == "claude") {
        claude_pid = current
        break
      }
      current = parents[current]
    }

    if (claude_pid > 0) {
      for (pid in pids) {
        if (parents[pid] == claude_pid && !ancestors[pid] &&
            index(commands[pid], snapshot_signature)) {
          candidates[pid] = 1
        }
      }

      for (candidate in candidates) {
        is_monitor = 0
        for (pid in pids) {
          if (!index(commands[pid], watch_signature)) {
            continue
          }
          current = pid
          for (depth = 0; current > 0 && depth < 128; depth++) {
            current = parents[current]
            if (current == candidate) {
              is_monitor = 1
              break
            }
          }
          if (is_monitor) {
            break
          }
        }
        if (!is_monitor) {
          bg_busy = 1
          break
        }
      }
    }

    print bridge_alive, bg_busy
  }
' 2>/dev/null) || PROCESS_STATE=""
read -r BRIDGE_ALIVE BG_BUSY <<< "$PROCESS_STATE"
[ "$BRIDGE_ALIVE" = "1" ] || BRIDGE_ALIVE=0
[ "$BG_BUSY" = "1" ] || BG_BUSY=0

RST='\e[0m'

# Powerline文字（UTF-8 octalエスケープ）
L_CAP=$(printf '\356\202\266')   # U+E0B6 左丸キャップ
R_CAP=$(printf '\356\202\264')   # U+E0B4 右丸キャップ
SEG_SEP=$(printf '\356\202\260') # U+E0B0 右向き三角（セクション遷移）

# hex→ANSIエスケープ文字列変換（printf '%b' で解釈される \e[...m 形式を返す）
hex2bg() {
  local h="${1:1}"
  printf '\\e[48;2;%d;%d;%dm' $((16#${h:0:2})) $((16#${h:2:2})) $((16#${h:4:2}))
}
hex2fg() {
  local h="${1:1}"
  printf '\\e[38;2;%d;%d;%dm' $((16#${h:0:2})) $((16#${h:2:2})) $((16#${h:4:2}))
}

# セクション本体テキスト用前景色
C_MODEL="\e[38;2;206;245;242m"   # #CEF5F2 foregrounds.main
C_DIR="\e[38;2;157;220;217m"     # #9DDCD9 foregrounds.heading
C_GIT="\e[38;2;108;216;211m"     # #6CD8D3 teals.bright
C_ADD="\e[38;2;52;149;148m"      # #349594 teals.deep
C_DEL="\e[38;2;163;122;167m"     # #a37aa7 purples.bright_purple
C_BUSY="\e[38;2;163;122;167m"    # #a37aa7 purples.bright_purple

# 使用率の色（閾値で変化）
pct=${USED_PCT%.*}
pct=${pct:-0}
if [ "$pct" -gt 75 ]; then
  C_PCT="\e[38;2;163;122;167m"   # #a37aa7
elif [ "$pct" -gt 50 ]; then
  C_PCT="\e[38;2;206;213;233m"   # #CED5E9
else
  C_PCT="\e[38;2;52;149;148m"    # #349594
fi

# リミット残量の色
if [ -n "$RATE_USED" ]; then
  rate_used_int=${RATE_USED%.*}
  rate_used_int=${rate_used_int:-0}
  rate_remaining=$((100 - rate_used_int))
  if [ "$rate_remaining" -lt 25 ]; then
    C_RATE="\e[38;2;163;122;167m"   # #a37aa7 警告（purple）
  elif [ "$rate_remaining" -lt 50 ]; then
    C_RATE="\e[38;2;206;213;233m"   # #CED5E9 注意（グレー）
  else
    C_RATE="\e[38;2;52;149;148m"    # #349594 安全（teal）
  fi
fi

# Powerlineバーを構築するヘルパー関数
build_bar() {
  local -n _segs=$1
  local n=${#_segs[@]}
  local out=""
  for i in "${!_segs[@]}"; do
    local bg_hex="${_segs[$i]%%|*}"
    local content="${_segs[$i]#*|}"
    local BG_CUR FG_CUR
    BG_CUR=$(hex2bg "$bg_hex")
    FG_CUR=$(hex2fg "$bg_hex")

    if [ "$i" -eq 0 ]; then
      out+="${RST}${FG_CUR}${L_CAP}"
    else
      local prev_bg_hex="${_segs[$((i-1))]%%|*}"
      local FG_PREV
      FG_PREV=$(hex2fg "$prev_bg_hex")
      out+="${BG_CUR}${FG_PREV}${SEG_SEP}"
    fi
    out+="${BG_CUR} ${content} "
  done
  local last_bg_hex="${_segs[$((n-1))]%%|*}"
  local FG_LAST
  FG_LAST=$(hex2fg "$last_bg_hex")
  out+="${RST}${FG_LAST}${R_CAP}${RST}"
  printf '%b' "$out"
}

# --- 1段目: モデル / コンテキスト使用率 / レート制限残量 ---
row1=()
row1+=("#1F3451|${C_MODEL}${MODEL}")
row1+=("#2B2D32|${C_PCT}󰍛 ${pct}%")
if [ -n "$RATE_USED" ]; then
  row1+=("#1E2A3A|${C_RATE}󰔛 ${rate_remaining}%")
fi

# --- 2段目: ディレクトリ / Gitブランチ / 行変更 ---
row2=()
row2+=("#152A2B|${C_DIR}${DIR_NAME}")
[ -n "$GIT_BRANCH" ] && row2+=("#1E1E24|${C_GIT}${GIT_BRANCH}")
if [ "$ADDED" -gt 0 ] || [ "$REMOVED" -gt 0 ]; then
  lines=""
  [ "$ADDED" -gt 0 ]                               && lines+="${C_ADD}+${ADDED}"
  [ "$ADDED" -gt 0 ] && [ "$REMOVED" -gt 0 ]       && lines+=" "
  [ "$REMOVED" -gt 0 ]                             && lines+="${C_DEL}-${REMOVED}"
  row2+=("#44363B|${lines}")
fi

[ "$CODEX_PENDING" = "1" ] && [ "$BRIDGE_ALIVE" = "1" ] && row2+=("#1E1E24|${C_BUSY}󰚩 codex")
[ "$BG_BUSY" = "1" ] && row2+=("#1E1E24|${C_BUSY}󰜎 bg")

build_bar row1
printf '\n'
build_bar row2
