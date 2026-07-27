package main

import (
	"crypto/sha256"
	"fmt"
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
			wantHash: "86472a6e5b21b853a7c8c18a9ee86f165175b90872073068680cab4ad8e06b3c",
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
