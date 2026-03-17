# path_helper (/etc/zprofile) 後に Nix を再度先頭に置く
if [ -e "$HOME/.nix-profile/bin" ]; then
  export PATH="$HOME/.nix-profile/bin:$PATH"
fi
