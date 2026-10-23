#!/bin/sh
# exec が1回数秒かかる環境では本体(statusline.sh)が20秒超になり表示が間に合わない。
# 前回の出力を即返し、本体の再計算は裏で1本だけ走らせる。
dir=${XDG_CACHE_HOME:-$HOME/.cache}/claude-statusline
out=$dir/out
lock=$dir/lock

IFS= read -r -d '' in
if [ -r "$out" ]; then
  IFS= read -r -d '' prev <"$out"
  printf '%s' "$prev"
fi

# 実行中の本体があれば何もしない。死んだロックは次回に消す
if [ -e "$lock" ]; then
  IFS= read -r pid <"$lock"
  kill -0 "$pid" 2>/dev/null || rm -f "$lock"
  exit 0
fi
(set -C; : >"$lock") 2>/dev/null || exit 0
{
  printf '%s' "$in" | "$HOME/.config/claude/statusline.sh" >"$out.tmp" 2>/dev/null && mv -f "$out.tmp" "$out"
  rm -f "$lock"
} >/dev/null 2>&1 &
echo $! >|"$lock"
