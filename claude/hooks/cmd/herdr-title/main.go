// herdr-title receives Claude Code hook events and delegates title generation
// to a singleton daemon shared by every Herdr workspace and tab.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	"syscall"
	"time"
	"unicode"
)

const (
	protocolVersion  = 1
	ackByte          = byte(0x06)
	nackByte         = byte(0x15)
	maxFrameSize     = 64 << 10
	maxHookInputSize = 8 << 20

	stateDirName         = "herdr-title"
	socketFileName       = "daemon.sock"
	lockFileName         = "daemon.lock"
	stateFileName        = "state.json"
	daemonLogFileName    = "daemon.log"
	maxDaemonLogSize     = 1 << 20
	fullReadLimit        = 32 << 20
	firstHistoryEntries  = 3
	recentHistoryEntries = 6
	maxInputRunes        = 600
	maxToolInputRunes    = 240
	maxTranscriptRunes   = 4096
	maxToolNameRunes     = 256
	maxWorkspaceIDRunes  = 256
	maxTabIDRunes        = 256
	maxTitleRunes        = 60
)

var (
	errAlreadyRunning = errors.New("herdr-title daemon is already running")
	errEmptyTitle     = errors.New("codex returned an empty title")

	systemReminderPattern = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)
	bearerPattern         = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`)
	knownTokenPattern     = regexp.MustCompile(`(?i)\b(?:sk-[A-Za-z0-9_-]{8,}|gh[pousr]_[A-Za-z0-9_]{8,}|glpat-[A-Za-z0-9_-]{8,}|xox[baprs]-[A-Za-z0-9-]{8,}|AKIA[A-Z0-9]{12,})\b`)
	longHexPattern        = regexp.MustCompile(`(?i)\b[0-9a-f]{32,}\b`)
	longBase64Pattern     = regexp.MustCompile(`\b[A-Za-z0-9+/]{40,}={0,2}\b`)
)

type actorKind string

const (
	tabKind       actorKind = "tab"
	workspaceKind actorKind = "workspace"
)

type actorState string

const (
	stateIdle         actorState = "idle"
	stateScheduled    actorState = "scheduled"
	stateQueued       actorState = "queued"
	stateRunning      actorState = "running"
	stateRetryPending actorState = "retry-pending"
)

type timerMode int

const (
	timerNone timerMode = iota
	timerGeneration
	timerGuardRetry
	timerCodexRetry
	timerRenameRetry
)

type runtimeConfig struct {
	tabThrottle       time.Duration
	workspaceThrottle time.Duration
	retryDelay        time.Duration
	guardRetry        time.Duration
	actorEviction     time.Duration
	daemonIdle        time.Duration
	idlePoll          time.Duration
	codexTimeout      time.Duration
	herdrTimeout      time.Duration
	ipcDeadline       time.Duration
	hookACKTimeout    time.Duration
	hookLatencyLimit  time.Duration
	handoverDelay     time.Duration
	spawnSettle       time.Duration
	queueSize         int
	actorInboxSize    int
	lockRetryInterval time.Duration
}

type clockTimer interface {
	Chan() <-chan time.Time
	Stop() bool
}

type clockTicker interface {
	Chan() <-chan time.Time
	Stop()
}

type realClockTimer struct {
	timer *time.Timer
}

func (t realClockTimer) Chan() <-chan time.Time { return t.timer.C }
func (t realClockTimer) Stop() bool             { return t.timer.Stop() }

type realClockTicker struct {
	ticker *time.Ticker
}

func (t realClockTicker) Chan() <-chan time.Time { return t.ticker.C }
func (t realClockTicker) Stop()                  { t.ticker.Stop() }

func defaultRuntimeConfig() runtimeConfig {
	return runtimeConfig{
		tabThrottle:       120 * time.Second,
		workspaceThrottle: 30 * time.Minute,
		retryDelay:        30 * time.Second,
		guardRetry:        30 * time.Second,
		actorEviction:     10 * time.Minute,
		daemonIdle:        60 * time.Minute,
		idlePoll:          time.Minute,
		codexTimeout:      60 * time.Second,
		herdrTimeout:      5 * time.Second,
		ipcDeadline:       time.Second,
		hookACKTimeout:    200 * time.Millisecond,
		hookLatencyLimit:  506 * time.Millisecond,
		handoverDelay:     300 * time.Millisecond,
		spawnSettle:       20 * time.Millisecond,
		queueSize:         256,
		actorInboxSize:    64,
		lockRetryInterval: 10 * time.Millisecond,
	}
}

type hookInput struct {
	HookEventName  string         `json:"hook_event_name"`
	Prompt         string         `json:"prompt"`
	TranscriptPath string         `json:"transcript_path"`
	AgentID        *string        `json:"agent_id"`
	ToolName       string         `json:"tool_name"`
	ToolInput      map[string]any `json:"tool_input"`
}

type pendingInput struct {
	Text           string `json:"text"`
	TranscriptPath string `json:"transcript_path"`
	TabID          string `json:"tab_id"`
	WorkspaceID    string `json:"workspace_id"`
	HookEventName  string `json:"hook_event_name"`
	ToolName       string `json:"tool_name,omitempty"`
}

type daemonEvent struct {
	ProtocolVersion int          `json:"protocol_version"`
	BuildHash       string       `json:"build_hash"`
	Kinds           []actorKind  `json:"kinds"`
	Input           pendingInput `json:"input"`
}

func (e daemonEvent) validate() error {
	if e.ProtocolVersion <= 0 || e.BuildHash == "" || len(e.Kinds) == 0 {
		return errors.New("missing protocol metadata or actor kinds")
	}
	if e.Input.TabID == "" && e.Input.WorkspaceID == "" {
		return errors.New("event has no Herdr identity")
	}
	seen := make(map[actorKind]bool, len(e.Kinds))
	for _, kind := range e.Kinds {
		if kind != tabKind && kind != workspaceKind {
			return fmt.Errorf("unknown actor kind %q", kind)
		}
		if seen[kind] {
			return fmt.Errorf("duplicate actor kind %q", kind)
		}
		seen[kind] = true
		if kind == tabKind && e.Input.TabID == "" {
			return errors.New("tab event has no tab ID")
		}
		if kind == workspaceKind && e.Input.WorkspaceID == "" {
			return errors.New("workspace event has no workspace ID")
		}
	}
	return nil
}

type actorKey struct {
	Kind        actorKind
	WorkspaceID string
	TabID       string
}

func (k actorKey) storageKey() string {
	data, _ := json.Marshal(k)
	return base64.RawURLEncoding.EncodeToString(data)
}

type persistedActor struct {
	Title       string `json:"title"`
	LastStarted int64  `json:"lastStarted"`
}

type diskState struct {
	ProtocolVersion int                       `json:"protocol_version"`
	Actors          map[string]persistedActor `json:"actors"`
}

type stateDelta struct {
	Title       *string
	LastStarted *int64
}

type titleRequest struct {
	Kind        actorKind
	Input       pendingInput
	Title       string
	ProjectsDir string
}

type commandRunner interface {
	GenerateTitle(context.Context, titleRequest) (string, error)
	RunHerdr(context.Context, ...string) ([]byte, error)
}

type execCommandRunner struct {
	codexCommand string
	herdrCommand string
	cacheDir     string
	environ      func() []string
	timeout      time.Duration
}

func codexArguments(outputPath string) []string {
	// The isolated work directory is not a Git repository. Without this flag,
	// Codex rejects every generation with exit status 1 before reading stdin.
	return []string{
		"exec", "--skip-git-repo-check", "-s", "read-only", "-m", "gpt-5.6-luna",
		"-c", "model_reasoning_effort=low", "-o", outputPath, "-",
	}
}

func (r execCommandRunner) GenerateTitle(parent context.Context, request titleRequest) (string, error) {
	workDir := filepath.Join(r.cacheDir, "work")
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(workDir, 0o700); err != nil {
		return "", err
	}
	output, err := os.CreateTemp(r.cacheDir, "output-*")
	if err != nil {
		return "", err
	}
	outputPath := output.Name()
	if err := output.Chmod(0o600); err != nil {
		output.Close()
		os.Remove(outputPath)
		return "", err
	}
	if err := output.Close(); err != nil {
		os.Remove(outputPath)
		return "", err
	}
	defer os.Remove(outputPath)

	ctx, cancel := context.WithTimeout(parent, r.timeout)
	defer cancel()
	cmd := exec.Command(r.codexCommand, codexArguments(outputPath)...)
	cmd.Dir = workDir
	cmd.Env = filteredEnvironment(r.environ())
	cmd.Stdin = strings.NewReader(titlePrompt(request))
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return "", err
		}
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
		return "", ctx.Err()
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", err
	}
	title := sanitizeTitle(string(data))
	if title == "" {
		return "", errEmptyTitle
	}
	return title, nil
}

func (r execCommandRunner) RunHerdr(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.Command(r.herdrCommand, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return output.Bytes(), err
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
		return output.Bytes(), ctx.Err()
	}
}

type application struct {
	cacheDir    string
	projectsDir string
	buildHash   string
	executable  string
	commands    commandRunner
	dial        func(string, string, time.Duration) (net.Conn, error)
	spawn       func(string, string) error
	now         func() time.Time
	sleep       func(time.Duration)
	newTimer    func(time.Duration) clockTimer
	newTicker   func(time.Duration) clockTicker
	stderr      io.Writer
	config      runtimeConfig
}

func newApplication() (*application, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	hash, err := hashExecutable(executable)
	if err != nil {
		return nil, err
	}
	config := defaultRuntimeConfig()
	cacheDir := filepath.Join(home, ".cache", stateDirName)
	return &application{
		cacheDir:    cacheDir,
		projectsDir: filepath.Join(home, ".claude", "projects"),
		buildHash:   hash,
		executable:  executable,
		commands: execCommandRunner{
			codexCommand: "codex",
			herdrCommand: "herdr",
			cacheDir:     cacheDir,
			environ:      os.Environ,
			timeout:      config.codexTimeout,
		},
		dial:  net.DialTimeout,
		spawn: startDetachedDaemon,
		now:   time.Now,
		sleep: time.Sleep,
		newTimer: func(duration time.Duration) clockTimer {
			return realClockTimer{timer: time.NewTimer(duration)}
		},
		newTicker: func(duration time.Duration) clockTicker {
			return realClockTicker{ticker: time.NewTicker(duration)}
		},
		stderr: os.Stderr,
		config: config,
	}, nil
}

func hashExecutable(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func startDetachedDaemon(executable, cacheDir string) error {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()
	stderr := io.Writer(devNull)
	daemonLog, logErr := openDaemonLog(cacheDir)
	if logErr == nil {
		defer daemonLog.Close()
		stderr = daemonLog
	}
	cmd := exec.Command(executable, "--daemon")
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func openDaemonLog(cacheDir string) (*os.File, error) {
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(cacheDir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(cacheDir, daemonLogFileName)
	// O_NOFOLLOW is authoritative. Lstat only improves diagnostics; O_NOFOLLOW
	// makes its TOCTOU window harmless. Removing O_NOFOLLOW reopens traversal.
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("daemon log must not be a symlink")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	closeOnError := func(err error) (*os.File, error) {
		_ = file.Close()
		return nil, err
	}
	info, err = file.Stat()
	if err != nil {
		return closeOnError(err)
	}
	if !info.Mode().IsRegular() {
		return closeOnError(errors.New("daemon log is not a regular file"))
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Nlink != 1 {
		return closeOnError(errors.New("daemon log must have exactly one link"))
	}
	if err := file.Chmod(0o600); err != nil {
		return closeOnError(err)
	}
	if info.Size() > maxDaemonLogSize {
		if err := file.Truncate(0); err != nil {
			return closeOnError(err)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return closeOnError(err)
		}
	}
	return file, nil
}

func filteredEnvironment(environment []string) []string {
	result := make([]string, 0, len(environment))
	for _, item := range environment {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "HERDR_") || strings.HasPrefix(key, "CLAUDE_CODE_") {
			continue
		}
		result = append(result, item)
	}
	return result
}

func buildDaemonEvent(input hookInput, tabID, workspaceID, buildHash string) (daemonEvent, bool) {
	if input.AgentID != nil {
		return daemonEvent{}, false
	}
	if len([]rune(tabID)) > maxTabIDRunes || len([]rune(workspaceID)) > maxWorkspaceIDRunes {
		return daemonEvent{}, false
	}
	kinds := make([]actorKind, 0, 2)
	text := ""
	switch input.HookEventName {
	case "UserPromptSubmit":
		text = strings.TrimSpace(input.Prompt)
		if tabID != "" {
			kinds = append(kinds, tabKind)
		}
		if workspaceID != "" {
			kinds = append(kinds, workspaceKind)
		}
	case "PreToolUse", "PostToolUse":
		if tabID == "" {
			return daemonEvent{}, false
		}
		kinds = append(kinds, tabKind)
		text = summarizeToolInput(input.ToolName, input.ToolInput)
	default:
		return daemonEvent{}, false
	}
	if len(kinds) == 0 {
		return daemonEvent{}, false
	}
	text = truncateRunes(text, maxInputRunes)
	return daemonEvent{
		ProtocolVersion: protocolVersion,
		BuildHash:       buildHash,
		Kinds:           kinds,
		Input: pendingInput{
			Text:           text,
			TranscriptPath: truncateRunes(input.TranscriptPath, maxTranscriptRunes),
			TabID:          tabID,
			WorkspaceID:    workspaceID,
			HookEventName:  input.HookEventName,
			ToolName:       truncateRunes(input.ToolName, maxToolNameRunes),
		},
	}, true
}

func summarizeToolInput(toolName string, input map[string]any) string {
	field := ""
	switch toolName {
	case "Bash":
		field = "description"
	case "Read", "Write", "Edit":
		field = "file_path"
	case "Grep", "Glob":
		field = "path"
	}
	value, _ := input[field].(string)
	value = redactToolValue(value)
	if value == "" {
		return toolName
	}
	return toolName + ": " + value
}

func redactToolValue(value string) string {
	value = bearerPattern.ReplaceAllString(value, "[REDACTED]")
	value = knownTokenPattern.ReplaceAllString(value, "[REDACTED]")
	value = longHexPattern.ReplaceAllString(value, "[REDACTED]")
	value = longBase64Pattern.ReplaceAllString(value, "[REDACTED]")
	return truncateRunes(strings.TrimSpace(value), maxToolInputRunes)
}

type deliveryResult int

const (
	deliveryFailed deliveryResult = iota
	deliveryACK
	deliveryNACK
)

func (a *application) runHook(reader io.Reader, tabID, workspaceID string) error {
	input, accepted, err := decodeHookInput(reader, a.stderr)
	if err != nil {
		return err
	}
	if !accepted {
		return nil
	}
	event, ok := buildDaemonEvent(input, tabID, workspaceID, a.buildHash)
	if !ok {
		return nil
	}
	frame, err := json.Marshal(event)
	if err != nil {
		return err
	}
	frame = append(frame, '\n')
	if len(frame) > maxFrameSize {
		_, _ = io.WriteString(a.stderr, "herdr-title: hook event exceeds IPC frame limit\n")
		return errors.New("hook event exceeds maximum frame size")
	}

	deadline := a.now().Add(a.config.hookLatencyLimit)
	result := a.deliver(frame, deadline)
	if result == deliveryACK {
		return nil
	}
	if result == deliveryNACK {
		a.sleepWithinDeadline(a.config.handoverDelay, deadline)
	}
	_ = a.spawn(a.executable, a.cacheDir)
	if result != deliveryNACK {
		a.sleepWithinDeadline(a.config.spawnSettle, deadline)
	}
	_ = a.deliver(frame, deadline)
	return nil
}

func decodeHookInput(reader io.Reader, stderr io.Writer) (hookInput, bool, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxHookInputSize+1))
	if err != nil {
		return hookInput{}, false, err
	}
	if len(data) > maxHookInputSize {
		_, _ = io.WriteString(stderr, "herdr-title: hook input too large\n")
		return hookInput{}, false, nil
	}
	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return hookInput{}, false, err
	}
	return input, true, nil
}

func (a *application) sleepWithinDeadline(duration time.Duration, deadline time.Time) {
	remaining := deadline.Sub(a.now())
	if remaining <= 0 {
		return
	}
	if duration > remaining {
		duration = remaining
	}
	a.sleep(duration)
}

func (a *application) deliver(frame []byte, overallDeadline time.Time) deliveryResult {
	remaining := overallDeadline.Sub(a.now())
	if remaining <= 0 {
		return deliveryFailed
	}
	timeout := a.config.hookACKTimeout
	if timeout > remaining {
		timeout = remaining
	}
	conn, err := a.dial("unix", filepath.Join(a.cacheDir, socketFileName), timeout)
	if err != nil {
		return deliveryFailed
	}
	defer conn.Close()
	deadline := a.now().Add(timeout)
	_ = conn.SetDeadline(deadline)
	if _, err := conn.Write(frame); err != nil {
		return deliveryFailed
	}
	response := []byte{0}
	if _, err := io.ReadFull(conn, response); err != nil {
		return deliveryFailed
	}
	switch response[0] {
	case ackByte:
		return deliveryACK
	case nackByte:
		return deliveryNACK
	default:
		return deliveryFailed
	}
}

type daemonLock struct {
	file *os.File
}

func acquireDaemonLock(cacheDir string) (*daemonLock, error) {
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(cacheDir, 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(cacheDir, lockFileName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, errAlreadyRunning
		}
		return nil, err
	}
	return &daemonLock{file: file}, nil
}

func acquireDaemonLockUntil(
	cacheDir string,
	deadline time.Time,
	now func() time.Time,
	sleep func(time.Duration),
	retryInterval time.Duration,
	stderr io.Writer,
	acquire func(string) (*daemonLock, error),
) (*daemonLock, error) {
	for {
		lock, err := acquire(cacheDir)
		if err == nil {
			return lock, nil
		}
		if !errors.Is(err, errAlreadyRunning) {
			return nil, err
		}
		remaining := deadline.Sub(now())
		if remaining <= 0 {
			_, _ = io.WriteString(stderr, "herdr-title: daemon lock deadline exceeded\n")
			return nil, context.DeadlineExceeded
		}
		delay := retryInterval
		if delay > remaining {
			delay = remaining
		}
		sleep(delay)
	}
}

func (l *daemonLock) release() {
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func prepareSocketPath(path string) error {
	return prepareSocketPathWith(path, func(path string) (os.FileMode, error) {
		info, err := os.Lstat(path)
		if err != nil {
			return 0, err
		}
		return info.Mode(), nil
	}, os.Remove)
}

func prepareSocketPathWith(
	path string,
	lstat func(string) (os.FileMode, error),
	remove func(string) error,
) error {
	mode, err := lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if mode&os.ModeSocket == 0 {
		return fmt.Errorf("refusing to remove non-socket at %s", path)
	}
	return remove(path)
}

type persistRequest struct {
	key   actorKey
	get   bool
	flush bool
	delta stateDelta
	reply chan persistResponse
}

type persistResponse struct {
	state persistedActor
	found bool
	err   error
}

type persister struct {
	requests chan persistRequest
	done     chan struct{}
}

func startPersister(ctx context.Context, cacheDir string) (*persister, error) {
	state, err := loadDiskState(filepath.Join(cacheDir, stateFileName))
	if err != nil {
		return nil, err
	}
	p := &persister{requests: make(chan persistRequest), done: make(chan struct{})}
	go p.run(ctx, cacheDir, state)
	return p, nil
}

func (p *persister) get(ctx context.Context, key actorKey) (persistedActor, bool, error) {
	reply := make(chan persistResponse, 1)
	request := persistRequest{key: key, get: true, reply: reply}
	select {
	case p.requests <- request:
	case <-ctx.Done():
		return persistedActor{}, false, ctx.Err()
	}
	select {
	case response := <-reply:
		return response.state, response.found, response.err
	case <-ctx.Done():
		return persistedActor{}, false, ctx.Err()
	}
}

func (p *persister) put(ctx context.Context, key actorKey, delta stateDelta) error {
	reply := make(chan persistResponse, 1)
	request := persistRequest{key: key, delta: delta, reply: reply}
	select {
	case p.requests <- request:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case response := <-reply:
		return response.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *persister) flush(ctx context.Context) error {
	reply := make(chan persistResponse, 1)
	request := persistRequest{flush: true, reply: reply}
	select {
	case p.requests <- request:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case response := <-reply:
		return response.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *persister) run(ctx context.Context, cacheDir string, state diskState) {
	defer close(p.done)
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-p.requests:
			if request.flush {
				request.reply <- persistResponse{}
				continue
			}
			storageKey := request.key.storageKey()
			if request.get {
				entry, found := state.Actors[storageKey]
				request.reply <- persistResponse{state: entry, found: found}
				continue
			}
			old, existed := state.Actors[storageKey]
			entry := old
			if request.delta.Title != nil {
				entry.Title = *request.delta.Title
			}
			if request.delta.LastStarted != nil {
				entry.LastStarted = *request.delta.LastStarted
			}
			state.Actors[storageKey] = entry
			err := writeDiskState(cacheDir, state)
			if err != nil {
				if existed {
					state.Actors[storageKey] = old
				} else {
					delete(state.Actors, storageKey)
				}
			}
			request.reply <- persistResponse{err: err}
		}
	}
}

func loadDiskState(path string) (diskState, error) {
	state := diskState{ProtocolVersion: protocolVersion, Actors: make(map[string]persistedActor)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return diskState{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return state, nil
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return diskState{ProtocolVersion: protocolVersion, Actors: make(map[string]persistedActor)}, nil
	}
	if state.ProtocolVersion != protocolVersion || state.Actors == nil {
		return diskState{ProtocolVersion: protocolVersion, Actors: make(map[string]persistedActor)}, nil
	}
	return state, nil
}

func writeDiskState(cacheDir string, state diskState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(cacheDir, "state-*")
	if err != nil {
		return err
	}
	path := temporary.Name()
	ok := false
	defer func() {
		_ = temporary.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(path, filepath.Join(cacheDir, stateFileName)); err != nil {
		return err
	}
	ok = true
	directory, err := os.Open(cacheDir)
	if err == nil {
		err = directory.Sync()
		_ = directory.Close()
	}
	return err
}

type generationJob struct {
	actor         *titleActor
	input         pendingInput
	previousTitle string
	mode          jobMode
	title         string
}

type jobMode int

const (
	jobFresh jobMode = iota
	jobGuardRetry
	jobCodexRetry
	jobRenameRetry
)

type actorPhase string

const (
	phaseNone               actorPhase = "none"
	phaseQueuedGuard        actorPhase = "queued-guard"
	phaseRunningGuard       actorPhase = "running-guard"
	phaseQueuedCodex        actorPhase = "queued-codex"
	phaseRunningCodex       actorPhase = "running-codex"
	phaseQueuedRenameGuard  actorPhase = "queued-rename-guard"
	phaseRunningRenameGuard actorPhase = "running-rename-guard"
	phaseQueuedRename       actorPhase = "queued-rename"
	phaseRunningRename      actorPhase = "running-rename"
)

type jobPicked struct {
	job   generationJob
	phase actorPhase
}

type jobQueued struct {
	job   generationJob
	phase actorPhase
}

type jobStart struct {
	job     generationJob
	started time.Time
	reply   chan error
}

type resultPhase int

const (
	resultSuccess resultPhase = iota
	resultGuardBlocked
	resultGuardFailure
	resultRenameGuardSkipped
	resultRenameGuardFailure
	resultCodexFailure
	resultRenameFailure
)

type jobResult struct {
	job   generationJob
	phase resultPhase
	title string
	err   error
	when  time.Time
}

type actorEvent struct {
	input pendingInput
}

type actorStatus struct {
	key                 actorKey
	actor               *titleActor
	state               actorState
	phase               actorPhase
	guardRetries        int
	generationRetried   bool
	renameRetryConsumed bool
	pendingText         string
}

type evictionRequest struct {
	key   actorKey
	actor *titleActor
	reply chan bool
}

type renameRetry struct {
	input pendingInput
	title string
}

type titleActor struct {
	key                 actorKey
	state               actorState
	phase               actorPhase
	pending             *pendingInput
	active              *pendingInput
	title               string
	lastStarted         time.Time
	coldStart           bool
	generationRetried   bool
	guardRetryCount     int
	renameRetryConsumed bool
	renameRetry         *renameRetry
	timer               clockTimer
	timerChannel        <-chan time.Time
	timerMode           timerMode
	evictionTimer       clockTimer
	evictionChannel     <-chan time.Time
	inbox               chan any
	jobs                chan<- generationJob
	statuses            chan<- actorStatus
	evictions           chan<- evictionRequest
	persister           *persister
	stderr              io.Writer
	ctx                 context.Context
	now                 func() time.Time
	newTimer            func(time.Duration) clockTimer
	config              runtimeConfig
}

func newTitleActor(
	ctx context.Context,
	key actorKey,
	state persistedActor,
	found bool,
	jobs chan<- generationJob,
	statuses chan<- actorStatus,
	evictions chan<- evictionRequest,
	persister *persister,
	stderr io.Writer,
	now func() time.Time,
	newTimer func(time.Duration) clockTimer,
	config runtimeConfig,
) *titleActor {
	actor := &titleActor{
		key: key, state: stateIdle, phase: phaseNone, title: state.Title, coldStart: !found,
		jobs: jobs, statuses: statuses, evictions: evictions, persister: persister,
		stderr: stderr, ctx: ctx, now: now, newTimer: newTimer, config: config,
		inbox: make(chan any, config.actorInboxSize),
	}
	if state.LastStarted > 0 {
		actor.lastStarted = time.Unix(0, state.LastStarted)
	}
	return actor
}

func (a *titleActor) run() {
	a.publishState()
	for {
		select {
		case <-a.ctx.Done():
			a.stopTimers()
			return
		case message := <-a.inbox:
			switch value := message.(type) {
			case actorEvent:
				a.onEvent(value.input)
			case jobPicked:
				a.onJobPicked(value)
			case jobQueued:
				a.onJobQueued(value)
			case jobStart:
				a.onJobStart(value)
			case jobResult:
				a.onJobResult(value)
			}
		case <-a.timerChannel:
			a.timer = nil
			a.timerChannel = nil
			mode := a.timerMode
			a.timerMode = timerNone
			a.onTimer(mode)
		case <-a.evictionChannel:
			a.evictionTimer = nil
			a.evictionChannel = nil
			if a.requestEviction() {
				return
			}
		}
	}
}

func (a *titleActor) onEvent(input pendingInput) {
	copy := input
	a.pending = &copy
	a.stopEvictionTimer()
	if a.state != stateIdle {
		a.publishState()
		return
	}
	a.schedulePending(a.now())
}

func (a *titleActor) onJobPicked(message jobPicked) {
	if !a.matchesActiveJob(message.job) {
		return
	}
	a.changeState(stateRunning, message.phase)
}

func (a *titleActor) onJobQueued(message jobQueued) {
	if !a.matchesActiveJob(message.job) {
		return
	}
	a.changeState(stateQueued, message.phase)
}

func (a *titleActor) onJobStart(message jobStart) {
	if !a.matchesActiveJob(message.job) {
		message.reply <- errors.New("stale generation job")
		return
	}
	a.lastStarted = message.started
	started := message.started.UnixNano()
	err := a.persister.put(a.ctx, a.key, stateDelta{LastStarted: &started})
	message.reply <- err
}

func (a *titleActor) onJobResult(result jobResult) {
	if !a.matchesActiveJob(result.job) {
		return
	}
	switch result.phase {
	case resultGuardBlocked:
		a.recordLastStarted(result.when)
		a.abandonJob()
	case resultGuardFailure:
		a.guardRetryCount++
		if a.guardRetryCount >= 3 {
			a.abandonJob()
			return
		}
		a.scheduleAt(a.now().Add(a.config.guardRetry), timerGuardRetry, stateRetryPending)
	case resultCodexFailure:
		if a.generationRetried {
			a.reportCodexRetryExhaustion(result.err)
			a.abandonJob()
			return
		}
		a.generationRetried = true
		now := a.now()
		target := now.Add(a.config.retryDelay)
		throttled := a.lastStarted.Add(a.throttle())
		if throttled.After(target) {
			target = throttled
		}
		a.scheduleAt(target, timerCodexRetry, stateRetryPending)
	case resultRenameGuardSkipped:
		a.abandonJob()
	case resultRenameGuardFailure, resultRenameFailure:
		if a.renameRetryConsumed {
			a.abandonJob()
			return
		}
		a.renameRetryConsumed = true
		a.renameRetry = &renameRetry{input: result.job.input, title: result.title}
		a.scheduleAt(a.now().Add(a.config.retryDelay), timerRenameRetry, stateRetryPending)
	case resultSuccess:
		a.title = result.title
		_ = a.persister.put(a.ctx, a.key, stateDelta{Title: &a.title})
		a.active = nil
		a.renameRetry = nil
		a.generationRetried = false
		a.guardRetryCount = 0
		a.renameRetryConsumed = false
		a.transitionAfterSuccess()
	}
}

func (a *titleActor) reportCodexRetryExhaustion(err error) {
	if a.stderr == nil {
		return
	}
	_, _ = fmt.Fprintf(
		a.stderr,
		"herdr-title: codex retries exhausted actor=%s exit_status=%s\n",
		a.key.storageKey(), codexFailureExitStatus(err),
	)
}

func codexFailureExitStatus(err error) string {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return strconv.Itoa(exitError.ExitCode())
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, errEmptyTitle) {
		return "empty"
	}
	return "unavailable"
}

func (a *titleActor) recordLastStarted(started time.Time) {
	if started.IsZero() {
		started = a.now()
	}
	a.lastStarted = started
	value := started.UnixNano()
	_ = a.persister.put(a.ctx, a.key, stateDelta{LastStarted: &value})
}

func (a *titleActor) abandonJob() {
	a.stopTimer()
	a.pending = nil
	a.active = nil
	a.renameRetry = nil
	a.guardRetryCount = 0
	a.generationRetried = false
	a.renameRetryConsumed = false
	a.becomeIdle()
}

func (a *titleActor) matchesActiveJob(job generationJob) bool {
	return a.active != nil && *a.active == job.input
}

func (a *titleActor) onTimer(mode timerMode) {
	switch mode {
	case timerGeneration:
		a.enqueueFreshGeneration()
	case timerGuardRetry:
		a.enqueueActiveRetry(jobGuardRetry)
	case timerCodexRetry:
		a.enqueueActiveRetry(jobCodexRetry)
	case timerRenameRetry:
		a.enqueueRenameRetry()
	}
}

func (a *titleActor) enqueueFreshGeneration() {
	if a.pending == nil {
		a.becomeIdle()
		return
	}
	input := *a.pending
	a.pending = nil
	a.active = &input
	a.guardRetryCount = 0
	a.generationRetried = false
	a.renameRetryConsumed = false
	a.renameRetry = nil
	job := generationJob{actor: a, input: input, previousTitle: a.title, mode: jobFresh}
	a.changeState(stateQueued, initialPhase(job))
	select {
	case a.jobs <- job:
	case <-a.ctx.Done():
	}
}

func (a *titleActor) enqueueActiveRetry(mode jobMode) {
	if a.active == nil {
		a.abandonJob()
		return
	}
	job := generationJob{actor: a, input: *a.active, previousTitle: a.title, mode: mode}
	a.changeState(stateQueued, initialPhase(job))
	select {
	case a.jobs <- job:
	case <-a.ctx.Done():
	}
}

func (a *titleActor) enqueueRenameRetry() {
	if a.renameRetry == nil {
		a.abandonJob()
		return
	}
	retry := *a.renameRetry
	if a.pending != nil && *a.pending != retry.input {
		a.active = nil
		a.renameRetry = nil
		a.schedulePending(a.now())
		return
	}
	if a.pending != nil {
		a.pending = nil
	}
	if a.active == nil {
		input := retry.input
		a.active = &input
	}
	job := generationJob{actor: a, input: retry.input, previousTitle: a.title, mode: jobRenameRetry, title: retry.title}
	a.changeState(stateQueued, initialPhase(job))
	select {
	case a.jobs <- job:
	case <-a.ctx.Done():
	}
}

func (a *titleActor) schedulePending(now time.Time) {
	if a.pending == nil {
		a.becomeIdle()
		return
	}
	target := now
	throttled := a.lastStarted.Add(a.throttle())
	if throttled.After(target) {
		target = throttled
	}
	if !target.After(now) {
		a.enqueueFreshGeneration()
		return
	}
	a.scheduleAt(target, timerGeneration, stateScheduled)
}

func initialPhase(job generationJob) actorPhase {
	switch job.mode {
	case jobGuardRetry:
		return phaseQueuedGuard
	case jobCodexRetry:
		return phaseQueuedCodex
	case jobRenameRetry:
		if job.actor.key.Kind == workspaceKind {
			return phaseQueuedRenameGuard
		}
		return phaseQueuedRename
	default:
		if job.actor.key.Kind == workspaceKind {
			return phaseQueuedGuard
		}
		return phaseQueuedCodex
	}
}

func (a *titleActor) scheduleAt(target time.Time, mode timerMode, state actorState) {
	a.stopTimer()
	delay := target.Sub(a.now())
	if delay < 0 {
		delay = 0
	}
	a.timer = a.newTimer(delay)
	a.timerChannel = a.timer.Chan()
	a.timerMode = mode
	a.changeState(state, phaseNone)
}

func (a *titleActor) transitionAfterSuccess() {
	if a.pending != nil {
		a.schedulePending(a.now())
		return
	}
	a.becomeIdle()
}

func (a *titleActor) becomeIdle() {
	a.stopTimer()
	a.changeState(stateIdle, phaseNone)
	a.stopEvictionTimer()
	a.evictionTimer = a.newTimer(a.config.actorEviction)
	a.evictionChannel = a.evictionTimer.Chan()
}

func (a *titleActor) requestEviction() bool {
	if a.state != stateIdle || a.pending != nil || a.timerMode != timerNone {
		a.becomeIdle()
		return false
	}
	started := a.lastStarted.UnixNano()
	if a.lastStarted.IsZero() {
		started = 0
	}
	if err := a.persister.put(a.ctx, a.key, stateDelta{Title: &a.title, LastStarted: &started}); err != nil {
		a.becomeIdle()
		return false
	}
	reply := make(chan bool, 1)
	request := evictionRequest{key: a.key, actor: a, reply: reply}
	select {
	case a.evictions <- request:
	case <-a.ctx.Done():
		return false
	}
	select {
	case evicted := <-reply:
		if evicted {
			a.stopTimers()
			return true
		}
		a.becomeIdle()
	case <-a.ctx.Done():
	}
	return false
}

func (a *titleActor) throttle() time.Duration {
	if a.key.Kind == workspaceKind {
		return a.config.workspaceThrottle
	}
	return a.config.tabThrottle
}

func (a *titleActor) changeState(state actorState, phase actorPhase) {
	a.state = state
	a.phase = phase
	a.publishState()
}

func (a *titleActor) publishState() {
	if a.statuses == nil {
		return
	}
	select {
	case a.statuses <- actorStatus{
		key: a.key, actor: a, state: a.state, phase: a.phase,
		guardRetries:        a.guardRetryCount,
		generationRetried:   a.generationRetried,
		renameRetryConsumed: a.renameRetryConsumed,
		pendingText:         pendingText(a.pending),
	}:
	case <-a.ctx.Done():
	}
}

func pendingText(input *pendingInput) string {
	if input == nil {
		return ""
	}
	return input.Text
}

func (a *titleActor) stopTimer() {
	if a.timer != nil {
		if !a.timer.Stop() {
			select {
			case <-a.timerChannel:
			default:
			}
		}
	}
	a.timer = nil
	a.timerChannel = nil
	a.timerMode = timerNone
}

func (a *titleActor) stopEvictionTimer() {
	if a.evictionTimer != nil {
		if !a.evictionTimer.Stop() {
			select {
			case <-a.evictionChannel:
			default:
			}
		}
	}
	a.evictionTimer = nil
	a.evictionChannel = nil
}

func (a *titleActor) stopTimers() {
	a.stopTimer()
	a.stopEvictionTimer()
}

type dispatchRequest struct {
	event daemonEvent
	reply chan error
}

type inspectRequest struct {
	reply chan runtimeSnapshot
}

type runtimeSnapshot struct {
	States              map[actorKey]actorState
	Phases              map[actorKey]actorPhase
	ColdStarts          map[actorKey]bool
	GuardRetries        map[actorKey]int
	GenerationRetried   map[actorKey]bool
	RenameRetryConsumed map[actorKey]bool
	PendingText         map[actorKey]string
}

type semaphoreWaiter struct {
	ready    chan struct{}
	granted  bool
	canceled bool
}

type fifoSemaphore struct {
	mu       sync.Mutex
	capacity int
	inUse    int
	waiters  []*semaphoreWaiter
}

func newFIFOSemaphore(capacity int) *fifoSemaphore {
	return &fifoSemaphore{capacity: capacity}
}

func (s *fifoSemaphore) Acquire(ctx context.Context) error {
	s.mu.Lock()
	if s.inUse < s.capacity && len(s.waiters) == 0 {
		s.inUse++
		s.mu.Unlock()
		return nil
	}
	waiter := &semaphoreWaiter{ready: make(chan struct{})}
	s.waiters = append(s.waiters, waiter)
	s.mu.Unlock()

	select {
	case <-waiter.ready:
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		if waiter.granted {
			s.mu.Unlock()
			return nil
		}
		waiter.canceled = true
		for index, queued := range s.waiters {
			if queued == waiter {
				s.waiters = append(s.waiters[:index], s.waiters[index+1:]...)
				break
			}
		}
		s.mu.Unlock()
		return ctx.Err()
	}
}

func (s *fifoSemaphore) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.waiters) > 0 {
		waiter := s.waiters[0]
		s.waiters = s.waiters[1:]
		if waiter.canceled {
			continue
		}
		waiter.granted = true
		close(waiter.ready)
		return
	}
	if s.inUse <= 0 {
		panic("herdr-title semaphore released without acquisition")
	}
	s.inUse--
}

func (s *fifoSemaphore) Stats() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inUse, len(s.waiters)
}

type semaphoreLease struct {
	codex bool
	herdr bool
}

func (l *semaphoreLease) acquireCodex(ctx context.Context, semaphore *fifoSemaphore) error {
	if l.herdr {
		panic("herdr-title attempted to hold codexSem and herdrSem together")
	}
	if err := semaphore.Acquire(ctx); err != nil {
		return err
	}
	l.codex = true
	return nil
}

func (l *semaphoreLease) releaseCodex(semaphore *fifoSemaphore) {
	if !l.codex {
		panic("herdr-title released an unheld codexSem")
	}
	l.codex = false
	semaphore.Release()
}

func (l *semaphoreLease) acquireHerdr(ctx context.Context, semaphore *fifoSemaphore) error {
	if l.codex {
		panic("herdr-title attempted to hold herdrSem and codexSem together")
	}
	if err := semaphore.Acquire(ctx); err != nil {
		return err
	}
	l.herdr = true
	return nil
}

func (l *semaphoreLease) releaseHerdr(semaphore *fifoSemaphore) {
	if !l.herdr {
		panic("herdr-title released an unheld herdrSem")
	}
	l.herdr = false
	semaphore.Release()
}

func (l *semaphoreLease) assertReleased() {
	if l.codex || l.herdr {
		panic("herdr-title job ended while holding a semaphore")
	}
}

type daemonRuntime struct {
	app       *application
	persister *persister
	ctx       context.Context
	cancel    context.CancelFunc
	dispatch  chan dispatchRequest
	inspect   chan inspectRequest
	shutdown  chan struct{}
	codexSem  *fifoSemaphore
	herdrSem  *fifoSemaphore
	stopOnce  sync.Once
	waitGroup sync.WaitGroup
}

func startDaemonRuntime(app *application) (*daemonRuntime, error) {
	if err := os.MkdirAll(app.cacheDir, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(app.cacheDir, 0o700); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	persister, err := startPersister(ctx, app.cacheDir)
	if err != nil {
		cancel()
		return nil, err
	}
	runtime := &daemonRuntime{
		app: app, persister: persister, ctx: ctx, cancel: cancel,
		dispatch: make(chan dispatchRequest), inspect: make(chan inspectRequest),
		shutdown: make(chan struct{}), codexSem: newFIFOSemaphore(2), herdrSem: newFIFOSemaphore(4),
	}
	jobs := make(chan generationJob, app.config.queueSize)
	statuses := make(chan actorStatus, app.config.queueSize)
	evictions := make(chan evictionRequest, app.config.queueSize)
	runtime.waitGroup.Add(1)
	go func() {
		defer runtime.waitGroup.Done()
		runtime.launchJobs(jobs)
	}()
	runtime.waitGroup.Add(1)
	go func() {
		defer runtime.waitGroup.Done()
		runtime.dispatcher(persister, jobs, statuses, evictions)
	}()
	return runtime, nil
}

func (r *daemonRuntime) Dispatch(event daemonEvent) error {
	reply := make(chan error, 1)
	select {
	case r.dispatch <- dispatchRequest{event: event, reply: reply}:
	case <-r.ctx.Done():
		return r.ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-r.ctx.Done():
		return r.ctx.Err()
	}
}

func (r *daemonRuntime) Snapshot() runtimeSnapshot {
	reply := make(chan runtimeSnapshot, 1)
	select {
	case r.inspect <- inspectRequest{reply: reply}:
	case <-r.ctx.Done():
		return runtimeSnapshot{}
	}
	select {
	case snapshot := <-reply:
		return snapshot
	case <-r.ctx.Done():
		return runtimeSnapshot{}
	}
}

func (r *daemonRuntime) Stop() {
	r.stopOnce.Do(func() { r.cancel() })
	r.waitGroup.Wait()
	<-r.persister.done
}

func (r *daemonRuntime) requestShutdown() {
	r.stopOnce.Do(func() {
		close(r.shutdown)
		r.cancel()
	})
}

func (r *daemonRuntime) dispatcher(
	persister *persister,
	jobs chan<- generationJob,
	statuses chan actorStatus,
	evictions chan evictionRequest,
) {
	actors := make(map[actorKey]*titleActor)
	states := make(map[actorKey]actorState)
	phases := make(map[actorKey]actorPhase)
	guardRetries := make(map[actorKey]int)
	generationRetried := make(map[actorKey]bool)
	renameRetryConsumed := make(map[actorKey]bool)
	pendingTexts := make(map[actorKey]string)
	lastEvent := r.app.now()
	ticker := r.app.newTicker(r.app.config.idlePoll)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case request := <-r.dispatch:
			lastEvent = r.app.now()
			var dispatchErr error
			for _, kind := range request.event.Kinds {
				key := actorKey{Kind: kind, WorkspaceID: request.event.Input.WorkspaceID}
				if kind == tabKind {
					key.TabID = request.event.Input.TabID
				}
				actor := actors[key]
				if actor == nil {
					persisted, found, err := persister.get(r.ctx, key)
					if err != nil {
						dispatchErr = err
						break
					}
					actor = newTitleActor(
						r.ctx, key, persisted, found, jobs, statuses, evictions,
						persister, r.app.stderr, r.app.now, r.app.newTimer, r.app.config,
					)
					actors[key] = actor
					states[key] = stateIdle
					phases[key] = phaseNone
					r.waitGroup.Add(1)
					go func(actor *titleActor) {
						defer r.waitGroup.Done()
						actor.run()
					}(actor)
				}
				select {
				case actor.inbox <- actorEvent{input: request.event.Input}:
				case <-r.ctx.Done():
					dispatchErr = r.ctx.Err()
				}
			}
			request.reply <- dispatchErr
		case status := <-statuses:
			if actors[status.key] == status.actor {
				states[status.key] = status.state
				phases[status.key] = status.phase
				guardRetries[status.key] = status.guardRetries
				generationRetried[status.key] = status.generationRetried
				renameRetryConsumed[status.key] = status.renameRetryConsumed
				pendingTexts[status.key] = status.pendingText
			}
		case request := <-evictions:
			actor := actors[request.key]
			if actor != request.actor {
				request.reply <- false
				continue
			}
			idle := states[request.key] == stateIdle && len(actor.inbox) == 0
			if idle {
				delete(actors, request.key)
				delete(states, request.key)
				delete(phases, request.key)
				delete(guardRetries, request.key)
				delete(generationRetried, request.key)
				delete(renameRetryConsumed, request.key)
				delete(pendingTexts, request.key)
			}
			request.reply <- idle
		case request := <-r.inspect:
			snapshot := runtimeSnapshot{
				States:              make(map[actorKey]actorState, len(states)),
				Phases:              make(map[actorKey]actorPhase, len(phases)),
				ColdStarts:          make(map[actorKey]bool, len(actors)),
				GuardRetries:        make(map[actorKey]int, len(guardRetries)),
				GenerationRetried:   make(map[actorKey]bool, len(generationRetried)),
				RenameRetryConsumed: make(map[actorKey]bool, len(renameRetryConsumed)),
				PendingText:         make(map[actorKey]string, len(pendingTexts)),
			}
			for key, state := range states {
				snapshot.States[key] = state
			}
			for key, actor := range actors {
				snapshot.ColdStarts[key] = actor.coldStart
			}
			for key, phase := range phases {
				snapshot.Phases[key] = phase
			}
			for key, count := range guardRetries {
				snapshot.GuardRetries[key] = count
			}
			for key, value := range generationRetried {
				snapshot.GenerationRetried[key] = value
			}
			for key, value := range renameRetryConsumed {
				snapshot.RenameRetryConsumed[key] = value
			}
			for key, value := range pendingTexts {
				snapshot.PendingText[key] = value
			}
			request.reply <- snapshot
		case <-ticker.Chan():
			if r.app.now().Sub(lastEvent) >= r.app.config.daemonIdle && allActorsIdle(states) {
				if err := persister.flush(r.ctx); err != nil {
					continue
				}
				r.requestShutdown()
				return
			}
		}
	}
}

func allActorsIdle(states map[actorKey]actorState) bool {
	for _, state := range states {
		if state != stateIdle {
			return false
		}
	}
	return true
}

func (r *daemonRuntime) launchJobs(jobs <-chan generationJob) {
	for {
		select {
		case <-r.ctx.Done():
			return
		case job := <-jobs:
			r.waitGroup.Add(1)
			go func(job generationJob) {
				defer r.waitGroup.Done()
				r.executeJob(job)
			}(job)
		}
	}
}

func (r *daemonRuntime) executeJob(job generationJob) {
	lease := &semaphoreLease{}
	defer lease.assertReleased()
	if job.mode == jobRenameRetry {
		r.executeRename(job, job.title, lease)
		return
	}
	if job.actor.key.Kind == workspaceKind && job.mode != jobCodexRetry {
		single, err := r.executeGuardCall(job, phaseRunningGuard, lease)
		if err != nil {
			r.sendActor(job.actor, jobResult{job: job, phase: resultGuardFailure, err: err})
			return
		}
		if !single {
			r.sendActor(job.actor, jobResult{job: job, phase: resultGuardBlocked, when: r.app.now()})
			return
		}
		r.sendActor(job.actor, jobQueued{job: job, phase: phaseQueuedCodex})
	}
	r.executeCodex(job, lease)
}

func (r *daemonRuntime) executeCodex(job generationJob, lease *semaphoreLease) {
	if err := lease.acquireCodex(r.ctx, r.codexSem); err != nil {
		return
	}
	r.sendActor(job.actor, jobPicked{job: job, phase: phaseRunningCodex})
	started := r.app.now()
	startReply := make(chan error, 1)
	r.sendActor(job.actor, jobStart{job: job, started: started, reply: startReply})
	select {
	case err := <-startReply:
		if err != nil {
			lease.releaseCodex(r.codexSem)
			r.sendActor(job.actor, jobResult{job: job, phase: resultCodexFailure, err: err})
			return
		}
	case <-r.ctx.Done():
		lease.releaseCodex(r.codexSem)
		return
	}
	title, err := r.app.commands.GenerateTitle(r.ctx, titleRequest{
		Kind: job.actor.key.Kind, Input: job.input, Title: job.previousTitle,
		ProjectsDir: r.app.projectsDir,
	})
	lease.releaseCodex(r.codexSem)
	title = sanitizeTitle(title)
	if err != nil || title == "" {
		if err == nil {
			err = errEmptyTitle
		}
		r.sendActor(job.actor, jobResult{job: job, phase: resultCodexFailure, err: err})
		return
	}
	job.title = title
	if job.actor.key.Kind == workspaceKind {
		r.sendActor(job.actor, jobQueued{job: job, phase: phaseQueuedRenameGuard})
	} else {
		r.sendActor(job.actor, jobQueued{job: job, phase: phaseQueuedRename})
	}
	r.executeRename(job, title, lease)
}

func (r *daemonRuntime) executeRename(job generationJob, title string, lease *semaphoreLease) {
	if job.actor.key.Kind == workspaceKind {
		single, err := r.executeGuardCall(job, phaseRunningRenameGuard, lease)
		if err != nil {
			r.sendActor(job.actor, jobResult{job: job, phase: resultRenameGuardFailure, title: title, err: err})
			return
		}
		if !single {
			r.sendActor(job.actor, jobResult{job: job, phase: resultRenameGuardSkipped, title: title})
			return
		}
		r.sendActor(job.actor, jobQueued{job: job, phase: phaseQueuedRename})
	}
	if err := lease.acquireHerdr(r.ctx, r.herdrSem); err != nil {
		return
	}
	r.sendActor(job.actor, jobPicked{job: job, phase: phaseRunningRename})
	ctx, cancel := context.WithTimeout(r.ctx, r.app.config.herdrTimeout)
	var err error
	if job.actor.key.Kind == tabKind {
		_, err = r.app.commands.RunHerdr(ctx, "tab", "rename", job.actor.key.TabID, title)
	} else {
		_, err = r.app.commands.RunHerdr(ctx, "workspace", "rename", job.actor.key.WorkspaceID, title)
	}
	cancel()
	lease.releaseHerdr(r.herdrSem)
	if err != nil {
		r.sendActor(job.actor, jobResult{job: job, phase: resultRenameFailure, title: title, err: err})
		return
	}
	r.sendActor(job.actor, jobResult{job: job, phase: resultSuccess, title: title})
}

func (r *daemonRuntime) executeGuardCall(job generationJob, phase actorPhase, lease *semaphoreLease) (bool, error) {
	if err := lease.acquireHerdr(r.ctx, r.herdrSem); err != nil {
		return false, err
	}
	r.sendActor(job.actor, jobPicked{job: job, phase: phase})
	single, err := r.singleTabGuard(job.actor.key.WorkspaceID)
	lease.releaseHerdr(r.herdrSem)
	return single, err
}

func (r *daemonRuntime) sendActor(actor *titleActor, message any) {
	select {
	case actor.inbox <- message:
	case <-r.ctx.Done():
	}
}

func (r *daemonRuntime) singleTabGuard(workspaceID string) (bool, error) {
	ctx, cancel := context.WithTimeout(r.ctx, r.app.config.herdrTimeout)
	defer cancel()
	data, err := r.app.commands.RunHerdr(ctx, "tab", "list")
	if err != nil {
		return false, err
	}
	tabs, err := parseTabList(data)
	if err != nil {
		return false, err
	}
	count := 0
	for _, tab := range tabs {
		if tab.WorkspaceID == workspaceID {
			count++
		}
	}
	return count == 1, nil
}

type tabInfo struct {
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
}

func parseTabList(data []byte) ([]tabInfo, error) {
	var tabs []tabInfo
	if err := json.Unmarshal(data, &tabs); err == nil {
		return tabs, nil
	}
	var response struct {
		Tabs   *[]tabInfo `json:"tabs"`
		Result *struct {
			Tabs *[]tabInfo `json:"tabs"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	if response.Result != nil && response.Result.Tabs != nil {
		return *response.Result.Tabs, nil
	}
	if response.Tabs != nil {
		return *response.Tabs, nil
	}
	return nil, errors.New("tab list response does not contain tabs")
}

func (a *application) runDaemon() error {
	lock, err := acquireDaemonLockUntil(
		a.cacheDir,
		a.now().Add(a.config.hookLatencyLimit),
		a.now,
		a.sleep,
		a.config.lockRetryInterval,
		a.stderr,
		acquireDaemonLock,
	)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	if err != nil {
		return err
	}
	defer lock.release()

	socketPath := filepath.Join(a.cacheDir, socketFileName)
	if err := prepareSocketPath(socketPath); err != nil {
		return err
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()
	if err := os.Chmod(socketPath, 0o600); err != nil {
		return err
	}
	runtime, err := startDaemonRuntime(a)
	if err != nil {
		return err
	}
	defer runtime.Stop()
	go func() {
		<-runtime.shutdown
		_ = listener.Close()
	}()
	for {
		connection, err := listener.Accept()
		if err != nil {
			select {
			case <-runtime.shutdown:
				return nil
			default:
				return err
			}
		}
		go a.handleConnection(connection, runtime)
	}
}

func (a *application) handleConnection(connection net.Conn, runtime *daemonRuntime) {
	defer connection.Close()
	_ = connection.SetDeadline(a.now().Add(a.config.ipcDeadline))
	reader := bufio.NewReaderSize(connection, maxFrameSize)
	frame, err := reader.ReadSlice('\n')
	if err != nil || len(frame) > maxFrameSize {
		_, _ = connection.Write([]byte{nackByte})
		return
	}
	frame = bytes.TrimSuffix(frame, []byte{'\n'})
	var event daemonEvent
	if json.Unmarshal(frame, &event) != nil || event.validate() != nil {
		_, _ = connection.Write([]byte{nackByte})
		return
	}
	if event.ProtocolVersion != protocolVersion || event.BuildHash != a.buildHash {
		_, _ = connection.Write([]byte{nackByte})
		runtime.requestShutdown()
		return
	}
	if err := runtime.Dispatch(event); err != nil {
		_, _ = connection.Write([]byte{nackByte})
		return
	}
	_, _ = connection.Write([]byte{ackByte})
}

func titlePrompt(request titleRequest) string {
	inputs := buildTitleInputs(request.ProjectsDir, request.Input.TranscriptPath, request.Input.Text)
	var instruction string
	if request.Kind == workspaceKind {
		instruction = "以下の発話から、セッション全体の大枠のtaskを表す日本語の名詞句を1つ作成してください。"
	} else {
		instruction = "以下の発話とtool要約から、現在進めている作業を表す日本語タイトルを1つ作成してください。動詞句でもかまいません。"
	}
	if request.Title != "" {
		instruction += "\n現在のタイトルは「" + request.Title + "」です。今も内容を正確に表すなら、そのまま出力してかまいません。"
	}
	return instruction + "\n60文字以内とし、改行、引用符、装飾を含めず、タイトルだけを出力してください。#はissue番号に使用できます。\n\n" + strings.Join(inputs, "\n")
}

func buildTitleInputs(projectsDir, transcriptPath, current string) []string {
	entries := readTranscriptEntries(projectsDir, transcriptPath)
	history := extractUserMessages(entries, current)
	if len(history) > firstHistoryEntries+recentHistoryEntries {
		history = append(
			append([]string(nil), history[:firstHistoryEntries]...),
			history[len(history)-recentHistoryEntries:]...,
		)
	}
	inputs := make([]string, 0, len(history)+1)
	for _, input := range history {
		if input = truncateRunes(strings.TrimSpace(input), maxInputRunes); input != "" {
			inputs = append(inputs, input)
		}
	}
	if current = truncateRunes(strings.TrimSpace(current), maxInputRunes); current != "" {
		inputs = append(inputs, current)
	}
	return inputs
}

type transcriptEntry struct {
	Type        string          `json:"type"`
	Message     json.RawMessage `json:"message"`
	IsMeta      bool            `json:"isMeta"`
	IsSynthetic bool            `json:"isSynthetic"`
	Meta        bool            `json:"meta"`
	Synthetic   bool            `json:"synthetic"`
	IsSidechain bool            `json:"isSidechain"`
}

func readTranscriptEntries(projectsDir, path string) []transcriptEntry {
	data, partial := readTranscriptRange(projectsDir, path, fullReadLimit)
	return parseTranscriptEntries(data, partial)
}

func readTranscriptRange(projectsDir, path string, limit int64) ([]byte, bool) {
	file, err := openValidatedTranscript(projectsDir, path)
	if err != nil {
		return nil, false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, false
	}
	start := info.Size() - limit
	if start < 0 {
		start = 0
	}
	partial := false
	if start > 0 {
		if _, err := file.Seek(start-1, io.SeekStart); err != nil {
			return nil, false
		}
		previous := []byte{0}
		if _, err := io.ReadFull(file, previous); err != nil {
			return nil, false
		}
		partial = previous[0] != '\n'
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, false
	}
	data, _ := io.ReadAll(io.LimitReader(file, limit))
	return data, partial
}

func openValidatedTranscript(projectsDir, path string) (*os.File, error) {
	if projectsDir == "" || path == "" || containsParentComponent(path) {
		return nil, errors.New("invalid transcript path")
	}
	root, err := filepath.Abs(projectsDir)
	if err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	lexicalRelative, err := filepath.Rel(root, absolute)
	if err != nil || lexicalRelative == ".." || strings.HasPrefix(lexicalRelative, ".."+string(os.PathSeparator)) {
		return nil, errors.New("transcript path escapes projects directory")
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	resolvedPath, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	resolvedRelative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(os.PathSeparator)) {
		return nil, errors.New("resolved transcript path escapes projects directory")
	}
	current := root
	if lexicalRelative != "." {
		for _, component := range strings.Split(lexicalRelative, string(os.PathSeparator)) {
			current = filepath.Join(current, component)
			info, err := os.Lstat(current)
			if err != nil {
				return nil, err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("symlink transcript path is not allowed")
			}
		}
	}
	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, errors.New("transcript is not a regular file")
	}
	return file, nil
}

func containsParentComponent(path string) bool {
	for _, component := range strings.Split(filepath.ToSlash(path), "/") {
		if component == ".." {
			return true
		}
	}
	return false
}

func parseTranscriptEntries(data []byte, partialFirstLine bool) []transcriptEntry {
	if partialFirstLine {
		if newline := bytes.IndexByte(data, '\n'); newline >= 0 {
			data = data[newline+1:]
		} else {
			return nil
		}
	}
	lines := bytes.Split(data, []byte{'\n'})
	entries := make([]transcriptEntry, 0, len(lines))
	for _, line := range lines {
		var entry transcriptEntry
		if len(bytes.TrimSpace(line)) != 0 && json.Unmarshal(line, &entry) == nil {
			entries = append(entries, entry)
		}
	}
	return entries
}

func extractUserMessages(entries []transcriptEntry, current string) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != "user" || entry.IsMeta || entry.IsSynthetic || entry.Meta || entry.Synthetic || entry.IsSidechain {
			continue
		}
		message := strings.TrimSpace(systemReminderPattern.ReplaceAllString(userMessageText(entry.Message), ""))
		if message != "" {
			result = append(result, message)
		}
	}
	for index := len(result) - 1; index >= 0; index-- {
		if result[index] == current {
			return append(result[:index], result[index+1:]...)
		}
	}
	return result
}

func userMessageText(raw json.RawMessage) string {
	var message struct {
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(raw, &message) != nil {
		return ""
	}
	var text string
	if json.Unmarshal(message.Content, &text) == nil {
		return text
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(message.Content, &blocks) != nil {
		return ""
	}
	parts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func sanitizeTitle(input string) string {
	filtered := make([]rune, 0, len([]rune(input)))
	baseAvailable := false
	for _, value := range []rune(input) {
		if value == '\n' || value == '\r' || value == '\t' || value == '\u2028' || value == '\u2029' {
			value = ' '
		}
		switch {
		case unicode.IsLetter(value) || unicode.IsNumber(value):
			filtered = append(filtered, value)
			baseAvailable = true
		case isAllowedMark(value):
			if baseAvailable {
				filtered = append(filtered, value)
			}
		case unicode.Is(unicode.Zs, value):
			filtered = append(filtered, ' ')
			baseAvailable = false
		case value == '-' || value == '_' || value == '·' || value == '#':
			filtered = append(filtered, value)
			baseAvailable = false
		default:
			baseAvailable = false
		}
	}
	label := strings.Join(strings.Fields(string(filtered)), " ")
	if label == "" || !containsLetterOrNumber(label) {
		return ""
	}
	label = cutTitleRunes(label, maxTitleRunes)
	if !containsLetterOrNumber(label) {
		return ""
	}
	return label
}

func containsLetterOrNumber(text string) bool {
	for _, value := range text {
		if unicode.IsLetter(value) || unicode.IsNumber(value) {
			return true
		}
	}
	return false
}

func isAllowedMark(value rune) bool {
	if !unicode.Is(unicode.Mn, value) {
		return false
	}
	return !(value >= '\uFE00' && value <= '\uFE0F') &&
		!(value >= '\U000E0100' && value <= '\U000E01EF')
}

func cutTitleRunes(label string, limit int) string {
	runes := []rune(label)
	if len(runes) <= limit {
		return label
	}
	cut := limit
	if isAllowedMark(runes[cut]) {
		for cut > 0 && isAllowedMark(runes[cut-1]) {
			cut--
		}
		if cut > 0 {
			cut--
		}
	}
	return strings.TrimSpace(string(runes[:cut]))
}

func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return text
}

func main() {
	app, err := newApplication()
	if err != nil {
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--daemon" {
		if err := app.runDaemon(); err != nil {
			_, _ = fmt.Fprintf(app.stderr, "herdr-title: daemon error: %v\n", err)
		}
		return
	}
	_ = app.runHook(os.Stdin, os.Getenv("HERDR_TAB_ID"), os.Getenv("HERDR_WORKSPACE_ID"))
}
