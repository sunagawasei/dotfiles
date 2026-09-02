package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

func TestTableCountsAndGoldenSpotChecks(t *testing.T) {
	values := loadCurrentValues(t)
	if len(ansiChecks) != 360 || len(namedChecks) != 99 || len(waivers) != 3 {
		t.Fatalf("table counts = %d/%d/%d, want 360/99/3", len(ansiChecks), len(namedChecks), len(waivers))
	}
	if err := validateTables(values); err != nil {
		t.Fatal(err)
	}
	spots := []struct {
		first, second, vision string
		baseline, proposed    float64
	}{
		{"ansi.bright_blue", "ansi.white", "protanopia", 21.459074, 18.794076},
		{"ansi.bright_blue", "ansi.white", "deuteranopia", 20.222503, 17.419557},
		{"ansi.bright_green", "ansi.bright_white", "protanopia", 23.167243, 17.655045},
		{"ansi.bright_cyan", "ansi.bright_green", "tritanopia", 17.569726, 8.913793},
		{"ansi.magenta", "ansi.red", "tritanopia", 13.051381, 6.073791},
		{"ansi.bright_green", "ansi.white", "deuteranopia", 19.213663, 14.082855},
		{"ansi.black", "ansi.blue", "protanopia", 57.854772, 60.315466},
		{"ansi.red", "ansi.green", "deuteranopia", 8.598644, 9.419540},
		{"ansi.cyan", "ansi.magenta", "tritanopia", 92.080611, 83.111447},
		{"ansi.bright_yellow", "ansi.bright_white", "tritanopia", 30.921512, 30.921512},
		{"ansi.blue", "ansi.magenta", "deuteranopia", 22.729617, 21.755118},
		{"ansi.green", "ansi.white", "protanopia", 27.826923, 27.826923},
		{"semantic.error", "semantic.success", "protanopia", 23.463923, 22.220447},
		{"git.changed", "git.deleted", "deuteranopia", 11.043981, 13.753410},
		{"semantic.keyword", "semantic.string", "deuteranopia", 22.440980, 22.963556},
		{"wezterm.active_tab", "wezterm.inactive_tab", "protanopia", 20.423750, 21.739886},
		{"semantic.variable", "semantic.constant", "tritanopia", 10.705617, 10.705617},
	}
	for _, spot := range spots {
		t.Run(spot.first+"/"+spot.second+"/"+spot.vision, func(t *testing.T) {
			item, ok := findCheck(append(ansiChecks, namedChecks...), spot.first, spot.second, spot.vision)
			if !ok {
				t.Fatal("spot-check is missing")
			}
			if math.Abs(item.BaselineDE-math.Floor(spot.baseline*10000)/10000) > 0.00001 {
				t.Errorf("baseline = %.4f, want %.4f", item.BaselineDE, math.Floor(spot.baseline*10000)/10000)
			}
			got, err := deltaFor(values, item)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-spot.proposed) > 0.000001 {
				t.Errorf("proposed dE = %.9f, want %.6f", got, spot.proposed)
			}
		})
	}
}

func TestEvaluateCriterionABoundary(t *testing.T) {
	values := loadCurrentValues(t)
	for _, name := range ansiNames {
		values[verifycolors.TokenRef(name)] = values[verifycolors.TokenRef("ansi.black")]
	}
	if err := evaluate(values); err == nil {
		t.Fatal("criterion A count violation accepted")
	}
}

func TestEvaluateUnwaivedViolationFails(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = []check{{First: "ansi.black", Second: "ansi.red", Vision: "protanopia", BaselineDE: 25}}
	namedChecks = nil
	values["ansi.red"] = values["ansi.black"]
	if err := evaluate(values); err == nil {
		t.Fatal("unwaived criterion B violation accepted")
	}
}

func TestEvaluateWaiverBoundaryAndViolation(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = []check{{First: "ansi.bright_blue", Second: "ansi.white", Vision: "protanopia", BaselineDE: 21.4590}}
	namedChecks = nil
	values["ansi.white"] = values["ansi.bright_blue"]
	original := waivers
	defer func() { waivers = original }()
	waivers = []waiver{{First: "ansi.bright_blue", Second: "ansi.white", Vision: "protanopia", ApprovedDE: 0.1}}
	if err := evaluate(values); err == nil {
		t.Fatal("waiver below approved threshold accepted")
	}
}

func TestEvaluateStaleWaiverFails(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = []check{{First: "ansi.bright_blue", Second: "ansi.white", Vision: "protanopia", BaselineDE: 21.4590}}
	namedChecks = nil
	values["ansi.white"] = values["ansi.black"]
	original := waivers
	defer func() { waivers = original }()
	waivers = []waiver{{First: "ansi.bright_blue", Second: "ansi.white", Vision: "protanopia", ApprovedDE: 0}}
	if err := evaluate(values); err == nil {
		t.Fatal("stale waiver accepted")
	}
}

func TestEvaluateCriterionCViolationFails(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = nil
	namedChecks = []check{{First: "semantic.error", Second: "semantic.success", Vision: "protanopia", BaselineDE: 20, Kind: KindHistoricalFloor}}
	values["semantic.success"] = values["semantic.error"]
	if err := evaluate(values); err == nil {
		t.Fatal("criterion C violation accepted")
	}
}

func TestEvaluateCurrentPaletteSucceeds(t *testing.T) {
	values := loadCurrentValues(t)
	if err := evaluate(values); err != nil {
		t.Fatalf("current palette rejected by evaluate: %v", err)
	}
}

func TestEvaluateHistoricalBelowTwentyCollapseFails(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = nil
	namedChecks = []check{
		{First: "semantic.error", Second: "semantic.warning", Vision: "protanopia", BaselineDE: 10.0475, Kind: KindHistoricalFloor},
		{First: "semantic.error", Second: "semantic.warning", Vision: "deuteranopia", BaselineDE: 5.9503, Kind: KindHistoricalFloor},
		{First: "semantic.error", Second: "semantic.warning", Vision: "tritanopia", BaselineDE: 15.7606, Kind: KindHistoricalFloor},
	}
	values[verifycolors.TokenRef("semantic.warning")] = values[verifycolors.TokenRef("semantic.error")]
	if err := evaluate(values); err == nil {
		t.Fatal("collapsed below-twenty historical pair accepted")
	}
}

func TestEvaluateMinimumSeparationBelowThreeFails(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = nil
	namedChecks = []check{{First: "foregrounds.heading", Second: "semantic.info", Vision: "protanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation}}
	values[verifycolors.TokenRef("semantic.info")] = values[verifycolors.TokenRef("foregrounds.heading")]
	if err := evaluate(values); err == nil {
		t.Fatal("collapsed minimum-separation pair accepted")
	}
}

func TestEvaluateEffectiveNamedLimitCap(t *testing.T) {
	values := loadCurrentValues(t)
	originalANSI, originalNamed := ansiChecks, namedChecks
	defer func() { ansiChecks, namedChecks = originalANSI, originalNamed }()
	ansiChecks = nil
	namedChecks = []check{{First: "semantic.error", Second: "semantic.success", Vision: "protanopia", BaselineDE: 23.4639, Kind: KindHistoricalFloor}}
	values[verifycolors.TokenRef("semantic.error")] = "#000000"
	values[verifycolors.TokenRef("semantic.success")] = findColorWithDelta(t, values, namedChecks[0], 20.0, 20.1)
	liveAtTwenty, err := deltaFor(values, namedChecks[0])
	if err != nil {
		t.Fatal(err)
	}
	if liveAtTwenty < 20.0 || liveAtTwenty >= 20.1 {
		t.Fatalf("live dE = %.6f, want approximately 20.0", liveAtTwenty)
	}
	if err := evaluate(values); err != nil {
		t.Fatalf("raw historical floor above 20 was not capped at 20: %v", err)
	}
	if namedChecks[0].BaselineDE != 23.4639 {
		t.Fatalf("validate record was changed, got %.4f", namedChecks[0].BaselineDE)
	}

	values[verifycolors.TokenRef("semantic.success")] = findColorWithDelta(t, values, namedChecks[0], 19.8, 19.9)
	liveBelowTwenty, err := deltaFor(values, namedChecks[0])
	if err != nil {
		t.Fatal(err)
	}
	if liveBelowTwenty < 19.8 || liveBelowTwenty >= 19.9 {
		t.Fatalf("live dE = %.6f, want approximately 19.9", liveBelowTwenty)
	}
	if err := evaluate(values); err == nil {
		t.Fatal("live dE below 20 incorrectly accepted for raw historical floor above 20")
	}
}

func TestTableIntegrityRejectsDuplicateTriple(t *testing.T) {
	values := loadCurrentValues(t)
	original := ansiChecks
	defer func() { ansiChecks = original }()
	ansiChecks = append([]check(nil), original...)
	ansiChecks[1] = ansiChecks[0]
	if err := validateTables(values); err == nil {
		t.Fatal("duplicate ANSI triple accepted")
	}
}

func TestTableIntegrityRejectsCountMismatch(t *testing.T) {
	values := loadCurrentValues(t)
	original := namedChecks
	defer func() { namedChecks = original }()
	namedChecks = append([]check(nil), original[:len(original)-1]...)
	if err := validateTables(values); err == nil {
		t.Fatal("namedChecks count mismatch accepted")
	}
}

func TestTableIntegrityRejectsMissingVision(t *testing.T) {
	values := loadCurrentValues(t)
	original := ansiChecks
	defer func() { ansiChecks = original }()
	ansiChecks = append([]check(nil), original...)
	ansiChecks[0].Vision = "invalid"
	if err := validateTables(values); err == nil {
		t.Fatal("invalid vision accepted")
	}
}

func TestTableIntegrityRejectsMissingToken(t *testing.T) {
	values := loadCurrentValues(t)
	delete(values, verifycolors.TokenRef("ansi.black"))
	if err := validateTables(values); err == nil {
		t.Fatal("missing palette token accepted")
	}
}

func TestLegacyFixtureSHA256RejectsOneByteChange(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	fixture[len(fixture)-1] ^= 1
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("one-byte legacy fixture change accepted")
	}
}

func TestHistoricalFloorsMatchLegacyRawDE(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	original := namedChecks
	defer func() { namedChecks = original }()
	namedChecks = append([]check(nil), original...)
	namedChecks[0].BaselineDE += 0.0001
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("historical floor mismatch accepted")
	}
}

func TestAchievedFreezeFloorsAreExact(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	original := namedChecks
	defer func() { namedChecks = original }()
	namedChecks = append([]check(nil), original...)
	namedChecks[89].BaselineDE = 18.7516
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("achieved freeze floor mismatch accepted")
	}
}

func TestMinimumSeparationFloorsAreExact(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	original := namedChecks
	defer func() { namedChecks = original }()
	namedChecks = append([]check(nil), original...)
	namedChecks[90].BaselineDE = 2.9999
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("minimum separation floor mismatch accepted")
	}
}

func TestNamedKindAllowlistRejectsReplacementPair(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	original := namedChecks
	defer func() { namedChecks = original }()
	namedChecks = append([]check(nil), original...)
	namedChecks[0].First = "semantic.error"
	namedChecks[0].Second = "git.changed"
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("named kind allowlist replacement accepted")
	}
}

func TestMinimumSeparationUsesLegacyTokenValues(t *testing.T) {
	values, fixture, legacyValues := loadFixtureInputs(t)
	legacyValues = cloneValues(legacyValues)
	legacyValues["semantic.info"] = legacyValues["semantic.info"]
	legacyValues["semantic.info"] = "#000000"
	if err := validateTablesWithFixture(values, fixture, legacyValues); err == nil {
		t.Fatal("minimum separation legacy token mismatch accepted")
	}
}

func TestFloor4AndEffectiveNamedLimitBoundaries(t *testing.T) {
	if got := floor4(1.2345); got != 1.2345 {
		t.Fatalf("floor4 exact boundary = %.4f, want 1.2345", got)
	}
	if got := floor4(1.23449); got != 1.2344 {
		t.Fatalf("floor4 below boundary = %.4f, want 1.2344", got)
	}
	if got := floor4(1.23451); got != 1.2345 {
		t.Fatalf("floor4 above boundary = %.4f, want 1.2345", got)
	}
	if got := effectiveNamedLimit(20.0001); got != 20 {
		t.Fatalf("effective limit above threshold = %.4f, want 20", got)
	}
	if got := effectiveNamedLimit(19.9999); got != 19.9999 {
		t.Fatalf("effective limit below threshold = %.4f, want 19.9999", got)
	}
}

func loadFixtureInputs(t *testing.T) (map[verifycolors.TokenRef]string, []byte, map[verifycolors.TokenRef]string) {
	t.Helper()
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	current, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	fixturePath := filepath.Join(root, "scripts", "testdata", "legacy-palette-8bf163c.toml")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := verifycolors.LoadPalette(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	return current.TokenValues(), fixture, legacy.TokenValues()
}

func cloneValues(values map[verifycolors.TokenRef]string) map[verifycolors.TokenRef]string {
	clone := make(map[verifycolors.TokenRef]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func findColorWithDelta(t *testing.T, values map[verifycolors.TokenRef]string, item check, lower, upper float64) string {
	t.Helper()
	candidateValues := cloneValues(values)
	for red := 0; red <= 255; red++ {
		for green := 0; green <= 255; green++ {
			for blue := 0; blue <= 255; blue++ {
				candidate := fmt.Sprintf("#%02x%02x%02x", red, green, blue)
				candidateValues[verifycolors.TokenRef(item.Second)] = candidate
				got, err := deltaFor(candidateValues, item)
				if err != nil {
					t.Fatal(err)
				}
				if got >= lower && got < upper {
					return candidate
				}
			}
		}
	}
	t.Fatalf("no color has dE in [%.4f, %.4f)", lower, upper)
	return ""
}

func findCheck(items []check, first, second, vision string) (check, bool) {
	key := triple(first, second, vision)
	for _, item := range items {
		if triple(item.First, item.Second, item.Vision) == key {
			return item, true
		}
	}
	return check{}, false
}

func loadCurrentValues(t *testing.T) map[verifycolors.TokenRef]string {
	t.Helper()
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	colorPalette, err := verifycolors.LoadPalette(filepath.Join(root, "colors", "ghost-visor.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return colorPalette.TokenValues()
}
