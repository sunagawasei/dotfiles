package verifycolors

import "fmt"

type pairOverride struct {
	Class      PairClass
	Role       PairRole
	Background *BackgroundRef
	Reason     string
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
	"nvim.highlight.FlashBackdrop": {
		Class:  ClassWaived,
		Reason: "Flash backdrop is intentionally dimmed to emphasize the active match",
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
