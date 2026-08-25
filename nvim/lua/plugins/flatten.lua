return {
  "willothy/flatten.nvim",
  lazy = false,
  priority = 1001,
  commit = "d92ca41e9c330f45c1b854a80c89c8488a9d730c",
  opts = {
    -- should_nestが唯一の実効ガード。この設定はcmd付き起動(+qa等)を守らない。
    nest_if_no_args = true,
    window = {
      open = function(opts)
        local focus = opts.files[1]
        if not focus then return nil, nil end
        local win = require("flatten.core").smart_open()
        if win then
          vim.api.nvim_win_set_buf(win, focus.bufnr)
          vim.api.nvim_set_current_win(win)
          _G.__flatten_had_candidate = true
          return focus.bufnr, win
        end
        _G.__flatten_had_candidate = false
        return focus.bufnr, nil
      end,
    },
    -- 既知の制限：nvimをEDITORに指定してブロッキング待機するツール(crontab -e等)は、
    -- この環境ではEDITOR=vim(home-manager/shell.nix)のため現状は該当しない。将来nvimを
    -- EDITORに指定するツールを追加する場合はblock_for/should_block設定の追加検討が必要。
    hooks = {
      -- vim.v.argvを見て、プレーンなファイルパス以外は全てローカルnestに倒す
      -- (委譲しない)。true=nest(ローカル), false=delegate。
      -- 対象: フラグ全般(-R/-M/-o/-O/-p等)・裸の"-"(stdin)・"--"・
      -- 全"+"開始token(+<数字>を含む)・引数なし。"--cmd"はNixラッパーが
      -- 注入する固定payloadとの厳密一致時だけ無視し、それ以外は即nestする。
      -- 対話的TUIではNeovim 0.12がcore argvへ注入する"--embed"を無条件で無視する。
      -- ユーザー指定の"nvim --embed <file>"も委譲される既知の制限を許容する。
      should_nest = function(_)
        return require("util.flatten_classify").classify(vim.v.argv)
      end,
      -- host側で、ファイルウィンドウ選択の前に呼ばれる(edit_files内)。
      -- 通常はこの時点のhostのカレントウィンドウが、ユーザーが入力したterminal
      -- ウィンドウと一致する(RPC受領直後・window.open処理より前のため)が、
      -- guest起動〜RPC到達までの間にhost側のフォーカスが変わっていれば一致しない
      -- (既知の残存リスク。無関係なterminalが閉じる可能性があるが<C-/>で回復可能)。
      pre_open = function(_)
        -- 前サイクルで例外によりpost_openへ到達しなかった場合のstale値を、次サイクル
        -- 開始時に必ずクリアする。
        _G.__flatten_had_candidate = nil
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
        local had_candidate = _G.__flatten_had_candidate
        _G.__flatten_had_candidate = nil
        if not had_candidate then
          _G.__flatten_pending_term_win = nil
          return
        end
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
