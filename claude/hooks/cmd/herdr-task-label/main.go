// herdr-task-label は Claude Code の hook として動作し、herdr に
// 2種類の情報を反映します:
//
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
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InputData は hook から渡される stdin JSON の必要フィールドを表します。
type InputData struct {
	HookEventName string `json:"hook_event_name"`
	Prompt        string `json:"prompt"`
	Cwd           string `json:"cwd"`
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
			reportWorkspaceDir(workspaceID, input.Cwd)
		}
	}
}
