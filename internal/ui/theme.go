package ui

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha palette — single source of truth for all colors.
// Changing flavor here re-themes the whole TUI in one swap.
var flavor = catppuccin.Mocha

func hex(c catppuccin.Color) lipgloss.Color { return lipgloss.Color(c.Hex) }

var (
	colBase     = hex(flavor.Base())
	colMantle   = hex(flavor.Mantle())
	colCrust    = hex(flavor.Crust())
	colSurface0 = hex(flavor.Surface0())
	colSurface1 = hex(flavor.Surface1())
	colSurface2 = hex(flavor.Surface2())
	colOverlay0 = hex(flavor.Overlay0())
	colOverlay1 = hex(flavor.Overlay1())
	colOverlay2 = hex(flavor.Overlay2())
	colSubtext0 = hex(flavor.Subtext0())
	colSubtext1 = hex(flavor.Subtext1())
	colText     = hex(flavor.Text())
	colMauve    = hex(flavor.Mauve())
	colLavender = hex(flavor.Lavender())
	colBlue     = hex(flavor.Blue())
	colSapphire = hex(flavor.Sapphire())
	colSky      = hex(flavor.Sky())
	colTeal     = hex(flavor.Teal())
	colGreen    = hex(flavor.Green())
	colYellow   = hex(flavor.Yellow())
	colPeach    = hex(flavor.Peach())
	colRed      = hex(flavor.Red())
	colPink     = hex(flavor.Pink())
)

// Semantic aliases — let screens reach for intent, not raw color names.
// Swap accent / accentAlt here and every pill / border / spinner follows.
var (
	accent      = colPeach
	accentAlt   = colSky
	muted       = colOverlay0
	borderBlur  = colSurface1
	borderFocus = colPeach
)

// Reusable lipgloss styles. Keep raw colour usage out of the rest of
// the package: every site should pick one of these (or build via the
// pill helpers below) so a palette change does not require chasing hex
// codes through the codebase.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Padding(0, 1)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentAlt).
			Padding(0, 1)

	subtleStyle = lipgloss.NewStyle().Foreground(colSubtext0)
	mutedStyle  = lipgloss.NewStyle().Foreground(colOverlay0).Italic(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colSubtext0).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colRed).
			Bold(true).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(colGreen).
			Bold(true).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(colSky).
			Bold(true).
			Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
			Foreground(colYellow).
			Bold(true).
			Padding(0, 1)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderFocus).
				Padding(0, 1)

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderBlur).
				Padding(0, 1)

	// outerFrameStyle wraps a whole screen (the persona selector, for
	// instance) in a subtle rounded frame. Surface2 is dark grey,
	// quiet enough to not compete with the inner peach pane.
	outerFrameStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colSurface2).
				Padding(0, 1)

	// Conversation bubbles. User goes Sky (accentAlt) so it does not
	// fight with the Peach status pills; Assistant goes Peach to mark
	// the active speaker.
	userBubbleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentAlt).
			Foreground(colText).
			Padding(0, 1)

	assistantBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderFocus).
				Foreground(colText).
				Padding(0, 1)

	systemBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colOverlay1).
				Foreground(colSubtext0).
				Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentAlt).
			Padding(0, 1)
)

// chevronPill wraps text with Powerline rounded chevrons so it renders
// as a capsule. bg is the pill fill (also the foreground of the two
// caps); fg is the text/icon colour painted on top of the fill. Bold is
// opt-in because semantic pills (success/error) want it while neutral
// meta pills do not.
func chevronPill(bg, fg lipgloss.Color, bold bool, content string) string {
	cap := lipgloss.NewStyle().Foreground(bg)
	body := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(bold).
		Render(" " + content + " ")
	return cap.Render(chevronLeft) + body + cap.Render(chevronRight)
}

// pillContent joins an icon and text with a double space, omitting any
// space when icon is empty. Two spaces give Nerd Font glyphs visual
// breathing room — many of them advance the cursor by less than a full
// cell so a single space looks cramped against the next character.
func pillContent(icon, text string) string {
	if icon == "" {
		return text
	}
	return icon + "  " + text
}

// accentPill — primary accent (Peach) capsule; titles, active states.
func accentPill(icon, text string) string {
	if text == "" {
		return ""
	}
	return chevronPill(accent, colBase, true, pillContent(icon, text))
}

// accentAltPill — secondary accent (Sky) capsule; brand, neutrals.
func accentAltPill(icon, text string) string {
	if text == "" {
		return ""
	}
	return chevronPill(accentAlt, colBase, true, pillContent(icon, text))
}

// metaPill — neutral grey pill for inert metadata.
func metaPill(icon, text string) string {
	if text == "" {
		return ""
	}
	return chevronPill(colSurface0, colText, false, pillContent(icon, text))
}

// successPill / errorPill / pendingPill / warningPill / infoPill —
// semantic status capsules.
func successPill(icon, text string) string {
	return chevronPill(colGreen, colBase, true, pillContent(icon, text))
}

func errorPill(icon, text string) string {
	return chevronPill(colRed, colBase, true, pillContent(icon, text))
}

func pendingPill(icon, text string) string {
	return chevronPill(colYellow, colBase, true, pillContent(icon, text))
}

func warningPill(icon, text string) string {
	return chevronPill(colYellow, colBase, true, pillContent(icon, text))
}

func infoPill(icon, text string) string {
	return chevronPill(colSky, colBase, true, pillContent(icon, text))
}

// mutedPill — surface0 background with a low-contrast text colour;
// for "Silent mode", "(no content)" markers.
func mutedPill(icon, text string) string {
	if text == "" {
		return ""
	}
	return chevronPill(colSurface0, colOverlay1, false, pillContent(icon, text))
}

// inactiveTabPill — surface1 capsule, kept for parity with EV-CLI even
// if persona has no tab bar yet.
func inactiveTabPill(icon, text string) string {
	return chevronPill(colSurface1, colOverlay1, false, pillContent(icon, text))
}

// listItemStyles applies the catppuccin palette to the default list
// item delegate, used by the persona selector list.
func listItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles()

	s.NormalTitle = s.NormalTitle.Foreground(colText)
	s.NormalDesc = s.NormalDesc.Foreground(colOverlay1)

	s.SelectedTitle = s.SelectedTitle.
		Foreground(accent).
		BorderForeground(accent).
		Bold(true)
	s.SelectedDesc = s.SelectedDesc.
		Foreground(accentAlt).
		BorderForeground(accent)

	s.DimmedTitle = s.DimmedTitle.Foreground(colOverlay1)
	s.DimmedDesc = s.DimmedDesc.Foreground(colOverlay0)

	s.FilterMatch = s.FilterMatch.Foreground(accent).Underline(true)

	return s
}
