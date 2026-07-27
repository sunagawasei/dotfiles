package verifycolors

import (
	"fmt"
	"sort"
	"strings"
)

const MinimumContrast = 4.5

type Contract struct {
	Pairs        []PairSpec
	Dispositions []TokenDisposition
	Coverage     []CoverageNote
}

type ContractSummary struct {
	Enforced        int
	ReportOnly      int
	Waived          int
	Disposition     int
	Unclassified    []TokenRef
	CoverageNotes   int
	ReferencedToken int
}

func DefaultContract() (Contract, error) {
	pairs, err := ContractPairs()
	if err != nil {
		return Contract{}, err
	}
	return Contract{
		Pairs:        pairs,
		Dispositions: tokenDispositions(),
		Coverage:     GeneratedCoverageNotes(),
	}, nil
}

func ValidateContract(contract Contract, palette *Palette) (ContractSummary, error) {
	summary := ContractSummary{
		Disposition:   len(contract.Dispositions),
		CoverageNotes: len(contract.Coverage),
	}
	values := palette.TokenValues()
	referenced := make(map[TokenRef]bool)
	consumerIDs := make(map[string]bool)
	var problems []string

	for _, pair := range contract.Pairs {
		if strings.TrimSpace(pair.ConsumerID) == "" {
			problems = append(problems, "pair has an empty consumer ID")
		} else if consumerIDs[pair.ConsumerID] {
			problems = append(problems, fmt.Sprintf("duplicate consumer ID %q", pair.ConsumerID))
		}
		consumerIDs[pair.ConsumerID] = true
		if _, ok := values[pair.Foreground]; !ok {
			problems = append(problems, fmt.Sprintf("%s references unknown foreground token %q", pair.ConsumerID, pair.Foreground))
		} else {
			referenced[pair.Foreground] = true
		}
		if pair.Background.Ambient && pair.Background.Token != "" {
			problems = append(problems, fmt.Sprintf("%s background cannot be both ambient and token-based", pair.ConsumerID))
		}
		if !pair.Background.Ambient {
			if pair.Background.Token == "" {
				problems = append(problems, fmt.Sprintf("%s has no background", pair.ConsumerID))
			} else if _, ok := values[pair.Background.Token]; !ok {
				problems = append(problems, fmt.Sprintf("%s references unknown background token %q", pair.ConsumerID, pair.Background.Token))
			} else {
				referenced[pair.Background.Token] = true
			}
		}
		switch pair.Class {
		case ClassEnforced:
			summary.Enforced++
			if pair.Background.Ambient {
				problems = append(problems, fmt.Sprintf("%s is enforced but has an ambient background", pair.ConsumerID))
			}
		case ClassReportOnly:
			summary.ReportOnly++
		case ClassWaived:
			summary.Waived++
			if strings.TrimSpace(pair.Reason) == "" {
				problems = append(problems, fmt.Sprintf("%s is waived without a reason", pair.ConsumerID))
			}
		default:
			problems = append(problems, fmt.Sprintf("%s has unknown class %q", pair.ConsumerID, pair.Class))
		}
		switch pair.Role {
		case RoleText, RoleBorder, RoleIndicator, RoleSurface:
		default:
			problems = append(problems, fmt.Sprintf("%s has unknown role %q", pair.ConsumerID, pair.Role))
		}
		if len(pair.Profiles) == 0 {
			problems = append(problems, fmt.Sprintf("%s has no render profile", pair.ConsumerID))
		}
		seenProfiles := make(map[RenderProfile]bool)
		for _, profile := range pair.Profiles {
			if seenProfiles[profile] {
				problems = append(problems, fmt.Sprintf("%s repeats profile %q", pair.ConsumerID, profile))
			}
			seenProfiles[profile] = true
			switch profile {
			case ProfileTruecolor:
			case ProfileCtermFGOnly, ProfileCtermFGBG:
				if !strings.HasPrefix(pair.ConsumerID, "vim.") {
					problems = append(problems, fmt.Sprintf("%s uses cterm profile outside Vim", pair.ConsumerID))
				}
			default:
				problems = append(problems, fmt.Sprintf("%s has unknown render profile %q", pair.ConsumerID, profile))
			}
		}
		if pair.Class == ClassEnforced && !seenProfiles[ProfileTruecolor] {
			problems = append(problems, fmt.Sprintf("%s is enforced without a primary truecolor profile", pair.ConsumerID))
		}
	}

	dispositions := make(map[TokenRef]bool)
	for _, disposition := range contract.Dispositions {
		if _, ok := values[disposition.Token]; !ok {
			problems = append(problems, fmt.Sprintf("disposition references unknown token %q", disposition.Token))
		}
		if dispositions[disposition.Token] {
			problems = append(problems, fmt.Sprintf("duplicate disposition for %q", disposition.Token))
		}
		dispositions[disposition.Token] = true
		if referenced[disposition.Token] {
			problems = append(problems, fmt.Sprintf("token %q is both referenced and dispositioned", disposition.Token))
		}
		switch disposition.Kind {
		case DispositionBackgroundNonText, DispositionUnused, DispositionOutOfScope, DispositionProjection:
		default:
			problems = append(problems, fmt.Sprintf("token %q has unknown disposition %q", disposition.Token, disposition.Kind))
		}
		if strings.TrimSpace(disposition.Reason) == "" {
			problems = append(problems, fmt.Sprintf("token %q disposition has no reason", disposition.Token))
		}
	}
	for token := range values {
		if !referenced[token] && !dispositions[token] {
			summary.Unclassified = append(summary.Unclassified, token)
		}
	}
	sort.Slice(summary.Unclassified, func(i, j int) bool {
		return summary.Unclassified[i] < summary.Unclassified[j]
	})
	summary.ReferencedToken = len(referenced)
	if len(summary.Unclassified) > 0 {
		problems = append(problems, fmt.Sprintf("unclassified palette tokens: %s", joinTokenRefs(summary.Unclassified)))
	}
	for _, note := range contract.Coverage {
		if strings.TrimSpace(note.ID) == "" || strings.TrimSpace(note.Reason) == "" {
			problems = append(problems, "coverage notes require an ID and reason")
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return summary, fmt.Errorf("invalid color contract:\n  %s", strings.Join(problems, "\n  "))
	}
	return summary, nil
}

func joinTokenRefs(tokens []TokenRef) string {
	values := make([]string, len(tokens))
	for index, token := range tokens {
		values[index] = string(token)
	}
	return strings.Join(values, ", ")
}

func tokenDispositions() []TokenDisposition {
	projectionReason := "generated alias/export of a canonical token that is evaluated at its canonical use sites"
	outOfScopeReason := "used by a configuration surface outside the static use-site inventory"
	unusedReason := "no use site was found in the repository"
	return []TokenDisposition{
		{Token: "nvim.bg", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.border", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.comment", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.dark_shadow", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.function", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.git_blame", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.header", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.keyword", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.number", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.operator", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.punctuation", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.selection", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.string", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.type", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "nvim.variable", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "semantic.git_blame", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "semantic.info", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "semantic.warning", Kind: DispositionProjection, Reason: projectionReason},
		{Token: "wezterm.tab_bar", Kind: DispositionProjection, Reason: projectionReason},

		{Token: "blues_slates.ocean_blue", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (LazyGit author color)"},
		{Token: "zsh.command", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},
		{Token: "zsh.comment", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},
		{Token: "zsh.default", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},
		{Token: "zsh.option", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},
		{Token: "zsh.path", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},
		{Token: "zsh.string", Kind: DispositionOutOfScope, Reason: outOfScopeReason + " (shell prompt/highlighting)"},

		{Token: "blues_slates.deep_ocean", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "purples.dark_purple", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "purples.deep_glitch", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "purples.soft_magenta", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "teals.deep", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "teals.muted", Kind: DispositionUnused, Reason: unusedReason},
		{Token: "teals.sea_green", Kind: DispositionUnused, Reason: unusedReason},
	}
}
