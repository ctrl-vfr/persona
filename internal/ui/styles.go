package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout constants — ratios of the terminal grid, not colours.
const (
	BORDER_SIZE       = 1
	HORIZONTAL_MARGIN = 2
	VERTICAL_MARGIN   = 1

	TITLE_HEIGHT        = 2
	STATUS_HEIGHT       = 2
	INPUT_HEIGHT        = 2
	HELP_TEXT_HEIGHT    = 1
	BOTTOM_MARGIN       = 2
	MIN_VIEWPORT_HEIGHT = 10

	MIN_TERMINAL_WIDTH  = 40
	MIN_TERMINAL_HEIGHT = 20
	MIN_MESSAGE_WIDTH   = 20
)

// Exported lipgloss styles. All colours come from theme.go; do not
// hard-code hex values here. Anything written below should reach for
// the palette aliases (`accent`, `accentAlt`, `colText`, …) so a
// flavor swap re-themes the whole UI.
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Transform(strings.ToUpper).
			MarginTop(1).
			MarginBottom(1).
			MarginLeft(2)

	ContentStyle = lipgloss.NewStyle().
			Foreground(colText).
			MarginLeft(4)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(accentAlt).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(muted).
			Italic(true)

	UserMessageStyle      = userBubbleStyle
	AssistantMessageStyle = assistantBubbleStyle
	SystemMessageStyle    = systemBubbleStyle

	SuccessStyle = successStyle
	ErrorStyle   = errorStyle
	WarningStyle = warningStyle
	InfoStyle    = infoStyle

	InputStyle = inputStyle

	ListItemStyle = lipgloss.NewStyle().
			Foreground(colText).
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(colBase).
				Background(accent).
				Bold(true).
				Padding(0, 2)

	ProgressBarStyle  = lipgloss.NewStyle().Foreground(accent)
	ProgressTextStyle = lipgloss.NewStyle().Foreground(muted)
)

// GetUserMessageStyle returns a width-aware copy of the user bubble.
func GetUserMessageStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	usableWidth := max(terminalWidth-(2*(BORDER_SIZE+HORIZONTAL_MARGIN)), MIN_MESSAGE_WIDTH)
	messageWidth := max((usableWidth*2)/3, MIN_MESSAGE_WIDTH)
	return userBubbleStyle.Width(messageWidth)
}

// GetAssistantMessageStyle returns a width-aware copy of the assistant bubble.
func GetAssistantMessageStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	usableWidth := max(terminalWidth-(2*BORDER_SIZE), MIN_MESSAGE_WIDTH)
	messageWidth := max((usableWidth*2)/3, MIN_MESSAGE_WIDTH)
	return assistantBubbleStyle.Width(messageWidth)
}

// GetStatusStyle Status message style — accentAlt (Sky) so it doesn't
// shout over the Peach pills.
func GetStatusStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	messageWidth := max(terminalWidth-(2*(BORDER_SIZE+HORIZONTAL_MARGIN)), MIN_MESSAGE_WIDTH)
	return lipgloss.NewStyle().
		Foreground(accentAlt).
		Bold(true).
		Width(messageWidth).
		Align(lipgloss.Center)
}

// GetChatBoxStyle Chat box styles
func GetChatBoxStyle(terminalWidth, terminalHeight int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	if terminalHeight < MIN_TERMINAL_HEIGHT {
		terminalHeight = MIN_TERMINAL_HEIGHT
	}
	viewportHeight := max(terminalHeight-(TITLE_HEIGHT+STATUS_HEIGHT+INPUT_HEIGHT+HELP_TEXT_HEIGHT+BOTTOM_MARGIN), MIN_VIEWPORT_HEIGHT)
	boxWidth := max(terminalWidth-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderFocus).
		Width(boxWidth).
		Height(viewportHeight)
}

// GetChatTitleStyle uses an accent foreground over the default surface.
func GetChatTitleStyle(terminalWidth int) lipgloss.Style {
	boxWidth := max(terminalWidth-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Width(boxWidth).
		Align(lipgloss.Center)
}

// GetInputBoxStyle box style
func GetInputBoxStyle(terminalWidth int) lipgloss.Style {
	boxWidth := max(max(terminalWidth, MIN_TERMINAL_WIDTH)-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentAlt).
		Width(boxWidth).
		Height(INPUT_HEIGHT).
		MarginBottom(1)
}

// RenderTitle helper functions
func RenderTitle(title string) string {
	return TitleStyle.Render(title)
}

func RenderSubtitle(subtitle string) string {
	return SubtitleStyle.Render(subtitle)
}

// Banner renderers — pill-styled status messages.

func RenderSuccess(message string) string {
	return successPill(iconSuccess, message)
}

func RenderError(message string) string {
	return errorPill(iconError, message)
}

func RenderWarning(message string) string {
	return warningPill(iconWarning, message)
}

func RenderInfo(message string) string {
	return infoPill(iconInfo, message)
}

func RenderMuted(message string) string {
	return MutedStyle.Render(message)
}

// RenderUserMessage renders a chat message from the user. Old messages
// get a clock pill prefix; the latest one stays clean.
func RenderUserMessage(message string, terminalWidth int, messageIndex int, isLatest bool) string {
	prefix := metaPill(iconUser, "you")
	if !isLatest {
		prefix = metaPill(iconClock, fmt.Sprintf("#%d", messageIndex+1)) + " " + prefix
	}
	body := GetUserMessageStyle(terminalWidth).Render(message)
	composed := prefix + "\n" + body
	return lipgloss.PlaceHorizontal(terminalWidth-(2*HORIZONTAL_MARGIN)-4, lipgloss.Right, composed)
}

// RenderAssistantMessage renders a chat message from the active persona.
// The body is passed through the Catppuccin glamour renderer (see
// markdown_style.go) so markdown in responses (lists, code, emphasis)
// gets coloured consistently with the rest of the TUI.
func RenderAssistantMessage(personaName, message string, terminalWidth int, messageIndex int, isLatest bool) string {
	prefix := accentPill(personaIcon(personaName), personaName)
	style := GetAssistantMessageStyle(terminalWidth)
	rendered := renderMarkdown(message, style.GetWidth())
	body := style.Render(rendered)
	composed := prefix + "\n" + body
	return lipgloss.PlaceHorizontal(terminalWidth-(2*HORIZONTAL_MARGIN), lipgloss.Left, composed)
}

// Status pills — peach for active states, sky for transitional, yellow
// for muted. Each is wrapped in GetStatusStyle so it centers on the
// status row.

func RenderRecordingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(accentPill(iconRecord, "Recording — speak now"))
}

func RenderTranscribingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(accentAltPill(iconTranscribe, "Transcribing"))
}

func RenderThinkingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(accentPill(iconThink, "Thinking"))
}

func RenderGeneratingAudioStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(accentAltPill(iconSpeak, "Generating audio"))
}

func RenderPlayingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(accentPill(iconPlaying, "Playing"))
}

func RenderMutedStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render(mutedPill(iconMute, "Silent mode on"))
}

// RenderMessageWithSeparator adds a low-contrast separator between
// stacked chat messages so they don't visually merge.
func RenderMessageWithSeparator(message string, isLast bool) string {
	if isLast {
		return message
	}
	separator := MutedStyle.Render("·····")
	return message + "\n\n" + lipgloss.PlaceHorizontal(80, lipgloss.Center, separator)
}

func RenderMessageSpacing() string {
	return "\n"
}

// RenderChatBoxTitle wraps the title in an accent pill on the left and
// pads the rest of the row with a thin separator. Cleaner than the
// previous block-fill (▓) approach.
func RenderChatBoxTitle(title string, terminalWidth int) string {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	pill := accentPill(iconAssistant, title)
	pillWidth := lipgloss.Width(pill)
	available := max(terminalWidth-(2*BORDER_SIZE)-pillWidth-2, 0)
	separator := lipgloss.NewStyle().Foreground(colSurface1).Render(strings.Repeat("─", available))
	return pill + " " + separator
}

func RenderChatBoxBorder(content string, terminalWidth, terminalHeight int) string {
	return GetChatBoxStyle(terminalWidth, terminalHeight).Render(content)
}

func RenderInputBox(content string, terminalWidth int) string {
	return GetInputBoxStyle(terminalWidth).Render(content)
}

// GetChatLayoutDimensions Responsive layout helper
func GetChatLayoutDimensions(terminalWidth, terminalHeight int) (viewportWidth, viewportHeight, inputHeight int) {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	if terminalHeight < MIN_TERMINAL_HEIGHT {
		terminalHeight = MIN_TERMINAL_HEIGHT
	}
	viewportWidth = max(terminalWidth-(2*BORDER_SIZE), MIN_MESSAGE_WIDTH)
	viewportHeight = max(terminalHeight-(TITLE_HEIGHT+STATUS_HEIGHT+INPUT_HEIGHT+HELP_TEXT_HEIGHT+BOTTOM_MARGIN), MIN_VIEWPORT_HEIGHT)
	inputHeight = INPUT_HEIGHT
	return viewportWidth, viewportHeight, inputHeight
}
