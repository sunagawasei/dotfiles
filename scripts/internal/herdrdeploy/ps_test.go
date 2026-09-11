package herdrdeploy

import "testing"

func TestClassifyHerdrProcesses(t *testing.T) {
	cases := []struct {
		name        string
		psOutput    string
		wantServers []Process
		wantOthers  []Process
	}{
		{
			name:     "no herdr process",
			psOutput: " 100 /bin/zsh\n 101 /usr/bin/ssh-agent\n",
		},
		{
			name:        "client and server",
			psOutput:    " 2449 herdr\n 2450 /etc/profiles/per-user/s23159/bin/herdr server\n",
			wantServers: []Process{{PID: 2450, Args: "/etc/profiles/per-user/s23159/bin/herdr server"}},
			wantOthers:  []Process{{PID: 2449, Args: "herdr"}},
		},
		{
			name:        "multiple servers",
			psOutput:    " 10 /bin/herdr server\n 20 /nix/store/x-herdr-0.8.0/bin/herdr server\n",
			wantServers: []Process{{PID: 10, Args: "/bin/herdr server"}, {PID: 20, Args: "/nix/store/x-herdr-0.8.0/bin/herdr server"}},
		},
		{
			name:       "herdr but not server",
			psOutput:   " 30 herdr attach main\n 31 herdr\n",
			wantOthers: []Process{{PID: 30, Args: "herdr attach main"}, {PID: 31, Args: "herdr"}},
		},
		{
			name:     "similarly named binary is not herdr",
			psOutput: " 40 /usr/local/bin/otherherdr server\n 41 ugrep -i herdr\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			servers, others := ClassifyHerdrProcesses(tc.psOutput)
			if !equalProcesses(servers, tc.wantServers) {
				t.Errorf("servers = %+v, want %+v", servers, tc.wantServers)
			}
			if !equalProcesses(others, tc.wantOthers) {
				t.Errorf("others = %+v, want %+v", others, tc.wantOthers)
			}
		})
	}
}

func equalProcesses(got, want []Process) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
