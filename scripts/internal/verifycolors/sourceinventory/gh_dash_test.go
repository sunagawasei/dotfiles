package sourceinventory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

func TestGhDashSpecExtractionHasEighteenSlots(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	specs, err := parseGhDashColorSpecs(root)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, entries := range specs {
		total += len(entries)
	}
	if total != 18 {
		t.Fatalf("gh-dash spec slot count = %d, want 18", total)
	}
	wantGroupSizes := map[string]int{"text": 8, "background": 1, "border": 3, "icon": 6}
	for group, want := range wantGroupSizes {
		if got := len(specs[group]); got != want {
			t.Errorf("gh-dash spec group %q size = %d, want %d", group, got, want)
		}
	}
}

// ghDashExpectedPair is an independently-authored expectation for one gh-dash.theme.* pair,
// checked against the real extraction in TestGhDashPairTableMatchesExpectedTable. It uses
// resolved token values (e.g. "foregrounds.main", not the slot name "text.primary") so this
// test also catches a ghDashColorSpecs token drift, not just a gh_dash.go logic bug.
type ghDashExpectedPair struct {
	foreground verifycolors.TokenRef
	background verifycolors.TokenRef // ignored when ambient is true
	ambient    bool
	class      verifycolors.PairClass
	role       verifycolors.PairRole
}

// ghDashExpectedPairs is the full expected gh-dash.theme.* pair table (keys are the
// consumer ID suffix after "gh-dash.theme."). It mirrors ghDashPairTable's 36 rows plus the
// background.selected surface pair (37 total; gh-dash.terminalDefault.taskStart is a 38th
// pair checked separately in TestGhDashTerminalDefaultTaskStartPair, since its consumer ID
// doesn't use the "gh-dash.theme." prefix this table is keyed on). Classes here are
// pre-override: 3 of these (text.inverted.selected, text.secondary.faintBorder,
// text.faint.faintBorder) are Enforced here but get demoted to report-only by
// inventory_overrides.go; see TestGhDashOverridesAreReportOnlyWithReason for their
// post-override state.
var ghDashExpectedPairs = map[string]ghDashExpectedPair{
	// background = background.selected
	"text.primary.selected":         {foreground: "foregrounds.main", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.secondary.selected":       {foreground: "foregrounds.dim", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.inverted.selected":        {foreground: "core.darkest_bg", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.faint.selected":           {foreground: "foregrounds.subdued", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.warning.selected":         {foreground: "ansi.bright_yellow", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.success.selected":         {foreground: "semantic.success", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.error.selected":           {foreground: "ansi.bright_red", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"icon.newcontributor.selected":  {foreground: "semantic.success", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.contributor.selected":     {foreground: "teals.mid_bright", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.collaborator.selected":    {foreground: "purples.lavender", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.member.selected":          {foreground: "purples.bright_purple", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.owner.selected":           {foreground: "ansi.bright_yellow", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"border.faint.selected":         {foreground: "blues_slates.slate_mid", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleBorder},
	"background.selected.helpLabel": {foreground: "core.active_line", background: "foregrounds.subdued", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},

	// background = border.faint
	"text.primary.faintBorder":   {foreground: "foregrounds.main", background: "blues_slates.slate_mid", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.secondary.faintBorder": {foreground: "foregrounds.dim", background: "blues_slates.slate_mid", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.faint.faintBorder":     {foreground: "foregrounds.subdued", background: "blues_slates.slate_mid", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"text.success.faintBorder":   {foreground: "semantic.success", background: "blues_slates.slate_mid", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"border.primary.faintBorder": {foreground: "teals.bright", background: "blues_slates.slate_mid", class: verifycolors.ClassEnforced, role: verifycolors.RoleBorder},

	// background = text.faint
	"text.inverted.draftPill": {foreground: "core.darkest_bg", background: "foregrounds.subdued", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},

	// background = background.selected, foreground = raw ANSI SGR (resolved through our own
	// WezTerm ANSI palette, not a gh-dash config token)
	"ansi.brightWhite.selected": {foreground: "ansi.bright_white", background: "core.active_line", class: verifycolors.ClassEnforced, role: verifycolors.RoleText},
	"ansi.green.selected":       {foreground: "ansi.green", background: "core.active_line", class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},

	// ambient
	"text.primary":        {foreground: "foregrounds.main", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"text.secondary":      {foreground: "foregrounds.dim", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"text.faint":          {foreground: "foregrounds.subdued", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"text.warning":        {foreground: "ansi.bright_yellow", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"text.success":        {foreground: "semantic.success", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"text.error":          {foreground: "ansi.bright_red", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleText},
	"border.primary":      {foreground: "teals.bright", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleBorder},
	"border.secondary":    {foreground: "teals.border", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleBorder},
	"border.faint":        {foreground: "blues_slates.slate_mid", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleBorder},
	"icon.newcontributor": {foreground: "semantic.success", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.contributor":    {foreground: "teals.mid_bright", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.collaborator":   {foreground: "purples.lavender", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.member":         {foreground: "purples.bright_purple", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},
	"icon.owner":          {foreground: "ansi.bright_yellow", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleIndicator},

	// surface
	"background.selected": {foreground: "core.active_line", ambient: true, class: verifycolors.ClassReportOnly, role: verifycolors.RoleSurface},
}

// TestGhDashPairTableMatchesExpectedTable checks every gh-dash.theme.* pair (foreground,
// background, class, role) against ghDashExpectedPairs, and that no gh-dash.theme.* pair
// exists outside that table (a stray/renamed row would show up as "unexpected").
func TestGhDashPairTableMatchesExpectedTable(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)

	for suffix, want := range ghDashExpectedPairs {
		consumerID := "gh-dash.theme." + suffix
		pair, ok := pairs[consumerID]
		if !ok {
			t.Errorf("pair %s not found", consumerID)
			continue
		}
		if pair.Foreground != want.foreground {
			t.Errorf("%s foreground = %q, want %q", consumerID, pair.Foreground, want.foreground)
		}
		if pair.Background.Ambient != want.ambient {
			t.Errorf("%s background.Ambient = %t, want %t", consumerID, pair.Background.Ambient, want.ambient)
		}
		if !want.ambient && pair.Background.Token != want.background {
			t.Errorf("%s background = %q, want %q", consumerID, pair.Background.Token, want.background)
		}
		if pair.Class != want.class {
			t.Errorf("%s class = %q, want %q", consumerID, pair.Class, want.class)
		}
		if pair.Role != want.role {
			t.Errorf("%s role = %q, want %q", consumerID, pair.Role, want.role)
		}
	}

	// gh-dash.terminalDefault.taskStart doesn't use the "gh-dash.theme." prefix (it isn't a
	// gh-dash config slot at all); it's allow-listed here and checked on its own in
	// TestGhDashTerminalDefaultTaskStartPair.
	knownNonThemePrefixed := map[string]bool{"gh-dash.terminalDefault.taskStart": true}

	var extra []string
	for _, pair := range result.Pairs {
		suffix := strings.TrimPrefix(pair.ConsumerID, "gh-dash.theme.")
		if suffix != pair.ConsumerID {
			if _, ok := ghDashExpectedPairs[suffix]; !ok {
				extra = append(extra, pair.ConsumerID)
			}
			continue
		}
		if strings.HasPrefix(pair.ConsumerID, "gh-dash.") && !knownNonThemePrefixed[pair.ConsumerID] {
			extra = append(extra, pair.ConsumerID)
		}
	}
	if len(extra) > 0 {
		t.Errorf("gh-dash.* pairs not accounted for: %v", extra)
	}
}

// TestGhDashTerminalDefaultTaskStartPair checks gh-dash.terminalDefault.taskStart: the
// running-task status line sets a background but never calls Foreground, so it renders in
// the terminal's own default foreground (wezterm/wezterm.lua:203 = foregrounds.main here),
// not any gh-dash config token.
func TestGhDashTerminalDefaultTaskStartPair(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)
	assertPair(t, pairs, "gh-dash.terminalDefault.taskStart", "foregrounds.main", "core.active_line", false, verifycolors.ClassEnforced)
	if pair := pairs["gh-dash.terminalDefault.taskStart"]; pair.Role != verifycolors.RoleText {
		t.Errorf("gh-dash.terminalDefault.taskStart role = %q, want %q", pair.Role, verifycolors.RoleText)
	}
}

// TestGhDashRemovedPairsAreNotRegistered checks the two use-sites this task removed: neither
// key has a renderer that reads it (see the text.actor.unused and icon.unknownrole.upstreamUnwired
// coverage notes), so no pair - ambient or selected-row - should exist for either.
func TestGhDashRemovedPairsAreNotRegistered(t *testing.T) {
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
		"gh-dash.theme.text.actor",
		"gh-dash.theme.text.actor.selected",
		"gh-dash.theme.icon.unknownrole",
		"gh-dash.theme.icon.unknownrole.selected",
	} {
		if _, ok := pairs[consumerID]; ok {
			t.Errorf("%s should not be registered", consumerID)
		}
	}
}

// TestGhDashIconMemberSelectedIsReportOnly is a focused regression check (in addition to the
// full-table check above) for the one icon.*.selected value codex's numeric sweep flagged as
// closest to the AA line if it were ever mistakenly enforced (member: 4.097:1).
func TestGhDashIconMemberSelectedIsReportOnly(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	pairs := pairMap(result.Pairs)
	pair, ok := pairs["gh-dash.theme.icon.member.selected"]
	if !ok {
		t.Fatal("gh-dash.theme.icon.member.selected not found")
	}
	if pair.Class != verifycolors.ClassReportOnly {
		t.Errorf("gh-dash.theme.icon.member.selected class = %q, want %q", pair.Class, verifycolors.ClassReportOnly)
	}
}

// TestGhDashCoverageNotes checks the 6 gh-dash coverage notes: the two pill/label notes that
// explain why parts of text.inverted can't be static pairs, the two "no pair" notes for the
// keys TestGhDashRemovedPairsAreNotRegistered checks, the meta-note documenting how the
// use-site survey was done, and the note tracking the reverse literal-foreground direction.
func TestGhDashCoverageNotes(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	notes := make(map[string]verifycolors.CoverageNote, len(result.CoverageNotes))
	for _, note := range result.CoverageNotes {
		notes[note.ID] = note
	}

	const rev = "rev 78b9ca5e21dcee4740502018d46dc5e4b613db86"
	wantSourceSubstrings := map[string][]string{
		"gh-dash.theme.text.inverted.statusPill": {
			"internal/tui/context/styles.go:120-122",
			"internal/tui/components/prview/prview.go:280-298",
			"internal/tui/components/issueview/issueview.go:311-326",
			"internal/tui/theme/theme.go:107-110",
			"gh-dash v4.23.2", rev,
		},
		"gh-dash.theme.text.inverted.labels": {
			"internal/tui/context/styles.go:120-122",
			"internal/tui/common/labels.go:9-22",
			"internal/tui/theme/theme.go:107-110",
			"gh-dash v4.23.2", rev,
		},
		"gh-dash.theme.icon.unknownrole.upstreamUnwired": {
			"internal/config/parser.go:253",
			"internal/tui/theme/theme.go:28",
			"internal/tui/theme/theme.go:147",
			"gh-dash v4.23.2", rev,
		},
		"gh-dash.theme.text.actor.unused": {
			"internal/tui/theme/theme.go:22",
			"internal/tui/theme/theme.go:123-126",
			"gh-dash v4.23.2", rev,
		},
		"gh-dash.inventory.survey": {
			"gh-dash v4.23.2", rev, "ghDashPairTable",
		},
		"gh-dash.theme.literalForegrounds": {
			"internal/tui/context/styles.go:96-116",
			"internal/tui/components/prrow/prrow.go:76-85",
			"internal/tui/components/branch/branch.go:58-64",
			"internal/tui/components/notificationrow/notificationrow.go:47-64",
			"internal/tui/components/issuerow/issuerow.go:87-92",
			"internal/tui/components/notificationrow/notificationrow.go:75-78",
			"internal/tui/components/notificationrow/notificationrow.go:78-81",
			"internal/tui/components/notificationrow/notificationrow.go:113-125",
			"internal/tui/components/footer/footer.go:138-152",
			"internal/tui/context/styles.go:227-230",
			"internal/tui/components/reposection/reposection.go:555-565",
			"internal/tui/context/styles.go:186-190",
			"internal/tui/theme/theme.go:79-82",
			"internal/tui/theme/theme.go:87-90",
			"#e0af68",
			"gh-dash v4.23.2", rev,
		},
	}
	for noteID, wantSubstrings := range wantSourceSubstrings {
		note, ok := notes[noteID]
		if !ok {
			t.Errorf("coverage note %s not found", noteID)
			continue
		}
		if note.Reason == "" {
			t.Errorf("coverage note %s has no reason", noteID)
		}
		for _, want := range wantSubstrings {
			if !strings.Contains(note.Source, want) {
				t.Errorf("coverage note %s source = %q, want it to contain %q", noteID, note.Source, want)
			}
		}
	}

	ghDashNoteCount := 0
	for id := range notes {
		if strings.HasPrefix(id, "gh-dash.") {
			ghDashNoteCount++
		}
	}
	if ghDashNoteCount != len(wantSourceSubstrings) {
		t.Errorf("gh-dash coverage note count = %d, want %d", ghDashNoteCount, len(wantSourceSubstrings))
	}

	// The survey note's Reason must describe all 3 scan passes (a 3rd, background-only pass
	// was added after 2 rounds of review found a misses from only running the first 2), and
	// literalForegrounds' Reason must record the whole-file raw-ANSI grep that ruled out a
	// third SGR code beyond the 2 gh-dash.theme.ansi.*.selected pairs.
	if survey := notes["gh-dash.inventory.survey"]; !strings.Contains(survey.Reason, "Pass 3") {
		t.Errorf("gh-dash.inventory.survey reason = %q, want it to mention \"Pass 3\"", survey.Reason)
	}
	if literal := notes["gh-dash.theme.literalForegrounds"]; !strings.Contains(literal.Reason, "grep") {
		t.Errorf("gh-dash.theme.literalForegrounds reason = %q, want it to mention the whole-file grep confirming only 2 raw ANSI escapes", literal.Reason)
	}
}

// TestGhDashOverridesAreReportOnlyWithReason checks the 3 manual overrides in
// inventory_overrides.go: each demotes an Enforced-by-default pair to report-only, and per
// .claude/skills/color-validation/SKILL.md ("手書きoverrideの出所と再検証") each Source must
// carry the consumer file:line, the consumer's version, its rev, and (where applicable) the
// config→Theme wiring that makes the token take effect at all.
func TestGhDashOverridesAreReportOnlyWithReason(t *testing.T) {
	pairs, err := verifycolors.ContractPairs()
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]verifycolors.PairSpec, len(pairs))
	for _, pair := range pairs {
		byID[pair.ConsumerID] = pair
	}

	wantSourceSubstrings := map[string][]string{
		"gh-dash.theme.text.inverted.selected": {
			"gh-dash v4.23.2",
			"rev 78b9ca5e21dcee4740502018d46dc5e4b613db86",
			"internal/tui/context/styles.go:222-225",
			"internal/tui/components/footer/footer.go:159-190",
			"internal/tui/context/styles.go:120-122",
			"internal/tui/theme/theme.go:107-110",
		},
		"gh-dash.theme.text.faint.faintBorder": {
			"gh-dash v4.23.2",
			"rev 78b9ca5e21dcee4740502018d46dc5e4b613db86",
			"internal/tui/context/styles.go:234-236",
			"internal/tui/components/footer/footer.go:155-157",
			"internal/tui/theme/theme.go:95-98",
			"internal/tui/theme/theme.go:87-90",
		},
		"gh-dash.theme.text.secondary.faintBorder": {
			"gh-dash v4.23.2",
			"rev 78b9ca5e21dcee4740502018d46dc5e4b613db86",
			"internal/tui/components/inputbox/inputbox.go:49-56",
			"internal/tui/theme/theme.go:103-106",
			"internal/tui/theme/theme.go:87-90",
		},
	}
	for consumerID, wantSubstrings := range wantSourceSubstrings {
		pair, ok := byID[consumerID]
		if !ok {
			t.Fatalf("pair %s not found", consumerID)
		}
		if pair.Class != verifycolors.ClassReportOnly {
			t.Errorf("%s class = %q, want %q", consumerID, pair.Class, verifycolors.ClassReportOnly)
		}
		if pair.Reason == "" {
			t.Errorf("%s reason is empty, want non-empty", consumerID)
		}
		for _, want := range wantSubstrings {
			if !strings.Contains(pair.Source, want) {
				t.Errorf("%s source = %q, want it to contain %q", consumerID, pair.Source, want)
			}
		}
	}
}

// TestGhDashFaintBorderOverridesRequireShowSeparator pins the precondition both border.faint
// text overrides depend on: the row separator (not these text-on-fill pairs) is the priority
// use of border.faint. If theme.ui.table.showSeparator ever flips to false, that justification
// no longer holds and both overrides need re-verification (see their Reason/Source in
// inventory_overrides.go).
func TestGhDashFaintBorderOverridesRequireShowSeparator(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "gh-dash", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\n            showSeparator: true\n") {
		t.Fatal("gh-dash/config.yml theme.ui.table.showSeparator is not \"true\"; " +
			"re-verify the gh-dash.theme.text.faint.faintBorder and gh-dash.theme.text.secondary.faintBorder " +
			"overrides in inventory_overrides.go")
	}
}

func TestGhDashIconColorsAreDistinct(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "gh-dash", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}

	colors := make(map[string]string)
	inIcon := false
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.TrimRight(line, "\n") == "        icon:":
			inIcon = true
		case inIcon && line != "" && !strings.HasPrefix(line, "            "):
			inIcon = false
		case inIcon && strings.HasPrefix(line, "            "):
			trimmed := strings.TrimSpace(line)
			idx := strings.Index(trimmed, ": ")
			if idx < 0 {
				continue // a "# ..." comment line, e.g. the unknownrole unwired note
			}
			key := trimmed[:idx]
			value := strings.Trim(strings.TrimSpace(trimmed[idx+2:]), `"`)
			colors[key] = value
		}
	}

	wantKeys := []string{"newcontributor", "contributor", "collaborator", "member", "owner", "unknownrole"}
	if len(colors) != len(wantKeys) {
		t.Fatalf("gh-dash config.yml icon color count = %d, want %d (found %v)", len(colors), len(wantKeys), colors)
	}

	seen := make(map[string]string, len(wantKeys))
	for _, key := range wantKeys {
		value, ok := colors[key]
		if !ok || value == "" {
			t.Fatalf("gh-dash config.yml icon.%s has no generated color", key)
		}
		if previousKey, ok := seen[value]; ok {
			t.Fatalf("gh-dash config.yml icon.%s and icon.%s both resolve to %s", previousKey, key, value)
		}
		seen[value] = key
	}
}

// TestGhDashGeneratedConfigHasExplanatoryComments checks (read-only, from the generated
// output) that the two YAML comment lines generate-colors/main.go now emits survive into
// gh-dash/config.yml next to their keys. The walker's own structural tests (comment lines
// don't break the 18-key exact-match check) live in cmd/generate-colors, out of this
// package's scope; go test ./cmd/generate-colors/... covers those.
func TestGhDashGeneratedConfigHasExplanatoryComments(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "gh-dash", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"gh-dash v4.23.2 は ParseTheme で未配線。設定値は反映されない",
		"v4.23.2 では描画箇所なし",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("gh-dash/config.yml does not contain comment %q", want)
		}
	}
}
