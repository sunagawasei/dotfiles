// share-project-skills exposes Claude's project skills through Codex's discovery path.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "share-project-skills:", err)
		os.Exit(1)
	}
}

func run() error {
	recursive := flag.Bool("recursive", false, "指定ディレクトリ配下のプロジェクトも設定する")
	check := flag.Bool("check", false, "リンクを変更せず共有状態を検査する")
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	var failures []error
	for _, root := range roots {
		root, err := filepath.Abs(root)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		info, err := os.Stat(root)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if !info.IsDir() {
			failures = append(failures, fmt.Errorf("%s はディレクトリではありません", root))
			continue
		}
		if !*recursive {
			failures = append(failures, share(root, *check))
			continue
		}
		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Name() == ".claude" {
				failures = append(failures, share(filepath.Dir(path), *check))
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				switch entry.Name() {
				case ".git", ".agents", "node_modules", "vendor", ".terraform", ".direnv", ".venv":
					return filepath.SkipDir
				}
			}
			return nil
		})
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

func share(project string, check bool) error {
	source := filepath.Join(project, ".claude", "skills")
	info, err := os.Stat(source)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s はディレクトリではありません", source)
	}
	agents := filepath.Join(project, ".agents")
	if info, err := os.Lstat(agents); err == nil && !info.IsDir() {
		return fmt.Errorf("%s は実ディレクトリではありません(変更なし)", agents)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	dest := filepath.Join(agents, "skills")
	info, err = os.Lstat(dest)
	if errors.Is(err, fs.ErrNotExist) {
		if check {
			return fmt.Errorf("%s の共有リンクがありません", dest)
		}
		if err := os.MkdirAll(agents, 0o755); err != nil {
			return err
		}
		return link(source, dest, false)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return link(source, dest, check)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	var failures []error
	for _, entry := range entries {
		path := filepath.Join(source, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if info.IsDir() {
			failures = append(failures, link(path, filepath.Join(dest, entry.Name()), check))
		}
	}
	return errors.Join(failures...)
}

func link(source, dest string, check bool) error {
	if _, err := os.Lstat(dest); err == nil {
		a, errA := filepath.EvalSymlinks(source)
		b, errB := filepath.EvalSymlinks(dest)
		if errA == nil && errB == nil && a == b {
			fmt.Println("共有済み:", dest)
			return nil
		}
		return fmt.Errorf("%s は別の設定と衝突しています(変更なし)", dest)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if check {
		return fmt.Errorf("%s の共有リンクがありません", dest)
	}
	target, err := filepath.Rel(filepath.Dir(dest), source)
	if err != nil {
		return err
	}
	if err := os.Symlink(target, dest); err != nil {
		return err
	}
	fmt.Println("共有リンク作成:", dest)
	return nil
}
