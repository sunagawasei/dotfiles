package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
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

	if _, err := app.reserveGeneration("workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.reserveGeneration("workspace", "other-tab"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.reserveGeneration("workspace-only", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := app.reserveGeneration("", "tab-only"); err != nil {
		t.Fatal(err)
	}

	state, err := app.readGenerations("workspace")
	if err != nil {
		t.Fatal(err)
	}
	if state["tab"] == 0 || state["other-tab"] == 0 || len(state) != 2 {
		t.Fatalf("unexpected tab-keyed state: %#v", state)
	}
	workspaceOnly, err := app.readGenerations("workspace-only")
	if err != nil || workspaceOnly[""] == 0 {
		t.Fatalf("workspace-only state = %#v, err = %v", workspaceOnly, err)
	}
	tabOnly, err := app.readGenerations("")
	if err != nil || tabOnly["tab-only"] == 0 {
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

func TestConcurrentGenerationIsUniqueAndMonotonic(t *testing.T) {
	t.Parallel()
	app := testApplication(t, &fakeCommandRunner{})
	start := make(chan struct{})
	results := make(chan int64, 2)
	errorsCh := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			generation, err := app.reserveGeneration("workspace", "tab")
			results <- generation
			errorsCh <- err
		}()
	}
	close(start)
	values := []int64{<-results, <-results}
	for range 2 {
		if err := <-errorsCh; err != nil {
			t.Fatal(err)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	wantFirst := app.now().UnixNano()
	if values[0] != wantFirst || values[1] != wantFirst+1 {
		t.Fatalf("generations = %v, want [%d %d]", values, wantFirst, wantFirst+1)
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
	if _, err := app.reserveGeneration("workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	lockInode := inodeOf(t, app.lockPath("workspace"))
	generationInode := inodeOf(t, app.generationPath("workspace"))
	if _, err := app.reserveGeneration("workspace", "tab"); err != nil {
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
		`{"type":"user","message":{"content":"current"}}`,
		`{"type":"user","message":{"content":"incomplete"}} trailing`,
	)
	want := []string{"first", "mixed", "visible", "current"}
	if got := buildTitleInputs(path, "current"); !equalStrings(got, want) {
		t.Fatalf("buildTitleInputs() = %#v, want %#v", got, want)
	}

	withoutCurrent := writeTranscript(t, `{"type":"user","message":{"content":"past"}}`)
	if got := buildTitleInputs(withoutCurrent, "current"); !equalStrings(got, []string{"past", "current"}) {
		t.Fatalf("current prompt missing from transcript: %#v", got)
	}
}

func TestTranscriptExtractionKeepsOnlyLatestFourHistoryEntries(t *testing.T) {
	t.Parallel()
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"one"}}`,
		`{"type":"user","message":{"content":"two"}}`,
		`{"type":"user","message":{"content":"three"}}`,
		`{"type":"user","message":{"content":"four"}}`,
		`{"type":"user","message":{"content":"five"}}`,
	)
	want := []string{"two", "three", "four", "five", "current"}
	if got := buildTitleInputs(path, "current"); !equalStrings(got, want) {
		t.Fatalf("buildTitleInputs() = %#v, want %#v", got, want)
	}
}

func TestTranscriptReadFallbackPastHugeLastLine(t *testing.T) {
	t.Parallel()
	huge := strings.Repeat("x", tailReadLimit+1024)
	path := writeTranscript(t,
		`{"type":"user","message":{"content":"first user"}}`,
		`{"type":"ai-title","aiTitle":"以前のタイトル"}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","content":"`+huge+`"}]}}`,
	)
	if got := buildTitleInputs(path, "current"); !equalStrings(got, []string{"first user", "current"}) {
		t.Fatalf("fallback inputs = %#v", got)
	}
	if got := extractConversationTitle(path); got != "以前のタイトル" {
		t.Fatalf("fallback title = %q", got)
	}
}

func TestTranscriptUnderTailLimitPreservesFirstLineAndTruncatesInputs(t *testing.T) {
	t.Parallel()
	longHistory := strings.Repeat("履", maxInputRunes+20)
	longCurrent := strings.Repeat("現", maxInputRunes+20)
	path := writeTranscript(t, `{"type":"user","message":{"content":"`+longHistory+`"}}`)
	got := buildTitleInputs(path, longCurrent)
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
			generation, err := app.reserveGeneration("workspace", "tab")
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
			generation, err := app.reserveGeneration("workspace", "")
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
	generation, err := app.reserveGeneration("", "tab")
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
	oldGeneration, err := app.reserveGeneration("workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.reserveGeneration("workspace", "tab"); err != nil {
		t.Fatal(err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: oldGeneration}, "Old"); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("herdr was called for a stale generation")
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	genA, err := app.reserveGeneration("workspace", "a")
	if err != nil {
		t.Fatal(err)
	}
	genB, err := app.reserveGeneration("workspace", "b")
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
	generation, err := app.reserveGeneration("workspace", "a")
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
	oldGeneration, err := app.reserveGeneration("workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	oldDone := make(chan error, 1)
	go func() {
		oldDone <- app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: oldGeneration}, "Old")
	}()
	<-oldRenameStarted

	type generationResult struct {
		value int64
		err   error
	}
	newGeneration := make(chan generationResult, 1)
	go func() {
		value, err := app.reserveGeneration("workspace", "tab")
		newGeneration <- generationResult{value: value, err: err}
	}()
	select {
	case result := <-newGeneration:
		t.Fatalf("new generation acquired while old worker held lock: %+v", result)
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseOldRename)
	if err := <-oldDone; err != nil {
		t.Fatal(err)
	}
	result := <-newGeneration
	if result.err != nil {
		t.Fatal(result.err)
	}
	if err := app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: result.value}, "New"); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if finalLabel != "New" {
		t.Fatalf("final tab label = %q, want New", finalLabel)
	}
}

func TestHookSkipsAfterOneSecondWhileHerdrHoldsLockAndRecovers(t *testing.T) {
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
	generation, err := app.reserveGeneration("workspace", "tab")
	if err != nil {
		t.Fatal(err)
	}
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- app.applyTitle(workerPayload{TabID: "tab", WorkspaceID: "workspace", Generation: generation}, "Title")
	}()
	<-started

	begin := time.Now()
	if _, err := app.reserveGeneration("workspace", "tab"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("reserveGeneration() error = %v, want deadline", err)
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
	if _, err := app.reserveGeneration("workspace", "tab"); err != nil {
		t.Fatalf("lock did not recover: %v", err)
	}
	callsMu.Lock()
	defer callsMu.Unlock()
	if len(calls) != 1 || !containsCall(calls, "tab", "rename", "tab", "Title") {
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
	generation, err := app.reserveGeneration("workspace", "tab")
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
	payload := workerPayload{Prompt: "prompt", TranscriptPath: "/path", TabID: "tab", WorkspaceID: "workspace", Generation: 42}
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
