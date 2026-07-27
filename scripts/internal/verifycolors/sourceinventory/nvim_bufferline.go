package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var bufferlineEntryPattern = regexp.MustCompile(`^\s*([a-z][a-z0-9_]*)\s*=\s*\{(.*)\}`)
var nvimPFGPattern = regexp.MustCompile(`\bfg\s*=\s*p\.([A-Za-z0-9_]+)`)
var nvimPBGPattern = regexp.MustCompile(`\bbg\s*=\s*p\.([A-Za-z0-9_]+)`)

func extractBufferline(root string, result *Result) error {
	const relative = "nvim/lua/plugins/ui.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inHighlights := false
	seenRuntimeKeys := false
	seenRuntimeBackgroundRule := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.Contains(line, "opts.highlights = {") {
			inHighlights = true
			continue
		}
		if inHighlights && strings.TrimSpace(line) == "}" {
			inHighlights = false
		}
		if strings.Contains(line, "vim.tbl_keys(cfgmod.highlights)") {
			seenRuntimeKeys = true
		}
		if strings.Contains(line, `highlight.bg = "none"`) {
			seenRuntimeBackgroundRule = true
		}
		if !inHighlights {
			continue
		}
		match := bufferlineEntryPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		key, attrs := match[1], match[2]
		fgMatch := nvimPFGPattern.FindStringSubmatch(attrs)
		if fgMatch == nil {
			continue
		}
		foreground, err := resolveNvimAlias(fgMatch[1])
		if err != nil {
			return fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
		}
		background := verifycolors.AmbientBackground()
		if bgMatch := nvimPBGPattern.FindStringSubmatch(attrs); bgMatch != nil {
			token, err := resolveNvimAlias(bgMatch[1])
			if err != nil {
				return fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
			}
			background = verifycolors.TokenBackground(token)
		}
		result.addPair(defaultTextPair(
			"nvim.bufferline."+key,
			foreground,
			background,
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			fmt.Sprintf("%s:%d", relative, lineNumber),
		))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !seenRuntimeKeys || !seenRuntimeBackgroundRule {
		return fmt.Errorf("runtime BufferLine highlight background rule was not found")
	}
	result.addCoverageNote(verifycolors.CoverageNote{
		ID:     "nvim.bufferline.runtime-default-groups",
		Reason: "BufferLine groups not present in opts.highlights are discovered at runtime; their background is ambient, but plugin-default foregrounds are not statically attributable to palette tokens",
		Source: relative,
	})
	return nil
}
