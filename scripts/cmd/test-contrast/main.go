package main

import (
	"fmt"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

func main() {
	bg := "#0A0C1A"
	oldFg := "#4A4953"
	newFg := "#8C97B5"

	oldContrast := colorutil.WCAG255ContrastWithHash(bg, oldFg)
	newContrast := colorutil.WCAG255ContrastWithHash(bg, newFg)

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
