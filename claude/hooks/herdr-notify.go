// herdr-notify は Claude Code の hook として動作し、herdr の通知（delivery=terminal）が
// カバーしない2つのギャップ — エラー終了と入力待ち放置 — だけを herdr の通知API
// （herdr notification show）で補います。herdr が使えない場合は旧実装
// claude-notify.go と同じ osascript 経路にフォールバックします。
// Stop / permission_prompt は herdr 純正の通知が担当するため、ここでは配線しません。
// hook 用途のため non-blocking（エラーは握りつぶして常に exit 0）。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InputData は hook から渡される stdin JSON の必要フィールドを表します。
type InputData struct {
	HookEventName    string `json:"hook_event_name"`
	NotificationType string `json:"notification_type"`
	Cwd              string `json:"cwd"`
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
}

// debugLogPath は HERDR_NOTIFY_DEBUG 有効時のログ出力先です。
const debugLogPath = "/tmp/herdr-notify.log"

// transcriptFilePath はトランスクリプトファイルのパスを解決します。
// in.TranscriptPath が空の場合は session_id + cwd から再構成します。
func transcriptFilePath(in InputData) string {
	if in.TranscriptPath != "" {
		return in.TranscriptPath
	}
	if in.SessionID == "" || in.Cwd == "" {
		return ""
	}
	// cwd の非英数字文字を '-' に置換してプロジェクトディレクトリ名を再構成
	// 例: /Users/foo/.config → -Users-foo--config
	var b strings.Builder
	for _, r := range in.Cwd {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	// configDir: 環境変数 CLAUDE_CONFIG_DIR > ~/.config/claude
	configDir := os.Getenv("CLAUDE_CONFIG_DIR")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(home, ".config", "claude")
	}
	return filepath.Join(configDir, "projects", b.String(), in.SessionID+".jsonl")
}

// readLinesReverse はファイルの末尾から最大 maxLines 行を逆順で返します。
func readLinesReverse(path string, maxLines int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := bytes.Split(bytes.TrimRight(data, "\n"), []byte("\n"))
	result := make([]string, 0, min(maxLines, len(lines)))
	for i := len(lines) - 1; i >= 0 && len(result) < maxLines; i-- {
		if len(lines[i]) > 0 {
			result = append(result, string(lines[i]))
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// titleEntry は JSONL の会話タイトルエントリを表します。
// Claude Code はバージョン/設定により以下のいずれかの形式で記録します:
//   - {"type":"ai-title","aiTitle":"..."}         AI 自動生成タイトル（現行の既定）
//   - {"type":"custom-title","customTitle":"..."}  ユーザー設定/ブランチ由来タイトル
//
// 両者は同一セッション内では相互排他的で、どちらも Claude Code が
// OSC 2 端末タイトルに設定する文字列（= WezTerm タブバー表示）と一致します。
type titleEntry struct {
	Type        string `json:"type"`
	AITitle     string `json:"aiTitle"`
	CustomTitle string `json:"customTitle"`
}

// extractConversationTitle はトランスクリプトから最新の会話タイトルを返します。
// ai-title / custom-title のどちらの形式にも対応し、末尾（最新）を優先します。
// タイトルは毎ターン末尾付近へ再出力されますが、巨大な単一ターンで
// 末尾から押し出される稀なケースに備え全行を走査します。
func extractConversationTitle(path string) string {
	lines := readLinesReverse(path, 1<<30)
	for _, raw := range lines {
		var entry titleEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			continue
		}
		var title string
		switch entry.Type {
		case "ai-title":
			title = entry.AITitle
		case "custom-title":
			title = entry.CustomTitle
		default:
			continue
		}
		if title = strings.TrimSpace(title); title != "" {
			return title
		}
	}
	return ""
}

// transcriptLine / contentBlock は pending tool_use 検出に必要な最小フィールドです。
type transcriptLine struct {
	Type        string          `json:"type"`
	IsSidechain bool            `json:"isSidechain"`
	Message     json.RawMessage `json:"message"`
}

type contentBlock struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// hasPendingToolUse は最新の assistant ターンに許可待ちの tool_use があるかを返します。
// idle_prompt がタスク完了後の単なるアイドルか、Claude が応答を求めているかの判別に使います。
func hasPendingToolUse(path string) bool {
	lines := readLinesReverse(path, 50)
	for _, raw := range lines {
		var tl transcriptLine
		if json.Unmarshal([]byte(raw), &tl) != nil {
			continue
		}
		if tl.Type != "assistant" || tl.IsSidechain {
			continue
		}
		// 末尾から最初に見つかった assistant ターンだけを評価し、これより古いターンは
		// 遡らない。完了後の `assistant(tool_use) → user(tool_result) → assistant(text)`
		// で古い tool_use を拾い、誤って true（＝完了後アイドルを通知）にしないため。
		var msg struct {
			Content json.RawMessage `json:"content"`
		}
		if json.Unmarshal(tl.Message, &msg) != nil {
			return false
		}
		var blocks []contentBlock
		if json.Unmarshal(msg.Content, &blocks) != nil {
			return false
		}
		for _, b := range blocks {
			if b.Type == "tool_use" && b.Name != "" {
				return true
			}
		}
		return false
	}
	return false
}

// truncate は s を max rune 数に制限し、超過時は末尾に … を付けます。
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// escapeAppleScript は AppleScript の文字列リテラルに安全に埋め込めるよう
// バックスラッシュ・ダブルクォートをエスケープします。
// 改行は & return & 連結に変換して複数行通知を実現します。
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r\n", "\" & return & \"")
	s = strings.ReplaceAll(s, "\n", "\" & return & \"")
	s = strings.ReplaceAll(s, "\r", "\" & return & \"")
	return s
}

// resolveBody は通知本文（会話タイトル、無ければ cwd のディレクトリ名）を返します。
func resolveBody(in InputData) string {
	title := ""
	if txPath := transcriptFilePath(in); txPath != "" {
		title = truncate(extractConversationTitle(txPath), 60)
	}
	if title == "" && in.Cwd != "" {
		title = filepath.Base(in.Cwd)
	}
	return title
}

// tryHerdr は herdr notification show で通知を出します。herdr が未インストール、
// 実行エラー、または応答に "shown":true が含まれない場合は false を返し、
// 呼び出し元に osascript フォールバックを促します。
func tryHerdr(title, body string) bool {
	if _, err := exec.LookPath("herdr"); err != nil {
		return false
	}
	out, err := exec.Command("herdr", "notification", "show", title, "--body", body, "--sound", "request").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), `"shown":true`)
}

// notifyViaOsascript は旧実装 claude-notify.go と同じ経路で macOS 通知を出します
// （herdr が使えない場合のフォールバック）。
func notifyViaOsascript(label, body, sound string) {
	script := fmt.Sprintf(
		`display notification "%s" with title "Claude Code" subtitle "%s" sound name "%s"`,
		escapeAppleScript(body),
		escapeAppleScript(label),
		escapeAppleScript(sound),
	)
	_ = exec.Command("osascript", "-e", script).Run()
}

// notify は herdr → osascript の2段構えで通知を配送します。
func notify(in InputData, label, sound string, debug bool) {
	title := "claude " + label
	body := resolveBody(in)

	if tryHerdr(title, body) {
		logResult(debug, "herdr ok")
		return
	}
	logResult(debug, "herdr failed→osascript fallback")
	notifyViaOsascript(label, body, sound)
}

// logRaw はデバッグ時に生の stdin JSON をログへ追記します。
func logRaw(debug bool, raw []byte) {
	if !debug {
		return
	}
	appendLog(fmt.Sprintf("stdin: %s", bytes.TrimRight(raw, "\n")))
}

// logResult はデバッグ時に配送結果をログへ追記します。
func logResult(debug bool, format string, args ...any) {
	if !debug {
		return
	}
	appendLog(fmt.Sprintf(format, args...))
}

func appendLog(line string) {
	f, err := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", line)
}

func main() {
	// 生バイトを先に全部読む（デバッグで未知フィールドを捨てないため）
	rawBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}

	debug := os.Getenv("HERDR_NOTIFY_DEBUG") != ""
	logRaw(debug, rawBytes)

	var input InputData
	// JSON デコードに失敗しても静かに終了（non-blocking hook）
	if err := json.Unmarshal(rawBytes, &input); err != nil {
		return
	}

	switch input.HookEventName {
	case "StopFailure":
		notify(input, "エラー", "Basso", debug)

	case "Notification":
		if input.NotificationType != "idle_prompt" {
			logResult(debug, "ignored event (notification_type=%s)", input.NotificationType)
			return
		}
		// pending tool_use が無ければタスク完了後の単なるアイドル。
		// herdr純正/Stop通知で完了は既に伝わっているため通知しない。
		if txPath := transcriptFilePath(input); txPath != "" && !hasPendingToolUse(txPath) {
			logResult(debug, "suppressed (idle_prompt without pending tool_use)")
			return
		}
		notify(input, "入力待ち", "Submarine", debug)

	default:
		// Stop / permission_prompt 等は herdr 純正の通知が担当するため配線しない
		logResult(debug, "ignored event (hook_event_name=%s)", input.HookEventName)
	}
}
