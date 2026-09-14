package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// hooksDir returns the repo-relative location verify-claude-hooks expects
// each hook binary to live at, given a fake repo root.
func hooksDir(repoRoot string) string {
	return filepath.Join(repoRoot, "claude", "hooks")
}

// setupFreshBinary places name's binary and a matching cmd/<name>/main.go
// source file under repoRoot, with the binary newer than the source, so
// checkBuildFreshness passes. Returns the binary path.
func setupFreshBinary(t *testing.T, repoRoot, name string) string {
	t.Helper()
	older := time.Now().Add(-time.Hour)
	newer := time.Now()

	srcDir := filepath.Join(repoRoot, "claude", "hooks", "cmd", name)
	srcFile := filepath.Join(srcDir, "main.go")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", srcDir, err)
	}
	if err := os.WriteFile(srcFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", srcFile, err)
	}
	if err := os.Chtimes(srcFile, older, older); err != nil {
		t.Fatalf("chtimes %s: %v", srcFile, err)
	}

	binPath := filepath.Join(hooksDir(repoRoot), name)
	writeExecutable(t, binPath)
	if err := os.Chtimes(binPath, newer, newer); err != nil {
		t.Fatalf("chtimes %s: %v", binPath, err)
	}
	return binPath
}

func validSettingsJSON(boundedBg, addNewline, herdrNotify string) string {
	return fmt.Sprintf(`{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": %q}]}
    ],
    "PostToolUse": [
      {"matcher": "Write|Edit|MultiEdit|NotebookEdit", "hooks": [{"type": "command", "command": %q}]}
    ],
    "Notification": [
      {"matcher": "idle_prompt", "hooks": [{"type": "command", "command": %q}]}
    ]
  }
}`, boundedBg, addNewline, herdrNotify)
}

func TestVerifyClaudeHooks(t *testing.T) {
	t.Run("all_hooks_registered_and_executable", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := setupFreshBinary(t, repoRoot, "bounded-background")
		addNewline := setupFreshBinary(t, repoRoot, "add_newline")
		herdrNotify := setupFreshBinary(t, repoRoot, "herdr-notify")
		setupFreshBinary(t, repoRoot, "bg-deadline")

		var buf bytes.Buffer
		code := verify(&buf, []byte(validSettingsJSON(boundedBg, addNewline, herdrNotify)), home, repoRoot)
		if code != exitSuccess {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitSuccess, buf.String())
		}
	})

	t.Run("missing_registration_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		settingsJSON := fmt.Sprintf(`{
  "hooks": {
    "PostToolUse": [
      {"matcher": "Write|Edit|MultiEdit|NotebookEdit", "hooks": [{"type": "command", "command": %q}]}
    ],
    "Notification": [
      {"matcher": "idle_prompt", "hooks": [{"type": "command", "command": %q}]}
    ]
  }
}`, addNewline, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(settingsJSON), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("bounded-background")) {
			t.Fatalf("output does not mention the missing hook:\n%s", buf.String())
		}
	})

	t.Run("registered_but_binary_missing", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := filepath.Join(hooksDir(repoRoot), "bounded-background") // not created
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(validSettingsJSON(boundedBg, addNewline, herdrNotify)), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("go build")) {
			t.Fatalf("output does not mention the rebuild command:\n%s", buf.String())
		}
	})

	t.Run("wrong_matcher_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := filepath.Join(hooksDir(repoRoot), "bounded-background")
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, boundedBg)
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		settingsJSON := fmt.Sprintf(`{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Write", "hooks": [{"type": "command", "command": %q}]}
    ],
    "PostToolUse": [
      {"matcher": "Write|Edit|MultiEdit|NotebookEdit", "hooks": [{"type": "command", "command": %q}]}
    ],
    "Notification": [
      {"matcher": "idle_prompt", "hooks": [{"type": "command", "command": %q}]}
    ]
  }
}`, boundedBg, addNewline, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(settingsJSON), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("matcher")) {
			t.Fatalf("output does not mention the matcher mismatch:\n%s", buf.String())
		}
	})

	t.Run("non_command_type_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := filepath.Join(hooksDir(repoRoot), "bounded-background")
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, boundedBg)
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		settingsJSON := fmt.Sprintf(`{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "prompt", "command": %q}]}
    ],
    "PostToolUse": [
      {"matcher": "Write|Edit|MultiEdit|NotebookEdit", "hooks": [{"type": "command", "command": %q}]}
    ],
    "Notification": [
      {"matcher": "idle_prompt", "hooks": [{"type": "command", "command": %q}]}
    ]
  }
}`, boundedBg, addNewline, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(settingsJSON), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("type")) {
			t.Fatalf("output does not mention the type mismatch:\n%s", buf.String())
		}
	})

	t.Run("stale_path_outside_repo_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		staleDir := t.TempDir()
		boundedBgStale := filepath.Join(staleDir, "bounded-background")
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, boundedBgStale)
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(validSettingsJSON(boundedBgStale, addNewline, herdrNotify)), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		expectedPath := filepath.Join(hooksDir(repoRoot), "bounded-background")
		if !bytes.Contains(buf.Bytes(), []byte(expectedPath)) {
			t.Fatalf("output does not mention the expected repo path %s:\n%s", expectedPath, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte(boundedBgStale)) {
			t.Fatalf("output does not mention the actual stale path %s:\n%s", boundedBgStale, buf.String())
		}
	})

	t.Run("command_with_extra_tokens_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := filepath.Join(hooksDir(repoRoot), "bounded-background")
		addNewline := filepath.Join(hooksDir(repoRoot), "add_newline")
		herdrNotify := filepath.Join(hooksDir(repoRoot), "herdr-notify")
		writeExecutable(t, boundedBg)
		writeExecutable(t, addNewline)
		writeExecutable(t, herdrNotify)

		boundedBgWithExtra := "echo " + boundedBg
		settingsJSON := validSettingsJSON(boundedBgWithExtra, addNewline, herdrNotify)

		var buf bytes.Buffer
		code := verify(&buf, []byte(settingsJSON), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte(boundedBg)) {
			t.Fatalf("output does not mention the expected command %s:\n%s", boundedBg, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte(boundedBgWithExtra)) {
			t.Fatalf("output does not mention the actual command %q:\n%s", boundedBgWithExtra, buf.String())
		}
	})

	t.Run("missing_bg_deadline_binary_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		boundedBg := setupFreshBinary(t, repoRoot, "bounded-background")
		addNewline := setupFreshBinary(t, repoRoot, "add_newline")
		herdrNotify := setupFreshBinary(t, repoRoot, "herdr-notify")
		// bg-deadline is intentionally not created.

		var buf bytes.Buffer
		code := verify(&buf, []byte(validSettingsJSON(boundedBg, addNewline, herdrNotify)), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("go build")) {
			t.Fatalf("output does not mention the rebuild command:\n%s", buf.String())
		}
	})

	t.Run("binary_older_than_source_is_reported", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		addNewline := setupFreshBinary(t, repoRoot, "add_newline")
		herdrNotify := setupFreshBinary(t, repoRoot, "herdr-notify")
		setupFreshBinary(t, repoRoot, "bg-deadline")

		// bounded-background: binary older than its source main.go.
		srcDir := filepath.Join(repoRoot, "claude", "hooks", "cmd", "bounded-background")
		srcFile := filepath.Join(srcDir, "main.go")
		if err := os.MkdirAll(srcDir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", srcDir, err)
		}
		if err := os.WriteFile(srcFile, []byte("package main\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", srcFile, err)
		}
		newer := time.Now()
		older := newer.Add(-time.Hour)
		if err := os.Chtimes(srcFile, newer, newer); err != nil {
			t.Fatalf("chtimes %s: %v", srcFile, err)
		}
		boundedBg := filepath.Join(hooksDir(repoRoot), "bounded-background")
		writeExecutable(t, boundedBg)
		if err := os.Chtimes(boundedBg, older, older); err != nil {
			t.Fatalf("chtimes %s: %v", boundedBg, err)
		}

		var buf bytes.Buffer
		code := verify(&buf, []byte(validSettingsJSON(boundedBg, addNewline, herdrNotify)), home, repoRoot)
		if code != exitPolicyNG {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitPolicyNG, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte("go build")) {
			t.Fatalf("output does not mention the rebuild command:\n%s", buf.String())
		}
	})

	t.Run("unreadable_settings_exits_2", func(t *testing.T) {
		home := t.TempDir()
		repoRoot := t.TempDir()
		missingPath := filepath.Join(home, "does-not-exist", "settings.json")

		var buf bytes.Buffer
		code := run(&buf, missingPath, home, repoRoot)
		if code != exitInput {
			t.Fatalf("code = %d, want %d\noutput:\n%s", code, exitInput, buf.String())
		}
	})
}
