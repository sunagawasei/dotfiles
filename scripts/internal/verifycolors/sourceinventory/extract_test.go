package sourceinventory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	expected := map[string]verifycolors.TokenRef{
		"plus-style":                    "nvim.diff_add_bg",
		"minus-style":                   "nvim.diff_delete_bg",
		"plus-emph-style":               "nvim.diff_add_inline_bg",
		"minus-emph-style":              "nvim.diff_delete_inline_bg",
		"line-numbers-plus-style":       "semantic.success",
		"line-numbers-minus-style":      "semantic.error",
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
	if count != len(expected) {
		t.Fatalf("delta style pair count = %d, want %d", count, len(expected))
	}
	for option, token := range expected {
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
