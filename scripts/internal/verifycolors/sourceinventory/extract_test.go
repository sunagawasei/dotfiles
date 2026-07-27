package sourceinventory

import (
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
