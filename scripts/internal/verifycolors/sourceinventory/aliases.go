package sourceinventory

import (
	"fmt"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

var nvimAliases = map[string]verifycolors.TokenRef{
	"bg":                    "core.background",
	"darkest_bg":            "core.darkest_bg",
	"panel_bg":              "core.panel_bg",
	"dark_shadow":           "core.ui_shadow",
	"gutter_bg":             "nvim.gutter_bg",
	"border":                "teals.border",
	"mid_gray":              "blues_slates.slate_mid",
	"subdued_fg":            "foregrounds.subdued",
	"light_gray":            "foregrounds.dim",
	"punctuation_gray":      "blues_slates.punctuation_gray",
	"comment_gray":          "blues_slates.comment_gray",
	"git_blame_gray":        "blues_slates.git_blame_gray",
	"operator":              "teals.mid_bright",
	"fg":                    "foregrounds.main",
	"near_white":            "foregrounds.bright",
	"highlight_white":       "ansi.bright_white",
	"cyan":                  "teals.bright",
	"bright_cyan":           "foregrounds.heading",
	"magenta":               "zsh.error",
	"bright_magenta":        "blues_slates.cloud_slate",
	"purple_accent":         "purples.muted_purple",
	"syntax_violet":         "semantic.keyword",
	"success":               "semantic.success",
	"ocean_blue":            "blues_slates.ocean_blue",
	"lavender":              "purples.lavender",
	"ansi_black":            "ansi.black",
	"ansi_red":              "ansi.red",
	"ansi_green":            "ansi.green",
	"ansi_yellow":           "ansi.yellow",
	"ansi_blue":             "ansi.blue",
	"ansi_magenta":          "ansi.magenta",
	"ansi_cyan":             "ansi.cyan",
	"ansi_white":            "ansi.white",
	"ansi_bright_black":     "ansi.bright_black",
	"ansi_bright_red":       "ansi.bright_red",
	"ansi_bright_green":     "ansi.bright_green",
	"ansi_bright_yellow":    "ansi.bright_yellow",
	"ansi_bright_blue":      "ansi.bright_blue",
	"ansi_bright_magenta":   "ansi.bright_magenta",
	"ansi_bright_cyan":      "ansi.bright_cyan",
	"ansi_bright_white":     "ansi.bright_white",
	"white":                 "blues_slates.sky_slate",
	"selection":             "core.selection_bg",
	"selection_fg":          "core.selection_fg",
	"string":                "semantic.string",
	"diff_add_bg":           "nvim.diff_add_bg",
	"diff_change_bg":        "nvim.diff_change_bg",
	"diff_delete_bg":        "nvim.diff_delete_bg",
	"diff_add_inline_bg":    "nvim.diff_add_inline_bg",
	"diff_change_inline_bg": "nvim.diff_change_inline_bg",
	"diff_delete_inline_bg": "nvim.diff_delete_inline_bg",
}

func resolveNvimAlias(alias string) (verifycolors.TokenRef, error) {
	token, ok := nvimAliases[alias]
	if !ok {
		return "", fmt.Errorf("unknown Neovim palette alias %q", alias)
	}
	return token, nil
}
