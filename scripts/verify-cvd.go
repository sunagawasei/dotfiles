// Command verify-cvd is a standalone, single-file numeric verification tool for
// colors/ghost-visor.toml. It is run individually (`go run verify-cvd.go`), not
// built together with the other scripts/*.go files (this directory intentionally
// hosts multiple independent `package main` files run one at a time — see
// scripts/generate-colors.go and scripts/validate-colors.go for the sibling tools).
//
// Methodology:
//   - Contrast: WCAG 2.x relative luminance (sRGB -> linear, Rec. 709 coefficients)
//     and the standard (L1+0.05)/(L2+0.05) contrast ratio formula.
//   - Color Vision Deficiency (CVD) simulation: Machado, Oliveira & Fernandes (2009)
//     linear-RGB transformation matrices at severity=1.0 (full dichromacy), applied
//     to linear sRGB primaries, then converted to CIELAB (D65) and compared with
//     CIE76 Euclidean Delta E. Delta E < 20 is used as a rough "may look identical
//     under this CVD type" screening threshold.
//   - These matrices/formulas were independently re-implemented by both a
//     codex-research agent and a Sonnet subagent during the 2026-07-17 ghost-visor
//     palette redesign; the two implementations' outputs matched to 0.00 across
//     all cross-checked pairs, which is the cross-verification this tool encodes.
//
// Usage:
//
//	go run verify-cvd.go [path/to/palette.toml]
//
// With no argument, resolves colors/ghost-visor.toml relative to the git
// repository root.
package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ---------- palette loading ----------

// rawPalette is a generic TOML section -> key -> value map. Every section in
// colors/ghost-visor.toml is a flat string table, so this unmarshals directly
// without needing a fixed struct of section names.
type rawPalette map[string]map[string]string

func loadPalette(path string) (rawPalette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var p rawPalette
	if err := toml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return p, nil
}

func resolvePalettePath() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("find repository root: %w", err)
	}
	root := strings.TrimSpace(string(out))
	return filepath.Join(root, "colors", "ghost-visor.toml"), nil
}

// ---------- color space plumbing ----------

func hexToRGB01(hex string) (r, g, b float64, err error) {
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

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// relLuminance computes WCAG relative luminance for a hex color.
func relLuminance(hex string) (float64, error) {
	r, g, b, err := hexToRGB01(hex)
	if err != nil {
		return 0, err
	}
	rl, gl, bl := srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)
	return 0.2126*rl + 0.7152*gl + 0.0722*bl, nil
}

// contrastRatio computes the WCAG contrast ratio between two hex colors.
func contrastRatio(hexA, hexB string) (float64, error) {
	l1, err := relLuminance(hexA)
	if err != nil {
		return 0, err
	}
	l2, err := relLuminance(hexB)
	if err != nil {
		return 0, err
	}
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05), nil
}

// ---------- CVD simulation (Machado 2009, severity=1.0, linear sRGB) ----------

var protanopiaM = [3][3]float64{
	{0.152286, 1.052583, -0.204868},
	{0.114503, 0.786281, 0.099216},
	{-0.003882, -0.048116, 1.051998},
}
var deuteranopiaM = [3][3]float64{
	{0.367322, 0.860646, -0.227968},
	{0.280085, 0.672501, 0.047413},
	{-0.011820, 0.042940, 0.968881},
}
var tritanopiaM = [3][3]float64{
	{1.255528, -0.076749, -0.178779},
	{-0.078411, 0.930809, 0.147602},
	{0.004733, 0.691367, 0.303900},
}

type cvdType struct {
	Name string
	M    [3][3]float64
}

var cvdTypes = []cvdType{
	{"protanopia", protanopiaM},
	{"deuteranopia", deuteranopiaM},
	{"tritanopia", tritanopiaM},
}

// applyCVD simulates the given CVD type on a hex color and returns the
// resulting color's CIELAB coordinates.
func applyCVD(hex string, M [3][3]float64) ([3]float64, error) {
	r, g, b, err := hexToRGB01(hex)
	if err != nil {
		return [3]float64{}, err
	}
	rl, gl, bl := srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)
	rl2 := M[0][0]*rl + M[0][1]*gl + M[0][2]*bl
	gl2 := M[1][0]*rl + M[1][1]*gl + M[1][2]*bl
	bl2 := M[2][0]*rl + M[2][1]*gl + M[2][2]*bl
	rl2, gl2, bl2 = clamp01(rl2), clamp01(gl2), clamp01(bl2)
	return linearRGBToLab(rl2, gl2, bl2), nil
}

// linearRGBToLab converts linear sRGB (D65 white point) to CIELAB.
func linearRGBToLab(r, g, b float64) [3]float64 {
	X := 0.4124564*r + 0.3575761*g + 0.1804375*b
	Y := 0.2126729*r + 0.7151522*g + 0.0721750*b
	Z := 0.0193339*r + 0.1191920*g + 0.9503041*b
	Xn, Yn, Zn := 0.95047, 1.0, 1.08883
	f := func(t float64) float64 {
		d := 6.0 / 29.0
		if t > d*d*d {
			return math.Cbrt(t)
		}
		return t/(3*d*d) + 4.0/29.0
	}
	fx, fy, fz := f(X/Xn), f(Y/Yn), f(Z/Zn)
	L := 116*fy - 16
	a := 500 * (fx - fy)
	bb := 200 * (fy - fz)
	return [3]float64{L, a, bb}
}

// deltaE76 computes the CIE76 Euclidean Delta E between two CIELAB colors.
func deltaE76(l1, l2 [3]float64) float64 {
	dl := l1[0] - l2[0]
	da := l1[1] - l2[1]
	db := l1[2] - l2[2]
	return math.Sqrt(dl*dl + da*da + db*db)
}

// ---------- main ----------

const contrastThreshold = 4.5
const cvdDeltaEThreshold = 20.0

type colorEntry struct {
	Section, Key, Hex string
}

func collectColors(p rawPalette) []colorEntry {
	var entries []colorEntry
	for section, kv := range p {
		for key, val := range kv {
			if strings.HasPrefix(strings.TrimSpace(val), "#") {
				entries = append(entries, colorEntry{section, key, val})
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Section != entries[j].Section {
			return entries[i].Section < entries[j].Section
		}
		return entries[i].Key < entries[j].Key
	})
	return entries
}

func main() {
	path, err := resolvePalettePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "verify-cvd:", err)
		os.Exit(1)
	}
	palette, err := loadPalette(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "verify-cvd:", err)
		os.Exit(1)
	}
	fmt.Printf("palette=%s\n", path)

	core, ok := palette["core"]
	if !ok {
		fmt.Fprintln(os.Stderr, "verify-cvd: palette has no [core] section")
		os.Exit(1)
	}
	type bgItem struct{ Name, Hex string }
	var bgs []bgItem
	for _, name := range []string{"background", "panel_bg", "active_line"} {
		hex, ok := core[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "verify-cvd: [core] missing %q\n", name)
			os.Exit(1)
		}
		bgs = append(bgs, bgItem{name, hex})
	}

	// ===== (a) WCAG contrast matrix: all colors x core 3 backgrounds =====
	entries := collectColors(palette)
	fmt.Println("\n=== CONTRAST MATRIX (all colors x core.background/panel_bg/active_line) ===")
	total, ngCount := 0, 0
	var ngRows []string
	for _, e := range entries {
		for _, bg := range bgs {
			ratio, err := contrastRatio(e.Hex, bg.Hex)
			if err != nil {
				fmt.Fprintf(os.Stderr, "verify-cvd: %s.%s: %v\n", e.Section, e.Key, err)
				os.Exit(1)
			}
			total++
			status := "PASS"
			if ratio < contrastThreshold {
				ngCount++
				status = "NG"
				ngRows = append(ngRows, fmt.Sprintf("%s.%s (%s) vs %s (%s): %.2f:1 [need %.1f:1]",
					e.Section, e.Key, e.Hex, bg.Name, bg.Hex, ratio, contrastThreshold))
			}
			fmt.Printf("%-28s %-8s vs %-12s %-8s -> %5.2f:1 [%s]\n",
				e.Section+"."+e.Key, e.Hex, bg.Name, bg.Hex, ratio, status)
		}
	}
	fmt.Printf("\ncontrast: total=%d ng=%d pass=%d (threshold=%.1f:1)\n", total, ngCount, total-ngCount, contrastThreshold)
	if ngCount > 0 {
		fmt.Println("NG (< threshold):")
		for _, row := range ngRows {
			fmt.Println(" -", row)
		}
	}

	// ===== (b) CVD Delta E for all ANSI 16-color pairs =====
	ansi, ok := palette["ansi"]
	if !ok {
		fmt.Fprintln(os.Stderr, "verify-cvd: palette has no [ansi] section")
		os.Exit(1)
	}
	ansiNames := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"bright_black", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	}
	for _, name := range ansiNames {
		if _, ok := ansi[name]; !ok {
			fmt.Fprintf(os.Stderr, "verify-cvd: [ansi] missing %q\n", name)
			os.Exit(1)
		}
	}

	fmt.Println("\n=== CVD DELTA E (ANSI 16-color pairs, Machado 2009 severity=1.0, CIE76) ===")
	cvdTotal, cvdNG := 0, 0
	var cvdNGRows []string
	for _, ct := range cvdTypes {
		labCache := map[string][3]float64{}
		getLab := func(name string) [3]float64 {
			if v, ok := labCache[name]; ok {
				return v
			}
			lab, err := applyCVD(ansi[name], ct.M)
			if err != nil {
				fmt.Fprintf(os.Stderr, "verify-cvd: ansi.%s: %v\n", name, err)
				os.Exit(1)
			}
			labCache[name] = lab
			return lab
		}
		for i := 0; i < len(ansiNames); i++ {
			for j := i + 1; j < len(ansiNames); j++ {
				a, b := ansiNames[i], ansiNames[j]
				de := deltaE76(getLab(a), getLab(b))
				cvdTotal++
				status := "PASS"
				if de < cvdDeltaEThreshold {
					cvdNG++
					status = "NG"
					cvdNGRows = append(cvdNGRows, fmt.Sprintf("%-12s %s(%s) vs %s(%s): dE=%.2f [need >=%.0f]",
						ct.Name, a, ansi[a], b, ansi[b], de, cvdDeltaEThreshold))
				}
				fmt.Printf("%-12s %-14s vs %-14s dE=%5.2f [%s]\n", ct.Name, a, b, de, status)
			}
		}
	}
	fmt.Printf("\ncvd: total=%d ng=%d pass=%d (deltaE threshold=%.0f)\n", cvdTotal, cvdNG, cvdTotal-cvdNG, cvdDeltaEThreshold)
	if cvdNG > 0 {
		fmt.Println("NG (dE < threshold):")
		for _, row := range cvdNGRows {
			fmt.Println(" -", row)
		}
	}

	fmt.Println("\n=== DONE ===")
}
