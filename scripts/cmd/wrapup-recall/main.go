// Command wrapup-recall prints the user's own utterances of the current
// session, for the wrapup skill. A Claude Code session comes from the recall
// archive; a Codex thread comes from its rollout files (codex.go). The exit
// code tells "recall is not installed" from "the fetch failed", because the
// skill must not treat a failed fetch as "nothing to learn".
//
// Usage:
//
//	wrapup-recall current    every utterance of the current session, untruncated
//
// The session is $CLAUDE_CODE_SESSION_ID, or $CODEX_THREAD_ID (else
// $CODEX_SESSION_ID) under Codex. Both families set is ambiguous and exits 5.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	exitOK        = 0
	exitUsage     = 2
	exitNoRecall  = 3
	exitFailed    = 4
	exitNoSession = 5
)

// Harness-injected user turns that carry no utterance of the user. A bare "<"
// is not here: text the user pastes arrives as <pasted_content>, and XML or
// HTML the user types starts with "<" too.
var skipPrefixes = []string{
	"<command-", "<local-command-", "<system-reminder", "<task-notification",
	"[Request interrupted", "This session is being", "[task:", "Caveat",
}

var errEmptyCurrent = errors.New("the current session has no user utterance")

type runner func(args ...string) ([]byte, error)

type message struct {
	Role      string `json:"role"`
	BlockType string `json:"blockType"`
	Content   string `json:"content"`
}

type exported struct {
	// A pointer tells a missing or null field (a format change) from an
	// empty list, which is a session without messages.
	Messages *[]message `json:"messages"`
}

// userUtterances keeps the text blocks the user typed.
func userUtterances(messages []message) []string {
	var out []string
	for _, m := range messages {
		if m.Role != "user" || m.BlockType != "text" {
			continue
		}
		if hasAnyPrefix(m.Content, skipPrefixes) {
			continue
		}
		out = append(out, m.Content)
	}
	return out
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if _, ok := strings.CutPrefix(s, p); ok {
			return true
		}
	}
	return false
}

func execRecall(args ...string) ([]byte, error) {
	return exec.Command("recall", args...).Output()
}

func call(rc runner, args ...string) ([]byte, error) {
	out, err := rc(args...)
	if err == nil {
		return out, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return nil, fmt.Errorf("recall %s: %w: %s", args[0], err, strings.TrimSpace(string(ee.Stderr)))
	}
	return nil, fmt.Errorf("recall %s: %w", args[0], err)
}

func exportMessages(rc runner, id string) ([]message, error) {
	out, err := call(rc, "export", id, "--format", "json")
	if err != nil {
		return nil, err
	}
	var e exported
	if err := json.Unmarshal(out, &e); err != nil {
		return nil, fmt.Errorf("recall export %s: unexpected output: %w", id, err)
	}
	if e.Messages == nil {
		return nil, fmt.Errorf("recall export %s: no messages field", id)
	}
	return *e.Messages, nil
}

func claudeUtterances(rc runner, id string) ([]string, error) {
	if _, err := call(rc, "import"); err != nil {
		return nil, err
	}
	messages, err := exportMessages(rc, id)
	if err != nil {
		return nil, err
	}
	return userUtterances(messages), nil
}

func printUtterances(utterances []string, w io.Writer) {
	for i, u := range utterances {
		fmt.Fprintf(w, "[%d] %s\n\n", i+1, u)
	}
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "wrapup-recall:", err)
	if errors.Is(err, exec.ErrNotFound) {
		return exitNoRecall
	}
	return exitFailed
}

// currentSession names the session this process runs in. Both families set
// cannot be told apart (a Codex worker spawned by Claude inherits
// CLAUDE_CODE_SESSION_ID), and a wrong guess prints another session's words.
func currentSession(getenv func(string) string) (codex bool, id string, err error) {
	claude := getenv("CLAUDE_CODE_SESSION_ID")
	thread := getenv("CODEX_THREAD_ID")
	if thread == "" {
		thread = getenv("CODEX_SESSION_ID")
	}
	switch {
	case claude != "" && thread != "":
		return false, "", errors.New("both CLAUDE_CODE_SESSION_ID and CODEX_THREAD_ID/CODEX_SESSION_ID are set; cannot tell which session is current")
	case claude != "":
		return false, claude, nil
	case thread != "":
		if !codexIDPattern.MatchString(thread) {
			return false, "", errors.New("CODEX_THREAD_ID/CODEX_SESSION_ID is not a thread id")
		}
		return true, thread, nil
	}
	return false, "", errors.New("CLAUDE_CODE_SESSION_ID and CODEX_THREAD_ID/CODEX_SESSION_ID are all empty")
}

func run(args []string, getenv func(string) string, rc runner, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "current" {
		fmt.Fprintln(stderr, "usage: wrapup-recall current")
		return exitUsage
	}
	codex, id, err := currentSession(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "wrapup-recall:", err)
		return exitNoSession
	}
	var utterances []string
	if codex {
		home, herr := codexHome(getenv)
		if herr != nil {
			return fail(stderr, herr)
		}
		utterances, err = codexUtterances(home, id)
	} else {
		utterances, err = claudeUtterances(rc, id)
	}
	if err != nil {
		return fail(stderr, err)
	}
	if len(utterances) == 0 {
		if !codex {
			return fail(stderr, fmt.Errorf("%w (recall import may not have picked it up)", errEmptyCurrent))
		}
		return fail(stderr, errEmptyCurrent)
	}
	printUtterances(utterances, stdout)
	return exitOK
}

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, execRecall, os.Stdout, os.Stderr))
}
