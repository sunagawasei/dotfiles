# Amazon Q pre block. Keep at the top of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh"
# ----------------------
# 基本設定
# ----------------------
# コアダンプを残さない
limit coredumpsize 0


# ----------------------
# パッケージ管理
# ----------------------
# Homebrew
eval "$(/opt/homebrew/bin/brew shellenv)"

# Antigen (Zshプラグイン管理)
source $HOME/.local/bin/antigen.zsh

# Load the oh-my-zsh's library
antigen use oh-my-zsh

# プラグインの読み込み
antigen bundles <<EOBUNDLES
    # シンタックスハイライト
    zsh-users/zsh-syntax-highlighting
    # ディレクトリ履歴移動
    rupa/z z.sh
EOBUNDLES


# Antigenの設定を適用
antigen apply

# Starship (プロンプトのカスタマイズ)
eval "$(starship init zsh)"

# ----------------------
# エイリアスと関数
# ----------------------
# 基本エイリアス
alias ll='ls -l'
alias k='kubectl'
alias ssh='TERM=xterm-256color \ssh'

# ----------------------
# 環境変数と PATH
# ----------------------
# AWSクライアント
export PATH="/opt/homebrew/opt/awscli@1/bin:$PATH"

# Docker認証ヘルパー
export PATH=$PATH:/Users/s23159/.config/cycloud/bin

# Go バイナリパス
export PATH="$PATH:$(go env GOPATH)/bin"

# ----------------------
# クラウド SDK
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


# ----------------------
# キーバインド設定
# ----------------------
# viモードを解除（デフォルトのemacsモードを使用）

# Amazon Q post block. Keep at the bottom of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh"
