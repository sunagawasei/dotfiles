package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func hexToRGB(hex string) (float64, float64, float64) {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return float64(r), float64(g), float64(b)
}

func sRGBtoLinear(c float64) float64 {
	c = c / 255.0
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func relativeLuminance(r, g, b float64) float64 {
	rL := sRGBtoLinear(r)
	gL := sRGBtoLinear(g)
	bL := sRGBtoLinear(b)
	return 0.2126*rL + 0.7152*gL + 0.0722*bL
}

func contrastRatio(hex1, hex2 string) float64 {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)
	
	L1 := relativeLuminance(r1, g1, b1)
	L2 := relativeLuminance(r2, g2, b2)
	
	lighter := math.Max(L1, L2)
	darker := math.Min(L1, L2)
	
	return (lighter + 0.05) / (darker + 0.05)
}

func main() {
	bg := "#0B0C0C"
	oldFg := "#525B65"
	newFg := "#8A97AD"
	
	oldContrast := contrastRatio(bg, oldFg)
	newContrast := contrastRatio(bg, newFg)
	
	fmt.Printf("Background: %s\n", bg)
	fmt.Printf("\nOld inactive_tab_fg: %s\n", oldFg)
	fmt.Printf("  Contrast Ratio: %.2f:1\n", oldContrast)
	fmt.Printf("  WCAG AA (4.5:1): %s\n", map[bool]string{true: "✅", false: "❌"}[oldContrast >= 4.5])
	
	fmt.Printf("\nNew inactive_tab_fg: %s (Git Blame Gray)\n", newFg)
	fmt.Printf("  Contrast Ratio: %.2f:1\n", newContrast)
	fmt.Printf("  WCAG AA (4.5:1): %s\n", map[bool]string{true: "✅", false: "❌"}[newContrast >= 4.5])
	
	improvement := ((newContrast - oldContrast) / oldContrast) * 100
	fmt.Printf("\nImprovement: +%.1f%%\n", improvement)
}
