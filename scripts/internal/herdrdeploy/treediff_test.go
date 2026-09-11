package herdrdeploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTreeDiffShape(t *testing.T) {
	cases := []struct {
		name string
		diff TreeDiff
		want DiffShape
	}{
		{
			name: "no diff",
			diff: TreeDiff{},
			want: ShapeNoDiff,
		},
		{
			name: "changed only",
			diff: TreeDiff{Changed: []string{"src/main.rs"}},
			want: ShapeChangedOnly,
		},
		{
			name: "only in config tree (dev tree missing a registered patch's effect)",
			diff: TreeDiff{OnlyInConfigTree: []string{"src/app/state.rs"}},
			want: ShapeOnlyInConfig,
		},
		{
			name: "only in dev tree (a dev commit with no registered patch yet)",
			diff: TreeDiff{OnlyInDevTree: []string{"src/ui/newmod/mod.rs"}},
			want: ShapeOnlyInDevTree,
		},
		{
			name: "mixed: changed plus only-in-config",
			diff: TreeDiff{Changed: []string{"src/main.rs"}, OnlyInConfigTree: []string{"src/app/state.rs"}},
			want: ShapeMixed,
		},
		{
			name: "mixed: only-in-config plus only-in-dev-tree",
			diff: TreeDiff{OnlyInConfigTree: []string{"a.rs"}, OnlyInDevTree: []string{"b.rs"}},
			want: ShapeMixed,
		},
		{
			name: "mixed: all three populated",
			diff: TreeDiff{OnlyInConfigTree: []string{"a.rs"}, OnlyInDevTree: []string{"b.rs"}, Changed: []string{"c.rs"}},
			want: ShapeMixed,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.diff.Shape(); got != tc.want {
				t.Fatalf("Shape() = %v, want %v", got, tc.want)
			}
			wantEmpty := tc.want == ShapeNoDiff
			if tc.diff.Empty() != wantEmpty {
				t.Fatalf("Empty() = %v, want %v", tc.diff.Empty(), wantEmpty)
			}
		})
	}
}

// TestDiffTrees exercises the actual filesystem walk, in particular that a
// file inside a brand-new subdirectory on one side is still detected: an
// earlier design that special-cased known directories missed exactly this.
func TestDiffTrees(t *testing.T) {
	configRoot := t.TempDir()
	devRoot := t.TempDir()

	writeFile(t, configRoot, "src/main.rs", "fn main() {}\n")
	writeFile(t, devRoot, "src/main.rs", "fn main() {}\n")

	writeFile(t, configRoot, "src/app/state.rs", "// config version\n")
	writeFile(t, devRoot, "src/app/state.rs", "// dev version\n")

	writeFile(t, configRoot, "src/ui/panes.rs", "// only in config\n")

	writeFile(t, devRoot, "src/ui/newmod/mod.rs", "// only in dev, new subdirectory\n")

	diff, err := DiffTrees(configRoot, devRoot)
	if err != nil {
		t.Fatalf("DiffTrees: %v", err)
	}

	assertContainsExactly(t, "Changed", diff.Changed, []string{"src/app/state.rs"})
	assertContainsExactly(t, "OnlyInConfigTree", diff.OnlyInConfigTree, []string{"src/ui/panes.rs"})
	assertContainsExactly(t, "OnlyInDevTree", diff.OnlyInDevTree, []string{"src/ui/newmod/mod.rs"})
	if got := diff.Shape(); got != ShapeMixed {
		t.Fatalf("Shape() = %v, want %v", got, ShapeMixed)
	}
}

// TestDiffTreesSymlinksAndMode covers the cases a plain content-hash
// comparison misses: symlinks (added, or present on both sides with
// different targets), a file/symlink kind swap at the same path, and a
// same-content file whose only difference is the executable bit.
func TestDiffTreesSymlinksAndMode(t *testing.T) {
	t.Run("symlink added only in dev tree", func(t *testing.T) {
		configRoot, devRoot := t.TempDir(), t.TempDir()
		writeFile(t, configRoot, "keep.txt", "same\n")
		writeFile(t, devRoot, "keep.txt", "same\n")
		mustSymlink(t, devRoot, "bin/tool", "../lib/tool-real")

		diff, err := DiffTrees(configRoot, devRoot)
		if err != nil {
			t.Fatalf("DiffTrees: %v", err)
		}
		assertContainsExactly(t, "OnlyInDevTree", diff.OnlyInDevTree, []string{"bin/tool"})
		if got := diff.Shape(); got != ShapeOnlyInDevTree {
			t.Fatalf("Shape() = %v, want %v", got, ShapeOnlyInDevTree)
		}
	})

	t.Run("symlink present on both sides with different targets", func(t *testing.T) {
		configRoot, devRoot := t.TempDir(), t.TempDir()
		mustSymlink(t, configRoot, "bin/tool", "../lib/a")
		mustSymlink(t, devRoot, "bin/tool", "../lib/b")

		diff, err := DiffTrees(configRoot, devRoot)
		if err != nil {
			t.Fatalf("DiffTrees: %v", err)
		}
		assertContainsExactly(t, "Changed", diff.Changed, []string{"bin/tool"})
	})

	t.Run("regular file replaced by a symlink at the same path", func(t *testing.T) {
		configRoot, devRoot := t.TempDir(), t.TempDir()
		writeFile(t, configRoot, "bin/tool", "#!/bin/sh\necho hi\n")
		mustSymlink(t, devRoot, "bin/tool", "../lib/tool-real")

		diff, err := DiffTrees(configRoot, devRoot)
		if err != nil {
			t.Fatalf("DiffTrees: %v", err)
		}
		assertContainsExactly(t, "Changed", diff.Changed, []string{"bin/tool"})
	})

	t.Run("same content, executable bit differs", func(t *testing.T) {
		configRoot, devRoot := t.TempDir(), t.TempDir()
		writeFileMode(t, configRoot, "bin/tool", "same content\n", 0o644)
		writeFileMode(t, devRoot, "bin/tool", "same content\n", 0o755)

		diff, err := DiffTrees(configRoot, devRoot)
		if err != nil {
			t.Fatalf("DiffTrees: %v", err)
		}
		assertContainsExactly(t, "Changed", diff.Changed, []string{"bin/tool"})
	})
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	writeFileMode(t, root, rel, content, 0o644)
}

func writeFileMode(t *testing.T, root, rel, content string, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func mustSymlink(t *testing.T, root, rel, target string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}

func assertContainsExactly(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}
