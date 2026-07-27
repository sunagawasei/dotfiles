package verifycolors

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	hexPattern = regexp.MustCompile(`#[0-9A-Fa-f]{6}(?:[0-9A-Fa-f]{2})?`)

	defaultScanFiles = []string{
		"wezterm/wezterm.lua",
		"wezterm/keybinds.lua",
		"nvim/lua/plugins/colorscheme.lua",
		"nvim/lua/plugins/render-markdown.lua",
		"nvim/lua/plugins/lualine.lua",
		"nvim/lua/plugins/scrollbar.lua",
		"zsh/.zshrc",
		"lazygit/config.yml",
	}

	allowedLiteralColors = map[string]bool{
		"#000000": true,
		"#FFFFFF": true,
		"#F8FCFD": true,
	}
)

type HexScanIssue struct {
	Path   string
	Colors []string
}

type HexScanResult struct {
	Checked int
	Issues  []HexScanIssue
}

func DefaultScanFiles() []string {
	files := make([]string, len(defaultScanFiles))
	copy(files, defaultScanFiles)
	return files
}

func ScanHexLiterals(root string, files []string, palette *Palette) (HexScanResult, error) {
	if len(files) == 0 {
		files = DefaultScanFiles()
	}
	validColors := make(map[string]bool)
	for _, value := range palette.TokenValues() {
		validColors[normalizeLiteral(value)] = true
	}
	for color := range allowedLiteralColors {
		validColors[color] = true
	}

	var result HexScanResult
	for _, relative := range files {
		if filepath.IsAbs(relative) {
			return result, fmt.Errorf("scan file must be relative to scan root: %q", relative)
		}
		path := filepath.Join(root, relative)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, fmt.Errorf("scan %s: %w", relative, err)
		}
		result.Checked++
		seen := make(map[string]bool)
		var unknown []string
		for _, match := range hexPattern.FindAllString(string(data), -1) {
			normalized := normalizeLiteral(match)
			if seen[normalized] {
				continue
			}
			seen[normalized] = true
			if !validColors[normalized] {
				unknown = append(unknown, normalized)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			result.Issues = append(result.Issues, HexScanIssue{
				Path:   filepath.ToSlash(relative),
				Colors: unknown,
			})
		}
	}
	sort.Slice(result.Issues, func(i, j int) bool {
		return result.Issues[i].Path < result.Issues[j].Path
	})
	return result, nil
}

func normalizeLiteral(value string) string {
	value = strings.ToUpper(strings.Trim(strings.TrimSpace(value), `"'`))
	if len(value) == 9 {
		value = value[:7]
	}
	return value
}
