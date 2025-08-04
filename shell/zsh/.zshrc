# Amazon Q pre block. Keep at the top of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.pre.zsh"
# ----------------------
# 基本設定
# ----------------------
# コアダンプを残さない
limit coredumpsize 0

# oh-my-zshの自動タイトル設定を無効化
# これによりターミナルのタイトルバーにユーザー名@ホスト名が表示されなくなる
# DISABLE_AUTO_TITLE="true"


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


# Amazon Q post block. Keep at the bottom of this file.
[[ -f "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh" ]] && builtin source "${HOME}/Library/Application Support/amazon-q/shell/zshrc.post.zsh"

