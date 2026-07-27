package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

type herdrThemeSpec struct {
	token   verifycolors.TokenRef
	fixed   string
	visible bool
	role    verifycolors.PairRole
}

var herdrThemeSpecs = map[string]herdrThemeSpec{
	"accent":      {token: "ansi.blue", visible: true, role: verifycolors.RoleBorder},
	"panel_bg":    {fixed: "reset"},
	"surface0":    {fixed: "reset"},
	"surface1":    {token: "core.active_line", visible: true, role: verifycolors.RoleSurface},
	"surface_dim": {token: "ansi.bright_black", visible: true, role: verifycolors.RoleSurface},
	"overlay0":    {token: "foregrounds.dim", visible: true, role: verifycolors.RoleText},
	"overlay1":    {token: "ansi.bright_white", visible: true, role: verifycolors.RoleText},
	"text":        {token: "foregrounds.main", visible: true, role: verifycolors.RoleText},
	"subtext0":    {token: "foregrounds.dim", visible: true, role: verifycolors.RoleText},
	"mauve":       {token: "foregrounds.heading", visible: true, role: verifycolors.RoleText},
	"green":       {token: "ansi.green", visible: true, role: verifycolors.RoleIndicator},
	"yellow":      {token: "ansi.yellow", visible: true, role: verifycolors.RoleIndicator},
	"red":         {token: "ansi.bright_red", visible: true, role: verifycolors.RoleIndicator},
	"blue":        {token: "ansi.blue", visible: true, role: verifycolors.RoleIndicator},
	"teal":        {token: "ansi.cyan", visible: true, role: verifycolors.RoleIndicator},
	"peach":       {token: "semantic.warning"},
}

var (
	herdrThemeEntryPattern = regexp.MustCompile(`^([A-Za-z0-9_-]+)\s*=\s*"([^"]+)"(?:\s+#.*)?$`)
	herdrThemeTokenPattern = regexp.MustCompile(`^\{\{([a-z_]+\.[a-z_]+)\}\}$`)
)

func extractHerdrTheme(root string, result *Result) error {
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
			if strings.HasPrefix(line, "const herdrTemplate = `") {
				inTemplate = true
				foundTemplate = true
			}
			continue
		}
		if line == "`" {
			inTemplate = false
			break
		}

		match := herdrThemeEntryPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		key, value := match[1], match[2]
		spec, ok := herdrThemeSpecs[key]
		if !ok {
			return fmt.Errorf("%s:%d: unknown herdr theme key %q", relative, lineNumber, key)
		}
		if seen[key] {
			return fmt.Errorf("%s:%d: duplicate herdr theme key %q", relative, lineNumber, key)
		}
		seen[key] = true
		source := fmt.Sprintf("%s:%d", relative, lineNumber)

		if spec.fixed != "" {
			if value != spec.fixed {
				return fmt.Errorf(
					"%s:%d: herdr theme key %q = %q, want fixed value %q",
					relative,
					lineNumber,
					key,
					value,
					spec.fixed,
				)
			}
			addHerdrResetCoverageNote(result, key, source)
			continue
		}

		tokenMatch := herdrThemeTokenPattern.FindStringSubmatch(value)
		if tokenMatch == nil {
			return fmt.Errorf("%s:%d: herdr theme key %q must use a color token placeholder", relative, lineNumber, key)
		}
		token := verifycolors.TokenRef(tokenMatch[1])
		if token != spec.token {
			return fmt.Errorf(
				"%s:%d: herdr theme key %q resolves to %q, want %q",
				relative,
				lineNumber,
				key,
				token,
				spec.token,
			)
		}
		if !spec.visible {
			result.addCoverageNote(verifycolors.CoverageNote{
				ID:     "herdr.theme.peach.unused",
				Reason: "herdr 0.7.4 has no consumer for peach because AgentState has no Interrupted variant and no rendering branch uses it; the key is still generated to keep all 16 custom theme keys managed, but it has no visible color pair",
				Source: source,
			})
			continue
		}

		consumerID := "herdr.theme." + key
		switch spec.role {
		case verifycolors.RoleSurface:
			result.addPair(surfacePair(consumerID, token, source))
		case verifycolors.RoleBorder, verifycolors.RoleIndicator:
			result.addPair(verifycolors.PairSpec{
				ConsumerID: consumerID,
				Foreground: token,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassReportOnly,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       spec.role,
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
		return fmt.Errorf("%s: herdrTemplate was not found", relative)
	}
	if inTemplate {
		return fmt.Errorf("%s: herdrTemplate is not terminated", relative)
	}
	for key := range herdrThemeSpecs {
		if !seen[key] {
			return fmt.Errorf("%s: herdr theme key %q was not found", relative, key)
		}
	}
	return nil
}

func addHerdrResetCoverageNote(result *Result, key, source string) {
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "herdr.theme." + key + ".reset",
		Reason: fmt.Sprintf(
			"%s is strictly checked as the fixed string \"reset\"; it restores the terminal default background after Clear, so the effective color depends on WezTerm #202A42 at 0.90 opacity composited with the wallpaper and cannot be represented as a static color pair",
			key,
		),
		Source: source,
	})
}
