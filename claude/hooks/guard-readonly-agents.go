// guard-readonly-agents は read-only 分析エージェント（codex/copilot/cursor）の
// PreToolUse フックとして動作し、CLI 起動と preflight 以外の生 Bash を拒否する。
//
// シェル合成の判定はクォートを認識して行う:
//   - シングルクォート内: 完全に literal（評価されない）→ 無害として無視
//   - ダブルクォート内: コマンド置換 `$(` / バッククォートのみ危険（bash が評価する）。
//     `|` `;` `&&` 等はダブルクォート内では literal なので無視
//   - クォート外: `;` `&&` `|` `$(` バッククォートはすべて危険
// これにより、プロンプト引数に含まれるメタ文字（| ; $() など）の誤検知を防ぎつつ、
// 実際のコマンド連結・インジェクションは確実に拒否する。
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

// splitQuoting walks cmd and returns:
//   - unquoted: the concatenation of all text outside any quotes
//   - dquoted:  the concatenation of all text inside double quotes
//
// Single-quoted segments are dropped entirely (literal, never evaluated).
// `$(pwd)` は事前に除去しておくこと（唯一許可するコマンド置換）。
func splitQuoting(cmd string) (unquoted, dquoted string) {
	var u, d strings.Builder
	i, n := 0, len(cmd)
	for i < n {
		c := cmd[i]
		switch c {
		case '\\':
			// escaped char outside quotes: bash treats `\X` as a literal X and
			// `\"` does NOT open a quote. Consume both bytes and emit a space so
			// the escaped char neither opens a quote nor reads as a metacharacter.
			i += 2
			u.WriteByte(' ')
		case '\'':
			// single-quoted: literal, skip to next single quote
			j := strings.IndexByte(cmd[i+1:], '\'')
			if j < 0 {
				i = n
			} else {
				i = i + 1 + j + 1
			}
		case '"':
			// double-quoted: capture until next unescaped double quote
			i++
			for i < n {
				if cmd[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				if cmd[i] == '"' {
					i++
					break
				}
				d.WriteByte(cmd[i])
				i++
			}
		default:
			u.WriteByte(c)
			i++
		}
	}
	return u.String(), d.String()
}

// removeAllQuoted strips both single- and double-quoted segments, leaving only
// the unquoted skeleton (command + flags). Used for token/flag inspection so
// that prompt text never trips the flag allow/deny rules.
func removeAllQuoted(cmd string) string {
	var b strings.Builder
	i, n := 0, len(cmd)
	for i < n {
		c := cmd[i]
		switch c {
		case '\\':
			// escaped char outside quotes: literal, never opens a quote.
			i += 2
			b.WriteByte(' ')
		case '\'':
			j := strings.IndexByte(cmd[i+1:], '\'')
			if j < 0 {
				i = n
			} else {
				i = i + 1 + j + 1
			}
			b.WriteByte(' ')
		case '"':
			i++
			for i < n {
				if cmd[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				if cmd[i] == '"' {
					i++
					break
				}
				i++
			}
			b.WriteByte(' ')
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// hasShellComposition reports whether cmd contains shell composition / command
// substitution that would let it run secondary commands. Quote-aware: metachars
// inside the (quoted) prompt argument are not flagged.
func hasShellComposition(cmd string) bool {
	// Remove the allowlisted constructs the canonical commands legitimately use,
	// so their `<`, `>`, `&`, `$(` do not trip the strict metachar check below.
	// Order matters: strip the more specific forms before the bare ones.
	for _, s := range []string{
		"$(pwd)",
		"2>&1",
		"2>/dev/null", "2> /dev/null",
		"< /dev/null", "</dev/null",
		">/dev/null", "> /dev/null",
	} {
		cmd = strings.ReplaceAll(cmd, s, " ")
	}

	unquoted, dquoted := splitQuoting(cmd)

	// Outside quotes: any shell-active character can start a second command,
	// background a job, or redirect to a file. Read-only guard → fail closed.
	// Set: ; | & < > $ ( ) { } backtick(\x60) newline carriage-return
	if strings.ContainsAny(unquoted, ";|&<>$(){}\x60\n\r") {
		return true
	}
	// Inside double quotes only command substitution is evaluated by bash.
	if strings.Contains(dquoted, "$(") || strings.Contains(dquoted, "\x60") {
		return true
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
	// Composition check is quote-aware (operates on the full command).
	if hasShellComposition(cmd) {
		return false
	}
	stripped := stripEnvPrefix(cmd)
	if isPreflight(stripped) {
		return true
	}
	// Token + flag checks run on the quote-stripped skeleton so prompt text
	// (which may mention flag-like strings) cannot trip the allow/deny rules.
	bare := removeAllQuoted(stripped)
	tok := firstToken(bare)
	switch tok {
	case "codex":
		return isCodexSafe(bare)
	case "copilot":
		return isCopilotSafe(bare)
	case "cursor-agent":
		return isCursorSafe(bare)
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
