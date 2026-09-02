package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
)

func TestMarkdownPreviewVendorCommitMatchesLock(t *testing.T) {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateMarkdownPreviewVendorCommit(root); err != nil {
		t.Fatal(err)
	}
}

func TestMarkdownPreviewVendorCommitRejectsUpdate(t *testing.T) {
	lock := []byte(`{"markdown-preview.nvim":{"commit":"updated-upstream-commit"}}`)
	err := validateMarkdownPreviewVendorLock(lock)
	if err == nil {
		t.Fatal("validateMarkdownPreviewVendorLock succeeded for a changed commit")
	}
	if !strings.Contains(err.Error(), "rebase the vendored base CSS") {
		t.Fatalf("error %q does not explain that the vendored CSS must be rebased", err)
	}
}

func TestMarkdownPreviewVendoredCSSPreservesUpstreamStructure(t *testing.T) {
	tests := map[string]struct {
		template string
		wantHash string
	}{
		"markdown.css": {
			template: markdownPreviewMarkdownTemplate,
			wantHash: "2fa9b970dd46eb3b1e246a757a09d430d6e2045ee8d0151dd65fc98a4f14939a",
		},
		"highlight.css": {
			template: markdownPreviewHighlightTemplate,
			wantHash: "b9334e6ac86794f9ce67adc6176d2ed3fffcdc1e948513d732fdc072f64898e6",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// The snapshot hashes the fixed upstream CSS plus approved color-only
			// additions after token names are normalized, so non-color changes fail.
			normalized := placeholderPattern.ReplaceAllString(test.template, "{{COLOR}}")
			gotHash := fmt.Sprintf("%x", sha256.Sum256([]byte(normalized)))
			if gotHash != test.wantHash {
				t.Fatalf("normalized vendor hash = %s, want %s; restore upstream selectors/non-color declarations or intentionally rebase the vendor snapshot", gotHash, test.wantHash)
			}
		})
	}
}

func TestHerdrTemplateRequiresRenderContract(t *testing.T) {
	if err := validateHerdrTemplate(herdrTemplate); err != nil {
		t.Fatalf("valid herdr template rejected: %v", err)
	}
	for _, test := range []struct {
		name string
		old  string
		new  string
	}{
		{name: "panel background", old: `panel_bg = "reset"`, new: `panel_bg = "#000000"`},
		{name: "surface dim", old: `surface_dim = "{{ansi.bright_black}}"`, new: `surface_dim = "{{foregrounds.dim}}"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := strings.Replace(herdrTemplate, test.old, test.new, 1)
			if err := validateHerdrTemplate(invalid); err == nil {
				t.Fatalf("validateHerdrTemplate accepted changed %s", test.name)
			}
		})
	}
}

func TestHerdrTemplateRejectsCommentedRequiredAssignment(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		`panel_bg = "reset"`,
		"# panel_bg = \"reset\"\npanel_bg = \"{{core.panel_bg}}\"",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("commented required assignment with wrong assignment accepted")
	}
}

func TestHerdrTemplateRejectsDuplicateAssignment(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		"# END GENERATED COLORS",
		"panel_bg = \"reset\"\n# END GENERATED COLORS",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("duplicate required assignment accepted")
	}
}

func TestHerdrTemplateRejectsWrongAccent(t *testing.T) {
	wrongAccent := "purples." + "deep_glitch"
	invalid := strings.Replace(herdrTemplate, `accent = "{{purples.lavender}}"`, `accent = "{{`+wrongAccent+`}}"`, 1)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("wrong accent assignment accepted")
	}
}

func TestHerdrTemplateRejectsInlineCommentDuplicateAssignment(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		"# END GENERATED COLORS",
		"panel_bg = \"{{core.panel_bg}}\"#x\n# END GENERATED COLORS",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("duplicate assignment with adjacent inline comment accepted")
	}
}

func TestHerdrTemplateRejectsUnparseableAssignment(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		"# END GENERATED COLORS",
		"panel_bg =\n# END GENERATED COLORS",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil || !strings.Contains(err.Error(), "parse herdr template") {
		t.Fatalf("unparseable assignment did not fail through TOML parser: %v", err)
	}
}

func TestHerdrTemplateRejectsQuotedKeyDuplicate(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		"# END GENERATED COLORS",
		"\"panel_bg\" = \"{{core.panel_bg}}\"\n# END GENERATED COLORS",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("duplicate quoted-key assignment accepted")
	}
}

func TestHerdrTemplateRejectsLiteralQuotedKeyDuplicate(t *testing.T) {
	invalid := strings.Replace(
		herdrTemplate,
		"# END GENERATED COLORS",
		"'panel_bg' = \"{{core.panel_bg}}\"\n# END GENERATED COLORS",
		1,
	)
	if err := validateHerdrTemplate(invalid); err == nil {
		t.Fatal("duplicate literal-quoted-key assignment accepted")
	}
}

func TestHerdrTemplateAccentMatchesInventoryOverride(t *testing.T) {
	const accentToken = "purples.lavender"
	if !strings.Contains(herdrTemplate, `accent = "{{`+accentToken+`}}"`) {
		t.Fatalf("herdrTemplate accent does not use %s", accentToken)
	}
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	override, err := os.ReadFile(filepath.Join(root, "scripts", "internal", "verifycolors", "inventory_overrides.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(override), `Background: TokenBackground("`+accentToken+`")`) {
		t.Fatalf("herdr tab label override does not use %s", accentToken)
	}
}
