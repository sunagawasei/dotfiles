#!/usr/bin/env bash

input=$(cat)

# jq失敗時も既存バーを描画できるようデフォルト値を設定
MODEL="?"
DIR=""
USED_PCT=0
RATE_USED=""
SESSION_ID=""

# jqで一括抽出
eval "$(printf '%s' "$input" | jq -r '
  @sh "MODEL=\(.model.display_name // "?")",
  @sh "DIR=\(.workspace.current_dir // "")",
  @sh "USED_PCT=\(.context_window.used_percentage // 0)",
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
CODEX_BUSY=0
CODEX_BRIDGE_PIDS=()
CODEX_BRIDGE_NAMES=()
CODEX_BRIDGE_LOGS=()
if [[ -n "$SESSION_ID" && "$SESSION_ID" =~ ^[0-9a-fA-F-]+$ ]]; then
  AGMSG_TEAM="s-${SESSION_ID}"
  AGMSG_RUN_DIR=/Users/s23159/.agents/skills/agmsg/run
  for metafile in "$AGMSG_RUN_DIR"/codex-bridge."$AGMSG_TEAM".*.meta; do
    [ -f "$metafile" ] || continue

    meta_team=""
    meta_type=""
    while IFS='=' read -r key value; do
      case "$key" in
        team) meta_team="$value" ;;
        type) meta_type="$value" ;;
      esac
    done < "$metafile"
    [ "$meta_team" = "$AGMSG_TEAM" ] && [ "$meta_type" = "codex" ] || continue

    bridge_name=${metafile#"$AGMSG_RUN_DIR/codex-bridge.${AGMSG_TEAM}."}
    bridge_name=${bridge_name%.meta}
    [ -n "$bridge_name" ] || continue

    pidfile=${metafile%.meta}.pid
    bridge_pid=""
    [ -r "$pidfile" ] && IFS= read -r bridge_pid < "$pidfile"
    [[ "$bridge_pid" =~ ^[0-9]+$ ]] || continue
    kill -0 "$bridge_pid" 2>/dev/null || continue

    CODEX_BRIDGE_PIDS+=("$bridge_pid")
    CODEX_BRIDGE_NAMES+=("$bridge_name")
    CODEX_BRIDGE_LOGS+=("${metafile%.meta}.log")
  done
fi

# psは1回だけ走査し、bridgeの実体確認とClaude配下のバックグラウンドタスクに再利用
BG_BUSY=0
PROCESS_SNAPSHOT=$(ps -ww -ax -o pid=,ppid=,command= 2>/dev/null) || PROCESS_SNAPSHOT=""
PROCESS_STATE=$(printf '%s\n' "$PROCESS_SNAPSHOT" | awk \
  -v self_pid="$$" '
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
    snapshot_signature = "shell-snapshots/" "snapshot-"
    watch_signature = "agmsg/scripts/" "watch.sh"
    bg_busy = 0

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

    print bg_busy
  }
' 2>/dev/null) || PROCESS_STATE=""
read -r BG_BUSY <<< "$PROCESS_STATE"
[ "$BG_BUSY" = "1" ] || BG_BUSY=0

# 実行中のサブエージェントが残したマーカーを同じbusy表示へ合流
if [[ -n "$SESSION_ID" && "$SESSION_ID" =~ ^[0-9a-fA-F-]+$ ]]; then
  SUBAGENT_RUN_DIR=/Users/s23159/.config/claude/run
  for marker in "$SUBAGENT_RUN_DIR"/subagent."$SESSION_ID".*; do
    [ -f "$marker" ] || continue
    marker_pid=""
    IFS= read -r marker_pid < "$marker"
    [[ "$marker_pid" =~ ^[0-9]+$ ]] || continue

    while read -r process_pid process_ppid process_command; do
      [ "$process_pid" = "$marker_pid" ] || continue
      command_token=${process_command%%[[:space:]]*}
      if [ "${command_token##*/}" = "claude" ]; then
        BG_BUSY=1
        break
      fi
    done <<< "$PROCESS_SNAPSHOT"
    [ "$BG_BUSY" = "1" ] && break
  done
fi

# meta/pidに対応するbridgeの実体をpsスナップショットで確認し、最後のlifecycle行を読む
for index in "${!CODEX_BRIDGE_PIDS[@]}"; do
  bridge_pid=${CODEX_BRIDGE_PIDS[$index]}
  bridge_name=${CODEX_BRIDGE_NAMES[$index]}
  logfile=${CODEX_BRIDGE_LOGS[$index]}
  bridge_command=""
  while read -r process_pid process_ppid process_command; do
    if [ "$process_pid" = "$bridge_pid" ]; then
      bridge_command="$process_command"
      break
    fi
  done <<< "$PROCESS_SNAPSHOT"

  padded_command=" $bridge_command "
  [[ "$bridge_command" == *"codex-bridge.js"* ]] || continue
  [[ "$padded_command" == *" --team ${AGMSG_TEAM} "* ]] || continue
  [[ "$padded_command" == *" --name ${bridge_name} "* ]] || continue
  [ -r "$logfile" ] || continue

  lifecycle_state=$(tail -n 400 "$logfile" 2>/dev/null | awk \
    -v identity="${AGMSG_TEAM}/${bridge_name}" '
    BEGIN {
      wakeup_prefix = "codex-bridge: wakeup "
      wakeup_suffix = " for " identity
      armed_line = "codex-bridge: armed " identity
      state = 0
    }
    $0 == armed_line {
      state = 0
      next
    }
    index($0, wakeup_prefix) == 1 &&
        substr($0, length($0) - length(wakeup_suffix) + 1) == wakeup_suffix {
      wakeup_number = substr($0, length(wakeup_prefix) + 1,
        length($0) - length(wakeup_prefix) - length(wakeup_suffix))
      if (wakeup_number ~ /^[0-9]+$/) {
        state = 1
      }
    }
    END {
      print state
    }
  ' 2>/dev/null) || lifecycle_state=0
  if [ "$lifecycle_state" = "1" ]; then
    CODEX_BUSY=1
    break
  fi
done

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

# BEGIN GENERATED COLORS: ANSI
# セクション本体テキスト用前景色
C_MODEL="\e[38;2;205;233;245m"   # #CDE9F5 foregrounds.main
C_DIR="\e[38;2;136;203;234m"     # #88CBEA foregrounds.heading
C_GIT="\e[38;2;88;202;248m"     # #58CAF8 teals.bright
C_BUSY="\e[38;2;176;122;224m"    # #b07ae0 purples.bright_purple

# 使用率の色（閾値で変化）
pct=${USED_PCT%.*}
pct=${pct:-0}
if [ "$pct" -gt 75 ]; then
  C_PCT="\e[38;2;176;122;224m"   # #b07ae0
elif [ "$pct" -gt 50 ]; then
  C_PCT="\e[38;2;208;212;240m"   # #D0D4F0
else
  C_PCT="\e[38;2;46;101;137m"    # #2E6589
fi

# リミット残量の色
if [ -n "$RATE_USED" ]; then
  rate_used_int=${RATE_USED%.*}
  rate_used_int=${rate_used_int:-0}
  rate_remaining=$((100 - rate_used_int))
  if [ "$rate_remaining" -lt 25 ]; then
    C_RATE="\e[38;2;176;122;224m"   # #b07ae0 警告（purple）
  elif [ "$rate_remaining" -lt 50 ]; then
    C_RATE="\e[38;2;208;212;240m"   # #D0D4F0 注意（グレー）
  else
    C_RATE="\e[38;2;46;101;137m"    # #2E6589 安全（teal）
  fi
fi

# END GENERATED COLORS: ANSI
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

# BEGIN GENERATED COLORS: SEGMENTS
# --- 1段目: モデル / コンテキスト使用率 / レート制限残量 / codex・bgマーカー ---
row1=()
row1+=("#1F3265|${C_MODEL}${MODEL}")
row1+=("#252A40|${C_PCT}󰍛 ${pct}%")
if [ -n "$RATE_USED" ]; then
  row1+=("#30314B|${C_RATE}󰔛 ${rate_remaining}%")
fi
[ "$CODEX_BUSY" = "1" ] && row1+=("#191D2B|${C_BUSY}󰚩")
[ "$BG_BUSY" = "1" ] && row1+=("#191D2B|${C_BUSY}󰜎")

# --- 2段目: ディレクトリ / Gitブランチ ---
row2=()
row2+=("#102337|${C_DIR}${DIR_NAME}")
[ -n "$GIT_BRANCH" ] && row2+=("#191D2B|${C_GIT}${GIT_BRANCH}")

# END GENERATED COLORS: SEGMENTS
build_bar row1
printf '\n'
build_bar row2
