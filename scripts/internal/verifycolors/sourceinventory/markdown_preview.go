package sourceinventory

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

type cssDeclaration struct {
	property string
	value    string
}

type cssRule struct {
	selectors    []string
	declarations []cssDeclaration
}

type markdownPreviewHighlightSpec struct {
	selectors []string
	token     verifycolors.TokenRef
	property  string
}

var markdownPreviewBaseVariables = map[string]verifycolors.TokenRef{
	"--color-bg-primary":                 "core.background",
	"--color-bg-secondary":               "core.ui_shadow",
	"--color-bg-tertiary":                "core.panel_bg",
	"--color-text-primary":               "foregrounds.main",
	"--color-text-tertiary":              "foregrounds.dim",
	"--color-text-link":                  "teals.bright",
	"--color-markdown-code-bg":           "core.panel_bg",
	"--color-markdown-blockquote-border": "teals.border",
	"--color-border-primary":             "teals.border",
	"--color-border-secondary":           "core.panel_bg",
	"--color-border-tertiary":            "core.ui_shadow",
	"--color-markdown-table-border":      "teals.border",
	"--color-markdown-table-tr-border":   "core.ui_shadow",
	"--color-kbd-foreground":             "foregrounds.dim",
}

var markdownPreviewPageVariables = map[string]verifycolors.TokenRef{
	"--foreground-color":           "foregrounds.main",
	"--background-color":           "core.background",
	"--secondary-background-color": "core.darkest_bg",
	"--border-color":               "teals.border",
}

var markdownPreviewHeadingTokens = map[string]verifycolors.TokenRef{
	".markdown-body h1": "foregrounds.heading",
	".markdown-body h2": "teals.bright",
	".markdown-body h3": "teals.mid_bright",
	".markdown-body h4": "foregrounds.main",
	".markdown-body h5": "blues_slates.cloud_slate",
	".markdown-body h6": "foregrounds.subdued",
}

var markdownPreviewHighlightSpecs = []markdownPreviewHighlightSpec{
	{selectors: []string{".hljs-comment", ".hljs-quote"}, token: "semantic.comment", property: "color"},
	{selectors: []string{".hljs-keyword", ".hljs-selector-tag", ".hljs-subst"}, token: "semantic.keyword", property: "color"},
	{selectors: []string{".hljs-number", ".hljs-literal"}, token: "semantic.number", property: "color"},
	{selectors: []string{".hljs-variable", ".hljs-template-variable", ".hljs-tag .hljs-attr"}, token: "semantic.variable", property: "color"},
	{selectors: []string{".hljs-string", ".hljs-doctag"}, token: "semantic.string", property: "color"},
	{selectors: []string{".hljs-title", ".hljs-section", ".hljs-selector-id"}, token: "semantic.function", property: "color"},
	{selectors: []string{".hljs-type", ".hljs-class .hljs-title"}, token: "semantic.type", property: "color"},
	{selectors: []string{".hljs-tag", ".hljs-name", ".hljs-attribute"}, token: "foregrounds.heading", property: "color"},
	{selectors: []string{".hljs-regexp", ".hljs-link"}, token: "teals.bright", property: "color"},
	{selectors: []string{".hljs-symbol", ".hljs-bullet"}, token: "semantic.punctuation", property: "color"},
	{selectors: []string{".hljs-built_in", ".hljs-builtin-name"}, token: "teals.mid_bright", property: "color"},
	{selectors: []string{".hljs-meta"}, token: "foregrounds.dim", property: "color"},
	{selectors: []string{".hljs-deletion"}, token: "nvim.diff_delete_bg", property: "background"},
	{selectors: []string{".hljs-addition"}, token: "nvim.diff_add_bg", property: "background"},
}

var cssTokenPlaceholderPattern = regexp.MustCompile(`^\{\{([a-z_]+\.[a-z_]+)\}\}$`)

func extractMarkdownPreviewTheme(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	path := sourcePath(root, relative)
	markdownTemplate, markdownLine, err := readRawStringConstant(path, "markdownPreviewMarkdownTemplate")
	if err != nil {
		return err
	}
	highlightTemplate, highlightLine, err := readRawStringConstant(path, "markdownPreviewHighlightTemplate")
	if err != nil {
		return err
	}
	if err := validateMarkdownPreviewConfig(sourcePath(root, "nvim/lua/plugins/markdown-preview.lua")); err != nil {
		return err
	}
	if err := validateGeneratedCSS(sourcePath(root, "nvim/generated/mkdp-markdown.css")); err != nil {
		return err
	}
	if err := validateGeneratedCSS(sourcePath(root, "nvim/generated/mkdp-highlight.css")); err != nil {
		return err
	}

	markdownRules, err := parseFlatCSS(markdownTemplate)
	if err != nil {
		return fmt.Errorf("%s:%d: parse markdown preview CSS template: %w", relative, markdownLine, err)
	}
	highlightRules, err := parseFlatCSS(highlightTemplate)
	if err != nil {
		return fmt.Errorf("%s:%d: parse markdown preview highlight template: %w", relative, highlightLine, err)
	}
	markdownSource := fmt.Sprintf("%s:%d", relative, markdownLine)
	highlightSource := fmt.Sprintf("%s:%d", relative, highlightLine)

	if err := validateMarkdownVariables(markdownRules); err != nil {
		return fmt.Errorf("%s:%d: %w", relative, markdownLine, err)
	}
	for selector, token := range markdownPreviewHeadingTokens {
		if err := requireCSSRuleToken(markdownRules, []string{selector}, "color", token); err != nil {
			return fmt.Errorf("%s:%d: %w", relative, markdownLine, err)
		}
	}
	if countTokenPlaceholders(markdownTemplate) != 38 {
		return fmt.Errorf("%s:%d: markdown preview CSS placeholder count = %d, want 38", relative, markdownLine, countTokenPlaceholders(markdownTemplate))
	}

	if err := requireCSSRuleToken(highlightRules, []string{":root"}, "--color-text-primary", "foregrounds.main"); err != nil {
		return fmt.Errorf("%s:%d: %w", relative, highlightLine, err)
	}
	if err := requireCSSRuleToken(highlightRules, []string{`[data-theme="dark"]`}, "--color-text-primary", "foregrounds.main"); err != nil {
		return fmt.Errorf("%s:%d: %w", relative, highlightLine, err)
	}
	if err := requireCSSRuleToken(highlightRules, []string{".hljs"}, "background", "core.panel_bg"); err != nil {
		return fmt.Errorf("%s:%d: %w", relative, highlightLine, err)
	}
	for _, spec := range markdownPreviewHighlightSpecs {
		if err := requireCSSRuleToken(highlightRules, spec.selectors, spec.property, spec.token); err != nil {
			return fmt.Errorf("%s:%d: %w", relative, highlightLine, err)
		}
	}
	if countTokenPlaceholders(highlightTemplate) != 17 {
		return fmt.Errorf("%s:%d: markdown preview highlight placeholder count = %d, want 17", relative, highlightLine, countTokenPlaceholders(highlightTemplate))
	}

	addMarkdownPreviewPairs(result, markdownSource)
	addMarkdownPreviewHighlightPairs(result, highlightSource)
	return nil
}

func validateMarkdownVariables(rules []cssRule) error {
	rootRule, err := findExactCSSRule(rules, []string{":root"})
	if err != nil {
		return err
	}
	darkRule, err := findExactCSSRule(rules, []string{`[data-theme="dark"]`})
	if err != nil {
		return err
	}
	if countCSSVariables(rootRule) != len(markdownPreviewBaseVariables) {
		return fmt.Errorf(":root CSS variable count = %d, want %d", countCSSVariables(rootRule), len(markdownPreviewBaseVariables))
	}
	if countCSSVariables(darkRule) != len(markdownPreviewBaseVariables)+len(markdownPreviewPageVariables) {
		return fmt.Errorf("[data-theme=\"dark\"] CSS variable count = %d, want %d", countCSSVariables(darkRule), len(markdownPreviewBaseVariables)+len(markdownPreviewPageVariables))
	}
	for variable, token := range markdownPreviewBaseVariables {
		if err := requireCSSDeclarationToken(rootRule, variable, token); err != nil {
			return err
		}
		if err := requireCSSDeclarationToken(darkRule, variable, token); err != nil {
			return err
		}
	}
	for variable, token := range markdownPreviewPageVariables {
		if err := requireCSSDeclarationToken(darkRule, variable, token); err != nil {
			return err
		}
	}
	return nil
}

func addMarkdownPreviewPairs(result *Result, source string) {
	background := verifycolors.TokenBackground("core.background")
	add := func(id string, foreground verifycolors.TokenRef, bg verifycolors.BackgroundRef) {
		result.addPair(defaultTextPair(
			"nvim.markdown-preview.markdown."+id,
			foreground,
			bg,
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			source,
		))
	}
	add("text-primary", "foregrounds.main", background)
	add("blockquote", "foregrounds.dim", background)
	add("link", "teals.bright", background)
	for heading := 1; heading <= 6; heading++ {
		selector := fmt.Sprintf(".markdown-body h%d", heading)
		add(fmt.Sprintf("h%d", heading), markdownPreviewHeadingTokens[selector], background)
	}
	add("table", "foregrounds.main", verifycolors.TokenBackground("core.panel_bg"))
	add("kbd", "foregrounds.dim", verifycolors.TokenBackground("core.ui_shadow"))
	add("page-header", "foregrounds.main", verifycolors.TokenBackground("core.background"))
}

func addMarkdownPreviewHighlightPairs(result *Result, source string) {
	panel := verifycolors.TokenBackground("core.panel_bg")
	add := func(id string, foreground verifycolors.TokenRef, background verifycolors.BackgroundRef) {
		result.addPair(defaultTextPair(
			"nvim.markdown-preview.highlight."+id,
			foreground,
			background,
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			source,
		))
	}
	add("hljs", "foregrounds.main", panel)
	for _, spec := range markdownPreviewHighlightSpecs {
		if spec.property == "background" {
			background := verifycolors.TokenBackground(spec.token)
			for _, selector := range spec.selectors {
				add(cssConsumerFragment(selector)+".inherited", "foregrounds.main", background)
			}
			continue
		}
		for _, selector := range spec.selectors {
			add(cssConsumerFragment(selector), spec.token, panel)
		}
	}
}

func validateMarkdownPreviewConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	required := []string{
		`vim.g.mkdp_theme = "dark"`,
		`vim.g.mkdp_markdown_css = vim.fn.expand("~/.config/nvim/generated/mkdp-markdown.css")`,
		`vim.g.mkdp_highlight_css = vim.fn.expand("~/.config/nvim/generated/mkdp-highlight.css")`,
	}
	for _, line := range required {
		if strings.Count(content, line) != 1 {
			return fmt.Errorf("%s: expected exactly one %q", path, line)
		}
	}
	for _, forbidden := range []string{`vim.g.mkdp_theme = "light"`, "AppleInterfaceStyle", `vim.g.mkdp_markdown_css = ""`, `vim.g.mkdp_highlight_css = ""`} {
		if strings.Contains(content, forbidden) {
			return fmt.Errorf("%s: forbidden stale Markdown Preview setting %q", path, forbidden)
		}
	}
	return nil
}

func validateGeneratedCSS(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), "{{") {
		return fmt.Errorf("%s: generated CSS contains an unresolved placeholder", path)
	}
	if _, err := parseFlatCSS(string(data)); err != nil {
		return fmt.Errorf("%s: invalid generated CSS: %w", path, err)
	}
	return nil
}

func parseFlatCSS(content string) ([]cssRule, error) {
	commentPattern := regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = commentPattern.ReplaceAllString(content, "")
	maskedTokenPattern := regexp.MustCompile(`\{\{([a-z_]+\.[a-z_]+)\}\}`)
	content = maskedTokenPattern.ReplaceAllString(content, "@@$1@@")
	var rules []cssRule
	for {
		content = strings.TrimSpace(content)
		if content == "" {
			return rules, nil
		}
		open := strings.IndexByte(content, '{')
		if open < 0 {
			return nil, fmt.Errorf("selector without opening brace: %q", content)
		}
		close := strings.IndexByte(content[open+1:], '}')
		if close < 0 {
			return nil, fmt.Errorf("selector %q has no closing brace", strings.TrimSpace(content[:open]))
		}
		close += open + 1
		selectorText := strings.TrimSpace(content[:open])
		if selectorText == "" {
			return nil, fmt.Errorf("empty selector")
		}
		var selectors []string
		for _, selector := range strings.Split(selectorText, ",") {
			selector = strings.TrimSpace(selector)
			if selector == "" {
				return nil, fmt.Errorf("empty selector in %q", selectorText)
			}
			selectors = append(selectors, selector)
		}
		var declarations []cssDeclaration
		for _, rawDeclaration := range strings.Split(content[open+1:close], ";") {
			rawDeclaration = strings.TrimSpace(rawDeclaration)
			if rawDeclaration == "" {
				continue
			}
			parts := strings.SplitN(rawDeclaration, ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
				return nil, fmt.Errorf("invalid declaration %q in selector %q", rawDeclaration, selectorText)
			}
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(value, "@@") && strings.HasSuffix(value, "@@") {
				value = "{{" + strings.TrimSuffix(strings.TrimPrefix(value, "@@"), "@@") + "}}"
			}
			declarations = append(declarations, cssDeclaration{
				property: strings.TrimSpace(parts[0]),
				value:    value,
			})
		}
		if len(declarations) == 0 {
			return nil, fmt.Errorf("selector %q has no declarations", selectorText)
		}
		rules = append(rules, cssRule{selectors: selectors, declarations: declarations})
		content = content[close+1:]
	}
}

func findExactCSSRule(rules []cssRule, selectors []string) (cssRule, error) {
	var found *cssRule
	for index := range rules {
		if !equalStrings(rules[index].selectors, selectors) {
			continue
		}
		if found != nil {
			return cssRule{}, fmt.Errorf("duplicate CSS rule for selectors %q", strings.Join(selectors, ", "))
		}
		found = &rules[index]
	}
	if found == nil {
		return cssRule{}, fmt.Errorf("CSS rule for selectors %q was not found", strings.Join(selectors, ", "))
	}
	return *found, nil
}

func requireCSSRuleToken(rules []cssRule, selectors []string, property string, token verifycolors.TokenRef) error {
	rule, err := findExactCSSRule(rules, selectors)
	if err != nil {
		return err
	}
	return requireCSSDeclarationToken(rule, property, token)
}

func requireCSSDeclarationToken(rule cssRule, property string, token verifycolors.TokenRef) error {
	count := 0
	for _, declaration := range rule.declarations {
		if declaration.property != property {
			continue
		}
		count++
		match := cssTokenPlaceholderPattern.FindStringSubmatch(declaration.value)
		if match == nil {
			return fmt.Errorf("%s in selectors %q must use a color token placeholder", property, strings.Join(rule.selectors, ", "))
		}
		if verifycolors.TokenRef(match[1]) != token {
			return fmt.Errorf("%s in selectors %q resolves to %q, want %q", property, strings.Join(rule.selectors, ", "), match[1], token)
		}
	}
	if count != 1 {
		return fmt.Errorf("%s declaration count in selectors %q = %d, want 1", property, strings.Join(rule.selectors, ", "), count)
	}
	return nil
}

func countCSSVariables(rule cssRule) int {
	count := 0
	for _, declaration := range rule.declarations {
		if strings.HasPrefix(declaration.property, "--") {
			count++
		}
	}
	return count
}

func countTokenPlaceholders(content string) int {
	return len(regexp.MustCompile(`\{\{[a-z_]+\.[a-z_]+\}\}`).FindAllStringIndex(content, -1))
}

func cssConsumerFragment(selector string) string {
	selector = strings.TrimPrefix(selector, ".")
	return strings.NewReplacer(" .", "_", ".", "_", " ", "_").Replace(selector)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
