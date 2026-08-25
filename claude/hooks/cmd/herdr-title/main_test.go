package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type fakeCommandRunner struct {
	generate func(context.Context, titleRequest) (string, error)
	herdr    func(context.Context, ...string) ([]byte, error)
}

func (f *fakeCommandRunner) GenerateTitle(ctx context.Context, request titleRequest) (string, error) {
	if f.generate == nil {
		return "", errors.New("unexpected codex call")
	}
	return f.generate(ctx, request)
}

func (f *fakeCommandRunner) RunHerdr(ctx context.Context, args ...string) ([]byte, error) {
	if f.herdr == nil {
		return nil, errors.New("unexpected herdr call")
	}
	return f.herdr(ctx, args...)
}

func testApplication(t *testing.T, runner commandRunner) *application {
	t.Helper()
	config := defaultRuntimeConfig()
	config.tabThrottle = 120 * time.Millisecond
	config.workspaceThrottle = 400 * time.Millisecond
	config.retryDelay = 30 * time.Millisecond
	config.guardRetry = 100 * time.Millisecond
	config.actorEviction = 2 * time.Second
	config.daemonIdle = 2 * time.Second
	config.idlePoll = 10 * time.Millisecond
	config.codexTimeout = time.Second
	config.herdrTimeout = time.Second
	config.ipcDeadline = time.Second
	config.hookACKTimeout = 100 * time.Millisecond
	config.hookLatencyLimit = 250 * time.Millisecond
	config.handoverDelay = 30 * time.Millisecond
	config.spawnSettle = 5 * time.Millisecond
	return &application{
		cacheDir:    t.TempDir(),
		projectsDir: t.TempDir(),
		buildHash:   "test-build",
		executable:  "/test/herdr-title",
		commands:    runner,
		dial:        net.DialTimeout,
		spawn:       func(string, string) error { return nil },
		now:         time.Now,
		sleep:       time.Sleep,
		newTimer: func(duration time.Duration) clockTimer {
			return realClockTimer{timer: time.NewTimer(duration)}
		},
		newTicker: func(duration time.Duration) clockTicker {
			return realClockTicker{ticker: time.NewTicker(duration)}
		},
		stderr: io.Discard,
		config: config,
	}
}

func eventFor(app *application, kinds []actorKind, text, tabID, workspaceID string) daemonEvent {
	return daemonEvent{
		ProtocolVersion: protocolVersion,
		BuildHash:       app.buildHash,
		Kinds:           kinds,
		Input: pendingInput{
			Text: text, TabID: tabID, WorkspaceID: workspaceID,
			HookEventName: "UserPromptSubmit",
		},
	}
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

type manualClock struct {
	mu      sync.Mutex
	now     time.Time
	timers  map[*manualTimer]struct{}
	tickers map[*manualTicker]struct{}
}

type manualTimer struct {
	clock    *manualClock
	deadline time.Time
	channel  chan time.Time
	active   bool
}

func (t *manualTimer) Chan() <-chan time.Time { return t.channel }

func (t *manualTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	wasActive := t.active
	t.active = false
	delete(t.clock.timers, t)
	return wasActive
}

type manualTicker struct {
	clock    *manualClock
	interval time.Duration
	next     time.Time
	channel  chan time.Time
	active   bool
}

func (t *manualTicker) Chan() <-chan time.Time { return t.channel }

func (t *manualTicker) Stop() {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	t.active = false
	delete(t.clock.tickers, t)
}

func newManualClock() *manualClock {
	return &manualClock{
		now:     time.Unix(1_800_000_000, 0),
		timers:  make(map[*manualTimer]struct{}),
		tickers: make(map[*manualTicker]struct{}),
	}
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *manualClock) Sleep(duration time.Duration) { c.Advance(duration) }

func (c *manualClock) NewTimer(duration time.Duration) clockTimer {
	c.mu.Lock()
	defer c.mu.Unlock()
	timer := &manualTimer{
		clock: c, deadline: c.now.Add(duration), channel: make(chan time.Time, 1), active: true,
	}
	c.timers[timer] = struct{}{}
	return timer
}

func (c *manualClock) NewTicker(interval time.Duration) clockTicker {
	c.mu.Lock()
	defer c.mu.Unlock()
	ticker := &manualTicker{
		clock: c, interval: interval, next: c.now.Add(interval), channel: make(chan time.Time, 1), active: true,
	}
	c.tickers[ticker] = struct{}{}
	return ticker
}

func (c *manualClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	now := c.now
	dueTimers := make([]*manualTimer, 0)
	for timer := range c.timers {
		if timer.active && !timer.deadline.After(now) {
			timer.active = false
			delete(c.timers, timer)
			dueTimers = append(dueTimers, timer)
		}
	}
	dueTickers := make([]*manualTicker, 0)
	for ticker := range c.tickers {
		if ticker.active && !ticker.next.After(now) {
			for !ticker.next.After(now) {
				ticker.next = ticker.next.Add(ticker.interval)
			}
			dueTickers = append(dueTickers, ticker)
		}
	}
	c.mu.Unlock()
	for _, timer := range dueTimers {
		timer.channel <- now
	}
	for _, ticker := range dueTickers {
		select {
		case ticker.channel <- now:
		default:
		}
	}
}

func useManualClock(app *application, clock *manualClock) {
	app.now = clock.Now
	app.sleep = clock.Sleep
	app.newTimer = clock.NewTimer
	app.newTicker = clock.NewTicker
}

func TestProductionTimingConstants(t *testing.T) {
	config := defaultRuntimeConfig()
	if config.tabThrottle != 120*time.Second {
		t.Fatalf("tab throttle = %s", config.tabThrottle)
	}
	if config.workspaceThrottle != 30*time.Minute {
		t.Fatalf("workspace throttle = %s", config.workspaceThrottle)
	}
	if config.retryDelay != 30*time.Second || config.guardRetry != 30*time.Second {
		t.Fatalf("retry durations = %s, %s", config.retryDelay, config.guardRetry)
	}
	if config.actorEviction != 10*time.Minute || config.daemonIdle != 60*time.Minute {
		t.Fatalf("idle durations = %s, %s", config.actorEviction, config.daemonIdle)
	}
	if config.ipcDeadline != time.Second || config.handoverDelay != 300*time.Millisecond {
		t.Fatalf("IPC durations = %s, %s", config.ipcDeadline, config.handoverDelay)
	}
}

func TestBuildDaemonEventTriggerMatrixAndSubagentGate(t *testing.T) {
	user := hookInput{HookEventName: "UserPromptSubmit", Prompt: "current", TranscriptPath: "/transcript"}
	event, ok := buildDaemonEvent(user, "tab", "workspace", "hash")
	if !ok || len(event.Kinds) != 2 || event.Kinds[0] != tabKind || event.Kinds[1] != workspaceKind {
		t.Fatalf("user event = %#v, ok = %v", event, ok)
	}

	pre := hookInput{
		HookEventName: "PreToolUse", ToolName: "Bash",
		ToolInput: map[string]any{"description": "deploy with Bearer abc.def.ghi"},
	}
	event, ok = buildDaemonEvent(pre, "tab", "workspace", "hash")
	if !ok || len(event.Kinds) != 1 || event.Kinds[0] != tabKind {
		t.Fatalf("tool event = %#v, ok = %v", event, ok)
	}
	if strings.Contains(event.Input.Text, "abc.def.ghi") || !strings.Contains(event.Input.Text, "[REDACTED]") {
		t.Fatalf("tool summary was not redacted: %q", event.Input.Text)
	}

	agentID := "agent-a123"
	if _, ok := buildDaemonEvent(hookInput{HookEventName: "PostToolUse", AgentID: &agentID}, "tab", "workspace", "hash"); ok {
		t.Fatal("subagent event passed the hook-side gate")
	}
	if _, ok := buildDaemonEvent(hookInput{HookEventName: "Stop"}, "tab", "workspace", "hash"); ok {
		t.Fatal("unsupported hook event was accepted")
	}
}

func TestBuildDaemonEventTruncatesPromptBeforeIPCFrame(t *testing.T) {
	prompt := strings.Repeat("p", maxFrameSize+1024)
	input := hookInput{HookEventName: "UserPromptSubmit", Prompt: prompt}
	event, ok := buildDaemonEvent(input, "tab", "workspace", "hash")
	if !ok || len([]rune(event.Input.Text)) != maxInputRunes {
		t.Fatalf("event accepted=%v prompt runes=%d", ok, len([]rune(event.Input.Text)))
	}
	frame, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame)+1 >= maxFrameSize {
		t.Fatalf("truncated prompt frame=%d limit=%d", len(frame)+1, maxFrameSize)
	}

	// A control rune has JSON's largest per-rune expansion. Even that prompt
	// cannot reach the IPC frame ceiling after the 600-rune truncation.
	worstCase := hookInput{HookEventName: "UserPromptSubmit", Prompt: strings.Repeat("\x01", maxInputRunes+1)}
	worstEvent, ok := buildDaemonEvent(worstCase, "tab", "workspace", "hash")
	if !ok {
		t.Fatal("worst-case prompt was rejected")
	}
	worstFrame, err := json.Marshal(worstEvent)
	if err != nil {
		t.Fatal(err)
	}
	if len(worstFrame)+1 >= maxFrameSize {
		t.Fatalf("worst-case truncated prompt frame=%d limit=%d", len(worstFrame)+1, maxFrameSize)
	}

	boundedInput := hookInput{
		HookEventName:  "PostToolUse",
		TranscriptPath: strings.Repeat("\x01", maxTranscriptRunes+1),
		ToolName:       strings.Repeat("\x01", maxToolNameRunes+1),
	}
	boundedEvent, ok := buildDaemonEvent(
		boundedInput,
		strings.Repeat("\x01", maxTabIDRunes),
		strings.Repeat("\x01", maxWorkspaceIDRunes),
		"hash",
	)
	if !ok {
		t.Fatal("maximum-field event was rejected")
	}
	if len([]rune(boundedEvent.Input.Text)) != maxToolNameRunes+1 ||
		len([]rune(boundedEvent.Input.TranscriptPath)) != maxTranscriptRunes ||
		len([]rune(boundedEvent.Input.ToolName)) != maxToolNameRunes ||
		len([]rune(boundedEvent.Input.TabID)) != maxTabIDRunes ||
		len([]rune(boundedEvent.Input.WorkspaceID)) != maxWorkspaceIDRunes {
		t.Fatalf("bounded event=%#v", boundedEvent.Input)
	}
	maximumFrameEvent := boundedEvent
	maximumFrameEvent.Input.Text = strings.Repeat("\x01", maxInputRunes)
	boundedFrame, err := json.Marshal(maximumFrameEvent)
	if err != nil {
		t.Fatal(err)
	}
	if len(boundedFrame)+1 >= maxFrameSize {
		t.Fatalf("all fields at maximum encoded size: frame=%d limit=%d", len(boundedFrame)+1, maxFrameSize)
	}

	app := testApplication(t, &fakeCommandRunner{})
	delivered := make(chan daemonEvent, 1)
	app.dial = func(string, string, time.Duration) (net.Conn, error) {
		server, client := net.Pipe()
		go func() {
			defer server.Close()
			received, readErr := bufioReadFrame(server)
			if readErr == nil {
				var decoded daemonEvent
				readErr = json.Unmarshal(bytes.TrimSpace(received), &decoded)
				if readErr == nil {
					delivered <- decoded
				}
			}
			if readErr != nil {
				close(delivered)
				return
			}
			_, _ = server.Write([]byte{ackByte})
		}()
		return client, nil
	}
	hookPayload, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(hookPayload) <= maxFrameSize {
		t.Fatalf("test hook payload=%d is not larger than IPC frame limit", len(hookPayload))
	}
	if err := app.runHook(bytes.NewReader(hookPayload), "tab", "workspace"); err != nil {
		t.Fatal(err)
	}
	received, open := <-delivered
	if !open || len([]rune(received.Input.Text)) != maxInputRunes {
		t.Fatalf("ACK delivery open=%v prompt runes=%d", open, len([]rune(received.Input.Text)))
	}
}

func TestBuildDaemonEventRejectsOversizedActorIdentifiers(t *testing.T) {
	prefix := strings.Repeat("a", maxTabIDRunes)
	input := hookInput{HookEventName: "UserPromptSubmit", Prompt: "dummy"}
	for _, test := range []struct {
		name        string
		tabID       string
		workspaceID string
	}{
		{name: "tab collision candidate x", tabID: prefix + "x", workspaceID: "workspace"},
		{name: "tab collision candidate y", tabID: prefix + "y", workspaceID: "workspace"},
		{name: "workspace collision candidate x", tabID: "tab", workspaceID: prefix + "x"},
		{name: "workspace collision candidate y", tabID: "tab", workspaceID: prefix + "y"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, ok := buildDaemonEvent(input, test.tabID, test.workspaceID, "hash"); ok {
				t.Fatal("oversized actor identifier was truncated and accepted")
			}
			app := testApplication(t, &fakeCommandRunner{})
			var deliveries atomic.Int32
			app.dial = func(string, string, time.Duration) (net.Conn, error) {
				deliveries.Add(1)
				return nil, errors.New("unexpected IPC delivery")
			}
			payload := `{"hook_event_name":"UserPromptSubmit","prompt":"dummy"}`
			if err := app.runHook(strings.NewReader(payload), test.tabID, test.workspaceID); err != nil {
				t.Fatal(err)
			}
			if deliveries.Load() != 0 {
				t.Fatalf("oversized actor event deliveries=%d", deliveries.Load())
			}
		})
	}
	event, ok := buildDaemonEvent(input, prefix, prefix, "hash")
	if !ok || event.Input.TabID != prefix || event.Input.WorkspaceID != prefix {
		t.Fatalf("identifier at exact limit event=%#v accepted=%v", event.Input, ok)
	}
}

func TestHookFrameOverflowWritesOneDiagnostic(t *testing.T) {
	app := testApplication(t, &fakeCommandRunner{})
	app.buildHash = strings.Repeat("h", maxFrameSize)
	var diagnostic bytes.Buffer
	app.stderr = &diagnostic
	var deliveries atomic.Int32
	app.dial = func(string, string, time.Duration) (net.Conn, error) {
		deliveries.Add(1)
		return nil, errors.New("unexpected IPC delivery")
	}
	payload := `{"hook_event_name":"UserPromptSubmit","prompt":"dummy"}`
	if err := app.runHook(strings.NewReader(payload), "tab", "workspace"); err == nil {
		t.Fatal("oversized IPC frame unexpectedly succeeded")
	}
	if deliveries.Load() != 0 {
		t.Fatalf("oversized frame deliveries=%d", deliveries.Load())
	}
	if !strings.Contains(diagnostic.String(), "IPC frame limit") || strings.Count(diagnostic.String(), "\n") != 1 {
		t.Fatalf("frame diagnostic=%q", diagnostic.String())
	}
}

func detachedMarkerExecutable(t *testing.T) (string, string) {
	t.Helper()
	directory := t.TempDir()
	executable := filepath.Join(directory, "fake-herdr-title")
	marker := filepath.Join(directory, "started")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n: > \"${0%/*}/started\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return executable, marker
}

func TestDetachedDaemonStderrUsesPrivateRotatedLog(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(cacheDir, daemonLogFileName)
	if err := os.WriteFile(logPath, bytes.Repeat([]byte{'x'}, maxDaemonLogSize+1), 0o644); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(t.TempDir(), "fake-herdr-title")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nprintf 'daemon-visible\\n' >&2\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := startDetachedDaemon(executable, cacheDir); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		data, err := os.ReadFile(logPath)
		return err == nil && bytes.Contains(data, []byte("daemon-visible"))
	}, "detached daemon diagnostic")
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("daemon log mode=%#o", info.Mode().Perm())
	}
	if info.Size() > maxDaemonLogSize || info.Size() >= int64(maxDaemonLogSize+1) {
		t.Fatalf("daemon log was not rotated: size=%d", info.Size())
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(strings.Repeat("x", 64))) {
		t.Fatal("daemon log retained old oversized content")
	}
}

func TestOpenDaemonLogRejectsSymlinkWithoutMutatingTarget(t *testing.T) {
	assertTargetUnchanged := func(t *testing.T, target string, wantContent []byte) {
		t.Helper()
		content, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(content, wantContent) || info.Mode().Perm() != 0o640 {
			t.Fatalf("target changed: size=%d mode=%#o", len(content), info.Mode().Perm())
		}
	}
	t.Run("symlink", func(t *testing.T) {
		cacheDir := filepath.Join(t.TempDir(), "cache")
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "target.log")
		wantContent := []byte("target-content")
		if err := os.WriteFile(target, wantContent, 0o640); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(target, 0o640); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(cacheDir, daemonLogFileName)); err != nil {
			t.Fatal(err)
		}
		if log, err := openDaemonLog(cacheDir); err == nil {
			_ = log.Close()
			t.Fatal("daemon log symlink was accepted")
		}
		assertTargetUnchanged(t, target, wantContent)
		executable, marker := detachedMarkerExecutable(t)
		if err := startDetachedDaemon(executable, cacheDir); err != nil {
			t.Fatalf("symlink fallback prevented spawn: %v", err)
		}
		waitFor(t, time.Second, func() bool {
			_, err := os.Stat(marker)
			return err == nil
		}, "daemon spawn after log symlink rejection")
		assertTargetUnchanged(t, target, wantContent)
	})
	t.Run("hardlink", func(t *testing.T) {
		cacheDir := filepath.Join(t.TempDir(), "cache")
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "target.log")
		wantContent := bytes.Repeat([]byte{'h'}, maxDaemonLogSize+1)
		if err := os.WriteFile(target, wantContent, 0o640); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(target, 0o640); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(target, filepath.Join(cacheDir, daemonLogFileName)); err != nil {
			t.Fatal(err)
		}
		if log, err := openDaemonLog(cacheDir); err == nil {
			_ = log.Close()
			t.Fatal("daemon log hardlink was accepted")
		}
		assertTargetUnchanged(t, target, wantContent)
	})
}

func TestOpenDaemonLogRejectsSpecialFilesBeforeChmod(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(string) error
	}{
		{name: "fifo", setup: func(path string) error { return syscall.Mkfifo(path, 0o640) }},
		{name: "directory", setup: func(path string) error { return os.Mkdir(path, 0o750) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			cacheDir := filepath.Join(t.TempDir(), "cache")
			if err := os.MkdirAll(cacheDir, 0o700); err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(cacheDir, daemonLogFileName)
			if err := test.setup(logPath); err != nil {
				t.Fatal(err)
			}
			wantMode := os.FileMode(0o640)
			if test.name == "directory" {
				wantMode = 0o750
			}
			if err := os.Chmod(logPath, wantMode); err != nil {
				t.Fatal(err)
			}
			if log, err := openDaemonLog(cacheDir); err == nil {
				_ = log.Close()
				t.Fatalf("daemon log %s was accepted", test.name)
			}
			info, err := os.Lstat(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != wantMode {
				t.Fatalf("%s mode changed to %#o", test.name, info.Mode().Perm())
			}
		})
	}
}

func TestDetachedDaemonSpawnsWhenLogIsDirectory(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(filepath.Join(cacheDir, daemonLogFileName), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, marker := detachedMarkerExecutable(t)
	if err := startDetachedDaemon(executable, cacheDir); err != nil {
		t.Fatalf("directory log fallback prevented spawn: %v", err)
	}
	waitFor(t, time.Second, func() bool {
		_, err := os.Stat(marker)
		return err == nil
	}, "daemon spawn with directory log")
}

func TestAgentIDPresenceIsDecodedForSubagentGate(t *testing.T) {
	var input hookInput
	if err := json.Unmarshal([]byte(`{"hook_event_name":"PreToolUse","agent_id":"","tool_name":"Read"}`), &input); err != nil {
		t.Fatal(err)
	}
	if input.AgentID == nil {
		t.Fatal("present agent_id field was not distinguishable from an absent field")
	}
	if _, ok := buildDaemonEvent(input, "tab", "workspace", "hash"); ok {
		t.Fatal("present empty agent_id passed the subagent gate")
	}
}

func TestToolSummaryAllowlistAndRedaction(t *testing.T) {
	cases := []struct {
		name  string
		tool  string
		input map[string]any
		want  string
	}{
		{name: "bash description", tool: "Bash", input: map[string]any{"description": "run tests", "command": "secret"}, want: "Bash: run tests"},
		{name: "read path", tool: "Read", input: map[string]any{"file_path": "/tmp/file", "other": "secret"}, want: "Read: /tmp/file"},
		{name: "write path", tool: "Write", input: map[string]any{"file_path": "/tmp/file"}, want: "Write: /tmp/file"},
		{name: "edit path", tool: "Edit", input: map[string]any{"file_path": "/tmp/file"}, want: "Edit: /tmp/file"},
		{name: "grep path", tool: "Grep", input: map[string]any{"path": "/repo", "pattern": "secret"}, want: "Grep: /repo"},
		{name: "glob path", tool: "Glob", input: map[string]any{"path": "/repo", "pattern": "secret"}, want: "Glob: /repo"},
		{name: "unknown tool", tool: "WebFetch", input: map[string]any{"url": "secret"}, want: "WebFetch"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := summarizeToolInput(test.tool, test.input); got != test.want {
				t.Fatalf("summary = %q, want %q", got, test.want)
			}
		})
	}
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	longHex := strings.Repeat("a", 40)
	longBase64 := strings.Repeat("Q", 48)
	redacted := redactToolValue(secret + " " + longHex + " " + longBase64)
	if strings.Contains(redacted, secret) || strings.Contains(redacted, longHex) || strings.Contains(redacted, longBase64) {
		t.Fatalf("secret remained in %q", redacted)
	}
	if len([]rune(redactToolValue(strings.Repeat("長", maxToolInputRunes+20)))) != maxToolInputRunes {
		t.Fatal("tool summary was not rune-truncated")
	}
}

func TestCodexArgumentsUseSpecifiedModelSandboxAndOutput(t *testing.T) {
	want := []string{
		"exec", "--skip-git-repo-check", "-s", "read-only", "-m", "gpt-5.6-luna",
		"-c", "model_reasoning_effort=low", "-o", "/output", "-",
	}
	got := codexArguments("/output")
	if !equalStrings(got, want) {
		t.Fatalf("codex arguments = %#v, want %#v", got, want)
	}
}

func TestCodexRetryExhaustionDiagnosticOmitsPayload(t *testing.T) {
	exitErr := exec.Command("sh", "-c", "exit 7").Run()
	var typedExitErr *exec.ExitError
	if !errors.As(exitErr, &typedExitErr) {
		t.Fatalf("exit error=%T %v", exitErr, exitErr)
	}
	for _, test := range []struct {
		name string
		err  error
		want string
	}{
		{name: "exit code", err: fmt.Errorf("SECRET_ERROR: %w", typedExitErr), want: "7"},
		{name: "timeout", err: fmt.Errorf("SECRET_ERROR: %w", context.DeadlineExceeded), want: "timeout"},
		{name: "empty", err: fmt.Errorf("SECRET_ERROR: %w", errEmptyTitle), want: "empty"},
		{name: "unavailable", err: errors.New("SECRET_ERROR"), want: "unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := codexFailureExitStatus(test.err); got != test.want {
				t.Fatalf("exit status=%q want=%q", got, test.want)
			}
		})
	}

	clock := newManualClock()
	input := pendingInput{
		Text: "SECRET_PROMPT", TranscriptPath: "/SECRET_TRANSCRIPT", TabID: "tab", WorkspaceID: "workspace",
	}
	var diagnostic bytes.Buffer
	actor := &titleActor{
		key:   actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"},
		state: stateRunning, phase: phaseRunningCodex, active: &input,
		statuses: make(chan actorStatus, 8), stderr: &diagnostic,
		ctx: context.Background(), now: clock.Now, newTimer: clock.NewTimer, config: defaultRuntimeConfig(),
	}
	job := generationJob{actor: actor, input: input}
	wrappedExitErr := fmt.Errorf("SECRET_ERROR: %w", typedExitErr)
	actor.onJobResult(jobResult{job: job, phase: resultCodexFailure, err: wrappedExitErr})
	if diagnostic.Len() != 0 {
		t.Fatalf("first codex failure diagnostic=%q", diagnostic.String())
	}
	actor.onJobResult(jobResult{job: job, phase: resultCodexFailure, err: wrappedExitErr})
	want := fmt.Sprintf(
		"herdr-title: codex retries exhausted actor=%s exit_status=7\n",
		actor.key.storageKey(),
	)
	if diagnostic.String() != want || strings.Contains(diagnostic.String(), "SECRET") {
		t.Fatalf("exhaustion diagnostic=%q want=%q", diagnostic.String(), want)
	}
}

func TestSanitizeTitleAllowsHashAndEnforcesRuneLimit(t *testing.T) {
	if got := sanitizeTitle("Fix #123\nwith「quotes」"); got != "Fix #123 withquotes" {
		t.Fatalf("sanitized title = %q", got)
	}
	input := strings.Repeat("題", maxTitleRunes+10)
	if got := sanitizeTitle(input); len([]rune(got)) != maxTitleRunes {
		t.Fatalf("title rune length = %d", len([]rune(got)))
	}
	if got := sanitizeTitle("👨‍💻"); got != "" {
		t.Fatalf("emoji-only title = %q", got)
	}
}

func TestTitlePromptsSeparateTabWorkAndWorkspaceTask(t *testing.T) {
	tab := titlePrompt(titleRequest{Kind: tabKind, Input: pendingInput{Text: "fix tests"}})
	workspace := titlePrompt(titleRequest{Kind: workspaceKind, Input: pendingInput{Text: "fix tests"}})
	if !strings.Contains(tab, "動詞句") || strings.Contains(workspace, "動詞句") {
		t.Fatalf("tab prompt = %q; workspace prompt = %q", tab, workspace)
	}
	if !strings.Contains(workspace, "名詞句") || !strings.Contains(tab, "60文字以内") || !strings.Contains(tab, "#") {
		t.Fatal("title output constraints are missing")
	}
}

func TestTranscriptInputFiltersSubagentsAndDeduplicatesCurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	data := strings.Join([]string{
		`{"type":"user","message":{"content":"first"}}`,
		`{"type":"user","isSidechain":true,"message":{"content":"subagent"}}`,
		`{"type":"user","message":{"content":"<system-reminder>hidden</system-reminder> visible"}}`,
		`{"type":"user","message":{"content":"current"}}`,
	}, "\n")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	got := buildTitleInputs(filepath.Dir(path), path, "current")
	want := []string{"first", "visible", "current"}
	if !equalStrings(got, want) {
		t.Fatalf("inputs = %#v, want %#v", got, want)
	}
}

func TestPersisterMergesActorsAndHydratesCanonicalState(t *testing.T) {
	cacheDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	persister, err := startPersister(ctx, cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	keyA := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "a"}
	keyB := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "b"}
	titleA, titleB := "A", "B"
	startedA, startedB := int64(100), int64(200)
	if err := persister.put(ctx, keyA, stateDelta{Title: &titleA, LastStarted: &startedA}); err != nil {
		t.Fatal(err)
	}
	if err := persister.put(ctx, keyB, stateDelta{Title: &titleB, LastStarted: &startedB}); err != nil {
		t.Fatal(err)
	}
	cancel()

	info, err := os.Stat(filepath.Join(cacheDir, stateFileName))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("state mode = %o", info.Mode().Perm())
	}
	loaded, err := loadDiskState(filepath.Join(cacheDir, stateFileName))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Actors[keyA.storageKey()] != (persistedActor{Title: titleA, LastStarted: startedA}) ||
		loaded.Actors[keyB.storageKey()] != (persistedActor{Title: titleB, LastStarted: startedB}) {
		t.Fatalf("canonical state = %#v", loaded.Actors)
	}
}

func TestDaemonLockIsSingletonAndCacheIs0700(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "herdr-title")
	first, err := acquireDaemonLock(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	defer first.release()
	if _, err := acquireDaemonLock(cacheDir); !errors.Is(err, errAlreadyRunning) {
		t.Fatalf("second lock error = %v", err)
	}
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("cache mode = %o", info.Mode().Perm())
	}
	lockInfo, err := os.Stat(filepath.Join(cacheDir, lockFileName))
	if err != nil || lockInfo.Mode().Perm() != 0o600 {
		t.Fatalf("lock info = %v, err = %v", lockInfo, err)
	}
}

func TestPrepareSocketOnlyRemovesSockets(t *testing.T) {
	dir := t.TempDir()
	regular := filepath.Join(dir, "daemon.sock")
	if err := os.WriteFile(regular, []byte("do not remove"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prepareSocketPath(regular); err == nil {
		t.Fatal("regular-file collision was removed")
	}
	if _, err := os.Stat(regular); err != nil {
		t.Fatalf("regular file disappeared: %v", err)
	}
	if err := os.Remove(regular); err != nil {
		t.Fatal(err)
	}
	removed := false
	if err := prepareSocketPathWith(
		regular,
		func(string) (os.FileMode, error) { return os.ModeSocket, nil },
		func(string) error { removed = true; return nil },
	); err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("stale socket was not removed")
	}
}

func TestSameWorkspaceMultipleTabsUseIndependentActors(t *testing.T) {
	var mu sync.Mutex
	var renames [][]string
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			return request.Input.Text, nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			mu.Lock()
			renames = append(renames, append([]string(nil), args...))
			mu.Unlock()
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	for _, item := range []struct{ tab, title string }{{"a", "Title A"}, {"b", "Title B"}} {
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, item.title, item.tab, "workspace")); err != nil {
			t.Fatal(err)
		}
	}
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return containsCall(renames, "tab", "rename", "a", "Title A") &&
			containsCall(renames, "tab", "rename", "b", "Title B")
	}, "both tab renames")
	snapshot := runtime.Snapshot()
	if len(snapshot.States) != 2 {
		t.Fatalf("actor registry = %#v", snapshot.States)
	}
}

func TestGlobalCodexConcurrencyIsTwo(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			current := active.Add(1)
			for {
				old := maximum.Load()
				if current <= old || maximum.CompareAndSwap(old, current) {
					break
				}
			}
			started <- struct{}{}
			<-release
			active.Add(-1)
			return request.Input.Text, nil
		},
		herdr: func(context.Context, ...string) ([]byte, error) { return nil, nil },
	}
	app := testApplication(t, runner)
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	for _, tabID := range []string{"a", "b", "c"} {
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "Title "+tabID, tabID, "workspace")); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("two codex jobs did not start")
		}
	}
	select {
	case <-started:
		t.Fatal("third codex job bypassed the global semaphore")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("third codex job did not start after a slot opened")
	}
	if got := maximum.Load(); got != 2 {
		t.Fatalf("maximum codex concurrency = %d", got)
	}
}

func TestAdditionalEventDuringGenerationCannotBypassThrottle(t *testing.T) {
	var mu sync.Mutex
	var starts []time.Time
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			mu.Lock()
			starts = append(starts, time.Now())
			call := len(starts)
			mu.Unlock()
			if call == 1 {
				close(firstStarted)
				<-releaseFirst
			}
			return request.Input.Text, nil
		},
		herdr: func(context.Context, ...string) ([]byte, error) { return nil, nil },
	}
	app := testApplication(t, runner)
	app.config.tabThrottle = 160 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "first", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	<-firstStarted
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "second", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	close(releaseFirst)
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(starts) == 2
	}, "second throttled generation")
	mu.Lock()
	delta := starts[1].Sub(starts[0])
	mu.Unlock()
	if delta < 140*time.Millisecond {
		t.Fatalf("generation interval = %s, throttle was bypassed", delta)
	}
}

func TestNonIdleStatesOnlyReplacePendingInput(t *testing.T) {
	for _, state := range []actorState{stateScheduled, stateQueued, stateRunning, stateRetryPending} {
		t.Run(string(state), func(t *testing.T) {
			jobs := make(chan generationJob, 1)
			actor := &titleActor{state: state, jobs: jobs}
			input := pendingInput{Text: "latest", TabID: "tab", WorkspaceID: "workspace"}
			actor.onEvent(input)
			if actor.state != state {
				t.Fatalf("state changed from %s to %s", state, actor.state)
			}
			if actor.pending == nil || *actor.pending != input {
				t.Fatalf("pending = %#v", actor.pending)
			}
			if len(jobs) != 0 {
				t.Fatal("non-idle event enqueued an additional job")
			}
		})
	}
}

func TestWorkspaceGuardBlocksBeforeGenerationAndBecomesIdle(t *testing.T) {
	var tabLists atomic.Int32
	var generations atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			generations.Add(1)
			return "workspace title", nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				tabLists.Add(1)
				return []byte(`[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]`), nil
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	app.config.guardRetry = 500 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "workspace task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool { return tabLists.Load() == 1 }, "pre-generation tab guard")
	time.Sleep(50 * time.Millisecond)
	if generations.Load() != 0 {
		t.Fatal("workspace generation ran with multiple tabs")
	}
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if got := runtime.Snapshot().States[key]; got != stateIdle {
		t.Fatalf("guard-blocked actor state = %s", got)
	}
	time.Sleep(80 * time.Millisecond)
	if tabLists.Load() != 1 {
		t.Fatalf("normal multi-tab result retried %d times", tabLists.Load())
	}
}

func TestWorkspaceGuardChecksAgainBeforeRename(t *testing.T) {
	var mu sync.Mutex
	tabListCalls := 0
	workspaceRenames := 0
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) { return "workspace title", nil },
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			mu.Lock()
			defer mu.Unlock()
			if equalStrings(args, []string{"tab", "list"}) {
				tabListCalls++
				if tabListCalls == 1 {
					return []byte(`[{"tab_id":"a","workspace_id":"workspace"}]`), nil
				}
				return []byte(`[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]`), nil
			}
			if len(args) >= 2 && args[0] == "workspace" && args[1] == "rename" {
				workspaceRenames++
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "workspace task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		calls := tabListCalls
		renames := workspaceRenames
		mu.Unlock()
		return calls >= 2 && renames == 0 && runtime.Snapshot().States[key] == stateIdle
	}, "post-generation workspace guard")
}

func TestCodexFailureRetryUsesThrottleFloorAndKeepsActiveSnapshot(t *testing.T) {
	var mu sync.Mutex
	var inputs []string
	var starts []time.Time
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			mu.Lock()
			inputs = append(inputs, request.Input.Text)
			starts = append(starts, time.Now())
			call := len(inputs)
			mu.Unlock()
			if call == 1 {
				close(firstStarted)
				<-releaseFirst
				return "", errors.New("codex failed")
			}
			return "latest title", nil
		},
		herdr: func(context.Context, ...string) ([]byte, error) { return nil, nil },
	}
	app := testApplication(t, runner)
	app.config.tabThrottle = 150 * time.Millisecond
	app.config.retryDelay = 30 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "old", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	<-firstStarted
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "latest", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	close(releaseFirst)
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(inputs) == 3
	}, "codex retry")
	mu.Lock()
	defer mu.Unlock()
	if !equalStrings(inputs, []string{"old", "old", "latest"}) {
		t.Fatalf("retry inputs = %#v", inputs)
	}
	if delta := starts[1].Sub(starts[0]); delta < 130*time.Millisecond {
		t.Fatalf("codex retry interval = %s, throttle floor was not applied", delta)
	}
}

func TestRenameFailureRetriesRenameOnlyAfterFixedDelay(t *testing.T) {
	var generations atomic.Int32
	var mu sync.Mutex
	var renameTimes []time.Time
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			generations.Add(1)
			return "generated title", nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if len(args) >= 2 && args[0] == "tab" && args[1] == "rename" {
				mu.Lock()
				renameTimes = append(renameTimes, time.Now())
				call := len(renameTimes)
				mu.Unlock()
				if call == 1 {
					return nil, errors.New("rename failed")
				}
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	app.config.tabThrottle = 300 * time.Millisecond
	app.config.retryDelay = 45 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "current", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(renameTimes) == 2
	}, "rename-only retry")
	mu.Lock()
	delta := renameTimes[1].Sub(renameTimes[0])
	mu.Unlock()
	if generations.Load() != 1 {
		t.Fatalf("codex generation count = %d", generations.Load())
	}
	if delta < 35*time.Millisecond || delta >= 200*time.Millisecond {
		t.Fatalf("rename retry interval = %s", delta)
	}
}

func TestRenameRetryDropsStaleTitleForNewPendingInput(t *testing.T) {
	var mu sync.Mutex
	var generatedInputs []string
	var renamedTitles []string
	firstRename := make(chan struct{})
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			mu.Lock()
			generatedInputs = append(generatedInputs, request.Input.Text)
			mu.Unlock()
			return request.Input.Text, nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if len(args) >= 4 && args[0] == "tab" && args[1] == "rename" {
				mu.Lock()
				renameed := args[3]
				renamedTitles = append(renamedTitles, renameed)
				call := len(renamedTitles)
				mu.Unlock()
				if call == 1 {
					close(firstRename)
					return nil, errors.New("rename failed")
				}
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	app.config.tabThrottle = 120 * time.Millisecond
	app.config.retryDelay = 35 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "old title", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	<-firstRename
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "new title", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(renamedTitles) >= 2
	}, "new-input generation after stale rename retry")
	mu.Lock()
	defer mu.Unlock()
	if !equalStrings(generatedInputs, []string{"old title", "new title"}) {
		t.Fatalf("generated inputs = %#v", generatedInputs)
	}
	if !equalStrings(renamedTitles, []string{"old title", "new title"}) {
		t.Fatalf("rename titles = %#v", renamedTitles)
	}
}

func TestActorEvictionPersistsAndRecreationHydratesWithoutColdStart(t *testing.T) {
	var generations atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			generations.Add(1)
			return "persisted title", nil
		},
		herdr: func(context.Context, ...string) ([]byte, error) { return nil, nil },
	}
	app := testApplication(t, runner)
	app.config.tabThrottle = 500 * time.Millisecond
	app.config.actorEviction = 50 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "first", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool { return generations.Load() == 1 && len(runtime.Snapshot().States) == 0 }, "actor eviction")
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "second", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		snapshot := runtime.Snapshot()
		cold, exists := snapshot.ColdStarts[key]
		return exists && !cold && snapshot.States[key] == stateScheduled
	}, "hydrated actor recreation")
	time.Sleep(80 * time.Millisecond)
	if generations.Load() != 1 {
		t.Fatalf("hydrated actor bypassed throttle; generations = %d", generations.Load())
	}
}

func TestIdleSuicideRequiresEveryActorIdle(t *testing.T) {
	t.Run("no actors", func(t *testing.T) {
		app := testApplication(t, &fakeCommandRunner{})
		app.config.daemonIdle = 35 * time.Millisecond
		app.config.idlePoll = 5 * time.Millisecond
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-runtime.shutdown:
		case <-time.After(time.Second):
			t.Fatal("idle daemon did not shut down")
		}
		runtime.Stop()
	})

	t.Run("scheduled actor", func(t *testing.T) {
		runner := &fakeCommandRunner{
			generate: func(context.Context, titleRequest) (string, error) { return "title", nil },
			herdr:    func(context.Context, ...string) ([]byte, error) { return nil, nil },
		}
		app := testApplication(t, runner)
		app.config.tabThrottle = 500 * time.Millisecond
		app.config.daemonIdle = 35 * time.Millisecond
		app.config.idlePoll = 5 * time.Millisecond
		key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
		state := diskState{ProtocolVersion: protocolVersion, Actors: map[string]persistedActor{
			key.storageKey(): {Title: "old", LastStarted: time.Now().UnixNano()},
		}}
		if err := writeDiskState(app.cacheDir, state); err != nil {
			t.Fatal(err)
		}
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Stop()
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "pending", "tab", "workspace")); err != nil {
			t.Fatal(err)
		}
		waitFor(t, time.Second, func() bool { return runtime.Snapshot().States[key] == stateScheduled }, "scheduled state")
		select {
		case <-runtime.shutdown:
			t.Fatal("daemon shut down while an actor was scheduled")
		case <-time.After(100 * time.Millisecond):
		}
	})
}

func TestIPCACKNACKAndVersionHandover(t *testing.T) {
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) { return "title", nil },
		herdr:    func(context.Context, ...string) ([]byte, error) { return nil, nil },
	}

	t.Run("ack after dispatch", func(t *testing.T) {
		app := testApplication(t, runner)
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Stop()
		server, client := net.Pipe()
		go app.handleConnection(server, runtime)
		frame, _ := json.Marshal(eventFor(app, []actorKind{tabKind}, "title", "tab", "workspace"))
		if _, err := client.Write(append(frame, '\n')); err != nil {
			t.Fatal(err)
		}
		response := []byte{0}
		if _, err := io.ReadFull(client, response); err != nil {
			t.Fatal(err)
		}
		_ = client.Close()
		if response[0] != ackByte {
			t.Fatalf("response = %#x", response[0])
		}
		if len(runtime.Snapshot().States) != 1 {
			t.Fatal("ACK arrived without dispatcher actor creation")
		}
	})

	t.Run("version skew nack then shutdown", func(t *testing.T) {
		app := testApplication(t, runner)
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		server, client := net.Pipe()
		go app.handleConnection(server, runtime)
		event := eventFor(app, []actorKind{tabKind}, "title", "tab", "workspace")
		event.BuildHash = "new-build"
		frame, _ := json.Marshal(event)
		if _, err := client.Write(append(frame, '\n')); err != nil {
			t.Fatal(err)
		}
		response := []byte{0}
		if _, err := io.ReadFull(client, response); err != nil {
			t.Fatal(err)
		}
		if response[0] != nackByte {
			t.Fatalf("response = %#x", response[0])
		}
		select {
		case <-runtime.shutdown:
		case <-time.After(time.Second):
			t.Fatal("version-skew daemon did not shut down")
		}
		_ = client.Close()
		runtime.Stop()
	})

	t.Run("malformed and oversized frames nack without shutdown", func(t *testing.T) {
		app := testApplication(t, runner)
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Stop()
		for _, frame := range [][]byte{[]byte("not-json\n"), append([]byte(strings.Repeat("x", maxFrameSize)), '\n')} {
			server, client := net.Pipe()
			go app.handleConnection(server, runtime)
			writeDone := make(chan error, 1)
			go func() {
				_, writeErr := client.Write(frame)
				writeDone <- writeErr
			}()
			response := []byte{0}
			if _, err := io.ReadFull(client, response); err != nil {
				t.Fatal(err)
			}
			_ = client.Close()
			if err := <-writeDone; err != nil && !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "closed pipe") {
				t.Fatal(err)
			}
			if response[0] != nackByte {
				t.Fatalf("response = %#x", response[0])
			}
		}
		select {
		case <-runtime.shutdown:
			t.Fatal("bad frame shut down daemon")
		default:
		}
	})
}

func TestHookNACKWaitsForHandoverAndSpawnsOnce(t *testing.T) {
	app := testApplication(t, &fakeCommandRunner{})
	var spawnCount atomic.Int32
	app.spawn = func(string, string) error {
		spawnCount.Add(1)
		return nil
	}
	app.config.handoverDelay = 35 * time.Millisecond
	var dialCount atomic.Int32
	app.dial = func(string, string, time.Duration) (net.Conn, error) {
		server, client := net.Pipe()
		index := dialCount.Add(1)
		go func() {
			defer server.Close()
			_, _ = bufioReadFrame(server)
			response := nackByte
			if index == 2 {
				response = ackByte
			}
			_, _ = server.Write([]byte{response})
		}()
		return client, nil
	}
	begin := time.Now()
	err := app.runHook(strings.NewReader(`{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"/tmp/file"}}`), "tab", "workspace")
	if err != nil {
		t.Fatal(err)
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("spawn count = %d", spawnCount.Load())
	}
	if elapsed := time.Since(begin); elapsed < 30*time.Millisecond {
		t.Fatalf("handover wait = %s", elapsed)
	}
}

func bufioReadFrame(connection net.Conn) ([]byte, error) {
	reader := make([]byte, maxFrameSize)
	n := 0
	for n < len(reader) {
		read, err := connection.Read(reader[n : n+1])
		if err != nil {
			return reader[:n], err
		}
		n += read
		if reader[n-1] == '\n' {
			return reader[:n], nil
		}
	}
	return reader[:n], errors.New("frame too large")
}

func TestParseTabListActualEnvelopeAndFailClosed(t *testing.T) {
	data := []byte(`{"id":"cli:tab:list","result":{"tabs":[{"tab_id":"tab","workspace_id":"workspace"}],"type":"tab_list"}}`)
	tabs, err := parseTabList(data)
	if err != nil || len(tabs) != 1 || tabs[0].TabID != "tab" || tabs[0].WorkspaceID != "workspace" {
		t.Fatalf("tabs = %#v, err = %v", tabs, err)
	}
	for _, input := range []string{"not json", `{"result":{"type":"tab_list"}}`} {
		if _, err := parseTabList([]byte(input)); err == nil {
			t.Fatalf("unusable tab list was accepted: %s", input)
		}
	}
}

func TestV10Required1MultiTabGuardIdleEvictionAndShutdown(t *testing.T) {
	clock := newManualClock()
	var tabLists atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			return "", errors.New("codex must not run")
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				tabLists.Add(1)
				return []byte(`{"result":{"tabs":[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]}}`), nil
			}
			return nil, errors.New("unexpected rename")
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.actorEviction = 10 * time.Minute
	app.config.daemonIdle = 60 * time.Minute
	app.config.idlePoll = time.Minute
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		return tabLists.Load() == 1 && runtime.Snapshot().States[key] == stateIdle
	}, "#6 normal multi-tab idle")
	clock.Advance(10 * time.Minute)
	waitFor(t, time.Second, func() bool { return len(runtime.Snapshot().States) == 0 }, "#21 eviction")
	clock.Advance(50 * time.Minute)
	select {
	case <-runtime.shutdown:
	case <-time.After(time.Second):
		t.Fatal("#22 idle daemon did not shut down")
	}
	if tabLists.Load() != 1 {
		t.Fatalf("normal multi-tab result left a retry timer: calls=%d", tabLists.Load())
	}
}

func TestV10Required2GuardFailureExhaustionThenShutdown(t *testing.T) {
	clock := newManualClock()
	var tabLists atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			return "", errors.New("codex must not run")
		},
		herdr: func(ctx context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				tabLists.Add(1)
				return nil, errors.New("tab list failed")
			}
			return nil, errors.New("unexpected command")
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.guardRetry = 30 * time.Second
	app.config.daemonIdle = 60 * time.Minute
	app.config.actorEviction = 10 * time.Minute
	app.config.idlePoll = time.Minute
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		snapshot := runtime.Snapshot()
		return tabLists.Load() == 1 && snapshot.States[key] == stateRetryPending && snapshot.GuardRetries[key] == 1
	}, "first guard failure")
	for want := int32(2); want <= 3; want++ {
		clock.Advance(30 * time.Second)
		waitFor(t, time.Second, func() bool {
			snapshot := runtime.Snapshot()
			if want == 3 {
				return tabLists.Load() == want && snapshot.States[key] == stateIdle
			}
			return tabLists.Load() == want && snapshot.States[key] == stateRetryPending && snapshot.GuardRetries[key] == int(want)
		}, "guard retry")
	}
	waitFor(t, time.Second, func() bool { return runtime.Snapshot().States[key] == stateIdle }, "#8 guard exhaustion")
	clock.Advance(59 * time.Minute)
	select {
	case <-runtime.shutdown:
	case <-time.After(time.Second):
		t.Fatal("guard-exhausted daemon did not reach 60 minute shutdown")
	}
}

func TestV10Required3WorkspaceGuardThrottledPerWindow(t *testing.T) {
	clock := newManualClock()
	var tabLists atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) { return "unused", nil },
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				tabLists.Add(1)
				return []byte(`[{"tab_id":"a","workspace_id":"workspace"},{"tab_id":"b","workspace_id":"workspace"}]`), nil
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.workspaceThrottle = 30 * time.Minute
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		return tabLists.Load() == 1 && runtime.Snapshot().States[key] == stateIdle
	}, "first guard window")
	for index := 0; index < 9; index++ {
		if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
			t.Fatal(err)
		}
	}
	waitFor(t, time.Second, func() bool { return runtime.Snapshot().States[key] == stateScheduled }, "throttled workspace actor")
	clock.Advance(29 * time.Minute)
	time.Sleep(10 * time.Millisecond)
	if tabLists.Load() != 1 {
		t.Fatalf("guard called inside throttle window: %d", tabLists.Load())
	}
	clock.Advance(time.Minute)
	waitFor(t, time.Second, func() bool { return tabLists.Load() == 2 }, "next guard window")
}

func TestV10Required4RenameExhaustionRearmsOnNextEvent(t *testing.T) {
	clock := newManualClock()
	var generations atomic.Int32
	var renames atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			generations.Add(1)
			return "generated", nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if len(args) >= 2 && args[0] == "tab" && args[1] == "rename" {
				call := renames.Add(1)
				if call <= 2 {
					return nil, errors.New("rename failed")
				}
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.tabThrottle = 0
	app.config.retryDelay = 30 * time.Second
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "first", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		return renames.Load() == 1 && runtime.Snapshot().States[key] == stateRetryPending
	}, "first rename failure")
	clock.Advance(30 * time.Second)
	waitFor(t, time.Second, func() bool {
		return renames.Load() == 2 && runtime.Snapshot().States[key] == stateIdle
	}, "rename retry exhaustion")
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "second", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		return generations.Load() == 2 && renames.Load() == 3 && runtime.Snapshot().States[key] == stateIdle
	}, "next-event recovery")
}

func TestV10Required5RenameRetryStaleAndMatchingBranches(t *testing.T) {
	t.Run("stale pending starts a new job", func(t *testing.T) {
		clock := newManualClock()
		var mu sync.Mutex
		var generated, renamed []string
		runner := &fakeCommandRunner{
			generate: func(_ context.Context, request titleRequest) (string, error) {
				mu.Lock()
				generated = append(generated, request.Input.Text)
				mu.Unlock()
				return request.Input.Text, nil
			},
			herdr: func(_ context.Context, args ...string) ([]byte, error) {
				if len(args) >= 4 && args[0] == "tab" && args[1] == "rename" {
					mu.Lock()
					renamed = append(renamed, args[3])
					call := len(renamed)
					mu.Unlock()
					if call == 1 {
						return nil, errors.New("rename failed")
					}
				}
				return nil, nil
			},
		}
		app := testApplication(t, runner)
		useManualClock(app, clock)
		app.config.tabThrottle = 0
		app.config.retryDelay = 30 * time.Second
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Stop()
		key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "old", "tab", "workspace")); err != nil {
			t.Fatal(err)
		}
		waitFor(t, time.Second, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(renamed) == 1 && runtime.Snapshot().States[key] == stateRetryPending
		}, "old rename failure")
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "new", "tab", "workspace")); err != nil {
			t.Fatal(err)
		}
		waitFor(t, time.Second, func() bool { return runtime.Snapshot().PendingText[key] == "new" }, "new pending snapshot")
		clock.Advance(30 * time.Second)
		waitFor(t, time.Second, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(generated) == 2 && len(renamed) == 2
		}, "stale retry replacement")
		mu.Lock()
		defer mu.Unlock()
		if !equalStrings(generated, []string{"old", "new"}) || !equalStrings(renamed, []string{"old", "new"}) {
			t.Fatalf("generated=%#v renamed=%#v", generated, renamed)
		}
	})

	t.Run("nil pending retries the same title", func(t *testing.T) {
		clock := newManualClock()
		var generations atomic.Int32
		var mu sync.Mutex
		var renamed []string
		runner := &fakeCommandRunner{
			generate: func(context.Context, titleRequest) (string, error) {
				generations.Add(1)
				return "same", nil
			},
			herdr: func(_ context.Context, args ...string) ([]byte, error) {
				if len(args) >= 4 && args[0] == "tab" && args[1] == "rename" {
					mu.Lock()
					renamed = append(renamed, args[3])
					call := len(renamed)
					mu.Unlock()
					if call == 1 {
						return nil, errors.New("rename failed")
					}
				}
				return nil, nil
			},
		}
		app := testApplication(t, runner)
		useManualClock(app, clock)
		app.config.retryDelay = 30 * time.Second
		runtime, err := startDaemonRuntime(app)
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Stop()
		key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
		if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "same", "tab", "workspace")); err != nil {
			t.Fatal(err)
		}
		waitFor(t, time.Second, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(renamed) == 1 && runtime.Snapshot().States[key] == stateRetryPending
		}, "first rename")
		clock.Advance(30 * time.Second)
		waitFor(t, time.Second, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(renamed) == 2
		}, "same-title retry")
		if generations.Load() != 1 {
			t.Fatalf("rename-only retry regenerated title %d times", generations.Load())
		}
	})
}

func TestV10Required6GenerationAndRenameCountersAreIndependent(t *testing.T) {
	clock := newManualClock()
	var generationCalls atomic.Int32
	var tabLists atomic.Int32
	var renameCalls atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			if generationCalls.Add(1) == 1 {
				return "", errors.New("codex failed")
			}
			return "workspace title", nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				tabLists.Add(1)
				return []byte(`[{"tab_id":"a","workspace_id":"workspace"}]`), nil
			}
			if len(args) >= 2 && args[0] == "workspace" && args[1] == "rename" {
				renameCalls.Add(1)
				return nil, errors.New("rename failed")
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.workspaceThrottle = 0
	app.config.retryDelay = 30 * time.Second
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		snapshot := runtime.Snapshot()
		return generationCalls.Load() == 1 && snapshot.GenerationRetried[key]
	}, "#11 generation retry")
	clock.Advance(30 * time.Second)
	waitFor(t, time.Second, func() bool {
		snapshot := runtime.Snapshot()
		return generationCalls.Load() == 2 && renameCalls.Load() == 1 && snapshot.RenameRetryConsumed[key]
	}, "#18 rename retry")
	clock.Advance(30 * time.Second)
	waitFor(t, time.Second, func() bool {
		return renameCalls.Load() == 2 && runtime.Snapshot().States[key] == stateIdle
	}, "#16 rename exhaustion")
	if tabLists.Load() != 3 {
		t.Fatalf("guard calls=%d, want fire+fresh rename+retry rename", tabLists.Load())
	}
}

func TestV10Transition15RenameGuardFailureRetriesGeneratedTitle(t *testing.T) {
	clock := newManualClock()
	var generations atomic.Int32
	var tabLists atomic.Int32
	var renames atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			generations.Add(1)
			return "generated title", nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if equalStrings(args, []string{"tab", "list"}) {
				call := tabLists.Add(1)
				if call == 2 {
					return nil, errors.New("rename guard CLI failed")
				}
				return []byte(`[{"tab_id":"a","workspace_id":"workspace"}]`), nil
			}
			if len(args) >= 2 && args[0] == "workspace" && args[1] == "rename" {
				renames.Add(1)
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.workspaceThrottle = 0
	app.config.retryDelay = 30 * time.Second
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: workspaceKind, WorkspaceID: "workspace"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "a", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		snapshot := runtime.Snapshot()
		return tabLists.Load() == 2 && snapshot.States[key] == stateRetryPending && snapshot.RenameRetryConsumed[key]
	}, "#15 rename guard retry")
	clock.Advance(30 * time.Second)
	waitFor(t, time.Second, func() bool {
		return tabLists.Load() == 3 && renames.Load() == 1 && runtime.Snapshot().States[key] == stateIdle
	}, "#19b rename guard recheck")
	if generations.Load() != 1 {
		t.Fatalf("rename guard retry regenerated title %d times", generations.Load())
	}
}

func TestV10Required7SyntheticPayloadAndHookInputBoundaries(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "hook_payloads_synthetic.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	if len(lines) != 3 {
		t.Fatalf("fixture lines=%d", len(lines))
	}
	wantEventAccepted := []bool{true, true, false}
	for index, line := range lines {
		input, accepted, decodeErr := decodeHookInput(bytes.NewReader(line), io.Discard)
		if decodeErr != nil || !accepted {
			t.Fatalf("fixture %d accepted=%v err=%v", index, accepted, decodeErr)
		}
		if index == 0 && (input.Prompt != "DUMMY_PROMPT" || input.AgentID != nil) {
			t.Fatalf("main prompt fixture input=%#v", input)
		}
		if index == 1 && input.AgentID != nil {
			t.Fatal("main-agent tool fixture unexpectedly contains agent_id")
		}
		if index == 2 && (input.AgentID == nil || *input.AgentID != "dummy-agent-id") {
			t.Fatalf("subagent fixture input=%#v", input)
		}
		event, eventAccepted := buildDaemonEvent(input, "tab", "workspace", "hash")
		if eventAccepted != wantEventAccepted[index] {
			t.Fatalf("fixture %d buildDaemonEvent accepted=%v want=%v", index, eventAccepted, wantEventAccepted[index])
		}
		if index == 1 {
			if strings.Count(event.Input.Text, "[REDACTED]") != 4 {
				t.Fatalf("main-agent tool sanitization=%q", event.Input.Text)
			}
		}
	}
	if len(lines[2]) <= 70<<10 {
		t.Fatalf("synthetic tool_response line=%d bytes", len(lines[2]))
	}

	prefix := `{"hook_event_name":"UserPromptSubmit","prompt":"`
	suffix := `"}`
	below := prefix + strings.Repeat("x", maxHookInputSize-len(prefix)-len(suffix)) + suffix
	input, accepted, err := decodeHookInput(strings.NewReader(below), io.Discard)
	if err != nil || !accepted || len(input.Prompt) == 0 || len(below) != maxHookInputSize {
		t.Fatalf("8MiB boundary accepted=%v size=%d err=%v", accepted, len(below), err)
	}
	var diagnostic bytes.Buffer
	_, accepted, err = decodeHookInput(strings.NewReader(below+" "), &diagnostic)
	if err != nil || accepted || !strings.Contains(diagnostic.String(), "too large") {
		t.Fatalf("8MiB+1 accepted=%v diagnostic=%q err=%v", accepted, diagnostic.String(), err)
	}
	app := testApplication(t, &fakeCommandRunner{})
	var hookDiagnostic bytes.Buffer
	var deliveries atomic.Int32
	app.stderr = &hookDiagnostic
	app.dial = func(string, string, time.Duration) (net.Conn, error) {
		deliveries.Add(1)
		return nil, errors.New("unexpected IPC delivery")
	}
	if err := app.runHook(strings.NewReader(below+" "), "tab", "workspace"); err != nil {
		t.Fatalf("oversized hook input was not fail-open: %v", err)
	}
	if deliveries.Load() != 0 || !strings.Contains(hookDiagnostic.String(), "too large") {
		t.Fatalf("oversized event deliveries=%d diagnostic=%q", deliveries.Load(), hookDiagnostic.String())
	}
}

func TestV10Required8SyntheticFixtureLeakInspection(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "hook_payloads_synthetic.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"captured_payloads", "/Users/s23159", "REAL_PROMPT", "REAL_TOOL_RESPONSE", "secret",
	} {
		if bytes.Contains(data, []byte(forbidden)) {
			t.Fatalf("synthetic fixture contains forbidden marker %q", forbidden)
		}
	}
	dummyPatterns := []struct {
		value   string
		pattern *regexp.Regexp
	}{
		{value: "Bearer DUMMY.BEARER.TOKEN", pattern: bearerPattern},
		{value: "ghp_DUMMYTOKEN123456", pattern: knownTokenPattern},
		{value: "0123456789abcdef0123456789abcdef", pattern: longHexPattern},
		{value: "QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVo0123456789", pattern: longBase64Pattern},
	}
	scrubbed := append([]byte(nil), data...)
	for _, dummy := range dummyPatterns {
		if !bytes.Contains(data, []byte(dummy.value)) || !dummy.pattern.MatchString(dummy.value) {
			t.Fatalf("synthetic sanitization marker is inactive: %q", dummy.value)
		}
		scrubbed = bytes.ReplaceAll(scrubbed, []byte(dummy.value), []byte("DUMMY_PATTERN"))
	}
	if bearerPattern.Match(scrubbed) || knownTokenPattern.Match(scrubbed) ||
		longHexPattern.Match(scrubbed) || longBase64Pattern.Match(scrubbed) {
		t.Fatal("synthetic fixture contains an undeclared credential-like value")
	}
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	var mainPayload, subagentPayload map[string]any
	if err := json.Unmarshal(lines[0], &mainPayload); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(lines[2], &subagentPayload); err != nil {
		t.Fatal(err)
	}
	if _, exists := mainPayload["agent_id"]; exists {
		t.Fatal("main-agent fixture unexpectedly contains agent_id")
	}
	if subagentPayload["agent_id"] != "dummy-agent-id" || subagentPayload["agent_type"] != "Explore" {
		t.Fatalf("subagent keys=%#v", subagentPayload)
	}
	response, ok := subagentPayload["tool_response"].(map[string]any)
	if !ok {
		t.Fatal("synthetic tool_response missing")
	}
	content, _ := response["content"].(string)
	if !strings.HasPrefix(content, "DUMMY_TOOL_RESPONSE_") || strings.Trim(content[len("DUMMY_TOOL_RESPONSE_"):], ".") != "" {
		t.Fatal("tool_response is not the expected dummy-only content")
	}
}

func TestV10Required9HerdrSemaphoreTimeoutsAreBounded(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	var calls atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) {
			return "", errors.New("codex must not run")
		},
		herdr: func(ctx context.Context, args ...string) ([]byte, error) {
			if !equalStrings(args, []string{"tab", "list"}) {
				return nil, errors.New("unexpected command")
			}
			calls.Add(1)
			current := active.Add(1)
			for {
				old := maximum.Load()
				if current <= old || maximum.CompareAndSwap(old, current) {
					break
				}
			}
			<-ctx.Done()
			active.Add(-1)
			return nil, ctx.Err()
		},
	}
	app := testApplication(t, runner)
	app.config.herdrTimeout = 15 * time.Millisecond
	app.config.guardRetry = time.Millisecond
	app.config.actorEviction = 20 * time.Millisecond
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	for index := 0; index < 8; index++ {
		workspace := string(rune('a' + index))
		if err := runtime.Dispatch(eventFor(app, []actorKind{workspaceKind}, "task", "tab", workspace)); err != nil {
			t.Fatal(err)
		}
	}
	waitFor(t, 2*time.Second, func() bool { return calls.Load() == 24 && active.Load() == 0 }, "three bounded guard attempts per actor")
	waitFor(t, time.Second, func() bool { return len(runtime.Snapshot().States) == 0 }, "timed-out actor eviction")
	herdrInUse, herdrWaiters := runtime.herdrSem.Stats()
	codexInUse, codexWaiters := runtime.codexSem.Stats()
	if maximum.Load() > 4 || maximum.Load() < 2 {
		t.Fatalf("herdr concurrency maximum=%d", maximum.Load())
	}
	if herdrInUse != 0 || herdrWaiters != 0 || codexInUse != 0 || codexWaiters != 0 {
		t.Fatalf("semaphore residue herdr=(%d,%d) codex=(%d,%d)", herdrInUse, herdrWaiters, codexInUse, codexWaiters)
	}
}

func TestV10Required9HerdrTimeoutKillsChildProcessGroup(t *testing.T) {
	runner := execCommandRunner{herdrCommand: "/bin/sh"}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	output, err := runner.RunHerdr(ctx, "-c", "sleep 30 & child=$!; echo $child; wait")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("herdr timeout error=%v output=%q", err, output)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil || pid <= 0 {
		t.Fatalf("child pid output=%q err=%v", output, err)
	}
	waitFor(t, time.Second, func() bool {
		return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH)
	}, "timed-out herdr child process removal")
}

func TestV10Required10TranscriptPathValidation(t *testing.T) {
	projects := t.TempDir()
	valid := filepath.Join(projects, "session.jsonl")
	if err := os.WriteFile(valid, []byte(`{"type":"user","message":{"content":"dummy"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := openValidatedTranscript(projects, valid)
	if err != nil {
		t.Fatalf("valid transcript rejected: %v", err)
	}
	_ = file.Close()

	symlink := filepath.Join(projects, "link.jsonl")
	if err := os.Symlink(valid, symlink); err != nil {
		t.Fatal(err)
	}
	if file, err := openValidatedTranscript(projects, symlink); err == nil {
		file.Close()
		t.Fatal("symlink transcript accepted")
	}
	parentPath := projects + string(os.PathSeparator) + "sub" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "session.jsonl"
	if file, err := openValidatedTranscript(projects, parentPath); err == nil {
		file.Close()
		t.Fatal("parent-component transcript accepted")
	}
	if file, err := openValidatedTranscript(projects, projects); err == nil {
		file.Close()
		t.Fatal("non-regular transcript accepted")
	}
	outside := filepath.Join(t.TempDir(), "outside.jsonl")
	if err := os.WriteFile(outside, []byte("dummy"), 0o600); err != nil {
		t.Fatal(err)
	}
	if file, err := openValidatedTranscript(projects, outside); err == nil {
		file.Close()
		t.Fatal("out-of-project transcript accepted")
	}
}

func TestV10Required11LockHandoverUsesAbsoluteDeadline(t *testing.T) {
	t.Run("lock becomes available before deadline", func(t *testing.T) {
		clock := newManualClock()
		deadline := clock.Now().Add(506 * time.Millisecond)
		available := deadline.Add(-6 * time.Millisecond)
		attempts := 0
		lock, err := acquireDaemonLockUntil(
			"/dummy", deadline, clock.Now, clock.Sleep, 10*time.Millisecond, io.Discard,
			func(string) (*daemonLock, error) {
				attempts++
				if clock.Now().Before(available) {
					return nil, errAlreadyRunning
				}
				return &daemonLock{}, nil
			},
		)
		if err != nil || lock == nil || clock.Now().After(deadline) || attempts < 2 {
			t.Fatalf("lock=%v attempts=%d now=%s deadline=%s err=%v", lock, attempts, clock.Now(), deadline, err)
		}
	})

	t.Run("lock remains held past deadline", func(t *testing.T) {
		clock := newManualClock()
		deadline := clock.Now().Add(506 * time.Millisecond)
		var diagnostic bytes.Buffer
		lock, err := acquireDaemonLockUntil(
			"/dummy", deadline, clock.Now, clock.Sleep, 10*time.Millisecond, &diagnostic,
			func(string) (*daemonLock, error) { return nil, errAlreadyRunning },
		)
		if lock != nil || !errors.Is(err, context.DeadlineExceeded) || !clock.Now().Equal(deadline) {
			t.Fatalf("lock=%v now=%s deadline=%s err=%v", lock, clock.Now(), deadline, err)
		}
		if !strings.Contains(diagnostic.String(), "deadline exceeded") {
			t.Fatalf("diagnostic=%q", diagnostic.String())
		}
	})
}

func TestV10Required12TransitionRowsAndTimerBranches(t *testing.T) {
	requiredRows := []string{
		"#1", "#2", "#3", "#4", "#5", "#6", "#7", "#7b", "#8", "#9a", "#9b",
		"#10", "#11", "#11b", "#12", "#14", "#15", "#17", "#18", "#19", "#19b",
		"#16/#19c", "#21", "#22",
	}
	// This literal table checks only that every canonical row has a structural
	// note. It does not prove that a named test exists or behaviorally covers
	// each row; the table-driven assertions below cover the timer branches.
	structuralNotes := map[string]string{
		"#1": "trailing generation schedule", "#2": "immediate fresh queue", "#3": "pending-only update",
		"#4": "generation timer", "#5": "semaphore pickup", "#6": "normal fire guard block",
		"#7": "fire guard retry", "#7b": "fire guard retry timer", "#8": "fire guard exhaustion",
		"#9a": "fire guard release", "#9b": "codex queue", "#10": "generation success",
		"#11": "generation retry", "#11b": "generation retry timer", "#12": "generation exhaustion",
		"#14": "normal rename guard skip", "#15": "rename guard retry", "#17": "rename success",
		"#18": "rename retry", "#19": "stale rename retry", "#19b": "matching rename retry",
		"#16/#19c": "rename exhaustion", "#21": "actor eviction", "#22": "daemon idle shutdown",
	}
	for _, row := range requiredRows {
		if structuralNotes[row] == "" {
			t.Fatalf("transition row %s has no structural note", row)
		}
	}

	clock := newManualClock()
	makeActor := func() (*titleActor, chan generationJob) {
		jobs := make(chan generationJob, 1)
		return &titleActor{
			key:   actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"},
			state: stateScheduled, phase: phaseNone, jobs: jobs,
			statuses: make(chan actorStatus, 8), ctx: context.Background(), now: clock.Now,
			newTimer: clock.NewTimer, config: defaultRuntimeConfig(),
		}, jobs
	}
	input := pendingInput{Text: "active", TabID: "tab", WorkspaceID: "workspace"}
	tests := []struct {
		name  string
		mode  timerMode
		setup func(*titleActor)
		want  jobMode
	}{
		{name: "#4 generation timer resets counters", mode: timerGeneration, setup: func(actor *titleActor) {
			actor.pending = &input
			actor.guardRetryCount = 2
			actor.generationRetried = true
			actor.renameRetryConsumed = true
		}, want: jobFresh},
		{name: "#7b guard timer preserves counter", mode: timerGuardRetry, setup: func(actor *titleActor) {
			actor.active = &input
			actor.guardRetryCount = 1
		}, want: jobGuardRetry},
		{name: "#11b codex timer preserves counter", mode: timerCodexRetry, setup: func(actor *titleActor) {
			actor.active = &input
			actor.generationRetried = true
		}, want: jobCodexRetry},
		{name: "#19b rename timer preserves counter", mode: timerRenameRetry, setup: func(actor *titleActor) {
			actor.active = &input
			actor.renameRetry = &renameRetry{input: input, title: "title"}
			actor.renameRetryConsumed = true
		}, want: jobRenameRetry},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actor, jobs := makeActor()
			test.setup(actor)
			actor.onTimer(test.mode)
			job := <-jobs
			if job.mode != test.want || actor.state != stateQueued {
				t.Fatalf("job mode=%d state=%s", job.mode, actor.state)
			}
			if test.want == jobFresh {
				if actor.guardRetryCount != 0 || actor.generationRetried || actor.renameRetryConsumed {
					t.Fatal("fresh #4 did not reset every counter")
				}
			} else if test.want == jobGuardRetry && actor.guardRetryCount != 1 {
				t.Fatal("guard retry reset its counter")
			} else if test.want == jobCodexRetry && !actor.generationRetried {
				t.Fatal("codex retry reset its counter")
			} else if test.want == jobRenameRetry && !actor.renameRetryConsumed {
				t.Fatal("rename retry reset its counter")
			}
		})
	}
}

func TestV10Required13FreshTitleIsRenamedBeforePendingRegeneration(t *testing.T) {
	clock := newManualClock()
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var generated, renamed []string
	runner := &fakeCommandRunner{
		generate: func(_ context.Context, request titleRequest) (string, error) {
			mu.Lock()
			generated = append(generated, request.Input.Text)
			call := len(generated)
			mu.Unlock()
			if call == 1 {
				close(firstStarted)
				<-releaseFirst
			}
			return request.Input.Text, nil
		},
		herdr: func(_ context.Context, args ...string) ([]byte, error) {
			if len(args) >= 4 && args[0] == "tab" && args[1] == "rename" {
				mu.Lock()
				renamed = append(renamed, args[3])
				mu.Unlock()
			}
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.tabThrottle = 0
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "paid", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	<-firstStarted
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "pending", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	close(releaseFirst)
	waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(generated) == 2 && len(renamed) == 2
	}, "fresh rename followed by pending generation")
	mu.Lock()
	defer mu.Unlock()
	if !equalStrings(generated, []string{"paid", "pending"}) || !equalStrings(renamed, []string{"paid", "pending"}) {
		t.Fatalf("generated=%#v renamed=%#v", generated, renamed)
	}
}

func TestV10Required14AbandoningTerminalsDiscardNewestPending(t *testing.T) {
	clock := newManualClock()
	input := pendingInput{Text: "active", TabID: "tab", WorkspaceID: "workspace"}
	pending := pendingInput{Text: "newest", TabID: "tab", WorkspaceID: "workspace"}
	tests := []struct {
		name  string
		phase resultPhase
		setup func(*titleActor)
	}{
		{name: "#6 fire guard normal block", phase: resultGuardBlocked, setup: func(*titleActor) {}},
		{name: "#8 guard exhaustion", phase: resultGuardFailure, setup: func(actor *titleActor) { actor.guardRetryCount = 2 }},
		{name: "#12 codex exhaustion", phase: resultCodexFailure, setup: func(actor *titleActor) { actor.generationRetried = true }},
		{name: "#14 rename guard normal skip", phase: resultRenameGuardSkipped, setup: func(*titleActor) {}},
		{name: "#16 rename exhaustion", phase: resultRenameFailure, setup: func(actor *titleActor) { actor.renameRetryConsumed = true }},
		{name: "#19c rename guard exhaustion", phase: resultRenameGuardFailure, setup: func(actor *titleActor) { actor.renameRetryConsumed = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			persister, err := startPersister(ctx, t.TempDir())
			if err != nil {
				cancel()
				t.Fatal(err)
			}
			defer func() {
				cancel()
				<-persister.done
			}()
			actor := &titleActor{
				key:   actorKey{Kind: workspaceKind, WorkspaceID: "workspace"},
				state: stateRunning, phase: phaseRunningCodex,
				active: &input, pending: &pending,
				statuses: make(chan actorStatus, 8), ctx: ctx, persister: persister, now: clock.Now,
				newTimer: clock.NewTimer, config: defaultRuntimeConfig(),
			}
			test.setup(actor)
			actor.onJobResult(jobResult{
				job: generationJob{actor: actor, input: input}, phase: test.phase, title: "title", err: errors.New("failed"),
			})
			if actor.state != stateIdle || actor.active != nil || actor.pending != nil || actor.renameRetry != nil {
				t.Fatalf("state=%s active=%#v pending=%#v retry=%#v", actor.state, actor.active, actor.pending, actor.renameRetry)
			}
			if test.phase == resultGuardBlocked {
				persisted, found, err := persister.get(ctx, actor.key)
				if err != nil || !found || persisted.LastStarted == 0 {
					t.Fatalf("#6 persisted=%#v found=%v err=%v", persisted, found, err)
				}
			}
		})
	}
}

func TestV10Required15RenamePersistencePrecedesIdleShutdown(t *testing.T) {
	clock := newManualClock()
	var renames atomic.Int32
	runner := &fakeCommandRunner{
		generate: func(context.Context, titleRequest) (string, error) { return "persisted title", nil },
		herdr: func(context.Context, ...string) ([]byte, error) {
			renames.Add(1)
			return nil, nil
		},
	}
	app := testApplication(t, runner)
	useManualClock(app, clock)
	app.config.tabThrottle = 0
	app.config.daemonIdle = 60 * time.Minute
	app.config.idlePoll = time.Minute
	runtime, err := startDaemonRuntime(app)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop()
	key := actorKey{Kind: tabKind, WorkspaceID: "workspace", TabID: "tab"}
	if err := runtime.Dispatch(eventFor(app, []actorKind{tabKind}, "task", "tab", "workspace")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, time.Second, func() bool {
		return renames.Load() == 1 && runtime.Snapshot().States[key] == stateIdle
	}, "rename success idle")
	state, err := loadDiskState(filepath.Join(app.cacheDir, stateFileName))
	if err != nil {
		t.Fatal(err)
	}
	if state.Actors[key.storageKey()].Title != "persisted title" {
		t.Fatalf("state before shutdown=%#v", state.Actors[key.storageKey()])
	}
	clock.Advance(60 * time.Minute)
	select {
	case <-runtime.shutdown:
	case <-time.After(time.Second):
		t.Fatal("persisted idle daemon did not shut down")
	}
}

func TestHashExecutableTestBinaryProxyCostAgainstHookBudget(t *testing.T) {
	// os.Executable points at the go test binary here. This is a proxy for the
	// production hook binary cost, not a measurement of the installed hook.
	// The reopened-cycle-1 proxy measurement was 2.672334ms.
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	hash, err := hashExecutable(executable)
	elapsed := time.Since(started)
	if err != nil || len(hash) != sha256.Size*2 {
		t.Fatalf("hash length=%d elapsed=%s err=%v", len(hash), elapsed, err)
	}
	budget := 50 * time.Millisecond
	if elapsed > budget {
		t.Fatalf("hashExecutable test-binary proxy cost=%s exceeds budget=%s", elapsed, budget)
	}
	t.Logf("hashExecutable test-binary proxy cost=%s budget=%s within_budget=true", elapsed, budget)
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

func containsCall(calls [][]string, want ...string) bool {
	for _, call := range calls {
		if equalStrings(call, want) {
			return true
		}
	}
	return false
}
