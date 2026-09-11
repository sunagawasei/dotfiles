package herdrdeploy

import (
	"regexp"
	"slices"
	"strings"
)

// TxtStatus is the outcome of looking for the herdr binary's txt (loaded
// code) entry in `lsof -p <pid>` output.
type TxtStatus int

const (
	TxtNotFound TxtStatus = iota
	TxtFound
	TxtUnreadable
)

func (s TxtStatus) String() string {
	switch s {
	case TxtFound:
		return "found"
	case TxtUnreadable:
		return "unreadable"
	default:
		return "not-found"
	}
}

var executableStorePath = regexp.MustCompile(`^/nix/store/[^/\s]+/bin/herdr$`)

// FindExecutableTxtPath scans `lsof -p <pid>` output for the txt-segment
// line that names the herdr binary itself, as opposed to a shared library.
// lsof's column widths aren't fixed, so a txt line is identified by "txt"
// appearing as its own whitespace-delimited token rather than by a fixed
// column index.
//
// If a txt line's NAME field isn't a well-formed absolute path (a stat/
// readlink permission error, or a "(deleted)" marker), that's reported as
// TxtUnreadable rather than silently treated as "no match": lsof racing
// against process exit or file permissions must not be mistaken for a
// clean absence of the herdr binary.
func FindExecutableTxtPath(lsofOutput string) (storePath string, status TxtStatus, rawLine string) {
	var unreadableLine string
	for line := range strings.SplitSeq(lsofOutput, "\n") {
		fields := strings.Fields(line)
		if !slices.Contains(fields, "txt") {
			continue
		}
		name := fields[len(fields)-1]
		if executableStorePath.MatchString(name) {
			return name, TxtFound, line
		}
		if strings.HasPrefix(name, "/") {
			// Some other loaded file (a shared library, /usr/lib/dyld, ...):
			// a real txt entry, just not the herdr binary itself.
			continue
		}
		if unreadableLine == "" {
			unreadableLine = line
		}
	}
	if unreadableLine != "" {
		return "", TxtUnreadable, unreadableLine
	}
	return "", TxtNotFound, ""
}
