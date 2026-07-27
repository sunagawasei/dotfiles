package main

import (
	"fmt"
	"math"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

// rgbToHex converts RGB to hex
func rgbToHex(r, g, b float64) string {
	return fmt.Sprintf("%02X%02X%02X", int(r), int(g), int(b))
}

// brightenColor increases brightness while maintaining hue
func brightenColor(hex string, factor float64) string {
	r, g, b := colorutil.LegacyNoHash(hex)

	// Increase RGB values proportionally
	r = math.Min(255, r*factor)
	g = math.Min(255, g*factor)
	b = math.Min(255, b*factor)

	return rgbToHex(r, g, b)
}

// findOptimalBrightness finds the brightness factor to achieve target contrast
func findOptimalBrightness(hexColor, bgColor string, targetRatio float64) (string, float64) {
	bestHex := hexColor
	bestRatio := colorutil.LegacyContrastNoHash(hexColor, bgColor)

	// Try increasing brightness factors
	for factor := 1.1; factor <= 2.0; factor += 0.05 {
		newHex := brightenColor(hexColor, factor)
		ratio := colorutil.LegacyContrastNoHash(newHex, bgColor)

		if ratio >= targetRatio && ratio < bestRatio+1.0 {
			bestHex = newHex
			bestRatio = ratio
		}

		if ratio >= targetRatio+0.5 {
			break
		}
	}

	return bestHex, bestRatio
}

func main() {
	bgColor := "0B0C0C"
	targetAA := 4.5
	targetAAA := 7.0

	colors := []struct {
		name    string
		hexOld  string
		purpose string
	}{
		{"Muted Purple", "5F698E", "キーワード"},
		{"Success", "4A8778", "成功インジケーター"},
		{"Punctuation", "525B65", "句読点"},
		{"UI Border", "275D62", "境界線"},
	}

	fmt.Println("=== Ghost Visor 最適化された色の提案 ===")
	fmt.Printf("背景色: #%s\n", bgColor)
	fmt.Printf("目標: WCAG AA (%.1f:1) 以上\n\n", targetAA)

	fmt.Println("| 色名 | 用途 | 現在 | 比率 | AA提案 | 比率 | AAA提案 | 比率 |")
	fmt.Println("|------|------|------|------|--------|------|---------|------|")

	for _, color := range colors {
		currentRatio := colorutil.LegacyContrastNoHash(color.hexOld, bgColor)
		aaOptimal, aaRatio := findOptimalBrightness(color.hexOld, bgColor, targetAA)
		aaaOptimal, aaaRatio := findOptimalBrightness(color.hexOld, bgColor, targetAAA)

		fmt.Printf("| %s | %s | #%s | %.2f:1 | #%s | %.2f:1 | #%s | %.2f:1 |\n",
			color.name,
			color.purpose,
			color.hexOld,
			currentRatio,
			aaOptimal,
			aaRatio,
			aaaOptimal,
			aaaRatio,
		)
	}

	fmt.Println("\n=== 推奨事項 ===")
	fmt.Println("1. コードキーワード（Muted Purple）: 視認性が最も重要 → AAA基準推奨")
	fmt.Println("2. 成功インジケーター（Success）: 既にAA合格だが、AAA推奨")
	fmt.Println("3. 句読点（Punctuation）: AA基準を満たす必要あり")
	fmt.Println("4. UI境界線（Border）: AA基準を満たすか、意図的に控えめにするか判断")
}
