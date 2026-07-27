package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var scrollbarColorPattern = regexp.MustCompile(`^\s*([A-Za-z][A-Za-z0-9_]*)\s*=\s*p\.([A-Za-z0-9_]+)`)

func extractScrollbar(root string, result *Result) error {
	const relative = "nvim/lua/plugins/scrollbar.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inColors := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.Contains(line, "local colors = {") {
			inColors = true
			continue
		}
		if inColors && strings.TrimSpace(line) == "}" {
			inColors = false
		}
		if !inColors {
			continue
		}
		match := scrollbarColorPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		token, err := resolveNvimAlias(match[2])
		if err != nil {
			return fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
		}
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		if match[1] == "handle" {
			result.addPair(surfacePair("nvim.scrollbar.handle", token, source))
			continue
		}
		result.addPair(verifycolors.PairSpec{
			ConsumerID: "nvim.scrollbar." + match[1],
			Foreground: token,
			Background: verifycolors.AmbientBackground(),
			Class:      verifycolors.ClassWaived,
			Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			Role:       verifycolors.RoleIndicator,
			Reason:     "scrollbar mark position and glyph provide a non-color cue",
			Source:     source,
		})
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
