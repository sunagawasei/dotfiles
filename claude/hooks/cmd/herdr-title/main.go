// herdr-title は UserPromptSubmit hook から worker を切り離し、直近の会話から
// herdr の tab と単一 tab workspace のタイトルを生成します。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"
)

const (
	stateDirName         = "herdr-title"
	tailReadLimit        = 1 << 20
	fullReadLimit        = 32 << 20
	maxHistoryEntries    = 4
	maxInputRunes        = 300
	maxTitleRunes        = 12
	lockTimeout          = time.Second
	herdrCommandTimeout  = 5 * time.Second
	cursorCommandTimeout = 60 * time.Second
	staleFileAge         = time.Hour
)

var systemReminderPattern = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)

type hookInput struct {
	HookEventName  string `json:"hook_event_name"`
	Prompt         string `json:"prompt"`
	TranscriptPath string `json:"transcript_path"`
}

type workerPayload struct {
	Prompt         string `json:"prompt"`
	TranscriptPath string `json:"transcript_path"`
	TabID          string `json:"tab_id"`
	WorkspaceID    string `json:"workspace_id"`
	Generation     int64  `json:"generation"`
}

type cursorRequest struct {
	Cwd    string
	Env    []string
	Output *os.File
	Prompt string
}

// commandRunner は外部 CLI を実行します。テストでは fake に差し替えます。
type commandRunner interface {
	GenerateTitle(context.Context, cursorRequest) error
	RunHerdr(context.Context, ...string) ([]byte, error)
}

type execCommandRunner struct {
	cursorCommand string
	herdrCommand  string
}

func (r execCommandRunner) GenerateTitle(ctx context.Context, request cursorRequest) error {
	cmd := exec.Command(r.cursorCommand, cursorArguments(request.Prompt)...)
	cmd.Dir = request.Cwd
	cmd.Env = request.Env
	cmd.Stdout = request.Output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
		return ctx.Err()
	}
}

func cursorArguments(prompt string) []string {
	return []string{
		"-p", "--trust", "--mode", "ask", "--model", "composer-2.5",
		"--output-format", "text", prompt,
	}
}

func (r execCommandRunner) RunHerdr(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, r.herdrCommand, args...).Output()
}

type workerStarter func(executable, payloadPath string) error

type application struct {
	tmpDir        string
	now           func() time.Time
	environ       func() []string
	executable    func() (string, error)
	commands      commandRunner
	startDetached workerStarter
}

func newApplication() *application {
	return &application{
		tmpDir:     os.TempDir(),
		now:        time.Now,
		environ:    os.Environ,
		executable: os.Executable,
		commands: execCommandRunner{
			cursorCommand: "cursor-agent",
			herdrCommand:  "herdr",
		},
		startDetached: startDetachedWorker,
	}
}

func startDetachedWorker(executable, payloadPath string) error {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()

	cmd := exec.Command(executable, "--worker", payloadPath)
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	return nil
}

func (a *application) stateDir() string {
	return filepath.Join(a.tmpDir, stateDirName)
}

func escapeWorkspaceID(id string) string {
	replacer := strings.NewReplacer("%", "%25", ":", "%3A", "/", "%2F")
	return replacer.Replace(id)
}

func (a *application) lockPath(workspaceID string) string {
	return filepath.Join(a.stateDir(), "ws_"+escapeWorkspaceID(workspaceID)+".lock")
}

func (a *application) generationPath(workspaceID string) string {
	return filepath.Join(a.stateDir(), "ws_"+escapeWorkspaceID(workspaceID)+".gen")
}

type workspaceLock struct {
	file *os.File
}

func (l *workspaceLock) release() {
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func (a *application) acquireWorkspaceLock(workspaceID string, timeout time.Duration) (*workspaceLock, error) {
	if err := os.MkdirAll(a.stateDir(), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(a.lockPath(workspaceID), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &workspaceLock{file: file}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			file.Close()
			return nil, err
		}
		if !time.Now().Before(deadline) {
			file.Close()
			return nil, context.DeadlineExceeded
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (a *application) readGenerations(workspaceID string) (map[string]int64, error) {
	file, err := os.OpenFile(a.generationPath(workspaceID), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	generations := make(map[string]int64)
	if len(bytes.TrimSpace(data)) == 0 {
		return generations, nil
	}
	if err := json.Unmarshal(data, &generations); err != nil {
		return nil, err
	}
	return generations, nil
}

func (a *application) writeGenerations(workspaceID string, generations map[string]int64) error {
	data, err := json.Marshal(generations)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(a.generationPath(workspaceID), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}

func (a *application) reserveGeneration(workspaceID, tabID string) (int64, error) {
	lock, err := a.acquireWorkspaceLock(workspaceID, lockTimeout)
	if err != nil {
		return 0, err
	}
	defer lock.release()

	generations, err := a.readGenerations(workspaceID)
	if err != nil {
		return 0, err
	}
	next := generations[tabID] + 1
	if now := a.now().UnixNano(); now > next {
		next = now
	}
	generations[tabID] = next
	if err := a.writeGenerations(workspaceID, generations); err != nil {
		return 0, err
	}
	return next, nil
}

func (a *application) runHook(reader io.Reader, tabID, workspaceID string) error {
	var input hookInput
	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		return err
	}
	if tabID == "" && workspaceID == "" {
		return nil
	}

	generation, err := a.reserveGeneration(workspaceID, tabID)
	if err != nil {
		return err
	}
	payload := workerPayload{
		Prompt: input.Prompt, TranscriptPath: input.TranscriptPath,
		TabID: tabID, WorkspaceID: workspaceID, Generation: generation,
	}
	payloadPath, err := a.writePayload(payload)
	if err != nil {
		return err
	}
	executable, err := a.executable()
	if err != nil {
		os.Remove(payloadPath)
		return err
	}
	if err := a.startDetached(executable, payloadPath); err != nil {
		os.Remove(payloadPath)
		return err
	}
	return nil
}

func (a *application) writePayload(payload workerPayload) (string, error) {
	if err := os.MkdirAll(a.stateDir(), 0o700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(a.stateDir(), "payload-*")
	if err != nil {
		return "", err
	}
	path := file.Name()
	ok := false
	defer func() {
		file.Close()
		if !ok {
			os.Remove(path)
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return "", err
	}
	if err := json.NewEncoder(file).Encode(payload); err != nil {
		return "", err
	}
	if err := file.Sync(); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	ok = true
	return path, nil
}

func (a *application) runWorker(payloadPath string) error {
	data, readErr := os.ReadFile(payloadPath)
	removeErr := os.Remove(payloadPath)
	if readErr != nil {
		return readErr
	}
	if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return removeErr
	}
	var payload workerPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	a.cleanupStaleFiles()
	title := a.generateTitle(payload)
	if title == "" {
		title = sanitizeTitle(extractConversationTitle(payload.TranscriptPath))
	}
	if title == "" {
		return nil
	}
	return a.applyTitle(payload, title)
}

func (a *application) cleanupStaleFiles() {
	entries, err := os.ReadDir(a.stateDir())
	if err != nil {
		return
	}
	cutoff := a.now().Add(-staleFileAge)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "payload-") && !strings.HasPrefix(name, "out-") {
			continue
		}
		info, err := entry.Info()
		if err == nil && !info.ModTime().After(cutoff) {
			_ = os.Remove(filepath.Join(a.stateDir(), name))
		}
	}
}

func (a *application) generateTitle(payload workerPayload) string {
	inputs := buildTitleInputs(payload.TranscriptPath, payload.Prompt)
	if len(inputs) == 0 {
		return ""
	}
	workspaceDir := filepath.Join(a.stateDir(), "ws")
	if err := os.MkdirAll(workspaceDir, 0o700); err != nil {
		return ""
	}
	output, err := os.CreateTemp(a.stateDir(), "out-*")
	if err != nil {
		return ""
	}
	outputPath := output.Name()
	defer os.Remove(outputPath)
	if err := output.Chmod(0o600); err != nil {
		output.Close()
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), cursorCommandTimeout)
	err = a.commands.GenerateTitle(ctx, cursorRequest{
		Cwd: workspaceDir, Env: filteredEnvironment(a.environ()), Output: output,
		Prompt: titlePrompt(inputs),
	})
	cancel()
	closeErr := output.Close()
	if err != nil || closeErr != nil {
		return ""
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return ""
	}
	return sanitizeTitle(string(data))
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

func titlePrompt(inputs []string) string {
	return "以下の直近のユーザー発話から、会話内容を表すタイトルを1つ作成してください。" +
		"日本語、12文字以内の名詞句とし、句読点、引用符、記号を含めず、ラベルだけを出力してください。\n\n" +
		strings.Join(inputs, "\n")
}

func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return text
}

func buildTitleInputs(transcriptPath, currentPrompt string) []string {
	entries := readTranscriptEntries(transcriptPath)
	history := extractUserMessages(entries, currentPrompt, maxHistoryEntries)
	inputs := make([]string, 0, len(history)+1)
	for _, item := range history {
		if item = truncateRunes(strings.TrimSpace(item), maxInputRunes); item != "" {
			inputs = append(inputs, item)
		}
	}
	if current := truncateRunes(strings.TrimSpace(currentPrompt), maxInputRunes); current != "" {
		inputs = append(inputs, current)
	}
	return inputs
}

type transcriptEntry struct {
	Type        string          `json:"type"`
	Message     json.RawMessage `json:"message"`
	AITitle     string          `json:"aiTitle"`
	CustomTitle string          `json:"customTitle"`
	IsMeta      bool            `json:"isMeta"`
	IsSynthetic bool            `json:"isSynthetic"`
	Meta        bool            `json:"meta"`
	Synthetic   bool            `json:"synthetic"`
}

func readTranscriptEntries(path string) []transcriptEntry {
	data, partial := readTranscriptRange(path, tailReadLimit)
	entries := parseTranscriptEntries(data, partial)
	if len(entries) == 0 {
		data, partial = readTranscriptRange(path, fullReadLimit)
		entries = parseTranscriptEntries(data, partial)
	}
	return entries
}

func readTranscriptRange(path string, limit int64) ([]byte, bool) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, false
	}
	start := info.Size() - limit
	partial := start > 0
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, false
	}
	data, _ := io.ReadAll(io.LimitReader(file, limit))
	return data, partial
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

func extractUserMessages(entries []transcriptEntry, currentPrompt string, limit int) []string {
	result := make([]string, 0, limit)
	for index := len(entries) - 1; index >= 0 && len(result) < limit; index-- {
		entry := entries[index]
		if entry.Type != "user" || entry.IsMeta || entry.IsSynthetic || entry.Meta || entry.Synthetic {
			continue
		}
		message := userMessageText(entry.Message)
		message = strings.TrimSpace(systemReminderPattern.ReplaceAllString(message, ""))
		if message == "" || message == currentPrompt {
			continue
		}
		result = append(result, message)
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
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

func extractConversationTitle(path string) string {
	entries := readTranscriptEntries(path)
	for index := len(entries) - 1; index >= 0; index-- {
		switch entries[index].Type {
		case "ai-title":
			if title := strings.TrimSpace(entries[index].AITitle); title != "" {
				return title
			}
		case "custom-title":
			if title := strings.TrimSpace(entries[index].CustomTitle); title != "" {
				return title
			}
		}
	}
	return ""
}

func sanitizeTitle(input string) string {
	runes := []rune(input)
	filtered := make([]rune, 0, len(runes))
	baseAvailable := false
	for _, value := range runes {
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
		case value == '-' || value == '_' || value == '·':
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

func (a *application) runHerdr(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), herdrCommandTimeout)
	defer cancel()
	return a.commands.RunHerdr(ctx, args...)
}

func (a *application) applyTitle(payload workerPayload, title string) error {
	lock, err := a.acquireWorkspaceLock(payload.WorkspaceID, lockTimeout)
	if err != nil {
		return err
	}
	defer lock.release()

	generations, err := a.readGenerations(payload.WorkspaceID)
	if err != nil || generations[payload.TabID] != payload.Generation {
		return err
	}
	if payload.TabID != "" {
		if _, err := a.runHerdr("tab", "rename", payload.TabID, title); err != nil {
			return nil
		}
	}
	if payload.WorkspaceID == "" {
		return nil
	}
	data, err := a.runHerdr("tab", "list")
	if err != nil {
		return nil
	}
	tabs, err := parseTabList(data)
	if err != nil {
		return nil
	}
	workspaceTabs := make([]tabInfo, 0, len(tabs))
	for _, tab := range tabs {
		if tab.WorkspaceID == payload.WorkspaceID {
			workspaceTabs = append(workspaceTabs, tab)
		}
	}
	if len(workspaceTabs) != 1 || (payload.TabID != "" && workspaceTabs[0].TabID != payload.TabID) {
		return nil
	}
	if _, err := a.runHerdr("workspace", "rename", payload.WorkspaceID, title); err != nil {
		_, _ = a.runHerdr("workspace", "rename", payload.WorkspaceID, title)
	}
	return nil
}

func main() {
	app := newApplication()
	if len(os.Args) == 3 && os.Args[1] == "--worker" {
		_ = app.runWorker(os.Args[2])
		return
	}
	_ = app.runHook(os.Stdin, os.Getenv("HERDR_TAB_ID"), os.Getenv("HERDR_WORKSPACE_ID"))
}
