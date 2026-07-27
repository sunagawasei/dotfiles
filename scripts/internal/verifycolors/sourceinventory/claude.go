package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var claudeOverridePattern = regexp.MustCompile(`^\s*"([^"]+)":\s*"\{\{(?:(?:lower|xterm|rgb):)?([a-z_]+\.[a-z_]+)\}\}"`)

var claudeSurfaceKeys = map[string]bool{
	"diffAdded":                  true,
	"diffRemoved":                true,
	"diffAddedDimmed":            true,
	"diffRemovedDimmed":          true,
	"diffAddedWord":              true,
	"diffRemovedWord":            true,
	"userMessageBackground":      true,
	"userMessageBackgroundHover": true,
	"bashMessageBackgroundColor": true,
	"memoryBackgroundColor":      true,
	"selectionBg":                true,
	"rate_limit_fill":            true,
	"rate_limit_empty":           true,
}

func extractClaudeTheme(root string, result *Result) error {
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
		if strings.HasPrefix(line, "const claudeThemeTemplate = `") {
			inTemplate = true
			continue
		}
		if inTemplate && line == "`" {
			break
		}
		if !inTemplate {
			continue
		}
		match := claudeOverridePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		key := match[1]
		token := verifycolors.TokenRef(match[2])
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "claude.theme." + key
		if claudeSurfaceKeys[key] {
			result.addPair(surfacePair(consumerID, token, source))
			continue
		}
		if strings.Contains(strings.ToLower(key), "border") {
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
	if !inTemplate {
		return fmt.Errorf("claudeThemeTemplate was not found")
	}
	return nil
}
