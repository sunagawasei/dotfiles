package palette

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type LoadStage string

const (
	LoadStageRead  LoadStage = "read"
	LoadStageParse LoadStage = "parse"
)

type LoadError struct {
	Stage LoadStage
	Err   error
}

func (e *LoadError) Error() string {
	return e.Err.Error()
}

func (e *LoadError) Unwrap() error {
	return e.Err
}

func Load[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &LoadError{Stage: LoadStageRead, Err: err}
	}
	var value T
	if err := toml.Unmarshal(data, &value); err != nil {
		return nil, &LoadError{Stage: LoadStageParse, Err: err}
	}
	return &value, nil
}

func FindRepositoryRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("find repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func ResolvePath(root string, args []string) (string, error) {
	if len(args) == 1 {
		path := args[0]
		if !filepath.IsAbs(path) {
			if _, err := os.Stat(path); err == nil {
				absolute, err := filepath.Abs(path)
				if err != nil {
					return "", fmt.Errorf("resolve palette path: %w", err)
				}
				return absolute, nil
			}
			path = filepath.Join(root, path)
		}
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("palette %q: %w", path, err)
		}
		return filepath.Clean(path), nil
	}

	matches, err := filepath.Glob(filepath.Join(root, "colors", "*.toml"))
	if err != nil {
		return "", fmt.Errorf("find default palette: %w", err)
	}
	sort.Strings(matches)
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one colors/*.toml palette, found %d; pass the path explicitly", len(matches))
	}
	return matches[0], nil
}

func ResolveDefaultPath(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	root, err := FindRepositoryRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "colors", "ghost-visor.toml"), nil
}
