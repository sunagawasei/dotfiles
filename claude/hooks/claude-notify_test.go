package main

import (
	"os"
	"path/filepath"
	"testing"
)

// このディレクトリには複数の package main フックが同居し go.mod を置けないため、
// テストは明示ファイル指定で実行する:
//   go test claude-notify.go claude-notify_test.go

// writeTranscript はテスト用の JSONL トランスクリプトを一時ファイルに書き出す。
func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	var data string
	for _, l := range lines {
		data += l + "\n"
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	return path
}

const (
	aiLine     = `{"type":"ai-title","aiTitle":"AI生成タイトル","sessionId":"s1"}`
	customLine = `{"type":"custom-title","customTitle":"ユーザー設定タイトル (Branch)","sessionId":"s1"}`
	userLine   = `{"type":"user","message":{"role":"user","content":"こんにちは"}}`
	toolLine   = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{}}]}}`
	textLine   = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"完了しました"}]}}`
)

func TestExtractConversationTitle(t *testing.T) {
	const (
		aiLineNew = `{"type":"ai-title","aiTitle":"更新後タイトル","sessionId":"s1"}`
		emptyAI   = `{"type":"ai-title","aiTitle":"  ","sessionId":"s1"}`
	)

	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"ai-title形式を抽出", []string{userLine, aiLine}, "AI生成タイトル"},
		{"custom-title形式を抽出", []string{userLine, customLine}, "ユーザー設定タイトル (Branch)"},
		{"複数のai-titleでは末尾(最新)を優先", []string{aiLine, userLine, aiLineNew}, "更新後タイトル"},
		{"タイトルエントリが無ければ空", []string{userLine, userLine}, ""},
		{"空白のみのai-titleはスキップし直前の有効値を返す", []string{aiLine, emptyAI}, "AI生成タイトル"},
		{"不正JSON行は無視して有効なタイトルを返す", []string{`{壊れたJSON`, aiLine}, "AI生成タイトル"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTranscript(t, tc.lines...)
			if got := extractConversationTitle(path); got != tc.want {
				t.Errorf("extractConversationTitle() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractConversationTitleMissingFile(t *testing.T) {
	if got := extractConversationTitle("/nonexistent/transcript.jsonl"); got != "" {
		t.Errorf("extractConversationTitle(missing) = %q, want empty", got)
	}
}

func TestHasPendingToolUse(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  bool
	}{
		{"末尾assistantにtool_useあり", []string{userLine, toolLine}, true},
		{"末尾assistantがテキストのみ", []string{userLine, textLine}, false},
		{"assistantターンが無い", []string{userLine}, false},
		// 回帰: 完了後(最新assistantがtext)に古いtool_useを拾わない
		{"古いtool_useは遡らない(最新がtext)", []string{toolLine, textLine}, false},
		{"完了後シーケンス(tool_use→tool_result→text)", []string{toolLine, userLine, textLine}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTranscript(t, tc.lines...)
			if got := hasPendingToolUse(path); got != tc.want {
				t.Errorf("hasPendingToolUse() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestResolve は状態=サブタイトル・会話タイトル=本文 のレイアウトを固定する。
func TestResolve(t *testing.T) {
	cases := []struct {
		name         string
		in           InputData
		lines        []string // トランスクリプト内容（空なら作成しない）
		wantOK       bool
		wantSubtitle string
		wantBody     string
	}{
		{
			name:         "Stop: 完了 + 会話タイトル",
			in:           InputData{HookEventName: "Stop", Cwd: "/Users/x/proj"},
			lines:        []string{userLine, aiLine},
			wantOK:       true,
			wantSubtitle: "完了",
			wantBody:     "AI生成タイトル",
		},
		{
			name:   "Stop: stop_hook_active なら通知しない",
			in:     InputData{HookEventName: "Stop", StopHookActive: true, Cwd: "/Users/x/proj"},
			lines:  []string{aiLine},
			wantOK: false,
		},
		{
			name:         "StopFailure: エラー",
			in:           InputData{HookEventName: "StopFailure", Cwd: "/Users/x/proj"},
			lines:        []string{aiLine},
			wantOK:       true,
			wantSubtitle: "エラー",
			wantBody:     "AI生成タイトル",
		},
		{
			name:         "permission_prompt: 許可待ち",
			in:           InputData{HookEventName: "Notification", NotificationType: "permission_prompt", Cwd: "/Users/x/proj"},
			lines:        []string{aiLine, toolLine},
			wantOK:       true,
			wantSubtitle: "許可待ち",
			wantBody:     "AI生成タイトル",
		},
		{
			name:         "idle_prompt: pending toolあり → 入力待ち",
			in:           InputData{HookEventName: "Notification", NotificationType: "idle_prompt", Cwd: "/Users/x/proj"},
			lines:        []string{aiLine, toolLine},
			wantOK:       true,
			wantSubtitle: "入力待ち",
			wantBody:     "AI生成タイトル",
		},
		{
			name:   "idle_prompt: pending toolなし → 通知しない",
			in:     InputData{HookEventName: "Notification", NotificationType: "idle_prompt", Cwd: "/Users/x/proj"},
			lines:  []string{aiLine, textLine},
			wantOK: false,
		},
		{
			name:   "idle_prompt: 完了後(最新text・古いtool_use)→通知しない",
			in:     InputData{HookEventName: "Notification", NotificationType: "idle_prompt", Cwd: "/Users/x/proj"},
			lines:  []string{aiLine, toolLine, userLine, textLine},
			wantOK: false,
		},
		{
			name:         "タイトル未生成時は cwd ディレクトリ名にフォールバック",
			in:           InputData{HookEventName: "Stop", Cwd: "/Users/x/myproj"},
			lines:        []string{userLine},
			wantOK:       true,
			wantSubtitle: "完了",
			wantBody:     "myproj",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.lines) > 0 {
				tc.in.TranscriptPath = writeTranscript(t, tc.lines...)
			}
			got, ok := resolve(tc.in)
			if ok != tc.wantOK {
				t.Fatalf("resolve() ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if got.subtitle != tc.wantSubtitle {
				t.Errorf("subtitle = %q, want %q", got.subtitle, tc.wantSubtitle)
			}
			if got.body != tc.wantBody {
				t.Errorf("body = %q, want %q", got.body, tc.wantBody)
			}
		})
	}
}
