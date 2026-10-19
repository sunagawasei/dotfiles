package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type call struct{ args []string }

type fake struct {
	agent   string
	paneID  string // pane_id returned by pane get; defaults to the requested id
	getErr  error
	sendErr error
	calls   []call
}

func (f *fake) run(bin string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, call{append([]string{bin}, args...)})
	if args[1] == "get" {
		if f.getErr != nil {
			return nil, f.getErr
		}
		id := f.paneID
		if id == "" {
			id = args[2]
		}
		return []byte(`{"result":{"pane":{"pane_id":"` + id + `","agent":"` + f.agent + `"}}}`), nil
	}
	return nil, f.sendErr
}

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

var baseEnv = map[string]string{"HERDR_ACTIVE_PANE_ID": "w1:p2", "HERDR_BIN_PATH": "/bin/herdr"}

func TestRun(t *testing.T) {
	cases := []struct {
		name  string
		agent string
		key   string
		want  []string // the send call, after "/bin/herdr pane"
	}{
		{"codex alt+j sends one-line wheel down", "codex", "alt+j", []string{"send-text", "w1:p2", "\x1b[<65;1;1M"}},
		{"codex alt+k sends one-line wheel up", "codex", "alt+k", []string{"send-text", "w1:p2", "\x1b[<64;1;1M"}},
		{"codex alt+d sends PageDown", "codex", "alt+d", []string{"send-text", "w1:p2", "\x1b[6~"}},
		{"codex alt+u sends PageUp", "codex", "alt+u", []string{"send-text", "w1:p2", "\x1b[5~"}},
		{"claude alt+j passes through", "claude", "alt+j", []string{"send-keys", "w1:p2", "alt+j"}},
		{"shell without agent alt+u passes through", "", "alt+u", []string{"send-keys", "w1:p2", "alt+u"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fake{agent: tc.agent}
			if err := run([]string{tc.key}, env(baseEnv), f.run); err != nil {
				t.Fatal(err)
			}
			if len(f.calls) != 2 {
				t.Fatalf("calls = %v", f.calls)
			}
			want := append([]string{"/bin/herdr", "pane"}, tc.want...)
			if !reflect.DeepEqual(f.calls[1].args, want) {
				t.Fatalf("send = %q, want %q", f.calls[1].args, want)
			}
		})
	}
}

func TestRunTargetsExplicitPaneOnly(t *testing.T) {
	f := &fake{agent: "codex"}
	if err := run([]string{"alt+j"}, env(baseEnv), f.run); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.calls {
		if c.args[3] != "w1:p2" {
			t.Fatalf("call %q does not target w1:p2", c.args)
		}
		for _, a := range c.args {
			if a == "--current" || a == "--focused" {
				t.Fatalf("call %q targets focus", c.args)
			}
		}
	}
}

func TestRunDoesNotSendOnFailure(t *testing.T) {
	cases := []struct {
		name string
		f    *fake
		env  map[string]string
		args []string
	}{
		{"pane get fails", &fake{getErr: errors.New("boom")}, baseEnv, []string{"alt+j"}},
		{"pane get returns another pane", &fake{agent: "codex", paneID: "w1:p9"}, baseEnv, []string{"alt+j"}},
		{"pane id env missing", &fake{agent: "codex"}, map[string]string{}, []string{"alt+j"}},
		{"unsupported key", &fake{agent: "codex"}, baseEnv, []string{"alt+x"}},
		{"no key", &fake{agent: "codex"}, baseEnv, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := run(tc.args, env(tc.env), tc.f.run); err == nil {
				t.Fatal("want error")
			}
			for _, c := range tc.f.calls {
				if strings.HasPrefix(c.args[2], "send-") {
					t.Fatalf("sent despite failure: %q", c.args)
				}
			}
		})
	}
}

func TestRunReportsSendFailure(t *testing.T) {
	f := &fake{agent: "codex", sendErr: errors.New("boom")}
	if err := run([]string{"alt+j"}, env(baseEnv), f.run); err == nil {
		t.Fatal("want error")
	}
}
