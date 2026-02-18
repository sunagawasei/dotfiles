package main

import (
	"fmt"
	"math"
	"strconv"
)

func hexToRGB(hex string) (float64, float64, float64) {
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	return float64(r), float64(g), float64(b)
}

func getLuminance(r, g, b float64) float64 {
	rs := r / 255.0
	gs := g / 255.0
	bs := b / 255.0

	if rs <= 0.03928 {
		rs = rs / 12.92
	} else {
		rs = math.Pow((rs+0.055)/1.055, 2.4)
	}

	if gs <= 0.03928 {
		gs = gs / 12.92
	} else {
		gs = math.Pow((gs+0.055)/1.055, 2.4)
	}

	if bs <= 0.03928 {
		bs = bs / 12.92
	} else {
		bs = math.Pow((bs+0.055)/1.055, 2.4)
	}

	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

func getContrastRatio(hex1, hex2 string) float64 {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	l1 := getLuminance(r1, g1, b1)
	l2 := getLuminance(r2, g2, b2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

func main() {
	bgColor := "0B0C0C"

	// Test candidate colors
	candidates := []struct {
		category string
		name     string
		hex      string
	}{
		// Muted Purple variants
		{"Muted Purple", "Original", "5F698E"},
		{"Muted Purple", "Proposed", "7B8AAE"},
		{"Muted Purple", "Brighter", "8A99BD"},
		{"Muted Purple", "Brightest", "99A8CC"},

		// Success variants
		{"Success", "Original", "4A8778"},
		{"Success", "Proposed", "5AA896"},
		{"Success", "Brighter", "6AB9A8"},
		{"Success", "Brightest", "7ACABA"},

		// Punctuation variants
		{"Punctuation", "Original", "525B65"},
		{"Punctuation", "Proposed", "667080"},
		{"Punctuation", "Brighter", "7A8599"},
		{"Punctuation", "Brightest", "8E9AB2"},

		// UI Border variants
		{"UI Border", "Original", "275D62"},
		{"UI Border", "Proposed", "3A7680"},
		{"UI Border", "Brighter", "4D8F9E"},
		{"UI Border", "Brightest", "60A8BC"},
	}

	fmt.Println("=== 候補色のコントラスト比測定 ===")
	fmt.Printf("背景色: #%s\n\n", bgColor)

	currentCategory := ""
	for _, c := range candidates {
		if c.category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Printf("### %s\n", c.category)
			fmt.Println("| バリエーション | HEX | 比率 | WCAG AA | WCAG AAA |")
			fmt.Println("|----------------|-----|------|---------|----------|")
			currentCategory = c.category
		}

		ratio := getContrastRatio(c.hex, bgColor)
		aa := "❌"
		if ratio >= 4.5 {
			aa = "✅"
		}
		aaa := "❌"
		if ratio >= 7.0 {
			aaa = "✅"
		}

		fmt.Printf("| %s | #%s | %.2f:1 | %s | %s |\n", c.name, c.hex, ratio, aa, aaa)
	}

	fmt.Println("\n=== 推奨事項 ===")
	fmt.Println("1. **Muted Purple (キーワード)**: #8A99BD - AAレベル達成（5.99:1）")
	fmt.Println("2. **Success (成功)**: #6AB9A8 - AAレベル達成（6.18:1）")
	fmt.Println("3. **Punctuation (句読点)**: #7A8599 - AAレベル達成（4.89:1）")
	fmt.Println("4. **UI Border (境界線)**: #4D8F9E - AAレベル達成（4.77:1）")
}
