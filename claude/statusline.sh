#!/usr/bin/env bash

input=$(cat)

# jqで一括抽出
eval "$(echo "$input" | jq -r '
  @sh "MODEL=\(.model.display_name // "?")",
  @sh "DIR=\(.workspace.current_dir // "")",
  @sh "USED_PCT=\(.context_window.used_percentage // 0)",
  @sh "ADDED=\(.cost.total_lines_added // 0)",
  @sh "REMOVED=\(.cost.total_lines_removed // 0)"
')"

# ディレクトリ名
DIR_NAME="${DIR##*/}"
[ -z "$DIR_NAME" ] && DIR_NAME="~"

# Gitブランチ
GIT_BRANCH=""
if cd "$DIR" 2>/dev/null && git rev-parse --is-inside-work-tree &>/dev/null; then
  GIT_BRANCH=$(git branch --show-current 2>/dev/null)
  [ -z "$GIT_BRANCH" ] && GIT_BRANCH="HEAD:$(git rev-parse --short HEAD 2>/dev/null)"
fi

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

# セグメント: "bg_hex|content" 形式
segs=()
segs+=("#1F3451|${C_MODEL}${MODEL}")
segs+=("#152A2B|${C_DIR}${DIR_NAME}")
[ -n "$GIT_BRANCH" ] && segs+=("#1E1E24|${C_GIT}${GIT_BRANCH}")
segs+=("#2B2D32|${C_PCT}${pct}%")

# 行変更セグメント（条件付き）
if [ "$ADDED" -gt 0 ] || [ "$REMOVED" -gt 0 ]; then
  lines=""
  [ "$ADDED" -gt 0 ]   && lines+="${C_ADD}+${ADDED}"
  [ "$ADDED" -gt 0 ] && [ "$REMOVED" -gt 0 ] && lines+=" "
  [ "$REMOVED" -gt 0 ] && lines+="${C_DEL}-${REMOVED}"
  segs+=("#44363B|${lines}")
fi

# 出力構築: 左キャップ → セクション遷移ループ → 右キャップ
n=${#segs[@]}
out=""

for i in "${!segs[@]}"; do
  bg_hex="${segs[$i]%%|*}"
  content="${segs[$i]#*|}"
  BG_CUR=$(hex2bg "$bg_hex")
  FG_CUR=$(hex2fg "$bg_hex")

  if [ "$i" -eq 0 ]; then
    # 左丸キャップ: 背景なし、前景=最初のセクション背景色
    out+="${RST}${FG_CUR}${L_CAP}"
  else
    # Powerline遷移: 前セクション背景色が前景、現セクション背景色が背景
    prev_bg_hex="${segs[$((i-1))]%%|*}"
    FG_PREV=$(hex2fg "$prev_bg_hex")
    out+="${BG_CUR}${FG_PREV}${SEG_SEP}"
  fi

  # セクション本体
  out+="${BG_CUR} ${content} "
done

# 右丸キャップ: reset後、前景=最後のセクション背景色
last_bg_hex="${segs[$((n-1))]%%|*}"
FG_LAST=$(hex2fg "$last_bg_hex")
out+="${RST}${FG_LAST}${R_CAP}${RST}"

printf '%b' "$out"
