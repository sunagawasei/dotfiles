package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// TOML構造体
type AbyssalTealPalette struct {
	Core        map[string]string `toml:"core"`
	Foregrounds map[string]string `toml:"foregrounds"`
	Teals       map[string]string `toml:"teals"`
	BluesSlates map[string]string `toml:"blues_slates"`
	Purples     map[string]string `toml:"purples"`
	Semantic    map[string]string `toml:"semantic"`
	Ansi        map[string]string `toml:"ansi"`
	Wezterm     map[string]string `toml:"wezterm"`
	Nvim        map[string]string `toml:"nvim"`
	Zsh         map[string]string `toml:"zsh"`
}

// 色定義
type ColorCheck struct {
	Name     string
	Hex      string
	Priority string
	Usage    string
	Category string
}

// hexToRGB converts hex color to RGB (既存のコードから再利用)
func hexToRGB(hex string) (float64, float64, float64) {
	// #を除去
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	return float64(r), float64(g), float64(b)
}

// getLuminance calculates relative luminance (既存のコードから再利用)
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

// getContrastRatio calculates contrast ratio between two colors (既存のコードから再利用)
func getContrastRatio(hex1, hex2 string) float64 {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	l1 := getLuminance(r1, g1, b1)
	l2 := getLuminance(r2, g2, b2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

// evaluateWCAG evaluates WCAG compliance (既存のコードから再利用)
func evaluateWCAG(ratio float64) (string, string) {
	var aa, aaa string

	if ratio >= 4.5 {
		aa = "✓"
	} else {
		aa = "✗"
	}

	if ratio >= 7.0 {
		aaa = "✓"
	} else {
		aaa = "✗"
	}

	return aa, aaa
}

func main() {
	// 1. TOMLファイル読み込み
	var palette AbyssalTealPalette
	tomlPath := "../colors/abyssal-teal.toml"
	if _, err := toml.DecodeFile(tomlPath, &palette); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading TOML: %v\n", err)
		os.Exit(1)
	}

	// 2. 背景色
	bgColor := "0B0C0C"

	// 3. チェック対象色の定義
	checks := []ColorCheck{
		// CRITICAL - AAA必須 (7.0:1)
		{"Main Text", palette.Foregrounds["main"], "CRITICAL", "全テキスト", "text"},
		{"Brightest", palette.Foregrounds["bright"], "CRITICAL", "強調テキスト", "text"},
		{"Vibrant Teal", palette.Teals["bright"], "CRITICAL", "関数・コマンド（最頻使用）", "accent"},
		{"Heading", palette.Foregrounds["heading"], "CRITICAL", "見出し・パス", "text"},
		{"ANSI Bright White", palette.Ansi["bright_white"], "CRITICAL", "最高強調", "ansi"},

		// HIGH - AA必須 (4.5:1)
		{"Comment", palette.Semantic["comment"], "HIGH", "コメント（改善済み）", "syntax"},
		{"Git Blame", palette.Semantic["git_blame"], "HIGH", "Git blame", "ui"},
		{"Punctuation", palette.Semantic["punctuation"], "HIGH", "句読点（改善済み）", "syntax"},
		{"Error", palette.Semantic["error"], "HIGH", "エラー", "semantic"},
		{"Success", palette.Semantic["success"], "HIGH", "成功（AAA明示）", "semantic"},
		{"Operator", palette.Semantic["operator"], "HIGH", "演算子・選択", "syntax"},

		// MEDIUM - 推奨 (3.0:1以上)
		{"Dim Text", palette.Foregrounds["dim"], "MEDIUM", "ディムテキスト", "text"},
		{"UI Border", palette.Teals["border"], "MEDIUM", "UI境界線", "ui"},
		{"Sky Slate", palette.BluesSlates["sky_slate"], "MEDIUM", "型・オプション", "syntax"},
		{"Base Teal", palette.Teals["standard"], "MEDIUM", "文字列", "syntax"},
		{"Muted Purple", palette.Purples["muted_purple"], "MEDIUM", "キーワード", "syntax"},

		// ANSI Standard Colors (0-7)
		{"ANSI Black", palette.Ansi["black"], "MEDIUM", "ANSI 0", "ansi"},
		{"ANSI Red", palette.Ansi["red"], "MEDIUM", "ANSI 1", "ansi"},
		{"ANSI Green", palette.Ansi["green"], "MEDIUM", "ANSI 2", "ansi"},
		{"ANSI Yellow", palette.Ansi["yellow"], "MEDIUM", "ANSI 3", "ansi"},
		{"ANSI Blue", palette.Ansi["blue"], "MEDIUM", "ANSI 4", "ansi"},
		{"ANSI Magenta", palette.Ansi["magenta"], "MEDIUM", "ANSI 5", "ansi"},
		{"ANSI Cyan", palette.Ansi["cyan"], "MEDIUM", "ANSI 6", "ansi"},
		{"ANSI White", palette.Ansi["white"], "MEDIUM", "ANSI 7", "ansi"},

		// ANSI Bright Colors (8-15)
		{"ANSI Bright Black", palette.Ansi["bright_black"], "MEDIUM", "ANSI 8", "ansi"},
		{"ANSI Bright Red", palette.Ansi["bright_red"], "MEDIUM", "ANSI 9", "ansi"},
		{"ANSI Bright Green", palette.Ansi["bright_green"], "MEDIUM", "ANSI 10", "ansi"},
		{"ANSI Bright Yellow", palette.Ansi["bright_yellow"], "MEDIUM", "ANSI 11", "ansi"},
		{"ANSI Bright Blue", palette.Ansi["bright_blue"], "MEDIUM", "ANSI 12", "ansi"},
		{"ANSI Bright Magenta", palette.Ansi["bright_magenta"], "MEDIUM", "ANSI 13", "ansi"},
		{"ANSI Bright Cyan", palette.Ansi["bright_cyan"], "MEDIUM", "ANSI 14", "ansi"},
		// ANSI Bright White (15) は既にCRITICALとして追加済み
	}

	// 4. ホワイトリスト
	whitelist := map[string]bool{
		"525B65": true, // slate_mid
		"111E16": true, // darkest_bg
		"152A2B": true, // panel_bg
		"1E1E24": true, // ui_shadow
		"2B2D32": true, // active_line
	}

	// 5. チェック実行・結果出力
	printResults(checks, bgColor, whitelist)
}

func printResults(checks []ColorCheck, bgColor string, whitelist map[string]bool) {
	fmt.Println("=== Abyssal Teal 包括的コントラスト比測定 ===")
	fmt.Printf("背景色: #%s (Main Background)\n\n", bgColor)

	// 優先度別にグループ化
	priorities := []string{"CRITICAL", "HIGH", "MEDIUM"}
	priorityLabels := map[string]string{
		"CRITICAL": "CRITICAL Priority (AAA: 7.0:1 Required)",
		"HIGH":     "HIGH Priority (AA: 4.5:1 Required)",
		"MEDIUM":   "MEDIUM Priority (Recommended: 3.0:1+)",
	}

	var failures []string

	for _, priority := range priorities {
		fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ %-67s │\n", priorityLabels[priority])
		fmt.Println("├─────────────────────────────────────────────────────────────────────┤")
		fmt.Println("│ 色名              │ HEX     │ 比率    │ AA │ AAA │ 用途              │")
		fmt.Println("├───────────────────┼─────────┼─────────┼────┼─────┼───────────────────┤")

		for _, check := range checks {
			if check.Priority != priority {
				continue
			}

			// ホワイトリストチェック
			hexClean := strings.TrimPrefix(check.Hex, "#")
			if whitelist[hexClean] {
				continue
			}

			ratio := getContrastRatio(check.Hex, bgColor)
			aa, aaa := evaluateWCAG(ratio)

			// 不合格の記録
			if priority == "CRITICAL" && aaa != "✓" {
				failures = append(failures, fmt.Sprintf("  - %s (%s): %.2f:1 (AAA: %s)", check.Name, check.Hex, ratio, aaa))
			} else if priority == "HIGH" && aa != "✓" {
				failures = append(failures, fmt.Sprintf("  - %s (%s): %.2f:1 (AA: %s)", check.Name, check.Hex, ratio, aa))
			}

			fmt.Printf("│ %-17s │ %-7s │ %6.2f:1 │ %-2s │ %-3s │ %-17s │\n",
				truncate(check.Name, 17),
				check.Hex,
				ratio,
				aa,
				aaa,
				truncate(check.Usage, 17),
			)
		}
		fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Markdown要素のチェック
	fmt.Println("┌─────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Markdown Elements (Post-Fix Status)                                │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 要素                              │ 色      │ 比率    │ 状態       │")
	fmt.Println("├───────────────────────────────────┼─────────┼─────────┼────────────┤")

	markdownElements := []struct {
		name   string
		hex    string
		status string
	}{
		{"@markup.list", "#92A2AB", "✓"},
		{"@markup.list.markdown", "#92A2AB", "✓"},
		{"@markup.list.unchecked", "#92A2AB", "✓"},
		{"@markup.quote", "#92A2AB", "✓"},
		{"@markup.heading.6", "#525B65", "✗"},
		{"@punctuation.special.markdown", "#92A2AB", "✓"},
	}

	for _, elem := range markdownElements {
		ratio := getContrastRatio(elem.hex, bgColor)
		fmt.Printf("│ %-33s │ %-7s │ %6.2f:1 │ %-10s │\n",
			elem.name,
			elem.hex,
			ratio,
			elem.status,
		)
	}
	fmt.Println("└─────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// 不合格色サマリー
	fmt.Println("=== 不合格色サマリー ===")
	if len(failures) > 0 {
		fmt.Println("以下の色がWCAG基準を満たしていません:\n")
		for _, failure := range failures {
			fmt.Println(failure)
		}
	} else {
		fmt.Println("全ての色がWCAG基準を満たしています！ ✓")
	}
	fmt.Println()

	// ホワイトリスト表示
	fmt.Println("=== ホワイトリスト（チェック除外） ===")
	fmt.Println("以下は意図的に低コントラストで保持:")
	fmt.Println("  - #525B65: LspCodeLens, Conceal, 非アクティブUI")
	fmt.Println("  - #111E16, #152A2B, #1E1E24: 背景・シャドウ")
	fmt.Println()

	// 評価基準
	fmt.Println("=== 評価基準 ===")
	fmt.Println("WCAG AA: 4.5:1以上（通常テキスト）")
	fmt.Println("WCAG AAA: 7.0:1以上（推奨）")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
