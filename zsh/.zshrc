# ---- 基本設定 ----
# Emacsキーバインドを明示的に設定（$EDITORにvimを設定しているためviモードになるのを防ぐ）
bindkey -e

# コアダンプファイルを作成しない
limit coredumpsize 0

# flow control無効化（Ctrl+S/Ctrl+Qによる端末停止を防止、zeno ^x^sに必要）
stty -ixon

# ---- プラグインマネージャー (Zinit) ----
ZINIT_HOME="${XDG_DATA_HOME:-${HOME}/.local/share}/zinit/zinit.git"
if [ ! -d "$ZINIT_HOME" ]; then
  mkdir -p "$(dirname $ZINIT_HOME)"
  git clone https://github.com/zdharma-continuum/zinit.git "$ZINIT_HOME"
fi
source "$ZINIT_HOME/zinit.zsh"

# ---- 補完定義の追加コレクション ----
zinit light zsh-users/zsh-completions

# AWS CDK CLI 補完
zinit ice as"completion"
zinit snippet https://raw.githubusercontent.com/msysh/aws-cdk-zsh-completion/main/_cdk

# ---- Pure プロンプト ----
# 環境変数（読み込み前に設定が必要）
PURE_CMD_MAX_EXEC_TIME=5
PURE_GIT_PULL=1
PURE_GIT_UNTRACKED_DIRTY=1
PURE_GIT_DELAY_DIRTY_CHECK=1800
PURE_PROMPT_SYMBOL="❯"
PURE_PROMPT_VICMD_SYMBOL="❮"
PURE_GIT_DOWN_ARROW="⇣"
PURE_GIT_UP_ARROW="⇡"
PURE_GIT_STASH_SYMBOL="≡"
PURE_SUSPENDED_JOBS_SYMBOL="✦"

# zstyle動作設定
zstyle ':prompt:pure:git:stash' show yes
zstyle ':prompt:pure:git:fetch' only_upstream yes
zstyle ':prompt:pure:environment:nix-shell' show yes

# zstyleカラー設定（Abyssal Teal）
# MAIN: Main Foreground (#CEF5F2)
zstyle ':prompt:pure:path' color '#CEF5F2'
zstyle ':prompt:pure:git:arrow' color '#CEF5F2'
zstyle ':prompt:pure:git:action' color '#CEF5F2'
zstyle ':prompt:pure:prompt:success' color '#CEF5F2'
zstyle ':prompt:pure:git:branch' color '#CEF5F2'
zstyle ':prompt:pure:git:stash' color '#CEF5F2'
zstyle ':prompt:pure:suspended_jobs' color '#CEF5F2'

# TERTIARY: Comment Gray (#7A869A)
zstyle ':prompt:pure:git:branch:cached' color '#7A869A'
zstyle ':prompt:pure:git:dirty' color '#7A869A'
zstyle ':prompt:pure:execution_time' color '#7A869A'
zstyle ':prompt:pure:prompt:continuation' color '#7A869A'
zstyle ':prompt:pure:virtualenv' color '#7A869A'
zstyle ':prompt:pure:host' color '#7A869A'
zstyle ':prompt:pure:user' color '#7A869A'

# ERROR: ❯ は常に統一色。rootログイン時のみ警告色
zstyle ':prompt:pure:prompt:error' color '#CEF5F2'
zstyle ':prompt:pure:user:root' color '#A37AA7'

zinit ice compile'(pure|async).zsh' pick'async.zsh' src'pure.zsh'
zinit light sindresorhus/pure

# ---- WezTerm OSC 7 対応 ----
# WezTermに現在のディレクトリを通知（タブタイトル表示用）
if [ "$TERM_PROGRAM" = "WezTerm" ] || [ -n "$WEZTERM_PANE" ]; then
  __wezterm_osc7() {
    printf '\e]7;file://%s%s\e\\' "$HOST" "$PWD"
  }
  # シェルidle/busy状態の通知（pane閉じ確認制御用）
  __wezterm_set_user_var() {
    printf '\033]1337;SetUserVar=%s=%s\007' "$1" "$(printf '%s' "$2" | base64)"
  }
  __wezterm_preexec() { __wezterm_set_user_var WEZTERM_BUSY 1; }
  __wezterm_precmd_busy() { __wezterm_set_user_var WEZTERM_BUSY 0; }
  autoload -Uz add-zsh-hook
  add-zsh-hook precmd __wezterm_osc7
  add-zsh-hook preexec __wezterm_preexec
  add-zsh-hook precmd  __wezterm_precmd_busy
fi

# ---- 履歴設定 ----
setopt EXTENDED_HISTORY
setopt hist_ignore_all_dups
setopt hist_find_no_dups
setopt hist_ignore_space
setopt hist_verify
setopt hist_reduce_blanks
setopt hist_save_no_dups
setopt hist_no_store
setopt hist_expand
setopt share_history

# ---- キーバインド ----
bindkey '^p' history-search-backward
bindkey '^n' history-search-forward

autoload -Uz edit-command-line
zle -N edit-command-line
bindkey '^x^e' edit-command-line

# zsh-autosuggestions 部分適用
bindkey '^[f' forward-word
bindkey '^[[1;5C' forward-word
bindkey '^[[C' forward-char

# 補完用
bindkey '^y' accept-and-menu-complete
bindkey '^]' accept-and-hold

# ---- インタラクティブシェル設定 ----
unsetopt bg_nice
setopt list_packed
setopt no_beep
unsetopt list_types

# ---- 補完機能 ----
# LS_COLORS（補完候補の色設定）
export LS_COLORS='ma=48;2;100;187;190;38;2;19;32;24:di=38;2;157;220;217:ln=38;2;147;105;151:ex=38;2;108;216;211;1:*.md=38;2;82;91;101:*.txt=38;2;82;91;101:*.go=38;2;164;171;203:*.ts=38;2;164;171;203:*.js=38;2;164;171;203'

# FZF カラー設定
export FZF_DEFAULT_OPTS='
  --color=bg+:#64BBBE,bg:-1,fg:#CEF5F2,fg+:#0B0C0C
  --color=hl:#6CD8D3,hl+:#0B0C0C,info:#7A869A,marker:#0B0C0C
  --color=prompt:#936997,spinner:#936997,pointer:#0B0C0C,header:#7A869A
  --color=border:#4D8F9E,gutter:-1
'

# fzf キーバインド: Ctrl+T (ファイル選択), Alt+C (ディレクトリ移動)
# Ctrl+R は後続の zeno keybindings で上書きされる
[[ -f "${HOME}/.nix-profile/share/fzf/key-bindings.zsh" ]] && source "${HOME}/.nix-profile/share/fzf/key-bindings.zsh"

# Worklog CLI completion
fpath=(/Users/s23159/.local/share/zsh/site-functions $fpath)

autoload -U compinit && compinit

# Zoxide（compinit後に配置: compdefが実体化している必要あり）
# cdコマンドを置き換え（Claude Codeでは無効化）
if [[ "$CLAUDECODE" != "1" ]]; then
  eval "$(zoxide init --cmd cd zsh)"
fi

# fzf-tab: compinitの後に読み込み必須
zinit light Aloxaf/fzf-tab

# zsh-autosuggestions: fzf-tabの後に読み込む（fzf-tabが^Iをフックするため）
zinit light zsh-users/zsh-autosuggestions
ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=#7A869A'
ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-word)
ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-char vi-forward-char)

# Zeno.zsh: fzf-tab/autosuggestionsの後に読み込む（初期化順序の安全性のため）
zinit light yuki-yano/zeno.zsh

# Zeno keybindings
if [[ -n $ZENO_LOADED ]]; then
  # Space wrapper: zeno-auto-snippetが「success+空結果」を返した場合にスペースを補完
  # （zoxideのcd補完でSpace-Tabパターンが機能しない問題を修正）
  __zeno_space_wrapper() {
    local orig_buffer="$BUFFER"
    zle zeno-auto-snippet
    if [[ "$BUFFER" == "$orig_buffer" ]]; then
      LBUFFER+=" "
    fi
  }
  zle -N __zeno_space_wrapper
  bindkey ' '  __zeno_space_wrapper
  bindkey '^m' zeno-auto-snippet-and-accept-line
  bindkey '^r' zeno-smart-history-selection
  bindkey '^x^s' zeno-insert-snippet
fi

# zsh-syntax-highlighting: 最後に読み込む（競合リスク低減）
zinit light zsh-users/zsh-syntax-highlighting

# シンタックスハイライトの色設定 (Expanded Abyssal Teal)
typeset -A ZSH_HIGHLIGHT_STYLES
ZSH_HIGHLIGHT_STYLES[default]='fg=#CEF5F2'
ZSH_HIGHLIGHT_STYLES[arg0]='fg=#CEF5F2'
ZSH_HIGHLIGHT_STYLES[command]='fg=#6CD8D3,bold'
ZSH_HIGHLIGHT_STYLES[builtin]='fg=#6CD8D3,bold'
ZSH_HIGHLIGHT_STYLES[alias]='fg=#6CD8D3,bold'
ZSH_HIGHLIGHT_STYLES[function]='fg=#6CD8D3,bold'
ZSH_HIGHLIGHT_STYLES[unknown-token]='fg=#936997,bold'
ZSH_HIGHLIGHT_STYLES[path]='fg=#9DDCD9,underline'
ZSH_HIGHLIGHT_STYLES[single-quoted-argument]='fg=#659D9E'
ZSH_HIGHLIGHT_STYLES[double-quoted-argument]='fg=#659D9E'
ZSH_HIGHLIGHT_STYLES[single-hyphen-option]='fg=#A4ABCB'
ZSH_HIGHLIGHT_STYLES[double-hyphen-option]='fg=#A4ABCB'
ZSH_HIGHLIGHT_STYLES[comment]='fg=#7A869A'

# fzf-tab設定
zstyle ':completion:*' menu no
zstyle ':fzf-tab:complete:cd:*' disabled-on any  # zoxideとの競合を回避

# 補完スタイル
zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'
zstyle ':completion:*' list-colors "${(s.:.)LS_COLORS}"
zstyle ':completion:*' completer _complete _match

# ---- エイリアス ----
alias ls='ls --color=auto'
alias vtmp='nvim "${TMPDIR%/}/$(date "+%Y%m%d_%H%M%S").md"'
alias ssh='TERM=xterm-256color \ssh'
alias delta='delta --dark --paging=never --line-numbers --syntax-theme base16-256 -s'

# ---- Google Cloud SDK ----
# Homebrew と Nix 両方のレイアウトに対応
if command -v gcloud &>/dev/null; then
  _gcloud_bin="$(command -v gcloud)"
  _gcloud_root="$(dirname "$(readlink -f "$_gcloud_bin")")/.."
  for f in path.zsh.inc completion.zsh.inc; do
    # Nix: .../google-cloud-sdk/google-cloud-sdk/
    [[ -f "$_gcloud_root/google-cloud-sdk/$f" ]] && source "$_gcloud_root/google-cloud-sdk/$f" && continue
    # Homebrew: .../google-cloud-sdk/
    [[ -f "$_gcloud_root/$f" ]] && source "$_gcloud_root/$f"
  done
  unset _gcloud_bin _gcloud_root
fi

# ---- 外部ツール補完 ----
autoload -U +X bashcompinit && bashcompinit
complete -C "$(which aws_completer)" aws
complete -o nospace -C "$(which terraform)" terraform

# ---- ローカル設定 ----
if [ -f ~/.zshrc.local ]; then source ~/.zshrc.local; fi
