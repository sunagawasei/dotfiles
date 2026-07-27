package main

import "testing"

func TestGeneratedInventoryCheck(t *testing.T) {
	if code := run([]string{"--check"}); code != 0 {
		t.Fatalf("generate-color-inventory --check exit code = %d, want 0", code)
	}
}
