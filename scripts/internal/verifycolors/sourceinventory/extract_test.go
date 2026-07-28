package sourceinventory

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
	"github.com/sunagawasei/dotfiles/scripts/internal/xterm"
)

func TestRepresentativeBackgroundResolution(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	assertPair(t, pairs, "nvim.highlight.BlinkCmpKind", "foregrounds.subdued", "core.ui_shadow", false, verifycolors.ClassEnforced)
	assertPair(t, pairs, "nvim.highlight.LspInlayHint", "foregrounds.subdued", "", true, verifycolors.ClassReportOnly)

	modeBackgrounds := map[string]verifycolors.TokenRef{
		"normal":   "foregrounds.subdued",
		"insert":   "semantic.success",
		"visual":   "teals.mid_bright",
		"replace":  "purples.lavender",
		"command":  "purples.muted_purple",
		"terminal": "semantic.string",
	}
	for mode, background := range modeBackgrounds {
		assertPair(
			t,
			pairs,
			"nvim.lualine."+mode+".a",
			"core.darkest_bg",
			background,
			false,
			verifycolors.ClassEnforced,
		)
	}
	for _, part := range []string{"a", "b", "c"} {
		assertPair(
			t,
			pairs,
			"nvim.lualine.inactive."+part,
			"foregrounds.dim",
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}

	for _, consumerID := range []string{
		"nvim.highlight.Normal",
		"nvim.highlight.NormalNC",
		"nvim.highlight.TabLine",
	} {
		pair, ok := pairs[consumerID]
		if !ok {
			t.Fatalf("pair %s not found", consumerID)
		}
		if !pair.Background.Ambient {
			t.Errorf("%s background = %q, want ambient after transparency reapplication", consumerID, pair.Background.Token)
		}
	}
}

func TestSyntaxAndUIRoleSeparation(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	for _, consumerID := range []string{
		"nvim.highlight.Statement",
		"nvim.highlight.Keyword",
		"nvim.highlight.@keyword",
		"nvim.highlight.@keyword.function",
		"nvim.highlight.@keyword.operator",
		"nvim.highlight.@keyword.return",
		"nvim.highlight.@boolean.yaml",
		"nvim.highlight.@constant.builtin.yaml",
		"hunk.theme.syntax.keyword",
		"vim.highlight.Keyword",
		"vim.highlight.Statement",
		"vim.highlight.Conditional",
		"vim.highlight.Repeat",
	} {
		assertPair(
			t,
			pairs,
			consumerID,
			"semantic.keyword",
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}

	assertPair(
		t,
		pairs,
		"nvim.lualine.command.a",
		"core.darkest_bg",
		"purples.muted_purple",
		false,
		verifycolors.ClassEnforced,
	)
	assertPair(
		t,
		pairs,
		"hunk.theme.fileRenamed",
		"purples.muted_purple",
		"",
		true,
		verifycolors.ClassReportOnly,
	)
}

func TestNvimSyntaxRolesUseCanonicalSemanticTokens(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expected := map[verifycolors.TokenRef][]string{
		"semantic.function": {
			"nvim.highlight.Function",
			"nvim.highlight.@function",
			"nvim.highlight.@function.call",
			"nvim.highlight.@function.method",
			"nvim.highlight.@function.method.call",
			"nvim.highlight.@function.builtin",
			"nvim.highlight.@function.macro",
		},
		"semantic.type": {
			"nvim.highlight.Type",
			"nvim.highlight.@type",
			"nvim.highlight.@type.builtin",
		},
		"semantic.number": {
			"nvim.highlight.Number",
			"nvim.highlight.@number",
		},
		"semantic.constant": {
			"nvim.highlight.Constant",
			"nvim.highlight.Boolean",
			"nvim.highlight.@boolean",
			"nvim.highlight.@constant",
			"nvim.highlight.@constant.builtin",
		},
		"semantic.variable": {
			"nvim.highlight.Identifier",
			"nvim.highlight.@variable",
			"nvim.highlight.@variable.member",
			"nvim.highlight.@variable.parameter",
			"nvim.highlight.@parameter",
			"nvim.highlight.@property",
			"nvim.highlight.@field",
		},
		"semantic.builtin_variable": {
			"nvim.highlight.@variable.builtin",
			"nvim.highlight.@variable.parameter.builtin",
		},
	}
	for token, consumerIDs := range expected {
		for _, consumerID := range consumerIDs {
			assertPair(
				t,
				pairs,
				consumerID,
				token,
				"",
				true,
				verifycolors.ClassReportOnly,
			)
		}
	}
	assertPair(
		t,
		pairs,
		"bat.theme.scope.variable.language",
		"semantic.builtin_variable",
		"core.background",
		false,
		verifycolors.ClassEnforced,
	)
}

func TestGitRolesUseCanonicalTokensAcrossConsumers(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	nvimExpected := map[string]verifycolors.TokenRef{
		"nvim.highlight.GitSignsAdd":        "git.added",
		"nvim.highlight.GitSignsChange":     "git.changed",
		"nvim.highlight.GitSignsDelete":     "git.deleted",
		"nvim.highlight.GitSignsAddNr":      "git.added",
		"nvim.highlight.GitSignsChangeNr":   "git.changed",
		"nvim.highlight.GitSignsDeleteNr":   "git.deleted",
		"nvim.highlight.DiffAdd":            "git.added",
		"nvim.highlight.DiffChange":         "git.changed",
		"nvim.highlight.DiffDelete":         "git.deleted",
		"nvim.scrollbar.GitAdd":             "git.added",
		"nvim.scrollbar.GitChange":          "git.changed",
		"nvim.scrollbar.GitDelete":          "git.deleted",
		"nvim.highlight.ScrollbarGitAdd":    "git.added",
		"nvim.highlight.ScrollbarGitChange": "git.changed",
		"nvim.highlight.ScrollbarGitDelete": "git.deleted",
		"nvim.lualine.diff.added":           "git.added",
		"nvim.lualine.diff.modified":        "git.changed",
		"nvim.lualine.diff.removed":         "git.deleted",
	}
	if len(nvimExpected) != 18 {
		t.Fatalf("Neovim git consumer count = %d, want 18", len(nvimExpected))
	}
	for consumerID, token := range nvimExpected {
		pair, ok := pairs[consumerID]
		if !ok {
			t.Fatalf("pair %s not found", consumerID)
		}
		if pair.Foreground != token {
			t.Errorf("%s foreground = %q, want %q", consumerID, pair.Foreground, token)
		}
	}

	for consumerID, expected := range map[string]struct {
		foreground verifycolors.TokenRef
		background verifycolors.TokenRef
	}{
		"vim.highlight.DiffAdd":                {"git.added", "nvim.diff_add_bg"},
		"vim.highlight.DiffChange":             {"git.changed", "nvim.diff_change_bg"},
		"vim.highlight.DiffDelete":             {"git.deleted", "nvim.diff_delete_bg"},
		"hunk.theme.addedSignColor":            {"git.added", ""},
		"hunk.theme.removedSignColor":          {"git.deleted", ""},
		"hunk.theme.badgeAdded":                {"git.added", ""},
		"hunk.theme.badgeRemoved":              {"git.deleted", ""},
		"hunk.theme.fileNew":                   {"git.added", ""},
		"hunk.theme.fileModified":              {"git.changed", ""},
		"hunk.theme.fileDeleted":               {"git.deleted", ""},
		"delta.style.line-numbers-plus-style":  {"git.added", ""},
		"delta.style.line-numbers-minus-style": {"git.deleted", ""},
		"eza.theme.git.new":                    {"git.added", ""},
		"eza.theme.git.modified":               {"git.changed", ""},
		"eza.theme.git.deleted":                {"git.deleted", ""},
		"lazygit.theme.unstagedChangesColor":   {"git.changed", ""},
	} {
		class := verifycolors.ClassReportOnly
		ambient := true
		if expected.background != "" {
			class = verifycolors.ClassEnforced
			ambient = false
		}
		assertPair(t, pairs, consumerID, expected.foreground, expected.background, ambient, class)
	}

	assertPair(
		t,
		pairs,
		"delta.style.whitespace-error-style",
		"semantic.error",
		"",
		true,
		verifycolors.ClassReportOnly,
	)
	assertPair(
		t,
		pairs,
		"eza.theme.git.conflicted",
		"semantic.error",
		"",
		true,
		verifycolors.ClassReportOnly,
	)
}

func TestErrorGroupsUseCanonicalToken(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"nvim.highlight.DiagnosticError",
		"nvim.highlight.ErrorMsg",
		"nvim.highlight.RenderMarkdownError",
		"nvim.highlight.DiagnosticVirtualTextError",
		"nvim.highlight.DiagnosticUnderlineError.indicator",
		"nvim.highlight.DiagnosticSignError",
		"nvim.highlight.DiagnosticFloatingError",
		"nvim.highlight.NotifyERRORBorder",
		"nvim.highlight.NotifyERRORIcon",
		"nvim.highlight.NotifyERRORTitle",
		"nvim.highlight.TroubleCount",
		"nvim.highlight.TroubleError",
		"nvim.highlight.NeotestFailed",
		"nvim.highlight.ScrollbarError",
		"nvim.scrollbar.error",
		"nvim.bufferline.error",
		"nvim.bufferline.error_visible",
		"nvim.bufferline.error_selected",
	}
	if len(expected) != 18 {
		t.Fatalf("error consumer count = %d, want 18", len(expected))
	}
	pairs := pairMap(result.Pairs)
	for _, consumerID := range expected {
		pair, ok := pairs[consumerID]
		if !ok {
			t.Errorf("canonical error pair %s not found", consumerID)
			continue
		}
		if pair.Foreground != "semantic.error" {
			t.Errorf("%s foreground = %q, want semantic.error", consumerID, pair.Foreground)
		}
	}
}

func TestDiagnosticFamiliesUseCanonicalTokens(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)
	severities := map[string]verifycolors.TokenRef{
		"Error": "semantic.error",
		"Warn":  "semantic.warning",
		"Info":  "semantic.info",
		"Hint":  "semantic.hint",
	}

	count := 0
	for severity, token := range severities {
		for _, family := range []string{"Diagnostic", "DiagnosticVirtualText", "DiagnosticSign", "DiagnosticFloating"} {
			assertPair(
				t,
				pairs,
				"nvim.highlight."+family+severity,
				token,
				"",
				true,
				verifycolors.ClassReportOnly,
			)
			count++
		}
		assertPair(
			t,
			pairs,
			"nvim.highlight.DiagnosticUnderline"+severity+".indicator",
			token,
			"",
			true,
			verifycolors.ClassWaived,
		)
		count++
	}
	if count != 20 {
		t.Fatalf("Diagnostic pair count = %d, want 20", count)
	}

	for consumerID, token := range map[string]verifycolors.TokenRef{
		"nvim.highlight.ErrorMsg":       "semantic.error",
		"nvim.highlight.WarningMsg":     "semantic.warning",
		"vim.highlight.ErrorMsg":        "semantic.error",
		"vim.highlight.WarningMsg":      "semantic.warning",
		"vim.highlight.SpellBad":        "semantic.error",
		"vim.highlight.SpellCap":        "semantic.warning",
		"nvim.highlight.ScrollbarError": "semantic.error",
		"nvim.highlight.ScrollbarWarn":  "semantic.warning",
		"nvim.highlight.ScrollbarInfo":  "semantic.info",
		"nvim.highlight.ScrollbarHint":  "semantic.hint",
	} {
		assertPair(t, pairs, consumerID, token, "", true, verifycolors.ClassReportOnly)
	}
	for consumerID, token := range map[string]verifycolors.TokenRef{
		"nvim.scrollbar.error": "semantic.error",
		"nvim.scrollbar.warn":  "semantic.warning",
		"nvim.scrollbar.info":  "semantic.info",
		"nvim.scrollbar.hint":  "semantic.hint",
	} {
		assertPair(t, pairs, consumerID, token, "", true, verifycolors.ClassWaived)
	}

	severityConsumers := map[verifycolors.TokenRef][]string{
		"semantic.error": {
			"nvim.highlight.RenderMarkdownError",
			"nvim.highlight.NotifyERRORBorder",
			"nvim.highlight.NotifyERRORIcon",
			"nvim.highlight.NotifyERRORTitle",
			"nvim.highlight.TroubleCount",
			"nvim.highlight.TroubleError",
			"nvim.bufferline.error",
			"nvim.bufferline.error_visible",
			"nvim.bufferline.error_selected",
		},
		"semantic.warning": {
			"nvim.highlight.RenderMarkdownWarn",
			"nvim.highlight.NotifyWARNBorder",
			"nvim.highlight.NotifyWARNIcon",
			"nvim.highlight.NotifyWARNTitle",
			"nvim.highlight.TroubleWarning",
			"nvim.bufferline.warning",
			"nvim.bufferline.warning_visible",
			"nvim.bufferline.warning_selected",
		},
		"semantic.info": {
			"nvim.highlight.RenderMarkdownInfo",
			"nvim.highlight.NotifyINFOBorder",
			"nvim.highlight.NotifyINFOIcon",
			"nvim.highlight.NotifyINFOTitle",
			"nvim.highlight.TroubleInformation",
			"nvim.bufferline.info",
			"nvim.bufferline.info_visible",
			"nvim.bufferline.info_selected",
		},
		"semantic.hint": {
			"nvim.highlight.RenderMarkdownHint",
			"nvim.highlight.TroubleHint",
			"nvim.bufferline.hint",
			"nvim.bufferline.hint_visible",
			"nvim.bufferline.hint_selected",
		},
	}
	severityConsumerCount := 0
	for token, consumerIDs := range severityConsumers {
		for _, consumerID := range consumerIDs {
			assertPair(t, pairs, consumerID, token, "", true, verifycolors.ClassReportOnly)
			severityConsumerCount++
		}
	}
	if severityConsumerCount != 30 {
		t.Fatalf("severity consumer count = %d, want 30", severityConsumerCount)
	}
}

func TestZshErrorTokenAndReferencesAreRemoved(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	removedToken := "zsh" + ".error"
	if _, ok := colorPalette.TokenValues()[verifycolors.TokenRef(removedToken)]; ok {
		t.Fatalf("%s still exists in the palette", removedToken)
	}
	for _, relative := range []string{
		"scripts/cmd/generate-colors/main.go",
		"scripts/internal/verifycolors/sourceinventory/aliases.go",
		"scripts/internal/verifycolors/inventory_generated.go",
		"home-manager/zsh.nix",
		"home-manager/colors.nix",
		"nvim/lua/config/palette.lua",
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), removedToken) {
			t.Errorf("%s still references %s", relative, removedToken)
		}
	}
}

func TestSemanticHintIsExported(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"home-manager/colors.nix", "wezterm/colors.lua"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `hint = "#58CAF8"`) {
			t.Errorf("%s does not export semantic.hint", relative)
		}
	}
}

func TestSemanticBuiltinVariableIsExported(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"home-manager/colors.nix", "wezterm/colors.lua"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `builtin_variable = "#88CBEA"`) {
			t.Errorf("%s does not export semantic.builtin_variable", relative)
		}
	}
}

func TestNvimSyntaxSemanticContrastOnBrightAmbient(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	values := colorPalette.TokenValues()
	background := values["core.panel_bg"]
	minimumToken := verifycolors.TokenRef("")
	minimumRatio := math.Inf(1)
	for _, token := range []verifycolors.TokenRef{
		"semantic.function",
		"semantic.type",
		"semantic.number",
		"semantic.constant",
		"semantic.variable",
		"semantic.builtin_variable",
	} {
		ratio, err := colorutil.ContrastRatio(values[token], background)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s on core.panel_bg = %.4f", token, ratio)
		if ratio < 4.5 {
			t.Errorf("%s on core.panel_bg contrast = %.4f, want >= 4.5", token, ratio)
		}
		if ratio < minimumRatio {
			minimumToken = token
			minimumRatio = ratio
		}
	}
	if minimumToken != "semantic.type" || math.Abs(minimumRatio-4.864) > 0.001 {
		t.Errorf("minimum contrast = %s %.4f, want semantic.type 4.864", minimumToken, minimumRatio)
	}
}

func TestUIAccentRolesAreSeparated(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	for _, consumerID := range []string{"nvim.highlight.Substitute", "nvim.highlight.FlashLabel"} {
		assertPair(
			t,
			pairs,
			consumerID,
			"core.background",
			"ui.target_bg",
			false,
			verifycolors.ClassEnforced,
		)
	}
	for _, consumerID := range []string{
		"nvim.highlight.RenderMarkdownTodo",
		"nvim.bufferline.close_button_selected",
		"nvim.bufferline.pick",
		"nvim.bufferline.pick_visible",
		"nvim.bufferline.pick_selected",
		"zsh.fzf.prompt",
		"zsh.fzf.spinner",
	} {
		assertPair(t, pairs, consumerID, "ui.accent_fg", "", true, verifycolors.ClassReportOnly)
	}
	assertPair(
		t,
		pairs,
		"nvim.highlight.OilLink",
		"semantic.keyword",
		"",
		true,
		verifycolors.ClassReportOnly,
	)
}

func TestGitAndUITargetContrastRatios(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	values := colorPalette.TokenValues()
	for _, test := range []struct {
		name       string
		foreground verifycolors.TokenRef
		background verifycolors.TokenRef
		want       float64
	}{
		{"DiffAdd", "git.added", "nvim.diff_add_bg", 6.109},
		{"DiffChange", "git.changed", "nvim.diff_change_bg", 7.146},
		{"DiffDelete", "git.deleted", "nvim.diff_delete_bg", 7.618},
		{"Substitute/FlashLabel", "core.background", "ui.target_bg", 7.265},
	} {
		t.Run(test.name, func(t *testing.T) {
			ratio, err := colorutil.ContrastRatio(values[test.foreground], values[test.background])
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s on %s = %.4f", test.foreground, test.background, ratio)
			if math.Abs(ratio-test.want) > 0.001 {
				t.Errorf("%s/%s contrast = %.4f, want %.3f", test.foreground, test.background, ratio, test.want)
			}
		})
	}
}

func TestCanonicalDiagnosticContrastProfiles(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	values := colorPalette.TokenValues()
	if values["semantic.error"] != values["git.changed"] {
		t.Fatalf("semantic.error = %s, git.changed = %s; want equal rendered colors", values["semantic.error"], values["git.changed"])
	}

	batRatio, err := colorutil.ContrastRatio(values["semantic.error"], values["core.background"])
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(batRatio-7.265) > 0.001 {
		t.Errorf("bat invalid.illegal truecolor ratio = %.4f, want 7.265", batRatio)
	}
	t.Logf("bat invalid.illegal truecolor: %s on %s = %.4f", values["semantic.error"], values["core.background"], batRatio)

	vimRatio, err := colorutil.ContrastRatio(values["git.changed"], values["nvim.diff_change_bg"])
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(vimRatio-7.146) > 0.001 {
		t.Errorf("Vim DiffChange truecolor ratio = %.4f, want 7.146", vimRatio)
	}
	t.Logf("Vim DiffChange truecolor: %s on %s = %.4f", values["git.changed"], values["nvim.diff_change_bg"], vimRatio)

	foregroundIndex, _, err := xterm.Nearest256(values["git.changed"])
	if err != nil {
		t.Fatal(err)
	}
	backgroundIndex, _, err := xterm.Nearest256(values["nvim.diff_change_bg"])
	if err != nil {
		t.Fatal(err)
	}
	if foregroundIndex != 182 || backgroundIndex != 237 {
		t.Errorf("Vim DiffChange cterm indices = %d/%d, want 182/237", foregroundIndex, backgroundIndex)
	}
	ansi, err := colorPalette.WezTermANSI()
	if err != nil {
		t.Fatal(err)
	}
	foreground, err := xterm.Resolve256(foregroundIndex, ansi)
	if err != nil {
		t.Fatal(err)
	}
	background, err := xterm.Resolve256(backgroundIndex, ansi)
	if err != nil {
		t.Fatal(err)
	}
	ctermRatio, err := colorutil.ContrastRatio(foreground.Hex(), background.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(ctermRatio-5.960) > 0.001 {
		t.Errorf("Vim DiffChange cterm ratio = %.4f, want 5.960", ctermRatio)
	}
	t.Logf("Vim DiffChange cterm: index %d %s on index %d %s = %.4f", foregroundIndex, foreground.Hex(), backgroundIndex, background.Hex(), ctermRatio)
}

func TestNvimTerminalColorsUseANSIOnly(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expected := []verifycolors.TokenRef{
		"ansi.black",
		"ansi.red",
		"ansi.green",
		"ansi.yellow",
		"ansi.blue",
		"ansi.magenta",
		"ansi.cyan",
		"ansi.white",
		"ansi.bright_black",
		"ansi.bright_red",
		"ansi.bright_green",
		"ansi.bright_yellow",
		"ansi.bright_blue",
		"ansi.bright_magenta",
		"ansi.bright_cyan",
		"ansi.bright_white",
	}
	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "nvim.terminal.color") {
			count++
		}
	}
	if count != len(expected) {
		t.Fatalf("Neovim terminal color pair count = %d, want %d", count, len(expected))
	}
	for index, token := range expected {
		assertPair(
			t,
			pairs,
			fmt.Sprintf("nvim.terminal.color%d", index),
			token,
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}
}

func TestGhDashThemeColorsUseExpectedTokens(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expected := map[string]verifycolors.TokenRef{
		"text.primary":        "foregrounds.main",
		"text.secondary":      "foregrounds.dim",
		"text.inverted":       "core.darkest_bg",
		"text.faint":          "foregrounds.subdued",
		"text.warning":        "semantic.warning",
		"text.success":        "semantic.success",
		"text.error":          "semantic.error",
		"text.actor":          "foregrounds.heading",
		"background.selected": "core.active_line",
		"border.primary":      "teals.border",
		"border.secondary":    "teals.border",
		"border.faint":        "core.ui_shadow",
		"icon.newcontributor": "semantic.success",
		"icon.contributor":    "foregrounds.heading",
		"icon.collaborator":   "semantic.warning",
		"icon.member":         "semantic.warning",
		"icon.owner":          "semantic.warning",
	}
	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "gh-dash.theme.") {
			count++
		}
	}
	if count != len(expected) {
		t.Fatalf("gh-dash theme pair count = %d, want %d", count, len(expected))
	}
	for fieldPath, token := range expected {
		assertPair(
			t,
			pairs,
			"gh-dash.theme."+fieldPath,
			token,
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}
}

func TestDeltaStylesUseExpectedTokens(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expectedAmbient := map[string]verifycolors.TokenRef{
		"plus-style":                    "nvim.diff_add_bg",
		"minus-style":                   "nvim.diff_delete_bg",
		"line-numbers-plus-style":       "git.added",
		"line-numbers-minus-style":      "git.deleted",
		"line-numbers-zero-style":       "foregrounds.subdued",
		"line-numbers-left-style":       "foregrounds.subdued",
		"line-numbers-right-style":      "foregrounds.subdued",
		"hunk-header-line-number-style": "foregrounds.subdued",
		"hunk-header-decoration-style":  "teals.border",
		"file-style":                    "foregrounds.heading",
		"whitespace-error-style":        "semantic.error",
	}
	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "delta.style.") {
			count++
		}
	}
	if count != len(expectedAmbient)+2 {
		t.Fatalf("delta style pair count = %d, want %d", count, len(expectedAmbient)+2)
	}
	for option, token := range expectedAmbient {
		assertPair(
			t,
			pairs,
			"delta.style."+option,
			token,
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}
	assertPair(
		t,
		pairs,
		"delta.style.plus-emph-style",
		"ansi.bright_white",
		"nvim.diff_add_inline_bg",
		false,
		verifycolors.ClassEnforced,
	)
	assertPair(
		t,
		pairs,
		"delta.style.minus-emph-style",
		"ansi.bright_white",
		"nvim.diff_delete_inline_bg",
		false,
		verifycolors.ClassEnforced,
	)
	if _, ok := pairs["delta.style.syntax-theme"]; ok {
		t.Error("delta.style.syntax-theme must not be represented as a color pair")
	}
	if _, ok := pairs["delta.style.hunk-header-file-style"]; ok {
		t.Error("delta.style.hunk-header-file-style must remain unset")
	}

	data, err := os.ReadFile(filepath.Join(root, "home-manager", "git.nix"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "hunk-header-file-style =") {
		t.Error("programs.delta.options must not define hunk-header-file-style")
	}
	if !strings.Contains(string(data), `syntax-theme = "ghost-visor";`) {
		t.Error(`programs.delta.options must define syntax-theme = "ghost-visor"`)
	}
	data, err = os.ReadFile(filepath.Join(root, "lazygit", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "--syntax-theme") {
		t.Error("lazygit delta pager must not override syntax-theme")
	}
}

func TestBatThemeUsesExpectedTokensAndExplicitBackgrounds(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expectedScopes := map[string]verifycolors.TokenRef{
		"comment":                        "semantic.comment",
		"string":                         "semantic.string",
		"constant.numeric":               "semantic.number",
		"constant.language":              "semantic.constant",
		"constant.character.escape":      "foregrounds.bright",
		"keyword":                        "semantic.keyword",
		"keyword.operator":               "semantic.operator",
		"keyword.control.import":         "teals.bright",
		"storage.type":                   "semantic.type",
		"storage.modifier":               "semantic.keyword",
		"entity.name.function":           "semantic.function",
		"variable.function":              "semantic.function",
		"support.function":               "semantic.function",
		"entity.name.class":              "semantic.type",
		"entity.name.type":               "semantic.type",
		"support.type":                   "semantic.type",
		"entity.name.tag":                "teals.bright",
		"entity.other.attribute-name":    "foregrounds.heading",
		"variable.parameter":             "semantic.variable",
		"variable.language":              "semantic.builtin_variable",
		"punctuation.separator":          "semantic.punctuation",
		"punctuation.terminator":         "semantic.punctuation",
		"punctuation.section":            "semantic.punctuation",
		"punctuation.definition":         "semantic.punctuation",
		"punctuation.accessor":           "semantic.punctuation",
		"invalid.illegal":                "semantic.error",
		"entity.name.tag.yaml":           "teals.bright",
		"constant.language.boolean.yaml": "semantic.keyword",
		"constant.language.null.yaml":    "semantic.keyword",
		"source.yaml constant.numeric":   "purples.lavender",
		"punctuation.definition.block.sequence.item.yaml": "semantic.punctuation",
		"markup.heading":        "foregrounds.heading",
		"markup.bold":           "ansi.bright_white",
		"markup.italic":         "foregrounds.bright",
		"markup.raw":            "semantic.string",
		"markup.quote":          "foregrounds.dim",
		"markup.list":           "semantic.punctuation",
		"markup.underline.link": "teals.bright",
	}
	if len(batGlobalTokens) != 4 {
		t.Fatalf("bat global setting count = %d, want 4", len(batGlobalTokens))
	}
	if len(batScopeTokens) != len(expectedScopes) {
		t.Fatalf("bat scope count = %d, want %d", len(batScopeTokens), len(expectedScopes))
	}

	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "bat.theme.") {
			count++
		}
	}
	if count != len(expectedScopes)+3 {
		t.Fatalf("bat theme pair count = %d, want %d", count, len(expectedScopes)+3)
	}

	assertPair(
		t,
		pairs,
		"bat.theme.foreground",
		"foregrounds.main",
		"core.background",
		false,
		verifycolors.ClassEnforced,
	)
	assertPair(
		t,
		pairs,
		"bat.theme.gutterForeground",
		"foregrounds.dim",
		"core.background",
		false,
		verifycolors.ClassEnforced,
	)
	assertPair(
		t,
		pairs,
		"bat.theme.gutterForeground.lineHighlight",
		"foregrounds.dim",
		"core.active_line",
		false,
		verifycolors.ClassEnforced,
	)
	for scope, token := range expectedScopes {
		assertPair(
			t,
			pairs,
			"bat.theme.scope."+strings.ReplaceAll(scope, " ", "_"),
			token,
			"core.background",
			false,
			verifycolors.ClassEnforced,
		)
	}
}

func TestEzaAndZshCompletionUseSharedExpectedTokens(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expectedByToken := map[verifycolors.TokenRef][]string{
		"foregrounds.main": {
			"filekinds.normal",
			"size.number_byte", "size.number_kilo", "size.number_mega", "size.number_giga", "size.number_huge",
		},
		"foregrounds.heading": {
			"filekinds.directory", "git_repo.branch_main", "header",
		},
		"semantic.keyword": {
			"filekinds.symlink", "git.renamed", "git.typechange", "git_repo.branch_other",
		},
		"semantic.operator": {
			"filekinds.pipe", "filekinds.socket",
		},
		"foregrounds.dim": {
			"filekinds.block_device", "filekinds.char_device",
			"perms.user_read", "perms.group_read", "perms.other_read",
			"size.major", "size.minor",
			"users.user_other", "users.group_other",
			"links.normal", "file_type.compiled", "date", "symlink_path",
		},
		"teals.mid_bright": {
			"filekinds.special", "filekinds.mount_point", "file_type.build",
		},
		"semantic.success": {
			"filekinds.executable",
			"perms.user_execute_file", "perms.user_execute_other", "perms.group_execute", "perms.other_execute",
			"users.user_you", "users.group_yours",
			"git_repo.git_clean",
		},
		"semantic.warning": {
			"perms.user_write", "perms.group_write", "perms.other_write",
			"links.multi_link_file", "git_repo.git_dirty",
		},
		"purples.bright_purple": {
			"perms.special_user_file", "perms.special_other", "file_type.music", "file_type.lossless",
		},
		"foregrounds.subdued": {
			"perms.attribute",
			"size.unit_byte", "size.unit_kilo", "size.unit_mega", "size.unit_giga", "size.unit_huge",
			"git.ignored",
			"security_context.none",
			"security_context.selinux.colon", "security_context.selinux.user", "security_context.selinux.role",
			"security_context.selinux.typ", "security_context.selinux.range",
			"file_type.temp",
			"inode", "blocks", "octal", "flags",
		},
		"semantic.error": {
			"users.user_root", "users.group_root",
			"git.conflicted",
			"control_char", "broken_symlink", "broken_path_overlay",
		},
		"git.added":                {"git.new"},
		"git.changed":              {"git.modified"},
		"git.deleted":              {"git.deleted"},
		"purples.lavender":         {"file_type.image"},
		"purples.muted_purple":     {"file_type.video"},
		"ansi.bright_yellow":       {"file_type.crypto"},
		"blues_slates.cloud_slate": {"file_type.document"},
		"ansi.bright_red":          {"file_type.compressed"},
		"semantic.punctuation":     {"punctuation"},
		"semantic.type":            {"file_type.source"},
	}

	ezaCount := 0
	expectedEzaCount := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "eza.theme.") {
			ezaCount++
		}
	}
	for token, paths := range expectedByToken {
		expectedEzaCount += len(paths)
		for _, path := range paths {
			assertPair(
				t,
				pairs,
				"eza.theme."+path,
				token,
				"",
				true,
				verifycolors.ClassReportOnly,
			)
		}
	}
	if expectedEzaCount != 82 {
		t.Fatalf("test eza schema key count = %d, want 82", expectedEzaCount)
	}
	if ezaCount != expectedEzaCount {
		t.Fatalf("eza theme pair count = %d, want %d", ezaCount, expectedEzaCount)
	}
	if ezaThemeSchema["header"].modifier != "bold" {
		t.Errorf("eza header modifier = %q, want bold", ezaThemeSchema["header"].modifier)
	}

	expectedClasses := map[string]verifycolors.TokenRef{
		"fi": "foregrounds.main",
		"di": "foregrounds.heading",
		"ln": "semantic.keyword",
		"ex": "semantic.success",
		"or": "semantic.error",
		"pi": "semantic.operator",
		"so": "semantic.operator",
		"bd": "foregrounds.dim",
		"cd": "foregrounds.dim",
	}
	expectedExtensions := map[verifycolors.TokenRef][]string{
		"purples.lavender":         {"png", "jpg", "svg"},
		"purples.muted_purple":     {"mp4", "mkv"},
		"purples.bright_purple":    {"mp3", "ogg", "flac", "wav"},
		"ansi.bright_yellow":       {"age", "pem"},
		"blues_slates.cloud_slate": {"pdf", "key"},
		"foregrounds.subdued":      {"tmp", "bak"},
		"ansi.bright_red":          {"zip", "gz", "tar", "tar.gz"},
		"foregrounds.dim":          {"so", "o"},
		"teals.mid_bright":         {"ninja"},
		"semantic.type":            {"go", "rs", "py", "ts", "lua", "js"},
	}
	zshCount := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "zsh.completion.") {
			zshCount++
		}
	}
	for code, token := range expectedClasses {
		assertPair(t, pairs, "zsh.completion.class."+code, token, "", true, verifycolors.ClassReportOnly)
	}
	expectedExtensionCount := 0
	for token, extensions := range expectedExtensions {
		expectedExtensionCount += len(extensions)
		for _, extension := range extensions {
			assertPair(
				t,
				pairs,
				"zsh.completion.extension."+strings.ReplaceAll(extension, ".", "_"),
				token,
				"",
				true,
				verifycolors.ClassReportOnly,
			)
		}
	}
	assertPair(
		t,
		pairs,
		"zsh.completion.menu-select",
		"core.selection_fg",
		"core.selection_bg",
		false,
		verifycolors.ClassEnforced,
	)
	if want := len(expectedClasses) + expectedExtensionCount + 1; zshCount != want {
		t.Fatalf("zsh completion pair count = %d, want %d", zshCount, want)
	}

	for _, note := range result.CoverageNotes {
		if note.ID == "zsh.completion.file-type-subset" && strings.Contains(note.Reason, "README") {
			return
		}
	}
	t.Fatal("zsh completion file-type subset coverage note not found")
}

func TestEzaFileTypeSemanticGroupsUseDistinctColors(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "eza", "theme.yml"))
	if err != nil {
		t.Fatal(err)
	}

	colors := make(map[string]string)
	inFileType := false
	currentClass := ""
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case line == "file_type:":
			inFileType = true
		case inFileType && line != "" && line[0] != ' ':
			inFileType = false
		case inFileType && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    "):
			currentClass = strings.TrimSuffix(strings.TrimSpace(line), ":")
		case inFileType && strings.HasPrefix(line, "    foreground: "):
			colors[currentClass] = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "    foreground: ")), `"`)
		}
	}
	if len(colors) != 11 {
		t.Fatalf("eza file_type color count = %d, want 11", len(colors))
	}

	semanticGroups := map[string][]string{
		"image": {"image"},
		"video": {"video"},
		// Audio files intentionally share one color regardless of lossy/lossless encoding.
		"audio":      {"music", "lossless"},
		"crypto":     {"crypto"},
		"document":   {"document"},
		"compressed": {"compressed"},
		"temp":       {"temp"},
		"compiled":   {"compiled"},
		"build":      {"build"},
		"source":     {"source"},
	}
	if len(semanticGroups) != 10 {
		t.Fatalf("eza file_type semantic group count = %d, want 10", len(semanticGroups))
	}

	seenColors := make(map[string]string)
	for group, classes := range semanticGroups {
		groupColor := colors[classes[0]]
		if groupColor == "" {
			t.Fatalf("eza file_type class %q has no generated color", classes[0])
		}
		for _, class := range classes[1:] {
			if colors[class] != groupColor {
				t.Fatalf("eza file_type semantic group %q: %s = %q, want %q", group, class, colors[class], groupColor)
			}
		}
		if previousGroup, ok := seenColors[groupColor]; ok {
			t.Fatalf("eza file_type semantic groups %q and %q both use %s", previousGroup, group, groupColor)
		}
		seenColors[groupColor] = group
	}
}

func TestEzaStyleSpecRejectsUnknownDuplicateAndMissingPaths(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	template, startLine, err := readRawStringConstant(
		filepath.Join(root, "scripts", "cmd", "generate-colors", "main.go"),
		"ezaStyleSpecTemplate",
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"unknown":   strings.Replace(template, "filekinds.normal|", "filekinds.typo|", 1),
		"duplicate": template + "\nfilekinds.normal|foregrounds.main||fi2|\n",
		"missing":   strings.Replace(template, "filekinds.normal|foregrounds.main||fi|\n", "", 1),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseEzaSourceSpecs(input, "test", startLine); err == nil {
				t.Fatal("parseEzaSourceSpecs succeeded, want error")
			}
		})
	}
}

func TestMarkdownPreviewUsesExpectedTokensAndBackgrounds(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	type expectedPair struct {
		foreground verifycolors.TokenRef
		background verifycolors.TokenRef
	}
	expected := map[string]expectedPair{
		"markdown.text-primary":             {foreground: "foregrounds.main", background: "core.background"},
		"markdown.blockquote":               {foreground: "foregrounds.dim", background: "core.background"},
		"markdown.link":                     {foreground: "teals.bright", background: "core.background"},
		"markdown.h1":                       {foreground: "foregrounds.heading", background: "core.background"},
		"markdown.h2":                       {foreground: "teals.bright", background: "core.background"},
		"markdown.h3":                       {foreground: "teals.mid_bright", background: "core.background"},
		"markdown.h4":                       {foreground: "foregrounds.main", background: "core.background"},
		"markdown.h5":                       {foreground: "blues_slates.cloud_slate", background: "core.background"},
		"markdown.h6":                       {foreground: "foregrounds.subdued", background: "core.background"},
		"markdown.table":                    {foreground: "foregrounds.main", background: "core.panel_bg"},
		"markdown.kbd":                      {foreground: "foregrounds.dim", background: "core.ui_shadow"},
		"markdown.page-header":              {foreground: "foregrounds.main", background: "core.background"},
		"highlight.hljs":                    {foreground: "foregrounds.main", background: "core.panel_bg"},
		"highlight.hljs-comment":            {foreground: "semantic.comment", background: "core.panel_bg"},
		"highlight.hljs-quote":              {foreground: "semantic.comment", background: "core.panel_bg"},
		"highlight.hljs-keyword":            {foreground: "semantic.keyword", background: "core.panel_bg"},
		"highlight.hljs-selector-tag":       {foreground: "semantic.keyword", background: "core.panel_bg"},
		"highlight.hljs-subst":              {foreground: "semantic.keyword", background: "core.panel_bg"},
		"highlight.hljs-number":             {foreground: "semantic.number", background: "core.panel_bg"},
		"highlight.hljs-literal":            {foreground: "semantic.number", background: "core.panel_bg"},
		"highlight.hljs-variable":           {foreground: "semantic.variable", background: "core.panel_bg"},
		"highlight.hljs-template-variable":  {foreground: "semantic.variable", background: "core.panel_bg"},
		"highlight.hljs-tag_hljs-attr":      {foreground: "semantic.variable", background: "core.panel_bg"},
		"highlight.hljs-string":             {foreground: "semantic.string", background: "core.panel_bg"},
		"highlight.hljs-doctag":             {foreground: "semantic.string", background: "core.panel_bg"},
		"highlight.hljs-title":              {foreground: "semantic.function", background: "core.panel_bg"},
		"highlight.hljs-section":            {foreground: "semantic.function", background: "core.panel_bg"},
		"highlight.hljs-selector-id":        {foreground: "semantic.function", background: "core.panel_bg"},
		"highlight.hljs-type":               {foreground: "semantic.type", background: "core.panel_bg"},
		"highlight.hljs-class_hljs-title":   {foreground: "semantic.type", background: "core.panel_bg"},
		"highlight.hljs-tag":                {foreground: "foregrounds.heading", background: "core.panel_bg"},
		"highlight.hljs-name":               {foreground: "foregrounds.heading", background: "core.panel_bg"},
		"highlight.hljs-attribute":          {foreground: "foregrounds.heading", background: "core.panel_bg"},
		"highlight.hljs-regexp":             {foreground: "teals.bright", background: "core.panel_bg"},
		"highlight.hljs-link":               {foreground: "teals.bright", background: "core.panel_bg"},
		"highlight.hljs-symbol":             {foreground: "semantic.punctuation", background: "core.panel_bg"},
		"highlight.hljs-bullet":             {foreground: "semantic.punctuation", background: "core.panel_bg"},
		"highlight.hljs-built_in":           {foreground: "teals.mid_bright", background: "core.panel_bg"},
		"highlight.hljs-builtin-name":       {foreground: "teals.mid_bright", background: "core.panel_bg"},
		"highlight.hljs-meta":               {foreground: "foregrounds.dim", background: "core.panel_bg"},
		"highlight.hljs-deletion.inherited": {foreground: "foregrounds.main", background: "nvim.diff_delete_bg"},
		"highlight.hljs-addition.inherited": {foreground: "foregrounds.main", background: "nvim.diff_add_bg"},
	}

	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "nvim.markdown-preview.") {
			count++
		}
	}
	if count != len(expected) {
		t.Fatalf("Markdown Preview pair count = %d, want %d", count, len(expected))
	}
	for suffix, pair := range expected {
		assertPair(
			t,
			pairs,
			"nvim.markdown-preview."+suffix,
			pair.foreground,
			pair.background,
			false,
			verifycolors.ClassEnforced,
		)
	}
}

func TestHerdrThemeUsesExpectedTokensAndFixedValues(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expected := map[string]verifycolors.TokenRef{
		"accent":      "ansi.blue",
		"surface1":    "core.active_line",
		"surface_dim": "ansi.bright_black",
		"overlay0":    "foregrounds.dim",
		"overlay1":    "ansi.bright_white",
		"text":        "foregrounds.main",
		"subtext0":    "foregrounds.dim",
		"mauve":       "foregrounds.heading",
		"green":       "ansi.green",
		"yellow":      "ansi.yellow",
		"red":         "ansi.bright_red",
		"blue":        "ansi.blue",
		"teal":        "ansi.cyan",
	}
	if len(herdrThemeSpecs) != 16 {
		t.Fatalf("herdr theme key count = %d, want 16", len(herdrThemeSpecs))
	}
	count := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "herdr.theme.") {
			count++
		}
	}
	if count != len(expected) {
		t.Fatalf("herdr theme pair count = %d, want %d", count, len(expected))
	}
	for key, token := range expected {
		assertPair(
			t,
			pairs,
			"herdr.theme."+key,
			token,
			"",
			true,
			verifycolors.ClassReportOnly,
		)
	}
	for _, key := range []string{"panel_bg", "surface0", "peach"} {
		if _, ok := pairs["herdr.theme."+key]; ok {
			t.Errorf("herdr.theme.%s must not be represented as a visible pair", key)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, "scripts", "cmd", "generate-colors", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	for _, assignment := range []string{
		`panel_bg = "reset"`,
		`surface0 = "reset"`,
		`peach = "{{semantic.warning}}" # herdr 0.7.4では未使用`,
	} {
		if !strings.Contains(contents, assignment) {
			t.Errorf("herdrTemplate does not contain %q", assignment)
		}
	}

	notes := make(map[string]verifycolors.CoverageNote, len(result.CoverageNotes))
	for _, note := range result.CoverageNotes {
		notes[note.ID] = note
	}
	for _, noteID := range []string{
		"herdr.theme.panel_bg.reset",
		"herdr.theme.surface0.reset",
		"herdr.theme.peach.unused",
	} {
		note, ok := notes[noteID]
		if !ok {
			t.Errorf("coverage note %s not found", noteID)
			continue
		}
		if note.Reason == "" {
			t.Errorf("coverage note %s has no reason", noteID)
		}
	}
}

func TestWezTermKeyTableIndicatorsUseExpectedPairs(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	expected := map[string]struct {
		foreground verifycolors.TokenRef
		background verifycolors.TokenRef
	}{
		"copy_mode":       {foreground: "ansi.bright_white", background: "blues_slates.slate_mid"},
		"resize_pane":     {foreground: "core.darkest_bg", background: "semantic.string"},
		"pane_navigation": {foreground: "core.darkest_bg", background: "foregrounds.dim"},
		"search_mode":     {foreground: "core.darkest_bg", background: "semantic.operator"},
		"other":           {foreground: "core.darkest_bg", background: "purples.muted_purple"},
	}
	pairCount := 0
	for _, pair := range result.Pairs {
		if strings.HasPrefix(pair.ConsumerID, "wezterm.key_table.") {
			pairCount++
		}
	}
	if pairCount != len(expected)*2 {
		t.Fatalf("WezTerm key-table pair count = %d, want %d", pairCount, len(expected)*2)
	}

	for name, want := range expected {
		consumerID := "wezterm.key_table." + name
		assertPair(
			t,
			pairs,
			consumerID,
			want.foreground,
			want.background,
			false,
			verifycolors.ClassEnforced,
		)
		pair := pairs[consumerID]
		if pair.Role != verifycolors.RoleText {
			t.Errorf("%s role = %q, want %q", consumerID, pair.Role, verifycolors.RoleText)
		}
		if len(pair.Profiles) != 1 || pair.Profiles[0] != verifycolors.ProfileTruecolor {
			t.Errorf("%s profiles = %v, want [truecolor]", consumerID, pair.Profiles)
		}

		surfaceID := consumerID + ".surface"
		assertPair(
			t,
			pairs,
			surfaceID,
			want.background,
			"",
			true,
			verifycolors.ClassReportOnly,
		)
		surface := pairs[surfaceID]
		if surface.Role != verifycolors.RoleSurface {
			t.Errorf("%s role = %q, want %q", surfaceID, surface.Role, verifycolors.RoleSurface)
		}
		if len(surface.Profiles) != 1 || surface.Profiles[0] != verifycolors.ProfileTruecolor {
			t.Errorf("%s profiles = %v, want [truecolor]", surfaceID, surface.Profiles)
		}
	}
}

func TestWezTermKeyTableExtractorRejectsMalformedBranches(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "wezterm", "keybinds.lua"))
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	copyForeground := `table.insert(elements, { Foreground = { Color = colors.ansi.bright_white } })`
	resizeBackground := `table.insert(elements, { Background = { Color = colors.semantic.string } })`
	tests := map[string]string{
		"missing": strings.Replace(contents, copyForeground, "", 1),
		"duplicate": strings.Replace(
			contents,
			copyForeground,
			copyForeground+"\n\t\t\t"+copyForeground,
			1,
		),
		"cross-branch": strings.Replace(
			strings.Replace(contents, copyForeground, "", 1),
			resizeBackground,
			resizeBackground+"\n\t\t\t"+copyForeground,
			1,
		),
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseWezTermKeyTables(input, "wezterm/keybinds.lua"); err == nil {
				t.Fatal("parseWezTermKeyTables succeeded, want error")
			}
		})
	}
}

func TestWezTermKeyTableIndicatorContrastRatios(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	values := colorPalette.TokenValues()
	for _, test := range []struct {
		name       string
		foreground verifycolors.TokenRef
		background verifycolors.TokenRef
		want       float64
	}{
		{"copy_mode", "ansi.bright_white", "blues_slates.slate_mid", 8.583},
		{"resize_pane", "core.darkest_bg", "semantic.string", 8.723},
		{"pane_navigation", "core.darkest_bg", "foregrounds.dim", 8.693},
		{"search_mode", "core.darkest_bg", "semantic.operator", 8.720},
		{"other", "core.darkest_bg", "purples.muted_purple", 8.693},
	} {
		t.Run(test.name, func(t *testing.T) {
			ratio, err := colorutil.ContrastRatio(values[test.foreground], values[test.background])
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s on %s = %.4f", test.foreground, test.background, ratio)
			if math.Abs(ratio-test.want) > 0.001 {
				t.Errorf("%s/%s contrast = %.4f, want %.3f", test.foreground, test.background, ratio, test.want)
			}
		})
	}
}

func TestWezTermRuntimeCoverageGapIsExplicit(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, note := range result.CoverageNotes {
		if note.ID != "wezterm.runtime-dynamic-colors" {
			continue
		}
		for _, phrase := range []string{
			"window_background_gradient",
			"compose_cursor",
			"format-tab-title",
			"Claude state icon",
			"dynamic Lua branches",
			"could diverge without detection",
		} {
			if !strings.Contains(note.Reason, phrase) {
				t.Errorf("coverage note reason does not mention %q: %s", phrase, note.Reason)
			}
		}
		return
	}
	t.Fatal("WezTerm runtime dynamic-color coverage note not found")
}

func TestBufferlineRuntimeCoverageGapIsExplicit(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	staticCount := 0
	for _, pair := range result.Pairs {
		if len(pair.ConsumerID) >= len("nvim.bufferline.") &&
			pair.ConsumerID[:len("nvim.bufferline.")] == "nvim.bufferline." {
			staticCount++
		}
	}
	if staticCount != 37 {
		t.Fatalf("static BufferLine pair count = %d, want 37", staticCount)
	}
	for _, note := range result.CoverageNotes {
		if note.ID == "nvim.bufferline.runtime-default-groups" {
			return
		}
	}
	t.Fatal("runtime BufferLine coverage note not found")
}

func pairMap(pairs []verifycolors.PairSpec) map[string]verifycolors.PairSpec {
	result := make(map[string]verifycolors.PairSpec, len(pairs))
	for _, pair := range pairs {
		result[pair.ConsumerID] = pair
	}
	return result
}

func assertPair(
	t *testing.T,
	pairs map[string]verifycolors.PairSpec,
	consumerID string,
	foreground verifycolors.TokenRef,
	background verifycolors.TokenRef,
	ambient bool,
	class verifycolors.PairClass,
) {
	t.Helper()
	pair, ok := pairs[consumerID]
	if !ok {
		t.Fatalf("pair %s not found", consumerID)
	}
	if pair.Foreground != foreground {
		t.Errorf("%s foreground = %q, want %q", consumerID, pair.Foreground, foreground)
	}
	if pair.Background.Token != background || pair.Background.Ambient != ambient {
		t.Errorf(
			"%s background = {token:%q ambient:%t}, want {token:%q ambient:%t}",
			consumerID,
			pair.Background.Token,
			pair.Background.Ambient,
			background,
			ambient,
		)
	}
	if pair.Class != class {
		t.Errorf("%s class = %q, want %q", consumerID, pair.Class, class)
	}
}
