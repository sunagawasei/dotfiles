package sourceinventory

import (
	"fmt"
	"path/filepath"
)

func Extract(root string) (Result, error) {
	var result Result
	extractors := []struct {
		name string
		run  func(string, *Result) error
	}{
		{"Neovim colorscheme", extractNvimColorscheme},
		{"Neovim terminal colors", extractNvimTerminalColors},
		{"Neovim bufferline", extractBufferline},
		{"Neovim lualine", extractLualine},
		{"Neovim scrollbar", extractScrollbar},
		{"Vim highlights", extractVimTemplate},
		{"Claude theme", extractClaudeTheme},
		{"Hunk theme", extractHunk},
		{"gh-dash theme", extractGhDashTheme},
		{"delta styles", extractDeltaStyles},
		{"LazyGit and FZF role colors", extractGitUIRoleColors},
		{"herdr theme", extractHerdrTheme},
		{"bat theme", extractBatTheme},
		{"eza and zsh completion themes", extractEzaTheme},
		{"Markdown Preview theme", extractMarkdownPreviewTheme},
		{"WezTerm UI", extractWezTerm},
		{"statusline", extractStatusline},
	}
	for _, extractor := range extractors {
		if err := extractor.run(root, &result); err != nil {
			return Result{}, fmt.Errorf("extract %s: %w", extractor.name, err)
		}
	}
	if err := result.normalize(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func sourcePath(root, relative string) string {
	return filepath.Join(root, filepath.FromSlash(relative))
}
