// Command herdr-codex-scroll handles Alt+J/K/D/U pressed in a herdr pane.
// Codex gets a one-line wheel event or PageUp/PageDown. Cursor's conversation
// lives in the host scrollback, so those keys move that viewport (one line for
// J/K, one page for D/U). Every other pane gets the original Alt key. The
// target is the pane id herdr passed at key press time, never the focused pane.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	codexAgent  = "codex"
	cursorAgent = "cursor"
)

// wheelCol is the pointer column of a synthesized wheel event.
const wheelCol = 5

// codexInput returns the raw input sent to a Codex pane for each Alt key.
// herdr's `pane send-keys` has no PageUp/PageDown name, so every key goes out
// as raw bytes via `pane send-text`. Codex ignores wheel events outside its
// transcript area, which excludes the top row and the composer at the bottom,
// so the pointer sits a third of the way down.
func codexInput(key string, viewportRows int) (string, bool) {
	row := max(3, viewportRows/3)
	switch key {
	case "alt+j":
		return fmt.Sprintf("\x1b[<65;%d;%dM", wheelCol, row), true // wheel down, one line
	case "alt+k":
		return fmt.Sprintf("\x1b[<64;%d;%dM", wheelCol, row), true // wheel up, one line
	case "alt+d":
		return "\x1b[6~", true // PageDown
	case "alt+u":
		return "\x1b[5~", true // PageUp
	}
	return "", false
}

// cursorScroll is the host-viewport movement for a cursor pane.
// Up reveals older lines. A page is the pane height, matching Claude's full page.
func cursorScroll(key string, viewportRows int) (direction string, lines int, ok bool) {
	page := viewportRows
	if page < 1 {
		page = 1
	}
	switch key {
	case "alt+j":
		return "down", 1, true
	case "alt+k":
		return "up", 1, true
	case "alt+d":
		return "down", page, true
	case "alt+u":
		return "up", page, true
	}
	return "", 0, false
}

// passThroughScroll reports errors from `pane scroll` that mean the host
// viewport was not moved, so the original Alt key should be delivered instead.
// That covers the alternate screen, and a server or CLI that does not have
// pane scroll yet.
func passThroughScroll(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, `"code":"alternate_screen"`) ||
		strings.Contains(msg, "unknown variant") ||
		strings.Contains(msg, "herdr pane commands:")
}

type runner func(bin string, args ...string) ([]byte, error)

func execRunner(bin string, args ...string) ([]byte, error) {
	out, err := exec.Command(bin, args...).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// herdr reports API errors as JSON on stdout
		err = fmt.Errorf("%w: %s%s", err, bytes.TrimSpace(out), bytes.TrimSpace(exitErr.Stderr))
	}
	return out, err
}

func run(args []string, getenv func(string) string, exe runner) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: herdr-codex-scroll <alt+j|alt+k|alt+d|alt+u>")
	}
	key := args[0]
	if _, ok := codexInput(key, 0); !ok {
		return fmt.Errorf("unsupported key %q", key)
	}
	pane := getenv("HERDR_ACTIVE_PANE_ID")
	if pane == "" {
		return fmt.Errorf("HERDR_ACTIVE_PANE_ID is not set")
	}
	bin := getenv("HERDR_BIN_PATH")
	if bin == "" {
		bin = "herdr"
	}

	out, err := exe(bin, "pane", "get", pane)
	if err != nil {
		return fmt.Errorf("pane get %s: %w", pane, err)
	}
	var resp struct {
		Result struct {
			Pane struct {
				PaneID string `json:"pane_id"`
				Agent  string `json:"agent"`
				Scroll struct {
					ViewportRows int `json:"viewport_rows"`
				} `json:"scroll"`
			} `json:"pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("pane get %s: %w", pane, err)
	}
	if resp.Result.Pane.PaneID != pane {
		return fmt.Errorf("pane get returned %q for %q", resp.Result.Pane.PaneID, pane)
	}

	switch resp.Result.Pane.Agent {
	case codexAgent:
		input, _ := codexInput(key, resp.Result.Pane.Scroll.ViewportRows)
		_, err = exe(bin, "pane", "send-text", pane, input)
	case cursorAgent:
		direction, lines, _ := cursorScroll(key, resp.Result.Pane.Scroll.ViewportRows)
		_, err = exe(bin, "pane", "scroll", pane, "--direction", direction, "--lines", strconv.Itoa(lines))
		if passThroughScroll(err) {
			_, err = exe(bin, "pane", "send-keys", pane, key)
		}
	default:
		_, err = exe(bin, "pane", "send-keys", pane, key)
	}
	return err
}

func main() {
	if err := run(os.Args[1:], os.Getenv, execRunner); err != nil {
		fmt.Fprintln(os.Stderr, "herdr-codex-scroll:", err)
		os.Exit(1)
	}
}
