package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixtureExitContracts(t *testing.T) {
	scriptsRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot := filepath.Dir(scriptsRoot)
	binary := filepath.Join(t.TempDir(), "verify-colors")
	build := exec.Command("go", "build", "-o", binary, "./cmd/verify-colors")
	build.Dir = scriptsRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build verify-colors: %v\n%s", err, output)
	}
	currentPalette := filepath.Join(repositoryRoot, "colors", "ghost-visor.toml")
	paletteData, err := os.ReadFile(currentPalette)
	if err != nil {
		t.Fatal(err)
	}
	cleanRoot := filepath.Join("testdata", "clean")
	unknownRoot := filepath.Join("testdata", "unknown")

	fixtures := []struct {
		name         string
		palette      string
		scanRoot     string
		extraArgs    []string
		wantExit     int
		wantContains []string
	}{
		{
			name:     "low contrast enforced failure",
			palette:  writePaletteFixture(t, paletteData, `main = "#CDE9F5"`, `main = "#202A42"`),
			scanRoot: cleanRoot,
			wantExit: exitPolicyNG,
			wantContains: []string{
				"[ENFORCED BELOW-AA]",
				"enforced_ng=",
			},
		},
		{
			name:     "subdued 8A93A6 regression",
			palette:  writePaletteFixture(t, paletteData, `subdued = "#949DB0"`, `subdued = "#8A93A6"`),
			scanRoot: cleanRoot,
			wantExit: exitPolicyNG,
			wantContains: []string{
				"[ENFORCED BELOW-AA] nvim.highlight.BlinkCmpKind",
				"foregrounds.subdued",
			},
		},
		{
			name:     "unknown HEX",
			palette:  currentPalette,
			scanRoot: unknownRoot,
			wantExit: exitPolicyNG,
			wantContains: []string{
				"[NG] fixture.lua: #123456",
			},
		},
		{
			name:     "intentional dim waiver",
			palette:  currentPalette,
			scanRoot: cleanRoot,
			wantExit: exitSuccess,
			wantContains: []string{
				"[WAIVED] nvim.highlight.Conceal",
				"intentionally dimmed",
			},
		},
		{
			name:     "environment dependent report only",
			palette:  currentPalette,
			scanRoot: cleanRoot,
			extraArgs: []string{
				"--environment-profile", "measured=#565F72",
			},
			wantExit: exitSuccess,
			wantContains: []string{
				"environment:measured",
				"does NOT mean every on-screen color is AA",
			},
		},
		{
			name:     "current palette",
			palette:  currentPalette,
			scanRoot: cleanRoot,
			wantExit: exitSuccess,
			wantContains: []string{
				"enforced_ng=0",
				"unclassified=0",
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			args := []string{
				"--palette", fixture.palette,
				"--scan-root", fixture.scanRoot,
				"--scan-file", "fixture.lua",
			}
			args = append(args, fixture.extraArgs...)
			output, exitCode := runBinary(binary, args)
			t.Logf("exit=%d", exitCode)
			if exitCode != fixture.wantExit {
				t.Fatalf("exit code = %d, want %d\n%s", exitCode, fixture.wantExit, output)
			}
			for _, expected := range fixture.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("output does not contain %q", expected)
				}
			}
		})
	}
}

func TestInputErrorUsesDistinctExitCode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"--palette", filepath.Join(t.TempDir(), "missing.toml"),
			"--scan-root", t.TempDir(),
		},
		&stdout,
		&stderr,
	)
	if exitCode != exitInput {
		t.Fatalf("exit code = %d, want %d; stderr=%s", exitCode, exitInput, stderr.String())
	}
	if !strings.Contains(stderr.String(), "contract/input error") {
		t.Fatalf("stderr does not identify contract/input error: %s", stderr.String())
	}
}

func writePaletteFixture(t *testing.T, source []byte, oldValue, newValue string) string {
	t.Helper()
	content := string(source)
	if strings.Count(content, oldValue) != 1 {
		t.Fatalf("palette fixture replacement %q matched %d times", oldValue, strings.Count(content, oldValue))
	}
	content = strings.Replace(content, oldValue, newValue, 1)
	path := filepath.Join(t.TempDir(), "ghost-visor.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runBinary(binary string, args []string) (string, int) {
	command := exec.Command(binary, args...)
	output, err := command.CombinedOutput()
	if err == nil {
		return string(output), exitSuccess
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return string(output) + "\n" + err.Error(), -1
	}
	return string(output), exitError.ExitCode()
}
