package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) string {
	dir := t.TempDir()
	td := filepath.Join(dir, ".claude", "tasks")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	a := `{"tasks":[{"id":"A2","title":"実装","status":"reviewing"},{"id":"A1","title":"x","status":"done"}]}`
	b := `{"tasks":[{"id":"T5","title":"y","status":"merged"}]}`
	os.WriteFile(filepath.Join(td, "a.json"), []byte(a), 0o644)
	os.WriteFile(filepath.Join(td, "b.json"), []byte(b), 0o644)
	return dir
}

func call(dir, cmd string) string {
	var out bytes.Buffer
	in := `{"tool_name":"Bash","cwd":"` + dir + `","tool_input":{"command":"` + cmd + `"}}`
	run(strings.NewReader(in), &out)
	return out.String()
}

func TestNudge(t *testing.T) {
	dir := setup(t)
	for _, c := range []string{"gh pr merge 427 --squash", "git commit -m x", "git push -u origin b", "gh pr create --base main"} {
		got := call(dir, c)
		if !strings.Contains(got, "a.json A2 [reviewing]") || strings.Contains(got, "A1") || strings.Contains(got, "T5") {
			t.Errorf("%s: %s", c, got)
		}
	}
	if got := call(dir, "git commit"); !strings.Contains(got, "A2 [reviewing] ⚠") {
		t.Errorf("unknown status not flagged: %s", got)
	}
	for _, c := range []string{"git status", "gh pr view 1", "ls"} {
		if got := call(dir, c); got != "" {
			t.Errorf("%s: unexpected %s", c, got)
		}
	}
	if got := call(t.TempDir(), "git commit"); got != "" {
		t.Errorf("no tasks dir: %s", got)
	}
}
