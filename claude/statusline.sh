#!/usr/bin/env bash

input=$(cat)

# jq失敗時も既存バーを描画できるようデフォルト値を設定
MODEL="?"
DIR=""
USED_PCT=0
RATE_USED=""
RATE_RESET=""
WEEK_USED=""
WEEK_RESET=""
SESSION_ID=""

# jqで一括抽出
eval "$(printf '%s' "$input" | jq -r '
  @sh "MODEL=\(.model.display_name // "?")",
  @sh "DIR=\(.workspace.current_dir // "")",
  @sh "USED_PCT=\(.context_window.used_percentage // 0)",
  @sh "RATE_USED=\(.rate_limits.five_hour.used_percentage // "")",
  @sh "RATE_RESET=\(.rate_limits.five_hour.resets_at // "")",
  @sh "WEEK_USED=\(.rate_limits.seven_day.used_percentage // "")",
  @sh "WEEK_RESET=\(.rate_limits.seven_day.resets_at // "")",
  @sh "SESSION_ID=\(.session_id // "")"
' 2>/dev/null)" 2>/dev/null

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
CODEX_BRIDGE_TYPES=()
if [[ -n "$SESSION_ID" && "$SESSION_ID" =~ ^[0-9a-fA-F-]+$ ]]; then
  AGMSG_TEAM="s-${SESSION_ID}"
  AGMSG_RUN_DIR=/Users/s23159/.agents/skills/agmsg/run
  # headless bridgeはtype毎に <type>-bridge.<team>.<name>.* を作る（codex/claude-code）
  for bridge_type in codex claude-code; do
    for metafile in "$AGMSG_RUN_DIR"/"$bridge_type"-bridge."$AGMSG_TEAM".*.meta; do
      [ -f "$metafile" ] || continue

      meta_team=""
      meta_type=""
      meta_pid=""
      while IFS='=' read -r key value; do
        case "$key" in
          team) meta_team="$value" ;;
          # 新形式: identities=<team>/<name>（team= 行が無い meta への追従）
          identities) [ -z "$meta_team" ] && meta_team="${value%%/*}" ;;
          type) meta_type="$value" ;;
          pid) meta_pid="$value" ;;
        esac
      done < "$metafile"
      [ "$meta_team" = "$AGMSG_TEAM" ] && [ "$meta_type" = "$bridge_type" ] || continue

      bridge_name=${metafile#"$AGMSG_RUN_DIR/${bridge_type}-bridge.${AGMSG_TEAM}."}
      bridge_name=${bridge_name%.meta}
      [ -n "$bridge_name" ] || continue

      # pidfileだけ消えてmetaが残る経路があるためmetaのpid=へfallbackする
      pidfile=${metafile%.meta}.pid
      bridge_pid=""
      [ -r "$pidfile" ] && IFS= read -r bridge_pid < "$pidfile"
      [[ "$bridge_pid" =~ ^[1-9][0-9]*$ ]] || bridge_pid="$meta_pid"
      [[ "$bridge_pid" =~ ^[1-9][0-9]*$ ]] || continue
      # sandbox下のEPERMは「生きているがsignal不可」。ESRCHだけをdeadとみなす
      if ! kill_err=$(export LC_ALL=C; kill -0 "$bridge_pid" 2>&1); then
        case "$kill_err" in
          *[Nn]'o such process'*) continue ;;
        esac
      fi

      CODEX_BRIDGE_PIDS+=("$bridge_pid")
      CODEX_BRIDGE_NAMES+=("$bridge_name")
      CODEX_BRIDGE_LOGS+=("${metafile%.meta}.log")
      CODEX_BRIDGE_TYPES+=("$bridge_type")
    done
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
  # 観測最長17分に対して45分を確保し、生存中の誤判定を避けつつ異常終了の残留を打ち切る
  SUBAGENT_MARKER_TTL=2700
  subagent_marker_now=$(date +%s)
  for marker in "$SUBAGENT_RUN_DIR"/subagent."$SESSION_ID".*; do
    [ -f "$marker" ] || continue
    marker_mtime=$(stat -f %m "$marker" 2>/dev/null) || continue
    (( subagent_marker_now - marker_mtime > SUBAGENT_MARKER_TTL )) && continue

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
  bridge_type=${CODEX_BRIDGE_TYPES[$index]}
  bridge_command=""
  while read -r process_pid process_ppid process_command; do
    if [ "$process_pid" = "$bridge_pid" ]; then
      bridge_command="$process_command"
      break
    fi
  done <<< "$PROCESS_SNAPSHOT"

  # ps列挙が対象PIDに届かない環境（sandbox）ではargvを取れない。liveness側でESRCHは
  # 除外済みなので、ここでdead扱いにすると生きているbridgeを見失う
  if [ -n "$bridge_command" ]; then
    padded_command=" $bridge_command "
    # bridge実体はtypeで拡張子が違う（codex-bridge.js / claude-code-bridge.sh）
    [[ "$bridge_command" == *"${bridge_type}-bridge."* ]] || continue
    # 旧: --team/--name 引数、新: --identity-key base64(team\tname): のどちらかで照合
    bridge_identity_key=$(printf '%s\t%s' "$AGMSG_TEAM" "$bridge_name" | base64 | tr -d '\r\n' | tr '+/' '-_')
    if [[ "$padded_command" != *" --identity-key ${bridge_identity_key}: "* ]] &&
       ! { [[ "$padded_command" == *" --team ${AGMSG_TEAM} "* ]] &&
           [[ "$padded_command" == *" --name ${bridge_name} "* ]]; }; then
      continue
    fi
  fi
  [ -r "$logfile" ] || continue

  lifecycle_state=$(tail -n 400 "$logfile" 2>/dev/null | awk \
    -v identity="${AGMSG_TEAM}/${bridge_name}" -v btype="$bridge_type" '
    BEGIN {
      wakeup_prefix = btype "-bridge: wakeup "
      wakeup_suffix = " for " identity
      armed_line = btype "-bridge: armed " identity
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

# 使用率を固定幅の塗り部分とtrack部分に量子化
build_meter() {
  local meter_pct=$1
  local meter_width=$2
  local meter_filled=$(((meter_pct * meter_width + 50) / 100))
  local i

  METER_FILLED=""
  METER_TRACK=""
  for ((i = 0; i < meter_width; i++)); do
    if [ "$i" -lt "$meter_filled" ]; then
      METER_FILLED+="━"
    else
      METER_TRACK+="━"
    fi
  done
}

# リセットまでの残り時間を最大単位1つで整形
format_remaining() {
  local resets_at=$1
  local now remaining_seconds

  FORMATTED_REMAINING=""
  now=$(date +%s)
  if [[ "$resets_at" =~ ^[0-9]+$ ]] && [ "$resets_at" -gt "$now" ]; then
    remaining_seconds=$((resets_at - now))
    if [ "$remaining_seconds" -ge 86400 ]; then
      FORMATTED_REMAINING="$((remaining_seconds / 86400))d"
    elif [ "$remaining_seconds" -ge 3600 ]; then
      FORMATTED_REMAINING="$((remaining_seconds / 3600))h"
    else
      FORMATTED_REMAINING="$((remaining_seconds / 60))m"
    fi
  fi
}

# BEGIN GENERATED COLORS: ANSI
# セクション本体テキスト用前景色
C_MODEL="\e[38;2;205;233;245m"   # #CDE9F5 foregrounds.main
C_DIR="\e[38;2;136;203;234m"     # #88CBEA foregrounds.heading
C_GIT="\e[38;2;88;202;248m"     # #58CAF8 teals.bright
C_BUSY="\e[38;2;205;172;236m"    # #cdacec purples.bright_purple

# 使用率の色（閾値で変化）
pct=${USED_PCT%.*}
pct=${pct:-0}
if [ "$pct" -gt 75 ]; then
  C_PCT="\e[38;2;205;172;236m"   # #cdacec
elif [ "$pct" -gt 50 ]; then
  C_PCT="\e[38;2;208;212;240m"   # #D0D4F0
else
  C_PCT="\e[38;2;146;191;217m"    # #92BFD9 安全（暗いtealは帯背景に沈む）
fi

# 5時間リミット使用率とバー
rate_used_int=""
if [[ "$RATE_USED" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
  rate_used_int=$((10#${RATE_USED%%.*}))
  [ "$rate_used_int" -gt 100 ] && rate_used_int=100

  if [ "$rate_used_int" -gt 75 ]; then
    C_RATE="\e[38;2;205;172;236m"   # #cdacec critical
  elif [ "$rate_used_int" -gt 50 ]; then
    C_RATE="\e[38;2;208;212;240m"   # #D0D4F0 warning
  else
    C_RATE="\e[38;2;146;191;217m"    # #92BFD9 safe
  fi
  C_RATETRACK="\e[38;2;50;70;100m"   # #324664 track

  build_meter "$rate_used_int" 8
  rate_bar_filled=$METER_FILLED
  rate_bar_track=$METER_TRACK

  format_remaining "$RATE_RESET"
  rate_remaining=$FORMATTED_REMAINING
fi

# 週次リミット使用率とバー
week_used_int=""
if [[ "$WEEK_USED" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
  week_used_int=$((10#${WEEK_USED%%.*}))
  [ "$week_used_int" -gt 100 ] && week_used_int=100

  if [ "$week_used_int" -gt 75 ]; then
    C_WEEK="\e[38;2;205;172;236m"   # #cdacec critical
  elif [ "$week_used_int" -gt 50 ]; then
    C_WEEK="\e[38;2;208;212;240m"   # #D0D4F0 warning
  else
    C_WEEK="\e[38;2;146;191;217m"    # #92BFD9 safe
  fi
  C_WEEKTRACK="\e[38;2;50;70;100m"   # #324664 track

  build_meter "$week_used_int" 8
  week_bar_filled=$METER_FILLED
  week_bar_track=$METER_TRACK

  format_remaining "$WEEK_RESET"
  week_remaining=$FORMATTED_REMAINING
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
# --- 1段目: モデル / コンテキスト使用率 / ディレクトリ / Gitブランチ ---
row1=()
row1+=("#141B2D|${C_MODEL}${MODEL}")
row1+=("#1A2235|${C_PCT}󰍛 ${pct}%")
row1+=("#202A42|${C_DIR}${DIR_NAME}")
[ -n "$GIT_BRANCH" ] && row1+=("#141B2D|${C_GIT}${GIT_BRANCH}")

# --- 2段目: busyマーカー / 5時間リミット / 週次リミット ---
row2=()
# herdrが画面下の非空3行だけを走査するため、busyマーカーを下段先頭に置く
[ "$CODEX_BUSY" = "1" ] && row2+=("#141B2D|${C_BUSY}󰚩")
[ "$BG_BUSY" = "1" ] && row2+=("#1A2235|${C_BUSY}󰜎")
if [ -n "$rate_used_int" ]; then
  row2+=("#202A42|${C_RATE}5h ${rate_bar_filled}${C_RATETRACK}${rate_bar_track}${C_RATE} ${rate_used_int}%${rate_remaining:+ ${rate_remaining}}")
fi
if [ -n "$week_used_int" ]; then
  row2+=("#141B2D|${C_WEEK}Week ${week_bar_filled}${C_WEEKTRACK}${week_bar_track}${C_WEEK} ${week_used_int}%${week_remaining:+ ${week_remaining}}")
fi

# END GENERATED COLORS: SEGMENTS
build_bar row1
if ((${#row2[@]} > 0)); then
  printf '\n'
  build_bar row2
fi
