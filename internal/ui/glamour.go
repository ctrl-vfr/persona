package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// RenderPersonaListBox renders persona list in a chat-style box
func RenderPersonaListBox(personas []string, terminalWidth int) string {
	var lines []string

	// Build the persona list content
	for _, persona := range personas {
		lines = append(lines, fmt.Sprintf("  🤖 %s", persona))
	}

	content := strings.Join(lines, "\n")

	// Use the same style as chat boxes
	return RenderChatBoxBorder(content, terminalWidth, len(lines)+4)
}

// RenderHelpBox renders help content in a chat-style box
func RenderHelpBox(title, content string, terminalWidth int) string {
	// Format content with title
	formattedContent := fmt.Sprintf("%s\n\n%s", RenderInfo(title), content)

	// Calculate appropriate height based on content
	lines := strings.Split(formattedContent, "\n")
	height := len(lines) + 4 // Add padding

	return RenderChatBoxBorder(formattedContent, terminalWidth, height)
}

// RenderPersonaDetails renders persona details in a chat-style box
func RenderPersonaDetails(persona, voice, instructions, prompt string, historyCount int, terminalWidth int) string {
	content := fmt.Sprintf(`%s %s

%s
%s

%s
%s

%s
%s

%s %d`,
		RenderInfo("Nom:"), persona,
		RenderInfo("Voix:"), voice,
		RenderInfo("Instructions vocales:"), instructions,
		RenderInfo("Prompt système:"), prompt,
		RenderInfo("Messages dans l'historique:"), historyCount)

	// Calculate height based on content
	lines := strings.Split(content, "\n")
	height := len(lines) + 4

	return RenderChatBoxBorder(content, terminalWidth, height)
}

// GetTerminalWidth returns the current terminal width with fallback
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		width = MIN_TERMINAL_WIDTH
	}
	if width < MIN_TERMINAL_WIDTH {
		width = MIN_TERMINAL_WIDTH
	}
	return width
}
