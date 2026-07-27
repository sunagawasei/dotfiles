package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	highlightTablePattern = regexp.MustCompile(`^\s*(?:\["([^"]+)"\]|([A-Za-z][A-Za-z0-9_]*))\s*=\s*\{(.*)\}`)
	setHighlightPattern   = regexp.MustCompile(`nvim_set_hl\(0,\s*"([^"]+)",\s*\{(.*)\}\)`)
	nvimFGPattern         = regexp.MustCompile(`\bfg\s*=\s*colors\.([A-Za-z0-9_]+)`)
	nvimBGPattern         = regexp.MustCompile(`\bbg\s*=\s*colors\.([A-Za-z0-9_]+)`)
	nvimNoneBGPattern     = regexp.MustCompile(`\bbg\s*=\s*"none"`)
	nvimSPPattern         = regexp.MustCompile(`\bsp\s*=\s*colors\.([A-Za-z0-9_]+)`)
)

type highlightState struct {
	foreground    verifycolors.TokenRef
	hasForeground bool
	background    verifycolors.BackgroundRef
	hasBackground bool
	indicator     verifycolors.TokenRef
	hasIndicator  bool
	source        string
}

func extractNvimColorscheme(root string, result *Result) error {
	const relative = "nvim/lua/plugins/colorscheme.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	states := make(map[string]highlightState)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		group, attrs, ok := parseHighlightLine(line)
		if !ok {
			continue
		}
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		state := states[group]
		if err := applyHighlightAttrs(&state, attrs, source); err != nil {
			return fmt.Errorf("%s: %w", source, err)
		}
		states[group] = state
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for group, state := range states {
		background := state.background
		if !state.hasBackground {
			background = verifycolors.AmbientBackground()
		}
		if state.hasForeground {
			result.addPair(defaultTextPair(
				"nvim.highlight."+group,
				state.foreground,
				background,
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				state.source,
			))
		}
		if state.hasBackground && !state.background.Ambient {
			result.addPair(surfacePair(
				"nvim.highlight."+group+".surface",
				state.background.Token,
				state.source,
			))
		}
		if state.hasIndicator {
			result.addPair(verifycolors.PairSpec{
				ConsumerID: "nvim.highlight." + group + ".indicator",
				Foreground: state.indicator,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassWaived,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleIndicator,
				Reason:     "undercurl or underline shape provides a non-color cue",
				Source:     state.source,
			})
		}
	}
	return nil
}

func parseHighlightLine(line string) (string, string, bool) {
	if match := setHighlightPattern.FindStringSubmatch(line); match != nil {
		return match[1], match[2], true
	}
	match := highlightTablePattern.FindStringSubmatch(line)
	if match == nil {
		return "", "", false
	}
	group := match[1]
	if group == "" {
		group = match[2]
	}
	attrs := match[3]
	if !nvimFGPattern.MatchString(attrs) &&
		!nvimBGPattern.MatchString(attrs) &&
		!nvimNoneBGPattern.MatchString(attrs) &&
		!nvimSPPattern.MatchString(attrs) {
		return "", "", false
	}
	return group, attrs, true
}

func applyHighlightAttrs(state *highlightState, attrs, source string) error {
	if match := nvimFGPattern.FindStringSubmatch(attrs); match != nil {
		token, err := resolveNvimAlias(match[1])
		if err != nil {
			return err
		}
		state.foreground = token
		state.hasForeground = true
		state.source = source
	}
	if match := nvimBGPattern.FindStringSubmatch(attrs); match != nil {
		token, err := resolveNvimAlias(match[1])
		if err != nil {
			return err
		}
		state.background = verifycolors.TokenBackground(token)
		state.hasBackground = true
		state.source = source
	} else if nvimNoneBGPattern.MatchString(attrs) {
		state.background = verifycolors.AmbientBackground()
		state.hasBackground = true
		state.source = source
	}
	if match := nvimSPPattern.FindStringSubmatch(attrs); match != nil {
		token, err := resolveNvimAlias(match[1])
		if err != nil {
			return err
		}
		state.indicator = token
		state.hasIndicator = true
		state.source = source
	}
	return nil
}
