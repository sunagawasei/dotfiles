# Homebrew (最初に評価 — 他のツールが依存)
[ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"

# Nix Home Manager（brew shellenv / path_helper より後に評価して優先させる）
if [ -e "$HOME/.nix-profile/bin" ]; then
  export PATH="$HOME/.nix-profile/bin:$PATH"
fi

# Docker認証ヘルパー
export PATH=$PATH:/Users/s23159/.config/cycloud/bin

# Go バイナリパス (ガード付き)
if command -v go &>/dev/null; then
  export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Rust
[ -f "$HOME/.cargo/env" ] && . "$HOME/.cargo/env"

# HM session variables（非ログインシェル対応）
if [ -e "$HOME/.nix-profile/etc/profile.d/hm-session-vars.sh" ]; then
  . "$HOME/.nix-profile/etc/profile.d/hm-session-vars.sh"
fi

