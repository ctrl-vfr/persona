package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ctrl-vfr/persona/internal/config"
	"github.com/ctrl-vfr/persona/internal/ffmpeg"
	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/persona"
	"github.com/ctrl-vfr/persona/internal/speak"
	"github.com/ctrl-vfr/persona/internal/storage"
	"github.com/ctrl-vfr/persona/internal/watcher"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type ChatState int

const (
	StateIdle ChatState = iota
	StateRecording
	StateTranscribing
	StateChatting
	StateGeneratingAudio
	StatePlaying
	StateError
)

type AppMode int

const (
	ModePersonaSelector AppMode = iota
	ModeChat
)

// PersonaItem pour la liste des personas
type PersonaItem struct {
	name        string
	description string
}

func (i PersonaItem) FilterValue() string { return i.name }
func (i PersonaItem) Title() string       { return i.name }
func (i PersonaItem) Description() string { return i.description }

type ChatModel struct {
	// Application mode
	mode AppMode

	// UI components
	viewport    viewport.Model
	textArea    textarea.Model
	spinner     spinner.Model
	personaList list.Model

	// Application state
	state   ChatState
	persona *persona.Persona
	ai      *openai.OpenAI
	manager *storage.Manager
	config  *config.Config

	// Configuration for multi-mode support
	openaiAPIKey string

	// File watching
	personaWatcher  *watcher.PersonaWatcher
	instanceManager *watcher.InstanceManager

	// Persona selector preview cache: name -> loaded persona. Filled
	// once at startup so the preview pane re-renders without disk I/O
	// on every keystroke.
	previewCache map[string]*persona.Persona

	// itemDelegate is kept here so we can re-bind it with the current
	// pane width on every selector render. list.Model owns the live
	// delegate as an unexported field, so we can't read it back.
	listDelegate itemDelegate

	// Display state
	messages  []string
	statusMsg string
	errorMsg  string
	width     int
	height    int

	// Configuration
	inputDevice      string
	silenceThreshold int
	silenceDuration  int

	// Audio settings
	isMuted bool

	// Lifecycle context cancelled in Cleanup. Used to abort in-flight
	// OpenAI requests when the user quits the TUI.
	rootCtx    context.Context
	cancelRoot context.CancelFunc
}

// ctx returns the model's lifecycle context. Created lazily so models
// constructed without explicit context (older callers, tests) still work.
func (m *ChatModel) ctx() context.Context {
	if m.rootCtx == nil {
		m.rootCtx, m.cancelRoot = context.WithCancel(context.Background())
	}
	return m.rootCtx
}

// Message types for async operations
type recordingFinishedMsg struct {
	filename string
	err      error
}

type transcriptionFinishedMsg struct {
	text string
	err  error
}

type chatFinishedMsg struct {
	response string
	err      error
}

type audioFinishedMsg struct {
	audioData []byte
	err       error
}

type historyUpdateMsg struct {
	history []persona.Message
}

type personaUpdateMsg struct {
	persona *persona.Persona
}

func NewChatModel(p *persona.Persona, ai *openai.OpenAI, manager *storage.Manager, inputDevice string, silenceThreshold int, silenceDuration int) *ChatModel {
	// Get terminal size with fallback to minimum dimensions
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 || height == 0 {
		width = MIN_TERMINAL_WIDTH
		height = MIN_TERMINAL_HEIGHT
	}

	width = max(width, MIN_TERMINAL_WIDTH)
	height = max(height, MIN_TERMINAL_HEIGHT)

	// Calculate responsive dimensions
	viewportWidth, viewportHeight, inputHeight := GetChatLayoutDimensions(width, height)

	// Initialize viewport with margins
	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = lipgloss.NewStyle().MarginLeft(HORIZONTAL_MARGIN).MarginRight(HORIZONTAL_MARGIN)

	// Initialize text area
	ta := textarea.New()
	ta.Placeholder = "💬 Tapez votre message ou Ctrl+R pour enregistrer..."
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.SetWidth(viewportWidth)
	ta.SetHeight(inputHeight)

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ProgressBarStyle

	model := &ChatModel{
		viewport:         vp,
		textArea:         ta,
		spinner:          s,
		state:            StateIdle,
		persona:          p,
		ai:               ai,
		manager:          manager,
		messages:         []string{},
		width:            width,
		height:           height,
		inputDevice:      inputDevice,
		silenceThreshold: silenceThreshold,
		silenceDuration:  silenceDuration,
		isMuted:          false,
	}

	// Initialize file watcher
	if personaWatcher, err := watcher.NewPersonaWatcher(manager, p.Name); err == nil {
		model.personaWatcher = personaWatcher
		personaWatcher.Start(model.ctx())
	}

	// Initialize instance manager
	model.instanceManager = watcher.NewInstanceManager(manager)
	if err := model.instanceManager.RegisterInstance(); err == nil {
		model.instanceManager.StartHeartbeat(model.ctx())
	}

	// Add initial greeting with responsive styling
	model.addWelcomeMessage()
	model.addHistoryMessages()

	return model
}

func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle common messages first
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	// Handle mode-specific updates
	switch m.mode {
	case ModePersonaSelector:
		return m.updatePersonaSelector(msg)
	case ModeChat:
		return m.updateChat(msg)
	default:
		return m, nil
	}
}

func (m *ChatModel) updatePersonaSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(msg.Width, MIN_TERMINAL_WIDTH)
		m.height = max(msg.Height, MIN_TERMINAL_HEIGHT)
		// Real list-pane dims are recomputed in viewPersonaSelector;
		// the SetSize call here just primes the list with reasonable
		// defaults so the first paint after resize is not blank.
		m.personaList.SetSize(max(m.width*40/100, 24), max(m.height-4, MIN_VIEWPORT_HEIGHT))

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Get selected persona
			if selectedItem := m.personaList.SelectedItem(); selectedItem != nil {
				if persona, ok := selectedItem.(PersonaItem); ok {
					err := m.SwitchToPersona(persona.name)
					if err != nil {
						m.errorMsg = fmt.Sprintf("Error changing persona: %v", err)
						return m, nil
					}
					return m, nil
				}
			}
		case "ctrl+s":
			// Toggle between persona selector and current chat
			if m.persona != nil {
				m.mode = ModeChat
				return m, nil
			}
		}
	}

	// Update persona list
	m.personaList, cmd = m.personaList.Update(msg)
	return m, cmd
}

func (m *ChatModel) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Ensure minimum dimensions
		m.width = msg.Width
		m.height = msg.Height
		if m.width < MIN_TERMINAL_WIDTH {
			m.width = MIN_TERMINAL_WIDTH
		}
		if m.height < MIN_TERMINAL_HEIGHT {
			m.height = MIN_TERMINAL_HEIGHT
		}

		// Recalculate responsive dimensions
		viewportWidth, viewportHeight, inputHeight := GetChatLayoutDimensions(m.width, m.height)

		// Update viewport
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight
		m.viewport.Style = lipgloss.NewStyle().
			MarginLeft(HORIZONTAL_MARGIN).
			MarginRight(HORIZONTAL_MARGIN)

		// Update text area
		m.textArea.SetWidth(viewportWidth)
		m.textArea.SetHeight(inputHeight)

		// Re-render all messages with new width
		m.reRenderMessages()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+l":
			// Clear conversation
			if m.state == StateIdle {
				m.clearConversation()
			}
		case "ctrl+m":
			// Toggle mute
			m.isMuted = !m.isMuted
			if m.isMuted {
				m.statusMsg = RenderMutedStatus(m.width)
			} else {
				m.statusMsg = ""
			}
		case "ctrl+s":
			// Switch back to persona selector
			m.mode = ModePersonaSelector
			return m, nil
		case "ctrl+r":
			if m.state == StateIdle {
				return m, m.startRecording()
			}
		case "enter":
			if m.state == StateIdle && m.textArea.Value() != "" {
				userMessage := strings.TrimSpace(m.textArea.Value())
				m.textArea.Reset()
				return m, m.sendTextMessage(userMessage)
			}
		}

	case recordingFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Recording error: %v", msg.err)
			return m, nil
		}
		m.state = StateTranscribing
		m.statusMsg = RenderTranscribingStatus(m.width)
		return m, m.transcribeAudio(msg.filename)

	case transcriptionFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Transcription error: %v", msg.err)
			return m, nil
		}
		m.addUserMessage(msg.text)
		m.state = StateChatting
		m.statusMsg = RenderThinkingStatus(m.width)
		return m, m.sendMessage(msg.text)

	case chatFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Chat error: %v", msg.err)
			return m, nil
		}
		m.addAssistantMessage(msg.response)
		m.state = StateGeneratingAudio
		m.statusMsg = RenderGeneratingAudioStatus(m.width)
		return m, m.generateAudio(msg.response)

	case audioFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Audio generation error: %v", msg.err)
			return m, nil
		}
		m.state = StatePlaying
		m.statusMsg = RenderPlayingStatus(m.width)
		return m, m.playAudio(msg.audioData)

	case historyUpdateMsg:
		// Handle real-time history updates from other instances
		m.persona.History = msg.history
		m.reRenderMessages()

	case personaUpdateMsg:
		// Handle persona updates
		m.persona = msg.persona

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update components
	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	switch m.mode {
	case ModePersonaSelector:
		return m.viewPersonaSelector()
	case ModeChat:
		return m.viewChat()
	default:
		return "Mode inconnu"
	}
}

func (m *ChatModel) viewPersonaSelector() string {
	// Lipgloss layering (style.go:415-447): Width/Height include
	// padding but NOT border. Border is applied last and adds 2 cells
	// per axis. So for an outer frame whose total dim must equal the
	// terminal:
	//   - frameStyle.Width  = m.width  - 2  (gives total m.width)
	//   - frameStyle.Height = m.height - 2  (gives total m.height)
	// The actual usable content area inside (after the frame's
	// horizontal padding of 1 on each side) is then m.width - 4 wide
	// and m.height - 2 tall.
	frameWidth := max(m.width-2, MIN_TERMINAL_WIDTH-2)
	frameHeight := max(m.height-2, MIN_TERMINAL_HEIGHT-2)
	innerWidth := max(m.width-4, MIN_TERMINAL_WIDTH-4)
	innerHeight := max(m.height-2, MIN_TERMINAL_HEIGHT-2)

	// Header: brand pill + count of available personas.
	count := fmt.Sprintf("%d personas", len(m.personaList.Items()))
	header := accentAltPill(iconBrand, "persona") + " " + metaPill(iconUser, count)
	headerLine := lipgloss.PlaceHorizontal(innerWidth, lipgloss.Left, header)

	// Footer: keybindings as accent pills.
	keys := []string{
		accentPill("↑↓", "navigate"),
		accentPill("⏎", "select"),
		accentPill("/", "search"),
	}
	if m.persona != nil {
		keys = append(keys, accentAltPill("⌃S", "back to chat"))
	}
	keys = append(keys, mutedPill("⌃C", "quit"))
	footer := lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, strings.Join(keys, " "))
	footerHeight := 1
	if m.errorMsg != "" {
		footer = lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, RenderError(m.errorMsg)) + "\n" + footer
		footerHeight = 2
		m.errorMsg = ""
	}

	// Body fills what's left between header and footer. Header(1) +
	// blank line(1) + footer(footerHeight) + blank line(1).
	bodyHeight := max(innerHeight-3-footerHeight, MIN_VIEWPORT_HEIGHT)

	// Split body horizontally: list on the left (~45%), preview on
	// the right (~55%) with a 1-column gap.
	rightWidth := max(innerWidth*55/100, 36)
	leftWidth := max(innerWidth-rightWidth-1, 24)

	// Re-bind the delegate width and resize the list every render so
	// items track the current pane width.
	m.listDelegate.width = leftWidth
	m.personaList.SetDelegate(m.listDelegate)
	m.personaList.SetSize(leftWidth, bodyHeight)
	listView := m.personaList.View()

	// Same layering rule for the inner peach pane: pass total - 2 to
	// Width/Height so border + padding land inside rightWidth × bodyHeight.
	previewFrameWidth := max(rightWidth-2, 20)
	previewFrameHeight := max(bodyHeight-2, 4)
	previewContent := max(rightWidth-4, 18) // total - 2 border - 2 padding
	previewView := focusedBorderStyle.
		Width(previewFrameWidth).
		Height(previewFrameHeight).
		Render(m.renderPersonaPreview(previewContent))

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, " ", previewView)

	content := strings.Join([]string{headerLine, "", body, "", footer}, "\n")

	// frameWidth / frameHeight = total - 2 (border) so the rendered
	// box matches the terminal exactly. MaxWidth/MaxHeight is a final
	// belt-and-braces clip in case some inner widget overshoots.
	return outerFrameStyle.
		Width(frameWidth).
		Height(frameHeight).
		MaxWidth(m.width).
		MaxHeight(m.height).
		Render(content)
}

// renderPersonaPreview renders the detail panel for the currently
// highlighted persona. Reads from previewCache to avoid disk I/O on
// every redraw. Falls back gracefully if the cache misses (e.g. a
// freshly-created persona not yet in the cache).
func (m *ChatModel) renderPersonaPreview(width int) string {
	item, ok := m.personaList.SelectedItem().(PersonaItem)
	if !ok {
		return mutedStyle.Render("No persona selected")
	}

	p := m.previewCache[item.name]
	if p == nil {
		// Lazy fallback: load on miss and store for next render.
		loaded, err := m.manager.GetPersona(item.name)
		if err != nil {
			return RenderError(fmt.Sprintf("Cannot load %s: %v", item.name, err))
		}
		if m.previewCache == nil {
			m.previewCache = map[string]*persona.Persona{}
		}
		m.previewCache[item.name] = loaded
		p = loaded
	}

	// Header: persona name as accent pill with its custom glyph.
	title := accentPill(personaIcon(p.Name), p.Name)

	// Metadata pills: voice + history count.
	historyText := fmt.Sprintf("%d msg", len(p.History))
	meta := lipgloss.JoinHorizontal(
		lipgloss.Top,
		metaPill(iconSpeak, p.Voice.Name),
		" ",
		metaPill(iconClock, historyText),
	)

	// Prompt body: rendered through the markdown pipeline so headings
	// and inline code in system prompts come out themed. Truncated to
	// 12 lines to keep the panel readable on small terminals.
	prompt := strings.TrimSpace(p.Prompt)
	rendered := renderMarkdown(prompt, width)
	rendered = truncateLines(rendered, 12)
	if rendered == "" {
		rendered = mutedStyle.Render("(no system prompt)")
	}

	// Voice instructions in a smaller subtle block underneath.
	var voiceBlock string
	if vi := strings.TrimSpace(p.Voice.Instructions); vi != "" {
		voiceBlock = subtleStyle.Render("Voice ─ ") + mutedStyle.Render(truncateLines(vi, 4))
	}

	parts := []string{title, meta, "", rendered}
	if voiceBlock != "" {
		parts = append(parts, "", voiceBlock)
	}
	return strings.Join(parts, "\n")
}

// truncateLines keeps at most n lines of s, appending an ellipsis pill
// when the input was longer.
func truncateLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n") + "\n" + mutedStyle.Render("…")
}

func (m *ChatModel) viewChat() string {
	var sections []string

	// Chat box title with decorative border
	title := fmt.Sprintf("Chat avec %s", m.persona.Name)
	if instances, err := m.instanceManager.GetActiveInstances(); err == nil && len(instances) > 1 {
		title += fmt.Sprintf(" %s %d instances", iconUser, len(instances))
	}
	if m.isMuted {
		title += " " + iconMute
	}
	sections = append(sections, RenderChatBoxTitle(title, m.width))

	// Chat history viewport wrapped in border
	chatContent := strings.Join(m.messages, "\n\n")
	m.viewport.SetContent(chatContent)
	sections = append(sections, RenderChatBoxBorder(m.viewport.View(), m.width, m.height))

	// Input area or status message in a box
	if m.state == StateIdle {
		sections = append(sections, RenderInputBox(m.textArea.View(), m.width))
		sections = append(sections, RenderMuted(iconInfo+" Ctrl+R record · Enter send · Ctrl+L clear · Ctrl+M mute · Ctrl+S switch · Ctrl+C quit"))
	} else {
		if m.errorMsg != "" {
			sections = append(sections, RenderInputBox(RenderError(m.errorMsg), m.width))
			m.errorMsg = "" // Clear after showing
		} else if m.statusMsg != "" {
			statusLine := m.spinner.View() + " " + m.statusMsg
			sections = append(sections, RenderInputBox(statusLine, m.width))
		}
	}

	return strings.Join(sections, "\n")
}

func (m *ChatModel) addMessage(message string) {
	m.messages = append(m.messages, message)

	// Mettre à jour le contenu du viewport avec espacement amélioré
	chatContent := strings.Join(m.messages, "\n\n"+RenderMessageSpacing())
	m.viewport.SetContent(chatContent)

	// Forcer le défilement vers le bas
	m.viewport.GotoBottom()
}

func (m *ChatModel) addUserMessage(message string) {
	totalMessages := len(m.persona.History)
	messageIndex := totalMessages - 1
	isLatest := true

	rendered := RenderUserMessage(message, m.width, messageIndex, isLatest)
	m.addMessage(rendered)
}

func (m *ChatModel) addAssistantMessage(message string) {
	totalMessages := len(m.persona.History)
	messageIndex := totalMessages - 1
	isLatest := true

	rendered := RenderAssistantMessage(m.persona.Name, message, m.width, messageIndex, isLatest)
	m.addMessage(rendered)
}

func (m *ChatModel) addWelcomeMessage() {
	welcomeMessage := "Bonjour ! Je suis prêt à discuter avec vous. " + iconRecord + " Tapez votre message ou utilisez Ctrl+R pour enregistrer un message vocal."
	rendered := RenderAssistantMessage(m.persona.Name, welcomeMessage, m.width, 0, false)
	m.addMessage(rendered)
}

func (m *ChatModel) addHistoryMessages() {
	for i, msg := range m.persona.History {
		isLatest := i == len(m.persona.History)-1

		switch msg.Role {
		case "user":
			rendered := RenderUserMessage(msg.Content, m.width, i, isLatest)
			m.addMessage(rendered)
		case "assistant":
			rendered := RenderAssistantMessage(m.persona.Name, msg.Content, m.width, i, isLatest)
			m.addMessage(rendered)
		}
	}
}

func (m *ChatModel) reRenderMessages() {
	m.messages = []string{}
	m.addWelcomeMessage()
	m.addHistoryMessages()
}

func (m *ChatModel) clearConversation() {
	m.persona.History = []persona.Message{}
	_, historyPath := m.manager.GetPersonaPath(m.persona.Name)
	err := m.persona.SaveHistory(historyPath)
	if err != nil {
		m.state = StateError
		m.errorMsg = fmt.Sprintf(iconError+" History save error: %v", err)
		return
	}
	m.messages = []string{}
	m.addWelcomeMessage()
	m.viewport.GotoTop()
}

func (m *ChatModel) startRecording() tea.Cmd {
	return func() tea.Msg {
		m.state = StateRecording
		m.statusMsg = RenderRecordingStatus(m.width)

		recorder := ffmpeg.New(m.inputDevice, m.silenceThreshold, m.silenceDuration)
		filename, err := recorder.Record()

		return recordingFinishedMsg{filename: filename, err: err}
	}
}

func (m *ChatModel) transcribeAudio(filename string) tea.Cmd {
	return func() tea.Msg {
		dataToTranscribe, err := os.Open(filename)
		if err != nil {
			return transcriptionFinishedMsg{err: err}
		}

		transcript, err := m.ai.Transcribe(m.ctx(), dataToTranscribe)
		if err != nil {
			return transcriptionFinishedMsg{err: err}
		}
		err = dataToTranscribe.Close()
		if err != nil {
			return transcriptionFinishedMsg{err: err}
		}
		err = os.Remove(filename)
		if err != nil {
			return transcriptionFinishedMsg{err: err}
		}
		return transcriptionFinishedMsg{text: transcript, err: err}
	}
}

func (m *ChatModel) sendTextMessage(message string) tea.Cmd {
	m.addUserMessage(message)
	m.state = StateChatting
	m.statusMsg = RenderThinkingStatus(m.width)

	// Réinitialiser complètement la zone de saisie
	m.textArea.Reset()
	m.textArea.Blur()
	m.textArea.Focus()

	return m.sendMessage(message)
}

func (m *ChatModel) sendMessage(message string) tea.Cmd {
	return func() tea.Msg {
		// Add to persona history
		m.persona.History = append(m.persona.History, persona.Message{
			Role:    "user",
			Content: message,
		})

		// Prepare messages for AI
		messages := m.persona.GetMessages()
		aiMessages := []openai.Message{}
		for _, msg := range messages {
			aiMessages = append(aiMessages, openai.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Get AI response
		response, err := m.ai.Chat(m.ctx(), aiMessages)
		if err != nil {
			return chatFinishedMsg{err: err}
		}

		// Add to persona history
		m.persona.History = append(m.persona.History, persona.Message{
			Role:    "assistant",
			Content: response,
		})

		// Save history (this will trigger file watcher in other instances)
		_, historyPath := m.manager.GetPersonaPath(m.persona.Name)
		err = m.persona.SaveHistory(historyPath)
		if err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" History save error: %v", err)
			return chatFinishedMsg{err: err}
		}

		return chatFinishedMsg{response: response, err: nil}
	}
}

func (m *ChatModel) generateAudio(text string) tea.Cmd {
	return func() tea.Msg {
		audio, err := m.ai.GenerateAudio(m.ctx(), text, m.persona.Voice.Instructions)
		if err != nil {
			return audioFinishedMsg{err: err}
		}
		return audioFinishedMsg{audioData: audio, err: nil}
	}
}

func (m *ChatModel) playAudio(audioData []byte) tea.Cmd {
	return func() tea.Msg {
		// Skip audio playback if muted
		if m.isMuted {
			m.state = StateIdle
			m.statusMsg = ""
			return nil
		}

		// Create temporary file. Close immediately so WriteFile owns
		// the descriptor; defer Remove only after successful write so
		// we don't delete a file we never wrote to.
		tempFile, err := os.CreateTemp("", "persona-*.mp3")
		if err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Temporary file creation error: %v", err)
			return nil
		}
		tempPath := tempFile.Name()
		if err := tempFile.Close(); err != nil {
			_ = os.Remove(tempPath)
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Temporary file close error: %v", err)
			return nil
		}

		if err := os.WriteFile(tempPath, audioData, 0o600); err != nil {
			_ = os.Remove(tempPath)
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Audio write error: %v", err)
			return nil
		}
		defer func() { _ = os.Remove(tempPath) }()

		if err := speak.Play(tempPath); err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf(iconError+" Audio read error: %v", err)
			return nil
		}

		// Reset to idle state
		m.state = StateIdle
		m.statusMsg = ""

		return nil
	}
}

// Cleanup cleans up resources when the chat is closed.
// cancelRoot stops the heartbeat and watcher goroutines via ctx.
func (m *ChatModel) Cleanup() {
	if m.cancelRoot != nil {
		m.cancelRoot()
	}
	if m.personaWatcher != nil {
		m.personaWatcher.Stop()
	}
	if m.instanceManager != nil {
		if err := m.instanceManager.UnregisterInstance(); err != nil {
			log.Printf("Error unsubscribing instance: %v", err)
		}
	}
}

// NewChatModelWithSelector creates a new chat model that starts with persona selection
func NewChatModelWithSelector(manager *storage.Manager, config *config.Config, openaiAPIKey string) *ChatModel {
	// Get terminal size with fallback to minimum dimensions
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 || height == 0 {
		width = MIN_TERMINAL_WIDTH
		height = MIN_TERMINAL_HEIGHT
	}

	width = max(width, MIN_TERMINAL_WIDTH)
	height = max(height, MIN_TERMINAL_HEIGHT)

	// Get available personas
	personas, err := manager.ListPersonas()
	if err != nil {
		personas = []string{}
	}

	// Pre-load every persona once so the preview pane and the list
	// description can read from memory instead of hitting disk on every
	// keystroke. Cost is negligible (a handful of small YAML files).
	previewCache := make(map[string]*persona.Persona, len(personas))

	// Create persona items for the list
	items := make([]list.Item, 0, len(personas))
	for _, p := range personas {
		personaData, err := manager.GetPersona(p)
		description := "AI Persona"
		if err == nil {
			previewCache[p] = personaData
			if len(personaData.Prompt) > 50 {
				description = personaData.Prompt[:50] + "..."
			} else {
				description = personaData.Prompt
			}
		}

		items = append(items, PersonaItem{
			name:        p,
			description: description,
		})
	}

	// Initialize persona list. Delegate width is updated per render
	// to match the current pane width.
	delegate := itemDelegate{width: width}
	l := list.New(items, delegate, width-4, height-4)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = l.Styles.Title.
		Foreground(colBase).
		Background(accent).
		Bold(true).
		Padding(0, 1)

	// Initialize other components
	vp := viewport.New(width-4, height-10)
	ta := textarea.New()
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ProgressBarStyle

	model := &ChatModel{
		mode:             ModePersonaSelector,
		viewport:         vp,
		textArea:         ta,
		spinner:          s,
		personaList:      l,
		state:            StateIdle,
		manager:          manager,
		config:           config,
		openaiAPIKey:     openaiAPIKey,
		messages:         []string{},
		width:            width,
		height:           height,
		inputDevice:      config.Audio.InputDevice,
		silenceThreshold: config.Audio.SilenceThreshold,
		silenceDuration:  config.Audio.SilenceDuration,
		previewCache:     previewCache,
		listDelegate:     delegate,
	}

	return model
}

// SwitchToPersona switches the chat model to a specific persona
func (m *ChatModel) SwitchToPersona(personaName string) error {
	// Load the persona
	persona, err := m.manager.GetPersona(personaName)
	if err != nil {
		return fmt.Errorf("unable to load persona '%s': %w", personaName, err)
	}

	// Create new OpenAI client for this persona
	ai := openai.New(
		m.openaiAPIKey,
		m.config.Models.Transcription,
		m.config.Models.Speech,
		m.config.Models.Chat,
		persona.Voice.Name,
	)

	// Update model state
	m.persona = persona
	m.ai = ai
	m.mode = ModeChat

	// Recalculate dimensions for chat mode
	viewportWidth, viewportHeight, inputHeight := GetChatLayoutDimensions(m.width, m.height)

	// Reconfigure viewport
	m.viewport = viewport.New(viewportWidth, viewportHeight)
	m.viewport.Style = lipgloss.NewStyle().
		MarginLeft(HORIZONTAL_MARGIN).
		MarginRight(HORIZONTAL_MARGIN)

	// Reconfigure text area
	m.textArea.Placeholder = "💬 Tapez votre message ou Ctrl+R pour enregistrer..."
	m.textArea.Focus()
	m.textArea.ShowLineNumbers = false
	m.textArea.SetWidth(viewportWidth)
	m.textArea.SetHeight(inputHeight)

	// Initialize file watchers and instance management
	err = m.initializeWatchers()
	if err != nil {
		log.Printf("Error initializing watchers: %v", err)
	}

	// Load and display persona history
	_, historyPath := m.manager.GetPersonaPath(m.persona.Name)
	err = m.persona.LoadHistory(historyPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading history: %v", err)
	}

	// Clear current messages and load history
	m.messages = []string{}
	m.loadHistoryToMessages()

	// Update viewport content
	chatContent := strings.Join(m.messages, "\n\n")
	m.viewport.SetContent(chatContent)
	m.viewport.GotoBottom()

	return nil
}

// initializeWatchers initializes the file watchers for the current persona
func (m *ChatModel) initializeWatchers() error {
	if m.persona == nil {
		return fmt.Errorf("aucun persona chargé")
	}

	// Initialize file watcher
	if personaWatcher, err := watcher.NewPersonaWatcher(m.manager, m.persona.Name); err == nil {
		m.personaWatcher = personaWatcher
		personaWatcher.Start(m.ctx())
	} else {
		return fmt.Errorf("unable to initialize persona watcher: %w", err)
	}

	// Initialize instance manager
	m.instanceManager = watcher.NewInstanceManager(m.manager)
	if err := m.instanceManager.RegisterInstance(); err != nil {
		return fmt.Errorf("unable to initialize instance manager: %w", err)
	}
	m.instanceManager.StartHeartbeat(m.ctx())

	return nil
}

// loadHistoryToMessages loads the persona history into the messages display
func (m *ChatModel) loadHistoryToMessages() {
	if m.persona == nil {
		return
	}

	for i, msg := range m.persona.History {
		totalMessages := len(m.persona.History)
		isLatest := i == totalMessages-1

		var rendered string
		if msg.Role == "user" {
			rendered = RenderUserMessage(msg.Content, m.width, i, isLatest)
		} else {
			rendered = RenderAssistantMessage(m.persona.Name, msg.Content, m.width, i, isLatest)
		}
		m.messages = append(m.messages, rendered)
	}
}
