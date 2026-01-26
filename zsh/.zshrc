# ============================================================================
# Zsh Configuration
# ============================================================================

# ----------------------
# 基本設定
# ----------------------
# Emacsキーバインドを明示的に設定（$EDITORにnvimを設定しているためviモードになるのを防ぐ）
bindkey -e

# コアダンプファイルを作成しない（ディスク容量節約）
limit coredumpsize 0

# ----------------------
# プラグインマネージャー (Zinit)
# ----------------------
# Zinitのインストール先を設定（XDG Base Directory仕様に準拠）
ZINIT_HOME="${XDG_DATA_HOME:-${HOME}/.local/share}/zinit/zinit.git"

# Zinitが未インストールの場合は自動インストール
if [ ! -d "$ZINIT_HOME" ]; then
  mkdir -p "$(dirname $ZINIT_HOME)"
  git clone https://github.com/zdharma-continuum/zinit.git "$ZINIT_HOME"
fi

# Zinitを読み込み
source "$ZINIT_HOME/zinit.zsh"

# ----------------------
# Zshプラグイン
# ----------------------
# シンタックスハイライト - コマンドをリアルタイムで色付け
zinit light zsh-users/zsh-syntax-highlighting

# シンタックスハイライトの色設定（Geist primary text）
# プラグイン読み込み後に設定する
typeset -A ZSH_HIGHLIGHT_STYLES
ZSH_HIGHLIGHT_STYLES[arg0]='fg=white'

# 補完定義の追加コレクション - より多くのコマンドの補完をサポート
zinit light zsh-users/zsh-completions

# AWS CDK CLI 補完
zinit ice as"completion"
zinit snippet https://raw.githubusercontent.com/msysh/aws-cdk-zsh-completion/main/_cdk

# 自動サジェスト - 履歴に基づいてコマンドを提案（薄い文字で表示）
zinit light zsh-users/zsh-autosuggestions

# zsh-autosuggestions 色設定（モノクロ基調 - 中間グレー）
ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=#7E8A89'

# zsh-autosuggestions 部分適用の設定
# forward-wordを部分適用ウィジェットとして使用
ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-word)
# forward-charも部分適用ウィジェットとして使用
ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-char vi-forward-char)

# FZFタブ補完 - タブ補完をfzfでインタラクティブに
zinit light Aloxaf/fzf-tab

# ----------------------
# プロンプトとナビゲーション
# ----------------------
# Starship - 高速でカスタマイズ可能なプロンプト
eval "$(starship init zsh)"

# Zoxide - スマートなcdコマンド（頻繁に使うディレクトリに素早く移動）
# cdコマンドを置き換え（Claude Codeでは無効化）
if [[ "$CLAUDECODE" != "1" ]]; then
  eval "$(zoxide init --cmd cd zsh)"
fi

# ----------------------
# WezTerm用のOSC 7対応
# ----------------------
# WezTermに現在のディレクトリ情報を送信（タブタイトルに表示）
# OSC 7エスケープシーケンスを使用してターミナルに現在のパスを通知
if [ "$TERM_PROGRAM" = "WezTerm" ] || [ -n "$WEZTERM_PANE" ]; then
  # プロンプトが表示される前に実行される関数
  precmd() {
    # OSC 7 - 現在のディレクトリをターミナルに通知
    printf '\e]7;file://%s%s\e\\' "$HOST" "$PWD"
  }
fi

# ----------------------
# 履歴設定
# ----------------------
# 履歴の詳細オプション（環境変数は.zshenvで設定）
setopt EXTENDED_HISTORY          # 開始と終了時刻を記録
setopt hist_ignore_dups          # 直前と同じコマンドは記録しない
setopt hist_ignore_all_dups      # 履歴内の古い重複コマンドを削除
setopt hist_find_no_dups         # 履歴検索時に重複を表示しない
setopt hist_ignore_space         # スペースで始まるコマンドは記録しない
setopt hist_verify               # 履歴展開後、実行前に編集可能にする
setopt hist_reduce_blanks        # 余分な空白を詰めて記録
setopt hist_save_no_dups         # 履歴ファイル保存時に重複を除去
setopt hist_no_store             # historyコマンド自体は記録しない
setopt hist_expand               # 履歴展開を有効化
setopt share_history             # 複数のzshセッション間で履歴を共有
setopt appendhistory             # 履歴を上書きではなく追加

# ----------------------
# キーバインド
# ----------------------
# Ctrl+P/N で履歴を前方/後方検索（入力済み文字列にマッチ）
bindkey '^p' history-search-backward
bindkey '^n' history-search-forward

# edit-command-line - コマンドラインをエディタで編集
autoload -Uz edit-command-line
zle -N edit-command-line
bindkey '^x^e' edit-command-line  # Ctrl+X Ctrl+E でエディタ起動

# zsh-autosuggestionsの部分適用キーバインド
bindkey '^[f' forward-word          # Alt+F で単語単位の部分適用
bindkey '^[[1;5C' forward-word      # Ctrl+Right で単語単位の部分適用
bindkey '^[[C' forward-char         # 右矢印で1文字ずつ適用

# 補完用のキーバインド（WezTermと競合しない代替キー）
# Ctrl+Y - 現在の単語だけを適用して次の補完位置へ（Alt+Enterの代替）
bindkey '^y' accept-and-menu-complete

# Ctrl+] - 現在の補完を確定して続けて入力可能に（Ctrl+Qの代替）
bindkey '^]' accept-and-hold


# ----------------------
# インタラクティブシェル設定
# ----------------------
# バックグラウンドジョブの優先度をフォアグラウンドと同じにする
unsetopt bg_nice

# 補完候補を詰めて表示（スペース効率化）
setopt list_packed

# ビープ音を無効化
setopt no_beep

# ファイル種別記号（*/=@など）を補完リストに表示しない
unsetopt list_types

# ----------------------
# 補完機能の設定
# ----------------------
# ========================================
# LS_COLORS（補完候補の色設定）
# ========================================
# モノクロ基調＋アクセント
# di=ディレクトリ(cyan), ln=シンボリックリンク(magenta), ex=実行可能(前景+太字)
# *.md/*.txt=中間グレー, *.go/*.ts/*.js=bright cyan
export LS_COLORS='di=38;2;90;175;173:ln=38;2;140;131;163:ex=38;2;215;226;225;1:*.md=38;2;126;138;137:*.txt=38;2;126;138;137:*.go=38;2;150;203;209:*.ts=38;2;150;203;209:*.js=38;2;150;203;209'

# ========================================
# FZF カラー設定（fzf-tab用）
# ========================================
export FZF_DEFAULT_OPTS='
  --color=bg+:#2A2F2E,bg:#0E1210,fg:#D7E2E1,fg+:#F2FFFF
  --color=hl:#5AAFAD,hl+:#96CBD1,info:#7E8A89,marker:#5AAFAD
  --color=prompt:#8C83A3,spinner:#8C83A3,pointer:#5AAFAD,header:#7E8A89
  --color=border:#3A3F3E,gutter:#0E1210
'

# Worklog CLI completion
fpath=(/Users/s23159/.local/share/zsh/site-functions $fpath)

# 補完機能を有効化
autoload -U compinit && compinit

# メニュー選択機能を有効化
zmodload zsh/complist

# メニュー選択モードでのキーバインド
bindkey -M menuselect '^[[C' forward-char                  # 右矢印で次の候補へ
bindkey -M menuselect '^[[D' backward-char                 # 左矢印で前の候補へ
bindkey -M menuselect '^[[B' down-line-or-history          # 下矢印で下の候補へ
bindkey -M menuselect '^[[A' up-line-or-history            # 上矢印で上の候補へ
bindkey -M menuselect '^i' menu-complete                   # Tabで次の候補
bindkey -M menuselect '^[[Z' reverse-menu-complete         # Shift+Tabで前の候補
bindkey -M menuselect ' ' accept-and-infer-next-history    # スペースでディレクトリ部分適用
bindkey -M menuselect '/' accept-and-infer-next-history    # スラッシュでディレクトリ部分適用
bindkey -M menuselect '^m' accept-line                     # Enterで確定
bindkey -M menuselect '^g' send-break                      # Ctrl+Gでキャンセル

# 補完スタイルの詳細設定
# 大文字小文字を区別しない補完
zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'

# 補完候補に色を付ける（LS_COLORS環境変数を使用）
zstyle ':completion:*' list-colors '${(s.:.)LS_COLORS}'

# 補完候補を矢印キーで選択可能にする
zstyle ':completion:*' menu select

# 補完方法の優先順位（通常補完→部分一致→曖昧補完）
zstyle ':completion:*' completer _complete _match _approximate

# 曖昧補完で許容する誤字数（1文字まで）
zstyle ':completion:*:approximate:*' max-errors 1

# fzf-tabプラグイン用の設定（デフォルトメニューを無効化）
zstyle ':completion:*' menu no

# ----------------------
# エイリアス
# ----------------------
# 基本コマンドのエイリアス
alias ls='ls --color=auto'           # lsに色を付ける
alias ll='ls -l'                     # 詳細表示
alias la='ls -la'                    # 隠しファイルも含めて詳細表示
alias l='ls -CF'                     # ファイル種別記号付き表示
alias vi='nvim'
alias cl='claude'
alias vtmp='nvim "${TMPDIR%/}/$(date "+%Y%m%d_%H%M%S").md"'  # 一時ディレクトリにタイムスタンプ付きmdファイルを作成して編集
alias now='date "+%Y/%m/%d %H:%M:%S"'  # 現在日時を表示
alias wl='worklog'                       # 作業ログツールの短縮形

# Kubernetesショートカット
alias k='kubectl'

# Lazygit
alias lg='lazygit'                       # Git操作用のTUIツールを起動

# SSH接続時のターミナルタイプを明示的に指定（カラー対応）
alias ssh='TERM=xterm-256color \ssh'

# Delta（diff表示ツール）のデフォルトオプション
alias delta='delta --dark --paging=never --line-numbers --syntax-theme base16-256 -s'

# Copilot CLIツール
alias awscp='/opt/homebrew/bin/copilot'                           # AWS Copilot (ECS/Fargateデプロイツール)
alias copilot='/opt/homebrew/Caskroom/copilot-cli/0.0.394/copilot'  # GitHub Copilot CLI (AIターミナルアシスタント)

# ----------------------
# Google Cloud SDK
# ----------------------
# GCPコマンドラインツールのPATH設定
if [ -f '/opt/homebrew/share/google-cloud-sdk/path.zsh.inc' ]; then
  . '/opt/homebrew/share/google-cloud-sdk/path.zsh.inc'
fi

# GCPコマンド補完機能
if [ -f '/opt/homebrew/share/google-cloud-sdk/completion.zsh.inc' ]; then
  . '/opt/homebrew/share/google-cloud-sdk/completion.zsh.inc'
fi

# ----------------------
# AWS CLI
# ----------------------
# AWS CLI v2の補完機能を有効化
complete -C '/opt/homebrew/bin/aws_completer' aws

# ----------------------
# Terraform
# ----------------------
autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C /opt/homebrew/bin/terraform terraform

# ----------------------
# ローカル設定
# ----------------------
# マシン固有の設定があれば読み込む
[ -f ~/.zshrc.local ] && source ~/.zshrc.local

