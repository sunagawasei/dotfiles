package verifycolors

import "fmt"

type pairOverride struct {
	Class      PairClass
	Role       PairRole
	Background *BackgroundRef
	Reason     string
	Source     string
}

var manualPairs = []PairSpec{
	{
		ConsumerID: "herdr.theme.accent.tabLabel",
		Foreground: "ansi.bright_black",
		Background: TokenBackground("purples.lavender"),
		Class:      ClassEnforced,
		Profiles:   []RenderProfile{ProfileTruecolor},
		Role:       RoleText,
		Source:     "manual override: scripts/internal/verifycolors/inventory_overrides.go; herdr v0.8.0 rev 346411fa21afd297f5ed3b3fa56f9e3fbf7654b7 src/ui/tabs.rs:325 + src/ui/widgets.rs:32-36; .bg(accent) 全18箇所が panel_contrast_fg を使用; herdrTemplate の panel_bg==\"reset\" かつ surface_dim=={{ansi.bright_black}} である間だけこの対が成立する(generate-colors の assert で強制); flake.lock で herdr を bump したら再検証すること",
	},
}

var pairOverrides = map[string]pairOverride{
	"nvim.highlight.@markup.strikethrough": {
		Class:  ClassWaived,
		Reason: "strikethrough shape is the primary cue; the foreground is intentionally dimmed",
	},
	"nvim.highlight.Conceal": {
		Class:  ClassWaived,
		Reason: "concealed text is intentionally dimmed",
	},
	"nvim.highlight.LazyDimmed": {
		Class:  ClassWaived,
		Reason: "Lazy UI marks this content as intentionally dimmed",
	},
	"vim.highlight.Conceal": {
		Class:  ClassWaived,
		Reason: "concealed text is intentionally dimmed",
	},
	"vim.highlight.Special": {
		Class:  ClassWaived,
		Reason: "legacy Special is retained for intentional dimming; visible Markdown delimiters use dedicated highlights",
	},
	"hunk.theme.lineNumberFg": {
		Class:      ClassEnforced,
		Background: backgroundPointer(TokenBackground("core.background")),
	},
	"hunk.theme.noteTitleText": {
		Class:      ClassEnforced,
		Background: backgroundPointer(TokenBackground("teals.dark_accent")),
	},
	"hunk.theme.text": {
		Class:      ClassEnforced,
		Background: backgroundPointer(TokenBackground("core.background")),
	},
	"hunk.theme.muted": {
		Class:      ClassEnforced,
		Background: backgroundPointer(TokenBackground("core.background")),
	},
	"gh-dash.theme.text.inverted.selected": {
		Class: ClassReportOnly,
		Reason: "gh-dash上流が ViewSwitcher.Root で InvertedText を SelectedBackground 上に置く" +
			"(internal/tui/context/styles.go:222-225、Background元は internal/tui/common/styles.go:50-52 の" +
			"FooterStyle、gh-dash v4.23.2)。gh-dash既定テーマも inverted #242347 on selected #39386B で" +
			"同程度の比。InvertedTextは PrView.PillStyle(internal/tui/context/styles.go:120-122)で" +
			"固定色バッジの上にも載るため暗値が必要で、token変更では解けない",
		Source: "manual override: scripts/internal/verifycolors/inventory_overrides.go; gh-dash v4.23.2 rev " +
			"78b9ca5e21dcee4740502018d46dc5e4b613db86; foreground definition internal/tui/context/styles.go:120-122 " +
			"(PillStyle Foreground(InvertedText)); background setting internal/tui/context/styles.go:222-225 " +
			"(ViewSwitcher.Root) + internal/tui/components/footer/footer.go:159-190 (renderViewSwitcher applies it); " +
			"config→Theme wiring internal/tui/theme/theme.go:107-110 (InvertedText shim)",
	},
	"gh-dash.theme.text.faint.faintBorder": {
		Class: ClassReportOnly,
		Reason: "border.faint は表の行区切り線(internal/tui/context/styles.go:202-204)と" +
			"ViewSwitcherの塗り(同:230-236)の二役。config が showSeparator: true のため区切り線の可視性を優先し、" +
			"塗りの上の補助テキストは 3.67:1 を受容する。gh-dash v4.23.2",
		Source: "manual override: scripts/internal/verifycolors/inventory_overrides.go; gh-dash v4.23.2 rev " +
			"78b9ca5e21dcee4740502018d46dc5e4b613db86; internal/tui/context/styles.go:234-236 " +
			"(ViewSwitcher.InactiveView Foreground(FaintText) on Background(FaintBorder)) + " +
			"internal/tui/components/footer/footer.go:155-157 (renders InactiveView) + " +
			"internal/tui/components/inputbox/inputbox.go:49-56; config→Theme wiring " +
			"internal/tui/theme/theme.go:95-98 (FaintText shim) and internal/tui/theme/theme.go:87-90 (FaintBorder shim); " +
			"requires gh-dash/config.yml theme.ui.table.showSeparator: true (asserted in sourceinventory/gh_dash_test.go)",
	},
	"gh-dash.theme.text.secondary.faintBorder": {
		Class: ClassReportOnly,
		Reason: "補助UIの行番号(inputbox cursor line number)。border.faint は表の行区切り線と塗りの二役で、" +
			"config が showSeparator: true のため区切り線の可視性を優先する。配色変更で解くには border.faint を" +
			"暗くするしかなく、それは先行taskで却下した方向。実測 4.281:1 で AA を割る",
		Source: "manual override: scripts/internal/verifycolors/inventory_overrides.go; gh-dash v4.23.2 rev " +
			"78b9ca5e21dcee4740502018d46dc5e4b613db86; internal/tui/components/inputbox/inputbox.go:49-56 " +
			"(CursorLineNumber Foreground(SecondaryText) on CursorLine's Background(FaintBorder); " +
			"Bubbles v0.21.0's textarea.Styles applies CursorLineNumber.Inherit(CursorLine) internally, so " +
			"the background isn't set at this call site directly - re-verify this inheritance on a Bubbles bump too); " +
			"config→Theme wiring internal/tui/theme/theme.go:103-106 (SecondaryText shim) and internal/tui/theme/theme.go:87-90 (FaintBorder shim)",
	},
}

func ContractPairs() ([]PairSpec, error) {
	pairs := GeneratedPairs()
	pairs = append(pairs, manualPairs...)
	byID := make(map[string]int, len(pairs))
	for index, pair := range pairs {
		if _, exists := byID[pair.ConsumerID]; exists {
			return nil, fmt.Errorf("duplicate generated consumer ID %q", pair.ConsumerID)
		}
		byID[pair.ConsumerID] = index
	}
	for consumerID, override := range pairOverrides {
		index, ok := byID[consumerID]
		if !ok {
			return nil, fmt.Errorf("override references unknown consumer ID %q", consumerID)
		}
		pair := &pairs[index]
		if override.Class != "" {
			pair.Class = override.Class
		}
		if override.Role != "" {
			pair.Role = override.Role
		}
		if override.Background != nil {
			pair.Background = *override.Background
		}
		if override.Reason != "" {
			pair.Reason = override.Reason
		}
		if override.Source != "" {
			pair.Source = override.Source
		}
	}
	return pairs, nil
}

func backgroundPointer(background BackgroundRef) *BackgroundRef {
	return &background
}
