// Command wrapup-recall prints the user's own utterances from the recall
// archive for the wrapup skill. The exit code tells "recall is not installed"
// from "recall failed", because the skill must not treat a failed fetch as
// "nothing to learn".
//
// Usage:
//
//	wrapup-recall current        every utterance of $CLAUDE_CODE_SESSION_ID, untruncated
//	wrapup-recall past [-n 6]    one line per utterance (200 runes) of the latest
//	                             sessions of this repo, minus the current session
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
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

	pastLineRunes = 200
)

// Harness-injected user turns that carry no utterance of the user.
var skipPrefixes = []string{"<", "[Request interrupted", "This session is being", "[task:", "Caveat"}

var errEmptyCurrent = errors.New("the current session has no user utterance (recall import may not have picked it up)")

type runner func(args ...string) ([]byte, error)

type message struct {
	Role      string `json:"role"`
	BlockType string `json:"blockType"`
	Content   string `json:"content"`
}

type exported struct {
	Messages []message `json:"messages"`
}

type session struct {
	SessionID string `json:"sessionId"`
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

// pastIDs returns up to n session ids, newest first, leaving out the current one.
func pastIDs(sessions []session, current string, n int) []string {
	var ids []string
	for _, s := range sessions {
		if s.SessionID == current {
			continue
		}
		ids = append(ids, s.SessionID)
	}
	return ids[:min(n, len(ids))]
}

// oneLine joins the lines of s with spaces and cuts it to limit runes.
func oneLine(s string, limit int) string {
	r := []rune(strings.ReplaceAll(s, "\n", " "))
	return string(r[:min(limit, len(r))])
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
	return e.Messages, nil
}

func printCurrent(rc runner, id string, w io.Writer) error {
	messages, err := exportMessages(rc, id)
	if err != nil {
		return err
	}
	utterances := userUtterances(messages)
	if len(utterances) == 0 {
		return errEmptyCurrent
	}
	for i, u := range utterances {
		fmt.Fprintf(w, "[%d] %s\n\n", i+1, u)
	}
	return nil
}

func printPast(rc runner, repo, current string, n int, w io.Writer) error {
	out, err := call(rc, "list", "--repo", repo, "--limit", fmt.Sprint(n+1), "--format", "json")
	if err != nil {
		return err
	}
	var sessions []session
	if err := json.Unmarshal(out, &sessions); err != nil {
		return fmt.Errorf("recall list: unexpected output: %w", err)
	}
	// Written only after every session was read, so a failure never leaves a
	// partial list that looks complete.
	var buf bytes.Buffer
	for _, id := range pastIDs(sessions, current, n) {
		messages, err := exportMessages(rc, id)
		if err != nil {
			return err
		}
		fmt.Fprintf(&buf, "## %s\n", id)
		for _, u := range userUtterances(messages) {
			fmt.Fprintln(&buf, oneLine(u, pastLineRunes))
		}
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "wrapup-recall:", err)
	if errors.Is(err, exec.ErrNotFound) {
		return exitNoRecall
	}
	return exitFailed
}

func run(args []string, getenv func(string) string, cwd string, rc runner, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: wrapup-recall current | past [-n N]")
		return exitUsage
	}
	n := 6
	switch args[0] {
	case "current":
	case "past":
		fs := flag.NewFlagSet("past", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.IntVar(&n, "n", n, "number of past sessions")
		if err := fs.Parse(args[1:]); err != nil || n < 1 {
			return exitUsage
		}
	default:
		fmt.Fprintf(stderr, "wrapup-recall: unknown command %q\n", args[0])
		return exitUsage
	}

	id := getenv("CLAUDE_CODE_SESSION_ID")
	if id == "" {
		fmt.Fprintln(stderr, "wrapup-recall: CLAUDE_CODE_SESSION_ID is empty")
		return exitNoSession
	}
	if _, err := call(rc, "import"); err != nil {
		return fail(stderr, err)
	}

	var err error
	if args[0] == "current" {
		err = printCurrent(rc, id, stdout)
	} else {
		err = printPast(rc, cwd, id, n, stdout)
	}
	if err != nil {
		return fail(stderr, err)
	}
	return exitOK
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "wrapup-recall:", err)
		os.Exit(exitFailed)
	}
	os.Exit(run(os.Args[1:], os.Getenv, cwd, execRecall, os.Stdout, os.Stderr))
}
