// Command verify-cvd is a standalone, single-file numeric verification tool for
// colors/ghost-visor.toml.
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
//	go run ./cmd/verify-cvd [path/to/palette.toml]
//
// With no argument, resolves colors/ghost-visor.toml relative to the git
// repository root.
package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
	"github.com/sunagawasei/dotfiles/scripts/internal/cvd"
	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
)

// ---------- palette loading ----------

// rawPalette is a generic TOML section -> key -> value map. Every section in
// colors/ghost-visor.toml is a flat string table, so this unmarshals directly
// without needing a fixed struct of section names.
type rawPalette map[string]map[string]string

func loadPalette(path string) (rawPalette, error) {
	loaded, err := palette.Load[rawPalette](path)
	if err != nil {
		var loadErr *palette.LoadError
		if errors.As(err, &loadErr) {
			switch loadErr.Stage {
			case palette.LoadStageRead:
				return nil, fmt.Errorf("read %s: %w", path, loadErr.Err)
			case palette.LoadStageParse:
				return nil, fmt.Errorf("parse %s: %w", path, loadErr.Err)
			}
		}
		return nil, err
	}
	return *loaded, nil
}

func resolvePalettePath() (string, error) {
	return palette.ResolveDefaultPath(os.Args[1:])
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
			ratio, err := colorutil.StrictContrastRatio(e.Hex, bg.Hex)
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
	for _, ct := range cvd.Types {
		labCache := map[string][3]float64{}
		getLab := func(name string) [3]float64 {
			if v, ok := labCache[name]; ok {
				return v
			}
			lab, err := cvd.Apply(ansi[name], ct.Matrix)
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
				de := cvd.DeltaE76(getLab(a), getLab(b))
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
