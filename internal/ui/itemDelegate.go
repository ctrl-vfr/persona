package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type itemDelegate struct {
	width int
}

// Height: 2 lines per item (title + description). The selected item's
// rounded border adds 2 extra lines on top of these, so we expose 4
// here to keep the row pitch consistent across selected / unselected.
func (d itemDelegate) Height() int {
	return 4
}

// Spacing: 0 — the 4-line height already includes vertical breathing
// room for the unselected items (they fill 2 lines + 2 blank). Adding
// extra spacing on top stretches the list past the body height when
// there are more than ~5 personas.
func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(PersonaItem)
	if !ok {
		return
	}

	title := item.Title()
	description := item.Description()
	contentWidth := d.width - 10

	selected := index == m.Index()

	var containerStyle lipgloss.Style
	if selected {
		containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderFocus).
			Width(d.width - 8)
	} else {
		containerStyle = lipgloss.NewStyle().Width(d.width)
	}

	titleFg := colText
	if selected {
		titleFg = accent
	}
	titleLine := lipgloss.NewStyle().
		Foreground(titleFg).
		Bold(true).
		Width(contentWidth).
		Render(personaIcon(title) + "  " + title)

	descLine := lipgloss.NewStyle().
		Foreground(muted).
		Italic(true).
		Width(contentWidth).
		Render(description)

	fmt.Fprint(w, containerStyle.Render(lipgloss.JoinVertical(lipgloss.Left, titleLine, descLine)))
}
