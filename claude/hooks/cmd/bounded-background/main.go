// Command bounded-background is a Claude Code PreToolUse hook. It rewrites
// background Bash tool calls to run under bg-deadline so no background
// process can outlive the timeout the harness intended for it.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultDeadlineSeconds = 600
	maxDeadlineSeconds     = 600
	bgDeadlineBinaryName   = "bg-deadline"

	// shellPath matches the harness's Bash-tool execution shell so the
	// wrapped command's quoting semantics line up with the original.
	shellPath = "/bin/zsh"

	// schemaHint is appended to deny reasons caused by input this hook
	// doesn't recognize, since that's the actionable next step for whoever
	// reads it.
	schemaHint = "フックの入力schemaが変わった可能性があるので claude/hooks/cmd/bounded-background を確認すること"
)

// hookInput is the PreToolUse payload. ToolInput is decoded as a raw map so
// unknown/future fields survive the round trip into updatedInput untouched.
type hookInput struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName            string         `json:"hookEventName"`
	UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
	PermissionDecision       string         `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
}

// bgDeadlineResolution is the result of locating the bg-deadline binary.
// Path is set whenever a candidate path could be computed at all, even if
// Err is non-nil, so callers can still recognize an already-wrapped command.
type bgDeadlineResolution struct {
	Path string
	Err  error
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, defaultBgDeadlineResolver))
}

// run executes the hook end-to-end and returns the process exit code. It's
// split out from main so tests can inject a failing writer without ending
// the test process via os.Exit.
func run(in io.Reader, out io.Writer, errOut io.Writer, resolve func() bgDeadlineResolution) int {
	output := decideFromReader(in, resolve)
	if output == nil {
		return 0
	}
	if err := json.NewEncoder(out).Encode(output); err != nil {
		// If the harness never receives our decision, the original command
		// (possibly an unbounded background one) would run unmodified, so
		// fail the tool call instead of exiting 0.
		fmt.Fprintf(errOut, "bounded-background: failed to write hook output: %v\n", err)
		return 2
	}
	return 0
}

// defaultBgDeadlineResolver looks for bg-deadline next to this executable.
func defaultBgDeadlineResolver() bgDeadlineResolution {
	exe, err := os.Executable()
	if err != nil {
		return bgDeadlineResolution{Err: fmt.Errorf("resolve own executable path: %w", err)}
	}
	path := filepath.Join(filepath.Dir(exe), bgDeadlineBinaryName)
	info, err := os.Stat(path)
	if err != nil {
		return bgDeadlineResolution{Path: path, Err: err}
	}
	if info.IsDir() || info.Mode()&0o111 == 0 {
		return bgDeadlineResolution{Path: path, Err: fmt.Errorf("%s is not executable", path)}
	}
	return bgDeadlineResolution{Path: path}
}

// decideFromReader reads the full PreToolUse payload before handing off to
// decide. A read failure denies fail-closed instead of letting an unbounded
// background command through untouched.
func decideFromReader(r io.Reader, resolve func() bgDeadlineResolution) *hookOutput {
	data, err := io.ReadAll(r)
	if err != nil {
		return denyOutput(fmt.Sprintf("stdin の読み取りに失敗した(%v)。%s", err, schemaHint))
	}
	return decide(data, resolve)
}

// decide implements the PreToolUse rewrite/deny logic. A nil result means
// "do nothing" (empty stdout, exit 0) and is reserved for recognized
// non-background cases; anything else this hook can't make sense of denies
// fail-closed rather than passing an unbounded background command through.
func decide(rawInput []byte, resolve func() bgDeadlineResolution) *hookOutput {
	var in hookInput
	if err := json.Unmarshal(rawInput, &in); err != nil {
		return denyOutput(fmt.Sprintf("stdin の JSON を解析できない(%v)。%s", err, schemaHint))
	}
	if in.ToolName != "Bash" {
		return nil
	}

	switch classifyRunInBackground(in.ToolInput) {
	case runInBackgroundAbsent, runInBackgroundFalse:
		return nil
	case runInBackgroundInvalid:
		return denyOutput(fmt.Sprintf("tool_input.run_in_background が bool 型でない。%s", schemaHint))
	}

	command, ok := in.ToolInput["command"].(string)
	if !ok || command == "" {
		return denyOutput("tool_input.command が文字列として取得できないため background 実行を安全にラップできない")
	}

	res := resolve()
	quotedPath := singleQuote(res.Path)
	deadline := computeDeadlineSeconds(in.ToolInput)

	if token, ok := isAlreadyWrapped(command, quotedPath); ok {
		rewritten, changed := clampAlreadyWrappedDeadline(command, quotedPath, token, deadline)
		if !changed {
			return nil
		}
		return rewriteCommandOutput(in.ToolInput, rewritten)
	}
	if res.Err != nil {
		return denyOutput(fmt.Sprintf(
			"bg-deadline が見つからないため background 実行を許可できない(%v)。cd claude/hooks && go build -o . ./... でビルドすること",
			res.Err,
		))
	}

	return rewriteCommandOutput(in.ToolInput, wrapCommand(res.Path, deadline, command))
}

func rewriteCommandOutput(toolInput map[string]any, command string) *hookOutput {
	updated := maps.Clone(toolInput)
	updated["command"] = command
	return &hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName: "PreToolUse",
			UpdatedInput:  updated,
		},
	}
}

func denyOutput(reason string) *hookOutput {
	return &hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		},
	}
}

type runInBackgroundState int

const (
	runInBackgroundAbsent runInBackgroundState = iota
	runInBackgroundFalse
	runInBackgroundTrue
	runInBackgroundInvalid
)

// classifyRunInBackground distinguishes "not requested" (absent/false) from
// "malformed" (present but not a bool), since only the former is safe to
// pass through silently.
func classifyRunInBackground(m map[string]any) runInBackgroundState {
	v, ok := m["run_in_background"]
	if !ok {
		return runInBackgroundAbsent
	}
	b, ok := v.(bool)
	if !ok {
		return runInBackgroundInvalid
	}
	if b {
		return runInBackgroundTrue
	}
	return runInBackgroundFalse
}

// isAlreadyWrapped reports whether command is exactly the canonical form
// wrapCommand produces for some deadline token: quotedBgDeadlinePath + " " +
// <token> + " " + shellPath + " -c " + <singly-quoted payload>, with
// nothing trailing after the closing quote. A prefix match isn't enough: a
// well-formed wrapper followed by ";", "&&", a newline, or "|" would let an
// unbounded command ride along outside bg-deadline's process group.
func isAlreadyWrapped(command, quotedBgDeadlinePath string) (token string, ok bool) {
	if quotedBgDeadlinePath == "''" {
		return "", false
	}
	rest, ok := strings.CutPrefix(command, quotedBgDeadlinePath+" ")
	if !ok {
		return "", false
	}

	idx := strings.IndexByte(rest, ' ')
	if idx == -1 {
		return "", false
	}
	token, rest = rest[:idx], rest[idx+1:]

	rest, ok = strings.CutPrefix(rest, shellPath+" -c ")
	if !ok {
		return "", false
	}

	_, trailing, ok := parseSingleQuoted(rest)
	if !ok || trailing != "" {
		return "", false
	}
	return token, true
}

// parseSingleQuoted decodes a POSIX-quoted string produced by singleQuote,
// starting at an opening single quote in s. It returns the decoded literal
// and whatever text follows the closing quote, so a caller can require
// nothing trails it.
func parseSingleQuoted(s string) (decoded, trailing string, ok bool) {
	if !strings.HasPrefix(s, "'") {
		return "", "", false
	}
	var b strings.Builder
	i := 1
	for i < len(s) {
		if s[i] != '\'' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if strings.HasPrefix(s[i:], `'\''`) {
			b.WriteByte('\'')
			i += 4
			continue
		}
		return b.String(), s[i+1:], true
	}
	return "", "", false
}

// clampAlreadyWrappedDeadline re-validates the deadline token of an
// already-wrapped command (as recognized by isAlreadyWrapped) against
// capSeconds (this call's own computeDeadlineSeconds result), replacing just
// that token if it's missing/non-numeric/<=0/greater than the cap. Without
// this, an already-wrapped command could carry a deadline longer than what
// tool_input.timeout allows for this call.
func clampAlreadyWrappedDeadline(command, quotedBgDeadlinePath, token string, capSeconds int) (rewritten string, changed bool) {
	if n, err := strconv.Atoi(token); err == nil && n >= 1 && n <= capSeconds {
		return command, false
	}
	rest := strings.TrimPrefix(command, quotedBgDeadlinePath+" "+token)
	return quotedBgDeadlinePath + " " + strconv.Itoa(capSeconds) + rest, true
}

// computeDeadlineSeconds derives the deadline from tool_input.timeout
// (milliseconds), rounding up so a sub-second timeout still gets at least
// one full second, then clamping to [1, maxDeadlineSeconds].
func computeDeadlineSeconds(toolInput map[string]any) int {
	raw, ok := toolInput["timeout"]
	if !ok {
		return defaultDeadlineSeconds
	}
	ms, ok := toFloat(raw)
	if !ok || ms <= 0 {
		return defaultDeadlineSeconds
	}
	sec := int(math.Ceil(ms / 1000))
	return min(max(sec, 1), maxDeadlineSeconds)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// wrapCommand routes the original command through bg-deadline under
// /bin/zsh -c so the deadline enforces its whole process group, not just a
// single argv the shell would otherwise split on our behalf. Both the
// bg-deadline path and the original command are quoted since either can
// contain shell metacharacters (e.g. a space in a checkout path).
func wrapCommand(bgDeadlinePath string, deadlineSeconds int, original string) string {
	return fmt.Sprintf("%s %d %s -c %s", singleQuote(bgDeadlinePath), deadlineSeconds, shellPath, singleQuote(original))
}

// singleQuote wraps s for POSIX sh/zsh -c using the standard
// close-escape-reopen quoting technique so the shell sees exactly the
// original text regardless of embedded quotes.
func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
