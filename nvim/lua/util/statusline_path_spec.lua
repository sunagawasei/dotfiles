local sp = require("util.statusline_path")

-- 再現パス(表示幅67、dirは表示幅32、nameは表示幅35)
local repro = "~/.config/.claude/docs/research/fff-codediff-neovim-code-reading.md"

local shorten_cases = {
  { name = "avail=200: 全長そのまま", full = repro, avail = 200, expected = repro },
  { name = "avail=80: 全長そのまま(67<=80)", full = repro, avail = 80, expected = repro },
  {
    name = "avail=53: dirを維持しnameの中間だけ省略",
    full = repro,
    avail = 53,
    expected = "~/.config/.claude/docs/research/fff-codedi…reading.md",
  },
  {
    name = "avail=40: セグメントを2つ落とす",
    full = repro,
    avail = 40,
    expected = "~/…/docs/research/fff-codedif…reading.md",
  },
  {
    name = "avail=21: dir2=seg1のみ(~/…/)でちょうど届く境界",
    full = repro,
    avail = 21,
    expected = "~/…/fff-code…ading.md",
  },
  {
    name = "avail=20: dir2すら届かずディレクトリ表示を諦める境界",
    full = repro,
    avail = 20,
    expected = "fff-codedi…eading.md",
  },
  { name = "avail=1: 停止して定義どおり先頭1文字", full = repro, avail = 1, expected = "f" },
  { name = "avail=0: 空文字", full = repro, avail = 0, expected = "" },
  { name = "avail=-5: 空文字", full = repro, avail = -5, expected = "" },
  {
    name = "短い名前は余計にディレクトリを落とさない",
    full = "~/.config/a.md",
    avail = 20,
    expected = "~/.config/a.md",
  },
  {
    name = "dotfileは全長そのまま",
    full = "~/.config/zsh/.zshrc",
    avail = 22,
    expected = "~/.config/zsh/.zshrc",
  },
  {
    name = "短い名前+min_w分岐: dirを落としてもnameは省略しない",
    full = "~/verylong/a.md",
    avail = 8,
    expected = "~/…/a.md",
  },
  {
    name = "dotfile+min_w分岐: dirを落としてもnameは省略しない",
    full = "~/verylong/.gitignore",
    avail = 14,
    expected = "~/…/.gitignore",
  },
  {
    name = ".gitignoreを含むパスは収まるavailでディレクトリが落ちない",
    full = "~/project/.gitignore/nested/very-long-filename-for-elision-test.md",
    avail = 45,
    expected = "~/project/.gitignore/nested/very-lon…-test.md",
  },
  {
    name = "末尾がドットの名前(avail=4): tailにドットが残る",
    full = "~/dir/file.",
    avail = 4,
    expected = "fi….",
  },
  {
    name = "末尾がドットの名前(avail=3): head/tailとも1文字",
    full = "~/dir/file.",
    avail = 3,
    expected = "f….",
  },
  {
    name = "絶対パスはdir2が/…/x/の形になる",
    full = "/very/long/path/to/x/file.md",
    avail = 15,
    expected = "/…/to/x/file.md",
  },
  {
    name = "dirが~/だけの場合は落とせるセグメントが無く手順7へ抜ける",
    full = "~/a.md",
    avail = 3,
    expected = "a…d",
  },
  {
    name = "/を含まない名前(dir=空文字)",
    full = "onlyname.md",
    avail = 5,
    expected = "on…md",
  },
  {
    name = "全角のみのファイル名(avail=20): 手順7でelide",
    full = "~/.config/日本語のファイル名です.md",
    avail = 20,
    expected = "日本語のフ…名です.md",
  },
  {
    name = "全角のみのファイル名(avail=21): dir2=~/…/で届く",
    full = "~/.config/日本語のファイル名です.md",
    avail = 21,
    expected = "~/…/日本語の…です.md",
  },
  {
    name = "結合文字(e+U+0301)を含む名前はskipccで分断されない",
    full = "~/a/cafe\u{0301}nofile.md",
    avail = 10,
    expected = "cafe\u{0301}n…e.md",
  },
}

local total = 0

---@param case { name: string, full: string, avail: integer, expected: string }
local function assert_shorten_case(case)
  local actual = sp.shorten(case.full, case.avail)
  assert(
    actual == case.expected,
    string.format("%s: expected %q, got %q", case.name, case.expected, actual)
  )
end

for _, case in ipairs(shorten_cases) do
  assert_shorten_case(case)
  total = total + 1
end

-- 全角のみのファイル名: 文字が半分に割れていないこと(1文字ずつの表示幅合計が全体の表示幅と一致する)
do
  local result = sp.shorten("~/.config/日本語のファイル名です.md", 20)
  local sum = 0
  local n = vim.fn.strchars(result)
  for i = 0, n - 1 do
    sum = sum + vim.fn.strdisplaywidth(vim.fn.strcharpart(result, i, 1, true))
  end
  assert(
    sum == vim.fn.strdisplaywidth(result),
    string.format("zenkaku width sum mismatch: sum=%d, total=%d", sum, vim.fn.strdisplaywidth(result))
  )
  total = total + 1
end

-- 結合文字: 基底文字"e"と結合アキュート(U+0301, UTF-8で0xCC 0x81)が隣接していること(分断されていない)
do
  local result = sp.shorten("~/a/cafe\u{0301}nofile.md", 10)
  assert(result:find("e\204\129") ~= nil, string.format("combining mark got separated from base: %q", result))
  total = total + 1
end

-- 統合テスト: M.current()
do
  vim.o.columns = 93
  vim.go.laststatus = 3

  -- expand('%:p:~')が~に縮約されるにはHOME配下である必要があるため、実行時のHOMEから組み立てる。
  -- ~以降のパス(表示幅67)はavail=53の期待文字列がこの長さに依存しているため固定のまま。
  local target = vim.env.HOME .. "/.config/.claude/docs/research/fff-codediff-neovim-code-reading.md"
  vim.cmd("edit " .. vim.fn.fnameescape(target))
  local unmodified = sp.current()
  assert(
    unmodified == "~/.config/.claude/docs/research/fff-codedi…reading.md",
    string.format("current() unmodified: got %q", unmodified)
  )
  total = total + 1

  vim.bo.modified = true
  local modified = sp.current()
  assert(modified:sub(-4) == " [+]", string.format("current() modified suffix: got %q", modified))
  assert(
    vim.fn.strdisplaywidth(modified) <= 93 - 40,
    string.format("current() modified width exceeds budget: got %q (w=%d)", modified, vim.fn.strdisplaywidth(modified))
  )
  total = total + 2

  vim.bo.modified = false
  vim.bo.readonly = true
  local readonly = sp.current()
  assert(readonly:sub(-4) == " [-]", string.format("current() readonly suffix: got %q", readonly))
  assert(
    vim.fn.strdisplaywidth(readonly) <= 93 - 40,
    string.format("current() readonly width exceeds budget: got %q (w=%d)", readonly, vim.fn.strdisplaywidth(readonly))
  )
  total = total + 2
  vim.bo.readonly = false

  vim.bo.modified = true
  vim.bo.readonly = true
  local both = sp.current()
  assert(both:sub(-7) == " [+][-]", string.format("current() modified+readonly suffix order: got %q", both))
  assert(
    vim.fn.strdisplaywidth(both) <= 93 - 40,
    string.format("current() modified+readonly width exceeds budget: got %q (w=%d)", both, vim.fn.strdisplaywidth(both))
  )
  total = total + 2
  vim.bo.modified = false
  vim.bo.readonly = false

  vim.bo.modifiable = false
  local unmodifiable = sp.current()
  assert(unmodifiable:sub(-4) == " [-]", string.format("current() modifiable=false suffix: got %q", unmodifiable))
  total = total + 1
  vim.bo.modifiable = true

  vim.cmd("enew!")
  local noname = sp.current()
  assert(noname == "[No Name]", string.format("current() no-name buffer: got %q", noname))
  total = total + 1
end

print(string.format("statusline_path_spec: %d/%d passed", total, total))
