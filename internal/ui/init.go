// Package ui provides the user interface for the application.
package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// InitTerminalSize returns width and height, using defaults on error.
func InitTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < MIN_TERMINAL_WIDTH {
		width = MIN_TERMINAL_WIDTH
	}
	if height < MIN_TERMINAL_HEIGHT {
		height = MIN_TERMINAL_HEIGHT
	}
	return width, height
}

// InitViewport initializes the viewport component.
func InitViewport(width, height int) viewport.Model {
	vw, vh, _ := GetChatLayoutDimensions(width, height)
	vp := viewport.New(vw, vh)
	vp.Style = lipgloss.NewStyle().MarginLeft(HORIZONTAL_MARGIN).MarginRight(HORIZONTAL_MARGIN)
	return vp
}

// InitTextArea initializes the textarea component.
func InitTextArea(width, height int) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "💬 Tapez votre message ou Ctrl+R pour enregistrer..."
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.SetWidth(width)
	ta.SetHeight(height)
	return ta
}

// InitSpinner initializes the spinner component.
func InitSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ProgressBarStyle
	return s
}
