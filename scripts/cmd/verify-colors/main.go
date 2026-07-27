package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

const (
	exitSuccess  = 0
	exitPolicyNG = 1
	exitInput    = 2
)

type repeatedFlag []string

func (values *repeatedFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *repeatedFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("verify-colors", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var palettePath string
	var scanRoot string
	var scanFiles repeatedFlag
	var environmentValues repeatedFlag
	flags.StringVar(&palettePath, "palette", "", "path to a color palette TOML")
	flags.StringVar(&scanRoot, "scan-root", "", "root directory for HEX literal scans")
	flags.Var(&scanFiles, "scan-file", "relative file to scan (repeatable; defaults to the legacy validation list)")
	flags.Var(&environmentValues, "environment-profile", "additional name=#RRGGBB ambient profile (repeatable)")
	if err := flags.Parse(args); err != nil {
		return exitInput
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "verify-colors: unexpected positional arguments: %s\n", strings.Join(flags.Args(), " "))
		return exitInput
	}

	repositoryRoot := ""
	var err error
	if palettePath == "" || scanRoot == "" {
		repositoryRoot, err = palette.FindRepositoryRoot()
		if err != nil {
			fmt.Fprintf(stderr, "verify-colors: resolve repository root: %v\n", err)
			return exitInput
		}
	}
	if palettePath == "" {
		palettePath = filepath.Join(repositoryRoot, "colors", "ghost-visor.toml")
	} else if !filepath.IsAbs(palettePath) {
		palettePath, err = filepath.Abs(palettePath)
		if err != nil {
			fmt.Fprintf(stderr, "verify-colors: resolve palette path: %v\n", err)
			return exitInput
		}
	}
	if scanRoot == "" {
		scanRoot = repositoryRoot
	} else if !filepath.IsAbs(scanRoot) {
		scanRoot, err = filepath.Abs(scanRoot)
		if err != nil {
			fmt.Fprintf(stderr, "verify-colors: resolve scan root: %v\n", err)
			return exitInput
		}
	}
	environments, err := parseEnvironments(environmentValues)
	if err != nil {
		fmt.Fprintf(stderr, "verify-colors: %v\n", err)
		return exitInput
	}
	verification, err := verifycolors.Verify(verifycolors.VerifyConfig{
		PalettePath:  palettePath,
		ScanRoot:     scanRoot,
		ScanFiles:    scanFiles,
		Environments: environments,
	})
	if err != nil {
		fmt.Fprintf(stderr, "verify-colors: contract/input error: %v\n", err)
		return exitInput
	}
	verifycolors.WriteReport(stdout, verification)
	if verification.PolicyFailures() > 0 {
		return exitPolicyNG
	}
	return exitSuccess
}

func parseEnvironments(values []string) ([]verifycolors.EnvironmentProfile, error) {
	environments := make([]verifycolors.EnvironmentProfile, 0, len(values))
	for _, value := range values {
		name, color, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(name) == "" || strings.TrimSpace(color) == "" {
			return nil, fmt.Errorf("environment profile must be name=#RRGGBB, got %q", value)
		}
		environments = append(environments, verifycolors.EnvironmentProfile{
			Name:  strings.TrimSpace(name),
			Color: strings.TrimSpace(color),
		})
	}
	return environments, nil
}
