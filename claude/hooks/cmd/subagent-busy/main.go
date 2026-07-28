// subagent-busy は Claude Code の hook として動作し、実行中のサブエージェントを
// herdr の busy 表示へ反映するためのマーカーファイルを管理します。
//
// SubagentStart でセッション・エージェント単位のマーカーを作成し、
// SubagentStop で削除します。SessionStart では resume 時に残った同一セッションの
// マーカーと、24時間を超えた古いマーカーを掃除します。
//
// hook 用途のため non-blocking（エラーは stderr へ出力して常に exit 0）。
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	runDir    = "/Users/s23159/.config/claude/run"
	markerTTL = 24 * time.Hour
)

var markerComponentPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// InputData は hook から渡される stdin JSON の必要フィールドを表します。
type InputData struct {
	HookEventName string `json:"hook_event_name"`
	SessionID     string `json:"session_id"`
	AgentID       string `json:"agent_id"`
}

// reportError は hook の処理をブロックせず、診断情報だけを stderr へ出力します。
func reportError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "subagent-busy: "+format+"\n", args...)
}

// validMarkerComponent はパスへ埋め込む識別子が安全な文字だけで構成されるか検証します。
func validMarkerComponent(value string) bool {
	return markerComponentPattern.MatchString(value)
}

func markerPath(sessionID, agentID string) string {
	return filepath.Join(runDir, "subagent."+sessionID+"."+agentID)
}

// processInfo は macOS の ps から pid の親 PID とコマンドの先頭トークンを取得します。
func processInfo(pid int) (int, string, error) {
	output, err := exec.Command(
		"ps", "-o", "ppid=,comm=", "-p", strconv.Itoa(pid),
	).Output()
	if err != nil {
		return 0, "", err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return 0, "", fmt.Errorf("unexpected ps output for pid %d", pid)
	}
	parentPID, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, "", fmt.Errorf("parse parent pid for %d: %w", pid, err)
	}
	return parentPID, fields[1], nil
}

// findClaudePID は自プロセスの祖先を辿り、実行コマンドが claude の PID を返します。
func findClaudePID() (int, error) {
	visited := make(map[int]bool)
	for pid := os.Getppid(); pid > 0 && !visited[pid]; {
		visited[pid] = true
		parentPID, command, err := processInfo(pid)
		if err != nil {
			return 0, err
		}
		if filepath.Base(command) == "claude" {
			return pid, nil
		}
		pid = parentPID
	}
	return 0, nil
}

// createMarker はサブエージェントのマーカーへメイン Claude プロセスの PID を保存します。
func createMarker(sessionID, agentID string) {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		reportError("create run directory: %v", err)
		return
	}

	contents := ""
	claudePID, err := findClaudePID()
	if err != nil {
		reportError("find claude process: %v", err)
	} else if claudePID > 0 {
		contents = strconv.Itoa(claudePID)
	}
	if err := os.WriteFile(markerPath(sessionID, agentID), []byte(contents), 0o644); err != nil {
		reportError("write marker: %v", err)
	}
}

func removeMarker(sessionID, agentID string) {
	if err := os.Remove(markerPath(sessionID, agentID)); err != nil && !os.IsNotExist(err) {
		reportError("remove marker: %v", err)
	}
}

// cleanupSessionMarkers は resume 時に残った同一セッションのマーカーを削除します。
func cleanupSessionMarkers(sessionID string) {
	pattern := filepath.Join(runDir, "subagent."+sessionID+".*")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		reportError("find session markers: %v", err)
		return
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			reportError("remove session marker %s: %v", path, err)
		}
	}
}

// cleanupExpiredMarkers は mtime が24時間より古いマーカーファイルを削除します。
func cleanupExpiredMarkers(now time.Time) {
	entries, err := os.ReadDir(runDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		reportError("read run directory: %v", err)
		return
	}

	cutoff := now.Add(-markerTTL)
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "subagent.") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			reportError("stat marker %s: %v", entry.Name(), err)
			continue
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		path := filepath.Join(runDir, entry.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			reportError("remove expired marker %s: %v", path, err)
		}
	}
}

func main() {
	rawBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		reportError("read input: %v", err)
		return
	}
	var input InputData
	if err := json.Unmarshal(rawBytes, &input); err != nil {
		reportError("parse input: %v", err)
		return
	}

	switch input.HookEventName {
	case "SubagentStart":
		if !validMarkerComponent(input.SessionID) || !validMarkerComponent(input.AgentID) {
			return
		}
		createMarker(input.SessionID, input.AgentID)
	case "SubagentStop":
		if !validMarkerComponent(input.SessionID) || !validMarkerComponent(input.AgentID) {
			return
		}
		removeMarker(input.SessionID, input.AgentID)
	case "SessionStart":
		if !validMarkerComponent(input.SessionID) {
			return
		}
		cleanupSessionMarkers(input.SessionID)
		cleanupExpiredMarkers(time.Now())
	}
}
