package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors/sourceinventory"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("generate-color-inventory", flag.ContinueOnError)
	check := flags.Bool("check", false, "check whether the generated inventory is up to date")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "generate-color-inventory: no positional arguments are accepted")
		return 2
	}

	root, err := palette.FindRepositoryRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate-color-inventory: %v\n", err)
		return 2
	}
	result, err := sourceinventory.Extract(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate-color-inventory: %v\n", err)
		return 2
	}
	generated, err := sourceinventory.Render(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate-color-inventory: %v\n", err)
		return 2
	}

	outputPath := filepath.Join(root, "scripts", "internal", "verifycolors", "inventory_generated.go")
	current, err := os.ReadFile(outputPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "generate-color-inventory: read output: %v\n", err)
		return 2
	}
	if bytes.Equal(current, generated) {
		fmt.Printf("Color inventory is up to date (%d pairs, %d coverage notes).\n", len(result.Pairs), len(result.CoverageNotes))
		return 0
	}
	if *check {
		fmt.Fprintf(os.Stderr, "generated color inventory is out of date: %s\n", outputPath)
		return 1
	}
	if err := os.WriteFile(outputPath, generated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "generate-color-inventory: write output: %v\n", err)
		return 2
	}
	fmt.Printf("Generated color inventory (%d pairs, %d coverage notes).\n", len(result.Pairs), len(result.CoverageNotes))
	return 0
}
