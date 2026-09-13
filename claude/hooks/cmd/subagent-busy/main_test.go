package main

import "testing"

func TestValidMarkerComponentAcceptsAllowedCharacters(t *testing.T) {
	for _, value := range []string{
		"abc-123",
		"aws-close-impl-7933fb5e11cb077e",
		"impl_worker-abcdef",
	} {
		t.Run(value, func(t *testing.T) {
			if !validMarkerComponent(value) {
				t.Fatalf("validMarkerComponent(%q) = false, want true", value)
			}
		})
	}
}

func TestValidMarkerComponentRejectsPathBreakingValues(t *testing.T) {
	for _, value := range []string{
		"../etc",
		"a/b",
		"a.b",
		"",
	} {
		t.Run(value, func(t *testing.T) {
			if validMarkerComponent(value) {
				t.Fatalf("validMarkerComponent(%q) = true, want false", value)
			}
		})
	}
}
