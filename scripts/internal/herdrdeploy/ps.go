package herdrdeploy

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Process is one line of `ps -axo pid=,args=` output.
type Process struct {
	PID  int
	Args string
}

var psLine = regexp.MustCompile(`^\s*(\d+)\s+(.*\S)\s*$`)

// ClassifyHerdrProcesses parses `ps -axo pid=,args=` output and splits herdr
// processes into servers (first arg's basename is "herdr", second arg is
// "server") and others (basename "herdr" but not a server, e.g. the client).
// Lines whose first argument's basename isn't "herdr" are ignored, so a
// process like "otherherdr server" or "ugrep ... herdr ..." never matches.
func ClassifyHerdrProcesses(psOutput string) (servers, others []Process) {
	for line := range strings.SplitSeq(psOutput, "\n") {
		match := psLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		pid, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		args := match[2]
		tokens := strings.Fields(args)
		if len(tokens) == 0 || filepath.Base(tokens[0]) != "herdr" {
			continue
		}
		process := Process{PID: pid, Args: args}
		if len(tokens) >= 2 && tokens[1] == "server" {
			servers = append(servers, process)
		} else {
			others = append(others, process)
		}
	}
	return servers, others
}
