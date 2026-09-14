package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func buildInput(t *testing.T, toolName string, toolInput map[string]any) []byte {
	t.Helper()
	in := map[string]any{"tool_name": toolName}
	if toolInput != nil {
		in["tool_input"] = toolInput
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	return data
}

func workingResolver(path string) func() bgDeadlineResolution {
	return func() bgDeadlineResolution { return bgDeadlineResolution{Path: path} }
}

// errReader always fails, simulating an unreadable stdin.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

// failingWriter always fails, simulating an unwritable stdout.
type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestDecide(t *testing.T) {
	t.Run("foreground_passes_through_untouched", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{"command": "echo hi"})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output for foreground command, got %+v", out)
		}
	})

	t.Run("background_command_is_wrapped", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 100",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		if out.HookSpecificOutput.HookEventName != "PreToolUse" {
			t.Fatalf("hookEventName = %q, want PreToolUse", out.HookSpecificOutput.HookEventName)
		}
		cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
		wantPrefix := "'/opt/bg-deadline' 600 /bin/zsh -c "
		if !strings.HasPrefix(cmd, wantPrefix) {
			t.Fatalf("command = %q, want prefix %q", cmd, wantPrefix)
		}
	})

	t.Run("unknown_tool_input_fields_are_preserved", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 5",
			"run_in_background": true,
			"description":       "sleep a bit",
			"future_field":      42.0,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		if got := out.HookSpecificOutput.UpdatedInput["description"]; got != "sleep a bit" {
			t.Fatalf("description = %v, want %q", got, "sleep a bit")
		}
		if got := out.HookSpecificOutput.UpdatedInput["future_field"]; got != 42.0 {
			t.Fatalf("future_field = %v, want 42", got)
		}
	})

	t.Run("deadline_defaults_to_600", func(t *testing.T) {
		if got := computeDeadlineSeconds(map[string]any{}); got != 600 {
			t.Fatalf("computeDeadlineSeconds() = %d, want 600", got)
		}
	})

	t.Run("deadline_uses_timeout_rounded_up", func(t *testing.T) {
		cases := []struct {
			timeout float64
			want    int
		}{
			{1999, 2},
			{5000, 5},
		}
		for _, c := range cases {
			got := computeDeadlineSeconds(map[string]any{"timeout": c.timeout})
			if got != c.want {
				t.Fatalf("computeDeadlineSeconds(timeout=%v) = %d, want %d", c.timeout, got, c.want)
			}
		}
	})

	t.Run("deadline_clamps_to_600", func(t *testing.T) {
		if got := computeDeadlineSeconds(map[string]any{"timeout": 900000.0}); got != 600 {
			t.Fatalf("computeDeadlineSeconds(timeout=900000) = %d, want 600", got)
		}
	})

	t.Run("single_quotes_in_command_are_escaped", func(t *testing.T) {
		original := `echo 'hello world'`
		wrapped := wrapCommand("/opt/bg-deadline", 600, original)
		want := `'/opt/bg-deadline' 600 /bin/zsh -c 'echo '\''hello world'\'''`
		if wrapped != want {
			t.Fatalf("wrapCommand() = %q, want %q", wrapped, want)
		}
	})

	t.Run("already_wrapped_command_passes_through", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' 600 /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output for already-wrapped command, got %+v", out)
		}
	})

	t.Run("missing_bg_deadline_denies", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 100",
			"run_in_background": true,
		})
		resolver := func() bgDeadlineResolution {
			return bgDeadlineResolution{Path: "/opt/bg-deadline", Err: os.ErrNotExist}
		}
		out := decide(input, resolver)
		if out == nil {
			t.Fatalf("expected deny output, got nil")
		}
		if out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("permissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
		}
		if !strings.Contains(out.HookSpecificOutput.PermissionDecisionReason, "go build") {
			t.Fatalf("reason = %q, want it to mention go build", out.HookSpecificOutput.PermissionDecisionReason)
		}
	})

	t.Run("malformed_json_denies", func(t *testing.T) {
		out := decide([]byte("{not json"), workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected deny output for malformed JSON, got nil")
		}
		if out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("permissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
		}
	})

	t.Run("unreadable_stdin_denies", func(t *testing.T) {
		out := decideFromReader(errReader{err: errors.New("boom")}, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected deny output for unreadable stdin, got nil")
		}
		if out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("permissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
		}
	})

	t.Run("non_bool_run_in_background_denies", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 5",
			"run_in_background": "true",
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected deny output, got nil")
		}
		if out.HookSpecificOutput.PermissionDecision != "deny" {
			t.Fatalf("permissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
		}
	})

	t.Run("missing_run_in_background_passes_through", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{"command": "sleep 5"})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output, got %+v", out)
		}
	})

	t.Run("already_wrapped_oversized_deadline_is_clamped", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' 999999 /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
		want := "'/opt/bg-deadline' 600 /bin/zsh -c 'sleep 5'"
		if cmd != want {
			t.Fatalf("command = %q, want %q", cmd, want)
		}
	})

	t.Run("already_wrapped_valid_deadline_passes_through", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' 300 /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output for valid deadline, got %+v", out)
		}
	})

	t.Run("already_wrapped_invalid_deadline_is_replaced", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' abc /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
		want := "'/opt/bg-deadline' 600 /bin/zsh -c 'sleep 5'"
		if cmd != want {
			t.Fatalf("command = %q, want %q", cmd, want)
		}
	})

	t.Run("bg_deadline_path_with_space_is_quoted", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 5",
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/with space/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
		wantPrefix := "'/opt/with space/bg-deadline' 600 /bin/zsh -c "
		if !strings.HasPrefix(cmd, wantPrefix) {
			t.Fatalf("command = %q, want prefix %q", cmd, wantPrefix)
		}
	})

	t.Run("already_wrapped_deadline_is_capped_by_tool_timeout", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' 600 /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
			"timeout":           5000.0,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out == nil {
			t.Fatalf("expected rewrite output, got nil")
		}
		cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
		want := "'/opt/bg-deadline' 5 /bin/zsh -c 'sleep 5'"
		if cmd != want {
			t.Fatalf("command = %q, want %q", cmd, want)
		}
	})

	t.Run("already_wrapped_deadline_within_tool_timeout_passes_through", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "'/opt/bg-deadline' 300 /bin/zsh -c 'sleep 5'",
			"run_in_background": true,
			"timeout":           600000.0,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output, got %+v", out)
		}
	})

	t.Run("canonical_wrapper_is_recognized", func(t *testing.T) {
		wrapped := wrapCommand("/opt/bg-deadline", 600, "sleep 5")
		input := buildInput(t, "Bash", map[string]any{
			"command":           wrapped,
			"run_in_background": true,
		})
		out := decide(input, workingResolver("/opt/bg-deadline"))
		if out != nil {
			t.Fatalf("expected no output for canonical wrapped command, got %+v", out)
		}
	})

	trailingCases := []struct {
		name    string
		command string
	}{
		{"trailing_command_after_wrapper_is_rewrapped", "'/opt/bg-deadline' 5 /bin/zsh -c 'true'; sleep 9999"},
		{"trailing_and_after_wrapper_is_rewrapped", "'/opt/bg-deadline' 5 /bin/zsh -c 'true' && sleep 9999"},
		{"trailing_newline_after_wrapper_is_rewrapped", "'/opt/bg-deadline' 5 /bin/zsh -c 'true'\nsleep 9999"},
		{"trailing_pipe_after_wrapper_is_rewrapped", "'/opt/bg-deadline' 5 /bin/zsh -c 'true' | sleep 9999"},
	}
	for _, c := range trailingCases {
		t.Run(c.name, func(t *testing.T) {
			input := buildInput(t, "Bash", map[string]any{
				"command":           c.command,
				"run_in_background": true,
			})
			out := decide(input, workingResolver("/opt/bg-deadline"))
			if out == nil {
				t.Fatalf("expected rewrite output for trailing command, got nil")
			}
			cmd, _ := out.HookSpecificOutput.UpdatedInput["command"].(string)
			want := wrapCommand("/opt/bg-deadline", 600, c.command)
			if cmd != want {
				t.Fatalf("command = %q, want %q", cmd, want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Run("write_failure_exits_nonzero", func(t *testing.T) {
		input := buildInput(t, "Bash", map[string]any{
			"command":           "sleep 5",
			"run_in_background": true,
		})
		var errBuf bytes.Buffer
		code := run(bytes.NewReader(input), failingWriter{err: errors.New("broken pipe")}, &errBuf, workingResolver("/opt/bg-deadline"))
		if code != 2 {
			t.Fatalf("exit code = %d, want 2", code)
		}
		if errBuf.Len() == 0 {
			t.Fatalf("expected a reason written to stderr, got empty output")
		}
	})
}
