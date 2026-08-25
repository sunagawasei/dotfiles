return {
  "willothy/flatten.nvim",
  lazy = false,
  priority = 1001,
  commit = "d92ca41e9c330f45c1b854a80c89c8488a9d730c",
  opts = {
    -- should_nestが唯一の実効ガード。この設定はcmd付き起動(+qa等)を守らない。
    nest_if_no_args = true,
    window = {
      open = "smart",
    },
    -- 既知の制限：nvimをEDITORに指定してブロッキング待機するツール(crontab -e等)は、
    -- この環境ではEDITOR=vim(home-manager/shell.nix)のため現状は該当しない。将来nvimを
    -- EDITORに指定するツールを追加する場合はblock_for/should_block設定の追加検討が必要。
    hooks = {
      -- vim.v.argvを見て、プレーンなファイルパス(+ "+<数字>")以外は
      -- 全てローカルnestに倒す(委譲しない)。true=nest(ローカル), false=delegate。
      -- 対象: フラグ全般(-R/-M/-o/-O/-p等)・裸の"-"(stdin)・"--"・
      -- 数字以外の"+cmd"(例: "+qa")・引数なし。ただし、nixラッパーが常に注入する
      -- "--cmd <lua>"ペアは無視する。ユーザー自身が--cmdを明示指定した場合も同様に
      -- 無視され、そのcmdはローカル実行されず委譲後host側で実行される。稀な起動法であり
      -- 許容する既知の制限。
      -- 対話的TUIではNeovim 0.12がcore argvへ注入する"--embed"も無視する。
      -- "--embed"スキップの既知の制限: 外部ツールがmsgpack-rpcサーバとして起動する
      -- "nvim --embed <file>"(GUIクライアント等)が$NVIMを継承している場合も委譲対象に
      -- 入る。Neovim 0.12のTUI→core子とは判別不能なため許容する。
      should_nest = function(_)
        local argv = vim.v.argv
        local has_file = false
        local skip_next = false
        for i = 2, #argv do
          local a = argv[i]
          if skip_next then
            skip_next = false
          elseif a == "--cmd" then
            skip_next = true
          elseif a == "--embed" then
            -- skip: 対話的TUI起動時、Neovim 0.12はTUI(親)+embed core(子)の2プロセス
            -- 構成になり、coreプロセスのargvには常に--embedが付く(ユーザー意図の
            -- フラグではないため無視する)。--headlessは意図的にスキップ対象外のまま
            -- (tool-spawnedなheadless子の委譲を防ぐため)。
          elseif a == "--" then
            return true
          elseif a:sub(1, 1) == "-" then
            return true
          elseif a:sub(1, 1) == "+" and not a:match("^%+%d+$") then
            return true
          elseif a:sub(1, 1) ~= "+" then
            has_file = true
          end
        end
        return not has_file
      end,
      -- host側で、ファイルウィンドウ選択の前に呼ばれる(edit_files内)。
      -- 通常はこの時点のhostのカレントウィンドウが、ユーザーが入力したterminal
      -- ウィンドウと一致する(RPC受領直後・window.open処理より前のため)が、
      -- guest起動〜RPC到達までの間にhost側のフォーカスが変わっていれば一致しない
      -- (既知の残存リスク。無関係なterminalが閉じる可能性があるが<C-/>で回復可能)。
      pre_open = function(_)
        local win = vim.api.nvim_get_current_win()
        local buf = vim.api.nvim_win_get_buf(win)
        _G.__flatten_pending_term_win = (vim.bo[buf].buftype == "terminal") and win or nil
      end,
      -- ファイルを開いた後、記録したterminalウィンドウをtoggleterm経由でclose
      -- (lazygit_hideと同じUX。<C-/>で復元可能)。対象は非hiddenのtoggleterm
      -- terminalのみ(hidden=trueのlazygit/hunkはflatten経由で委譲されないため対象外、
      -- 素の:terminalはhideされず委譲のみ行われる)。閉じた後、flattenが選んだ
      -- ファイルウィンドウ(ctx.winnr)へ明示的にフォーカスを戻す
      -- (toggletermのclose()がorigin_windowへフォーカスを戻してしまうのを打ち消す)。
      post_open = function(ctx)
        local win = _G.__flatten_pending_term_win
        _G.__flatten_pending_term_win = nil
        if win and win ~= ctx.winnr and vim.api.nvim_win_is_valid(win) then
          local terms = require("toggleterm.terminal").get_all()
          for _, term in ipairs(terms) do
            if term.window == win then
              term:close()
              break
            end
          end
        end
        if ctx.winnr and vim.api.nvim_win_is_valid(ctx.winnr) then
          vim.api.nvim_set_current_win(ctx.winnr)
        end
      end,
    },
  },
}
