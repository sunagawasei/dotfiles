// Command task-board-nudge is a Claude Code PostToolUse hook for Bash. After
// a milestone command (git commit/push, gh pr create/merge/close) it reminds
// the model to bring .claude/tasks/*.json up to date, listing every
// unfinished task across all files so a stale row in a second file is seen.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxListed = 12

var milestone = regexp.MustCompile(`\bgit\s+(-C\s+\S+\s+)?(commit|push)\b|\bgh\s+pr\s+(create|merge|close)\b`)

type hookInput struct {
	ToolName  string `json:"tool_name"`
	Cwd       string `json:"cwd"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type taskFile struct {
	Tasks []struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	} `json:"tasks"`
}

type output struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

func isDone(status string) bool {
	head, _, _ := strings.Cut(strings.TrimSpace(status), " ")
	return head == "done" || head == "merged"
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

// unfinished returns one line per not-done task in dir/.claude/tasks/*.json.
func unfinished(dir string) []string {
	files, _ := filepath.Glob(filepath.Join(dir, ".claude", "tasks", "*.json"))
	var lines []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var tf taskFile
		if json.Unmarshal(b, &tf) != nil {
			continue
		}
		for _, t := range tf.Tasks {
			if !isDone(t.Status) {
				lines = append(lines, fmt.Sprintf("%s %s [%s] %s", filepath.Base(f), t.ID, t.Status, truncate(t.Title, 40)))
			}
		}
	}
	return lines
}

func run(in io.Reader, out io.Writer) {
	var h hookInput
	if json.NewDecoder(in).Decode(&h) != nil || h.ToolName != "Bash" || !milestone.MatchString(h.ToolInput.Command) {
		return
	}
	lines := unfinished(h.Cwd)
	if len(lines) == 0 {
		return
	}
	extra := ""
	if len(lines) > maxListed {
		extra = fmt.Sprintf("\n…他%d件", len(lines)-maxListed)
		lines = lines[:maxListed]
	}
	var o output
	o.HookSpecificOutput.HookEventName = "PostToolUse"
	o.HookSpecificOutput.AdditionalContext = "節目コマンド(commit/push/PR作成・merge・close)を実行した。次のタスクの status が実態と合っているか確認し、ずれていれば今すぐ .claude/tasks/*.json を更新すること(複数ファイルにまたがる):\n" + strings.Join(lines, "\n") + extra
	_ = json.NewEncoder(out).Encode(o)
}

func main() { run(os.Stdin, os.Stdout) }
