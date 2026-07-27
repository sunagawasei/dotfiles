package colorutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

// LegacyNoHash preserves the old tools' unchecked parsing of RRGGBB without a
// leading '#'. Parse errors are intentionally ignored for compatibility.
func LegacyNoHash(hex string) (float64, float64, float64) {
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	return float64(r), float64(g), float64(b)
}

// LegacyWithHash preserves the enhanced checker's optional '#' removal and
// unchecked parsing. It uses the legacy 0.03928 sRGB branch point.
func LegacyWithHash(hex string) (float64, float64, float64) {
	return LegacyNoHash(strings.TrimPrefix(hex, "#"))
}

// WCAG255WithHash preserves test_contrast's optional '#' removal and unchecked
// 8-bit parsing. It uses the 0.04045 sRGB branch point.
func WCAG255WithHash(hex string) (float64, float64, float64) {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return float64(r), float64(g), float64(b)
}

// StrictRGB01 preserves verify-cvd's whitespace handling, optional '#' removal,
// validation errors, and normalized RGB result.
func StrictRGB01(hex string) (r, g, b float64, err error) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) < 6 {
		return 0, 0, 0, fmt.Errorf("not a 6-digit hex color: %q", hex)
	}
	ri, err1 := strconv.ParseInt(hex[0:2], 16, 64)
	gi, err2 := strconv.ParseInt(hex[2:4], 16, 64)
	bi, err3 := strconv.ParseInt(hex[4:6], 16, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255, nil
}

// LegacyLuminance uses the 0.03928 branch point from the old contrast tools.
func LegacyLuminance(r, g, b float64) float64 {
	return 0.2126*legacyLinear(r/255) + 0.7152*legacyLinear(g/255) + 0.0722*legacyLinear(b/255)
}

// WCAG255Luminance uses the 0.04045 branch point with 8-bit RGB inputs.
func WCAG255Luminance(r, g, b float64) float64 {
	return 0.2126*SRGBToLinear(r/255) + 0.7152*SRGBToLinear(g/255) + 0.0722*SRGBToLinear(b/255)
}

func LegacyContrastNoHash(hexA, hexB string) float64 {
	r1, g1, b1 := LegacyNoHash(hexA)
	r2, g2, b2 := LegacyNoHash(hexB)
	return contrastFromLuminance(LegacyLuminance(r1, g1, b1), LegacyLuminance(r2, g2, b2))
}

func LegacyContrastWithHash(hexA, hexB string) float64 {
	r1, g1, b1 := LegacyWithHash(hexA)
	r2, g2, b2 := LegacyWithHash(hexB)
	return contrastFromLuminance(LegacyLuminance(r1, g1, b1), LegacyLuminance(r2, g2, b2))
}

func WCAG255ContrastWithHash(hexA, hexB string) float64 {
	r1, g1, b1 := WCAG255WithHash(hexA)
	r2, g2, b2 := WCAG255WithHash(hexB)
	return contrastFromLuminance(WCAG255Luminance(r1, g1, b1), WCAG255Luminance(r2, g2, b2))
}

func StrictRelativeLuminance(hex string) (float64, error) {
	r, g, b, err := StrictRGB01(hex)
	if err != nil {
		return 0, err
	}
	return 0.2126*SRGBToLinear(r) + 0.7152*SRGBToLinear(g) + 0.0722*SRGBToLinear(b), nil
}

func StrictContrastRatio(hexA, hexB string) (float64, error) {
	l1, err := StrictRelativeLuminance(hexA)
	if err != nil {
		return 0, err
	}
	l2, err := StrictRelativeLuminance(hexB)
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

func legacyLinear(value float64) float64 {
	if value <= 0.03928 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

func contrastFromLuminance(first, second float64) float64 {
	lighter := math.Max(first, second)
	darker := math.Min(first, second)
	return (lighter + 0.05) / (darker + 0.05)
}
