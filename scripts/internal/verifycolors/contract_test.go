package verifycolors

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/xterm"
)

func TestDefaultContractCoversEveryPaletteToken(t *testing.T) {
	palettePath := filepath.Join("..", "..", "..", "colors", "ghost-visor.toml")
	colorPalette, err := LoadPalette(palettePath)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := DefaultContract()
	if err != nil {
		t.Fatal(err)
	}
	summary, err := ValidateContract(contract, colorPalette)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Unclassified) != 0 {
		t.Fatalf("unclassified tokens = %v", summary.Unclassified)
	}
}

func TestWezTermCustomANSIResolution(t *testing.T) {
	palettePath := filepath.Join("..", "..", "..", "colors", "ghost-visor.toml")
	colorPalette, err := LoadPalette(palettePath)
	if err != nil {
		t.Fatal(err)
	}
	ansi, err := colorPalette.WezTermANSI()
	if err != nil {
		t.Fatal(err)
	}
	resolvedBlack, err := xterm.Resolve256(0, ansi)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedBlack.Hex() != colorPalette.ANSI["black"] {
		t.Errorf("index 0 = %s, want custom ANSI %s", resolvedBlack.Hex(), colorPalette.ANSI["black"])
	}
	resolvedWhite, err := xterm.Resolve256(7, ansi)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedWhite.Hex() != colorPalette.ANSI["white"] {
		t.Errorf("index 7 = %s, want custom ANSI %s", resolvedWhite.Hex(), colorPalette.ANSI["white"])
	}
}

func TestCurrentPaletteHasNoEnforcedContrastFailures(t *testing.T) {
	palettePath := filepath.Join("..", "..", "..", "colors", "ghost-visor.toml")
	colorPalette, err := LoadPalette(palettePath)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := DefaultContract()
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := Evaluate(contract, colorPalette, nil)
	if err != nil {
		t.Fatal(err)
	}
	var failures []string
	for _, result := range evaluation.Results {
		if result.Class == ClassEnforced && !result.Pass {
			failures = append(failures, fmt.Sprintf(
				"%s[%s] %s/%s = %.2f",
				result.ConsumerID,
				result.Profile,
				result.ForegroundToken,
				result.BackgroundToken,
				result.Ratio,
			))
		}
	}
	if len(failures) > 0 {
		t.Fatalf("enforced contrast failures (%d):\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

func TestMatchParenProfileRatios(t *testing.T) {
	palettePath := filepath.Join("..", "..", "..", "colors", "ghost-visor.toml")
	colorPalette, err := LoadPalette(palettePath)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := DefaultContract()
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := Evaluate(contract, colorPalette, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, result := range evaluation.Results {
		if result.ConsumerID == "nvim.highlight.MatchParen" || result.ConsumerID == "vim.highlight.MatchParen" {
			t.Logf(
				"%s[%s] fg=%s bg=%s ratio=%.4f pass=%t",
				result.ConsumerID,
				result.Profile,
				result.ForegroundColor,
				result.BackgroundColor,
				result.Ratio,
				result.Pass,
			)
			if result.ConsumerID == "vim.highlight.MatchParen" && result.Profile == ProfileCtermFGBG {
				if result.DeclaredClass != ClassEnforced || result.Class != ClassReportOnly {
					t.Errorf(
						"cterm MatchParen classes = declared:%s effective:%s, want enforced/report-only",
						result.DeclaredClass,
						result.Class,
					)
				}
				if result.ForegroundIndex == nil || *result.ForegroundIndex != 15 {
					t.Errorf("cterm MatchParen foreground index = %v, want 15", result.ForegroundIndex)
				}
				if result.BackgroundIndex == nil || *result.BackgroundIndex != 67 {
					t.Errorf("cterm MatchParen background index = %v, want 67", result.BackgroundIndex)
				}
				if result.ForegroundColor != "#F8FCFD" || result.BackgroundColor != "#5F87AF" {
					t.Errorf(
						"cterm MatchParen colors = %s/%s, want #F8FCFD/#5F87AF",
						result.ForegroundColor,
						result.BackgroundColor,
					)
				}
				if math.Abs(result.Ratio-3.6498) > 0.0001 {
					t.Errorf("cterm MatchParen ratio = %.4f, want 3.6498", result.Ratio)
				}
			}
			found++
		}
	}
	if found != 3 {
		t.Fatalf("MatchParen profile results = %d, want 3", found)
	}
}
