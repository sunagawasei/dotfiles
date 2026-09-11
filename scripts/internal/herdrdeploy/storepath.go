package herdrdeploy

import "strings"

const binHerdrSuffix = "/bin/herdr"

// TrimBinHerdrSuffix strips a trailing "/bin/herdr" from an absolute path,
// returning the store root (e.g. "/nix/store/xxx-herdr-0.8.0") and whether
// the suffix was present. It's used to compare a resolved herdr binary path
// (from realpath or lsof) against a derivation's outPath.
func TrimBinHerdrSuffix(path string) (root string, ok bool) {
	if !strings.HasSuffix(path, binHerdrSuffix) {
		return "", false
	}
	return strings.TrimSuffix(path, binHerdrSuffix), true
}
