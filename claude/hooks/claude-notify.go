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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// InputData は hook から渡される stdin JSON の必要フィールドを表します。
type InputData struct {
	HookEventName        string `json:"hook_event_name"`
	NotificationType     string `json:"notification_type"`
	Message              string `json:"message"`
	Cwd                  string `json:"cwd"`
	SessionID            string `json:"session_id"`
	TranscriptPath       string `json:"transcript_path"`
	LastAssistantMessage string `json:"last_assistant_message"`
	StopHookActive       bool   `json:"stop_hook_active"`
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

// customTitleEntry は JSONL の "custom-title" タイプエントリを表します。
type customTitleEntry struct {
	Type        string `json:"type"`
	CustomTitle string `json:"customTitle"`
}

// extractCustomTitle はトランスクリプトから会話タイトル(customTitle)を返します。
// これは Claude Code が OSC 2 端末タイトルに設定する要約と同一で、
// WezTerm のタブバーに表示される文字列に一致します（先頭グリフは含まない）。
func extractCustomTitle(path string) string {
	lines := readLinesReverse(path, 300)
	for _, raw := range lines {
		var entry customTitleEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			continue
		}
		if entry.Type != "custom-title" {
			continue
		}
		title := strings.TrimSpace(entry.CustomTitle)
		if title != "" {
			return title
		}
	}
	return ""
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
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Name  string          `json:"name"`  // tool_use
	Input json.RawMessage `json:"input"` // tool_use
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

// toolInputSummary はツール名と input JSON からサマリ文字列を組み立てます。
func toolInputSummary(name string, input json.RawMessage) string {
	var m map[string]json.RawMessage
	if json.Unmarshal(input, &m) != nil {
		return name
	}
	getString := func(key string) string {
		v, ok := m[key]
		if !ok {
			return ""
		}
		var s string
		if json.Unmarshal(v, &s) == nil {
			return s
		}
		return ""
	}
	switch name {
	case "Bash":
		if d := getString("description"); d != "" {
			return name + ": " + d
		}
		if c := getString("command"); c != "" {
			// コマンド先頭だけ表示（改行前まで）
			if idx := strings.IndexByte(c, '\n'); idx > 0 {
				c = c[:idx]
			}
			return name + ": " + c
		}
	case "AskUserQuestion":
		// questions[0].question を取り出す
		if raw, ok := m["questions"]; ok {
			var qs []struct {
				Question string `json:"question"`
			}
			if json.Unmarshal(raw, &qs) == nil && len(qs) > 0 && qs[0].Question != "" {
				return name + ": " + qs[0].Question
			}
		}
	case "Read", "Write", "Edit", "MultiEdit":
		if fp := getString("file_path"); fp != "" {
			return name + ": " + filepath.Base(fp)
		}
	case "WebFetch":
		if u := getString("url"); u != "" {
			return name + ": " + u
		}
	case "WebSearch":
		if q := getString("query"); q != "" {
			return name + ": " + q
		}
	}
	return name
}

// extractPendingToolUse はトランスクリプトの最後の assistant ターンから
// 許可待ちの tool_use のサマリを返します。
func extractPendingToolUse(path string) string {
	lines := readLinesReverse(path, 50)
	for _, raw := range lines {
		var tl transcriptLine
		if json.Unmarshal([]byte(raw), &tl) != nil {
			continue
		}
		if tl.Type != "assistant" || tl.IsSidechain {
			continue
		}
		var blocks []contentBlock
		if json.Unmarshal(tl.Message.Content, &blocks) != nil {
			continue
		}
		for _, b := range blocks {
			if b.Type == "tool_use" && b.Name != "" {
				return toolInputSummary(b.Name, b.Input)
			}
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

// truncateLine は s から改行を除去し、max rune 数に制限します。
func truncateLine(s string, max int) string {
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
	return truncate(s, max)
}

// formatMultiline はテキストを最大 maxLines 行・各行 maxRunes 文字に制限して返します。
// 空行はスキップします。
func formatMultiline(s string, maxLines, maxRunes int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, truncate(line, maxRunes))
		if len(result) >= maxLines {
			break
		}
	}
	return strings.Join(result, "\n")
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
	txPath := transcriptFilePath(in)

	// サブタイトル右側の識別子。WezTerm タブバーと同じ会話タイトル(customTitle)を
	// 優先し、無ければ cwd のディレクトリ名にフォールバックする。
	// （新規/直後の /clear セッションは custom-title 未生成のため空になりうる）
	context := ""
	if txPath != "" {
		context = truncate(extractCustomTitle(txPath), 40)
	}
	if context == "" && in.Cwd != "" {
		context = filepath.Base(in.Cwd)
	}

	withProject := func(label string) string {
		if context == "" {
			return label
		}
		return label + " · " + context
	}

	getUserPrompt := func() string {
		if txPath == "" {
			return ""
		}
		return extractLastUserPrompt(txPath)
	}

	getAssistantText := func() string {
		if txPath == "" {
			return ""
		}
		return truncate(extractLastAssistantText(txPath), 100)
	}

	getPendingTool := func() string {
		if txPath == "" {
			return ""
		}
		return extractPendingToolUse(txPath)
	}

	// permission_prompt の message が汎用テキストか否かを判定する
	isGenericMessage := func(msg string) bool {
		return strings.EqualFold(strings.TrimSpace(msg), "claude needs your permission")
	}

	switch in.HookEventName {
	case "Stop":
		// stop_hook_active=true はループ継続中の中間発火なので通知しない
		if in.StopHookActive {
			return notification{}, false
		}
		body := firstNonEmpty(truncateLine(getUserPrompt(), 80), "タスクが完了しました")
		return notification{withProject("完了"), body, "Glass"}, true

	case "StopFailure":
		body := firstNonEmpty(in.Message, "処理がエラーで中断しました")
		return notification{withProject("エラー"), body, "Basso"}, true

	case "Notification":
		switch in.NotificationType {
		case "permission_prompt":
			// 「どの会話か（1行目）→ 何のツールか（2行目）」の2行形式で組み立て
			userPrompt := getUserPrompt()
			toolSummary := getPendingTool()
			var body string
			switch {
			case userPrompt != "" && toolSummary != "":
				body = truncateLine(userPrompt, 45) + "\n→ " + truncateLine(toolSummary, 45)
			case toolSummary != "":
				body = truncateLine(toolSummary, 80)
			case !isGenericMessage(in.Message) && in.Message != "":
				body = in.Message
			default:
				body = firstNonEmpty(getAssistantText(), "権限の確認が必要です")
			}
			return notification{withProject("許可待ち"), body, "Submarine"}, true

		case "idle_prompt":
			// pending tool_use がなければタスク完了後のアイドルなので通知しない
			// （Stop 通知で既に完了を伝えており、追加通知は不要）
			if txPath != "" && extractPendingToolUse(txPath) == "" {
				return notification{}, false
			}
			// Claude が明示的に応答を求めている（AskUserQuestion 等）ケースのみ通知
			// 最大2行・各行45文字で表示
			assistantText := ""
			if txPath != "" {
				assistantText = formatMultiline(extractLastAssistantText(txPath), 2, 45)
			}
			body := firstNonEmpty(assistantText, in.Message, "入力を待機しています")
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
