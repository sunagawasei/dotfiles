package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var hunkColorPattern = regexp.MustCompile(`^\s*([A-Za-z][A-Za-z0-9_]*)\s*=\s*colors\.([a-z_]+\.[a-z_]+);`)

var hunkSurfaceFields = map[string]bool{
	"background":          true,
	"panel":               true,
	"panelAlt":            true,
	"addedBg":             true,
	"removedBg":           true,
	"movedAddedBg":        true,
	"movedRemovedBg":      true,
	"contextBg":           true,
	"addedContentBg":      true,
	"removedContentBg":    true,
	"contextContentBg":    true,
	"lineNumberBg":        true,
	"selectedHunk":        true,
	"noteBackground":      true,
	"noteTitleBackground": true,
}

func extractHunk(root string, result *Result) error {
	const relative = "home-manager/hunk.nix"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inSyntax := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.Contains(line, "syntax = {") {
			inSyntax = true
			continue
		}
		if inSyntax && strings.TrimSpace(line) == "};" {
			inSyntax = false
			continue
		}
		match := hunkColorPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		field := match[1]
		if strings.HasPrefix(match[2], "metadata.") {
			continue
		}
		fieldPath := field
		if inSyntax {
			fieldPath = "syntax." + field
		}
		token := verifycolors.TokenRef(match[2])
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "hunk.theme." + fieldPath
		if hunkSurfaceFields[field] {
			result.addPair(surfacePair(consumerID, token, source))
			continue
		}
		if strings.Contains(strings.ToLower(field), "border") {
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: token,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassWaived,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleBorder,
				Reason:     "border shape provides a non-color cue",
				Source:     source,
			})
			continue
		}
		result.addPair(defaultTextPair(
			consumerID,
			token,
			verifycolors.AmbientBackground(),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			source,
		))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
