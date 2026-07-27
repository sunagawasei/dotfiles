package sourceinventory

import (
	"fmt"
	"sort"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

type Result struct {
	Pairs         []verifycolors.PairSpec
	CoverageNotes []verifycolors.CoverageNote
}

func (r *Result) addPair(pair verifycolors.PairSpec) {
	r.Pairs = append(r.Pairs, pair)
}

func (r *Result) addCoverageNote(note verifycolors.CoverageNote) {
	r.CoverageNotes = append(r.CoverageNotes, note)
}

func (r *Result) normalize() error {
	sort.Slice(r.Pairs, func(i, j int) bool {
		return r.Pairs[i].ConsumerID < r.Pairs[j].ConsumerID
	})
	for i := 1; i < len(r.Pairs); i++ {
		if r.Pairs[i-1].ConsumerID == r.Pairs[i].ConsumerID {
			return fmt.Errorf("duplicate consumer ID %q", r.Pairs[i].ConsumerID)
		}
	}
	sort.Slice(r.CoverageNotes, func(i, j int) bool {
		return r.CoverageNotes[i].ID < r.CoverageNotes[j].ID
	})
	return nil
}

func defaultTextPair(
	consumerID string,
	foreground verifycolors.TokenRef,
	background verifycolors.BackgroundRef,
	profiles []verifycolors.RenderProfile,
	source string,
) verifycolors.PairSpec {
	class := verifycolors.ClassReportOnly
	if !background.Ambient {
		class = verifycolors.ClassEnforced
	}
	return verifycolors.PairSpec{
		ConsumerID: consumerID,
		Foreground: foreground,
		Background: background,
		Class:      class,
		Profiles:   profiles,
		Role:       verifycolors.RoleText,
		Source:     source,
	}
}

func surfacePair(
	consumerID string,
	surface verifycolors.TokenRef,
	source string,
) verifycolors.PairSpec {
	return verifycolors.PairSpec{
		ConsumerID: consumerID,
		Foreground: surface,
		Background: verifycolors.AmbientBackground(),
		Class:      verifycolors.ClassReportOnly,
		Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		Role:       verifycolors.RoleSurface,
		Source:     source,
	}
}
