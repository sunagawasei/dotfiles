package verifycolors

type TokenRef string

type PairClass string

const (
	ClassEnforced   PairClass = "enforced"
	ClassReportOnly PairClass = "report-only"
	ClassWaived     PairClass = "waived"
)

type PairRole string

const (
	RoleText      PairRole = "text"
	RoleBorder    PairRole = "border"
	RoleIndicator PairRole = "indicator"
	RoleSurface   PairRole = "surface"
)

type RenderProfile string

const (
	ProfileTruecolor   RenderProfile = "truecolor"
	ProfileCtermFGOnly RenderProfile = "cterm-fg-only"
	ProfileCtermFGBG   RenderProfile = "cterm-fg-bg"
)

type BackgroundRef struct {
	Token   TokenRef
	Ambient bool
}

func AmbientBackground() BackgroundRef {
	return BackgroundRef{Ambient: true}
}

func TokenBackground(token TokenRef) BackgroundRef {
	return BackgroundRef{Token: token}
}

type PairSpec struct {
	ConsumerID string
	Foreground TokenRef
	Background BackgroundRef
	Class      PairClass
	Profiles   []RenderProfile
	Role       PairRole
	Reason     string
	Source     string
}

type CoverageNote struct {
	ID     string
	Reason string
	Source string
}

type TokenDispositionKind string

const (
	DispositionBackgroundNonText TokenDispositionKind = "background-non-text"
	DispositionUnused            TokenDispositionKind = "unused"
	DispositionOutOfScope        TokenDispositionKind = "out-of-scope"
	DispositionProjection        TokenDispositionKind = "projection"
)

type TokenDisposition struct {
	Token  TokenRef
	Kind   TokenDispositionKind
	Reason string
}

func GeneratedPairs() []PairSpec {
	pairs := make([]PairSpec, len(generatedPairs))
	copy(pairs, generatedPairs)
	return pairs
}

func GeneratedCoverageNotes() []CoverageNote {
	notes := make([]CoverageNote, len(generatedCoverageNotes))
	copy(notes, generatedCoverageNotes)
	return notes
}
