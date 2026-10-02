package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

var sec = secrets{home: "/Users/alice", uid: "501234567", user: "alice"}

func TestPublicize(t *testing.T) {
	t.Run("ホームのパスを ~ に置換する", func(t *testing.T) {
		out, err := publicize(decode(t, `{"a":"/Users/alice/.config/x","b":["Read(//Users/alice/.cache/**)"]}`), sec)
		if err != nil {
			t.Fatal(err)
		}
		got := string(out)
		if !strings.Contains(got, `"~/.config/x"`) || !strings.Contains(got, `Read(~/.cache/**)`) {
			t.Errorf("置換されていない: %s", got)
		}
	})

	t.Run("除外語を含む配列要素を落とす", func(t *testing.T) {
		out, err := publicize(decode(t, `{"p":["Bash(ls:*)","Bash(cycloud:*)","Bash(pup logs:*)","Bash(gws:*)","/tmp/x/poc/y"]}`), sec)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "cycloud") || strings.Contains(string(out), "pup") || strings.Contains(string(out), "gws") {
			t.Errorf("除外語が残っている: %s", out)
		}
		if !strings.Contains(string(out), "Bash(ls:*)") {
			t.Errorf("無関係な要素まで落ちている: %s", out)
		}
	})

	t.Run("uid を含む配列要素を落とす", func(t *testing.T) {
		out, err := publicize(decode(t, `{"w":["/tmp","/private/tmp/agent-501234567"]}`), sec)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "501234567") || !strings.Contains(string(out), `"/tmp"`) {
			t.Errorf("uid 要素の扱いが違う: %s", out)
		}
	})

	t.Run("autoMode.environment を落とし allow は残す", func(t *testing.T) {
		out, err := publicize(decode(t, `{"autoMode":{"allow":["x"],"environment":["bucket"]}}`), sec)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "environment") || !strings.Contains(string(out), `"allow"`) {
			t.Errorf("autoMode の扱いが違う: %s", out)
		}
	})

	t.Run("配列の外に個人値が残るとエラーにする", func(t *testing.T) {
		if _, err := publicize(decode(t, `{"statusLine":"run-501234567"}`), sec); err == nil {
			t.Error("uid が残ったのにエラーにならない")
		}
		if _, err := publicize(decode(t, `{"k":"cycloud"}`), sec); err == nil {
			t.Error("除外語が配列の外に残ったのにエラーにならない")
		}
	})
}
