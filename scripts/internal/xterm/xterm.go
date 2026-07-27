package xterm

import (
	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

type RGB struct {
	R int
	G int
	B int
}

// Nearest256 returns both the xterm-256 index and the RGB value represented by
// that index.
func Nearest256(value string) (int, RGB, error) {
	r, g, b, err := colorutil.ParseHexRGB8(value)
	if err != nil {
		return 0, RGB{}, err
	}
	palette := Palette256()
	bestIndex := 0
	bestDistance := int(^uint(0) >> 1)
	for index, candidate := range palette {
		dr := r - candidate.R
		dg := g - candidate.G
		db := b - candidate.B
		distance := dr*dr + dg*dg + db*db
		if distance < bestDistance {
			bestDistance = distance
			bestIndex = index
		}
	}
	return bestIndex, palette[bestIndex], nil
}

func Palette256() []RGB {
	palette := []RGB{
		{0, 0, 0}, {128, 0, 0}, {0, 128, 0}, {128, 128, 0},
		{0, 0, 128}, {128, 0, 128}, {0, 128, 128}, {192, 192, 192},
		{128, 128, 128}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
		{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
	}
	levels := []int{0, 95, 135, 175, 215, 255}
	for _, r := range levels {
		for _, g := range levels {
			for _, b := range levels {
				palette = append(palette, RGB{r, g, b})
			}
		}
	}
	for gray := 8; gray <= 238; gray += 10 {
		palette = append(palette, RGB{gray, gray, gray})
	}
	return palette
}
