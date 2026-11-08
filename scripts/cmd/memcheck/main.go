// Command memcheck checks the consistency of one Claude Code memory directory.
// The wrapup and claude-audit skills both call it, so the rules for the
// MEMORY.md index live here and nowhere else.
//
// Usage:
//
//	memcheck check <dir>   index vs files, size limit; exit 1 when action is needed
//	memcheck pairs <dir>   unresolved [[link]] pairs as "file|name", sorted
//
// Exit codes: 0 nothing to act on, 1 action needed, 2 usage, 3 invalid target
// (nothing was checked), 4 read error (nothing was checked).
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const (
	exitOK       = 0
	exitProblem  = 1
	exitUsage    = 2
	exitInvalid  = 3
	exitReadFail = 4

	// MEMORY.md is loaded only up to 200 lines or 25600 bytes; "near" is the
	// point where the skills stop trusting an index gap to be a plain omission.
	nearLines = 180
	nearBytes = 23000
	maxLines  = 200
	maxBytes  = 25600
)

var (
	indexEntry = regexp.MustCompile(`^- \[[^\]]*\]\(([^)]*)\)`)
	wikiLink   = regexp.MustCompile(`\[\[([A-Za-z0-9_-]+)\]\]`)
)

type report struct {
	NotIndexed  []string
	Dangling    []string
	Duplicate   []string
	Unresolved  []string // "file|name"
	Undecided   []string // gaps that may be intentional; see check
	Lines       int
	Bytes       int
	HasArchived bool
}

func (r report) near() bool { return r.Lines >= nearLines || r.Bytes >= nearBytes }
func (r report) over() bool { return r.Lines >= maxLines || r.Bytes >= maxBytes }

// problem reports whether the caller has something to act on. Undecided gaps
// and links that were already unresolved are information only.
func (r report) problem() bool {
	return len(r.NotIndexed)+len(r.Dangling)+len(r.Duplicate) > 0 || r.near()
}

// validateTarget accepts only <home>/.config/claude/projects/*/memory that
// holds a MEMORY.md, so a typo cannot point the check at an arbitrary tree.
func validateTarget(dir, home string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", errors.New("not an absolute path")
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	base, err := filepath.EvalSymlinks(filepath.Join(home, ".config", "claude", "projects"))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, real)
	if err != nil {
		return "", err
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) != 2 || parts[1] != "memory" || parts[0] == ".." {
		return "", fmt.Errorf("not under %s/*/memory", base)
	}
	if _, err := os.Stat(filepath.Join(real, "MEMORY.md")); err != nil {
		return "", errors.New("no MEMORY.md")
	}
	return real, nil
}

func memoryFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func check(dir string) (report, error) {
	var r report
	index, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		return r, err
	}
	r.Lines = bytes.Count(index, []byte("\n"))
	r.Bytes = len(index)

	archived := map[string]bool{}
	if data, err := os.ReadFile(filepath.Join(dir, "ARCHIVED")); err == nil {
		r.HasArchived = true
		for line := range strings.SplitSeq(string(data), "\n") {
			archived[strings.TrimSpace(line)] = true
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return r, err
	}

	files, err := memoryFiles(dir)
	if err != nil {
		return r, err
	}
	text := string(index)
	for _, f := range files {
		if f == "MEMORY.md" || archived[f] || strings.Contains(text, "("+f+")") {
			continue
		}
		// Without ARCHIVED, a gap near the limit may be an intentional omission.
		if !r.HasArchived && r.near() {
			r.Undecided = append(r.Undecided, f)
		} else {
			r.NotIndexed = append(r.NotIndexed, f)
		}
	}

	seen := map[string]int{}
	for line := range strings.SplitSeq(text, "\n") {
		m := indexEntry.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		seen[m[1]]++
		if seen[m[1]] == 2 {
			r.Duplicate = append(r.Duplicate, m[1])
		}
		if _, err := os.Stat(filepath.Join(dir, m[1])); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return r, err
			}
			if !slices.Contains(r.Dangling, m[1]) {
				r.Dangling = append(r.Dangling, m[1])
			}
		}
	}

	r.Unresolved, err = unresolved(dir, files)
	return r, err
}

// unresolved returns the sorted "file|name" pairs whose [[name]] has no
// name.md. Pairs, not a count: fixing one dangling link while breaking another
// leaves the count unchanged.
func unresolved(dir string, files []string) ([]string, error) {
	set := map[string]bool{}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		for _, m := range wikiLink.FindAllStringSubmatch(string(data), -1) {
			if _, err := os.Stat(filepath.Join(dir, m[1]+".md")); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return nil, err
				}
				set[f+"|"+m[1]] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	slices.Sort(out)
	return out, nil
}

func printReport(w io.Writer, r report) {
	for _, f := range r.NotIndexed {
		fmt.Fprintln(w, "NOT_INDEXED", f)
	}
	for _, f := range r.Dangling {
		fmt.Fprintln(w, "DANGLING", f)
	}
	for _, f := range r.Duplicate {
		fmt.Fprintln(w, "DUPLICATE", f)
	}
	switch {
	case r.over():
		fmt.Fprintf(w, "OVER_LIMIT lines=%d bytes=%d\n", r.Lines, r.Bytes)
	case r.near():
		fmt.Fprintf(w, "NEAR_LIMIT lines=%d bytes=%d\n", r.Lines, r.Bytes)
	}
	if len(r.Undecided) > 0 {
		fmt.Fprintf(w, "UNDECIDED %d (no ARCHIVED and near the limit: omissions cannot be told from leaks)\n", len(r.Undecided))
	}
	for _, p := range r.Unresolved {
		fmt.Fprintln(w, "UNRESOLVED", p)
	}
	fmt.Fprintf(w, "SUMMARY lines=%d bytes=%d archived=%t\n", r.Lines, r.Bytes, r.HasArchived)
}

func run(args []string, home string, stdout, stderr io.Writer) int {
	if len(args) != 2 || (args[0] != "check" && args[0] != "pairs") {
		fmt.Fprintln(stderr, "usage: memcheck check <dir> | pairs <dir>")
		return exitUsage
	}
	dir, err := validateTarget(args[1], home)
	if err != nil {
		fmt.Fprintf(stderr, "memcheck: invalid target %s: %v\n", args[1], err)
		return exitInvalid
	}
	if args[0] == "pairs" {
		files, err := memoryFiles(dir)
		var pairs []string
		if err == nil {
			pairs, err = unresolved(dir, files)
		}
		if err != nil {
			fmt.Fprintln(stderr, "memcheck:", err)
			return exitReadFail
		}
		for _, p := range pairs {
			fmt.Fprintln(stdout, p)
		}
		return exitOK
	}
	r, err := check(dir)
	if err != nil {
		fmt.Fprintln(stderr, "memcheck:", err)
		return exitReadFail
	}
	printReport(stdout, r)
	if r.problem() {
		return exitProblem
	}
	return exitOK
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "memcheck:", err)
		os.Exit(exitReadFail)
	}
	os.Exit(run(os.Args[1:], home, os.Stdout, os.Stderr))
}
