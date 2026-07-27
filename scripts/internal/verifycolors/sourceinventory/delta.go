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
	prefix  string
	suffix  string
	surface bool
	role    verifycolors.PairRole
}

var deltaStyleSpecs = map[string]deltaStyleSpec{
	"plus-style":                    {prefix: "syntax ", surface: true},
	"minus-style":                   {prefix: "syntax ", surface: true},
	"plus-emph-style":               {prefix: "syntax ", surface: true},
	"minus-emph-style":              {prefix: "syntax ", surface: true},
	"line-numbers-plus-style":       {},
	"line-numbers-minus-style":      {},
	"line-numbers-zero-style":       {},
	"line-numbers-left-style":       {},
	"line-numbers-right-style":      {},
	"hunk-header-line-number-style": {},
	"hunk-header-decoration-style":  {suffix: " box", role: verifycolors.RoleBorder},
	"file-style":                    {},
	"whitespace-error-style":        {},
}

var (
	deltaOptionPattern = regexp.MustCompile(`^\s*([a-z][a-z0-9-]+)\s*=\s*"([^"]*)";\s*$`)
	deltaTokenPattern  = regexp.MustCompile(`^(.*?)\$\{colors\.([a-z_]+\.[a-z_]+)\}(.*?)$`)
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

		tokenMatch := deltaTokenPattern.FindStringSubmatch(match[2])
		if tokenMatch == nil {
			return fmt.Errorf("%s:%d: delta option %q must contain one colors token", relative, lineNumber, option)
		}
		if tokenMatch[1] != spec.prefix || tokenMatch[3] != spec.suffix {
			return fmt.Errorf(
				"%s:%d: delta option %q has special values %q and %q, want %q and %q",
				relative,
				lineNumber,
				option,
				tokenMatch[1],
				tokenMatch[3],
				spec.prefix,
				spec.suffix,
			)
		}
		seen[option] = true

		token := verifycolors.TokenRef(tokenMatch[2])
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		consumerID := "delta.style." + option
		switch {
		case spec.surface:
			result.addPair(surfacePair(consumerID, token, source))
		case spec.role == verifycolors.RoleBorder:
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
	return nil
}
