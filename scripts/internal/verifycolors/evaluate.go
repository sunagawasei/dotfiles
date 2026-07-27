package verifycolors

import (
	"fmt"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
	"github.com/sunagawasei/dotfiles/scripts/internal/xterm"
)

type EnvironmentProfile struct {
	Name  string
	Color string
}

type ProfileResult struct {
	ConsumerID      string
	DeclaredClass   PairClass
	Class           PairClass
	Role            PairRole
	Profile         RenderProfile
	Environment     string
	ForegroundToken TokenRef
	BackgroundToken TokenRef
	ForegroundColor string
	BackgroundColor string
	ForegroundIndex *int
	BackgroundIndex *int
	Ratio           float64
	Pass            bool
}

type Evaluation struct {
	Results      []ProfileResult
	EnforcedNG   int
	EnforcedPass int
	ReportOnly   int
	Waived       int
}

func Evaluate(contract Contract, palette *Palette, environments []EnvironmentProfile) (Evaluation, error) {
	values := palette.TokenValues()
	ansi, err := palette.WezTermANSI()
	if err != nil {
		return Evaluation{}, err
	}
	namedEnvironments := []EnvironmentProfile{{
		Name:  "nominal",
		Color: values["core.background"],
	}}
	environmentNames := map[string]bool{"nominal": true}
	for _, environment := range environments {
		if environmentNames[environment.Name] {
			return Evaluation{}, fmt.Errorf("duplicate environment profile %q", environment.Name)
		}
		environmentNames[environment.Name] = true
		namedEnvironments = append(namedEnvironments, environment)
	}
	environments = namedEnvironments
	for _, environment := range environments {
		if environment.Name == "" {
			return Evaluation{}, fmt.Errorf("environment profile name cannot be empty")
		}
		if _, _, _, err := colorutil.ParseHexRGB8(environment.Color); err != nil {
			return Evaluation{}, fmt.Errorf("environment profile %q: %w", environment.Name, err)
		}
	}

	var evaluation Evaluation
	for _, pair := range contract.Pairs {
		if pair.Class == ClassWaived {
			evaluation.Waived++
			continue
		}
		for _, profile := range pair.Profiles {
			backgrounds := environments
			if !pair.Background.Ambient {
				backgrounds = []EnvironmentProfile{{
					Name:  "declared",
					Color: values[pair.Background.Token],
				}}
			}
			for _, background := range backgrounds {
				result, err := evaluateProfile(pair, profile, background, values, ansi)
				if err != nil {
					return Evaluation{}, err
				}
				evaluation.Results = append(evaluation.Results, result)
				switch result.Class {
				case ClassEnforced:
					if result.Pass {
						evaluation.EnforcedPass++
					} else {
						evaluation.EnforcedNG++
					}
				case ClassReportOnly:
					evaluation.ReportOnly++
				}
			}
		}
	}
	return evaluation, nil
}

func evaluateProfile(
	pair PairSpec,
	profile RenderProfile,
	background EnvironmentProfile,
	values map[TokenRef]string,
	ansi [16]xterm.RGB,
) (ProfileResult, error) {
	foregroundColor := values[pair.Foreground]
	backgroundColor := background.Color
	result := ProfileResult{
		ConsumerID:      pair.ConsumerID,
		DeclaredClass:   pair.Class,
		Class:           pair.Class,
		Role:            pair.Role,
		Profile:         profile,
		Environment:     background.Name,
		ForegroundToken: pair.Foreground,
		BackgroundToken: pair.Background.Token,
		ForegroundColor: foregroundColor,
		BackgroundColor: backgroundColor,
	}
	switch profile {
	case ProfileTruecolor:
	case ProfileCtermFGOnly:
		// Index selection follows the generator's standard xterm-256 nearest
		// match. Resolve256 then substitutes WezTerm's custom RGB for 0-15.
		// For example, #F8FCFD ties at 15 and 231; the generator's stable
		// first-match rule selects 15, which WezTerm renders as #F8FCFD.
		index, _, err := xterm.Nearest256(foregroundColor)
		if err != nil {
			return result, fmt.Errorf("%s foreground: %w", pair.ConsumerID, err)
		}
		resolved, err := xterm.Resolve256(index, ansi)
		if err != nil {
			return result, fmt.Errorf("%s foreground: %w", pair.ConsumerID, err)
		}
		result.ForegroundIndex = intPointer(index)
		result.ForegroundColor = resolved.Hex()
	case ProfileCtermFGBG:
		foregroundIndex, _, err := xterm.Nearest256(foregroundColor)
		if err != nil {
			return result, fmt.Errorf("%s foreground: %w", pair.ConsumerID, err)
		}
		backgroundIndex, _, err := xterm.Nearest256(backgroundColor)
		if err != nil {
			return result, fmt.Errorf("%s background: %w", pair.ConsumerID, err)
		}
		resolvedForeground, err := xterm.Resolve256(foregroundIndex, ansi)
		if err != nil {
			return result, fmt.Errorf("%s foreground: %w", pair.ConsumerID, err)
		}
		resolvedBackground, err := xterm.Resolve256(backgroundIndex, ansi)
		if err != nil {
			return result, fmt.Errorf("%s background: %w", pair.ConsumerID, err)
		}
		result.ForegroundIndex = intPointer(foregroundIndex)
		result.BackgroundIndex = intPointer(backgroundIndex)
		result.ForegroundColor = resolvedForeground.Hex()
		result.BackgroundColor = resolvedBackground.Hex()
	default:
		return result, fmt.Errorf("%s has unsupported profile %q", pair.ConsumerID, profile)
	}
	if profile != ProfileTruecolor && result.Class == ClassEnforced {
		result.Class = ClassReportOnly
	}
	ratio, err := colorutil.StrictContrastRatio(result.ForegroundColor, result.BackgroundColor)
	if err != nil {
		return result, fmt.Errorf("%s contrast: %w", pair.ConsumerID, err)
	}
	result.Ratio = ratio
	result.Pass = ratio >= MinimumContrast
	return result, nil
}

func intPointer(value int) *int {
	return &value
}
