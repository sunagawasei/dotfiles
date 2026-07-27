package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var (
	modeColorPattern    = regexp.MustCompile(`^\s*(normal|insert|visual|replace|command|terminal|inactive)\s*=\s*p\.([A-Za-z0-9_]+)`)
	lualineModePattern  = regexp.MustCompile(`^\s*(normal|insert|visual|replace|command|terminal|inactive)\s*=\s*\{\s*$`)
	lualinePartPattern  = regexp.MustCompile(`^\s*([abc])\s*=\s*\{\s*fg\s*=\s*([^,]+),\s*bg\s*=\s*([^,}]+)`)
	lualineDiffPattern  = regexp.MustCompile(`^\s*(added|modified|removed)\s*=\s*\{\s*fg\s*=\s*p\.([A-Za-z0-9_]+)`)
	pExpressionPattern  = regexp.MustCompile(`^p\.([A-Za-z0-9_]+)$`)
	modeExpressionMatch = regexp.MustCompile(`^mode_colors\.([A-Za-z0-9_]+)$`)
)

func extractLualine(root string, result *Result) error {
	const relative = "nvim/lua/plugins/lualine.lua"
	path := sourcePath(root, relative)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	modeColors := make(map[string]verifycolors.TokenRef)
	currentMode := ""
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		source := fmt.Sprintf("%s:%d", relative, lineNumber)

		if match := modeColorPattern.FindStringSubmatch(line); match != nil {
			token, err := resolveNvimAlias(match[2])
			if err != nil {
				return fmt.Errorf("%s: %w", source, err)
			}
			modeColors[match[1]] = token
			continue
		}
		if match := lualineModePattern.FindStringSubmatch(line); match != nil {
			currentMode = match[1]
			continue
		}
		if currentMode != "" {
			if match := lualinePartPattern.FindStringSubmatch(line); match != nil {
				foreground, err := resolveLualineExpression(strings.TrimSpace(match[2]), modeColors)
				if err != nil {
					return fmt.Errorf("%s: %w", source, err)
				}
				background, err := resolveLualineBackground(strings.TrimSpace(match[3]), modeColors)
				if err != nil {
					return fmt.Errorf("%s: %w", source, err)
				}
				consumerID := "nvim.lualine." + currentMode + "." + match[1]
				result.addPair(defaultTextPair(
					consumerID,
					foreground,
					background,
					[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
					source,
				))
				if !background.Ambient {
					result.addPair(surfacePair(consumerID+".surface", background.Token, source))
				}
				continue
			}
			if strings.TrimSpace(line) == "}," {
				currentMode = ""
			}
		}
		if match := lualineDiffPattern.FindStringSubmatch(line); match != nil {
			foreground, err := resolveNvimAlias(match[2])
			if err != nil {
				return fmt.Errorf("%s: %w", source, err)
			}
			result.addPair(defaultTextPair(
				"nvim.lualine.diff."+match[1],
				foreground,
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				source,
			))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func resolveLualineExpression(
	expression string,
	modeColors map[string]verifycolors.TokenRef,
) (verifycolors.TokenRef, error) {
	if match := pExpressionPattern.FindStringSubmatch(expression); match != nil {
		return resolveNvimAlias(match[1])
	}
	if match := modeExpressionMatch.FindStringSubmatch(expression); match != nil {
		token, ok := modeColors[match[1]]
		if !ok {
			return "", fmt.Errorf("unknown mode_colors entry %q", match[1])
		}
		return token, nil
	}
	return "", fmt.Errorf("unsupported lualine color expression %q", expression)
}

func resolveLualineBackground(
	expression string,
	modeColors map[string]verifycolors.TokenRef,
) (verifycolors.BackgroundRef, error) {
	if expression == `"none"` || expression == `"NONE"` {
		return verifycolors.AmbientBackground(), nil
	}
	token, err := resolveLualineExpression(expression, modeColors)
	if err != nil {
		return verifycolors.BackgroundRef{}, err
	}
	return verifycolors.TokenBackground(token), nil
}
