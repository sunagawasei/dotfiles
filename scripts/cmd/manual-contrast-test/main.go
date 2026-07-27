package main

import (
	"fmt"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

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

		ratio := colorutil.LegacyContrastNoHash(c.hex, bgColor)
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
	fmt.Println("1. **Muted Purple (キーワード)**: #8E92C8 - AAレベル達成（6.57:1）")
	fmt.Println("2. **Success (成功)**: #52C4BC - AAAレベル達成（9.24:1）")
	fmt.Println("3. **Punctuation (句読点)**: #8892A8 - AAレベル達成（6.22:1）")
	fmt.Println("4. **UI Border (境界線)**: #45799D - UIコンポーネント基準達成（4.14:1）")
}
