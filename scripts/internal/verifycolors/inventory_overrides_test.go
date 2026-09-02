package verifycolors

import "testing"

func TestContractPairsAppliesMeaningOverrides(t *testing.T) {
	pairs, err := ContractPairs()
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]PairSpec, len(pairs))
	for _, pair := range pairs {
		byID[pair.ConsumerID] = pair
	}

	conceal := byID["nvim.highlight.Conceal"]
	if conceal.Class != ClassWaived || conceal.Reason == "" {
		t.Fatalf("Conceal override = class %q reason %q, want waived with reason", conceal.Class, conceal.Reason)
	}

	lineNumber := byID["hunk.theme.lineNumberFg"]
	if lineNumber.Class != ClassEnforced ||
		lineNumber.Background.Ambient ||
		lineNumber.Background.Token != "core.background" {
		t.Fatalf("lineNumberFg override = %#v, want enforced on core.background", lineNumber)
	}

	tabLabel := byID["herdr.theme.accent.tabLabel"]
	if tabLabel.Class != ClassEnforced ||
		tabLabel.Background.Ambient ||
		tabLabel.Background.Token != "purples.lavender" ||
		tabLabel.Foreground != "ansi.bright_black" {
		t.Fatalf("herdr accent tab label = %#v, want enforced ansi.bright_black on purples.lavender", tabLabel)
	}
}
