package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	threadID = "01a1c2f6-19c1-7d23-93f6-ae37aa5ee67f"
	otherID  = "01a1c2f6-0000-7d23-93f6-ffffffffffff"
)

func env(kv map[string]string) func(string) string {
	return func(k string) string { return kv[k] }
}

func noRecall(args ...string) ([]byte, error) {
	panic("recall must not be called: " + strings.Join(args, " "))
}

func exec1(t *testing.T, e map[string]string, rc runner) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run([]string{"current"}, env(e), rc, &out, &errb)
	return code, out.String(), errb.String()
}

// --- Claude Code path (recall) ---

func recallStub(exportJSON string, failOn string) runner {
	return func(args ...string) ([]byte, error) {
		if args[0] == failOn {
			return nil, &exec.ExitError{Stderr: []byte("boom")}
		}
		if args[0] == "export" {
			return []byte(exportJSON), nil
		}
		return nil, nil
	}
}

func TestClaudePath(t *testing.T) {
	claude := map[string]string{"CLAUDE_CODE_SESSION_ID": "abc"}
	exported := `{"messages":[
		{"role":"user","blockType":"text","content":"first"},
		{"role":"assistant","blockType":"text","content":"no"},
		{"role":"user","blockType":"text","content":"<system-reminder>x"},
		{"role":"user","blockType":"tool_result","content":"no"},
		{"role":"user","blockType":"text","content":"second"}]}`

	t.Run("prints the user text blocks and skips harness injections", func(t *testing.T) {
		code, out, _ := exec1(t, claude, recallStub(exported, ""))
		if code != exitOK || out != "[1] first\n\n[2] second\n\n" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})
	t.Run("recall not installed exits 3", func(t *testing.T) {
		rc := func(...string) ([]byte, error) { return nil, exec.ErrNotFound }
		if code, _, _ := exec1(t, claude, rc); code != exitNoRecall {
			t.Fatalf("code=%d", code)
		}
	})
	t.Run("import failure exits 4", func(t *testing.T) {
		if code, _, _ := exec1(t, claude, recallStub(exported, "import")); code != exitFailed {
			t.Fatalf("code=%d", code)
		}
	})
	t.Run("export without messages field exits 4", func(t *testing.T) {
		if code, _, _ := exec1(t, claude, recallStub(`{}`, "")); code != exitFailed {
			t.Fatalf("code=%d", code)
		}
	})
	t.Run("session without utterances exits 4 with the recall hint", func(t *testing.T) {
		code, _, errs := exec1(t, claude, recallStub(`{"messages":[]}`, ""))
		if code != exitFailed || !strings.Contains(errs, "recall import") {
			t.Fatalf("code=%d err=%q", code, errs)
		}
	})
	t.Run("a wrong argument exits 2", func(t *testing.T) {
		var out, errb bytes.Buffer
		if code := run([]string{"past"}, env(claude), noRecall, &out, &errb); code != exitUsage {
			t.Fatalf("code=%d", code)
		}
	})
}

// --- session selection ---

func TestSessionSelection(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want int
	}{
		{"no id at all exits 5", map[string]string{}, exitNoSession},
		{"Claude and Codex ids together are ambiguous and exit 5", map[string]string{"CLAUDE_CODE_SESSION_ID": "a", "CODEX_THREAD_ID": threadID}, exitNoSession},
		{"Claude and CODEX_SESSION_ID together are ambiguous and exit 5", map[string]string{"CLAUDE_CODE_SESSION_ID": "a", "CODEX_SESSION_ID": threadID}, exitNoSession},
		{"a Codex id with path characters exits 5", map[string]string{"CODEX_THREAD_ID": "../x"}, exitNoSession},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, _, _ := exec1(t, c.env, noRecall)
			if code != c.want {
				t.Fatalf("code=%d want %d", code, c.want)
			}
		})
	}
}

// --- Codex path (rollout files) ---

type rec = map[string]any

func line(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func meta(id, sessionID, ts string, extra rec) rec {
	p := rec{"id": id, "session_id": sessionID, "timestamp": ts}
	maps.Copy(p, extra)
	return rec{"type": "session_meta", "timestamp": ts, "payload": p}
}

// user builds a response_item user message. kinds are content_item_kinds;
// nil omits the passthrough entirely.
func user(id string, at float64, kinds []string, texts ...string) rec {
	var parts []rec
	for _, s := range texts {
		parts = append(parts, rec{"type": "input_text", "text": s})
	}
	p := rec{"type": "message", "role": "user", "id": id, "content": parts}
	if kinds != nil {
		p["internal_chat_message_metadata_passthrough"] = rec{"create_time": at, "content_item_kinds": kinds}
	}
	return rec{"type": "response_item", "timestamp": "2026-11-10T01:00:00.000Z", "payload": p}
}

func typed(id string, at float64, text string) rec {
	return user(id, at, []string{"user.text"}, text)
}

func assistant(text string) rec {
	return rec{"type": "response_item", "timestamp": "2026-11-10T01:00:00.000Z", "payload": rec{
		"type": "message", "role": "assistant", "content": []rec{{"type": "output_text", "text": text}}}}
}

func writeRollout(t *testing.T, home, rel string, lines []string, trailing string) string {
	t.Helper()
	p := filepath.Join(home, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"+trailing), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func codexEnv(home string) map[string]string {
	return map[string]string{"CODEX_HOME": home, "CODEX_THREAD_ID": threadID, "CODEX_SESSION_ID": threadID}
}

func rolloutName(id string) string {
	return "sessions/2026/11/10/rollout-2026-11-10T09-57-50-" + id + ".jsonl"
}

const (
	t0 = 1794260000.0 // create_time base; 2026-11-09T21:33:20Z
)

func TestCodexPath(t *testing.T) {
	startTS := "2026-11-10T00:57:50.865Z"

	t.Run("extracts typed utterances and drops harness injections, assistant and tool records", func(t *testing.T) {
		home := t.TempDir()
		hugeReplay := strings.Repeat("x", 300_000) // a line far beyond bufio.Scanner's 64KB
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, user("m0", t0, []string{"agents_md.instructions", "environments.environment_context"}, "# AGENTS.md instructions for /x", "<environment_context>")),
			line(t, typed("m1", t0+1, "最初の発言")),
			line(t, assistant("返答")),
			line(t, user("m2", t0+2, []string{"skills.selected_skill_instructions"}, "<skill>\n<name>handoff</name>")),
			line(t, rec{"type": "compacted", "payload": rec{"replacement_history": []rec{{"type": "message", "role": "user", "content": []rec{{"type": "input_text", "text": hugeReplay}}}}}}),
			line(t, rec{"type": "response_item", "payload": rec{"type": "function_call_output", "output": "tool says hi"}}),
			line(t, typed("m3", t0+3, "二つ目")),
			line(t, rec{"type": "event_msg", "payload": rec{"type": "item_completed", "item": rec{"type": "UserMessage"}}}),
		}, "")
		code, out, errs := exec1(t, codexEnv(home), noRecall)
		if code != exitOK || out != "[1] 最初の発言\n\n[2] 二つ目\n\n" {
			t.Fatalf("code=%d out=%q err=%q", code, out, errs)
		}
	})

	t.Run("keeps repeated identical text sent as different messages", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "はい")),
			line(t, typed("b", t0+2, "はい")),
		}, "")
		if _, out, _ := exec1(t, codexEnv(home), noRecall); out != "[1] はい\n\n[2] はい\n\n" {
			t.Fatalf("out=%q", out)
		}
	})

	t.Run("a Codex id with only CODEX_SESSION_ID set still resolves", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "only session id")),
		}, "")
		e := map[string]string{"CODEX_HOME": home, "CODEX_SESSION_ID": threadID}
		if code, out, _ := exec1(t, e, noRecall); code != exitOK || out != "[1] only session id\n\n" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})

	t.Run("a subagent file sharing session_id is not mixed in", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "mine")),
		}, "")
		// Another thread whose session_id is ours (a child) and whose name also contains our id.
		writeRollout(t, home, "sessions/2026/11/10/rollout-2026-11-10T10-00-00-"+otherID+"_"+threadID+".jsonl", []string{
			line(t, meta(otherID, threadID, startTS, rec{"parent_thread_id": threadID})),
			line(t, typed("c1", t0+5, "child words")),
		}, "")
		if _, out, _ := exec1(t, codexEnv(home), noRecall); out != "[1] mine\n\n" {
			t.Fatalf("out=%q", out)
		}
	})

	t.Run("other files of the same thread are merged in time order without duplicates", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "one")),
			line(t, typed("c", t0+30, "three")),
		}, "")
		writeRollout(t, home, "sessions/2026/11/10/rollout-2026-11-10T18-13-18-"+threadID+"_01a1c4bb.jsonl", []string{
			line(t, meta(threadID, threadID, "2026-11-10T09:13:18.217Z", nil)),
			line(t, typed("b", t0+10, "two")),
			line(t, typed("c", t0+30, "three")), // same message id replayed
		}, "")
		writeRollout(t, home, "archived_sessions/rollout-2026-11-09T01-00-00-"+threadID+".jsonl", []string{
			line(t, meta(threadID, threadID, "2026-11-09T01:00:00.000Z", nil)),
			line(t, typed("z", t0-100, "zero")),
		}, "")
		if _, out, _ := exec1(t, codexEnv(home), noRecall); out != "[1] zero\n\n[2] one\n\n[3] two\n\n[4] three\n\n" {
			t.Fatalf("out=%q", out)
		}
	})

	t.Run("a fork drops the history it replays from its parent", func(t *testing.T) {
		home := t.TempDir()
		forkTS := "2026-11-09T21:40:00.000Z" // after t0+1, before t0+1000
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, forkTS, rec{"forked_from_id": otherID, "parent_thread_id": otherID})),
			line(t, meta(otherID, otherID, "2026-11-09T21:30:00.000Z", nil)),
			line(t, typed("p1", t0+1, "parent said")),
			line(t, typed("f1", t0+1000, "fork said")),
		}, "")
		if code, out, _ := exec1(t, codexEnv(home), noRecall); code != exitOK || out != "[1] fork said\n\n" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})

	t.Run("without content_item_kinds the injection prefixes decide", func(t *testing.T) {
		home := t.TempDir()
		m := user("a", 0, nil, "実発言")
		inj := user("b", 0, nil, "<environment_context>\n<cwd>/x</cwd>")
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)), line(t, inj), line(t, m),
		}, "")
		if _, out, _ := exec1(t, codexEnv(home), noRecall); out != "[1] 実発言\n\n" {
			t.Fatalf("out=%q", out)
		}
	})

	t.Run("an unfinished last line from the live session is tolerated", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "kept")),
		}, `{"type":"response_item","payload":{"type":"mess`)
		if code, out, _ := exec1(t, codexEnv(home), noRecall); code != exitOK || out != "[1] kept\n\n" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})

	t.Run("a damaged line in the middle exits 4 without echoing content", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			`{"type":"response_item","payload": SECRET-LOOKING-TEXT`,
			line(t, typed("a", t0+1, "after")),
		}, "")
		code, out, errs := exec1(t, codexEnv(home), noRecall)
		if code != exitFailed || out != "" || strings.Contains(errs, "SECRET") {
			t.Fatalf("code=%d out=%q err=%q", code, out, errs)
		}
	})

	t.Run("no rollout for the thread exits 4", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(otherID), []string{
			line(t, meta(otherID, otherID, startTS, nil)),
			line(t, typed("a", t0+1, "someone else")),
		}, "")
		code, out, _ := exec1(t, codexEnv(home), noRecall)
		if code != exitFailed || out != "" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})

	t.Run("a name match whose session_meta holds another id is not used", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(otherID, otherID, startTS, nil)),
			line(t, typed("a", t0+1, "wrong thread")),
		}, "")
		if code, out, _ := exec1(t, codexEnv(home), noRecall); code != exitFailed || out != "" {
			t.Fatalf("code=%d out=%q", code, out)
		}
	})

	t.Run("a thread with only injections exits 4", func(t *testing.T) {
		home := t.TempDir()
		writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, user("m0", t0, []string{"agents_md.instructions"}, "# AGENTS.md instructions")),
		}, "")
		if code, _, errs := exec1(t, codexEnv(home), noRecall); code != exitFailed || !strings.Contains(errs, "no user utterance") {
			t.Fatalf("code=%d err=%q", code, errs)
		}
	})

	t.Run("an unreadable rollout exits 4", func(t *testing.T) {
		home := t.TempDir()
		p := writeRollout(t, home, rolloutName(threadID), []string{
			line(t, meta(threadID, threadID, startTS, nil)),
			line(t, typed("a", t0+1, "x")),
		}, "")
		if err := os.Chmod(p, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(p, 0o644) })
		if f, err := os.Open(p); err == nil { // running as root
			f.Close()
			t.Skip("file is readable despite mode 0")
		}
		if code, _, _ := exec1(t, codexEnv(home), noRecall); code != exitFailed {
			t.Fatalf("code=%d", code)
		}
	})

	t.Run("an empty CODEX_HOME exits 4, not 5", func(t *testing.T) {
		if code, _, _ := exec1(t, codexEnv(t.TempDir()), noRecall); code != exitFailed {
			t.Fatalf("code=%d", code)
		}
	})
}

func TestFailMapsNotFound(t *testing.T) {
	var b bytes.Buffer
	if code := fail(&b, errors.Join(exec.ErrNotFound)); code != exitNoRecall {
		t.Fatalf("code=%d", code)
	}
}
