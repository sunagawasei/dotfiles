// claude-notify は Claude Code の hook として動作し、
// stdin の JSON を解析して macOS のデスクトップ通知を出します。
//
// 通知は「状態」と「会話タイトル」の2点だけを表示します:
//   - サブタイトル: 状態（完了 / 入力待ち / 許可待ち / エラー）
//   - 本文        : 会話タイトル（= WezTerm タブバー表示。未生成時はプロジェクト名）
//
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
	Message          string `json:"message"`
	Cwd              string `json:"cwd"`
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
	StopHookActive   bool   `json:"stop_hook_active"`
}

// notification は組み立てた通知内容を表します。
type notification struct {
	subtitle string // 状態ラベル
	body     string // 会話タイトル
	sound    string
}

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

// resolve はイベント種別から通知内容を決定します。
// サブタイトルに状態、本文に会話タイトルを載せます。
func resolve(in InputData) (notification, bool) {
	txPath := transcriptFilePath(in)

	// 本文に表示する会話タイトル。WezTerm タブバーと同一の文字列を優先し、
	// 無ければ cwd のディレクトリ名にフォールバックする。
	// （新規/直後の /clear セッションはタイトル未生成のため空になりうる）
	title := ""
	if txPath != "" {
		title = truncate(extractConversationTitle(txPath), 60)
	}
	if title == "" && in.Cwd != "" {
		title = filepath.Base(in.Cwd)
	}

	switch in.HookEventName {
	case "Stop":
		// stop_hook_active=true はループ継続中の中間発火なので通知しない
		if in.StopHookActive {
			return notification{}, false
		}
		return notification{"完了", title, "Glass"}, true

	case "StopFailure":
		return notification{"エラー", title, "Basso"}, true

	case "Notification":
		switch in.NotificationType {
		case "permission_prompt":
			return notification{"許可待ち", title, "Submarine"}, true

		case "idle_prompt":
			// pending tool_use が無ければタスク完了後の単なるアイドル。
			// 完了は Stop 通知で既に伝えているため追加通知しない。
			if txPath != "" && !hasPendingToolUse(txPath) {
				return notification{}, false
			}
			return notification{"入力待ち", title, "Submarine"}, true

		default:
			// 未知の Notification でも message があれば状態として出す
			if msg := strings.TrimSpace(in.Message); msg != "" {
				return notification{"通知", title, "Submarine"}, true
			}
		}
	}
	return notification{}, false
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

func main() {
	// 生バイトを先に全部読む（デバッグで未知フィールドを捨てないため）
	rawBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}

	// デバッグ用: CLAUDE_NOTIFY_DEBUG が設定されていれば生 JSON をログに残す
	// （未知フィールドを含む実ペイロードを確認するため。通常運用では未設定）
	if os.Getenv("CLAUDE_NOTIFY_DEBUG") != "" {
		if f, ferr := os.OpenFile("/tmp/claude-notify.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
			fmt.Fprintf(f, "%s\n", rawBytes)
			f.Close()
		}
	}

	var input InputData
	// JSON デコードに失敗しても静かに終了（non-blocking hook）
	if err := json.Unmarshal(rawBytes, &input); err != nil {
		return
	}

	n, ok := resolve(input)
	if !ok {
		return
	}

	script := fmt.Sprintf(
		`display notification "%s" with title "Claude Code" subtitle "%s" sound name "%s"`,
		escapeAppleScript(n.body),
		escapeAppleScript(n.subtitle),
		escapeAppleScript(n.sound),
	)

	// エラーがあっても静かに終了（non-blocking hook）
	_ = exec.Command("osascript", "-e", script).Run()
}
