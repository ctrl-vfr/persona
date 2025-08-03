package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout constants
const (
	// Borders and padding
	BORDER_SIZE       = 1
	HORIZONTAL_MARGIN = 2
	VERTICAL_MARGIN   = 1

	// Reserved space
	TITLE_HEIGHT        = 2  // Title + border
	STATUS_HEIGHT       = 2  // Status message
	INPUT_HEIGHT        = 2  // Input box + border
	HELP_TEXT_HEIGHT    = 1  // Help text
	BOTTOM_MARGIN       = 2  // Extra margin at bottom
	MIN_VIEWPORT_HEIGHT = 10 // Minimum height for chat viewport

	// Minimum dimensions
	MIN_TERMINAL_WIDTH  = 40
	MIN_TERMINAL_HEIGHT = 20
	MIN_MESSAGE_WIDTH   = 20
)

// Theme colors
var (
	PrimaryColor    = lipgloss.Color("#FF6B9D")
	SecondaryColor  = lipgloss.Color("#B8E6B8")
	AccentColor     = lipgloss.Color("#FFE66D")
	BackgroundColor = lipgloss.Color("#1A1A2E")
	TextColor       = lipgloss.Color("#FFFFFF")
	MutedColor      = lipgloss.Color("#8E8E93")
	ErrorColor      = lipgloss.Color("#FF4757")
	SuccessColor    = lipgloss.Color("#2ECC71")
	WarningColor    = lipgloss.Color("#F39C12")
	InfoColor       = lipgloss.Color("#3498DB")

	TitleColor   = lipgloss.Color("#6143df")
	ContentColor = lipgloss.Color("#ff79d0")
	// Message background colors (improved for better contrast)
	UserBgColor      = lipgloss.Color("#2563eb") // Modern blue
	AssistantBgColor = lipgloss.Color("#6143df") // Modern purple
	SystemBgColor    = lipgloss.Color("#6b7280") // Neutral gray
)

// ClockEmojis for message age indication
var ClockEmojis = []string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕘", "🕙"}

// Common styles
var (
	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(TitleColor).
			Bold(true).Transform(strings.ToUpper).MarginBottom(1).MarginTop(1).MarginLeft(2)

	ContentStyle = lipgloss.NewStyle().
			Foreground(ContentColor).MarginLeft(4)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			MarginBottom(0)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	// Message styles
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentColor).
				BorderLeft(true)

	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SecondaryColor).
				BorderLeft(true)

	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				Italic(true)
	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(InfoColor).
			Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			MarginBottom(1)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(BackgroundColor).
				Background(PrimaryColor).
				Bold(true).
				Padding(0, 2)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(MutedColor)
)

// GetUserMessageStyle styles with proper width
func GetUserMessageStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}

	// Calculate usable width (terminal width minus borders and margins)
	usableWidth := max(terminalWidth-(2*(BORDER_SIZE+HORIZONTAL_MARGIN)), MIN_MESSAGE_WIDTH)

	// messageWidth := (usableWidth * 2) / 3 // 2/3 of usable width
	messageWidth := max((usableWidth*2)/3, MIN_MESSAGE_WIDTH) // 2/3 of usable width

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(UserBgColor).
		Padding(1, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(UserBgColor).
		Width(messageWidth)
}

func GetAssistantMessageStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}

	// usableWidth := terminalWidth - (2 * (BORDER_SIZE))
	usableWidth := max(terminalWidth-(2*(BORDER_SIZE)), MIN_MESSAGE_WIDTH)

	// messageWidth := (usableWidth * 2) / 3
	messageWidth := max((usableWidth*2)/3, MIN_MESSAGE_WIDTH)

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(AssistantBgColor).
		Padding(1, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AssistantBgColor).
		Width(messageWidth)
}

// GetStatusStyle Status message style
func GetStatusStyle(terminalWidth int) lipgloss.Style {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}

	messageWidth := max(terminalWidth-(2*(BORDER_SIZE+HORIZONTAL_MARGIN)), MIN_MESSAGE_WIDTH)

	return lipgloss.NewStyle().
		Foreground(InfoColor).
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

	// Calculate viewport height (minimum height for chat viewport)
	viewportHeight := max(terminalHeight-(TITLE_HEIGHT+STATUS_HEIGHT+INPUT_HEIGHT+HELP_TEXT_HEIGHT+BOTTOM_MARGIN), MIN_VIEWPORT_HEIGHT)
	// Calculate box width
	boxWidth := max(terminalWidth-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Width(boxWidth).
		Height(viewportHeight)
}

func GetChatTitleStyle(terminalWidth int) lipgloss.Style {
	boxWidth := max(terminalWidth-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Background(BackgroundColor).
		Width(boxWidth).
		Align(lipgloss.Center)
}

// GetInputBoxStyle box style
func GetInputBoxStyle(terminalWidth int) lipgloss.Style {
	boxWidth := max(max(terminalWidth, MIN_TERMINAL_WIDTH)-(2*BORDER_SIZE), MIN_TERMINAL_WIDTH)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
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

func RenderSuccess(message string) string {
	return SuccessStyle.Render("✅ " + message)
}

func RenderError(message string) string {
	return ErrorStyle.Render("❌ " + message)
}

func RenderWarning(message string) string {
	return WarningStyle.Render("⚠️  " + message)
}

func RenderInfo(message string) string {
	return InfoStyle.Render("ℹ️  " + message)
}

// RenderUserMessage message renderers with PlaceHorizontal:
func RenderUserMessage(message string, terminalWidth int, messageIndex int, isLatest bool) string {
	// Add subtle time indicator for older messages
	timeEmoji := ""
	if !isLatest && messageIndex < len(ClockEmojis) {
		timeEmoji = ClockEmojis[messageIndex%len(ClockEmojis)] + " "
	}

	// Format content with improved emoji and spacing
	content := fmt.Sprintf("%s👤 %s", timeEmoji, message)
	styledMessage := GetUserMessageStyle(terminalWidth).Render(content)

	// Right-align user messages for chat-like appearance
	return lipgloss.PlaceHorizontal(terminalWidth-(2*HORIZONTAL_MARGIN)-4, lipgloss.Right, styledMessage)
}

func RenderAssistantMessage(personaName, message string, terminalWidth int, messageIndex int, isLatest bool) string {
	// Add persona indicator and format content
	content := fmt.Sprintf("🤖 %s", message)
	styledMessage := GetAssistantMessageStyle(terminalWidth).Render(content)

	// Left-align assistant messages
	return lipgloss.PlaceHorizontal(terminalWidth-(2*HORIZONTAL_MARGIN), lipgloss.Left, styledMessage)
}

// RenderRecordingStatus Status messages with animated emojis
func RenderRecordingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("🎤 🔴 Enregistrement en cours... Parlez maintenant!")
}

// RenderTranscribingStatus Status messages with animated emojis
func RenderTranscribingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("📝 ✍️  Transcription en cours...")
}

// RenderThinkingStatus Status messages with animated emojis
func RenderThinkingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("🤔 💭 Réflexion en cours...")
}

// RenderGeneratingAudioStatus Status messages with animated emojis
func RenderGeneratingAudioStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("🎵 🔊 Génération audio en cours...")
}

// RenderPlayingStatus Status messages with animated emojis
func RenderPlayingStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("🔈 🎶 Lecture en cours...")
}

// RenderMutedStatus Status messages with animated emojis
func RenderMutedStatus(terminalWidth int) string {
	return GetStatusStyle(terminalWidth).Render("🔇 Mode silencieux activé")
}

// RenderMuted helper function
func RenderMuted(message string) string {
	return MutedStyle.Render(message)
}

// RenderMessageWithSeparator Enhanced message rendering with better visual separation
func RenderMessageWithSeparator(message string, isLast bool) string {
	if isLast {
		return message
	}
	// Add subtle separator between messages
	separator := MutedStyle.Render("─────")
	return message + "\n\n" + lipgloss.PlaceHorizontal(80, lipgloss.Center, separator)
}

// Helper function to add visual spacing between message groups
func RenderMessageSpacing() string {
	return "\n"
}

// RenderChatBoxTitle Chat box with decorative border
func RenderChatBoxTitle(title string, terminalWidth int) string {
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}

	// Calculate available width for decorative characters
	titleLength := len(title)
	availableWidth := max(terminalWidth-(2*BORDER_SIZE)-titleLength-4, MIN_TERMINAL_WIDTH) // -4 for spaces around title

	// Create decorative fill
	leftPadding := availableWidth / 2
	rightPadding := availableWidth - leftPadding

	var titleContent string
	if leftPadding > 0 {
		leftFill := strings.Repeat("▓", leftPadding)
		rightFill := strings.Repeat("▓", rightPadding)
		titleContent = leftFill + " " + title + " " + rightFill
	} else {
		titleContent = " " + title + " "
	}

	return GetChatTitleStyle(terminalWidth).Render(titleContent)
}

func RenderChatBoxBorder(content string, terminalWidth, terminalHeight int) string {
	boxStyle := GetChatBoxStyle(terminalWidth, terminalHeight)
	return boxStyle.Render(content)
}

func RenderInputBox(content string, terminalWidth int) string {
	inputStyle := GetInputBoxStyle(terminalWidth)
	return inputStyle.Render(content)
}

// GetChatLayoutDimensions Responsive layout helper
func GetChatLayoutDimensions(terminalWidth, terminalHeight int) (viewportWidth, viewportHeight, inputHeight int) {
	// Ensure minimum dimensions
	if terminalWidth < MIN_TERMINAL_WIDTH {
		terminalWidth = MIN_TERMINAL_WIDTH
	}
	if terminalHeight < MIN_TERMINAL_HEIGHT {
		terminalHeight = MIN_TERMINAL_HEIGHT
	}

	// Calculate usable width (terminal width minus borders and margins)
	viewportWidth = max(terminalWidth-(2*(BORDER_SIZE)), MIN_MESSAGE_WIDTH)

	// Calculate viewport height
	viewportHeight = max(terminalHeight-(TITLE_HEIGHT+STATUS_HEIGHT+INPUT_HEIGHT+HELP_TEXT_HEIGHT+BOTTOM_MARGIN), MIN_VIEWPORT_HEIGHT)

	// Input height is fixed
	inputHeight = INPUT_HEIGHT // Account for input box borders

	return viewportWidth, viewportHeight, inputHeight
}
