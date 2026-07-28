package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	lazygitUnstagedPattern = regexp.MustCompile(`(?m)^\s*unstagedChangesColor:\n\s*-\s*'\{\{([a-z_]+\.[a-z_]+)\}\}'`)
	fzfAccentPattern       = regexp.MustCompile(`--color=prompt:\$\{colors\.([a-z_]+\.[a-z_]+)\},spinner:\$\{colors\.([a-z_]+\.[a-z_]+)\},pointer:\$\{colors\.core\.selection_fg\},header:\$\{colors\.blues_slates\.comment_gray\}`)
)

func extractGitUIRoleColors(root string, result *Result) error {
	if err := extractLazygitGitColor(root, result); err != nil {
		return err
	}
	return extractFZFAccentColors(root, result)
}

func extractLazygitGitColor(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	template, startLine, err := readRawStringConstant(sourcePath(root, relative), "lazygitTemplate")
	if err != nil {
		return err
	}
	matches := lazygitUnstagedPattern.FindAllStringSubmatch(template, -1)
	if len(matches) != 1 {
		return fmt.Errorf("%s:%d: lazygit unstagedChangesColor match count = %d, want 1", relative, startLine, len(matches))
	}
	token := verifycolors.TokenRef(matches[0][1])
	if token != "git.changed" {
		return fmt.Errorf("%s:%d: lazygit unstagedChangesColor = %q, want %q", relative, startLine, token, "git.changed")
	}
	result.addPair(defaultTextPair(
		"lazygit.theme.unstagedChangesColor",
		token,
		verifycolors.AmbientBackground(),
		[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		fmt.Sprintf("%s:%d", relative, startLine),
	))
	return nil
}

func extractFZFAccentColors(root string, result *Result) error {
	const relative = "home-manager/zsh.nix"
	file, err := os.Open(sourcePath(root, relative))
	if err != nil {
		return err
	}
	defer file.Close()

	found := false
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		match := fzfAccentPattern.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		if found {
			return fmt.Errorf("%s:%d: duplicate FZF prompt/spinner color definition", relative, lineNumber)
		}
		found = true
		source := fmt.Sprintf("%s:%d", relative, lineNumber)
		for index, role := range []string{"prompt", "spinner"} {
			token := verifycolors.TokenRef(match[index+1])
			if token != "ui.accent_fg" {
				return fmt.Errorf("%s: FZF %s = %q, want %q", source, role, token, "ui.accent_fg")
			}
			result.addPair(defaultTextPair(
				"zsh.fzf."+role,
				token,
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				source,
			))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%s: FZF prompt/spinner color definition was not found", relative)
	}
	return nil
}
