package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type fakeCommandRunner struct {
	generate func(context.Context, cursorRequest) error
	herdr    func(context.Context, ...string) ([]byte, error)
}

func (f *fakeCommandRunner) GenerateTitle(ctx context.Context, request cursorRequest) error {
	if f.generate == nil {
		return errors.New("unexpected cursor-agent call")
	}
	return f.generate(ctx, request)
}

func (f *fakeCommandRunner) RunHerdr(ctx context.Context, args ...string) ([]byte, error) {
	if f.herdr == nil {
		return nil, errors.New("unexpected herdr call")
	}
	return f.herdr(ctx, args...)
}

func testApplication(t *testing.T, commands commandRunner) *application {
	t.Helper()
	return &application{
		tmpDir:     t.TempDir(),
		now:        func() time.Time { return time.Unix(1_700_000_000, 123) },
		environ:    func() []string { return []string{"PATH=/bin", "KEEP=value"} },
		executable: func() (string, error) { return "/test/herdr-title", nil },
		commands:   commands,
		startDetached: func(string, string) error {
			return errors.New("unexpected worker start")
		},
	}
}

func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	data := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeRawState(t *testing.T, app *application, workspaceID, data string) {
	t.Helper()
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(app.generationPath(workspaceID), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readPayload(t *testing.T, path string) workerPayload {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload workerPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func hookJSON(prompt, transcriptPath string) string {
	data, _ := json.Marshal(hookInput{Prompt: prompt, TranscriptPath: transcriptPath})
	return string(data)
}

func seedGenerationForTest(app *application, workspaceID, tabID string) (int64, error) {
	lock, err := app.acquireWorkspaceLock(workspaceID, lockTimeout)
	if err != nil {
		return 0, err
	}
	defer lock.release()
	state, err := app.readState(workspaceID)
	if err != nil {
		return 0, err
	}
	entry := state[tabID]
	entry.Generation = nextGeneration(entry, app.now().UnixNano())
	state[tabID] = entry
	if err := app.writeState(workspaceID, state); err != nil {
		return 0, err
	}
	return entry.Generation, nil
}

func TestSanitizeTitle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "CRLF tab and separators", input: "A\r\nB\tC\u2028D\u2029E", want: "A B C D E"},
		{name: "C0 and C1 controls", input: "A\x00\x1f\x7f\u0085B", want: "AB"},
		{name: "bidi and zero width", input: "A\u202eB\u200bC", want: "ABC"},
		{name: "NFC latin", input: "Café", want: "Café"},
		{name: "NFD latin", input: "Cafe\u0301", want: "Cafe\u0301"},
		{name: "combining dakuten", input: "は\u3099", want: "は\u3099"},
		{name: "consecutive marks", input: "e\u0301\u0327\u0304", want: "e\u0301\u0327\u0304"},
		{name: "leading mark", input: "\u0301A", want: "A"},
		{name: "symbol then variation selector", input: "☀️", want: ""},
		{name: "extended variation selector", input: "A\U000E0100", want: "A"},
		{name: "spacing mark", input: "A\u093e", want: "A"},
		{name: "keycap", input: "1️⃣", want: "1"},
		{name: "emoji sequence", input: "開発👨‍💻改善", want: "開発改善"},
		{name: "quotes only", input: "「」\"'", want: ""},
		{name: "spaces only", input: " \u3000\t\n", want: ""},
		{name: "allowed symbols without letter", input: "-_·", want: ""},
		{name: "allowed symbols", input: "Go-1_test·完了", want: "Go-1_test·完了"},
		{name: "collapse whitespace", input: "A  \u3000 B", want: "A B"},
		{name: "mark boundary drops entire base", input: "abcdefghijke\u0301\u0327", want: "abcdefghijk"},
		{name: "mark boundary cannot leave symbols only", input: "-----------a\u0301", want: ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := sanitizeTitle(test.input); got != test.want {
				t.Fatalf("sanitizeTitle(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestWorkspaceIDEscapingIsReversibleAndCollisionFree(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"a:b/c%": "a%3Ab%2Fc%25",
		"a/b":    "a%2Fb",
		"a%2Fb":  "a%252Fb",
		"a:b":    "a%3Ab",
		"a%3Ab":  "a%253Ab",
	}
	seen := make(map[string]string)
	for input, want := range cases {
		got := escapeWorkspaceID(input)
		if got != want {
			t.Errorf("escapeWorkspaceID(%q) = %q, want %q", input, got, want)
		}
		if previous, exists := seen[got]; exists {
			t.Errorf("%q and %q collide as %q", previous, input, got)
		}
		seen[got] = input
	}
}

func TestGenerationStateKeysAndNoOp(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})

	if _, err := seedGenerationForTest(app, "workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	if _, err := seedGenerationForTest(app, "workspace", "other-tab"); err != nil {
		t.Fatal(err)
	}
	if _, err := seedGenerationForTest(app, "workspace-only", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := seedGenerationForTest(app, "", "tab-only"); err != nil {
		t.Fatal(err)
	}

	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["tab"].Generation == 0 || state["other-tab"].Generation == 0 || len(state) != 2 {
		t.Fatalf("unexpected tab-keyed state: %#v", state)
	}
	workspaceOnly, err := app.readState("workspace-only")
	if err != nil || workspaceOnly[""].Generation == 0 {
		t.Fatalf("workspace-only state = %#v, err = %v", workspaceOnly, err)
	}
	tabOnly, err := app.readState("")
	if err != nil || tabOnly["tab-only"].Generation == 0 {
		t.Fatalf("tab-only state = %#v, err = %v", tabOnly, err)
	}

	noOp := testApplication(t, &fakeCommandRunner{})
	if err := noOp.runHook(strings.NewReader(`{"prompt":"secret"}`), "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(noOp.stateDir()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("state directory was created for no-op: %v", err)
	}
}

func TestStateReadsNewFormatAndMigratesLegacyEntries(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	now := app.now().UnixNano()
	legacyGeneration := now + 100
	writeRawState(t, app, "workspace", `{
		"legacy": `+jsonNumber(legacyGeneration)+`,
		"other_legacy": 42,
		"current": {"generation": 84, "title": "安定タイトル", "last_started_at": 83, "skipped_count": 2}
	}`)

	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if got := state["legacy"]; got != (tabState{Generation: legacyGeneration}) {
		t.Fatalf("legacy state = %#v", got)
	}
	if got := state["other_legacy"]; got != (tabState{Generation: 42}) {
		t.Fatalf("second legacy state = %#v", got)
	}
	wantCurrent := tabState{Generation: 84, Title: "安定タイトル", LastStartedAt: 83, SkippedCount: 2}
	if got := state["current"]; got != wantCurrent {
		t.Fatalf("current state = %#v, want %#v", got, wantCurrent)
	}

	var payload workerPayload
	app.startDetached = func(_ string, path string) error {
		payload = readPayload(t, path)
		return os.Remove(path)
	}
	if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "legacy", "workspace"); err != nil {
		t.Fatal(err)
	}
	if payload.Generation != legacyGeneration+1 {
		t.Fatalf("migrated generation = %d, want %d", payload.Generation, legacyGeneration+1)
	}
	state, err = app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["other_legacy"].Generation != 42 || state["current"] != wantCurrent {
		t.Fatalf("other entries were lost during migration: %#v", state)
	}
}

func jsonNumber(value int64) string {
	return strconv.FormatInt(value, 10)
}

func TestStateDropsOnlyInvalidEntries(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	now := app.now().UnixNano()
	writeRawState(t, app, "workspace", `{
		"valid": {"generation": 10, "title": "keep", "last_started_at": 9, "skipped_count": 1},
		"missing_generation": {"title": "bad"},
		"zero_generation": {"generation": 0},
		"negative_generation": {"generation": -1},
		"typed_generation": {"generation": "10"},
		"negative_started": {"generation": 10, "last_started_at": -1},
		"future_started": {"generation": 10, "last_started_at": `+jsonNumber(now+1)+`},
		"typed_started": {"generation": 10, "last_started_at": "9"},
		"negative_skipped": {"generation": 10, "skipped_count": -1},
		"typed_skipped": {"generation": 10, "skipped_count": "1"},
		"invalid_legacy": 0
	}`)

	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	want := tabState{Generation: 10, Title: "keep", LastStartedAt: 9, SkippedCount: 1}
	if len(state) != 1 || state["valid"] != want {
		t.Fatalf("filtered state = %#v, want only %#v", state, want)
	}
}

func TestBrokenStateFileBecomesEmptyMap(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	writeRawState(t, app, "workspace", `{not JSON`)
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if len(state) != 0 {
		t.Fatalf("broken state = %#v, want empty", state)
	}
}

func TestNextGenerationUsesMaximumOfCurrentPlusOneAndUnixNano(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		entry tabState
		now   int64
		want  int64
	}{
		{name: "clock wins", entry: tabState{Generation: 10}, now: 20, want: 20},
		{name: "current plus one wins", entry: tabState{Generation: 20}, now: 10, want: 21},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := nextGeneration(test.entry, test.now); got != test.want {
				t.Fatalf("nextGeneration(%#v, %d) = %d, want %d", test.entry, test.now, got, test.want)
			}
		})
	}
}

func inodeOf(t *testing.T, path string) uint64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("unsupported stat type %T", info.Sys())
	}
	return stat.Ino
}

func TestStateUpdatesPreserveLockAndGenerationInodes(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	if _, err := seedGenerationForTest(app, "workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	lockInode := inodeOf(t, app.lockPath("workspace"))
	generationInode := inodeOf(t, app.generationPath("workspace"))
	if _, err := seedGenerationForTest(app, "workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	if got := inodeOf(t, app.lockPath("workspace")); got != lockInode {
		t.Fatalf("lock inode changed from %d to %d", lockInode, got)
	}
	if got := inodeOf(t, app.generationPath("workspace")); got != generationInode {
		t.Fatalf("generation inode changed from %d to %d", generationInode, got)
	}
	for _, path := range []string{app.lockPath("workspace"), app.generationPath("workspace")} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode = %o, want 600", path, info.Mode().Perm())
		}
	}
}

func TestCleanupOnlyRemovesStalePayloadAndOutput(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	names := []string{"payload-old", "out-old", "ws_x.lock", "ws_x.gen", "unrelated"}
	old := app.now().Add(-staleFileAge)
	for _, name := range names {
		path := filepath.Join(app.stateDir(), name)
		if err := os.WriteFile(path, []byte("sensitive"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	recent := filepath.Join(app.stateDir(), "payload-recent")
	if err := os.WriteFile(recent, []byte("recent"), 0o600); err != nil {
		t.Fatal(err)
	}

	app.cleanupStaleFiles()
	for _, name := range []string{"payload-old", "out-old"} {
		if _, err := os.Stat(filepath.Join(app.stateDir(), name)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s was not removed: %v", name, err)
		}
	}
	for _, name := range []string{"ws_x.lock", "ws_x.gen", "unrelated", "payload-recent"} {
		if _, err := os.Stat(filepath.Join(app.stateDir(), name)); err != nil {
			t.Errorf("%s was removed: %v", name, err)
		}
	}
}

func TestTranscriptExtractionFiltersAndDeduplicates(t *testing.T) {
	t.Parallel()
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"first"}}`,
		`{"type":"user","message":{"content":[{"type":"text","text":"mixed"},{"type":"tool_result","content":"ignored"}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","content":"ignored"}]}}`,
		`{"type":"user","message":{"content":"<system-reminder>hidden</system-reminder> visible"}}`,
		`{"type":"user","message":{"content":"<system-reminder>only hidden</system-reminder>"}}`,
		`{"type":"user","isMeta":true,"message":{"content":"meta"}}`,
		`{"type":"user","isSynthetic":true,"message":{"content":"synthetic"}}`,
		`{"type":"user","isSidechain":true,"message":{"content":"subagent instruction"}}`,
		`{"type":"user","message":{"content":"current"}}`,
		`{"type":"user","message":{"content":"incomplete"}} trailing`,
	)
	want := []string{"first", "mixed", "visible", "current"}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, want) {
		t.Fatalf("buildTitleInputs() = %#v, want %#v", got, want)
	}

	withoutCurrent := writeTranscript(t, `{"type":"user","message":{"content":"past"}}`)
	if got := buildTitleInputs(withoutCurrent, "current", 0); !equalStrings(got, []string{"past", "current"}) {
		t.Fatalf("current prompt missing from transcript: %#v", got)
	}
}

func TestTranscriptExtractionKeepsAllOverlappingEdgeEntries(t *testing.T) {
	t.Parallel()
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"one"}}`,
		`{"type":"user","message":{"content":"two"}}`,
		`{"type":"user","message":{"content":"three"}}`,
		`{"type":"user","message":{"content":"four"}}`,
		`{"type":"user","message":{"content":"five"}}`,
	)
	want := []string{"one", "two", "three", "four", "five", "current"}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, want) {
		t.Fatalf("buildTitleInputs() = %#v, want %#v", got, want)
	}
}

func TestTitleInputsUseSessionEdgesInChronologicalOrder(t *testing.T) {
	t.Parallel()
	lines := make([]string, 0, 9)
	for index := 1; index <= 8; index++ {
		lines = append(lines, `{"type":"user","message":{"content":"message-`+strconv.Itoa(index)+`"}}`)
	}
	path := writeTranscript(t, lines...)
	want := []string{"message-1", "message-2", "message-3", "message-6", "message-7", "message-8", "current"}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, want) {
		t.Fatalf("edge inputs = %#v, want %#v", got, want)
	}
}

func TestTitleInputsWithThreeOrFewerHistoryEntries(t *testing.T) {
	t.Parallel()
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"one"}}`,
		`{"type":"user","message":{"content":"two"}}`,
		`{"type":"user","message":{"content":"three"}}`,
	)
	want := []string{"one", "two", "three", "current"}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, want) {
		t.Fatalf("short inputs = %#v, want %#v", got, want)
	}
}

func TestCurrentPromptIsRemovedBeforeSelectingEdges(t *testing.T) {
	t.Parallel()
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"one"}}`,
		`{"type":"user","message":{"content":"two"}}`,
		`{"type":"user","message":{"content":"three"}}`,
		`{"type":"user","message":{"content":"four"}}`,
		`{"type":"user","message":{"content":"five"}}`,
		`{"type":"user","message":{"content":"six"}}`,
		`{"type":"user","message":{"content":"seven"}}`,
		`{"type":"user","message":{"content":"current"}}`,
	)
	want := []string{"one", "two", "three", "five", "six", "seven", "current"}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, want) {
		t.Fatalf("deduplicated edge inputs = %#v, want %#v", got, want)
	}
}

func TestSkippedCountExpandsRecentInputsUpToTen(t *testing.T) {
	t.Parallel()
	lines := make([]string, 0, 15)
	for index := 1; index <= 15; index++ {
		lines = append(lines, `{"type":"user","message":{"content":"message-`+strconv.Itoa(index)+`"}}`)
	}
	path := writeTranscript(t, lines...)

	wantFiveRecent := []string{"message-1", "message-2", "message-3", "message-11", "message-12", "message-13", "message-14", "message-15", "current"}
	if got := buildTitleInputs(path, "current", 2); !equalStrings(got, wantFiveRecent) {
		t.Fatalf("expanded inputs = %#v, want %#v", got, wantFiveRecent)
	}
	wantCapped := []string{"message-1", "message-2", "message-3"}
	for index := 6; index <= 15; index++ {
		wantCapped = append(wantCapped, "message-"+strconv.Itoa(index))
	}
	wantCapped = append(wantCapped, "current")
	if got := buildTitleInputs(path, "current", 99); !equalStrings(got, wantCapped) {
		t.Fatalf("capped inputs = %#v, want %#v", got, wantCapped)
	}
}

func TestTranscriptReadsWholeFilePastHugeLastLine(t *testing.T) {
	t.Parallel()
	huge := strings.Repeat("x", (1<<20)+1024)
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"first user"}}`,
		`{"type":"ai-title","aiTitle":"以前のタイトル"}`,
		`{"type":"user","message":{"content":"recent user"}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","content":"`+huge+`"}]}}`,
	)
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, []string{"first user", "recent user", "current"}) {
		t.Fatalf("fallback inputs = %#v", got)
	}
	if got := extractConversationTitle(path); got != "以前のタイトル" {
		t.Fatalf("fallback title = %q", got)
	}
}

func TestTranscriptOverLimitReadsTailAndDropsPartialFirstLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	data := `{"type":"user","message":{"content":"too old"}}` + "\n" +
		strings.Repeat("x", fullReadLimit+1024) + "\n" +
		`{"type":"user","message":{"content":"recent"}}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, []string{"recent", "current"}) {
		t.Fatalf("over-limit inputs = %#v", got)
	}
}

func TestTranscriptOverLimitPreservesCompleteFirstTailLine(t *testing.T) {
	boundaryLine := `{"type":"user","message":{"content":"boundary"}}`
	recentLine := `{"type":"user","message":{"content":"recent"}}`
	fillerLength := fullReadLimit - len(boundaryLine) - 1 - 1 - len(recentLine)
	if fillerLength <= 0 {
		t.Fatal("invalid fixture sizes")
	}
	tail := boundaryLine + "\n" + strings.Repeat("x", fillerLength) + "\n" + recentLine
	if len(tail) != fullReadLimit {
		t.Fatalf("tail size = %d, want %d", len(tail), fullReadLimit)
	}
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	data := "outside tail\n" + tail
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := buildTitleInputs(path, "current", 0); !equalStrings(got, []string{"boundary", "recent", "current"}) {
		t.Fatalf("boundary-aligned inputs = %#v", got)
	}
}

func TestTranscriptUnderTailLimitPreservesFirstLineAndTruncatesInputs(t *testing.T) {
	t.Parallel()
	longHistory := strings.Repeat("履", maxInputRunes+20)
	longCurrent := strings.Repeat("現", maxInputRunes+20)
	path := writeTranscript(t, `{"type":"user","message":{"content":"`+longHistory+`"}}`)
	got := buildTitleInputs(path, longCurrent, 0)
	if len(got) != 2 || len([]rune(got[0])) != maxInputRunes || len([]rune(got[1])) != maxInputRunes {
		t.Fatalf("truncated inputs lengths = %d, %d", len([]rune(got[0])), len([]rune(got[1])))
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestGenerateTitleUsesIsolatedRequestAndCleansOutput(t *testing.T) {
	t.Parallel()
	var captured cursorRequest
	fake := &fakeCommandRunner{
		generate: func(ctx context.Context, request cursorRequest) error {
			captured = request
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > cursorCommandTimeout {
				return errors.New("cursor context does not have a 60-second deadline")
			}
			info, err := request.Output.Stat()
			if err != nil {
				return err
			}
			if info.Mode().Perm() != 0o600 {
				return errors.New("output mode is not 0600")
			}
			_, err = request.Output.WriteString("生成タイトル\n")
			return err
		},
	}
	app := testApplication(t, fake)
	app.environ = func() []string {
		return []string{
			"PATH=/bin", "HERDR_TAB_ID=tab", "HERDR_OTHER=x",
			"CLAUDE_CODE_CHILD_SESSION=1", "CLAUDE_CODE_X=y", "KEEP=value",
		}
	}
	path := writeTranscript(t, `{"type":"user","message":{"content":"past"}}`)
	if got := app.generateTitle(workerPayload{Prompt: "current", TranscriptPath: path}); got != "生成タイトル" {
		t.Fatalf("generated title = %q", got)
	}
	if captured.Cwd != filepath.Join(app.stateDir(), "ws") {
		t.Errorf("cursor cwd = %q", captured.Cwd)
	}
	if !equalStrings(captured.Env, []string{"PATH=/bin", "KEEP=value"}) {
		t.Errorf("cursor env = %#v", captured.Env)
	}
	if strings.Count(captured.Prompt, "current") != 1 || !strings.Contains(captured.Prompt, "past") {
		t.Errorf("cursor prompt lacks deduplicated input: %q", captured.Prompt)
	}
	assertNoTemporaryFiles(t, app)
}

func TestCursorArguments(t *testing.T) {
	t.Parallel()
	want := []string{
		"-p", "--trust", "--mode", "ask", "--model", "composer-2.5",
		"--output-format", "text", "prompt",
	}
	if got := cursorArguments("prompt"); !equalStrings(got, want) {
		t.Fatalf("cursorArguments() = %#v, want %#v", got, want)
	}
}

func TestTitlePromptChangesWhenPreviousTitleExists(t *testing.T) {
	t.Parallel()
	withoutPrevious := titlePrompt([]string{"first", "recent"}, "")
	if !strings.Contains(withoutPrevious, "前半がセッション開始時、後半が直近") ||
		!strings.Contains(withoutPrevious, "セッション全体で取り組んでいる作業") ||
		strings.Contains(withoutPrevious, "現在のタイトル") {
		t.Fatalf("prompt without previous title = %q", withoutPrevious)
	}

	withPrevious := titlePrompt([]string{"first", "recent"}, "前回タイトル")
	if !strings.Contains(withPrevious, "現在のタイトルは「前回タイトル」") ||
		!strings.Contains(withPrevious, "今もセッションの主題を表しているなら") {
		t.Fatalf("prompt with previous title = %q", withPrevious)
	}
	for _, rule := range []string{"日本語", "12文字以内", "名詞句", "句読点", "引用符", "記号", "ラベルだけ"} {
		if !strings.Contains(withPrevious, rule) {
			t.Errorf("output rule %q is missing from prompt", rule)
		}
	}
}

func TestGenerateTitleFailureAndTimeoutCleanOutput(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "failure", err: errors.New("cursor failed")},
		{name: "timeout", err: context.DeadlineExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeCommandRunner{generate: func(_ context.Context, request cursorRequest) error {
				_, _ = request.Output.WriteString("untrusted result")
				return test.err
			}}
			app := testApplication(t, fake)
			if got := app.generateTitle(workerPayload{Prompt: "current"}); got != "" {
				t.Fatalf("generateTitle() = %q", got)
			}
			assertNoTemporaryFiles(t, app)
		})
	}
}

func TestHookStartFailureRemoves0600Payload(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	var payloadPath string
	app.startDetached = func(_ string, path string) error {
		payloadPath = path
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode().Perm() != 0o600 {
			return errors.New("payload mode is not 0600")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), "sensitive prompt") {
			return errors.New("payload does not contain prompt")
		}
		return errors.New("start failed")
	}
	err := app.runHook(strings.NewReader(`{"prompt":"sensitive prompt","transcript_path":"/tmp/x"}`), "tab", "workspace")
	if err == nil {
		t.Fatal("runHook() succeeded")
	}
	if _, err := os.Stat(payloadPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("payload remains after Start failure: %v", err)
	}
}

func TestHookStartSuccessLeavesPayloadForWorker(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	var payloadPath string
	app.startDetached = func(executable, path string) error {
		if executable != "/test/herdr-title" {
			return errors.New("unexpected executable")
		}
		payloadPath = path
		return nil
	}
	if err := app.runHook(strings.NewReader(`{"prompt":"current"}`), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(payloadPath); err != nil {
		t.Fatalf("payload was removed after successful Start: %v", err)
	}
	if err := os.Remove(payloadPath); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationIntervalBoundaries(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_700_000_000, 0).UnixNano()
	interval := generationInterval.Nanoseconds()
	cases := []struct {
		name          string
		lastStartedAt int64
		wantWithin    bool
	}{
		{name: "unset", lastStartedAt: 0},
		{name: "five minutes minus one nanosecond", lastStartedAt: now - interval + 1, wantWithin: true},
		{name: "exactly five minutes", lastStartedAt: now - interval},
		{name: "five minutes plus one nanosecond", lastStartedAt: now - interval - 1},
		{name: "negative elapsed", lastStartedAt: now + 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := withinGenerationInterval(test.lastStartedAt, now); got != test.wantWithin {
				t.Fatalf("withinGenerationInterval(%d, %d) = %v, want %v", test.lastStartedAt, now, got, test.wantWithin)
			}
		})
	}
}

func TestHookWithinFiveMinutesOnlyIncrementsSkippedCount(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	now := app.now().UnixNano()
	wantBefore := tabState{
		Generation: now - 100, Title: "前回タイトル",
		LastStartedAt: now - time.Minute.Nanoseconds(), SkippedCount: 2,
	}
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := app.writeState("workspace", map[string]tabState{"tab": wantBefore}); err != nil {
		t.Fatal(err)
	}
	if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	wantAfter := wantBefore
	wantAfter.SkippedCount++
	if got := state["tab"]; got != wantAfter {
		t.Fatalf("gated state = %#v, want %#v", got, wantAfter)
	}
	assertNoTemporaryFiles(t, app)
}

func TestEligibleHookStartsWorkerAndResetsSkippedCount(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	now := app.now().UnixNano()
	before := tabState{
		Generation: now - 100, Title: "前回タイトル",
		LastStartedAt: now - generationInterval.Nanoseconds(), SkippedCount: 4,
	}
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := app.writeState("workspace", map[string]tabState{"tab": before}); err != nil {
		t.Fatal(err)
	}
	var captured workerPayload
	app.startDetached = func(_ string, path string) error {
		captured = readPayload(t, path)
		return os.Remove(path)
	}
	if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	got := state["tab"]
	if got.Generation != now || got.Title != before.Title || got.LastStartedAt != now || got.SkippedCount != 0 {
		t.Fatalf("started state = %#v", got)
	}
	if captured.Generation != now || captured.PreviousTitle != before.Title || captured.SkippedCount != 4 {
		t.Fatalf("worker payload = %#v", captured)
	}
}

func TestHookStartsForUnsetExpiredAndFutureStartTimes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		setup func(*testing.T, *application, int64)
	}{
		{name: "unset"},
		{
			name: "more than five minutes old",
			setup: func(t *testing.T, app *application, now int64) {
				if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
					t.Fatal(err)
				}
				entry := tabState{Generation: 10, LastStartedAt: now - generationInterval.Nanoseconds() - 1}
				if err := app.writeState("workspace", map[string]tabState{"tab": entry}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "future value is discarded and gate opens",
			setup: func(t *testing.T, app *application, now int64) {
				writeRawState(t, app, "workspace", `{"tab":{"generation":10,"title":"corrupt","last_started_at":`+jsonNumber(now+1)+`}}`)
			},
		},
		{
			name: "legacy entry has no start time",
			setup: func(t *testing.T, app *application, _ int64) {
				writeRawState(t, app, "workspace", `{"tab":10}`)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := testApplication(t, &fakeCommandRunner{})
			now := app.now().UnixNano()
			if test.setup != nil {
				test.setup(t, app, now)
			}
			starts := 0
			app.startDetached = func(_ string, path string) error {
				starts++
				return os.Remove(path)
			}
			if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "tab", "workspace"); err != nil {
				t.Fatal(err)
			}
			if starts != 1 {
				t.Fatalf("worker starts = %d, want 1", starts)
			}
			state, err := app.readState("workspace")
			if err != nil {
				t.Fatal(err)
			}
			if state["tab"].LastStartedAt != now || state["tab"].SkippedCount != 0 {
				t.Fatalf("eligible state = %#v", state["tab"])
			}
		})
	}
}

func TestHookFailuresDoNotUpdateGenerationOrStartTime(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		setup func(*application)
	}{
		{
			name: "payload creation failure",
			setup: func(app *application) {
				app.payloadWriter = func(workerPayload) (string, error) {
					return "", errors.New("payload failed")
				}
			},
		},
		{
			name: "worker start failure",
			setup: func(app *application) {
				app.startDetached = func(string, string) error { return errors.New("start failed") }
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := testApplication(t, &fakeCommandRunner{})
			now := app.now().UnixNano()
			before := tabState{
				Generation: now - 100, Title: "前回タイトル",
				LastStartedAt: now - generationInterval.Nanoseconds(), SkippedCount: 3,
			}
			if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := app.writeState("workspace", map[string]tabState{"tab": before}); err != nil {
				t.Fatal(err)
			}
			test.setup(app)
			if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "tab", "workspace"); err == nil {
				t.Fatal("runHook() succeeded")
			}
			state, err := app.readState("workspace")
			if err != nil {
				t.Fatal(err)
			}
			if got := state["tab"]; got != before {
				t.Fatalf("state changed after failure: %#v, want %#v", got, before)
			}
			assertNoTemporaryFiles(t, app)
		})
	}
}

func TestTwoConcurrentEligibleHooksStartOnlyOneWorker(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	var mu sync.Mutex
	starts := 0
	app.startDetached = func(_ string, path string) error {
		mu.Lock()
		starts++
		mu.Unlock()
		return os.Remove(path)
	}
	start := make(chan struct{})
	errorsCh := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			errorsCh <- app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "tab", "workspace")
		}()
	}
	close(start)
	for range 2 {
		if err := <-errorsCh; err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	gotStarts := starts
	mu.Unlock()
	if gotStarts != 1 {
		t.Fatalf("worker starts = %d, want 1", gotStarts)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	entry := state["tab"]
	if entry.Generation != app.now().UnixNano() || entry.LastStartedAt != app.now().UnixNano() || entry.SkippedCount != 1 {
		t.Fatalf("concurrent hook state = %#v", entry)
	}
}

func TestPreviousTitleSurvivesMultipleEligibleGenerationsBeforeApply(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	now := app.now()
	currentNow := now
	app.now = func() time.Time { return currentNow }
	initial := tabState{
		Generation: now.UnixNano() - 10, Title: "安定タイトル",
		LastStartedAt: now.Add(-generationInterval).UnixNano(), SkippedCount: 2,
	}
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := app.writeState("workspace", map[string]tabState{"tab": initial}); err != nil {
		t.Fatal(err)
	}
	var payloads []workerPayload
	app.startDetached = func(_ string, path string) error {
		payloads = append(payloads, readPayload(t, path))
		return os.Remove(path)
	}
	if err := app.runHook(strings.NewReader(hookJSON("first", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	currentNow = currentNow.Add(generationInterval)
	if err := app.runHook(strings.NewReader(hookJSON("second", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	if len(payloads) != 2 || payloads[0].PreviousTitle != initial.Title || payloads[1].PreviousTitle != initial.Title {
		t.Fatalf("previous titles in payloads = %#v", payloads)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["tab"].Title != initial.Title {
		t.Fatalf("title was cleared before apply: %#v", state["tab"])
	}
}

func TestSkippedHooksExpandInputsForNextWorker(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	currentNow := app.now()
	app.now = func() time.Time { return currentNow }
	initial := tabState{Generation: 10, LastStartedAt: currentNow.UnixNano()}
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := app.writeState("workspace", map[string]tabState{"tab": initial}); err != nil {
		t.Fatal(err)
	}
	for _, prompt := range []string{"skipped one", "skipped two"} {
		if err := app.runHook(strings.NewReader(hookJSON(prompt, "/transcript")), "tab", "workspace"); err != nil {
			t.Fatal(err)
		}
	}
	currentNow = currentNow.Add(generationInterval)
	var captured workerPayload
	app.startDetached = func(_ string, path string) error {
		captured = readPayload(t, path)
		return os.Remove(path)
	}
	if err := app.runHook(strings.NewReader(hookJSON("eligible", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	if captured.SkippedCount != 2 {
		t.Fatalf("payload skipped count = %d, want 2", captured.SkippedCount)
	}
	if recent := recentHistoryEntries(captured.SkippedCount); recent != 5 {
		t.Fatalf("recent input count = %d, want 5", recent)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["tab"].SkippedCount != 0 {
		t.Fatalf("skipped count was not reset: %#v", state["tab"])
	}
}

func TestWorkerFallbackRemovesPayloadBeforeCommands(t *testing.T) {
	t.Parallel()
	var app *application
	var calls [][]string
	fake := &fakeCommandRunner{
		generate: func(_ context.Context, _ cursorRequest) error {
			entries, err := os.ReadDir(app.stateDir())
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), "payload-") {
					return errors.New("payload still exists during cursor call")
				}
			}
			return errors.New("cursor failed")
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			calls = append(calls, append([]string(nil), args...))
			if equalStrings(args, []string{"tab", "list"}) {
				return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
			}
			return nil, nil
		},
	}
	app = testApplication(t, fake)
	transcript := writeTranscript(t,
		`{"type":"user","message":{"content":"past"}}`,
		`{"type":"ai-title","aiTitle":"Fallback「題」"}`,
	)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	payloadPath, err := app.writePayload(workerPayload{
		Prompt: "current", TranscriptPath: transcript, TabID: "tab",
		WorkspaceID: "workspace", Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.runWorker(payloadPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(payloadPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("payload remains after worker: %v", err)
	}
	if !containsCall(calls, "tab", "rename", "tab", "Fallback題") ||
		!containsCall(calls, "workspace", "rename", "workspace", "Fallback題") {
		t.Fatalf("fallback rename calls = %#v", calls)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["tab"].Title != "Fallback題" {
		t.Fatalf("fallback title was not saved: %#v", state["tab"])
	}
	assertNoTemporaryFiles(t, app)
}

func TestWorkerUsesFallbackWhenThereAreNoInputs(t *testing.T) {
	t.Parallel()
	generated := false
	var calls [][]string
	fake := &fakeCommandRunner{
		generate: func(context.Context, cursorRequest) error {
			generated = true
			return nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			calls = append(calls, append([]string(nil), args...))
			if equalStrings(args, []string{"tab", "list"}) {
				return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
			}
			return nil, nil
		},
	}
	app := testApplication(t, fake)
	transcript := writeTranscript(t, `{"type":"custom-title","customTitle":"既存タイトル"}`)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	payloadPath, err := app.writePayload(workerPayload{
		TranscriptPath: transcript, TabID: "tab", WorkspaceID: "workspace", Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.runWorker(payloadPath); err != nil {
		t.Fatal(err)
	}
	if generated {
		t.Fatal("cursor-agent was called with zero inputs")
	}
	if !containsCall(calls, "tab", "rename", "tab", "既存タイトル") {
		t.Fatalf("fallback was not applied: %#v", calls)
	}
}

func TestWorkerUsesFallbackForEmptyCursorOutput(t *testing.T) {
	t.Parallel()
	var calls [][]string
	fake := &fakeCommandRunner{
		generate: func(context.Context, cursorRequest) error { return nil },
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			calls = append(calls, append([]string(nil), args...))
			if equalStrings(args, []string{"tab", "list"}) {
				return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
			}
			return nil, nil
		},
	}
	app := testApplication(t, fake)
	transcript := writeTranscript(t, `{"type":"ai-title","aiTitle":"空出力fallback"}`)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	payloadPath, err := app.writePayload(workerPayload{
		Prompt: "current", TranscriptPath: transcript, TabID: "tab",
		WorkspaceID: "workspace", Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.runWorker(payloadPath); err != nil {
		t.Fatal(err)
	}
	if !containsCall(calls, "tab", "rename", "tab", "空出力fallback") {
		t.Fatalf("empty-output fallback calls = %#v", calls)
	}
}

func TestWorkerWithoutGeneratedOrFallbackTitleKeepsExistingLabels(t *testing.T) {
	t.Parallel()
	called := false
	fake := &fakeCommandRunner{
		generate: func(context.Context, cursorRequest) error { return errors.New("failed") },
		herdr: func(context.Context, ...string) ([]byte, error) {
			called = true
			return nil, nil
		},
	}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	payloadPath, err := app.writePayload(workerPayload{
		Prompt: "current", TabID: "tab", WorkspaceID: "workspace", Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.runWorker(payloadPath); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("herdr was called without generated or fallback title")
	}
}

func TestPreviousTitleControlsCursorFailureFallback(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		generate   func(context.Context, cursorRequest) error
		wantTitle  string
		wantRename bool
	}{
		{
			name: "success replaces previous title",
			generate: func(_ context.Context, request cursorRequest) error {
				if !strings.Contains(request.Prompt, "現在のタイトルは「安定タイトル」") {
					return errors.New("previous title is missing from prompt")
				}
				_, err := request.Output.WriteString("新しい主題")
				return err
			},
			wantTitle: "新しい主題", wantRename: true,
		},
		{
			name:      "timeout keeps previous title",
			generate:  func(context.Context, cursorRequest) error { return context.DeadlineExceeded },
			wantTitle: "安定タイトル",
		},
		{
			name:      "authentication failure keeps previous title",
			generate:  func(context.Context, cursorRequest) error { return errors.New("authentication failed") },
			wantTitle: "安定タイトル",
		},
		{
			name:      "empty output keeps previous title",
			generate:  func(context.Context, cursorRequest) error { return nil },
			wantTitle: "安定タイトル",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var calls [][]string
			fake := &fakeCommandRunner{
				generate: test.generate,
				herdr: func(_ context.Context, args ...string) ([]byte, error) {
					calls = append(calls, append([]string(nil), args...))
					if equalStrings(args, []string{"tab", "list"}) {
						return []byte(`{"id":"cli:tab:list","result":{"tabs":[{"tab_id":"tab","workspace_id":"workspace"}],"type":"tab_list"}}`), nil
					}
					return nil, nil
				},
			}
			app := testApplication(t, fake)
			entry := tabState{Generation: 42, Title: "安定タイトル", LastStartedAt: 41, SkippedCount: 3}
			if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := app.writeState("workspace", map[string]tabState{"tab": entry}); err != nil {
				t.Fatal(err)
			}
			transcript := writeTranscript(t,
				`{"type":"user","message":{"content":"session task"}}`,
				`{"type":"ai-title","aiTitle":"fallback title"}`,
			)
			payloadPath, err := app.writePayload(workerPayload{
				Prompt: "current", TranscriptPath: transcript, TabID: "tab", WorkspaceID: "workspace",
				Generation: 42, PreviousTitle: "安定タイトル", SkippedCount: 3,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := app.runWorker(payloadPath); err != nil {
				t.Fatal(err)
			}
			state, err := app.readState("workspace")
			if err != nil {
				t.Fatal(err)
			}
			got := state["tab"]
			if got.Title != test.wantTitle || got.Generation != entry.Generation ||
				got.LastStartedAt != entry.LastStartedAt || got.SkippedCount != entry.SkippedCount {
				t.Fatalf("state after cursor result = %#v", got)
			}
			gotRename := containsCall(calls, "tab", "rename", "tab", "新しい主題")
			if gotRename != test.wantRename {
				t.Fatalf("rename = %v, want %v; calls = %#v", gotRename, test.wantRename, calls)
			}
			if !test.wantRename && len(calls) != 0 {
				t.Fatalf("cursor failure changed herdr labels: %#v", calls)
			}
		})
	}
}

func TestCursorFailureAfterWorkerStartKeepsThrottleState(t *testing.T) {
	t.Parallel()
	fake := &fakeCommandRunner{
		generate: func(context.Context, cursorRequest) error { return errors.New("cursor failed") },
	}
	app := testApplication(t, fake)
	var payloadPath string
	app.startDetached = func(_ string, path string) error {
		payloadPath = path
		return nil
	}
	if err := app.runHook(strings.NewReader(hookJSON("current", "")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	before, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.runWorker(payloadPath); err != nil {
		t.Fatal(err)
	}
	after, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if after["tab"] != before["tab"] || after["tab"].LastStartedAt == 0 {
		t.Fatalf("cursor failure rolled back throttle state: before=%#v after=%#v", before["tab"], after["tab"])
	}
}

func assertNoTemporaryFiles(t *testing.T, app *application) {
	t.Helper()
	entries, err := os.ReadDir(app.stateDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "payload-") || strings.HasPrefix(entry.Name(), "out-") {
			t.Errorf("temporary file remains: %s", entry.Name())
		}
	}
}

func containsCall(calls [][]string, want ...string) bool {
	for _, call := range calls {
		if equalStrings(call, want) {
			return true
		}
	}
	return false
}

func TestApplyTitleSingleTabAndWorkspaceRetry(t *testing.T) {
	t.Parallel()
	var calls [][]string
	workspaceAttempts := 0
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`{"id":"cli:tab:list","result":{"tabs":[{"agent_status":"working","focused":false,"label":"1","number":1,"pane_count":1,"tab_id":"tab","workspace_id":"workspace"}],"type":"tab_list"}}`), nil
		}
		if len(args) >= 2 && args[0] == "workspace" && args[1] == "rename" {
			workspaceAttempts++
			if workspaceAttempts == 1 {
				return nil, errors.New("temporary failure")
			}
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: generation}, "Title"); err != nil {
		t.Fatal(err)
	}
	if !containsCall(calls, "tab", "rename", "tab", "Title") || workspaceAttempts != 2 {
		t.Fatalf("calls = %#v, workspace attempts = %d", calls, workspaceAttempts)
	}
}

func TestApplyTitlePreservesThrottleFields(t *testing.T) {
	t.Parallel()
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`{"id":"cli:tab:list","result":{"tabs":[{"tab_id":"tab","workspace_id":"workspace"}],"type":"tab_list"}}`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	before := tabState{Generation: 42, Title: "Old", LastStartedAt: 41, SkippedCount: 3}
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := app.writeState("workspace", map[string]tabState{"tab": before}); err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: 42}, "New"); err != nil {
		t.Fatal(err)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	want := before
	want.Title = "New"
	if got := state["tab"]; got != want {
		t.Fatalf("applied state = %#v, want %#v", got, want)
	}
}

func TestTabListFailureSkipsWorkspaceRename(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		response string
	}{
		{name: "invalid JSON", response: "not json"},
		{name: "missing result tabs", response: `{"id":"cli:tab:list","result":{"type":"tab_list"}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls [][]string
			fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
				calls = append(calls, append([]string(nil), args...))
				if equalStrings(args, []string{"tab", "list"}) {
					return []byte(test.response), nil
				}
				return nil, nil
			}}
			app := testApplication(t, fake)
			generation, err := seedGenerationForTest(app, "workspace", "tab")
			if err != nil {
				t.Fatal(err)
			}
			if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: generation}, "Title"); err != nil {
				t.Fatal(err)
			}
			for _, call := range calls {
				if len(call) >= 2 && call[0] == "workspace" && call[1] == "rename" {
					t.Fatalf("workspace was renamed for unusable tab list: %#v", calls)
				}
			}
		})
	}
}

func TestApplyTitleWorkspaceOnlyStillChecksTabCount(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name          string
		tabs          string
		wantWorkspace bool
	}{
		{name: "single tab", tabs: `[{"tab_id":"any","workspace_id":"workspace"}]`, wantWorkspace: true},
		{name: "multiple tabs", tabs: `[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls [][]string
			fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
				calls = append(calls, append([]string(nil), args...))
				if equalStrings(args, []string{"tab", "list"}) {
					return []byte(test.tabs), nil
				}
				return nil, nil
			}}
			app := testApplication(t, fake)
			generation, err := seedGenerationForTest(app, "workspace", "")
			if err != nil {
				t.Fatal(err)
			}
			if err := app.applyTitle(workerPayload{WorkspaceID: "workspace", Generation: generation}, "Title"); err != nil {
				t.Fatal(err)
			}
			if containsCall(calls, "tab", "rename", "", "Title") {
				t.Fatal("tab rename was called without a tab ID")
			}
			gotWorkspace := containsCall(calls, "workspace", "rename", "workspace", "Title")
			if gotWorkspace != test.wantWorkspace {
				t.Fatalf("workspace rename = %v, want %v; calls = %#v", gotWorkspace, test.wantWorkspace, calls)
			}
		})
	}
}

func TestApplyTitleTabOnlyDoesNotTryWorkspaceCommands(t *testing.T) {
	t.Parallel()
	var calls [][]string
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		return nil, nil
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", Generation: generation}, "Title"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || !containsCall(calls, "tab", "rename", "tab", "Title") {
		t.Fatalf("tab-only calls = %#v", calls)
	}
}

func TestApplyTitleRejectsStaleGeneration(t *testing.T) {
	t.Parallel()
	called := false
	app := testApplication(t, &fakeCommandRunner{herdr: func(context.Context, ...string) ([]byte, error) {
		called = true
		return nil, nil
	}})
	oldGeneration, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seedGenerationForTest(app, "workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	before, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: oldGeneration}, "Old"); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("herdr was called for a stale generation")
	}
	after, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if after["tab"] != before["tab"] {
		t.Fatalf("stale apply changed state: before=%#v after=%#v", before["tab"], after["tab"])
	}
}

func TestTabRenameFailureStopsBeforeWorkspaceRename(t *testing.T) {
	t.Parallel()
	var calls [][]string
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		return nil, errors.New("tab rename failed")
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: generation}, "Title"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || !containsCall(calls, "tab", "rename", "tab", "Title") {
		t.Fatalf("calls after tab rename failure = %#v", calls)
	}
}

func TestApplyTitleSerializesMultipleTabsAndSkipsWorkspace(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	active := 0
	maxActive := 0
	var calls [][]string
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		mu.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		calls = append(calls, append([]string(nil), args...))
		mu.Unlock()
		time.Sleep(15 * time.Millisecond)
		mu.Lock()
		active--
		mu.Unlock()
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	genA, err := seedGenerationForTest(app, "workspace", "a")
	if err != nil {
		t.Fatal(err)
	}
	genB, err := seedGenerationForTest(app, "workspace", "b")
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	errorsCh := make(chan error, 2)
	for _, payload := range []workerPayload{
		{TabID: "a", WorkspaceID: "workspace", Generation: genA},
		{TabID: "b", WorkspaceID: "workspace", Generation: genB},
	} {
		payload := payload
		go func() {
			<-start
			errorsCh <- app.applyTitle(payload, strings.ToUpper(payload.TabID))
		}()
	}
	close(start)
	for range 2 {
		if err := <-errorsCh; err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if maxActive != 1 {
		t.Errorf("maximum concurrent herdr calls = %d, want 1", maxActive)
	}
	if !containsCall(calls, "tab", "rename", "a", "A") || !containsCall(calls, "tab", "rename", "b", "B") {
		t.Errorf("tab rename calls = %#v", calls)
	}
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "workspace" && call[1] == "rename" {
			t.Errorf("workspace was renamed in a multiple-tab workspace: %#v", call)
		}
	}
}

func TestClosedTabWorkerDoesNotRenameRemainingTabWorkspace(t *testing.T) {
	t.Parallel()
	var calls [][]string
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`[{"tab_id":"b","workspace_id":"workspace"}]`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "a")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "a", WorkspaceID: "workspace", Generation: generation}, "Old A"); err != nil {
		t.Fatal(err)
	}
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "workspace" && call[1] == "rename" {
			t.Fatalf("closed tab renamed workspace: %#v", calls)
		}
	}
}

func TestTOCTOUOldWorkerCannotOverwriteNewGeneration(t *testing.T) {
	t.Parallel()
	oldRenameStarted := make(chan struct{})
	releaseOldRename := make(chan struct{})
	var once sync.Once
	var mu sync.Mutex
	finalLabel := ""
	fake := &fakeCommandRunner{herdr: func(_ context.Context, args ...string) ([]byte, error) {
		if len(args) >= 4 && args[0] == "tab" && args[1] == "rename" {
			if args[3] == "Old" {
				once.Do(func() { close(oldRenameStarted) })
				<-releaseOldRename
			}
			mu.Lock()
			finalLabel = args[3]
			mu.Unlock()
			return nil, nil
		}
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	currentNow := app.now()
	app.now = func() time.Time { return currentNow }
	var payloadsMu sync.Mutex
	var payloads []workerPayload
	app.startDetached = func(_ string, path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var payload workerPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		payloadsMu.Lock()
		payloads = append(payloads, payload)
		payloadsMu.Unlock()
		return nil
	}
	if err := app.runHook(strings.NewReader(hookJSON("old", "/transcript")), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	payloadsMu.Lock()
	oldPayload := payloads[0]
	payloadsMu.Unlock()
	oldDone := make(chan error, 1)
	go func() {
		oldDone <- app.applyTitle(oldPayload, "Old")
	}()
	<-oldRenameStarted

	currentNow = currentNow.Add(generationInterval)
	newHookDone := make(chan error, 1)
	go func() {
		newHookDone <- app.runHook(strings.NewReader(hookJSON("new", "/transcript")), "tab", "workspace")
	}()
	select {
	case err := <-newHookDone:
		t.Fatalf("new hook completed while old worker held lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseOldRename)
	if err := <-oldDone; err != nil {
		t.Fatal(err)
	}
	if err := <-newHookDone; err != nil {
		t.Fatal(err)
	}
	payloadsMu.Lock()
	if len(payloads) != 2 {
		payloadsMu.Unlock()
		t.Fatalf("payload count = %d, want 2", len(payloads))
	}
	newPayload := payloads[1]
	payloadsMu.Unlock()
	if newPayload.Generation <= oldPayload.Generation {
		t.Fatalf("new generation = %d, old = %d", newPayload.Generation, oldPayload.Generation)
	}
	if err := app.applyTitle(newPayload, "New"); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if finalLabel != "New" {
		t.Fatalf("final tab label = %q, want New", finalLabel)
	}
}

func TestEligibleHookLeavesStateUnchangedWhenAnotherTabHoldsLock(t *testing.T) {
	started := make(chan struct{})
	var once sync.Once
	var callsMu sync.Mutex
	var calls [][]string
	fake := &fakeCommandRunner{herdr: func(ctx context.Context, args ...string) ([]byte, error) {
		callsMu.Lock()
		calls = append(calls, append([]string(nil), args...))
		callsMu.Unlock()
		if len(args) >= 2 && args[0] == "tab" && args[1] == "rename" {
			once.Do(func() { close(started) })
			<-ctx.Done()
			return nil, ctx.Err()
		}
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "a")
	if err != nil {
		t.Fatal(err)
	}
	state, err := app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	wantB := tabState{Generation: 10, Title: "B title"}
	state["b"] = wantB
	if err := app.writeState("workspace", state); err != nil {
		t.Fatal(err)
	}
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- app.applyTitle(workerPayload{TabID: "a", WorkspaceID: "workspace", Generation: generation}, "Title")
	}()
	<-started

	begin := time.Now()
	if err := app.runHook(strings.NewReader(hookJSON("current", "/transcript")), "b", "workspace"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runHook() error = %v, want deadline", err)
	}
	elapsed := time.Since(begin)
	if elapsed < 900*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("lock timeout elapsed = %s, want about 1 second", elapsed)
	}
	select {
	case err := <-workerDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("herdr timeout did not release workspace lock")
	}
	state, err = app.readState("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["b"] != wantB {
		t.Fatalf("eligible hook changed state while lock was held: %#v", state["b"])
	}
	if _, err := seedGenerationForTest(app, "workspace", "b"); err != nil {
		t.Fatalf("lock did not recover: %v", err)
	}
	callsMu.Lock()
	defer callsMu.Unlock()
	if len(calls) != 1 || !containsCall(calls, "tab", "rename", "a", "Title") {
		t.Fatalf("tab timeout continued to workspace operations: %#v", calls)
	}
}

func TestEveryHerdrCallGetsDeadline(t *testing.T) {
	t.Parallel()
	var deadlines []time.Duration
	fake := &fakeCommandRunner{herdr: func(ctx context.Context, args ...string) ([]byte, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			return nil, errors.New("missing deadline")
		}
		deadlines = append(deadlines, time.Until(deadline))
		if equalStrings(args, []string{"tab", "list"}) {
			return []byte(`[{"tab_id":"tab","workspace_id":"workspace"}]`), nil
		}
		return nil, nil
	}}
	app := testApplication(t, fake)
	generation, err := seedGenerationForTest(app, "workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: generation}, "Title"); err != nil {
		t.Fatal(err)
	}
	if len(deadlines) != 3 {
		t.Fatalf("deadline count = %d, want 3", len(deadlines))
	}
	for _, remaining := range deadlines {
		if remaining <= 0 || remaining > herdrCommandTimeout {
			t.Errorf("invalid herdr timeout: %s", remaining)
		}
	}
}

func TestPayloadAndOutputFromCrashedWorkerAreCleanedAfterOneHour(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	if err := os.MkdirAll(app.stateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	old := app.now().Add(-staleFileAge - time.Second)
	for _, name := range []string{"payload-crashed", "out-crashed"} {
		path := filepath.Join(app.stateDir(), name)
		if err := os.WriteFile(path, []byte("sensitive"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	app.cleanupStaleFiles()
	assertNoTemporaryFiles(t, app)
}

func TestParseTabListFormats(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		data string
	}{
		{name: "actual CLI envelope", data: `{"id":"cli:tab:list","result":{"tabs":[{"tab_id":"tab","workspace_id":"workspace"}],"type":"tab_list"}}`},
		{name: "top-level tabs", data: `{"tabs":[{"tab_id":"tab","workspace_id":"workspace"}]}`},
		{name: "bare array", data: `[{"tab_id":"tab","workspace_id":"workspace"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			tabs, err := parseTabList([]byte(test.data))
			if err != nil {
				t.Fatal(err)
			}
			if len(tabs) != 1 || tabs[0].TabID != "tab" || tabs[0].WorkspaceID != "workspace" {
				t.Fatalf("tabs = %#v", tabs)
			}
		})
	}
	for _, data := range []string{"not json", `{"id":"cli:tab:list","result":{"type":"tab_list"}}`} {
		if _, err := parseTabList([]byte(data)); err == nil {
			t.Errorf("parseTabList() accepted unusable output %q", data)
		}
	}
}

func TestPayloadJSONRoundTrip(t *testing.T) {
	t.Parallel()
	payload := workerPayload{
		Prompt: "prompt", TranscriptPath: "/path", TabID: "tab", WorkspaceID: "workspace",
		Generation: 42, PreviousTitle: "previous", SkippedCount: 3,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded workerPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != payload {
		t.Fatalf("decoded payload = %#v, want %#v", decoded, payload)
	}
}
