package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	ghDashSectionPattern = regexp.MustCompile(`^        ([a-z][a-z0-9]*):$`)
	ghDashColorPattern   = regexp.MustCompile(`^            ([a-z][a-z0-9]*):\s*"\{\{([a-z_]+\.[a-z_]+)\}\}"$`)
)

func extractGhDashTheme(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inTemplate := false
	foundTemplate := false
	section := ""
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !inTemplate {
			if line == "const ghDashTemplate = `        # BEGIN GENERATED COLORS" {
				inTemplate = true
				foundTemplate = true
			}
			continue
		}
		if line == "`" {
			inTemplate = false
			break
		}
		if match := ghDashSectionPattern.FindStringSubmatch(line); match != nil {
			section = match[1]
			continue
		}
		match := ghDashColorPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if section == "" {
			return fmt.Errorf("%s:%d: gh-dash color %q has no section", relative, lineNumber, match[1])
		}

		fieldPath := section + "." + match[1]
		if seen[fieldPath] {
			return fmt.Errorf("%s:%d: duplicate gh-dash color %q", relative, lineNumber, fieldPath)
		}
		seen[fieldPath] = true

		token := verifycolors.TokenRef(match[2])
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "gh-dash.theme." + fieldPath
		switch section {
		case "background":
			result.addPair(surfacePair(consumerID, token, source))
		case "border":
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: token,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassReportOnly,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleBorder,
				Source:     source,
			})
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
		return fmt.Errorf("%s: ghDashTemplate was not found", relative)
	}
	if inTemplate {
		return fmt.Errorf("%s: ghDashTemplate is not terminated", relative)
	}
	if len(seen) == 0 {
		return fmt.Errorf("%s: ghDashTemplate contains no color placeholders", relative)
	}
	return nil
}
