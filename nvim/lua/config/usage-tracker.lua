-- キーバインド・プラグイン使用状況トラッカー（純観測方式・データ収集専任）
-- vim.keymap.set のラップは行わず、vim.on_key + nvim_get_keymap の突き合わせだけで観測する。
-- 1ヶ月分のデータを stdpath("data") の usage-tracker.json に貯める。棚卸しはこのJSONを
-- 直接読んで分析する運用のため、レポート生成コマンドはここには置かない。

local M = {}

local data_path = vim.fn.stdpath("data") .. "/usage-tracker.json"
local lock_path = data_path .. ".lock"
local ns = vim.api.nvim_create_namespace("usage_tracker")
local MODES = { "n", "v", "x", "s", "o", "i", "t", "c" }

-- flush まではメモリ上の増分のみ保持し、flush時にディスクへ読み直しマージする
-- （複数nvimインスタンスが同時に書いてもカウントが上書きで消えにくいようにするため）
M._delta = { keys = {}, plugins = {} }

local registry, registry_max_len, prefix_set = {}, 0, {}
local buf_registry, buf_registry_max_len, buf_prefix_set = {}, 0, {}
local loaded_this_session = {}
local did_setup = false

-- ローリングバッファと、prefix上で足踏み中の「未確定な最長一致」
-- pending.ext は最初のマッチ文字列から続く実際の継続バイト列（直接ルックアップに使う）
local keybuf = ""
local last_hr = nil
local pending = nil -- { mode, display, ext }

local function today()
  return os.date("%Y-%m-%d")
end

local function is_excluded(lhs)
  if not lhs or lhs == "" then
    return true
  end
  if lhs:sub(1, 6) == "<Plug>" then
    return true
  end
  if lhs:find("<SNR>", 1, true) then
    return true
  end
  return false
end

-- getter(mode) は nvim_get_keymap / nvim_buf_get_keymap(0, mode) を渡す想定。
-- 戻り値のprefixesは各lhsrawの真の接頭辞ごとの「その接頭辞を持つ全lhsのmodes和集合」
-- （gc/gccのような曖昧性判定を、別modeにしか無い長いマッピングで誤爆させないため）。
local function build_registry(getter)
  local reg = {}
  local prefixes = {}
  local max_len = 0
  for _, m in ipairs(MODES) do
    local ok, maps = pcall(getter, m)
    if ok and maps then
      for _, mp in ipairs(maps) do
        local lhs = mp.lhs
        if not is_excluded(lhs) then
          local lhsraw = mp.lhsraw
          if not lhsraw or lhsraw == "" then
            lhsraw = vim.api.nvim_replace_termcodes(lhs, true, true, true)
          end
          local entry = reg[lhsraw]
          if not entry then
            entry = { modes = {}, display = lhs, nowait = false }
            reg[lhsraw] = entry
            if #lhsraw > max_len then
              max_len = #lhsraw
            end
          end
          entry.modes[m] = true
          if mp.nowait == 1 then
            entry.nowait = true
          end
          -- 同一lhsrawが複数modeに跨る場合があるため、prefix集合はmodeループの
          -- 出現ごとに毎回更新する（"if not entry"の中に閉じ込めるとunionが漏れる）
          for k = 1, #lhsraw - 1 do
            local pfx = lhsraw:sub(1, k)
            local pset = prefixes[pfx]
            if not pset then
              pset = {}
              prefixes[pfx] = pset
            end
            pset[m] = true
          end
        end
      end
    end
  end
  return reg, max_len, prefixes
end

local function rebuild_global_registry()
  registry, registry_max_len, prefix_set = build_registry(vim.api.nvim_get_keymap)
end

local function rebuild_buf_registry()
  buf_registry, buf_registry_max_len, buf_prefix_set = build_registry(function(m)
    return vim.api.nvim_buf_get_keymap(0, m)
  end)
end

-- Visual/Select系(v/V/^V/s/S/^S/x)は v(Visual+Select) 登録側で拾えるようにまとめる。
-- operator-pending(no/nov/noV/no^V)は先頭1文字が"n"になり誤分類するため"o"に正規化する。
-- それ以外はそのまま先頭1文字。
local function current_mode_char()
  local full = vim.api.nvim_get_mode().mode
  local head = full:sub(1, 1)
  if head == "v" or head == "V" or head == "\22" or head == "s" or head == "S" or head == "\19" or head == "x" then
    return "v"
  end
  if full:sub(1, 2) == "no" then
    return "o"
  end
  return head
end

-- modes_set(entry.modesあるいはprefixのmodes和集合)に対して、正規化後のmodeを
-- 受理するかどうかの共通規則。v→v/x/sいずれか、o→o、それ以外は完全一致。
local function modes_set_accepts(modes_set, mode)
  if not modes_set then
    return false
  end
  if mode == "v" then
    return modes_set.v or modes_set.x or modes_set.s
  end
  if mode == "o" then
    return modes_set.o
  end
  return modes_set[mode]
end

local function mode_accepts(entry, mode)
  return modes_set_accepts(entry.modes, mode)
end

-- sが「別modeにしか存在しない長いマッピング」のprefixで誤爆しないよう、
-- prefix集合側もmode_acceptsと同じ受理規則で判定する。
local function is_prefix_candidate(s, mode)
  return modes_set_accepts(prefix_set[s], mode) or modes_set_accepts(buf_prefix_set[s], mode)
end

-- 各長さでbuffer-local優先(vimのマッピング解決セマンティクスに合わせる)、
-- mode_acceptsを通るものだけ候補にする。entry自体を返す(nowait参照のため)。
local function find_longest_match(buf, mode)
  local max_l = math.min(#buf, math.max(registry_max_len, buf_registry_max_len))
  for l = max_l, 1, -1 do
    local candidate = buf:sub(-l)
    local buf_entry = buf_registry[candidate]
    if buf_entry and mode_accepts(buf_entry, mode) then
      return candidate, buf_entry
    end
    local g_entry = registry[candidate]
    if g_entry and mode_accepts(g_entry, mode) then
      return candidate, g_entry
    end
  end
  return nil
end

-- pending.ext向けの完全一致版（サフィックス探索ではなく単一文字列の直接ルックアップ）。
-- buffer-local優先の順序はfind_longest_matchと揃える。
local function lookup_exact(s, mode)
  local buf_entry = buf_registry[s]
  if buf_entry and mode_accepts(buf_entry, mode) then
    return buf_entry
  end
  local g_entry = registry[s]
  if g_entry and mode_accepts(g_entry, mode) then
    return g_entry
  end
  return nil
end

local function record_key(mode, display)
  local key = mode .. ":" .. display
  local d = M._delta.keys[key]
  if not d then
    d = { count = 0 }
    M._delta.keys[key] = d
  end
  d.count = d.count + 1
  d.last = today()
end

-- sが他のより長いマッピングのprefixなら即記録せず保留する（gc/gccのような組の曖昧性回避）。
-- nowaitな登録はvim自身がprefixを待たず即実行するため、prefix判定自体をスキップする。
local function handle_candidate(mode, s, entry)
  if not entry.nowait and is_prefix_candidate(s, mode) then
    pending = { mode = mode, display = entry.display, ext = s }
  else
    record_key(mode, entry.display)
    keybuf = ""
    pending = nil
  end
end

-- pendingを「vimが確定実行した」とみなして記録し、ローリングバッファを空にする。
-- タイムアウト経過時と、CursorMoved等の解決シグナル時の両方から呼ばれる共通処理。
local function flush_pending_and_clear()
  if pending then
    record_key(pending.mode, pending.display)
    pending = nil
  end
  keybuf = ""
end

local function evaluate(this_chunk)
  local mode = current_mode_char()
  if pending then
    pending.ext = pending.ext .. this_chunk
    local entry = lookup_exact(pending.ext, mode)
    if entry then
      -- 完全一致: 新candidateとして同じprefix判定(nowait考慮込み)を適用する
      handle_candidate(mode, pending.ext, entry)
    elseif is_prefix_candidate(pending.ext, mode) then
      -- まだ曖昧: 完全一致ではないがより長いマッピングのprefixではある。記録せず待つ。
    else
      -- 破断: vimは古いpending(mode/displayは生成時のまま)を確定実行済みなので記録し、
      -- 今回分のchunkだけで取り直して非pending経路を一度だけ適用する
      record_key(pending.mode, pending.display)
      pending = nil
      keybuf = this_chunk
      local s2, entry2 = find_longest_match(keybuf, mode)
      if s2 then
        handle_candidate(mode, s2, entry2)
      end
    end
  else
    local s, entry = find_longest_match(keybuf, mode)
    if s then
      handle_candidate(mode, s, entry)
    end
  end
end

-- ローリングバッファ。typed由来のキーのみ連結し、最長一致優先でレジストリ照合する。
local function on_key_body(typed)
  if not typed or typed == "" then
    return
  end
  local now = vim.uv.hrtime()
  if last_hr and (now - last_hr) / 1e6 >= vim.o.timeoutlen then
    flush_pending_and_clear()
  end
  last_hr = now
  keybuf = keybuf .. typed
  local cap = math.max(64, registry_max_len, buf_registry_max_len)
  if #keybuf > cap then
    keybuf = keybuf:sub(-cap)
  end
  evaluate(typed)
end

-- ディスクI/O。「不存在=空データで開始」と「存在するが読めない/壊れている=データ破損」を
-- 区別する。破損時は黙って上書きせず.corruptへ退避してから空データで継続する。
-- 退避自体が失敗した場合はエラーを投げてflush_once側でflushを中止させる
-- （中途半端な空データ書き込みで既存集計を消さないため）。
local function read_disk_safe()
  local stat_ok, stat = pcall(vim.uv.fs_stat, data_path)
  local exists = stat_ok and stat ~= nil
  if not exists then
    return { keys = {}, plugins = {}, started = today() }
  end

  local ok, result = pcall(function()
    local fd = io.open(data_path, "r")
    if not fd then
      error("usage-tracker: file exists but could not be opened")
    end
    local content = fd:read("*a")
    fd:close()
    if not content or content == "" then
      error("usage-tracker: file exists but is empty/unreadable")
    end
    return vim.json.decode(content)
  end)

  if ok and type(result) == "table" then
    result.keys = result.keys or {}
    result.plugins = result.plugins or {}
    result.started = result.started or today()
    return result
  end

  if not os.rename(data_path, data_path .. ".corrupt") then
    error("usage-tracker: corrupt data file could not be moved aside")
  end
  return { keys = {}, plugins = {}, started = today() }
end

-- tmpファイル書き込み→renameのアトミック書き込み。write/closeの失敗も検査する。
local function write_disk(data)
  return pcall(function()
    local tmp = string.format("%s.tmp.%d", data_path, vim.uv.os_getpid())
    local fd = assert(io.open(tmp, "w"))
    local write_ok, write_err = fd:write(vim.json.encode(data))
    if not write_ok then
      fd:close()
      os.remove(tmp)
      error("usage-tracker: write failed: " .. tostring(write_err))
    end
    local close_ok, close_err = fd:close()
    if not close_ok then
      os.remove(tmp)
      error("usage-tracker: close failed: " .. tostring(close_err))
    end
    if not os.rename(tmp, data_path) then
      os.remove(tmp)
      error("usage-tracker: rename failed")
    end
  end)
end

local function merge_delta(disk, delta)
  for k, v in pairs(delta.keys) do
    local cur = disk.keys[k] or { count = 0 }
    cur.count = cur.count + v.count
    cur.last = (cur.last and cur.last > v.last) and cur.last or v.last
    disk.keys[k] = cur
  end
  for k, v in pairs(delta.plugins) do
    local cur = disk.plugins[k] or { count = 0 }
    cur.count = cur.count + v.count
    cur.last = (cur.last and cur.last > v.last) and cur.last or v.last
    disk.plugins[k] = cur
  end
end

-- lockファイルの存在＝保持中とみなす排他制御。stale(30秒超)なら奪って再取得する。
-- 30秒のstale判定を誤って有効なlockごと奪うリスクはある（理論上は）が、critical
-- section(read→merge→write)はms級で終わるため実際には起こらない。本ツールは
-- 統計収集用途でありシビアな整合性は求めないため、この簡易判定で許容する。
local function acquire_lock()
  local stat_ok, stat = pcall(vim.uv.fs_stat, lock_path)
  if stat_ok and stat and stat.mtime and stat.mtime.sec then
    if os.time() - stat.mtime.sec > 30 then
      pcall(os.remove, lock_path)
    end
  end
  local open_ok, fd = pcall(vim.uv.fs_open, lock_path, "wx", 384)
  if open_ok and fd then
    pcall(vim.uv.fs_close, fd)
    return true
  end
  return false
end

local function release_lock()
  pcall(os.remove, lock_path)
end

-- read→merge→writeの成否をそのまま返す。write失敗時はdeltaを保持したままにする。
local function flush_once()
  if not next(M._delta.keys) and not next(M._delta.plugins) then
    return true
  end
  if not acquire_lock() then
    return false -- 他インスタンスが保持中。deltaは保持したまま次回に回す
  end
  local ok = pcall(function()
    local disk = read_disk_safe()
    merge_delta(disk, M._delta)
    if not write_disk(disk) then
      error("usage-tracker: write_disk failed")
    end
    M._delta = { keys = {}, plugins = {} }
  end)
  release_lock()
  return ok
end

-- is_vim_leave時のみロック競合(またはwrite失敗)を50ms間隔で最大3回リトライする
-- （プロセス終了前の最後の砦）
local function flush(is_vim_leave)
  if flush_once() then
    return
  end
  if not is_vim_leave then
    return
  end
  for _ = 1, 3 do
    vim.uv.sleep(50)
    if flush_once() then
      return
    end
  end
end

local function record_plugin_load(name)
  if type(name) ~= "string" or loaded_this_session[name] then
    return
  end
  loaded_this_session[name] = true
  local d = M._delta.plugins[name] or { count = 0 }
  d.count = d.count + 1
  d.last = today()
  M._delta.plugins[name] = d
end

function M.setup()
  if vim.g.usage_tracker_disable then
    return
  end
  if did_setup then
    return
  end
  did_setup = true

  local function initial_build()
    pcall(rebuild_global_registry)
    pcall(rebuild_buf_registry)
    -- setup前(lazy.nvim起動シーケンス中)に既にロード済みのプラグインはLazyLoadを
    -- 一度も発火しないため、ここで一括して初回記録する
    pcall(function()
      local ok, lazy = pcall(require, "lazy")
      if not ok then
        return
      end
      for _, p in ipairs(lazy.plugins()) do
        if p._ and p._.loaded then
          record_plugin_load(p.name)
        end
      end
    end)
  end

  local group = vim.api.nvim_create_augroup("usage_tracker", { clear = true })

  vim.api.nvim_create_autocmd("User", {
    group = group,
    pattern = "VeryLazy",
    once = true,
    callback = initial_build,
  })
  -- 引数無し起動(`nvim`単体)だとVeryLazyがこのsetup呼び出し中に発火済みで
  -- 上のautocmd登録が間に合わないため、次tickでの初回構築を保険として掛ける
  vim.schedule(initial_build)

  vim.api.nvim_create_autocmd("User", {
    group = group,
    pattern = "LazyLoad",
    callback = function(ev)
      pcall(rebuild_global_registry)
      record_plugin_load(ev.data)
    end,
  })

  vim.api.nvim_create_autocmd("BufEnter", {
    group = group,
    callback = function()
      pcall(rebuild_buf_registry)
    end,
  })

  -- LSP/ファイルタイプ/ターミナル起動時のバッファローカルマップ登録は、プラグイン側の
  -- 自前autocmdより後に確定することが多いため、schedule_wrapで次tickに遅らせて拾う
  vim.api.nvim_create_autocmd({ "LspAttach", "FileType", "TermOpen" }, {
    group = group,
    callback = vim.schedule_wrap(function()
      pcall(rebuild_buf_registry)
    end),
  })

  -- ネイティブコマンド境界を跨いだ誤マッチ（例: "gg"直後の"y"が"ggy"として
  -- サフィックス"gy"に誤カウントされる）を減らす。マッピング解決待ち(pending)中は
  -- これらのイベントは発火しないため、複数キーシーケンスの途中で切れる心配は無い。
  -- pendingが残っている場合はここが「vimが確定実行した」ことを知る唯一の機会になり
  -- うるため、クリア前に記録する（on_keyは次のキー入力が無いと発火しないため）。
  vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI", "TextChanged", "TextChangedI", "ModeChanged" }, {
    group = group,
    callback = function()
      pcall(flush_pending_and_clear)
    end,
  })

  vim.on_key(function(_, typed)
    local ok = pcall(on_key_body, typed)
    if not ok then
      keybuf = "" -- エラー時はバッファ状態を破棄し次の入力から再開する
      pending = nil
    end
  end, ns)

  local timer_ok, timer = pcall(vim.uv.new_timer)
  if timer_ok and timer then
    timer:start(
      5 * 60 * 1000,
      5 * 60 * 1000,
      vim.schedule_wrap(function()
        flush(false)
      end)
    )
  end

  vim.api.nvim_create_autocmd("VimLeavePre", {
    group = group,
    callback = function()
      pending = nil -- 未解決pendingは記録せず破棄する
      if timer_ok and timer then
        pcall(function()
          timer:stop()
          timer:close()
        end)
      end
      flush(true)
    end,
  })
end

-- テスト用エクスポート（挙動には影響しない。単体駆動確認のためだけに存在）
M._test = {
  current_mode_char = current_mode_char,
  mode_accepts = mode_accepts,
  is_prefix_candidate = is_prefix_candidate,
  on_key_body = on_key_body,
  merge_delta = merge_delta,
  flush_once = flush_once,
  acquire_lock = acquire_lock,
  release_lock = release_lock,
  flush_pending_and_clear = flush_pending_and_clear,
  get_delta = function()
    return M._delta
  end,
  get_pending = function()
    return pending
  end,
  set_registry = function(reg, max_len, prefixes)
    registry = reg or {}
    registry_max_len = max_len or 0
    prefix_set = prefixes or {}
  end,
  set_buf_registry = function(reg, max_len, prefixes)
    buf_registry = reg or {}
    buf_registry_max_len = max_len or 0
    buf_prefix_set = prefixes or {}
  end,
  reset_state = function()
    keybuf = ""
    last_hr = nil
    pending = nil
    M._delta = { keys = {}, plugins = {} }
  end,
  get_lock_path = function()
    return lock_path
  end,
  get_data_path = function()
    return data_path
  end,
  set_data_path = function(path)
    data_path = path
    lock_path = path .. ".lock"
  end,
  -- lockとdataを別々の場所に向けたい単体テスト向け（例: lock成功/write失敗を分離したいケース）
  set_lock_path = function(path)
    lock_path = path
  end,
}

return M
