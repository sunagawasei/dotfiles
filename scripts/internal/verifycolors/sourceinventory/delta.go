package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

type deltaStyleSpec struct {
	pattern    string
	fixedValue string
	surface    bool
	enforced   bool
	role       verifycolors.PairRole
}

var deltaStyleSpecs = map[string]deltaStyleSpec{
	"plus-style":                    {pattern: "syntax <color>", surface: true},
	"minus-style":                   {pattern: "syntax <color>", surface: true},
	"plus-emph-style":               {pattern: "<color> <color>", enforced: true},
	"minus-emph-style":              {pattern: "<color> <color>", enforced: true},
	"syntax-theme":                  {fixedValue: "ghost-visor"},
	"line-numbers-plus-style":       {pattern: "<color>"},
	"line-numbers-minus-style":      {pattern: "<color>"},
	"line-numbers-zero-style":       {pattern: "<color>"},
	"line-numbers-left-style":       {pattern: "<color>"},
	"line-numbers-right-style":      {pattern: "<color>"},
	"hunk-header-line-number-style": {pattern: "<color>"},
	"hunk-header-decoration-style":  {pattern: "<color> box", role: verifycolors.RoleBorder},
	"file-style":                    {pattern: "<color>"},
	"whitespace-error-style":        {pattern: "<color>"},
}

var (
	deltaOptionPattern = regexp.MustCompile(`^\s*([a-z][a-z0-9-]+)\s*=\s*"([^"]*)";\s*$`)
	deltaTokenPattern  = regexp.MustCompile(`\$\{colors\.([a-z_]+\.[a-z_]+)\}`)
)

func extractDeltaStyles(root string, result *Result) error {
	const relative = "home-manager/git.nix"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inDelta := false
	inOptions := false
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if !inDelta {
			if trimmed == "programs.delta = {" {
				inDelta = true
			}
			continue
		}
		if !inOptions {
			if trimmed == "options = {" {
				inOptions = true
			} else if trimmed == "};" {
				break
			}
			continue
		}
		if trimmed == "};" {
			inOptions = false
			continue
		}

		match := deltaOptionPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		option := match[1]
		if option == "hunk-header-file-style" {
			return fmt.Errorf("%s:%d: unsupported delta option %q must remain unset", relative, lineNumber, option)
		}
		spec, ok := deltaStyleSpecs[option]
		if !ok {
			continue
		}
		if seen[option] {
			return fmt.Errorf("%s:%d: duplicate delta option %q", relative, lineNumber, option)
		}

		if spec.fixedValue != "" {
			if match[2] != spec.fixedValue {
				return fmt.Errorf(
					"%s:%d: delta option %q = %q, want fixed value %q",
					relative,
					lineNumber,
					option,
					match[2],
					spec.fixedValue,
				)
			}
			seen[option] = true
			continue
		}

		tokenMatches := deltaTokenPattern.FindAllStringSubmatch(match[2], -1)
		wantTokenCount := strings.Count(spec.pattern, "<color>")
		if len(tokenMatches) != wantTokenCount {
			return fmt.Errorf(
				"%s:%d: delta option %q contains %d colors tokens, want %d",
				relative,
				lineNumber,
				option,
				len(tokenMatches),
				wantTokenCount,
			)
		}
		normalized := deltaTokenPattern.ReplaceAllString(match[2], "<color>")
		if normalized != spec.pattern {
			return fmt.Errorf(
				"%s:%d: delta option %q has structure %q, want %q",
				relative,
				lineNumber,
				option,
				normalized,
				spec.pattern,
			)
		}
		seen[option] = true

		tokens := make([]verifycolors.TokenRef, len(tokenMatches))
		for index, tokenMatch := range tokenMatches {
			tokens[index] = verifycolors.TokenRef(tokenMatch[1])
		}
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "delta.style." + option
		switch {
		// emph has explicit fg/bg and is enforced; non-emph keeps dynamic syntax fg and a report-only surface.
		case spec.enforced:
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: tokens[0],
				Background: verifycolors.TokenBackground(tokens[1]),
				Class:      verifycolors.ClassEnforced,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleText,
				Source:     source,
			})
		case spec.surface:
			result.addPair(surfacePair(consumerID, tokens[0], source))
		case spec.role == verifycolors.RoleBorder:
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: tokens[0],
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassReportOnly,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleBorder,
				Source:     source,
			})
		default:
			result.addPair(defaultTextPair(
				consumerID,
				tokens[0],
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				source,
			))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !inDelta {
		return fmt.Errorf("%s: programs.delta was not found", relative)
	}
	if inOptions {
		return fmt.Errorf("%s: programs.delta.options is not terminated", relative)
	}
	for option := range deltaStyleSpecs {
		if !seen[option] {
			return fmt.Errorf("%s: delta option %q was not found", relative, option)
		}
	}
	return validateLazygitDeltaPager(root)
}

func validateLazygitDeltaPager(root string) error {
	const relative = "lazygit/config.yml"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	found := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "pager:") || !strings.Contains(line, "delta ") {
			continue
		}
		found = true
		if strings.Contains(line, "--syntax-theme") {
			return fmt.Errorf("%s:%d: delta pager must not override syntax-theme", relative, lineNumber)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%s: delta pager was not found", relative)
	}
	return nil
}
