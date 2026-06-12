// guard-readonly-agents は read-only 分析エージェント（codex/copilot/cursor）の
// PreToolUse フックとして動作し、CLI 起動と preflight 以外の生 Bash を拒否する。
package main

import (
	"encoding/json"
	"os"
	"strings"
)

type hookInput struct {
	ToolName  string    `json:"tool_name"`
	ToolInput toolInput `json:"tool_input"`
}

type toolInput struct {
	Command string `json:"command"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

const denyReason = "このエージェントのBashはCLI起動(codex/copilot/cursor-agent)とpreflightのみ許可。調査・編集・gitはCLIに委譲するか、結果の提案として出力すること"

func writeDeny() {
	out := hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: denyReason,
		},
	}
	json.NewEncoder(os.Stdout).Encode(out) //nolint:errcheck
}

// firstToken returns the first whitespace-delimited token of s.
func firstToken(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

// stripEnvPrefix removes leading VAR=val assignments, e.g. "FOO=bar BAZ=qux cmd" → "cmd".
func stripEnvPrefix(cmd string) string {
	for {
		cmd = strings.TrimSpace(cmd)
		// Match token of the form NAME=... (NAME is uppercase-ish, contains no space)
		eq := strings.IndexByte(cmd, '=')
		if eq < 0 {
			break
		}
		name := cmd[:eq]
		if !isEnvName(name) {
			break
		}
		// skip past the value (may be quoted or bare)
		rest := cmd[eq+1:]
		if len(rest) == 0 {
			cmd = ""
			break
		}
		if rest[0] == '"' {
			end := strings.Index(rest[1:], "\"")
			if end < 0 {
				break
			}
			cmd = rest[end+2:]
		} else if rest[0] == '\'' {
			end := strings.Index(rest[1:], "'")
			if end < 0 {
				break
			}
			cmd = rest[end+2:]
		} else {
			sp := strings.IndexAny(rest, " \t")
			if sp < 0 {
				cmd = ""
			} else {
				cmd = rest[sp:]
			}
		}
	}
	return strings.TrimSpace(cmd)
}

func isEnvName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// hasShellComposition returns true if cmd contains shell composition operators
// that would allow running arbitrary secondary commands.
func hasShellComposition(cmd string) bool {
	// Remove known-safe constructs before checking
	safe := []string{"$(pwd)", "< /dev/null", "2>&1", ">/dev/null", "> /dev/null"}
	clean := cmd
	for _, s := range safe {
		clean = strings.ReplaceAll(clean, s, "")
	}
	// Shell operators that chain or substitute commands
	dangerous := []string{";", "&&", " || ", " | ", "`", "$("}
	for _, d := range dangerous {
		if strings.Contains(clean, d) {
			return true
		}
	}
	return false
}

// isPreflight returns true for safe version/status queries.
func isPreflight(cmd string) bool {
	tok := firstToken(cmd)
	if tok == "which" || tok == "command" {
		return true
	}
	prefixes := []string{
		"codex --version", "codex login status",
		"copilot --binary-version", "copilot --version", "copilot --help",
		"cursor-agent --version", "cursor-agent --help",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(cmd, p) {
			return true
		}
	}
	return false
}

func isCodexSafe(cmd string) bool {
	if !strings.Contains(cmd, "--sandbox read-only") {
		return false
	}
	forbidden := []string{"workspace-write", "--full-auto", "--dangerously-"}
	for _, f := range forbidden {
		if strings.Contains(cmd, f) {
			return false
		}
	}
	return true
}

func isCopilotSafe(cmd string) bool {
	forbidden := []string{
		"--allow-all-tools", "--allow-all-paths",
		"--allow-all ", "--allow-all\t",
		"--allow-tool", "--allow-url",
	}
	for _, f := range forbidden {
		if strings.Contains(cmd, f) {
			return false
		}
	}
	return true
}

func isCursorSafe(cmd string) bool {
	if !strings.Contains(cmd, "--mode plan") && !strings.Contains(cmd, "--mode ask") {
		return false
	}
	forbidden := []string{"--force", "--yolo", "--sandbox disabled"}
	for _, f := range forbidden {
		if strings.Contains(cmd, f) {
			return false
		}
	}
	return true
}

func isAllowed(rawCmd string) bool {
	cmd := strings.TrimSpace(rawCmd)
	if cmd == "" {
		return true
	}
	if hasShellComposition(cmd) {
		return false
	}
	if isPreflight(cmd) {
		return true
	}
	stripped := stripEnvPrefix(cmd)
	tok := firstToken(stripped)
	switch tok {
	case "codex":
		return isCodexSafe(stripped)
	case "copilot":
		return isCopilotSafe(stripped)
	case "cursor-agent":
		return isCursorSafe(stripped)
	default:
		return false
	}
}

func main() {
	var input hookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		os.Exit(0)
	}
	if input.ToolName != "Bash" {
		os.Exit(0)
	}
	if !isAllowed(input.ToolInput.Command) {
		writeDeny()
	}
}
