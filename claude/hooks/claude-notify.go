// claude-notify は Claude Code の hook として動作し、
// stdin の JSON を解析して macOS のデスクトップ通知を出します。
//
// 通知本文にトランスクリプトから抽出したコンテキストを含めることで、
// 複数セッション並行時でもどの作業の通知かを判別できます。
//
//   - Stop        : 直前のユーザープロンプト（「何を頼んだか」）を本文に表示
//   - idle_prompt : Claude の最後の発言（「何を求めているか」）を本文に表示
//   - permission  : message（許可対象）→ Claude の最後の発言 の順で表示
//
// hook 用途のため non-blocking（エラーは握りつぶして常に exit 0）。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
}

// notification は組み立てた通知内容を表します。
type notification struct {
	subtitle string
	body     string
	sound    string
}

// imageTagRe は [Image #N] / [Image: ...] マーカーを除去するための正規表現です。
var imageTagRe = regexp.MustCompile(`(?i)\[Image[^\]]*\]\s*`)

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

// lastPromptEntry は JSONL の "last-prompt" タイプエントリを表します。
type lastPromptEntry struct {
	Type       string `json:"type"`
	LastPrompt string `json:"lastPrompt"`
}

// extractLastUserPrompt はトランスクリプトから最後のユーザープロンプトを返します。
// "last-prompt" エントリを優先し、なければ user ターンを直接解析します。
func extractLastUserPrompt(path string) string {
	lines := readLinesReverse(path, 300)

	// まず "last-prompt" タイプのエントリを探す（最も信頼性が高い）
	for _, raw := range lines {
		var entry lastPromptEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			continue
		}
		if entry.Type != "last-prompt" {
			continue
		}
		text := strings.TrimSpace(entry.LastPrompt)
		if text == "" {
			continue
		}
		// [Image #N] / [Image: ...] マーカーを除去
		text = imageTagRe.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		return text
	}

	// フォールバック: user ターンを直接解析
	return extractLastUserTurn(lines)
}

// transcriptLine は JSONL の user/assistant ターンを表す最小構造体です。
type transcriptLine struct {
	Type        string         `json:"type"`
	IsMeta      bool           `json:"isMeta"`
	IsSidechain bool           `json:"isSidechain"`
	Message     transcriptMsg  `json:"message"`
}

type transcriptMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// extractTextFromContent は content（string または []block）からテキストを取り出します。
// hasToolResult は tool_result ブロックが存在するかを示します。
func extractTextFromContent(raw json.RawMessage) (text string, hasToolResult bool) {
	if len(raw) == 0 {
		return "", false
	}
	// string として試みる
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, false
	}
	// []block として試みる
	var blocks []contentBlock
	if json.Unmarshal(raw, &blocks) != nil {
		return "", false
	}
	for _, b := range blocks {
		switch b.Type {
		case "text":
			text += b.Text
		case "tool_result":
			hasToolResult = true
		}
	}
	return text, hasToolResult
}

// isCommandMessage はスラッシュコマンド合成ターンのテキストかを判定します。
func isCommandMessage(text string) bool {
	return strings.Contains(text, "<command-name>") ||
		strings.Contains(text, "<command-message>") ||
		strings.Contains(text, "<local-command-stdout>")
}

// extractLastUserTurn は逆順行リストから最後の人間入力ターンのテキストを返します。
// "last-prompt" エントリが見つからない場合のフォールバックです。
func extractLastUserTurn(lines []string) string {
	for _, raw := range lines {
		var tl transcriptLine
		if err := json.Unmarshal([]byte(raw), &tl); err != nil {
			continue
		}
		if tl.Type != "user" {
			continue
		}
		if tl.IsMeta || tl.IsSidechain {
			continue
		}
		if tl.Message.Role != "user" {
			continue
		}
		text, hasToolResult := extractTextFromContent(tl.Message.Content)
		if hasToolResult {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" || isCommandMessage(text) {
			continue
		}
		text = imageTagRe.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
	}
	return ""
}

// extractLastAssistantText はトランスクリプトから最後の assistant テキストを返します。
func extractLastAssistantText(path string) string {
	lines := readLinesReverse(path, 200)
	for _, raw := range lines {
		var tl transcriptLine
		if err := json.Unmarshal([]byte(raw), &tl); err != nil {
			continue
		}
		if tl.Type != "assistant" || tl.IsSidechain {
			continue
		}
		text, _ := extractTextFromContent(tl.Message.Content)
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
	}
	return ""
}

// truncate は s を max rune 数に制限し、超過時は末尾に … を付けます。
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// firstNonEmpty は空白でない最初の値を返します。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// resolve はイベント種別から通知内容を決定します。
func resolve(in InputData) (notification, bool) {
	project := ""
	if in.Cwd != "" {
		project = filepath.Base(in.Cwd)
	}

	withProject := func(label string) string {
		if project == "" {
			return label
		}
		return label + " · " + project
	}

	txPath := transcriptFilePath(in)

	getUserPrompt := func() string {
		if txPath == "" {
			return ""
		}
		return truncate(extractLastUserPrompt(txPath), 100)
	}

	getAssistantText := func() string {
		if txPath == "" {
			return ""
		}
		return truncate(extractLastAssistantText(txPath), 100)
	}

	switch in.HookEventName {
	case "Stop":
		body := firstNonEmpty(getUserPrompt(), "タスクが完了しました")
		return notification{withProject("完了"), body, "Glass"}, true

	case "StopFailure":
		body := firstNonEmpty(in.Message, "処理がエラーで中断しました")
		return notification{withProject("エラー"), body, "Basso"}, true

	case "Notification":
		switch in.NotificationType {
		case "permission_prompt":
			// message（許可対象の説明）を優先し、なければ Claude の最後の発言
			body := firstNonEmpty(in.Message, getAssistantText(), "権限の確認が必要です")
			return notification{withProject("許可待ち"), body, "Submarine"}, true

		case "idle_prompt":
			// Claude の最後の発言（何を求めているか）を表示
			body := firstNonEmpty(getAssistantText(), in.Message, "入力を待機しています")
			return notification{withProject("入力待ち"), body, "Submarine"}, true

		default:
			// 未知の Notification でも message があれば出す
			if strings.TrimSpace(in.Message) != "" {
				return notification{withProject("通知"), in.Message, "Submarine"}, true
			}
		}
	}
	return notification{}, false
}

// escapeAppleScript は AppleScript の文字列リテラルに安全に埋め込めるよう
// バックスラッシュ・ダブルクォートをエスケープし、改行をスペースに変換します。
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func main() {
	var input InputData
	decoder := json.NewDecoder(os.Stdin)

	// JSON デコードに失敗しても静かに終了（non-blocking hook）
	if err := decoder.Decode(&input); err != nil {
		return
	}

	// デバッグ用: CLAUDE_NOTIFY_DEBUG が設定されていれば raw 値をログに残す
	// （フィールド名の実機検証用。通常運用では未設定）
	if os.Getenv("CLAUDE_NOTIFY_DEBUG") != "" {
		if f, err := os.OpenFile("/tmp/claude-notify.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			raw, _ := json.Marshal(input)
			fmt.Fprintf(f, "%s\n", raw)
			f.Close()
		}
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
