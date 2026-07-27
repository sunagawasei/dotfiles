package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	weztermFieldPattern = regexp.MustCompile(`^\s*([a-z_]+)\s*=\s*(?:\{\s*Color\s*=\s*)?colors\.([a-z_]+\.[a-z_]+)`)
	weztermTabPattern   = regexp.MustCompile(`^\s*(active_tab|inactive_tab|inactive_tab_hover|new_tab|new_tab_hover)\s*=\s*\{`)
	weztermArrayPattern = regexp.MustCompile(`^\s*colors\.([a-z_]+\.[a-z_]+),`)
)

type weztermField struct {
	token  verifycolors.TokenRef
	source string
}

func extractWezTerm(root string, result *Result) error {
	const relative = "wezterm/wezterm.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fields := make(map[string]weztermField)
	inColors := false
	arrayName := ""
	tabState := ""
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		if strings.HasPrefix(line, "config.colors = {") {
			inColors = true
			continue
		}
		if match := regexp.MustCompile(`^config\.(char_select_(?:bg|fg)_color)\s*=\s*colors\.([a-z_]+\.[a-z_]+)`).FindStringSubmatch(line); match != nil {
			fields[match[1]] = weztermField{token: verifycolors.TokenRef(match[2]), source: source}
			continue
		}
		if !inColors {
			continue
		}
		if match := weztermTabPattern.FindStringSubmatch(line); match != nil {
			tabState = match[1]
			continue
		}
		if trimmed == "ansi = {" {
			arrayName = "ansi"
			continue
		}
		if trimmed == "brights = {" {
			arrayName = "brights"
			continue
		}
		if arrayName != "" {
			if trimmed == "}," {
				arrayName = ""
				continue
			}
			if match := weztermArrayPattern.FindStringSubmatch(line); match != nil {
				token := verifycolors.TokenRef(match[1])
				result.addPair(defaultTextPair(
					"wezterm."+arrayName+"."+strings.TrimPrefix(match[1], "ansi."),
					token,
					verifycolors.AmbientBackground(),
					[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
					source,
				))
			}
			continue
		}
		if tabState != "" {
			if trimmed == "}," {
				tabState = ""
				continue
			}
			if match := weztermFieldPattern.FindStringSubmatch(line); match != nil {
				fields["tab_bar."+tabState+"."+match[1]] = weztermField{
					token:  verifycolors.TokenRef(match[2]),
					source: source,
				}
			}
			continue
		}
		if match := weztermFieldPattern.FindStringSubmatch(line); match != nil {
			fields[match[1]] = weztermField{token: verifycolors.TokenRef(match[2]), source: source}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	pairs := []struct {
		id string
		fg string
		bg string
	}{
		{"wezterm.colors.foreground", "foreground", "background"},
		{"wezterm.colors.cursor", "cursor_fg", "cursor_bg"},
		{"wezterm.colors.selection", "selection_fg", "selection_bg"},
		{"wezterm.colors.copy_mode_active", "copy_mode_active_highlight_fg", "copy_mode_active_highlight_bg"},
		{"wezterm.colors.copy_mode_inactive", "copy_mode_inactive_highlight_fg", "copy_mode_inactive_highlight_bg"},
		{"wezterm.colors.quick_select_label", "quick_select_label_fg", "quick_select_label_bg"},
		{"wezterm.colors.quick_select_match", "quick_select_match_fg", "quick_select_match_bg"},
		{"wezterm.colors.tab_bar.active_tab", "tab_bar.active_tab.fg_color", "tab_bar.active_tab.bg_color"},
		{"wezterm.colors.tab_bar.inactive_tab", "tab_bar.inactive_tab.fg_color", "tab_bar.inactive_tab.bg_color"},
		{"wezterm.colors.tab_bar.inactive_tab_hover", "tab_bar.inactive_tab_hover.fg_color", "tab_bar.inactive_tab_hover.bg_color"},
		{"wezterm.colors.tab_bar.new_tab_hover", "tab_bar.new_tab_hover.fg_color", "tab_bar.new_tab_hover.bg_color"},
		{"wezterm.colors.char_select", "char_select_fg_color", "char_select_bg_color"},
	}
	for _, pair := range pairs {
		fg, err := requireWezTermField(fields, pair.fg)
		if err != nil {
			return err
		}
		bg, err := requireWezTermField(fields, pair.bg)
		if err != nil {
			return err
		}
		result.addPair(defaultTextPair(
			pair.id,
			fg.token,
			verifycolors.TokenBackground(bg.token),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			fg.source,
		))
		result.addPair(surfacePair(pair.id+".surface", bg.token, bg.source))
	}

	if field, ok := fields["tab_bar.new_tab.fg_color"]; ok {
		result.addPair(defaultTextPair(
			"wezterm.colors.tab_bar.new_tab",
			field.token,
			verifycolors.AmbientBackground(),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			field.source,
		))
	}
	for _, fieldName := range []string{"background", "scrollbar_thumb", "visual_bell"} {
		if field, ok := fields[fieldName]; ok {
			result.addPair(surfacePair("wezterm.colors."+fieldName, field.token, field.source))
		}
	}
	for _, fieldName := range []string{"cursor_border", "split", "inactive_tab_edge"} {
		if field, ok := fields[fieldName]; ok {
			consumerField := fieldName
			if fieldName == "inactive_tab_edge" {
				consumerField = "tab_bar.inactive_tab_edge"
			}
			result.addPair(verifycolors.PairSpec{
				ConsumerID: "wezterm.colors." + consumerField,
				Foreground: field.token,
				Background: verifycolors.AmbientBackground(),
				Class:      verifycolors.ClassWaived,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleBorder,
				Reason:     "border or divider shape provides a non-color cue",
				Source:     field.source,
			})
		}
	}
	return nil
}

func requireWezTermField(fields map[string]weztermField, name string) (weztermField, error) {
	field, ok := fields[name]
	if !ok {
		return weztermField{}, fmt.Errorf("required WezTerm color field %q was not found", name)
	}
	return field, nil
}
