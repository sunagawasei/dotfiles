// Command rift-chrome-nudge brings a freshly launched Chrome to the front so
// Rift starts managing its window. Rift 0.6.7 gets no launch notification for
// Chrome and only learns the window when it is first focused, so a relaunched
// Chrome stays unmanaged until the first click.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	chromeBundleID = "com.google.Chrome"
	// pgrep -f matches the full command line. The trailing $ leaves out the
	// automation Chrome, whose command line carries extra flags.
	chromePattern = `Google Chrome\.app/Contents/MacOS/Google Chrome$`

	pollInterval = time.Second
	// A window needs about a second after launch before activation raises it.
	nudgeInterval = time.Second
	nudgeTimeout  = 15 * time.Second
)

// newPids returns the pids in current that are not in seen.
func newPids(seen map[int]bool, current []int) []int {
	var fresh []int
	for _, pid := range current {
		if !seen[pid] {
			fresh = append(fresh, pid)
		}
	}
	slices.Sort(fresh)
	return fresh
}

// managed reports whether the `rift-cli query workspaces` output holds a
// window of pid.
func managed(workspacesJSON []byte, pid int) (bool, error) {
	var workspaces []struct {
		Windows []struct {
			ID struct {
				Pid int `json:"pid"`
			} `json:"id"`
		} `json:"windows"`
	}
	if err := json.Unmarshal(workspacesJSON, &workspaces); err != nil {
		return false, err
	}
	for _, ws := range workspaces {
		for _, w := range ws.Windows {
			if w.ID.Pid == pid {
				return true, nil
			}
		}
	}
	return false, nil
}

func chromePids() ([]int, error) {
	out, err := exec.Command("pgrep", "-f", chromePattern).Output()
	if err != nil {
		// pgrep exits 1 when nothing matches
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	var pids []int
	for field := range strings.FieldsSeq(string(out)) {
		pid, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("pgrep output %q: %w", field, err)
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// nudge activates Chrome until Rift lists a window for pid or the timeout passes.
func nudge(riftCLI string, pid int) {
	deadline := time.Now().Add(nudgeTimeout)
	for time.Now().Before(deadline) {
		out, err := exec.Command(riftCLI, "query", "workspaces").Output()
		if err != nil {
			log.Printf("rift-cli query workspaces: %v", err)
			return
		}
		ok, err := managed(out, pid)
		if err != nil {
			log.Printf("parse workspaces: %v", err)
			return
		}
		if ok {
			log.Printf("chrome pid %d is managed", pid)
			return
		}
		if err := exec.Command("open", "-b", chromeBundleID).Run(); err != nil {
			log.Printf("open -b %s: %v", chromeBundleID, err)
			return
		}
		time.Sleep(nudgeInterval)
	}
	log.Printf("chrome pid %d not managed within %s", pid, nudgeTimeout)
}

func main() {
	riftCLI := flag.String("rift-cli", "rift-cli", "path to rift-cli")
	flag.Parse()

	// Chrome already running at start is left alone.
	seen := map[int]bool{}
	if pids, err := chromePids(); err == nil {
		for _, pid := range pids {
			seen[pid] = true
		}
	}

	for range time.Tick(pollInterval) {
		pids, err := chromePids()
		if err != nil {
			log.Printf("pgrep: %v", err)
			continue
		}
		for _, pid := range newPids(seen, pids) {
			seen[pid] = true
			log.Printf("chrome launched: pid %d", pid)
			nudge(*riftCLI, pid)
		}
	}
}
