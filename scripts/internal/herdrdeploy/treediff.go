package herdrdeploy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// TreeDiff is a bidirectional file-level diff between two directory trees.
// Stage 1 uses it to compare the config-side tree (upstream v0.8.0 with the
// registered patches applied) against the same fileset applied to the dev
// tree, so it holds no notion of which side is "correct" — only which
// relative paths differ and how.
type TreeDiff struct {
	OnlyInConfigTree []string // present in the config-side tree, absent from the dev tree
	OnlyInDevTree    []string // present in the dev tree, absent from the config-side tree
	Changed          []string // present in both, content/target/executable-bit differs
}

func (d TreeDiff) Empty() bool {
	return len(d.OnlyInConfigTree) == 0 && len(d.OnlyInDevTree) == 0 && len(d.Changed) == 0
}

// DiffShape classifies the overall shape of a TreeDiff so callers can print
// a one-line summary before the file lists.
type DiffShape string

const (
	ShapeNoDiff        DiffShape = "no-diff"
	ShapeOnlyInConfig  DiffShape = "only-in-config"   // registered patches produce files/content the dev tree doesn't have
	ShapeOnlyInDevTree DiffShape = "only-in-dev-tree" // dev tree has files/content not covered by any registered patch
	ShapeChangedOnly   DiffShape = "changed-only"
	ShapeMixed         DiffShape = "mixed"
)

func (d TreeDiff) Shape() DiffShape {
	hasOnlyConfig := len(d.OnlyInConfigTree) > 0
	hasOnlyDev := len(d.OnlyInDevTree) > 0
	hasChanged := len(d.Changed) > 0
	switch {
	case !hasOnlyConfig && !hasOnlyDev && !hasChanged:
		return ShapeNoDiff
	case hasOnlyConfig && !hasOnlyDev && !hasChanged:
		return ShapeOnlyInConfig
	case !hasOnlyConfig && hasOnlyDev && !hasChanged:
		return ShapeOnlyInDevTree
	case !hasOnlyConfig && !hasOnlyDev && hasChanged:
		return ShapeChangedOnly
	default:
		return ShapeMixed
	}
}

// fileState is the comparable fingerprint of one relative path: a regular
// file's content hash and executable bit, or a symlink's target string. Two
// paths with different fileState (including a kind swap between file and
// symlink) count as Changed.
type fileState struct {
	kind       string // "file", "symlink", or "other:<type>"
	executable bool   // meaningful only when kind == "file"
	payload    string // content hash (file) or link target (symlink)
}

// DiffTrees walks two directory trees (typically /nix/store paths) and
// returns every relative path that differs between them, in either
// direction. It has no notion of "which files matter" — that's decided
// entirely by what the fileset already put into each tree before this runs.
func DiffTrees(configRoot, devRoot string) (TreeDiff, error) {
	configFiles, err := scanTree(configRoot)
	if err != nil {
		return TreeDiff{}, fmt.Errorf("scan config tree %s: %w", configRoot, err)
	}
	devFiles, err := scanTree(devRoot)
	if err != nil {
		return TreeDiff{}, fmt.Errorf("scan dev tree %s: %w", devRoot, err)
	}

	var diff TreeDiff
	for rel, state := range configFiles {
		devState, ok := devFiles[rel]
		if !ok {
			diff.OnlyInConfigTree = append(diff.OnlyInConfigTree, rel)
			continue
		}
		if devState != state {
			diff.Changed = append(diff.Changed, rel)
		}
	}
	for rel := range devFiles {
		if _, ok := configFiles[rel]; !ok {
			diff.OnlyInDevTree = append(diff.OnlyInDevTree, rel)
		}
	}
	sort.Strings(diff.OnlyInConfigTree)
	sort.Strings(diff.OnlyInDevTree)
	sort.Strings(diff.Changed)
	return diff, nil
}

func scanTree(root string) (map[string]fileState, error) {
	states := make(map[string]fileState)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info() // Lstat-based: reports the symlink itself, not its target
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			states[rel] = fileState{kind: "symlink", payload: target}
		case info.Mode().IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(data)
			states[rel] = fileState{
				kind:       "file",
				executable: info.Mode().Perm()&0o100 != 0,
				payload:    hex.EncodeToString(sum[:]),
			}
		default:
			// Unexpected in a source tree (device, socket, ...): still record
			// something so it isn't silently dropped from the comparison.
			states[rel] = fileState{kind: "other:" + info.Mode().Type().String()}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return states, nil
}
