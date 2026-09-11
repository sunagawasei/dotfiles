// Command verify-herdr-deploy checks that herdr's deployment pipeline
// (dev tree -> patch files -> nix build -> running server) is actually in
// sync, one stage at a time, so a failing stage maps 1:1 to the next
// operation to run (patch + register / darwin-apply / restart herdr server).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sunagawasei/dotfiles/scripts/internal/herdrdeploy"
)

const (
	exitSuccess  = 0
	exitPolicyNG = 1
	exitInput    = 2
)

const (
	herdrLockNode  = "herdr"
	nixSystem      = "aarch64-darwin"
	maxListedFiles = 50
)

// unsupportedPathChars breaks either nix installable syntax ("#", "?") or
// git+file: URL syntax ("%", space) if they appear literally in a path we
// splice into those strings. Rejecting them outright is simpler and safer
// than trying to percent-encode an arbitrary local path correctly for both
// syntaxes at once.
const unsupportedPathChars = " #?%"

// commandRunner abstracts a subprocess invocation so stage 1's two-eval
// stability check can be exercised in tests without real nix/git.
type commandRunner func(name string, args ...string) (stdout, stderr string, err error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("verify-herdr-deploy", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var devTreeFlag, flakeFlag string
	var stageOnly int
	flags.StringVar(&devTreeFlag, "dev-tree", "", "herdrの開発ツリー(git worktree)のパス。環境変数 HERDR_DEV_TREE でも指定可")
	flags.StringVar(&flakeFlag, "flake", "", "設定リポジトリのパス(既定: $HOME/.config)")
	flags.IntVar(&stageOnly, "stage", 0, "指定した段だけ実行する(0=全段。段0は常に前提として実行する)")
	if err := flags.Parse(args); err != nil {
		return exitInput
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "verify-herdr-deploy: unexpected positional arguments: %s\n", strings.Join(flags.Args(), " "))
		return exitInput
	}
	if stageOnly < 0 || stageOnly > 3 {
		fmt.Fprintln(stderr, "verify-herdr-deploy: --stage は0〜3を指定する")
		return exitInput
	}

	devTree := devTreeFlag
	if devTree == "" {
		devTree = os.Getenv("HERDR_DEV_TREE")
	}
	if devTree == "" {
		fmt.Fprintln(stderr, "verify-herdr-deploy: 開発ツリーのパスが未指定")
		fmt.Fprintln(stderr, "  --dev-tree <path> を指定するか、環境変数 HERDR_DEV_TREE を設定する")
		return exitInput
	}
	if abs, err := filepath.Abs(devTree); err == nil {
		devTree = abs
	}

	flakeRoot := flakeFlag
	if flakeRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(stderr, "verify-herdr-deploy: ホームディレクトリの解決に失敗: %v\n", err)
			return exitInput
		}
		flakeRoot = filepath.Join(home, ".config")
	}
	if abs, err := filepath.Abs(flakeRoot); err == nil {
		flakeRoot = abs
	}

	if err := rejectUnsupportedPathChars("--dev-tree", devTree); err != nil {
		fmt.Fprintln(stderr, "verify-herdr-deploy: "+err.Error())
		return exitInput
	}
	if err := rejectUnsupportedPathChars("--flake", flakeRoot); err != nil {
		fmt.Fprintln(stderr, "verify-herdr-deploy: "+err.Error())
		return exitInput
	}

	if !runStage0(stdout, devTree, flakeRoot) {
		fmt.Fprintln(stdout, "段0がngのため、段1〜3は実行しない。")
		return exitPolicyNG
	}

	runAll := stageOnly == 0
	overallOK := true

	if runAll || stageOnly == 1 {
		if !runStage1(stdout, devTree, flakeRoot, runCommand) {
			overallOK = false
		}
	}

	var actualRoot string
	stage2Ran := false
	if runAll || stageOnly == 2 {
		var ok2 bool
		ok2, actualRoot = runStage2(stdout, flakeRoot)
		stage2Ran = true
		if !ok2 {
			overallOK = false
		}
	}

	if runAll || stageOnly == 3 {
		if !stage2Ran {
			if root, _, err := resolveActiveBinary(); err == nil {
				actualRoot = root
			}
		}
		if !runStage3(stdout, actualRoot) {
			overallOK = false
		}
	}

	if overallOK {
		return exitSuccess
	}
	return exitPolicyNG
}

// --- 段0: 前提の検査 ---

func runStage0(stdout io.Writer, devTree, flakeRoot string) bool {
	fmt.Fprintln(stdout, "--- 段0: 前提の検査 ---")
	ok := true

	if _, _, err := runCommand("git", "-C", devTree, "rev-parse", "--show-toplevel"); err != nil {
		fmt.Fprintf(stdout, "ng: 開発ツリー %s が git の作業ツリーとして認識できない: %v\n", devTree, err)
		fmt.Fprintln(stdout, "次の操作: --dev-tree に正しい git 作業ツリーのパスを指定する")
		return false
	}
	fmt.Fprintf(stdout, "ok: 開発ツリー %s は git の作業ツリー\n", devTree)

	// git+file: の override は追跡済みファイルの未commit変更も含む(実測確認済み)ため、
	// 未追跡ファイルだけでなく status --porcelain 全体を見る。
	statusOut, _, err := runCommand("git", "-C", devTree, "status", "--porcelain")
	if err != nil {
		fmt.Fprintf(stdout, "ng: 開発ツリーの git status 取得に失敗: %v\n", err)
		ok = false
	} else if strings.TrimSpace(statusOut) != "" {
		fmt.Fprintln(stdout, "ng: 開発ツリーに未commitの変更(未追跡または変更済み追跡ファイル)がある:")
		printRawLines(stdout, statusOut)
		fmt.Fprintln(stdout, "次の操作: 開発ツリーをcommitするか、対象ファイルを元に戻す")
		ok = false
	} else {
		fmt.Fprintln(stdout, "ok: 開発ツリーに未commitの変更なし")
	}

	// --exclude-standard は無い: home-manager/patches/*.patch がグローバルの
	// exclude で無視される設定になっていても、Nix flakeはgit追跡状態だけを見る
	// (ignore有無は無関係)ため、ignore対象でも未追跡なら検出しなければならない。
	patchesRel := filepath.Join("home-manager", "patches")
	herdrNixRel := filepath.Join("home-manager", "herdr.nix")
	untrackedOut, _, err := runCommand("git", "-C", flakeRoot, "ls-files", "--others", patchesRel, herdrNixRel)
	if err != nil {
		fmt.Fprintf(stdout, "ng: %s の未追跡ファイル確認に失敗: %v\n", flakeRoot, err)
		ok = false
	} else if strings.TrimSpace(untrackedOut) != "" {
		fmt.Fprintln(stdout, "ng: 以下がgit未追跡のため、Nix flakeの評価に乗らない:")
		printRawLines(stdout, untrackedOut)
		fmt.Fprintln(stdout, "次の操作: 上記を git add してから再実行する")
		ok = false
	} else {
		fmt.Fprintln(stdout, "ok: home-manager/patches, home-manager/herdr.nix に未追跡ファイルなし")
	}

	lockPath := filepath.Join(flakeRoot, "flake.lock")
	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		fmt.Fprintf(stdout, "ng: %s を読めない: %v\n", lockPath, err)
		return false
	}
	rev, err := herdrdeploy.LockedRev(lockData, herdrLockNode)
	if err != nil {
		fmt.Fprintf(stdout, "ng: flake.lock から herdr の locked.rev を取得できない: %v\n", err)
		return false
	}

	if _, ancestorStderr, err := runCommand("git", "-C", devTree, "merge-base", "--is-ancestor", rev, "HEAD"); err != nil {
		fmt.Fprintf(stdout, "ng: 開発ツリーが upstream rev %s を含まない\n", rev)
		if strings.TrimSpace(ancestorStderr) != "" {
			fmt.Fprintf(stdout, "  詳細: %s\n", strings.TrimSpace(ancestorStderr))
		}
		fmt.Fprintln(stdout, "次の操作: --dev-tree が指す先が正しいか確認する(upstream v0.8.0を含まないbranchを指していないか)")
		return false
	}
	fmt.Fprintf(stdout, "ok: 開発ツリーは upstream rev %s を含む\n", rev)

	_, diffStderr, diffErr := runCommand("git", "-C", devTree, "diff", "--quiet", rev, "HEAD", "--", "nix", "flake.nix", "flake.lock")
	if diffErr == nil {
		fmt.Fprintln(stdout, "ok: 開発ツリーの nix/, flake.nix, flake.lock は upstream rev と同一")
	} else if exitErr, isExit := diffErr.(*exec.ExitError); isExit && exitErr.ExitCode() == 1 {
		fmt.Fprintln(stdout, "ng: 開発ツリーの nix/, flake.nix, flake.lock が upstream rev と異なる(fileset評価の土台が違う)")
		statOut, _, _ := runCommand("git", "-C", devTree, "diff", "--stat", rev, "HEAD", "--", "nix", "flake.nix", "flake.lock")
		printRawLines(stdout, statOut)
		ok = false
	} else {
		fmt.Fprintf(stdout, "ng: nix/, flake.nix, flake.lock の差分確認に失敗: %v %s\n", diffErr, strings.TrimSpace(diffStderr))
		ok = false
	}

	return ok
}

// --- 段1: 開発ツリー ↔ パッチファイル ---

func runStage1(stdout io.Writer, devTree, flakeRoot string, run commandRunner) bool {
	fmt.Fprintln(stdout, "--- 段1: 開発ツリー ↔ パッチファイル ---")

	configPath, configStderr, err := run("nix", "build",
		flakeAttr(flakeRoot, "herdr-src-patched"),
		"--no-link", "--print-out-paths", "--no-write-lock-file")
	if err != nil {
		fmt.Fprintf(stdout, "ng: 設定側ツリー(herdr-src-patched)のbuildに失敗: %v\n%s\n", err, configStderr)
		return false
	}

	overrideArg := "git+file://" + devTree
	evalAttr := flakeAttr(flakeRoot, "herdr-src-unpatched") + ".outPath"

	start := time.Now()
	devPath1, devStderr, err := run("nix", "eval", evalAttr, "--raw", "--no-write-lock-file", "--override-input", "herdr", overrideArg)
	firstDuration := time.Since(start)
	if err != nil {
		fmt.Fprintf(stdout, "ng: 開発ツリー側(--override-input)のNix評価(1回目)に失敗: %v\n%s\n", err, devStderr)
		return false
	}

	start = time.Now()
	devPath2, devStderr, err := run("nix", "eval", evalAttr, "--raw", "--no-write-lock-file", "--override-input", "herdr", overrideArg)
	secondDuration := time.Since(start)
	if err != nil {
		fmt.Fprintf(stdout, "ng: 開発ツリー側(--override-input)のNix評価(2回目)に失敗: %v\n%s\n", err, devStderr)
		return false
	}
	fmt.Fprintf(stdout, "所要時間: 1回目 %s / 2回目 %s\n", firstDuration.Round(time.Millisecond), secondDuration.Round(time.Millisecond))
	if devPath1 != devPath2 {
		// 段0の後(2回の評価の間)に開発ツリーが変化したことを示す。どちらのstore
		// pathも信用できないので、警告に留めず不安定として非ゼロにする。
		fmt.Fprintf(stdout, "ng: 同一overrideの評価が2回で異なるstore pathを返した(%s / %s)\n", devPath1, devPath2)
		fmt.Fprintln(stdout, "次の操作: 開発ツリーへの書き込みが無い状態で再実行する")
		return false
	}

	diff, err := herdrdeploy.DiffTrees(configPath, devPath2)
	if err != nil {
		fmt.Fprintf(stdout, "ng: ツリー比較に失敗: %v\n", err)
		return false
	}

	return printStage1Diff(stdout, diff)
}

// printStage1Diff prints the stage 1 diff and, per direction, the operation
// that follows from it. The three directions imply different — sometimes
// opposite — next actions, so a single blanket "パッチ化して登録する" is
// wrong whenever the extra content on the dev-tree side isn't a new,
// not-yet-registered change but something intentionally dropped from
// deployment (its patch removed from home-manager/herdr.nix while the
// commit that produced it is still sitting in the dev branch).
func printStage1Diff(stdout io.Writer, diff herdrdeploy.TreeDiff) bool {
	if diff.Empty() {
		fmt.Fprintln(stdout, "ok: 設定側ツリーと開発ツリーに差分なし")
		return true
	}

	fmt.Fprintf(stdout, "ng: 差分あり(%s)\n", diff.Shape())

	if len(diff.OnlyInConfigTree) > 0 {
		printFileList(stdout, "設定側のみに存在(登録済みpatchesの効果に対応するcommitが開発ツリーに無い)", diff.OnlyInConfigTree)
		fmt.Fprintln(stdout, "  次の操作: 対応するcommitを開発branchへ戻すか、この変更を配備から外す(home-manager/herdr.nixのpatchesリストから該当patchを除く)")
	}
	if len(diff.OnlyInDevTree) > 0 {
		printFileList(stdout, "開発ツリーのみに存在(パッチ未登録の新しい変更、または配備から意図的に外した変更のcommitが開発ツリーに残っている)", diff.OnlyInDevTree)
		fmt.Fprintln(stdout, "  次の操作: 新しい変更ならパッチ化してhome-manager/patches/へ置きherdr.nixへ登録する。配備から意図的に外した変更が開発ツリーに残っているだけなら、開発branch側のcommitを外す(無条件でパッチ登録しない)")
	}
	if len(diff.Changed) > 0 {
		printFileList(stdout, "両方に存在するが内容が異なる(差分の向きだけでは正誤が決まらない)", diff.Changed)
		fmt.Fprintln(stdout, "  次の操作: 設定側(登録済みpatches)と開発ツリーのどちらを正とするか人が判断し、正しい側に合わせてpatchまたは開発branchのcommitを直す")
	}

	return false
}

// --- 段2: 設定 ↔ ビルド済みシステム ---

func runStage2(stdout io.Writer, flakeRoot string) (bool, string) {
	fmt.Fprintln(stdout, "--- 段2: 設定 ↔ ビルド済みシステム ---")

	expected, expectedStderr, err := runCommand("nix", "eval",
		flakeAttr(flakeRoot, "herdr-patched")+".outPath", "--raw", "--no-write-lock-file")
	if err != nil {
		fmt.Fprintf(stdout, "ng: 期待するstore pathの評価に失敗: %v\n%s\n", err, expectedStderr)
		return false, ""
	}

	actualRoot, actualFull, err := resolveActiveBinary()
	if err != nil {
		fmt.Fprintf(stdout, "ng: 実際に有効なバイナリの解決に失敗: %v\n", err)
		return false, ""
	}
	fmt.Fprintf(stdout, "期待: %s\n実際: %s (%s)\n", expected, actualRoot, actualFull)

	ok := expected == actualRoot
	if ok {
		fmt.Fprintln(stdout, "ok: 設定とビルド済みシステムが一致")
	} else {
		fmt.Fprintln(stdout, "ng: 設定とビルド済みシステムが不一致")
		fmt.Fprintln(stdout, "次の操作: darwin-apply を実行する")
	}

	if lookedUp, err := exec.LookPath("herdr"); err == nil {
		if resolved, err := filepath.EvalSymlinks(lookedUp); err == nil {
			if root, isHerdrBin := herdrdeploy.TrimBinHerdrSuffix(resolved); isHerdrBin && root != actualRoot {
				fmt.Fprintf(stdout, "警告: command -v herdr の解決先(%s)が判定基準と異なる(shellのPATH次第で変わるため判定には使わない)\n", resolved)
			}
		}
	}

	return ok, actualRoot
}

func resolveActiveBinary() (storeRoot, resolvedFull string, err error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", "", fmt.Errorf("現在のユーザーを解決できない: %w", err)
	}
	linkPath := filepath.Join("/etc/profiles/per-user", currentUser.Username, "bin", "herdr")
	resolved, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		return "", "", fmt.Errorf("%s の解決に失敗: %w", linkPath, err)
	}
	root, ok := herdrdeploy.TrimBinHerdrSuffix(resolved)
	if !ok {
		return "", resolved, fmt.Errorf("解決先 %s が /bin/herdr で終わらない", resolved)
	}
	return root, resolved, nil
}

// --- 段3: ビルド済みシステム ↔ 稼働プロセス ---

func runStage3(stdout io.Writer, actualRoot string) bool {
	fmt.Fprintln(stdout, "--- 段3: ビルド済みシステム ↔ 稼働プロセス ---")

	psOutput, psStderr, err := runCommand("ps", "-axo", "pid=,args=")
	if err != nil {
		fmt.Fprintf(stdout, "ng: ps の実行に失敗: %v\n%s\n", err, psStderr)
		return false
	}
	servers, others := herdrdeploy.ClassifyHerdrProcesses(psOutput)
	return evaluateStage3(stdout, actualRoot, servers, others)
}

// evaluateStage3 takes already-classified processes so it can be tested
// without shelling out to a real ps. Judgment needs only the pid,
// server-or-not classification, and the resolved store path — never the
// full argv, which a report-worthy verify-herdr-deploy invocation could
// otherwise leak (herdr's argv isn't guaranteed free of secret-looking
// values in every invocation shape).
func evaluateStage3(stdout io.Writer, actualRoot string, servers, others []herdrdeploy.Process) bool {
	if len(servers) == 0 {
		fmt.Fprintln(stdout, "skip: herdr server プロセスが無い")
		listOtherProcesses(stdout, others)
		return true
	}

	if actualRoot == "" {
		fmt.Fprintln(stdout, "ng: 段2の実際に有効なバイナリを解決できていないため判定不能")
		listOtherProcesses(stdout, others)
		return false
	}

	ok := true
	for _, server := range servers {
		lsofOutput, lsofStderr, err := runCommand("lsof", "-p", strconv.Itoa(server.PID))
		if err != nil {
			// lsofはプロセス消滅・権限不足でも非ゼロを返す。「一致」と誤認せず不明として扱う。
			fmt.Fprintf(stdout, "ng: pid %d (server) の lsof 実行に失敗(不明。processの終了などのraceの可能性): %v %s\n", server.PID, err, strings.TrimSpace(lsofStderr))
			ok = false
			continue
		}
		path, status, _ := herdrdeploy.FindExecutableTxtPath(lsofOutput)
		switch status {
		case herdrdeploy.TxtFound:
			root, _ := herdrdeploy.TrimBinHerdrSuffix(path)
			if root == actualRoot {
				fmt.Fprintf(stdout, "ok: pid %d (server) は現在有効なバイナリと一致\n", server.PID)
			} else {
				fmt.Fprintf(stdout, "ng: pid %d (server) は現在有効なバイナリと不一致\n  稼働中: %s\n  現在有効: %s/bin/herdr\n", server.PID, path, actualRoot)
				fmt.Fprintln(stdout, "次の操作: herdr server を再起動する")
				ok = false
			}
		case herdrdeploy.TxtUnreadable:
			fmt.Fprintf(stdout, "ng: pid %d (server) の実体を読めなかった(不明)\n", server.PID)
			ok = false
		default:
			fmt.Fprintf(stdout, "ng: pid %d (server) の lsof 出力に herdr バイナリのtxt行が見つからない\n", server.PID)
			ok = false
		}
	}

	listOtherProcesses(stdout, others)
	return ok
}

func listOtherProcesses(stdout io.Writer, others []herdrdeploy.Process) {
	if len(others) == 0 {
		return
	}
	fmt.Fprintln(stdout, "herdr の他のプロセス(判定には使わない):")
	for _, p := range others {
		lsofOutput, _, err := runCommand("lsof", "-p", strconv.Itoa(p.PID))
		if err != nil {
			fmt.Fprintf(stdout, "  pid %d: lsof失敗\n", p.PID)
			continue
		}
		path, status, _ := herdrdeploy.FindExecutableTxtPath(lsofOutput)
		if status == herdrdeploy.TxtFound {
			fmt.Fprintf(stdout, "  pid %d: %s\n", p.PID, path)
		} else {
			fmt.Fprintf(stdout, "  pid %d: 実体不明\n", p.PID)
		}
	}
}

// --- 共通ヘルパー ---

func runCommand(name string, args ...string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), errBuf.String(), err
}

func printRawLines(stdout io.Writer, block string) {
	for line := range strings.SplitSeq(strings.TrimRight(block, "\n"), "\n") {
		if line != "" {
			fmt.Fprintf(stdout, "  %s\n", line)
		}
	}
}

func rejectUnsupportedPathChars(label, path string) error {
	if strings.ContainsAny(path, unsupportedPathChars) {
		return fmt.Errorf("%s のパスに未対応の文字(space, '#', '?', '%%')が含まれる: %s", label, path)
	}
	return nil
}

func flakeAttr(flakeRoot, name string) string {
	return fmt.Sprintf("%s#packages.%s.%s", flakeRoot, nixSystem, name)
}

func printFileList(stdout io.Writer, label string, files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Fprintf(stdout, "  %s (%d件):\n", label, len(files))
	shown := files
	truncated := false
	if len(shown) > maxListedFiles {
		shown = shown[:maxListedFiles]
		truncated = true
	}
	for _, f := range shown {
		fmt.Fprintf(stdout, "    %s\n", f)
	}
	if truncated {
		fmt.Fprintf(stdout, "    ...他%d件\n", len(files)-maxListedFiles)
	}
}
