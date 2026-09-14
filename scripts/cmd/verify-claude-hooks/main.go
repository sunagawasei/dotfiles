// Command verify-claude-hooks checks that ~/.config/claude/settings.json still
// registers the hooks this repo depends on, and that the binaries they point
// to actually exist and are executable. settings.json is .gitignore'd (it
// carries a local path), so recreating it silently drops guard hooks like
// bounded-background with no other signal.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	exitSuccess  = 0
	exitPolicyNG = 1
	exitInput    = 2
)

// expectedHook describes one hook registration this repo requires to be
// present in settings.json. Other events (SessionStart等) are out of scope
// on purpose since they change independently of this list.
type expectedHook struct {
	Event      string
	Matcher    string
	BinaryName string
}

var expectedHooks = []expectedHook{
	{Event: "PreToolUse", Matcher: "Bash", BinaryName: "bounded-background"},
	{Event: "PostToolUse", Matcher: "Write|Edit|MultiEdit|NotebookEdit", BinaryName: "add_newline"},
	{Event: "Notification", Matcher: "idle_prompt", BinaryName: "herdr-notify"},
}

// requiredBinaries lists binaries settings.json never registers directly but
// a registered hook still needs at runtime (bounded-background execs
// bg-deadline as a subprocess).
var requiredBinaries = []string{"bg-deadline"}

type settings struct {
	Hooks map[string][]hookMatcherEntry `json:"hooks"`
}

type hookMatcherEntry struct {
	Matcher string    `json:"matcher"`
	Hooks   []hookCmd `json:"hooks"`
}

type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func main() {
	configDir := os.Getenv("CLAUDE_CONFIG_DIR")
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "verify-claude-hooks: ホームディレクトリの解決に失敗:", err)
		os.Exit(exitInput)
	}
	if configDir == "" {
		configDir = filepath.Join(home, ".config", "claude")
	}
	repoRoot, err := resolveRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "verify-claude-hooks: repo rootの解決に失敗(git rev-parse --show-toplevel):", err)
		os.Exit(exitInput)
	}
	os.Exit(run(os.Stdout, filepath.Join(configDir, "settings.json"), home, repoRoot))
}

// resolveRepoRoot asks git for the toplevel of whatever repo the current
// working directory sits in, so the expected hook location below never
// bakes in a hardcoded personal path.
func resolveRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func run(stdout io.Writer, settingsPath, homeDir, repoRoot string) int {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Fprintf(stdout, "ng: %s を読めない: %v\n", settingsPath, err)
		return exitInput
	}
	return verify(stdout, data, homeDir, repoRoot)
}

func verify(stdout io.Writer, data []byte, homeDir, repoRoot string) int {
	var s settings
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintf(stdout, "ng: settings.json のparseに失敗: %v\n", err)
		return exitInput
	}

	overallOK := true
	for _, exp := range expectedHooks {
		if !checkHook(stdout, s, exp, homeDir, repoRoot) {
			overallOK = false
		}
	}
	for _, name := range requiredBinaries {
		if !checkRequiredBinary(stdout, repoRoot, name) {
			overallOK = false
		}
	}

	if overallOK {
		fmt.Fprintln(stdout, "ok: 期待する3件のフックとその依存バイナリが登録されており、バイナリはソースより新しい(mtime比較。ソースとの対応までは保証しない)")
		return exitSuccess
	}
	return exitPolicyNG
}

// checkHook evaluates the stages (registered / command型 / correct matcher /
// command文字列の完全一致 / 実行可能 / ビルドの新しさ) in order and stops at
// the first that fails, since each stage implies a different next action.
func checkHook(stdout io.Writer, s settings, exp expectedHook, homeDir, repoRoot string) bool {
	label := fmt.Sprintf("%s (event=%s)", exp.BinaryName, exp.Event)

	var foundMatcher, foundType, foundCmd string
	found := false
	for _, entry := range s.Hooks[exp.Event] {
		for _, h := range entry.Hooks {
			if commandBinaryName(h.Command) == exp.BinaryName {
				foundMatcher = entry.Matcher
				foundType = h.Type
				foundCmd = h.Command
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		fmt.Fprintf(stdout, "ng: %s の登録が無い\n", label)
		fmt.Fprintf(stdout, "  次の操作: settings.json の hooks.%s へ matcher %q で追記する\n", exp.Event, exp.Matcher)
		return false
	}

	if foundType != "command" {
		fmt.Fprintf(stdout, "ng: %s の type が command でない(実際: %q)\n", label, foundType)
		fmt.Fprintln(stdout, "  次の操作: settings.json の該当エントリの type を \"command\" に修正する")
		return false
	}

	if foundMatcher != exp.Matcher {
		fmt.Fprintf(stdout, "ng: %s の matcher が不一致(実際: %q, 期待: %q)\n", label, foundMatcher, exp.Matcher)
		fmt.Fprintln(stdout, "  次の操作: settings.json の該当エントリの matcher を修正する")
		return false
	}

	expectedPath := filepath.Join(repoRoot, "claude", "hooks", exp.BinaryName)
	if !commandMatchesExpectedPath(foundCmd, expectedPath, homeDir) {
		fmt.Fprintf(stdout, "ng: %s の command 文字列が期待と不一致(引数・コメント・別パスなど余計な内容がある)\n", label)
		fmt.Fprintf(stdout, "  期待する command: %s (先頭を \"~/\" にして$HOMEから同じ場所を指す形も可)\n  実際の command: %s\n", expectedPath, strings.TrimSpace(foundCmd))
		fmt.Fprintln(stdout, "  次の操作: settings.json の該当エントリの command を期待する形に直す")
		return false
	}

	if err := checkExecutable(expectedPath); err != nil {
		fmt.Fprintf(stdout, "ng: %s のバイナリが実行できない(%s): %v\n", label, expectedPath, err)
		fmt.Fprintln(stdout, "  次の操作: cd claude/hooks && go build -o . ./...")
		return false
	}

	if err := checkBuildFreshness(repoRoot, exp.BinaryName, expectedPath); err != nil {
		fmt.Fprintf(stdout, "ng: %s はソースの方が新しい(ビルドが古い): %v\n", label, err)
		fmt.Fprintln(stdout, "  次の操作: cd claude/hooks && go build -o . ./...")
		return false
	}

	fmt.Fprintf(stdout, "ok: %s -> %s\n", label, expectedPath)
	return true
}

// checkRequiredBinary checks a binary that settings.json never references
// directly (bounded-backgroundがexecで呼ぶbg-deadline) but this repo still
// needs present, executable, and built from its current source.
func checkRequiredBinary(stdout io.Writer, repoRoot, name string) bool {
	path := filepath.Join(repoRoot, "claude", "hooks", name)

	if err := checkExecutable(path); err != nil {
		fmt.Fprintf(stdout, "ng: 必要なバイナリ %s が無い(bounded-backgroundの実行に必要): %v\n", name, err)
		fmt.Fprintln(stdout, "  次の操作: cd claude/hooks && go build -o . ./...")
		return false
	}

	if err := checkBuildFreshness(repoRoot, name, path); err != nil {
		fmt.Fprintf(stdout, "ng: 必要なバイナリ %s はソースの方が新しい(ビルドが古い): %v\n", name, err)
		fmt.Fprintln(stdout, "  次の操作: cd claude/hooks && go build -o . ./...")
		return false
	}

	fmt.Fprintf(stdout, "ok: 必要なバイナリ %s -> %s\n", name, path)
	return true
}

// commandBinaryName extracts the trailing binary name from a hook command
// string, since some entries prefix it with an interpreter (e.g. "bash
// ~/.claude/hooks/foo.sh"). It is only used to locate the entry meant for a
// given hook so a mismatch further down can be reported with useful detail;
// commandMatchesExpectedPath below decides pass/fail.
func commandBinaryName(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	return filepath.Base(fields[len(fields)-1])
}

// commandMatchesExpectedPath requires the whole command string (after
// trimming outer whitespace and expanding a leading "~") to equal
// expectedPath exactly, so an interpreter prefix, trailing argument, or
// comment defeats it instead of silently passing.
func commandMatchesExpectedPath(command, expectedPath, homeDir string) bool {
	trimmed := strings.TrimSpace(command)
	if trimmed == "~" {
		return homeDir == expectedPath
	}
	if rest, ok := strings.CutPrefix(trimmed, "~/"); ok {
		return filepath.Join(homeDir, rest) == expectedPath
	}
	return trimmed == expectedPath
}

func checkExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("ディレクトリであり実行ファイルではない")
	}
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("実行権限が無い")
	}
	return nil
}

// checkBuildFreshness reports an error when binaryPath is older than any
// *.go file in its source directory, catching a stale build that a plain
// existence+executable check can't see.
func checkBuildFreshness(repoRoot, name, binaryPath string) error {
	sourceDir := filepath.Join(repoRoot, "claude", "hooks", "cmd", name)
	sources, err := filepath.Glob(filepath.Join(sourceDir, "*.go"))
	if err != nil {
		return fmt.Errorf("ソースファイルの列挙に失敗(%s): %w", sourceDir, err)
	}
	if len(sources) == 0 {
		return fmt.Errorf("ソースファイルが見つからない: %s", sourceDir)
	}

	var newestSource time.Time
	for _, src := range sources {
		info, err := os.Stat(src)
		if err != nil {
			return err
		}
		if info.ModTime().After(newestSource) {
			newestSource = info.ModTime()
		}
	}

	binInfo, err := os.Stat(binaryPath)
	if err != nil {
		return err
	}
	if binInfo.ModTime().Before(newestSource) {
		return fmt.Errorf("ソース更新 %s > ビルド更新 %s", newestSource.Format(time.RFC3339), binInfo.ModTime().Format(time.RFC3339))
	}
	return nil
}
