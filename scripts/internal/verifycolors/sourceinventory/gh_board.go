package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var ghBoardColorPattern = regexp.MustCompile(`^([a-z][a-z0-9_]*)\s*=\s*"\{\{([a-z_]+\.[a-z_]+)\}\}"$`)

func extractGhBoardTheme(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inTemplate := false
	foundTemplate := false
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !inTemplate {
			if line == "const ghBoardTemplate = `# BEGIN GENERATED COLORS" {
				inTemplate = true
				foundTemplate = true
			}
			continue
		}
		if line == "`" {
			inTemplate = false
			break
		}
		match := ghBoardColorPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		key := match[1]
		if seen[key] {
			return fmt.Errorf("%s:%d: duplicate gh-board color %q", relative, lineNumber, key)
		}
		seen[key] = true

		token := verifycolors.TokenRef(match[2])
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "gh-board.theme." + key
		switch key {
		case "border_focused", "border_unfocused":
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: token,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassReportOnly,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleBorder,
				Source:     source,
			})
		case "shadow_bg", "shadow_fg", "text_inverted":
			result.addPair(surfacePair(consumerID, token, source))
		default:
			result.addPair(defaultTextPair(
				consumerID,
				token,
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				source,
			))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !foundTemplate {
		return fmt.Errorf("%s: ghBoardTemplate was not found", relative)
	}
	if inTemplate {
		return fmt.Errorf("%s: ghBoardTemplate is not terminated", relative)
	}
	if len(seen) == 0 {
		return fmt.Errorf("%s: ghBoardTemplate contains no color placeholders", relative)
	}
	return nil
}
