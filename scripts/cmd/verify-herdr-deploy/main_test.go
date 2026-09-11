package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/herdrdeploy"
)

// TestPrintStage1DiffGuidancePerDirection covers finding 1: stage 1's next
// action must follow the direction of the diff, not a single blanket
// "パッチ化して登録する". In particular, content that exists only on the
// dev-tree side isn't necessarily an unregistered new change — it can be a
// commit for something intentionally dropped from deployment — so that
// direction's guidance must not read as an unconditional instruction to
// re-register it.
func TestPrintStage1DiffGuidancePerDirection(t *testing.T) {
	cases := []struct {
		name            string
		diff            herdrdeploy.TreeDiff
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "only in config tree: guidance is to restore the commit or remove it from deployment",
			diff:         herdrdeploy.TreeDiff{OnlyInConfigTree: []string{"src/app/state.rs"}},
			wantContains: []string{"開発branchへ戻す", "配備から外す"},
		},
		{
			name: "only in dev tree: guidance covers both registering a new change and removing an intentionally-excluded one",
			diff: herdrdeploy.TreeDiff{OnlyInDevTree: []string{"src/ui/widgets.rs"}},
			wantContains: []string{
				"パッチ化して", "意図的に外した", "開発branch側のcommitを外す",
			},
			wantNotContains: []string{
				// must not read as an unconditional re-registration instruction
				"次の操作: パッチ化してhome-manager/patches/へ置きherdr.nixへ登録する\n",
			},
		},
		{
			name:         "changed: both candidates shown, a human decides",
			diff:         herdrdeploy.TreeDiff{Changed: []string{"src/main.rs"}},
			wantContains: []string{"人が判断"},
		},
		{
			name: "mixed: config-only and dev-only guidance both appear, independently",
			diff: herdrdeploy.TreeDiff{
				OnlyInConfigTree: []string{"src/app/mod.rs"},
				OnlyInDevTree:    []string{"src/ui/widgets.rs"},
			},
			wantContains: []string{"配備から外す", "開発branch側のcommitを外す"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			ok := printStage1Diff(&buf, tc.diff)
			if ok {
				t.Fatalf("printStage1Diff() = true, want false (all cases here are non-empty diffs)")
			}
			out := buf.String()
			for _, want := range tc.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q\n--- output ---\n%s", want, out)
				}
			}
			for _, notWant := range tc.wantNotContains {
				if strings.Contains(out, notWant) {
					t.Errorf("output must not contain %q\n--- output ---\n%s", notWant, out)
				}
			}
		})
	}
}

// TestEvaluateStage3DoesNotLeakProcessArgs covers finding 3: judgment needs
// only pid, server/other classification, and the resolved store path.
// Full argv must never reach stdout, since this tool's output is meant to
// be pasted into a report.
func TestEvaluateStage3DoesNotLeakProcessArgs(t *testing.T) {
	const secret = "SECRET_TOKEN_9f8e7d6c5b4a"
	// Pids far outside any real process table: lsof against them fails,
	// exercising exactly the branch where a naive fix might still
	// interpolate Args into the message.
	servers := []herdrdeploy.Process{{PID: 999999999, Args: "herdr server --token " + secret}}
	others := []herdrdeploy.Process{{PID: 999999998, Args: "herdr --auth " + secret}}

	var buf bytes.Buffer
	evaluateStage3(&buf, "/nix/store/xxxx-herdr-0.8.0", servers, others)
	out := buf.String()

	if strings.Contains(out, secret) {
		t.Fatalf("stdout leaked the secret value:\n%s", out)
	}
	if strings.Contains(out, "--token") || strings.Contains(out, "--auth") {
		t.Fatalf("stdout leaked argv flags:\n%s", out)
	}
}

// TestRunStage1FailsOnUnstableOverride covers finding 4: if the same
// --override-input eval returns two different store paths across the two
// calls, the dev tree changed between them and neither result can be
// trusted — that must be a hard failure, not a warning that still lets the
// comparison (and a possible exit 0) go through.
func TestRunStage1FailsOnUnstableOverride(t *testing.T) {
	calls := 0
	fake := func(name string, args ...string) (string, string, error) {
		calls++
		switch {
		case name == "nix" && len(args) > 0 && args[0] == "build":
			return "/nix/store/does-not-need-to-exist-config", "", nil
		case name == "nix" && len(args) > 0 && args[0] == "eval":
			if calls == 2 {
				return "/nix/store/aaaa-source", "", nil
			}
			return "/nix/store/bbbb-source", "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", "", nil
		}
	}

	var buf bytes.Buffer
	ok := runStage1(&buf, "/fake/dev-tree", "/fake/flake", fake)
	if ok {
		t.Fatalf("runStage1() = true, want false when the two eval calls disagree")
	}
	out := buf.String()
	if !strings.Contains(out, "ng:") {
		t.Errorf("output doesn't report ng:\n%s", out)
	}
	if strings.Contains(out, "警告:") {
		t.Errorf("instability must be a hard failure, not just a warning:\n%s", out)
	}
}

// TestRunRejectsPathsWithNixSpecialChars covers finding 7: a --dev-tree or
// --flake path containing characters that nix installable syntax or
// git+file: URL syntax treat specially must be rejected up front, rather
// than silently misparsed later.
func TestRunRejectsPathsWithNixSpecialChars(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"dev-tree contains '#'", []string{"--dev-tree", "/tmp/weird#tree", "--flake", "/tmp/flake"}},
		{"flake contains '?'", []string{"--dev-tree", "/tmp/tree", "--flake", "/tmp/weird?flake"}},
		{"dev-tree contains a space", []string{"--dev-tree", "/tmp/weird tree", "--flake", "/tmp/flake"}},
		{"flake contains '%'", []string{"--dev-tree", "/tmp/tree", "--flake", "/tmp/weird%2f"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tc.args, &stdout, &stderr)
			if code != exitInput {
				t.Fatalf("exit = %d, want %d (stdout=%q stderr=%q)", code, exitInput, stdout.String(), stderr.String())
			}
		})
	}
}
