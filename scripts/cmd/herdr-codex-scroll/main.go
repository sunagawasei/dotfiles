// Command herdr-codex-scroll handles Alt+J/K/D/U pressed in a herdr pane.
// For a Codex pane it sends a one-line wheel event or PageUp/PageDown, which
// Codex's normal conversation view accepts; for any other pane it forwards the
// original Alt key unchanged. The target is the pane id herdr passed at key
// press time, never the currently focused pane.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const codexAgent = "codex"

// codexInput is the raw input sent to a Codex pane for each Alt key.
// herdr's `pane send-keys` has no PageUp/PageDown name, so every key goes out
// as raw bytes via `pane send-text`.
var codexInput = map[string]string{
	"alt+j": "\x1b[<65;1;1M", // SGR wheel down, one line
	"alt+k": "\x1b[<64;1;1M", // SGR wheel up, one line
	"alt+d": "\x1b[6~",       // PageDown
	"alt+u": "\x1b[5~",       // PageUp
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
	input, ok := codexInput[key]
	if !ok {
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
			} `json:"pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("pane get %s: %w", pane, err)
	}
	if resp.Result.Pane.PaneID != pane {
		return fmt.Errorf("pane get returned %q for %q", resp.Result.Pane.PaneID, pane)
	}

	if resp.Result.Pane.Agent != codexAgent {
		_, err = exe(bin, "pane", "send-keys", pane, key)
	} else {
		_, err = exe(bin, "pane", "send-text", pane, input)
	}
	return err
}

func main() {
	if err := run(os.Args[1:], os.Getenv, execRunner); err != nil {
		fmt.Fprintln(os.Stderr, "herdr-codex-scroll:", err)
		os.Exit(1)
	}
}
