package main

import (
	"fmt"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

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

	fmt.Println("=== Ghost Visor コントラスト比測定 ===")
	fmt.Printf("背景色: #%s\n\n", bgColor)

	fmt.Println("| 色名 | 用途 | 変更前HEX | 比率 | AA | AAA | 変更後HEX | 比率 | AA | AAA |")
	fmt.Println("|------|------|-----------|------|----|----|-----------|------|----|----|")

	for _, color := range colors {
		ratioOld := colorutil.LegacyContrastNoHash(color.hexOld, bgColor)
		aaOld, aaaOld := evaluateWCAG(ratioOld)

		ratioNew := colorutil.LegacyContrastNoHash(color.hexNew, bgColor)
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
		ratio := colorutil.LegacyContrastNoHash(color.hex, bgColor)
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
