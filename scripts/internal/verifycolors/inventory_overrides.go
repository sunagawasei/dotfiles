package verifycolors

import "fmt"

type pairOverride struct {
	Class      PairClass
	Role       PairRole
	Background *BackgroundRef
	Reason     string
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
	}
	return pairs, nil
}

func backgroundPointer(background BackgroundRef) *BackgroundRef {
	return &background
}
