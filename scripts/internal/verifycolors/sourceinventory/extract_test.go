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

func TestMagentaAliasRemainsDiagnosticOnly(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"nvim.highlight.DiagnosticError":                    true,
		"nvim.highlight.ErrorMsg":                           true,
		"nvim.highlight.RenderMarkdownError":                true,
		"nvim.highlight.DiagnosticVirtualTextError":         true,
		"nvim.highlight.DiagnosticUnderlineError.indicator": true,
		"nvim.highlight.DiagnosticSignError":                true,
		"nvim.highlight.DiagnosticFloatingError":            true,
		"nvim.highlight.NotifyERRORBorder":                  true,
		"nvim.highlight.NotifyERRORIcon":                    true,
		"nvim.highlight.NotifyERRORTitle":                   true,
		"nvim.highlight.TroubleCount":                       true,
		"nvim.highlight.TroubleError":                       true,
		"nvim.highlight.NeotestFailed":                      true,
		"nvim.highlight.ScrollbarError":                     true,
		"nvim.scrollbar.error":                              true,
		"nvim.bufferline.error":                             true,
		"nvim.bufferline.error_visible":                     true,
		"nvim.bufferline.error_selected":                    true,
	}
	actual := make(map[string]bool)
	for _, pair := range result.Pairs {
		if pair.Foreground == "zsh.error" {
			actual[pair.ConsumerID] = true
		}
	}
	if len(actual) != len(expected) {
		t.Fatalf("zsh.error pair count = %d, want %d: %v", len(actual), len(expected), actual)
	}
	for consumerID := range expected {
		if !actual[consumerID] {
			t.Errorf("zsh.error diagnostic pair %s not found", consumerID)
		}
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
		"variable.language":              "foregrounds.heading",
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

func TestKeybindsModeIndicatorUsesMutedPurple(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "wezterm", "keybinds.lua"))
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)

	// wezterm/keybinds.lua は sourceinventory の抽出対象外のため、参照文字列を直接固定する。
	if !strings.Contains(contents, "colors.purples.muted_purple") {
		t.Error("key-table mode indicator does not reference colors.purples.muted_purple")
	}
	if strings.Contains(contents, "colors.semantic.keyword") {
		t.Error("key-table mode indicator still references colors.semantic.keyword")
	}
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
