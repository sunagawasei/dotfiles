package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var vimHighlightPattern = regexp.MustCompile(`^highlight\s+(\S+)\s+(.*)$`)
var placeholderPattern = regexp.MustCompile(`\{\{(?:(?:lower|xterm|rgb):)?([a-z_]+\.[a-z_]+)\}\}`)

func extractVimTemplate(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inTemplate := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.HasPrefix(line, "const vimTemplate = `") {
			inTemplate = true
			continue
		}
		if inTemplate && line == "`" {
			break
		}
		if !inTemplate {
			continue
		}
		match := vimHighlightPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		group, attrs := match[1], match[2]
		foreground, hasForeground := placeholderAfter(attrs, "guifg=")
		backgroundToken, hasBackground := placeholderAfter(attrs, "guibg=")
		background := verifycolors.AmbientBackground()
		profiles := []verifycolors.RenderProfile{
			verifycolors.ProfileTruecolor,
			verifycolors.ProfileCtermFGOnly,
		}
		if hasBackground {
			background = verifycolors.TokenBackground(backgroundToken)
			profiles[1] = verifycolors.ProfileCtermFGBG
		}
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		if hasForeground {
			result.addPair(defaultTextPair(
				"vim.highlight."+group,
				foreground,
				background,
				profiles,
				source,
			))
		}
		if hasBackground {
			result.addPair(surfacePair("vim.highlight."+group+".surface", backgroundToken, source))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !inTemplate {
		return fmt.Errorf("vimTemplate was not found")
	}
	return nil
}

func placeholderAfter(attrs, prefix string) (verifycolors.TokenRef, bool) {
	index := strings.Index(attrs, prefix)
	if index < 0 {
		return "", false
	}
	rest := attrs[index+len(prefix):]
	if strings.HasPrefix(rest, "NONE") {
		return "", false
	}
	match := placeholderPattern.FindStringSubmatch(rest)
	if match == nil {
		return "", false
	}
	return verifycolors.TokenRef(match[1]), true
}
