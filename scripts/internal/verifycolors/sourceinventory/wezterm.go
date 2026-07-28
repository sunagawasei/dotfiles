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
	weztermFieldPattern          = regexp.MustCompile(`^\s*([a-z_]+)\s*=\s*(?:\{\s*Color\s*=\s*)?colors\.([a-z_]+\.[a-z_]+)`)
	weztermTabPattern            = regexp.MustCompile(`^\s*(active_tab|inactive_tab|inactive_tab_hover|new_tab|new_tab_hover)\s*=\s*\{`)
	weztermArrayPattern          = regexp.MustCompile(`^\s*colors\.([a-z_]+\.[a-z_]+),`)
	weztermKeyTableStartPattern  = regexp.MustCompile(`^\s*wezterm\.on\("update-right-status",\s*function\(window,\s*pane\)\s*$`)
	weztermKeyTableBranchPattern = regexp.MustCompile(`^\s*(?:if|elseif)\s+name\s*==\s*"([a-z_]+)"\s+then\s*$`)
	weztermKeyTableColorPattern  = regexp.MustCompile(`^\s*table\.insert\(elements,\s*\{\s*(Background|Foreground)\s*=\s*\{\s*Color\s*=\s*colors\.([a-z_]+\.[a-z_]+)\s*\}\s*\}\)\s*$`)
)

type weztermField struct {
	token  verifycolors.TokenRef
	source string
}

type weztermKeyTableStyle struct {
	background weztermField
	foreground weztermField
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
	if err := extractWezTermKeyTables(root, result); err != nil {
		return err
	}
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "wezterm.runtime-dynamic-colors",
		Reason: "window_background_gradient, compose_cursor, and runtime format-tab-title pairs are not statically extracted; " +
			"format-tab-title requires parsing dynamic Lua branches, including Claude state icon teals.bright/foregrounds.bright " +
			"on explicit inactive or hover tab backgrounds, so runtime pairs matching config.colors today could diverge without detection",
		Source: "wezterm/wezterm.lua:165,220,365-420",
	})
	return nil
}

func extractWezTermKeyTables(root string, result *Result) error {
	const relative = "wezterm/keybinds.lua"
	data, err := os.ReadFile(sourcePath(root, relative))
	if err != nil {
		return err
	}
	styles, err := parseWezTermKeyTables(string(data), relative)
	if err != nil {
		return err
	}
	for _, name := range []string{"copy_mode", "resize_pane", "pane_navigation", "search_mode", "other"} {
		style := styles[name]
		consumerID := "wezterm.key_table." + name
		result.addPair(defaultTextPair(
			consumerID,
			style.foreground.token,
			verifycolors.TokenBackground(style.background.token),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			style.foreground.source,
		))
		result.addPair(surfacePair(consumerID+".surface", style.background.token, style.background.source))
	}
	return nil
}

func parseWezTermKeyTables(contents, relative string) (map[string]weztermKeyTableStyle, error) {
	expectedBranches := map[string]bool{
		"copy_mode":       true,
		"resize_pane":     true,
		"pane_navigation": true,
		"search_mode":     true,
		"other":           true,
	}
	styles := make(map[string]weztermKeyTableStyle, len(expectedBranches))
	inHandler := false
	handlerComplete := false
	currentBranch := ""
	handlerCount := 0

	scanner := bufio.NewScanner(strings.NewReader(contents))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		if weztermKeyTableStartPattern.MatchString(line) {
			handlerCount++
			if handlerCount > 1 {
				return nil, fmt.Errorf("%s: duplicate update-right-status handler", source)
			}
			inHandler = true
			continue
		}
		if !inHandler {
			continue
		}
		if strings.Contains(line, "window:set_right_status(wezterm.format(elements))") {
			inHandler = false
			handlerComplete = true
			currentBranch = ""
			continue
		}
		if match := weztermKeyTableBranchPattern.FindStringSubmatch(line); match != nil {
			currentBranch = match[1]
			if !expectedBranches[currentBranch] {
				return nil, fmt.Errorf("%s: unexpected key-table branch %q", source, currentBranch)
			}
			if _, exists := styles[currentBranch]; exists {
				return nil, fmt.Errorf("%s: duplicate key-table branch %q", source, currentBranch)
			}
			styles[currentBranch] = weztermKeyTableStyle{}
			continue
		}
		if strings.TrimSpace(line) == "else" {
			currentBranch = "other"
			if _, exists := styles[currentBranch]; exists {
				return nil, fmt.Errorf("%s: duplicate key-table branch %q", source, currentBranch)
			}
			styles[currentBranch] = weztermKeyTableStyle{}
			continue
		}
		match := weztermKeyTableColorPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if currentBranch == "" {
			return nil, fmt.Errorf("%s: key-table %s color is outside a named branch", source, match[1])
		}
		style := styles[currentBranch]
		field := weztermField{token: verifycolors.TokenRef(match[2]), source: source}
		switch match[1] {
		case "Background":
			if style.background.token != "" {
				return nil, fmt.Errorf("%s: duplicate Background for key-table branch %q", source, currentBranch)
			}
			style.background = field
		case "Foreground":
			if style.foreground.token != "" {
				return nil, fmt.Errorf("%s: duplicate Foreground for key-table branch %q", source, currentBranch)
			}
			style.foreground = field
		}
		styles[currentBranch] = style
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if handlerCount == 0 {
		return nil, fmt.Errorf("%s: update-right-status handler was not found", relative)
	}
	if !handlerComplete {
		return nil, fmt.Errorf("%s: update-right-status handler did not format key-table elements", relative)
	}
	for branch := range expectedBranches {
		style, ok := styles[branch]
		if !ok {
			return nil, fmt.Errorf("%s: required key-table branch %q was not found", relative, branch)
		}
		if style.background.token == "" {
			return nil, fmt.Errorf("%s: key-table branch %q is missing Background", relative, branch)
		}
		if style.foreground.token == "" {
			return nil, fmt.Errorf("%s: key-table branch %q is missing Foreground", relative, branch)
		}
	}
	return styles, nil
}

func requireWezTermField(fields map[string]weztermField, name string) (weztermField, error) {
	field, ok := fields[name]
	if !ok {
		return weztermField{}, fmt.Errorf("required WezTerm color field %q was not found", name)
	}
	return field, nil
}
