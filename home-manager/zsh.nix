{ config, lib, pkgs, ... }:
{
  programs.zsh = {
    enable = true;
    dotDir = "${config.xdg.configHome}/zsh";

    defaultKeymap = "emacs";

    # ---- 履歴 ----
    history = {
      size = 50000;
      save = 50000;
      path = "${config.home.homeDirectory}/.config/zsh/.zsh_history";
      extended = true;       # EXTENDED_HISTORY
      ignoreAllDups = true;  # hist_ignore_all_dups
      findNoDups = true;     # hist_find_no_dups
      ignoreSpace = true;    # hist_ignore_space
      saveNoDups = true;     # hist_save_no_dups
      share = true;          # share_history
    };

    # HM に対応 option がない setopt
    setOptions = [
      "HIST_VERIFY"
      "HIST_REDUCE_BLANKS"
      "HIST_NO_STORE"
      "HIST_EXPAND"
      "NO_BG_NICE"      # unsetopt bg_nice
      "LIST_PACKED"
      "NO_BEEP"
      "NO_LIST_TYPES"   # unsetopt list_types
    ];

    # ---- エイリアス ----
    shellAliases = {
      ls = "ls --color=auto";
      vtmp = ''nvim "''${TMPDIR%/}/$(date "+%Y%m%d_%H%M%S").md"'';
      ssh = ''TERM=xterm-256color \ssh'';
      delta = "delta --dark --paging=never --line-numbers --syntax-theme base16-256 -s";
      nswitch = "sudo darwin-rebuild switch --flake ~/.config";
    };

    # Zinit の zsh-completions が compinit より前にロードされるため
    # HM の自動 compinit を抑制し、initContent 内で手動制御
    completionInit = "";

    # ---- .zshenv (envExtra) ----
    # HM が自動で ZDOTDIR export と hm-session-vars.sh の source を追加するため
    # それらはここに含めない
    envExtra = ''
      # Homebrew (最初に評価 — 他のツールが依存)
      [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"

      # Homebrew GitHub API token (cycloud-io/tap に必要)
      # gh auth token --hostname でローカルキーリングから直接読む（GHEへのネットワークアクセスなし）
      if command -v gh &>/dev/null; then
        HOMEBREW_GITHUB_API_TOKEN=$(gh auth token --hostname github.com 2>/dev/null) && export HOMEBREW_GITHUB_API_TOKEN
      fi

      # Nix Home Manager（brew shellenv / path_helper より後に評価して優先させる）
      if [ -e "$HOME/.nix-profile/bin" ]; then
        export PATH="$HOME/.nix-profile/bin:$PATH"
      fi

      # Docker認証ヘルパー
      export PATH=$PATH:${config.home.homeDirectory}/.config/cycloud/bin

      # Go バイナリパス (ガード付き)
      if command -v go &>/dev/null; then
        export PATH="$PATH:$(go env GOPATH)/bin"
      fi

      # Rust
      [ -f "$HOME/.cargo/env" ] && . "$HOME/.cargo/env"
    '';

    # ---- .zprofile (profileExtra) ----
    profileExtra = ''
      # path_helper (/etc/zprofile) 後に Nix を再度先頭に置く
      if [ -e "$HOME/.nix-profile/bin" ]; then
        export PATH="$HOME/.nix-profile/bin:$PATH"
      fi
    '';

    # ---- .zshrc (initContent) ----
    # initExtra/initExtraFirst は deprecated → initContent + lib.mkOrder を使用
    initContent = lib.mkMerge [
      (lib.mkOrder 500 ''
        # コアダンプファイルを作成しない
        limit coredumpsize 0
        # flow control無効化（Ctrl+S/Ctrl+Qによる端末停止を防止、zeno ^x^sに必要）
        stty -ixon
      '')
      (lib.mkOrder 1000 ''
        # ---- プラグインマネージャー (Zinit) ----
        ZINIT_HOME="''${XDG_DATA_HOME:-''${HOME}/.local/share}/zinit/zinit.git"
        if [ ! -d "$ZINIT_HOME" ]; then
          mkdir -p "$(dirname $ZINIT_HOME)"
          git clone https://github.com/zdharma-continuum/zinit.git "$ZINIT_HOME"
        fi
        source "$ZINIT_HOME/zinit.zsh"

        # ---- 補完初期化ヘルパー（turbo wait"0b" の atinit から呼ばれる）----
        __setup_completions() {
          zicompinit
          zicdreplay
          # zoxide（compdef が compinit 後に正常動作）
          if [[ "$CLAUDECODE" != "1" ]]; then
            eval "$(zoxide init --cmd cd zsh)"
          fi
          # bashcompinit + 外部ツール補完
          autoload -U +X bashcompinit && bashcompinit
          complete -C "$(command -v aws_completer)" aws
          complete -o nospace -C "$(command -v terraform)" terraform
          # Google Cloud SDK 補完
          if command -v gcloud &>/dev/null; then
            _gcloud_bin="$(command -v gcloud)"
            _gcloud_root="$(dirname "$(readlink -f "$_gcloud_bin")")/.."
            for f in completion.zsh.inc; do
              [[ -f "$_gcloud_root/google-cloud-sdk/$f" ]] && source "$_gcloud_root/google-cloud-sdk/$f" && continue
              [[ -f "$_gcloud_root/$f" ]] && source "$_gcloud_root/$f"
            done
            unset _gcloud_bin _gcloud_root
          fi
        }

        # ---- Pure プロンプト ----
        # 環境変数（読み込み前に設定が必要）
        PURE_CMD_MAX_EXEC_TIME=5
        PURE_GIT_PULL=0
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
        [[ -f "''${HOME}/.nix-profile/share/fzf/key-bindings.zsh" ]] && source "''${HOME}/.nix-profile/share/fzf/key-bindings.zsh"

        # Worklog CLI completion
        fpath=(${config.home.homeDirectory}/.local/share/zsh/site-functions $fpath)

        # fzf-tab・補完スタイル設定（compinit前に宣言しても有効）
        zstyle ':completion:*' menu no
        zstyle ':fzf-tab:complete:cd:*' disabled-on any  # zoxideとの競合を回避
        zstyle ':completion:*' matcher-list 'm:{a-z}={A-Z}'
        zstyle ':completion:*' list-colors "''${(s.:.)LS_COLORS}"
        zstyle ':completion:*' completer _complete _match

        # ---- Google Cloud SDK (PATH設定のみ同期) ----
        if command -v gcloud &>/dev/null; then
          _gcloud_bin="$(command -v gcloud)"
          _gcloud_root="$(dirname "$(readlink -f "$_gcloud_bin")")/.."
          for f in path.zsh.inc; do
            [[ -f "$_gcloud_root/google-cloud-sdk/$f" ]] && source "$_gcloud_root/google-cloud-sdk/$f" && continue
            [[ -f "$_gcloud_root/$f" ]] && source "$_gcloud_root/$f"
          done
          unset _gcloud_bin _gcloud_root
        fi

        # ---- Turbo mode zinit ----
        # wait"0a": fpath追加（compinit前）
        zinit ice wait"0a" lucid blockf atpull'zinit creinstall -q .'
        zinit light zsh-users/zsh-completions

        zinit ice wait"0a" lucid as"completion"
        zinit snippet https://raw.githubusercontent.com/msysh/aws-cdk-zsh-completion/main/_cdk

        # wait"0b": compinit実行 → fzf-tab
        zinit ice wait"0b" lucid atinit'__setup_completions'
        zinit light Aloxaf/fzf-tab

        # wait"0c": autosuggestions, zeno（fzf-tab の後）
        zinit ice wait"0c" lucid atload'
          _zsh_autosuggest_start
          ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE="fg=#7A869A"
          ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-word)
          ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-char vi-forward-char)
        '
        zinit light zsh-users/zsh-autosuggestions

        zinit ice wait"0c" lucid atload'
          __zeno_space_wrapper() {
            local orig_buffer="$BUFFER"
            zle zeno-auto-snippet
            if [[ "$BUFFER" == "$orig_buffer" ]]; then
              LBUFFER+=" "
            fi
          }
          zle -N __zeno_space_wrapper
          bindkey " " __zeno_space_wrapper
          bindkey "^m" zeno-auto-snippet-and-accept-line
          bindkey "^r" zeno-smart-history-selection
          bindkey "^x^s" zeno-insert-snippet
        '
        zinit light yuki-yano/zeno.zsh

        # wait"0d": syntax-highlighting（最後）
        zinit ice wait"0d" lucid atload'
          typeset -A ZSH_HIGHLIGHT_STYLES
          ZSH_HIGHLIGHT_STYLES[default]="fg=#CEF5F2"
          ZSH_HIGHLIGHT_STYLES[arg0]="fg=#CEF5F2"
          ZSH_HIGHLIGHT_STYLES[command]="fg=#6CD8D3,bold"
          ZSH_HIGHLIGHT_STYLES[builtin]="fg=#6CD8D3,bold"
          ZSH_HIGHLIGHT_STYLES[alias]="fg=#6CD8D3,bold"
          ZSH_HIGHLIGHT_STYLES[function]="fg=#6CD8D3,bold"
          ZSH_HIGHLIGHT_STYLES[unknown-token]="fg=#936997,bold"
          ZSH_HIGHLIGHT_STYLES[path]="fg=#9DDCD9,underline"
          ZSH_HIGHLIGHT_STYLES[single-quoted-argument]="fg=#659D9E"
          ZSH_HIGHLIGHT_STYLES[double-quoted-argument]="fg=#659D9E"
          ZSH_HIGHLIGHT_STYLES[single-hyphen-option]="fg=#A4ABCB"
          ZSH_HIGHLIGHT_STYLES[double-hyphen-option]="fg=#A4ABCB"
          ZSH_HIGHLIGHT_STYLES[comment]="fg=#7A869A"
        '
        zinit light zsh-users/zsh-syntax-highlighting

        # ---- ローカル設定 ----
        if [ -f ~/.zshrc.local ]; then source ~/.zshrc.local; fi
      '')
    ];
  };
}
