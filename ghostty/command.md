```bash
▲ ~/.config ghostty +list-keybinds --default

# スクリーンファイルを開く
super + alt   + shift + j       write_screen_file:open
# 全てのウィンドウを閉じる
super + alt   + shift + w       close_all_windows
# インスペクターの表示切り替え
super + alt   + i               inspector:toggle
# タブを閉じる
super + alt   + w               close_tab
# 上の分割ペインに移動
super + alt   + up              goto_split:up
# 下の分割ペインに移動
super + alt   + down            goto_split:down
# 右の分割ペインに移動
super + alt   + right           goto_split:right
# 左の分割ペインに移動
super + alt   + left            goto_split:left
# フルスクリーン切り替え
super + ctrl  + f               toggle_fullscreen
# 分割ペインのサイズを均等化
super + ctrl  + equal           equalize_splits
# 分割ペインを上に10拡張
super + ctrl  + up              resize_split:up,10
# 分割ペインを下に10拡張
super + ctrl  + down            resize_split:down,10
# 分割ペインを右に10拡張
super + ctrl  + right           resize_split:right,10
# 分割ペインを左に10拡張
super + ctrl  + left            resize_split:left,10
# 下に分割
super + shift + d               new_split:down
# スクリーンファイルを貼り付け
super + shift + j               write_screen_file:paste
# 選択範囲から貼り付け
super + shift + v               paste_from_selection
# ウィンドウを閉じる
super + shift + w               close_window
# 設定を再読み込み
super + shift + comma           reload_config
# 前のタブに移動
super + shift + left_bracket    previous_tab
# 次のタブに移動
super + shift + right_bracket   next_tab
# 前のプロンプトに移動
super + shift + up              jump_to_prompt:-1
# 次のプロンプトに移動
super + shift + down            jump_to_prompt:1
# 分割ペインのズーム切り替え
super + shift + enter           toggle_split_zoom
# 前のタブに移動
ctrl  + shift + tab             previous_tab
# 全選択
super + a                       select_all
# クリップボードにコピー
super + c                       copy_to_clipboard
# 右に分割
super + d                       new_split:right
# 画面をクリア
super + k                       clear_screen
# 新しいウィンドウを開く
super + n                       new_window
# 終了
super + q                       quit
# 新しいタブを開く
super + t                       new_tab
# クリップボードから貼り付け
super + v                       paste_from_clipboard
# サーフェスを閉じる
super + w                       close_surface
# フォントサイズをリセット
super + zero                    reset_font_size
# タブ1に移動
super + physical:one            goto_tab:1
# タブ2に移動
super + physical:two            goto_tab:2
# タブ3に移動
super + physical:three          goto_tab:3
# タブ4に移動
super + physical:four           goto_tab:4
# タブ5に移動
super + physical:five           goto_tab:5
# タブ6に移動
super + physical:six            goto_tab:6
# タブ7に移動
super + physical:seven          goto_tab:7
# タブ8に移動
super + physical:eight          goto_tab:8
# 最後のタブに移動
super + physical:nine           last_tab
# 設定を開く
super + comma                   open_config
# フォントサイズを縮小
super + minus                   decrease_font_size:1
# フォントサイズを拡大
super + plus                    increase_font_size:1
# フォントサイズを拡大
super + equal                   increase_font_size:1
# 前の分割ペインに移動
super + left_bracket            goto_split:previous
# 次の分割ペインに移動
super + right_bracket           goto_split:next
# 前のプロンプトに移動
super + up                      jump_to_prompt:-1
# 次のプロンプトに移動
super + down                    jump_to_prompt:1
# 行末に移動 (Ctrl+E)
super + right                   text:\x05
# 行頭に移動 (Ctrl+A)
super + left                    text:\x01
# 最上部にスクロール
super + home                    scroll_to_top
# 最下部にスクロール
super + end                     scroll_to_bottom
# 1ページ上にスクロール
super + page_up                 scroll_page_up
# 1ページ下にスクロール
super + page_down               scroll_page_down
# フルスクリーン切り替え
super + enter                   toggle_fullscreen
# 行をクリア (Ctrl+U)
super + backspace               text:\x15
# 次の単語に移動 (Esc+F)
alt   + right                   esc:f
# 前の単語に移動 (Esc+B)
alt   + left                    esc:b
# 次のタブに移動
ctrl  + tab                     next_tab
# 選択範囲を上に拡張
shift + up                      adjust_selection:up
# 選択範囲を下に拡張
shift + down                    adjust_selection:down
# 選択範囲を右に拡張
shift + right                   adjust_selection:right
# 選択範囲を左に拡張
shift + left                    adjust_selection:left
# 選択範囲を行頭まで拡張
shift + home                    adjust_selection:home
# 選択範囲を行末まで拡張
shift + end                     adjust_selection:end
# 選択範囲を1ページ上まで拡張
shift + page_up                 adjust_selection:page_up
# 選択範囲を1ページ下まで拡張
shift + page_down               adjust_selection:page_down
```
