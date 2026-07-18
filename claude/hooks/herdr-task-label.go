// herdr-task-label は Claude Code の hook として動作し、herdr の agents パネルに
// 3種類の情報を反映します:
//
//   - workspace名（1行目）: Claude Code 自身が生成する会話タイトル
//     （transcript の ai-title/custom-title。WezTerm タブタイトルと同じ仕組み）を
//     タイトルのみで反映。Stop フックで、ターン完了ごとに最新のタイトルへ
//     同期する（毎回作り直すので前回分は積み上がらない）。
//   - workspaceのdirメタデータ: 作業ディレクトリのbasenameを workspace へ
//     反映。SessionStart と Stop フックで報告し、herdr server 再起動後も
//     自動回復する。表示位置は herdr 側の [ui.sidebar.spaces] rows 設定が
//     担う。
//   - agentsパネルの状態表示（2行目、display_agent）: 直近のユーザー指示を
//     短く整形して反映。UserPromptSubmit フックで、プロンプトが送られるたびに
//     最新の内容へ更新する。
//
// herdr の pane/workspace 外（対応する環境変数が未設定）では何もしません。
// hook 用途のため non-blocking（エラーは握りつぶして常に exit 0）。
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InputData は hook から渡される stdin JSON の必要フィールドを表します。
type InputData struct {
	HookEventName  string `json:"hook_event_name"`
	Prompt         string `json:"prompt"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
}

// truncate は s の空白類を1つのスペースに畳み込み、max rune を超える場合は
// 末尾に省略記号を付けます。
func truncate(s string, max int) string {
	joined := strings.Join(strings.Fields(s), " ")
	if joined == "" {
		return ""
	}
	runes := []rune(joined)
	if len(runes) > max {
		return string(runes[:max]) + "…"
	}
	return joined
}

// maxPromptRunes は agents パネル2行目に載せる直近指示の最大文字数です。
// サイドバーは既定26列程度しかなく、"idle · " 分も同じ行を共有するため
// 実際に見える幅に合わせて短めにしています（herdr側に折り返し機能は無く、
// 収まらない分は herdr 側で単純に切り詰められる）。
const maxPromptRunes = 16

// updateLatestPrompt は直近のユーザー指示を herdr の display_agent に反映します。
func updateLatestPrompt(paneID, prompt string) {
	label := truncate(prompt, maxPromptRunes)
	if label == "" {
		return
	}
	_ = exec.Command(
		"herdr", "pane", "report-metadata", paneID,
		"--source", "claude-task-hook",
		"--display-agent", label,
	).Run()
}

// titleEntry は JSONL の会話タイトルエントリを表します。
// Claude Code はバージョン/設定により以下のいずれかの形式で記録します:
//   - {"type":"ai-title","aiTitle":"..."}         AI 自動生成タイトル（現行の既定）
//   - {"type":"custom-title","customTitle":"..."}  ユーザー設定/ブランチ由来タイトル
type titleEntry struct {
	Type        string `json:"type"`
	AITitle     string `json:"aiTitle"`
	CustomTitle string `json:"customTitle"`
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

// extractConversationTitle はトランスクリプトから最新の会話タイトルを返します。
// ai-title / custom-title のどちらの形式にも対応し、末尾（最新）を優先します。
func extractConversationTitle(path string) string {
	for _, raw := range readLinesReverse(path, 1<<30) {
		var entry titleEntry
		if json.Unmarshal([]byte(raw), &entry) != nil {
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

// maxTitleRunes はworkspace名に載せる会話タイトルの最大文字数です。
// ディレクトリ名が2行目へ移り1行目をタイトルが占有できるようになったため、
// 旧値18から拡大（herdr側に折り返し機能は無く、収まらない分は
// herdr側で単純に切り詰められる）。
const maxTitleRunes = 28

// updateSessionTitle は Claude 生成の会話タイトルを herdr の workspace 名に
// 反映します。毎回タイトルから作り直すため、以前の内容が積み上がることは
// ありません。
func updateSessionTitle(workspaceID, transcriptPath string) {
	if transcriptPath == "" {
		return
	}
	title := extractConversationTitle(transcriptPath)
	if title == "" {
		return
	}
	_ = exec.Command("herdr", "workspace", "rename", workspaceID, truncate(title, maxTitleRunes)).Run()
}

// reportWorkspaceDir は作業ディレクトリ名を workspace の dir メタデータに反映します。
func reportWorkspaceDir(workspaceID, cwd string) {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return
		}
	}
	dir := filepath.Base(cwd)
	if dir == "" || dir == "." || dir == "/" {
		return
	}
	_ = exec.Command(
		"herdr", "workspace", "report-metadata", workspaceID,
		"--source", "claude-task-hook",
		"--token", "dir="+dir,
	).Run()
}

func main() {
	rawBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	var input InputData
	if json.Unmarshal(rawBytes, &input) != nil {
		return
	}

	switch input.HookEventName {
	case "UserPromptSubmit":
		if paneID := os.Getenv("HERDR_PANE_ID"); paneID != "" {
			updateLatestPrompt(paneID, input.Prompt)
		}
	case "SessionStart":
		if workspaceID := os.Getenv("HERDR_WORKSPACE_ID"); workspaceID != "" {
			reportWorkspaceDir(workspaceID, input.Cwd)
		}
	case "Stop":
		if workspaceID := os.Getenv("HERDR_WORKSPACE_ID"); workspaceID != "" {
			updateSessionTitle(workspaceID, input.TranscriptPath)
			reportWorkspaceDir(workspaceID, input.Cwd)
		}
	}
}
