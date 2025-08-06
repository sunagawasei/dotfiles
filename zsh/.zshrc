# Amazon Q pre block. Keep at the top of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh"
# ----------------------
# 基本設定
# ----------------------
# コアダンプを残さない
limit coredumpsize 0


# Starship (プロンプトのカスタマイズ)
eval "$(starship init zsh)"

# ----------------------
# 履歴設定
# ----------------------
# 履歴ファイルの保存先
export HISTFILE=${HOME}/.config/zsh/.zsh_history
# メモリに保存される履歴の件数
export HISTSIZE=1000
# 履歴ファイルに保存される履歴の件数
export SAVEHIST=100000
export HISTFILESIZE=100000
# 重複を記録しない
setopt hist_ignore_dups
# 開始と終了を記録
setopt EXTENDED_HISTORY
# ヒストリに追加されるコマンド行が古いものと同じなら古いものを削除
setopt hist_ignore_all_dups
# スペースで始まるコマンド行はヒストリリストから削除
setopt hist_ignore_space
# ヒストリを呼び出してから実行する間に一旦編集可能
setopt hist_verify
# 余分な空白は詰めて記録
setopt hist_reduce_blanks
# 古いコマンドと同じものは無視
setopt hist_save_no_dups
# historyコマンドは履歴に登録しない
setopt hist_no_store
# 保管時にヒストリを自動的に展開
setopt hist_expand
# history共有
setopt share_history

# ----------------------
# インタラクティブシェル設定
# ----------------------
# zshの補完候補が画面から溢れ出る時、それでも表示するか確認
export LISTMAX=50
# バックグラウンドジョブの優先度(ionice)をbashと同じ挙動に
unsetopt bg_nice
# 補完候補を詰めて表示
setopt list_packed
# ピープオンを鳴らさない
setopt no_beep
# ファイル種別起動を補完候補の末尾に表示しない
unsetopt list_types

# ----------------------
# 補完機能の強化
# ----------------------
# 補完機能を有効化
autoload -U compinit && compinit

# 補完のスタイル設定
zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'  # 大文字小文字を区別しない
zstyle ':completion:*' list-colors ''                 # 補完候補に色を付ける
zstyle ':completion:*' menu select                    # 補完候補をメニュー形式で表示
zstyle ':completion:*' completer _complete _match _approximate  # 補完方法の設定
zstyle ':completion:*:approximate:*' max-errors 1    # 曖昧補完で許容するエラー数

# ----------------------
# エイリアスと関数
# ----------------------
# 基本エイリアス
alias ll='ls -l'
alias k='kubectl'
alias ssh='TERM=xterm-256color \ssh'

# ----------------------
# クラウド SDK 補完機能
# ----------------------
# Google Cloud SDK - PATH
# GCPコマンドラインツールのPATHを設定
if [ -f '/opt/homebrew/share/google-cloud-sdk/path.zsh.inc' ]; then
  . '/opt/homebrew/share/google-cloud-sdk/path.zsh.inc'
fi

# Google Cloud SDK - 補完
# GCPコマンド補完機能の有効化
if [ -f '/opt/homebrew/share/google-cloud-sdk/completion.zsh.inc' ]; then
  . '/opt/homebrew/share/google-cloud-sdk/completion.zsh.inc'
fi

# AWSコマンド補完機能の有効化
complete -C '/opt/homebrew/opt/awscli@1/bin/aws_completer' aws


# Amazon Q post block. Keep at the bottom of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh"

