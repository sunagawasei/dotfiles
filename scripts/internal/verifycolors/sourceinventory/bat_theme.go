package sourceinventory

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var batGlobalTokens = map[string]verifycolors.TokenRef{
	"background":       "core.background",
	"foreground":       "foregrounds.main",
	"lineHighlight":    "core.active_line",
	"gutterForeground": "foregrounds.dim",
}

var batScopeTokens = map[string]verifycolors.TokenRef{
	"comment":                        "semantic.comment",
	"string":                         "semantic.string",
	"constant.numeric":               "semantic.number",
	"constant.language":              "semantic.constant",
	"constant.character.escape":      "foregrounds.bright",
	"keyword":                        "semantic.keyword",
	"keyword.operator":               "semantic.operator",
	"keyword.control.import":         "teals.bright",
	"storage.type":                   "semantic.type",
	"storage.modifier":               "semantic.keyword",
	"entity.name.function":           "semantic.function",
	"variable.function":              "semantic.function",
	"support.function":               "semantic.function",
	"entity.name.class":              "semantic.type",
	"entity.name.type":               "semantic.type",
	"support.type":                   "semantic.type",
	"entity.name.tag":                "teals.bright",
	"entity.other.attribute-name":    "foregrounds.heading",
	"variable.parameter":             "semantic.variable",
	"variable.language":              "foregrounds.heading",
	"punctuation.separator":          "semantic.punctuation",
	"punctuation.terminator":         "semantic.punctuation",
	"punctuation.section":            "semantic.punctuation",
	"punctuation.definition":         "semantic.punctuation",
	"punctuation.accessor":           "semantic.punctuation",
	"invalid.illegal":                "semantic.error",
	"entity.name.tag.yaml":           "teals.bright",
	"constant.language.boolean.yaml": "semantic.keyword",
	"constant.language.null.yaml":    "semantic.keyword",
	"source.yaml constant.numeric":   "purples.lavender",
	"punctuation.definition.block.sequence.item.yaml": "semantic.punctuation",
	"markup.heading":        "foregrounds.heading",
	"markup.bold":           "ansi.bright_white",
	"markup.italic":         "foregrounds.bright",
	"markup.raw":            "semantic.string",
	"markup.quote":          "foregrounds.dim",
	"markup.list":           "semantic.punctuation",
	"markup.underline.link": "teals.bright",
}

func extractBatTheme(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	template, lineNumber, err := readBatThemeTemplate(sourcePath(root, relative))
	if err != nil {
		return err
	}
	document, err := decodePlistDocument(template)
	if err != nil {
		return fmt.Errorf("%s:%d: parse batThemeTemplate: %w", relative, lineNumber, err)
	}

	name, ok := document["name"].(string)
	if !ok || name != "ghost-visor" {
		return fmt.Errorf("%s:%d: bat theme name = %q, want %q", relative, lineNumber, name, "ghost-visor")
	}
	rawEntries, ok := document["settings"].([]any)
	if !ok {
		return fmt.Errorf("%s:%d: bat theme settings array was not found", relative, lineNumber)
	}

	source := fmt.Sprintf("%s:%d", relative, lineNumber)
	globalTokens := make(map[string]verifycolors.TokenRef)
	scopeTokens := make(map[string]verifycolors.TokenRef)
	for index, rawEntry := range rawEntries {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return fmt.Errorf("%s:%d: bat theme entry %d is not a dictionary", relative, lineNumber, index)
		}
		settings, ok := entry["settings"].(map[string]any)
		if !ok {
			return fmt.Errorf("%s:%d: bat theme entry %d has no settings dictionary", relative, lineNumber, index)
		}
		rawScope, hasScope := entry["scope"]
		if !hasScope {
			if len(globalTokens) != 0 {
				return fmt.Errorf("%s:%d: duplicate bat global settings entry", relative, lineNumber)
			}
			if err := collectBatGlobalTokens(settings, globalTokens); err != nil {
				return fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
			}
			continue
		}

		scope, ok := rawScope.(string)
		if !ok || scope == "" {
			return fmt.Errorf("%s:%d: bat theme entry %d has an invalid scope", relative, lineNumber, index)
		}
		expected, ok := batScopeTokens[scope]
		if !ok {
			return fmt.Errorf("%s:%d: unknown bat theme scope %q", relative, lineNumber, scope)
		}
		if _, duplicate := scopeTokens[scope]; duplicate {
			return fmt.Errorf("%s:%d: duplicate bat theme scope %q", relative, lineNumber, scope)
		}
		if len(settings) != 1 {
			return fmt.Errorf("%s:%d: bat theme scope %q must contain only foreground", relative, lineNumber, scope)
		}
		value, ok := settings["foreground"].(string)
		if !ok {
			return fmt.Errorf("%s:%d: bat theme scope %q has no foreground", relative, lineNumber, scope)
		}
		token, err := parseBatTokenPlaceholder(value)
		if err != nil {
			return fmt.Errorf("%s:%d: bat theme scope %q: %w", relative, lineNumber, scope, err)
		}
		if token != expected {
			return fmt.Errorf(
				"%s:%d: bat theme scope %q resolves to %q, want %q",
				relative,
				lineNumber,
				scope,
				token,
				expected,
			)
		}
		scopeTokens[scope] = token
	}

	for key := range batGlobalTokens {
		if _, ok := globalTokens[key]; !ok {
			return fmt.Errorf("%s:%d: bat global setting %q was not found", relative, lineNumber, key)
		}
	}
	for scope := range batScopeTokens {
		if _, ok := scopeTokens[scope]; !ok {
			return fmt.Errorf("%s:%d: bat theme scope %q was not found", relative, lineNumber, scope)
		}
	}

	background := globalTokens["background"]
	lineHighlight := globalTokens["lineHighlight"]
	result.addPair(defaultTextPair(
		"bat.theme.foreground",
		globalTokens["foreground"],
		verifycolors.TokenBackground(background),
		[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		source,
	))
	result.addPair(defaultTextPair(
		"bat.theme.gutterForeground",
		globalTokens["gutterForeground"],
		verifycolors.TokenBackground(background),
		[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		source,
	))
	result.addPair(defaultTextPair(
		"bat.theme.gutterForeground.lineHighlight",
		globalTokens["gutterForeground"],
		verifycolors.TokenBackground(lineHighlight),
		[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		source,
	))
	for scope, token := range scopeTokens {
		result.addPair(defaultTextPair(
			"bat.theme.scope."+strings.ReplaceAll(scope, " ", "_"),
			token,
			verifycolors.TokenBackground(background),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			source,
		))
	}
	return nil
}

func collectBatGlobalTokens(settings map[string]any, tokens map[string]verifycolors.TokenRef) error {
	if len(settings) != len(batGlobalTokens) {
		return fmt.Errorf("bat global settings count = %d, want %d", len(settings), len(batGlobalTokens))
	}
	for key, rawValue := range settings {
		expected, ok := batGlobalTokens[key]
		if !ok {
			return fmt.Errorf("unknown bat global setting %q", key)
		}
		value, ok := rawValue.(string)
		if !ok {
			return fmt.Errorf("bat global setting %q is not a string", key)
		}
		token, err := parseBatTokenPlaceholder(value)
		if err != nil {
			return fmt.Errorf("bat global setting %q: %w", key, err)
		}
		if token != expected {
			return fmt.Errorf("bat global setting %q resolves to %q, want %q", key, token, expected)
		}
		tokens[key] = token
	}
	return nil
}

func parseBatTokenPlaceholder(value string) (verifycolors.TokenRef, error) {
	if !strings.HasPrefix(value, "{{") || !strings.HasSuffix(value, "}}") {
		return "", fmt.Errorf("value %q is not a color token placeholder", value)
	}
	token := strings.TrimSuffix(strings.TrimPrefix(value, "{{"), "}}")
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("value %q is not a simple color token placeholder", value)
	}
	return verifycolors.TokenRef(token), nil
}

func readBatThemeTemplate(path string) (string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	const prefix = "const batThemeTemplate = `"
	var builder strings.Builder
	found := false
	startLine := 0
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !found {
			if strings.HasPrefix(line, prefix) {
				found = true
				startLine = lineNumber
				builder.WriteString(strings.TrimPrefix(line, prefix))
				builder.WriteByte('\n')
			}
			continue
		}
		if line == "`" {
			return builder.String(), startLine, nil
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", 0, err
	}
	if !found {
		return "", 0, fmt.Errorf("batThemeTemplate was not found")
	}
	return "", 0, fmt.Errorf("batThemeTemplate is not terminated")
}

func decodePlistDocument(content string) (map[string]any, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("plist element was not found")
			}
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "plist" {
			continue
		}
		value, err := decodeNextPlistValue(decoder)
		if err != nil {
			return nil, err
		}
		document, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("plist root is not a dictionary")
		}
		return document, nil
	}
}

func decodeNextPlistValue(decoder *xml.Decoder) (any, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		return decodePlistValue(decoder, start)
	}
}

func decodePlistValue(decoder *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "string", "key":
		var value string
		if err := decoder.DecodeElement(&value, &start); err != nil {
			return nil, err
		}
		return value, nil
	case "array":
		var values []any
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch typed := token.(type) {
			case xml.StartElement:
				value, err := decodePlistValue(decoder, typed)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			case xml.EndElement:
				if typed.Name.Local == "array" {
					return values, nil
				}
			}
		}
	case "dict":
		values := make(map[string]any)
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch typed := token.(type) {
			case xml.StartElement:
				if typed.Name.Local != "key" {
					return nil, fmt.Errorf("dict entry starts with %q, want key", typed.Name.Local)
				}
				rawKey, err := decodePlistValue(decoder, typed)
				if err != nil {
					return nil, err
				}
				key := rawKey.(string)
				if _, duplicate := values[key]; duplicate {
					return nil, fmt.Errorf("duplicate plist key %q", key)
				}
				value, err := decodeNextPlistValue(decoder)
				if err != nil {
					return nil, err
				}
				values[key] = value
			case xml.EndElement:
				if typed.Name.Local == "dict" {
					return values, nil
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported plist element %q", start.Name.Local)
	}
}
