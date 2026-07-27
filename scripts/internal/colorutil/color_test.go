package colorutil

import (
	"math"
	"testing"
)

func TestContrastRatioUsesWCAGSRGBConversion(t *testing.T) {
	ratio, err := ContrastRatio("#000000", "#FFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(ratio-21) > 1e-12 {
		t.Fatalf("contrast ratio = %.12f, want 21", ratio)
	}

	atThreshold := SRGBToLinear(0.04045)
	wantThreshold := 0.04045 / 12.92
	if math.Abs(atThreshold-wantThreshold) > 1e-12 {
		t.Fatalf("sRGB threshold conversion = %.12f, want %.12f", atThreshold, wantThreshold)
	}

	aboveThreshold := 0.04046
	wantAbove := math.Pow((aboveThreshold+0.055)/1.055, 2.4)
	if math.Abs(SRGBToLinear(aboveThreshold)-wantAbove) > 1e-12 {
		t.Fatalf("sRGB value above 0.04045 did not use the exponential branch")
	}
}

func TestParseHexRGB01RejectsLegacyFormats(t *testing.T) {
	for _, value := range []string{"FFFFFF", " #FFFFFF", "#FFFFFF00", "#GGGGGG"} {
		if _, _, _, err := ParseHexRGB01(value); err == nil {
			t.Errorf("ParseHexRGB01(%q) unexpectedly succeeded", value)
		}
	}
}
