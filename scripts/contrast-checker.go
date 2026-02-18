package main

import (
	"fmt"
	"math"
	"strconv"
)

// hexToRGB converts hex color to RGB
func hexToRGB(hex string) (float64, float64, float64) {
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	return float64(r), float64(g), float64(b)
}

// getLuminance calculates relative luminance
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

// getContrastRatio calculates contrast ratio between two colors
func getContrastRatio(hex1, hex2 string) float64 {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	l1 := getLuminance(r1, g1, b1)
	l2 := getLuminance(r2, g2, b2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

// evaluateWCAG evaluates WCAG compliance
func evaluateWCAG(ratio float64) (string, string) {
	var aa, aaa string

	if ratio >= 4.5 {
		aa = "✓ PASS"
	} else {
		aa = "✗ FAIL"
	}

	if ratio >= 7.0 {
		aaa = "✓ PASS"
	} else {
		aaa = "✗ FAIL"
	}

	return aa, aaa
}

func main() {
	bgColor := "0B0C0C"

	colors := []struct {
		name       string
		hexOld     string
		hexNew     string
		colorLabel string
	}{
		{"Muted Purple", "5F698E", "7B8AAE", "キーワード"},
		{"Success", "4A8778", "5AA896", "成功インジケーター"},
		{"Punctuation (Slate Mid)", "525B65", "667080", "句読点"},
		{"UI Border", "275D62", "3A7680", "境界線"},
	}

	fmt.Println("=== Abyssal Teal コントラスト比測定 ===")
	fmt.Printf("背景色: #%s\n\n", bgColor)

	fmt.Println("| 色名 | 用途 | 変更前HEX | 比率 | AA | AAA | 変更後HEX | 比率 | AA | AAA |")
	fmt.Println("|------|------|-----------|------|----|----|-----------|------|----|----|")

	for _, color := range colors {
		ratioOld := getContrastRatio(color.hexOld, bgColor)
		aaOld, aaaOld := evaluateWCAG(ratioOld)

		ratioNew := getContrastRatio(color.hexNew, bgColor)
		aaNew, aaaNew := evaluateWCAG(ratioNew)

		fmt.Printf("| %s | %s | #%s | %.2f:1 | %s | %s | #%s | %.2f:1 | %s | %s |\n",
			color.name,
			color.colorLabel,
			color.hexOld,
			ratioOld,
			aaOld,
			aaaOld,
			color.hexNew,
			ratioNew,
			aaNew,
			aaaNew,
		)
	}

	fmt.Println("\n=== 評価基準 ===")
	fmt.Println("WCAG AA: 4.5:1以上（通常テキスト）")
	fmt.Println("WCAG AAA: 7.0:1以上（推奨）")

	// Additional check for intentionally dim colors
	fmt.Println("\n=== 意図的に暗く保つ色（参考） ===")
	dimColors := []struct {
		name string
		hex  string
		use  string
	}{
		{"LspCodeLens", "525B65", "補助情報"},
		{"Conceal", "525B65", "隠しテキスト"},
		{"Inactive Tab", "525B65", "非アクティブ表示"},
	}

	fmt.Println("| 色名 | HEX | 用途 | 比率 | AA | AAA |")
	fmt.Println("|------|-----|------|------|----|-----|")
	for _, color := range dimColors {
		ratio := getContrastRatio(color.hex, bgColor)
		aa, aaa := evaluateWCAG(ratio)
		fmt.Printf("| %s | #%s | %s | %.2f:1 | %s | %s |\n",
			color.name,
			color.hex,
			color.use,
			ratio,
			aa,
			aaa,
		)
	}
}
