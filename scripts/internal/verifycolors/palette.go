package verifycolors

import (
	"fmt"
	"sort"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/xterm"
)

type Palette struct {
	Metadata    map[string]string `toml:"metadata"`
	Core        map[string]string `toml:"core"`
	Foregrounds map[string]string `toml:"foregrounds"`
	Teals       map[string]string `toml:"teals"`
	BluesSlates map[string]string `toml:"blues_slates"`
	Purples     map[string]string `toml:"purples"`
	Semantic    map[string]string `toml:"semantic"`
	Git         map[string]string `toml:"git"`
	UI          map[string]string `toml:"ui"`
	ANSI        map[string]string `toml:"ansi"`
	WezTerm     map[string]string `toml:"wezterm"`
	Nvim        map[string]string `toml:"nvim"`
	Zsh         map[string]string `toml:"zsh"`
}

func LoadPalette(path string) (*Palette, error) {
	loaded, err := palette.Load[Palette](path)
	if err != nil {
		return nil, err
	}
	if err := loaded.ValidateColors(); err != nil {
		return nil, err
	}
	return loaded, nil
}

func (p *Palette) TokenValues() map[TokenRef]string {
	values := make(map[TokenRef]string)
	for section, tokens := range p.sections() {
		for name, value := range tokens {
			values[TokenRef(section+"."+name)] = value
		}
	}
	return values
}

func (p *Palette) SortedTokens() []TokenRef {
	values := p.TokenValues()
	tokens := make([]TokenRef, 0, len(values))
	for token := range values {
		tokens = append(tokens, token)
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i] < tokens[j] })
	return tokens
}

func (p *Palette) ValidateColors() error {
	for token, value := range p.TokenValues() {
		if _, _, _, err := colorutil.ParseHexRGB8(value); err != nil {
			return fmt.Errorf("palette token %s: %w", token, err)
		}
	}
	return nil
}

func (p *Palette) WezTermANSI() ([16]xterm.RGB, error) {
	names := []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"bright_black", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	}
	var ansi [16]xterm.RGB
	for index, name := range names {
		value, ok := p.ANSI[name]
		if !ok {
			return ansi, fmt.Errorf("palette is missing ansi.%s", name)
		}
		r, g, b, err := colorutil.ParseHexRGB8(value)
		if err != nil {
			return ansi, fmt.Errorf("palette token ansi.%s: %w", name, err)
		}
		ansi[index] = xterm.RGB{R: r, G: g, B: b}
	}
	return ansi, nil
}

func (p *Palette) sections() map[string]map[string]string {
	return map[string]map[string]string{
		"core":         p.Core,
		"foregrounds":  p.Foregrounds,
		"teals":        p.Teals,
		"blues_slates": p.BluesSlates,
		"purples":      p.Purples,
		"semantic":     p.Semantic,
		"git":          p.Git,
		"ui":           p.UI,
		"ansi":         p.ANSI,
		"wezterm":      p.WezTerm,
		"nvim":         p.Nvim,
		"zsh":          p.Zsh,
	}
}
