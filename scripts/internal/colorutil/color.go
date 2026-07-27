package colorutil

import (
	"fmt"
	"math"
	"strconv"
)

func ParseHexRGB8(value string) (int, int, int, error) {
	if len(value) != 7 || value[0] != '#' {
		return 0, 0, 0, fmt.Errorf("expected #RRGGBB, got %q", value)
	}
	parsed, err := strconv.ParseUint(value[1:], 16, 24)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse %q: %w", value, err)
	}
	return int(parsed >> 16), int((parsed >> 8) & 0xff), int(parsed & 0xff), nil
}

func ParseHexRGB01(value string) (float64, float64, float64, error) {
	r, g, b, err := ParseHexRGB8(value)
	if err != nil {
		return 0, 0, 0, err
	}
	return float64(r) / 255, float64(g) / 255, float64(b) / 255, nil
}

func RelativeLuminance(hex string) (float64, error) {
	r, g, b, err := ParseHexRGB01(hex)
	if err != nil {
		return 0, err
	}
	return 0.2126*SRGBToLinear(r) + 0.7152*SRGBToLinear(g) + 0.0722*SRGBToLinear(b), nil
}

func ContrastRatio(hexA, hexB string) (float64, error) {
	l1, err := RelativeLuminance(hexA)
	if err != nil {
		return 0, err
	}
	l2, err := RelativeLuminance(hexB)
	if err != nil {
		return 0, err
	}
	return contrastFromLuminance(l1, l2), nil
}

func Clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func SRGBToLinear(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

func contrastFromLuminance(first, second float64) float64 {
	lighter := math.Max(first, second)
	darker := math.Min(first, second)
	return (lighter + 0.05) / (darker + 0.05)
}
