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

func (d itemDelegate) Height() int {
	return 4
}

func (d itemDelegate) Spacing() int {
	return 1
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

	contentWidth := d.width - 10 // -6 pour les bordures et padding

	var containerStyle lipgloss.Style
	if index == m.Index() {
		containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B9D")).
			Width(d.width - 8)
	} else {
		containerStyle = lipgloss.NewStyle().
			Width(d.width)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6143df")).
		Bold(true).
		Width(contentWidth)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8E8E93")).
		Italic(true).
		Width(contentWidth)

	styledTitle := titleStyle.Render("🤖 " + title)
	styledDesc := descStyle.Render(description)
	content := lipgloss.JoinVertical(lipgloss.Left, styledTitle, styledDesc)

	_, err := fmt.Fprint(w, containerStyle.Render(content))
	if err != nil {
		fmt.Println(err)
	}
}
