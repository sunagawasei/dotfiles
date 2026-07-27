package verifycolors

import (
	"fmt"
	"sort"

	"github.com/sunagawasei/dotfiles/scripts/internal/cvd"
)

const CVDDeltaEThreshold = 20.0

type CVDResult struct {
	Type        string
	First       TokenRef
	Second      TokenRef
	DeltaE      float64
	BelowScreen bool
}

type CVDReport struct {
	Results     []CVDResult
	BelowScreen int
	Enforced    int
}

func EvaluateANSIColorVision(palette *Palette) (CVDReport, error) {
	names := make([]string, 0, len(palette.ANSI))
	for name := range palette.ANSI {
		names = append(names, name)
	}
	sort.Strings(names)
	var report CVDReport
	for _, kind := range cvd.Types {
		labs := make(map[string][3]float64, len(names))
		for _, name := range names {
			lab, err := cvd.Apply(palette.ANSI[name], kind.Matrix)
			if err != nil {
				return report, fmt.Errorf("%s ansi.%s: %w", kind.Name, name, err)
			}
			labs[name] = lab
		}
		for firstIndex := 0; firstIndex < len(names); firstIndex++ {
			for secondIndex := firstIndex + 1; secondIndex < len(names); secondIndex++ {
				first := names[firstIndex]
				second := names[secondIndex]
				delta := cvd.DeltaE76(labs[first], labs[second])
				result := CVDResult{
					Type:        kind.Name,
					First:       TokenRef("ansi." + first),
					Second:      TokenRef("ansi." + second),
					DeltaE:      delta,
					BelowScreen: delta < CVDDeltaEThreshold,
				}
				if result.BelowScreen {
					report.BelowScreen++
				}
				report.Results = append(report.Results, result)
			}
		}
	}
	// No color-only semantic distinction has been identified that can be
	// justified as an enforced CVD pair. The initial enforced set is empty.
	report.Enforced = 0
	return report, nil
}
