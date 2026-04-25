package cmd

import (
	"image/color"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
)

// fangFlavor mirrors the Catppuccin flavor used by internal/ui/theme.go
// (Mocha) so the CLI help output, error banners and `--help` formatting
// match the TUI palette. Keep this in sync with internal/ui/theme.go.
var fangFlavor = catppuccin.Mocha

// fangHex converts a catppuccin.Color to a color.Color usable by
// fang. We keep this local helper instead of reusing internal/ui/theme.go's
// `hex()` because that one returns lipgloss/v1 colors and fang depends on
// lipgloss/v2 (where lipgloss.Color is a constructor func, not a type).
func fangHex(c catppuccin.Color) color.Color {
	return lipgloss.Color(c.Hex)
}

// CatppuccinFangScheme returns a fang ColorScheme themed with Catppuccin
// Mocha — peach as the primary accent (titles, command names) and sky as
// the secondary accent (program name in the synopsis). The `_` ignores
// fang's light/dark dispatch because we always render against Mocha.
func CatppuccinFangScheme(_ lipgloss.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           fangHex(fangFlavor.Text()),
		Title:          fangHex(fangFlavor.Peach()),    // section headings — accent
		Description:    fangHex(fangFlavor.Subtext0()), // flag / command descriptions
		Codeblock:      fangHex(fangFlavor.Surface0()),
		Program:        fangHex(fangFlavor.Sky()),  // program name in usage line — accentAlt
		Command:        fangHex(fangFlavor.Peach()), // sub-command names
		DimmedArgument: fangHex(fangFlavor.Overlay0()),
		Comment:        fangHex(fangFlavor.Overlay0()),
		Flag:           fangHex(fangFlavor.Green()),
		FlagDefault:    fangHex(fangFlavor.Yellow()),
		Argument:       fangHex(fangFlavor.Text()),
		QuotedString:   fangHex(fangFlavor.Green()),
		Help:           fangHex(fangFlavor.Subtext0()),
		Dash:           fangHex(fangFlavor.Overlay1()),
		ErrorHeader: [2]color.Color{
			fangHex(fangFlavor.Base()), // fg on the red banner
			fangHex(fangFlavor.Red()),  // bg
		},
		ErrorDetails: fangHex(fangFlavor.Red()),
	}
}
