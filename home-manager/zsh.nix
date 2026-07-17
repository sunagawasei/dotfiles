{ config, lib, pkgs, ... }:
let
  colors = import ./colors.nix;
in
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
      nswitch = "sudo darwin-rebuild switch --flake ~/.config#CA-20021145";
      nupdate = "nix flake update --flake ~/.config && sudo darwin-rebuild switch --flake ~/.config#CA-20021145";
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

      # mysql-client (keg-only)
      export PATH="/opt/homebrew/opt/mysql-client/bin:$PATH"

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

        # zstyleカラー設定（Ghost Visor）
        # MAIN: foregrounds.main
        zstyle ':prompt:pure:path' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:git:arrow' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:git:action' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:prompt:success' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:git:branch' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:git:stash' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:suspended_jobs' color '${colors.foregrounds.main}'

        # TERTIARY: blues_slates.comment_gray
        zstyle ':prompt:pure:git:branch:cached' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:git:dirty' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:execution_time' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:prompt:continuation' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:virtualenv' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:host' color '${colors.blues_slates.comment_gray}'
        zstyle ':prompt:pure:user' color '${colors.blues_slates.comment_gray}'

        # ERROR: ❯ は常に統一色。rootログイン時のみ警告色
        zstyle ':prompt:pure:prompt:error' color '${colors.foregrounds.main}'
        zstyle ':prompt:pure:user:root' color '${colors.ansi.red}'

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
          # Claude Code 終了後にpaneタイトル(OSC 2)へ残る状態グリフを消去する。
          # プロンプト復帰時のみ発火する(claude動作中はzshがブロックされ発火しない)ため、
          # claude終了直後にタブがcwd表示へ戻り、状態インジケータの残留を防ぐ。
          __wezterm_reset_title() { printf '\e]2;\e\\'; }
          autoload -Uz add-zsh-hook
          add-zsh-hook precmd __wezterm_osc7
          add-zsh-hook preexec __wezterm_preexec
          add-zsh-hook precmd  __wezterm_precmd_busy
          add-zsh-hook precmd  __wezterm_reset_title
        fi

        # ---- Kube current context (Pure preprompt セグメント / psvar[21]) ----
        # kubectl を呼ばず KUBECONFIG が指す kubeconfig の current-context を直接パースし、
        # Pure 1行目の git 情報塊の直後（exec_time の前）へ表示する。
        # 表示は KUBECONFIG がセットされている時のみ（direnv 等でプロジェクト単位に出し分ける運用）。
        # 常時パスは zstat ビルトインのみ（fork なし）。config 変化時だけ awk で再パース。
        zmodload zsh/stat 2>/dev/null
        typeset -g __kube_ctx_sig="" __kube_ctx=""

        # Pure の PROMPT へ psvar[21] セグメントを1度だけ差し込む（マーカー %(19V の直前）。
        # 冪等＆Pure 再 setup 対策として precmd 先頭でも再確認する。glob を避けた split-rejoin 方式。
        __kube_prompt_patch() {
          [[ $PROMPT == *'%21v'* ]] && return     # 既にパッチ済み
          [[ $PROMPT == *'%(19V'* ]] || return    # マーカー無し（Pure 仕様変更等）は no-op
          local seg='%(21V. %F{${colors.teals.mid_bright}}⎈ %21v%f.)'
          local pre="''${PROMPT%%'%(19V'*}"
          local suf="''${PROMPT#*'%(19V'}"
          PROMPT="''${pre}''${seg}%(19V''${suf}"
        }

        # current-context を mtime+size signature キャッシュ付きでパース（変化時のみ再パース）。
        __kube_ctx_update() {
          # 表示は KUBECONFIG セット時のみなので source は KUBECONFIG（colon 区切りで複数可）。
          local -a files=( "''${(@s.:.)KUBECONFIG}" )

          local sig="" f
          typeset -A st
          for f in $files; do
            [[ -n "$f" && -r "$f" ]] || continue
            zstat -H st "$f" 2>/dev/null && sig+="$f:$st[mtime]:$st[size];"
          done

          [[ "$sig" == "$__kube_ctx_sig" ]] && return
          __kube_ctx_sig="$sig"

          local ctx=""
          for f in $files; do
            [[ -n "$f" && -r "$f" ]] || continue
            ctx=$(awk -F'current-context:[[:space:]]*' '/^current-context:/{print $2; exit}' "$f")
            [[ -n "$ctx" ]] && break
          done
          __kube_ctx="$ctx"
        }

        __kube_precmd() {
          __kube_prompt_patch
          # 判定基準は direnv 等による KUBECONFIG のセットのみ。未セットなら非表示。
          if [[ -z "$KUBECONFIG" ]]; then
            psvar[21]=
            return
          fi
          __kube_ctx_update
          # @ より前で短縮。psvar は %v 描画時に % を再解釈しないためエスケープ不要。
          [[ -n "$__kube_ctx" ]] && psvar[21]="''${__kube_ctx%%@*}" || psvar[21]=
        }
        autoload -Uz add-zsh-hook
        __kube_prompt_patch
        add-zsh-hook precmd __kube_precmd

        # ---- キーバインド ----
        bindkey '^p' history-search-backward
        bindkey '^n' history-search-forward

        autoload -Uz edit-command-line
        zle -N edit-command-line
        bindkey '^x^e' edit-command-line

        # zsh-autosuggestions 部分適用
        bindkey '^[f' forward-word
        bindkey '^[[C' forward-char

        # 1文字ずつ受諾（Ctrl+F）— デフォルトの forward-char(全受諾) より細かい粒度
        partial-accept-char() { zle .forward-char }
        zle -N partial-accept-char
        bindkey '^f' partial-accept-char

        # 非英数字（/・空白・.・-等）を区切りに受諾（Ctrl+O）
        partial-accept-subword() {
          local WORDCHARS=""
          zle .forward-word
        }
        zle -N partial-accept-subword
        bindkey '^o' partial-accept-subword

        # 補完用
        bindkey '^y' accept-and-menu-complete
        bindkey '^]' accept-and-hold

        # ---- 補完機能 ----
        # LS_COLORS（補完候補の色設定）
        export LS_COLORS='ma=48;2;100;187;190;38;2;19;32;24:di=38;2;157;220;217:ln=38;2;147;105;151:ex=38;2;108;216;211;1:*.md=38;2;82;91;101:*.txt=38;2;82;91;101:*.go=38;2;164;171;203:*.ts=38;2;164;171;203:*.js=38;2;164;171;203'

        # FZF カラー設定
        export FZF_DEFAULT_OPTS='
          --color=bg+:${colors.core.selection_bg},bg:-1,fg:${colors.foregrounds.main},fg+:${colors.core.selection_fg}
          --color=hl:${colors.teals.bright},hl+:${colors.core.selection_fg},info:${colors.blues_slates.comment_gray},marker:${colors.core.selection_fg}
          --color=prompt:${colors.zsh.error},spinner:${colors.zsh.error},pointer:${colors.core.selection_fg},header:${colors.blues_slates.comment_gray}
          --color=border:${colors.teals.border},gutter:-1
        '

        # fzf キーバインド: Ctrl+T (ファイル選択), Alt+C (ディレクトリ移動), Ctrl+R (履歴検索)
        # Ctrl+R は fzf-history-widget が担当(HISTFILE=矢印キーと同一ソース)。zeno は ^R を奪わない
        [[ -f "${pkgs.fzf}/share/fzf/key-bindings.zsh" ]] && source "${pkgs.fzf}/share/fzf/key-bindings.zsh"

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

        # wait"0c": zeno（fzf-tab の後）
        zinit ice wait"0c" lucid \
          atload'
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
          # Ctrl+R は zeno の smart-history(別SQLite DB・書き込み停止済み)ではなく
          # fzf-history-widget(HISTFILE=矢印キーと同一ソース)を使う
          bindkey "^x^s" zeno-insert-snippet
        '
        zinit light yuki-yano/zeno.zsh

        # wait"0d": syntax-highlighting（autosuggestions より先に読み込む）
        zinit ice wait"0d" lucid atload'
          typeset -A ZSH_HIGHLIGHT_STYLES
          ZSH_HIGHLIGHT_STYLES[default]="fg=${colors.zsh.default}"
          ZSH_HIGHLIGHT_STYLES[arg0]="fg=${colors.zsh.default}"
          ZSH_HIGHLIGHT_STYLES[command]="fg=${colors.zsh.command},bold"
          ZSH_HIGHLIGHT_STYLES[builtin]="fg=${colors.zsh.command},bold"
          ZSH_HIGHLIGHT_STYLES[alias]="fg=${colors.zsh.command},bold"
          ZSH_HIGHLIGHT_STYLES[function]="fg=${colors.zsh.command},bold"
          ZSH_HIGHLIGHT_STYLES[unknown-token]="fg=${colors.zsh.error},bold"
          ZSH_HIGHLIGHT_STYLES[path]="fg=${colors.zsh.path},underline"
          ZSH_HIGHLIGHT_STYLES[single-quoted-argument]="fg=${colors.zsh.string}"
          ZSH_HIGHLIGHT_STYLES[double-quoted-argument]="fg=${colors.zsh.string}"
          ZSH_HIGHLIGHT_STYLES[single-hyphen-option]="fg=${colors.zsh.option}"
          ZSH_HIGHLIGHT_STYLES[double-hyphen-option]="fg=${colors.zsh.option}"
          ZSH_HIGHLIGHT_STYLES[comment]="fg=${colors.zsh.comment}"
        '
        zinit light zsh-users/zsh-syntax-highlighting

        # autosuggestions: 同期ロード（turbo だと初回プロンプトで widget がバインドされない）
        zinit ice lucid atload'
          # hexは小文字必須: zsh 5.9はregion_highlightを読み戻し時に小文字へ正規化し、
          # autosuggestionsのhighlight_resetは完全一致でエントリを除去するため、
          # 大文字だと右矢印受け入れ後もグレーが残る (zsh-autosuggestions#698と同根)
          ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE="fg=${lib.toLower colors.zsh.comment}"
          ZSH_AUTOSUGGEST_PARTIAL_ACCEPT_WIDGETS+=(forward-word partial-accept-char partial-accept-subword)
        '
        zinit light zsh-users/zsh-autosuggestions

        # ---- ローカル設定 ----
        if [ -f ~/.zshrc.local ]; then source ~/.zshrc.local; fi
      '')
    ];
  };
}
