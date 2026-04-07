//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dirs := []string{
		filepath.Join(home, ".config/codex/skills"),
		filepath.Join(home, ".config/claude/skills"),
		filepath.Join(home, ".config/claude/agents"),
	}

	broken := []string{}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.Type()&os.ModeSymlink == 0 {
				continue
			}
			target, err := os.Readlink(path)
			if err != nil {
				broken = append(broken, fmt.Sprintf("  %s (unreadable link)", path))
				continue
			}
			resolvedTarget := target
			if !filepath.IsAbs(target) {
				resolvedTarget = filepath.Join(filepath.Dir(path), target)
			}
			if _, err := os.Stat(resolvedTarget); err != nil {
				broken = append(broken, fmt.Sprintf("  %s -> %s (target not found)", path, target))
			}
		}
	}

	if len(broken) == 0 {
		fmt.Println("OK: no broken symlinks found")
		return
	}

	fmt.Printf("FAIL: %d broken symlink(s) found:\n", len(broken))
	for _, b := range broken {
		fmt.Println(b)
	}
	os.Exit(1)
}
