package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

const nvimTerminalColorCount = 16

var nvimTerminalColorPattern = regexp.MustCompile(
	`^\s*vim\.g\.terminal_color_([0-9]+)\s*=\s*colors\.([A-Za-z0-9_]+)\b`,
)

func extractNvimTerminalColors(root string, result *Result) error {
	const relative = "nvim/lua/plugins/colorscheme.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	seen := make([]bool, nvimTerminalColorCount)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		match := nvimTerminalColorPattern.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}

		index, err := strconv.Atoi(match[1])
		if err != nil {
			return fmt.Errorf("%s:%d: parse terminal color index: %w", relative, lineNumber, err)
		}
		if index < 0 || index >= nvimTerminalColorCount {
			return fmt.Errorf("%s:%d: terminal color index %d is out of range", relative, lineNumber, index)
		}
		if seen[index] {
			return fmt.Errorf("%s:%d: duplicate terminal color index %d", relative, lineNumber, index)
		}
		seen[index] = true

		token, err := resolveNvimAlias(match[2])
		if err != nil {
			return fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
		}
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		result.addPair(defaultTextPair(
			fmt.Sprintf("nvim.terminal.color%d", index),
			token,
			verifycolors.AmbientBackground(),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			source,
		))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for index, found := range seen {
		if !found {
			return fmt.Errorf("%s: terminal color index %d was not found", relative, index)
		}
	}
	return nil
}
