package herdrdeploy

import (
	"strings"
	"testing"
)

func TestFindExecutableTxtPath(t *testing.T) {
	cases := []struct {
		name           string
		lsofOutput     string
		wantStorePath  string
		wantStatus     TxtStatus
		wantRawLineHas string // substring the raw line must contain, if any
	}{
		{
			name:       "no txt line at all",
			lsofOutput: "herdr 2450 s23159 cwd DIR 1,17 96 2 /\n",
			wantStatus: TxtNotFound,
		},
		{
			name: "multiple txt lines, herdr binary among shared libs",
			lsofOutput: "herdr   2450 s23159  txt   REG   1,17  18323136  5131237 /nix/store/69wn2w9ih51klij3zafhjh8rqfj7x2q0-herdr-0.8.0/bin/herdr\n" +
				"herdr   2450 s23159  txt   REG   1,17     92352  2315450 /nix/store/wjb6smwmv8cynnr776j01kjjzdks1863-libiconv-113/lib/libiconv.2.dylib\n" +
				"herdr   2450 s23159  txt   REG   1,13   2562000        1 /usr/lib/dyld\n",
			wantStorePath: "/nix/store/69wn2w9ih51klij3zafhjh8rqfj7x2q0-herdr-0.8.0/bin/herdr",
			wantStatus:    TxtFound,
		},
		{
			name: "only shared libs, no herdr binary line",
			lsofOutput: "herdr   2450 s23159  txt   REG   1,17     92352  2315450 /nix/store/wjb6smwmv8cynnr776j01kjjzdks1863-libiconv-113/lib/libiconv.2.dylib\n" +
				"herdr   2450 s23159  txt   REG   1,13   2562000        1 /usr/lib/dyld\n",
			wantStatus: TxtNotFound,
		},
		{
			name: "permission error line mixed in, herdr binary still resolvable",
			lsofOutput: "herdr   2450 s23159  txt   REG   1,17  18323136  5131237 /nix/store/69wn2w9ih51klij3zafhjh8rqfj7x2q0-herdr-0.8.0/bin/herdr\n" +
				"herdr   2450 s23159  txt    ???                             (readlink: Permission denied)\n",
			wantStorePath: "/nix/store/69wn2w9ih51klij3zafhjh8rqfj7x2q0-herdr-0.8.0/bin/herdr",
			wantStatus:    TxtFound,
		},
		{
			name:           "only a permission error line, herdr binary unresolvable",
			lsofOutput:     "herdr   2450 s23159  txt    ???                             (readlink: Permission denied)\n",
			wantStatus:     TxtUnreadable,
			wantRawLineHas: "Permission denied",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, status, raw := FindExecutableTxtPath(tc.lsofOutput)
			if path != tc.wantStorePath {
				t.Errorf("storePath = %q, want %q", path, tc.wantStorePath)
			}
			if status != tc.wantStatus {
				t.Errorf("status = %v, want %v", status, tc.wantStatus)
			}
			if tc.wantRawLineHas != "" && !strings.Contains(raw, tc.wantRawLineHas) {
				t.Errorf("rawLine = %q, want substring %q", raw, tc.wantRawLineHas)
			}
		})
	}
}
