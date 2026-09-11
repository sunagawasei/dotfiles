package herdrdeploy

import "testing"

func TestTrimBinHerdrSuffix(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		wantRoot string
		wantOK   bool
	}{
		{
			name:     "well-formed store path",
			path:     "/nix/store/af1frp2ja5mjwapf0n3qk3r1c154z2g3-herdr-0.8.0/bin/herdr",
			wantRoot: "/nix/store/af1frp2ja5mjwapf0n3qk3r1c154z2g3-herdr-0.8.0",
			wantOK:   true,
		},
		{
			name:   "not a herdr binary path",
			path:   "/nix/store/af1frp2ja5mjwapf0n3qk3r1c154z2g3-herdr-0.8.0/bin/herdr-server",
			wantOK: false,
		},
		{
			name:   "already a store root, no suffix",
			path:   "/nix/store/af1frp2ja5mjwapf0n3qk3r1c154z2g3-herdr-0.8.0",
			wantOK: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, ok := TrimBinHerdrSuffix(tc.path)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && root != tc.wantRoot {
				t.Fatalf("root = %q, want %q", root, tc.wantRoot)
			}
		})
	}
}
