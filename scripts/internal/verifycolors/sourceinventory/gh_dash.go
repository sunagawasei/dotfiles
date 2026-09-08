package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

// ghDashSpecEntry is one row of generate-colors/main.go's ghDashColorSpecs: a
// theme.colors.<group>.<key> path and the palette token that fills it, plus where it was
// found so pairs built from it carry an accurate Source.
type ghDashSpecEntry struct {
	token  verifycolors.TokenRef
	source string
}

// ghDashExpectedKeysByGroup pins the exact key set ghDashColorSpecs (generate-colors/main.go)
// must declare per group. Checking against this (rather than only a total count) catches a
// key rename or group move that would otherwise still sum to 18.
var ghDashExpectedKeysByGroup = map[string][]string{
	"text":       {"primary", "secondary", "inverted", "faint", "warning", "success", "error", "actor"},
	"background": {"selected"},
	"border":     {"primary", "secondary", "faint"},
	"icon":       {"newcontributor", "contributor", "collaborator", "member", "owner", "unknownrole"},
}

var ghDashSpecLinePattern = regexp.MustCompile(`^\s*\{"([a-z]+)",\s*"([a-zA-Z0-9]+)",\s*"([a-z_]+\.[a-z_]+)"\},$`)

// parseGhDashColorSpecs reads the ghDashColorSpecs slice literal in generate-colors/main.go
// directly. Unlike ghBoardTemplate/herdrTemplate, gh-dash has no rendered template string to
// scan: buildGhDashTemplate assembles theme.colors.<group>.<key> YAML from this Go literal at
// generate time, so the literal itself is the source of truth.
func parseGhDashColorSpecs(root string) (map[string]map[string]ghDashSpecEntry, error) {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	inSpecs := false
	foundSpecs := false
	specs := make(map[string]map[string]ghDashSpecEntry)
	total := 0
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !inSpecs {
			if line == "var ghDashColorSpecs = []ghDashColorSpec{" {
				inSpecs = true
				foundSpecs = true
			}
			continue
		}
		if line == "}" {
			inSpecs = false
			break
		}
		match := ghDashSpecLinePattern.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("%s:%d: unrecognized ghDashColorSpecs entry %q", relative, lineNumber, line)
		}
		group, key, token := match[1], match[2], match[3]
		if specs[group] == nil {
			specs[group] = make(map[string]ghDashSpecEntry)
		}
		if _, ok := specs[group][key]; ok {
			return nil, fmt.Errorf("%s:%d: duplicate gh-dash theme key %s.%s", relative, lineNumber, group, key)
		}
		specs[group][key] = ghDashSpecEntry{
			token:  verifycolors.TokenRef(token),
			source: fmt.Sprintf("%s:%d", relative, lineNumber),
		}
		total++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !foundSpecs {
		return nil, fmt.Errorf("%s: ghDashColorSpecs was not found", relative)
	}
	if inSpecs {
		return nil, fmt.Errorf("%s: ghDashColorSpecs is not terminated", relative)
	}

	var expectedTotal int
	for _, keys := range ghDashExpectedKeysByGroup {
		expectedTotal += len(keys)
	}
	if total != expectedTotal {
		return nil, fmt.Errorf("%s: ghDashColorSpecs has %d entries, want %d", relative, total, expectedTotal)
	}

	var mismatches []string
	for group, keys := range ghDashExpectedKeysByGroup {
		for _, key := range keys {
			if _, ok := specs[group][key]; !ok {
				mismatches = append(mismatches, fmt.Sprintf("missing %s.%s", group, key))
			}
		}
	}
	for group, entries := range specs {
		expectedKeys := ghDashExpectedKeysByGroup[group]
		if expectedKeys == nil {
			mismatches = append(mismatches, fmt.Sprintf("unexpected group %q", group))
			continue
		}
		allowed := make(map[string]bool, len(expectedKeys))
		for _, key := range expectedKeys {
			allowed[key] = true
		}
		for key := range entries {
			if !allowed[key] {
				mismatches = append(mismatches, fmt.Sprintf("unexpected %s.%s", group, key))
			}
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		return nil, fmt.Errorf("%s: ghDashColorSpecs key set mismatch: %s", relative, strings.Join(mismatches, ", "))
	}

	return specs, nil
}

// ghDashV423Rev pins the gh-dash version and commit this file's use-site citations were
// read against. Re-verify ghDashPairTable and the coverage notes below whenever flake.lock
// bumps gh-dash (see the gh-dash.inventory.survey coverage note for how the scan was done).
const ghDashV423Rev = "gh-dash v4.23.2 rev 78b9ca5e21dcee4740502018d46dc5e4b613db86"

// ghDashPairRow is one row of the gh-dash use-site table: a foreground slot ("group.key")
// rendered on a background slot ("group.key", or "" for ambient). The table was built from
// an exhaustive scan of gh-dash v4.23.2 production Go source (see the gh-dash.inventory.survey
// coverage note in extractGhDashTheme for the scan's method), not from the token template
// alone: several rows here don't correspond 1:1 to a single ghDashColorSpecs entry (e.g.
// background.selected.helpLabel uses one theme slot's color as a foreground on another).
//
// class is explicit per row rather than derived from background != "" because icon.*.selected
// and border.faint.selected stay report-only despite a non-ambient background (a 1-character
// glyph or a 1px rule isn't a good AA-enforcement candidate).
//
// Consumer ID (row.id, appended after "gh-dash.theme.") follows one of two conventions: when
// several foregrounds share one background (".selected" holds text.* + icon.* + border.faint,
// ".faintBorder" holds 4 rows), the suffix names the background slot. When a suffix names a
// single, specific use-site instead (".draftPill", ".helpLabel", or herdr.theme.accent.tabLabel),
// it's because that pair has exactly one foreground and doesn't generalize the way the shared
// ones do.
type ghDashPairRow struct {
	id         string
	foreground string
	background string
	role       verifycolors.PairRole
	class      verifycolors.PairClass
	source     string

	// foregroundToken, when set, is used directly instead of resolving foreground through
	// ghDashColorSpecs. It's for the 2 rows where gh-dash doesn't read a config slot at
	// all: it writes a raw ANSI SGR escape, and our terminal's ANSI palette (not any
	// gh-dash token) is what resolves it to a color.
	foregroundToken verifycolors.TokenRef
}

var ghDashPairTable = []ghDashPairRow{
	// -- background = background.selected (the panel's selected-row fill) --
	{
		id: "text.primary.selected", foreground: "text.primary", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/common/styles.go:101-109; internal/tui/context/styles.go:154-157,194-195,209-213,242-245; " +
			"internal/tui/components/table/table.go:304-374; internal/tui/components/section/section.go:415-420; " +
			"internal/tui/components/tabs/tabs.go:66-74,100-107; internal/tui/components/autocomplete/autocomplete.go:229-275; " +
			"internal/tui/components/prview/prview.go:536-540",
	},
	{
		id: "text.secondary.selected", foreground: "text.secondary", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/common/styles.go:90-98; internal/tui/components/table/table.go:304-374; " +
			"internal/tui/components/prrow/prrow.go:185-212; internal/tui/components/branch/branch.go:309-315; " +
			"internal/tui/components/notificationrow/notificationrow.go:83,110,134-170; internal/tui/components/utils.go:31-55",
	},
	{
		// Overridden to report-only in inventory_overrides.go: ViewSwitcher.Root places
		// InvertedText on SelectedBackground upstream, and that pair (not this token) is fixed.
		id: "text.inverted.selected", foreground: "text.inverted", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/context/styles.go:222-225; internal/tui/components/footer/footer.go:159-190; " +
			"internal/tui/common/styles.go:50-52",
	},
	{
		id: "text.faint.selected", foreground: "text.faint", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/context/styles.go:186-190; internal/tui/components/table/table.go:304-374; " +
			"internal/tui/components/prrow/prrow.go:30-40,81-93,127-128,230-281; internal/tui/components/branch/branch.go:60-72,174-203; " +
			"internal/tui/components/notificationrow/notificationrow.go:58-60,94-99,161-170; internal/tui/components/tabs/tabs.go:112-138; " +
			"internal/tui/components/footer/footer.go:184; internal/tui/ui.go:1568-1573",
	},
	{
		id: "text.warning.selected", foreground: "text.warning", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/table/table.go:304-374; internal/tui/components/prrow/prrow.go:76-93; " +
			"internal/tui/components/branch/branch.go:292-306; internal/tui/components/notificationrow/notificationrow.go:100-103,145-149; " +
			"internal/tui/components/footer/footer.go:56-61",
	},
	{
		id: "text.success.selected", foreground: "text.success", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/table/table.go:304-374; internal/tui/components/prrow/prrow.go:43-52,69-94,109-152; " +
			"internal/tui/components/branch/branch.go:29-55,85-138,258-270; internal/tui/components/notificationrow/notificationrow.go:88-103; " +
			"internal/tui/components/reposection/reposection.go:555-565; internal/tui/ui.go:1554-1558",
	},
	{
		id: "text.error.selected", foreground: "text.error", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/table/table.go:304-374; internal/tui/components/prrow/prrow.go:43-59,109-152; " +
			"internal/tui/components/branch/branch.go:29-45,85-138; internal/tui/components/notificationrow/notificationrow.go:88-106; " +
			"internal/tui/components/reposection/reposection.go:555-565; internal/tui/ui.go:1549-1553",
	},
	// icon.*.selected stay report-only: a 1-character glyph isn't a good fit for AA enforcement.
	{
		id: "icon.newcontributor.selected", foreground: "icon.newcontributor", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-24; internal/data/issueapi.go:61-65; internal/data/prapi.go:393-397; " +
			"internal/tui/components/issuerow/issuerow.go:22-33,75-77; internal/tui/components/prrow/prrow.go:185-216,324-353; " +
			"internal/tui/components/table/table.go:304-374",
	},
	{
		id: "icon.contributor.selected", foreground: "icon.contributor", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-30; internal/data/issueapi.go:61-65; internal/data/prapi.go:393-397; " +
			"internal/tui/components/issuerow/issuerow.go:22-33,75-77; internal/tui/components/prrow/prrow.go:185-216,324-353; " +
			"internal/tui/components/table/table.go:304-374",
	},
	{
		id: "icon.collaborator.selected", foreground: "icon.collaborator", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-27; internal/data/issueapi.go:61-65; internal/data/prapi.go:393-397; " +
			"internal/tui/components/issuerow/issuerow.go:22-33,75-77; internal/tui/components/prrow/prrow.go:185-216,324-353; " +
			"internal/tui/components/table/table.go:304-374",
	},
	{
		id: "icon.member.selected", foreground: "icon.member", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-32; internal/data/issueapi.go:61-65; internal/data/prapi.go:393-397; " +
			"internal/tui/components/issuerow/issuerow.go:22-33,75-77; internal/tui/components/prrow/prrow.go:185-216,324-353; " +
			"internal/tui/components/table/table.go:304-374",
	},
	{
		id: "icon.owner.selected", foreground: "icon.owner", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/data/issueapi.go:61-65; internal/data/prapi.go:393-397; " +
			"internal/tui/components/issuerow/issuerow.go:22-33,75-77; internal/tui/components/prrow/prrow.go:185-216,324-353; " +
			"internal/tui/components/table/table.go:304-374",
	},
	// border.faint.selected stays report-only: it's a 1px border rule, not body text.
	{
		id: "border.faint.selected", foreground: "border.faint", background: "background.selected",
		role: verifycolors.RoleBorder, class: verifycolors.ClassReportOnly,
		source: "internal/tui/components/footer/footer.go:174-189; internal/tui/common/styles.go:50-52",
	},
	// The footer's "? help" label renders SelectedBackground's color as a foreground on top
	// of FaintText used as a background: the two theme slots swap roles at this one site.
	{
		id: "background.selected.helpLabel", foreground: "background.selected", background: "text.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/footer/footer.go:51-55",
	},

	// -- background = border.faint (the row separator / ViewSwitcher panel fill) --
	{
		id: "text.primary.faintBorder", foreground: "text.primary", background: "border.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/inputbox/inputbox.go:43-56; internal/tui/ui.go:1110-1139",
	},
	{
		// Overridden to report-only in inventory_overrides.go (4.281:1, below AA).
		id: "text.secondary.faintBorder", foreground: "text.secondary", background: "border.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/inputbox/inputbox.go:49-56",
	},
	{
		// Overridden to report-only in inventory_overrides.go (3.67:1, below AA).
		id: "text.faint.faintBorder", foreground: "text.faint", background: "border.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/context/styles.go:234-236; internal/tui/components/footer/footer.go:155-157; " +
			"internal/tui/components/inputbox/inputbox.go:49-56",
	},
	{
		id: "text.success.faintBorder", foreground: "text.success", background: "border.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/footer/footer.go:138-152; internal/tui/context/styles.go:227-230",
	},
	{
		id: "border.primary.faintBorder", foreground: "border.primary", background: "border.faint",
		role: verifycolors.RoleBorder, class: verifycolors.ClassEnforced,
		source: "internal/tui/context/styles.go:231-233; internal/tui/components/footer/footer.go:174-182",
	},

	// -- background = text.faint (text.inverted's one statically-determinable use-site) --
	{
		id: "text.inverted.draftPill", foreground: "text.inverted", background: "text.faint",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/prview/prview.go:280-298; internal/tui/context/styles.go:120-122",
	},

	// -- background = background.selected, foreground = raw ANSI SGR (not a gh-dash config
	// slot at all): notificationrow.go's "new comments" count writes \x1b[97m / \x1b[32m
	// directly without a trailing reset, so it inherits whatever background the table cell
	// around it has. Our WezTerm ANSI palette (generated from ghost-visor) is what resolves
	// those SGR codes to a color, so these 2 pairs hold only under that terminal palette -
	// they're not "gh-dash reads token X" the way every other row is. brightWhite colors the
	// "+%d " count itself (variable-length text, not a glyph), so it's Enforced like any
	// other text pair; green only colors the 1-character CommentsIcon glyph next to it, so it
	// stays report-only like the other icon pairs.
	{
		id: "ansi.brightWhite.selected", foregroundToken: "ansi.bright_white", background: "background.selected",
		role: verifycolors.RoleText, class: verifycolors.ClassEnforced,
		source: "internal/tui/components/notificationrow/notificationrow.go:206-217 (renderActivity, raw \\x1b[97m " +
			"colors the \"+%d \" new-comments count text, not a glyph; no gh-dash token read - resolved by our " +
			"WezTerm ANSI palette instead); background: internal/tui/context/styles.go:191-195 (Table.SelectedCellStyle) + " +
			"internal/tui/components/table/table.go:304-365 (renderRow applies it to the selected row); " + ghDashV423Rev,
	},
	{
		id: "ansi.green.selected", foregroundToken: "ansi.green", background: "background.selected",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/tui/components/notificationrow/notificationrow.go:206-217 (renderActivity, raw \\x1b[32m, " +
			"no gh-dash token read - resolved by our WezTerm ANSI palette instead); background: " +
			"internal/tui/context/styles.go:191-195 (Table.SelectedCellStyle) + " +
			"internal/tui/components/table/table.go:304-365 (renderRow applies it to the selected row); " + ghDashV423Rev,
	},

	// -- ambient: no consumer renders these on a determinable background --
	{
		id: "text.primary", foreground: "text.primary", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/common/styles.go:45-47 (and other MainTextStyle use-sites)",
	},
	{
		id: "text.secondary", foreground: "text.secondary", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/context/styles.go:124-134 (and other SecondaryText use-sites)",
	},
	{
		id: "text.faint", foreground: "text.faint", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/common/styles.go:48-59 (and other FaintText use-sites)",
	},
	{
		id: "text.warning", foreground: "text.warning", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/common/styles.go:60-65,84-86 (and other WarningText use-sites)",
	},
	{
		id: "text.success", foreground: "text.success", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/common/styles.go:55,69-71 (and other SuccessText use-sites)",
	},
	{
		id: "text.error", foreground: "text.error", background: "",
		role: verifycolors.RoleText, class: verifycolors.ClassReportOnly,
		source: "internal/tui/common/styles.go:54,66-68 (and other ErrorText use-sites)",
	},
	{
		id: "border.primary", foreground: "border.primary", background: "",
		role: verifycolors.RoleBorder, class: verifycolors.ClassReportOnly,
		source: "internal/tui/context/styles.go:168-180,217-221; internal/tui/components/search/search.go:59-65; " +
			"internal/tui/components/sidebar/sidebar.go:51-73; internal/tui/components/tabs/tabs.go:66-74",
	},
	{
		id: "border.secondary", foreground: "border.secondary", background: "",
		role: verifycolors.RoleBorder, class: verifycolors.ClassReportOnly,
		source: "internal/tui/context/styles.go:126-134,215-216,237-240; internal/tui/components/autocomplete/autocomplete.go:220-275; " +
			"internal/tui/components/inputbox/inputbox.go:128-168; internal/tui/components/tabs/tabs.go:100-107",
	},
	{
		id: "border.faint", foreground: "border.faint", background: "",
		role: verifycolors.RoleBorder, class: verifycolors.ClassReportOnly,
		source: "internal/tui/context/styles.go:201-204; internal/tui/components/branchsidebar/branchsidebar.go:59-62; " +
			"internal/tui/components/prview/* (multiple files); internal/tui/components/table/table.go:371-374",
	},
	{
		id: "icon.newcontributor", foreground: "icon.newcontributor", background: "",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/tui/components/issueview/issueview.go:329-345; " +
			"internal/tui/components/prview/prview.go:483-499",
	},
	{
		id: "icon.contributor", foreground: "icon.contributor", background: "",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/tui/components/issueview/issueview.go:329-345; " +
			"internal/tui/components/prview/prview.go:483-499",
	},
	{
		id: "icon.collaborator", foreground: "icon.collaborator", background: "",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/tui/components/issueview/issueview.go:329-345; " +
			"internal/tui/components/prview/prview.go:483-499",
	},
	{
		id: "icon.member", foreground: "icon.member", background: "",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/tui/components/issueview/issueview.go:329-345; " +
			"internal/tui/components/prview/prview.go:483-499",
	},
	{
		id: "icon.owner", foreground: "icon.owner", background: "",
		role: verifycolors.RoleIndicator, class: verifycolors.ClassReportOnly,
		source: "internal/data/utils.go:19-34; internal/tui/components/issueview/issueview.go:329-345; " +
			"internal/tui/components/prview/prview.go:483-499",
	},
	// text.actor and icon.unknownrole have no rows: see the .unused / .upstreamUnwired
	// coverage notes below for why.
}

// ghDashResolveSlot looks up a "group.key" path (e.g. "text.primary") in the parsed
// ghDashColorSpecs. It panics on a malformed path or unknown slot: both are a bug in
// ghDashPairTable, not a runtime condition, and TestGhDashPairTableSlotsResolve catches it.
func ghDashResolveSlot(specs map[string]map[string]ghDashSpecEntry, slot string) ghDashSpecEntry {
	dot := strings.IndexByte(slot, '.')
	if dot < 0 {
		panic(fmt.Sprintf("gh-dash pair table: malformed slot reference %q", slot))
	}
	group, key := slot[:dot], slot[dot+1:]
	entry, ok := specs[group][key]
	if !ok {
		panic(fmt.Sprintf("gh-dash pair table: unknown slot reference %q", slot))
	}
	return entry
}

// extractGhDashTheme registers contrast pairs for gh-dash/config.yml's generated
// theme.colors consumers, from ghDashPairTable (the use-site table) plus the
// background.selected surface pair. See gh_dash_test.go for the expected pair table and
// inventory_overrides.go for the three pairs demoted to report-only after generation.
func extractGhDashTheme(root string, result *Result) error {
	specs, err := parseGhDashColorSpecs(root)
	if err != nil {
		return err
	}

	for _, row := range ghDashPairTable {
		background := verifycolors.AmbientBackground()
		if row.background != "" {
			background = verifycolors.TokenBackground(ghDashResolveSlot(specs, row.background).token)
		}
		foreground := row.foregroundToken
		if foreground == "" {
			foreground = ghDashResolveSlot(specs, row.foreground).token
		}
		result.addPair(verifycolors.PairSpec{
			ConsumerID: "gh-dash.theme." + row.id,
			Foreground: foreground,
			Background: background,
			Class:      row.class,
			Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			Role:       row.role,
			Source:     row.source,
		})
	}

	backgroundSlots := specs["background"]
	result.addPair(surfacePair("gh-dash.theme.background.selected", backgroundSlots["selected"].token, backgroundSlots["selected"].source))

	// gh-dash.terminalDefault.taskStart: the running-task status line sets
	// Background(SelectedBackground) but never calls Foreground, so it renders in whatever
	// the terminal's own default foreground is - not a gh-dash config slot at all (same shape
	// as the 2 raw-ANSI pairs above). Consumer ID uses "gh-dash.terminalDefault." rather than
	// "gh-dash.theme." to make that explicit.
	result.addPair(verifycolors.PairSpec{
		ConsumerID: "gh-dash.terminalDefault.taskStart",
		Foreground: "foregrounds.main",
		Background: verifycolors.TokenBackground(backgroundSlots["selected"].token),
		Class:      verifycolors.ClassEnforced,
		Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		Role:       verifycolors.RoleText,
		Source: "internal/tui/ui.go:1541-1548 (TaskStart case: Background(SelectedBackground), no Foreground call) + " +
			"internal/tui/ui.go:92-93 (taskSpinner.Style, same shape); background wiring " +
			"internal/tui/theme/theme.go:79-82 (SelectedBackground shim); gh-dash sets no foreground here at all, " +
			"so the terminal's default foreground applies - in our environment that's wezterm/wezterm.lua:203 " +
			"(foreground = colors.foregrounds.main), which is why this pair uses foregrounds.main directly " +
			"rather than resolving through ghDashColorSpecs. The ratio (9.85) currently matches " +
			"gh-dash.theme.text.primary.selected because both resolve to foregrounds.main today; that's a " +
			"coincidence of the current mapping, not a shared use-site - a future divergence between gh-dash's " +
			"text.primary token and our terminal's default foreground would only be caught here; " + ghDashV423Rev,
	})

	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.theme.text.inverted.statusPill",
		Reason: "the non-draft status pill (PR: OPEN non-draft/CLOSED/MERGED; issue: OPEN/CLOSED) renders " +
			"InvertedText on a hardcoded background, not a palette token, so it can't be a static pair. " +
			"gh-dash v4.23.2 hardcodes 4 such colors: OpenPR/OpenIssue #42A0FA, ClosedPR #656C76, " +
			"MergedPR #A371F7, ClosedIssue #C38080; re-check at the next gh-dash bump in case any of them moves to a token",
		Source: "internal/tui/context/styles.go:96-116 (Colors.OpenPR/ClosedPR/MergedPR/OpenIssue/ClosedIssue literals) + " +
			"internal/tui/context/styles.go:120-122 (PillStyle Foreground(InvertedText)) + " +
			"internal/tui/components/prview/prview.go:280-298 (PR status pill) + " +
			"internal/tui/components/issueview/issueview.go:311-326 (issue status pill); " +
			"config→Theme wiring internal/tui/theme/theme.go:107-110 (InvertedText shim); " + ghDashV423Rev,
	})
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.theme.text.inverted.labels",
		Reason: "label pill backgrounds are GitHub label colors fetched at runtime, not palette tokens, " +
			"so contrast can't be statically guaranteed",
		Source: "internal/tui/context/styles.go:120-122 (PillStyle Foreground(InvertedText)) + " +
			"internal/tui/components/prview/prview.go:301-315 (renderLabels) + " +
			"internal/tui/common/labels.go:9-22 (label background wiring); " +
			"config→Theme wiring internal/tui/theme/theme.go:107-110 (InvertedText shim); " + ghDashV423Rev,
	})
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.theme.icon.unknownrole.upstreamUnwired",
		Reason: "gh-dash v4.23.2's ParseTheme has no assignment for UnknownRoleIconColor (the icon color " +
			"shim block ends right after Owner); the config value parses successfully but is never " +
			"reflected into the running theme. Generation keeps producing this key alongside the other " +
			"5 icon colors, but no contrast pair is registered for it",
		Source: "internal/config/parser.go:253 (ColorThemeIcon.UnknownRole yaml field) + " +
			"internal/tui/theme/theme.go:28 (UnknownRoleIconColor field) + " +
			"internal/tui/theme/theme.go:147 (ParseTheme's icon color shim block ends without it); " + ghDashV423Rev,
	})
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.theme.text.actor.unused",
		Reason: "ActorText has zero production renderer references in gh-dash v4.23.2: it's defined on " +
			"Theme and assigned in ParseTheme, but no component reads theme.ActorText to render anything; " +
			"re-check at the next gh-dash bump, and when upstream starts rendering it, re-check whether " +
			"foregrounds.heading collides with any other gh-dash slot (an unused slot's color is the hardest " +
			"collision to notice, because nothing renders it until upstream wires it up). " +
			"Precedent: herdr.theme.peach.unused (herdr.go:119-123)",
		Source: "internal/tui/theme/theme.go:22 (field) + internal/tui/theme/theme.go:49 (default) + " +
			"internal/tui/theme/theme.go:123-126 (ParseTheme shim); " + ghDashV423Rev,
	})
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.inventory.survey",
		Reason: "gh-dash v4.23.2's production Go source (excluding _test.go) was scanned exhaustively for " +
			"the gh-dash.theme.* use-site table in gh_dash.go, in three passes over one direction each - the " +
			"first review round only covered pass 1, and a later round added pass 2; both misses are why all " +
			"3 passes are now run on every re-verification. Pass 1 (foreground is a theme slot): classified by " +
			"what the background is - (a) a theme slot, 21 sites, static pair; " +
			"(b) an upstream-hardcoded literal, 4 sites, coverage note; (c) runtime/external data (e.g. GitHub " +
			"label colors), 1 site, coverage note; (d) ambient/undetermined at the call site, 15 sites, " +
			"report-only. Pass 2 (foreground is not a theme slot, background is a theme slot - the reverse " +
			"direction, tracked in gh-dash.theme.literalForegrounds): 19 sites, of which 2 resolve through our " +
			"own WezTerm ANSI palette (raw ANSI SGR codes) and are registered as static pairs; the remaining " +
			"17 are hardcoded literals or state-based literal choices and stay a coverage note. Pass 3 " +
			"(background is a theme slot, no Foreground call at all in the same style chain, so the terminal's " +
			"own default foreground applies): 1 site so far (gh-dash.terminalDefault.taskStart) - scan for " +
			"Background( calls whose style chain never calls Foreground(. Scan targets: " +
			"every Foreground(/Border*Foreground(/Styles.Colors.* reference, every literal AdaptiveColor{}/" +
			"lipgloss.Color(...), every raw \\x1b[ escape, and every .Inherit(...) call (0 found in gh-dash's " +
			"own source; Bubbles v0.21.0's internal use in textarea is the only known instance and is handled " +
			"per-call-site, e.g. the text.secondary.faintBorder override). Every row-producing type (PR, branch, " +
			"Issue, notification - the only 4 that implement ToTableRow/table.Row{}) was checked exhaustively",
		Source: ghDashV423Rev + "; row producers: internal/tui/components/prrow/prrow.go:324-353, " +
			"internal/tui/components/branch/branch.go:230-255, internal/tui/components/issuerow/issuerow.go:22-33, " +
			"internal/tui/components/notificationrow/notificationrow.go:22-29; checked and excluded as out of " +
			"scope: ClosedIssue (pill background only, already in the statusPill note), Styles.Colors.SuccessText " +
			"(background is the terminal's own canvas, not a theme slot), the checks component's border, " +
			"MergedGlyph, the carousel's default ANSI index 212 (overridden by the theme style), LogoColor, the " +
			"CLI's sponsor colors, and Bubbles textinput's default index 240 (its prompt component isn't " +
			"rendered by gh-dash); re-run all 3 passes and re-diff against ghDashPairTable in sourceinventory/" +
			"gh_dash.go whenever flake.lock bumps gh-dash. Always fetch files at the pinned rev: GitHub's " +
			"code search API answers for HEAD, so a slot that HEAD renders can look used when the pinned rev " +
			"never renders it (ActorText is exactly that case)",
	})
	result.addCoverageNote(verifycolors.CoverageNote{
		ID: "gh-dash.theme.literalForegrounds",
		Reason: "sites where the rendered foreground is a literal or runtime value while the background is " +
			"a theme slot (the reverse of the usual token-foreground-on-literal-background direction); these " +
			"can't be static pairs from the token side either, since the foreground isn't a token (2 sibling " +
			"sites where the foreground turned out to be measurable through our own ANSI palette are registered " +
			"as pairs instead: gh-dash.theme.ansi.brightWhite.selected and .ansi.green.selected). None of the " +
			"17 sites below is a runtime/external color (e.g. a GitHub label); all pick among a fixed set of " +
			"hardcoded literals depending on state. 12 use Styles.Colors.* (OpenPR and OpenIssue are the same " +
			"alias, #42A0FA) on SelectedBackground across the PR/branch/issue/notification rows: OpenPR/OpenIssue " +
			"#42A0FA, ClosedPR #656C76 (notification rows use ClosedPR for closed issues too, not ClosedIssue), " +
			"MergedPR #A371F7. 3 are literal AdaptiveColor{} values: the Discussion icon (black/white), the " +
			"Release icon (blue), and the active NotificationsView icon (gold, on FaintBorder - the one " +
			"originally known here). 2 are literal lipgloss.Color(...) values: ANSI 256-color-cube index 33 " +
			"(the unread blue dot; not representable in our 16-color palette) and hex literal #e0af68 (the repo " +
			"section footer's modified-file count). notificationrow.go's raw ANSI is confirmed to be only the " +
			"2 escapes covered by gh-dash.theme.ansi.brightWhite.selected / .ansi.green.selected (\\x1b[97m and " +
			"\\x1b[32m) - a whole-file grep for \\x1b found exactly those 2 matches, no \\x1b[90m or others",
		Source: "Styles.Colors.* literals: internal/tui/context/styles.go:96-116; OpenPR x selected: " +
			"internal/tui/components/prrow/prrow.go:76-85, internal/tui/components/branch/branch.go:58-64, " +
			"internal/tui/components/notificationrow/notificationrow.go:47-64; ClosedPR x selected: " +
			"internal/tui/components/prrow/prrow.go:86-88, internal/tui/components/branch/branch.go:65-67, " +
			"internal/tui/components/notificationrow/notificationrow.go:50-56,66-73; MergedPR x selected: " +
			"internal/tui/components/prrow/prrow.go:89-91, internal/tui/components/branch/branch.go:68-70, " +
			"internal/tui/components/notificationrow/notificationrow.go:50-53; OpenIssue x selected: " +
			"internal/tui/components/issuerow/issuerow.go:87-92, internal/tui/components/notificationrow/notificationrow.go:66-74; " +
			"Discussion icon (AdaptiveColor{Light:#000000,Dark:#ffffff}) x selected: " +
			"internal/tui/components/notificationrow/notificationrow.go:75-78; Release icon " +
			"(AdaptiveColor{Light:#0969da,Dark:#58a6ff}) x selected: " +
			"internal/tui/components/notificationrow/notificationrow.go:78-81; unread blue dot (lipgloss.Color(\"33\")) " +
			"x selected: internal/tui/components/notificationrow/notificationrow.go:113-125; active NotificationsView icon " +
			"(AdaptiveColor{Light:#B8860B,Dark:#FFD700}) x FaintBorder: internal/tui/components/footer/footer.go:138-152 + " +
			"internal/tui/context/styles.go:227-230; repo footer modified-count (lipgloss.Color(\"#e0af68\")) x selected: " +
			"internal/tui/components/reposection/reposection.go:555-565 + internal/tui/context/styles.go:186-190; " +
			"config→Theme wiring: SelectedBackground = internal/tui/theme/theme.go:79-82, FaintBorder = " +
			"internal/tui/theme/theme.go:87-90; " + ghDashV423Rev,
	})

	return nil
}
