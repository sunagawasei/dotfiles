// claude-notify は Claude Code の hook として動作し、
// stdin の JSON を解析して macOS のデスクトップ通知を出します。
//
// 通知表示そのものは macOS 標準の osascript に委譲し、
// このバイナリは「イベント種別の判定」「動的メッセージの取り込みと
// フォールバック」「AppleScript 文字列のエスケープ」を担当します。
//
// hook 用途のため non-blocking（エラーは握りつぶして常に exit 0）。
package main

import (
	"encoding/json"
	"fmt"
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
}

// notification は組み立てた通知内容を表します。
type notification struct {
	subtitle string
	body     string
	sound    string
}

// resolve はイベント種別から通知内容を決定します。
// 動的メッセージ（message）が空の場合は固定文言にフォールバックし、
// 空通知を絶対に出さないようにします。
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

	dynamicOr := func(fallback string) string {
		if strings.TrimSpace(in.Message) != "" {
			return in.Message
		}
		return fallback
	}

	switch in.HookEventName {
	case "Stop":
		return notification{withProject("完了"), "タスクが完了しました", "Glass"}, true
	case "StopFailure":
		return notification{withProject("エラー"), dynamicOr("処理がエラーで中断しました"), "Basso"}, true
	case "Notification":
		switch in.NotificationType {
		case "permission_prompt":
			return notification{withProject("許可待ち"), dynamicOr("権限の確認が必要です"), "Submarine"}, true
		case "idle_prompt":
			return notification{withProject("入力待ち"), dynamicOr("入力を待機しています"), "Submarine"}, true
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
