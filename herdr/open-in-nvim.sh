#!/bin/sh

set -u

usage() {
  echo "Usage: $0 <file>" >&2
}

warn() {
  echo "open-in-nvim: $*" >&2
}

absolute_path() {
  case $1 in
    /*) path=$1 ;;
    *) path=$PWD/$1 ;;
  esac

  parent=${path%/*}
  [ -n "$parent" ] || parent=/
  name=${path##*/}
  if normalized_parent=$(cd "$parent" 2>/dev/null && pwd -P); then
    path=$normalized_parent/$name
  fi

  printf '%s\n' "$path"
}

shell_quote() {
  printf "'%s'" "$(printf '%s' "$1" | sed "s/'/'\\\\''/g")"
}

resolve_nvim_socket() {
  pid=$1
  tmpdir=${TMPDIR:-/tmp}
  socket_dir=${tmpdir%/}/nvim.$(id -un)
  child_pids=""

  # 単一プロセス形態とnvim 0.12以降のTUI→embed core形態の両方を探す。
  socket=$(find "$socket_dir" -type s -name "nvim.$pid.0" 2>/dev/null | head -n 1)
  if [ -z "$socket" ] && command -v pgrep >/dev/null 2>&1; then
    child_pids=$(pgrep -P "$pid" -x nvim 2>/dev/null) || child_pids=""
    socket=$(printf '%s\n' "$child_pids" | while IFS= read -r child_pid; do
      [ -n "$child_pid" ] || continue
      child_socket=$(find "$socket_dir" -type s -name "nvim.$child_pid.0" 2>/dev/null | head -n 1)
      if [ -n "$child_socket" ]; then
        printf '%s\n' "$child_socket"
        break
      fi
    done)
  fi

  if [ -z "$socket" ] && command -v lsof >/dev/null 2>&1; then
    socket=$({
      printf '%s\n' "$pid"
      [ -n "$child_pids" ] && printf '%s\n' "$child_pids"
    } | while IFS= read -r socket_pid; do
      [ -n "$socket_pid" ] || continue
      lsof_socket=$(lsof -p "$socket_pid" 2>/dev/null | awk -v needle="/nvim.$socket_pid.0" '
        tolower($5) == "unix" && index($0, needle) {
          for (i = 1; i <= NF; i++) {
            if (index($i, needle)) {
              print $i
              exit
            }
          }
        }
      ')
      if [ -n "$lsof_socket" ]; then
        printf '%s\n' "$lsof_socket"
        break
      fi
    done)
  fi

  printf '%s\n' "$socket"
}

focus_pane() {
  pane_id=$1
  request=$(jq -nc \
    --arg id "open-in-nvim-focus-$$" \
    --arg pane_id "$pane_id" \
    '{id: $id, method: "pane.focus", params: {pane_id: $pane_id}}')

  if response=$(printf '%s\n' "$request" | /usr/bin/nc -U "$HERDR_SOCKET_PATH" 2>/dev/null) &&
      printf '%s\n' "$response" | jq -e 'type == "object" and has("result")' >/dev/null 2>&1; then
    return 0
  fi

  warn "nvim pane $pane_id へのフォーカスに失敗しました"
  return 1
}

close_popup() {
  request=$(jq -nc \
    --arg id "open-in-nvim-close-$$" \
    '{id: $id, method: "popup.close", params: {}}')
  printf '%s\n' "$request" | /usr/bin/nc -U "$HERDR_SOCKET_PATH" >/dev/null
}

file=${1-}
if [ -z "$file" ]; then
  usage
  exit 1
fi
if [ -z "${HERDR_ACTIVE_WORKSPACE_ID:-}" ] || [ -z "${HERDR_SOCKET_PATH:-}" ]; then
  warn "herdr popup外では実行できません"
  exit 1
fi

herdr_bin=${HERDR_BIN_PATH:-herdr}
active_pane_id=${HERDR_ACTIVE_PANE_ID:-}
abs_path=$(absolute_path "$file")
nvim_pane_id=""
nvim_pid=""

if pane_list_json=$("$herdr_bin" pane list --workspace "$HERDR_ACTIVE_WORKSPACE_ID"); then
  if ! pane_ids=$(printf '%s\n' "$pane_list_json" | jq -r \
    --arg active "$active_pane_id" '
      (.result.panes // []) as $panes
      | (($panes | map(select(.pane_id == $active)))
        + ($panes | map(select(.pane_id != $active))))
      | .[].pane_id
    '); then
    warn "pane listのJSONを解析できませんでした"
    pane_ids=""
  fi
else
  warn "workspace $HERDR_ACTIVE_WORKSPACE_ID のpane一覧を取得できませんでした"
  pane_ids=""
fi

for pane_id in $pane_ids; do
  if ! process_info_json=$("$herdr_bin" pane process-info --pane "$pane_id"); then
    warn "pane $pane_id のプロセス情報を取得できませんでした"
    continue
  fi
  if ! pid=$(printf '%s\n' "$process_info_json" | jq -r '
    [
      .result.process_info.foreground_processes[]?
      | select(
          (.name // "") == "nvim"
          or (((.argv0 // "") | split("/") | last) == "nvim")
        )
    ]
    | first
    | .pid // empty
  '); then
    warn "pane $pane_id のプロセスJSONを解析できませんでした"
    continue
  fi
  case $pid in
    ''|*[!0-9]*) continue ;;
  esac
  nvim_pane_id=$pane_id
  nvim_pid=$pid
  break
done

opened_existing=0
if [ -n "$nvim_pane_id" ]; then
  socket=$(resolve_nvim_socket "$nvim_pid")
  if [ -n "$socket" ] && nvim --server "$socket" --remote-expr '1' >/dev/null 2>&1; then
    if ! nvim --server "$socket" --remote "$abs_path"; then
      warn "$abs_path を既存nvimで開けませんでした"
    fi
    focus_pane "$nvim_pane_id" || :
    opened_existing=1
  fi
fi

if [ "$opened_existing" -ne 1 ]; then
  if [ -z "$active_pane_id" ]; then
    warn "HERDR_ACTIVE_PANE_IDが設定されていません"
    exit 1
  fi
  if ! split_json=$("$herdr_bin" pane split "$active_pane_id" \
    --direction right --cwd "$PWD" --focus); then
    warn "nvim用paneを作成できませんでした"
    exit 1
  fi
  if ! new_pane_id=$(printf '%s\n' "$split_json" | jq -r '.result.pane.pane_id // empty') ||
      [ -z "$new_pane_id" ]; then
    warn "作成したpaneのIDを取得できませんでした"
    exit 1
  fi
  sleep 0.5
  quoted_path=$(shell_quote "$abs_path")
  if ! "$herdr_bin" pane run "$new_pane_id" "nvim $quoted_path"; then
    warn "pane $new_pane_id でnvimを起動できませんでした"
    exit 1
  fi
fi

# popup.closeはpopup内の全プロセスを終了させるため、必ず最後に送る。
close_popup
exit 0
