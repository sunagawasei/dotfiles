package herdrdeploy

import "testing"

func TestLockedRev(t *testing.T) {
	cases := []struct {
		name    string
		data    string
		node    string
		want    string
		wantErr bool
	}{
		{
			name: "root input name equals node ID",
			data: `{"root":"root","nodes":{
				"root":{"inputs":{"herdr":"herdr"}},
				"herdr":{"locked":{"rev":"346411fa21afd297f5ed3b3fa56f9e3fbf7654b7"}}
			}}`,
			node: "herdr",
			want: "346411fa21afd297f5ed3b3fa56f9e3fbf7654b7",
		},
		{
			name: "root input resolves to a differently-named node (dedup case, e.g. nixpkgs -> nixpkgs_3)",
			data: `{"root":"root","nodes":{
				"root":{"inputs":{"herdr":"herdr_2"}},
				"herdr":{"locked":{"rev":"wrong-node-same-name-must-not-be-used"}},
				"herdr_2":{"locked":{"rev":"346411fa21afd297f5ed3b3fa56f9e3fbf7654b7"}}
			}}`,
			node: "herdr",
			want: "346411fa21afd297f5ed3b3fa56f9e3fbf7654b7",
		},
		{
			name:    "root input missing",
			data:    `{"root":"root","nodes":{"root":{"inputs":{"nixpkgs":"nixpkgs_3"}}}}`,
			node:    "herdr",
			wantErr: true,
		},
		{
			name:    "referenced node missing",
			data:    `{"root":"root","nodes":{"root":{"inputs":{"herdr":"herdr_2"}}}}`,
			node:    "herdr",
			wantErr: true,
		},
		{
			name:    "root node missing",
			data:    `{"root":"root","nodes":{"herdr":{"locked":{"rev":"abc"}}}}`,
			node:    "herdr",
			wantErr: true,
		},
		{
			name: "root input is a follows chain (array), not a plain node reference",
			data: `{"root":"root","nodes":{
				"root":{"inputs":{"herdr":["home-manager","herdr"]}},
				"herdr":{"locked":{"rev":"abc"}}
			}}`,
			node:    "herdr",
			wantErr: true,
		},
		{
			name:    "rev missing on resolved node",
			data:    `{"root":"root","nodes":{"root":{"inputs":{"herdr":"herdr"}},"herdr":{"locked":{"type":"github"}}}}`,
			node:    "herdr",
			wantErr: true,
		},
		{
			name: "root field absent, defaults to node \"root\"",
			data: `{"nodes":{"root":{"inputs":{"herdr":"herdr"}},"herdr":{"locked":{"rev":"abc"}}}}`,
			node: "herdr",
			want: "abc",
		},
		{
			name:    "invalid json",
			data:    `not json`,
			node:    "herdr",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LockedRev([]byte(tc.data), tc.node)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("LockedRev(%q) = %q, want error", tc.name, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("LockedRev(%q) unexpected error: %v", tc.name, err)
			}
			if got != tc.want {
				t.Fatalf("LockedRev(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}
