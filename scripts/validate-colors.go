package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ColorPalette represents the TOML color definition file
type ColorPalette struct {
	Core     map[string]string `toml:"core"`
	ANSI     map[string]string `toml:"ansi"`
	Palette  map[string]string `toml:"palette"`
	Semantic map[string]string `toml:"semantic"`
	WezTerm  map[string]string `toml:"wezterm"`
	Nvim     map[string]string `toml:"nvim"`
	LazyGit  map[string]string `toml:"lazygit"`
	Zsh      map[string]string `toml:"zsh"`
	Starship map[string]string `toml:"starship"`
}

// Allowed colors that are not in the palette (common colors)
var allowedColors = map[string]bool{
	"#000000": true, // Pure black
	"#FFFFFF": true, // Pure white
	"#F2FFFF": true, // Bright white (from palette but may not be in all sections)
}

func main() {
	configDir := "/Users/s23159/.config"
	tomlPath := filepath.Join(configDir, "colors", "abyssal-teal.toml")

	// Load color palette
	palette, err := loadPalette(tomlPath)
	if err != nil {
		fmt.Printf("❌ Failed to load palette: %v\n", err)
		os.Exit(1)
	}

	// Extract all valid colors
	validColors := extractValidColors(palette)
	fmt.Printf("✓ Loaded %d colors from colors/abyssal-teal.toml\n", len(validColors))

	// Files to check
	filesToCheck := []string{
		"wezterm/wezterm.lua",
		"nvim/lua/plugins/colorscheme.lua",
		"nvim/lua/plugins/render-markdown.lua",
		"lazygit/config.yml",
		"zsh/.zshrc",
		"starship.toml",
	}

	hasErrors := false
	for _, relPath := range filesToCheck {
		filePath := filepath.Join(configDir, relPath)
		// Check if file exists before checking
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}
		
		if err := checkFile(filePath, validColors); err != nil {
			fmt.Printf("⚠ Found undefined colors in %s:\n%v\n", relPath, err)
			hasErrors = true
		} else {
			fmt.Printf("✓ All colors in %s are valid\n", relPath)
		}
	}

	if !hasErrors {
		fmt.Println("\n✅ No undefined colors found!")
	} else {
		fmt.Println("\n⚠ Some files contain undefined colors. Please review.")
		os.Exit(1)
	}
}

func loadPalette(path string) (*ColorPalette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var palette ColorPalette
	if err := toml.Unmarshal(data, &palette); err != nil {
		return nil, err
	}

	return &palette, nil
}

func extractValidColors(palette *ColorPalette) map[string]bool {
	colors := make(map[string]bool)

	// Add all colors from all sections
	addColors(colors, palette.Core)
	addColors(colors, palette.ANSI)
	addColors(colors, palette.Palette)
	addColors(colors, palette.Semantic)
	addColors(colors, palette.WezTerm)
	addColors(colors, palette.Nvim)
	addColors(colors, palette.LazyGit)
	addColors(colors, palette.Zsh)
	addColors(colors, palette.Starship)

	// Add allowed colors
	for color := range allowedColors {
		colors[strings.ToUpper(color)] = true
	}

	return colors
}

func addColors(dst map[string]bool, src map[string]string) {
	for _, color := range src {
		// Normalize to uppercase and remove transparency
		normalized := normalizeColor(color)
		if normalized != "" {
			dst[normalized] = true
		}
	}
}

func normalizeColor(color string) string {
	// Remove quotes and whitespace
	color = strings.Trim(strings.TrimSpace(color), `"'`) // Corrected: Removed unnecessary escaping of quotes

	// Check if it's a hex color
	if !strings.HasPrefix(color, "#") {
		return ""
	}

	// Convert to uppercase
	color = strings.ToUpper(color)

	// Remove transparency suffix (e.g., #10172CCC -> #10172C)
	if len(color) == 9 {
		color = color[:7]
	}

	return color
}

func checkFile(path string, validColors map[string]bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// Regex to find hex colors
	hexPattern := regexp.MustCompile(`#[0-9A-Fa-f]{6}(?:[0-9A-Fa-f]{2})?`)
	matches := hexPattern.FindAllString(content, -1)

	var undefinedColors []string
	seenColors := make(map[string]bool)

	for _, match := range matches {
		normalized := normalizeColor(match)
		if normalized == "" {
			continue
		}

		// Skip if already reported
		if seenColors[normalized] {
			continue
		}
		seenColors[normalized] = true

		// Check if color is valid
		if !validColors[normalized] {
			undefinedColors = append(undefinedColors, normalized)
		}
	}

	if len(undefinedColors) > 0 {
		return fmt.Errorf("  %s", strings.Join(undefinedColors, "\n  "))
	}

	return nil
}
