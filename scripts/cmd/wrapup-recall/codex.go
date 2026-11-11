package main

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Codex keeps each thread in rollout JSONL files under $CODEX_HOME:
//
//	sessions/YYYY/MM/DD/rollout-<ts>-<thread id>.jsonl
//	sessions/YYYY/MM/DD/rollout-<ts>-<thread id>_<suffix>.jsonl   same thread, another file
//	archived_sessions/rollout-<ts>-<thread id>.jsonl
//
// The first line is a session_meta whose payload.id is the thread id. A
// subagent's file carries its parent's id in payload.session_id, so the file is
// matched on payload.id and never on session_id.

var codexIDPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z-]*$`)

// Fallback for a user message without content_item_kinds. With the kinds
// present only "user.text" counts, so these prefixes matter only for a schema
// that lacks them.
var codexSkipPrefixes = []string{
	"# AGENTS.md instructions", "<environment_context", "<skill>", "<guardian_tool",
	"<turn_aborted", "[agmsg:",
}

type rolloutLine struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type sessionMeta struct {
	ID           string `json:"id"`
	Timestamp    string `json:"timestamp"`
	ForkedFromID string `json:"forked_from_id"`
}

type responseHead struct {
	Type string `json:"type"`
	Role string `json:"role"`
}

type contentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type passthrough struct {
	CreateTime *float64 `json:"create_time"`
	Kinds      []string `json:"content_item_kinds"`
}

type userMessage struct {
	ID      string        `json:"id"`
	Content []contentPart `json:"content"`
	Meta    *passthrough  `json:"internal_chat_message_metadata_passthrough"`
}

type utterance struct {
	at   float64
	seq  int
	id   string
	text string
}

func codexHome(getenv func(string) string) (string, error) {
	if h := getenv("CODEX_HOME"); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

// findRollouts lists the files whose name carries the thread id. The name is
// only a prefilter: readRollout confirms the id from the session_meta line.
func findRollouts(home, id string) ([]string, error) {
	var found []string
	for _, sub := range []string{"sessions", "archived_sessions"} {
		root := filepath.Join(home, sub)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == root && errors.Is(err, fs.ErrNotExist) {
					return filepath.SkipAll
				}
				return err
			}
			name := d.Name()
			if !d.IsDir() && strings.HasPrefix(name, "rollout-") && strings.HasSuffix(name, ".jsonl") && strings.Contains(name, id) {
				found = append(found, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", sub, err)
		}
	}
	return found, nil
}

func isCodexUtterance(m userMessage, text string) bool {
	if m.Meta != nil && len(m.Meta.Kinds) > 0 {
		return !slices.ContainsFunc(m.Meta.Kinds, func(k string) bool { return k != "user.text" })
	}
	return !hasAnyPrefix(text, codexSkipPrefixes)
}

func messageText(parts []contentPart) string {
	var texts []string
	for _, p := range parts {
		if p.Type == "input_text" && strings.TrimSpace(p.Text) != "" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func recordTime(l rolloutLine) float64 {
	t, err := time.Parse(time.RFC3339Nano, l.Timestamp)
	if err != nil {
		return 0
	}
	return float64(t.UnixNano()) / 1e9
}

// readRollout returns the user utterances of one file. matched is false when
// the file is not the thread's (another id, or nothing written yet). Errors
// never carry file content.
func readRollout(path, id string, seq *int) (out []utterance, matched bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	base := filepath.Base(path)

	r := bufio.NewReader(f)
	var forkCutoff float64
	forked := false
	for lineNo := 1; ; lineNo++ {
		raw, rerr := r.ReadBytes('\n')
		partial := rerr == io.EOF && len(raw) > 0
		if rerr != nil && rerr != io.EOF {
			return nil, false, fmt.Errorf("%s:%d: %w", base, lineNo, rerr)
		}
		if line := bytes.TrimSpace(raw); len(line) > 0 {
			var l rolloutLine
			if jerr := json.Unmarshal(line, &l); jerr != nil {
				// The live session may be appending: an unfinished last line is not damage.
				if partial {
					return out, matched, nil
				}
				return nil, false, fmt.Errorf("%s:%d: invalid JSON line", base, lineNo)
			}
			if lineNo == 1 {
				var meta sessionMeta
				if l.Type != "session_meta" || json.Unmarshal(l.Payload, &meta) != nil {
					return nil, false, fmt.Errorf("%s: first line is not a session_meta", base)
				}
				if meta.ID != id {
					return nil, false, nil
				}
				matched = true
				if meta.ForkedFromID != "" {
					t, terr := time.Parse(time.RFC3339Nano, meta.Timestamp)
					if terr != nil {
						return nil, false, fmt.Errorf("%s: fork without a usable timestamp", base)
					}
					forked, forkCutoff = true, float64(t.UnixNano())/1e9
				}
			} else if l.Type == "response_item" {
				u, ok, perr := userUtteranceOf(l, forked, forkCutoff)
				if perr != nil {
					return nil, false, fmt.Errorf("%s:%d: %w", base, lineNo, perr)
				}
				if ok {
					*seq++
					u.seq = *seq
					out = append(out, u)
				}
			}
		}
		if rerr == io.EOF {
			return out, matched, nil
		}
	}
}

func userUtteranceOf(l rolloutLine, forked bool, forkCutoff float64) (utterance, bool, error) {
	var head responseHead
	if err := json.Unmarshal(l.Payload, &head); err != nil {
		return utterance{}, false, errors.New("unexpected response_item payload")
	}
	if head.Type != "message" || head.Role != "user" {
		return utterance{}, false, nil
	}
	var m userMessage
	if err := json.Unmarshal(l.Payload, &m); err != nil {
		return utterance{}, false, errors.New("unexpected user message shape")
	}
	text := messageText(m.Content)
	if text == "" || !isCodexUtterance(m, text) {
		return utterance{}, false, nil
	}
	at := recordTime(l)
	if m.Meta != nil && m.Meta.CreateTime != nil {
		at = *m.Meta.CreateTime
	}
	if forked {
		// A fork replays its parent's history under the fork's own timestamps;
		// only the original create_time tells whose turn it was.
		if m.Meta == nil || m.Meta.CreateTime == nil || *m.Meta.CreateTime < forkCutoff {
			return utterance{}, false, nil
		}
	}
	return utterance{at: at, id: m.ID, text: text}, true, nil
}

func codexUtterances(home, id string) ([]string, error) {
	paths, err := findRollouts(home, id)
	if err != nil {
		return nil, err
	}
	var all []utterance
	seq := 0
	matchedFiles := 0
	for _, p := range paths {
		us, matched, err := readRollout(p, id, &seq)
		if err != nil {
			return nil, err
		}
		if matched {
			matchedFiles++
			all = append(all, us...)
		}
	}
	if matchedFiles == 0 {
		return nil, fmt.Errorf("no Codex rollout file for thread %s under %s", id, home)
	}
	slices.SortStableFunc(all, func(a, b utterance) int {
		if c := cmp.Compare(a.at, b.at); c != 0 {
			return c
		}
		return cmp.Compare(a.seq, b.seq)
	})
	seen := map[string]bool{}
	var out []string
	for _, u := range all {
		if u.id != "" {
			if seen[u.id] {
				continue
			}
			seen[u.id] = true
		}
		out = append(out, u.text)
	}
	return out, nil
}
