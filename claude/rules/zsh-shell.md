# Bashツールのシェルはzsh

Bashツールのコマンドはzshで動く。bashの癖で書くと、エラーにならず別の意味で実行される。

- unquoted `$VAR` は単語分割されない。`A="--profile x"; aws $A ...` は `--profile x` を1引数で渡して失敗する。オプションやパスの組は配列 (`A=(--profile x); aws "${A[@]}"`) にするか、コマンドに直接書く
- コマンドを変数に入れて `$N args` で呼ばない。関数で包む

他のzshの罠(`"$VAR:path"` の修飾子解釈・`$status` 予約変数など)は `~/.config` プロジェクトのメモリ `reference_zsh_colon_modifier_pitfall.md` にある。
