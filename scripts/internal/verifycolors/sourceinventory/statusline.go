package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	statuslineColorPattern   = regexp.MustCompile(`^\s*(C_[A-Z]+)=.*\{\{rgb:([a-z_]+\.[a-z_]+)\}\}`)
	statuslineSegmentPattern = regexp.MustCompile(`row[12]\+=\("\{\{([a-z_]+\.[a-z_]+)\}\}\|\$\{(C_[A-Z]+)\}`)
	statuslineTrackPattern   = regexp.MustCompile(`row[12]\+=\("\{\{([a-z_]+\.[a-z_]+)\}\}\|.*\$\{(C_[A-Z]+TRACK)\}`)
)

type statuslineColor struct {
	token  verifycolors.TokenRef
	source string
}

type statuslineSegment struct {
	id         string
	colorVar   string
	background verifycolors.TokenRef
	source     string
}

func extractStatusline(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	colors := make(map[string][]statuslineColor)
	var segments []statuslineSegment
	var tracks []statuslineSegment
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		if match := statuslineColorPattern.FindStringSubmatch(line); match != nil {
			colors[match[1]] = append(colors[match[1]], statuslineColor{
				token:  verifycolors.TokenRef(match[2]),
				source: source,
			})
		}
		if match := statuslineSegmentPattern.FindStringSubmatch(line); match != nil {
			id := statuslineSegmentID(match[2], line)
			segments = append(segments, statuslineSegment{
				id:         id,
				colorVar:   match[2],
				background: verifycolors.TokenRef(match[1]),
				source:     source,
			})
		}
		if match := statuslineTrackPattern.FindStringSubmatch(line); match != nil {
			baseColorVar := strings.TrimSuffix(match[2], "TRACK")
			tracks = append(tracks, statuslineSegment{
				id:         statuslineSegmentID(baseColorVar, line),
				colorVar:   match[2],
				background: verifycolors.TokenRef(match[1]),
				source:     source,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for _, segment := range segments {
		segmentColors := colors[segment.colorVar]
		if len(segmentColors) == 0 {
			return fmt.Errorf("%s: no foreground token found for %s", segment.source, segment.colorVar)
		}
		stateNames := statuslineStateNames(len(segmentColors))
		for i, color := range segmentColors {
			consumerID := "statusline." + segment.id
			if stateNames[i] != "" {
				consumerID += "." + stateNames[i]
			}
			result.addPair(defaultTextPair(
				consumerID,
				color.token,
				verifycolors.TokenBackground(segment.background),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				segment.source,
			))
		}
		result.addPair(surfacePair(
			"statusline."+segment.id+".surface",
			segment.background,
			segment.source,
		))
	}
	for _, track := range tracks {
		trackColors := colors[track.colorVar]
		if len(trackColors) == 0 {
			return fmt.Errorf("%s: no foreground token found for %s", track.source, track.colorVar)
		}
		for _, color := range trackColors {
			result.addPair(verifycolors.PairSpec{
				ConsumerID: "statusline." + track.id + ".track",
				Foreground: color.token,
				Background: verifycolors.TokenBackground(track.background),
				Class:      verifycolors.ClassReportOnly,
				Profiles:   []verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				Role:       verifycolors.RoleSurface,
				Source:     track.source,
			})
		}
	}
	return nil
}

func statuslineSegmentID(colorVar, line string) string {
	switch colorVar {
	case "C_MODEL":
		return "model"
	case "C_PCT":
		return "context"
	case "C_RATE":
		return "rate"
	case "C_WEEK":
		return "week"
	case "C_DIR":
		return "directory"
	case "C_GIT":
		return "git"
	case "C_BUSY":
		if strings.Contains(line, "CODEX_BUSY") {
			return "busy.codex"
		}
		return "busy.background"
	default:
		return strings.ToLower(strings.TrimPrefix(colorVar, "C_"))
	}
}

func statuslineStateNames(count int) []string {
	if count == 3 {
		return []string{"critical", "warning", "safe"}
	}
	names := make([]string, count)
	return names
}
