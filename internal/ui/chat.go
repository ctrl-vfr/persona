package ui

import (
	"fmt"
	"io"
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
	heartbeatStop   chan bool

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

	// Ensure minimum dimensions
	if width < MIN_TERMINAL_WIDTH {
		width = MIN_TERMINAL_WIDTH
	}
	if height < MIN_TERMINAL_HEIGHT {
		height = MIN_TERMINAL_HEIGHT
	}

	// Calculate responsive dimensions
	viewportWidth, viewportHeight, inputHeight := GetChatLayoutDimensions(width, height)

	// Initialize viewport with margins
	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = lipgloss.NewStyle().
		MarginLeft(HORIZONTAL_MARGIN).
		MarginRight(HORIZONTAL_MARGIN)

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
		personaWatcher.Start()
	}

	// Initialize instance manager
	model.instanceManager = watcher.NewInstanceManager(manager)
	if err := model.instanceManager.RegisterInstance(); err == nil {
		model.heartbeatStop = model.instanceManager.StartHeartbeat()
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
		m.width = msg.Width
		m.height = msg.Height
		if m.width < MIN_TERMINAL_WIDTH {
			m.width = MIN_TERMINAL_WIDTH
		}
		if m.height < MIN_TERMINAL_HEIGHT {
			m.height = MIN_TERMINAL_HEIGHT
		}
		m.personaList.SetSize(m.width-4, m.height-4)

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
			m.errorMsg = fmt.Sprintf("❌ Recording error: %v", msg.err)
			return m, nil
		}
		m.state = StateTranscribing
		m.statusMsg = RenderTranscribingStatus(m.width)
		return m, m.transcribeAudio(msg.filename)

	case transcriptionFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Transcription error: %v", msg.err)
			return m, nil
		}
		m.addUserMessage(msg.text)
		m.state = StateChatting
		m.statusMsg = RenderThinkingStatus(m.width)
		return m, m.sendMessage(msg.text)

	case chatFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Chat error: %v", msg.err)
			return m, nil
		}
		m.addAssistantMessage(msg.response)
		m.state = StateGeneratingAudio
		m.statusMsg = RenderGeneratingAudioStatus(m.width)
		return m, m.generateAudio(msg.response)

	case audioFinishedMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Audio generation error: %v", msg.err)
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
	var sections []string

	// Title
	sections = append(sections, RenderChatBoxTitle("🎭 Sélection de Persona", m.width))

	// Persona list in a chat-style box with dynamic height
	listHeight := m.height - 8 // Reserve space for title, help and margins
	if listHeight < 10 {
		listHeight = 10
	}
	sections = append(sections, RenderChatBoxBorder(m.personaList.View(), m.width, listHeight))

	// Simple help without box
	var helpLines []string
	helpLines = append(helpLines, "💡 Ctrl+C: Quitter | ↑/↓: Naviguer | Enter: Sélectionner | /: Rechercher")
	if m.persona != nil {
		helpLines = append(helpLines, "   Ctrl+S: Retourner au chat")
	}

	if m.errorMsg != "" {
		helpLines = append(helpLines, RenderError(m.errorMsg))
		m.errorMsg = "" // Clear after showing
	}

	sections = append(sections, strings.Join(helpLines, " | "))

	return strings.Join(sections, "\n")
}

func (m *ChatModel) viewChat() string {
	var sections []string

	// Chat box title with decorative border
	title := fmt.Sprintf("Chat avec %s", m.persona.Name)
	if instances, err := m.instanceManager.GetActiveInstances(); err == nil && len(instances) > 1 {
		title += fmt.Sprintf(" 👥 (%d instances)", len(instances))
	}
	if m.isMuted {
		title += " 🔇"
	}
	sections = append(sections, RenderChatBoxTitle(title, m.width))

	// Chat history viewport wrapped in border
	chatContent := strings.Join(m.messages, "\n\n")
	m.viewport.SetContent(chatContent)
	sections = append(sections, RenderChatBoxBorder(m.viewport.View(), m.width, m.height))

	// Input area or status message in a box
	if m.state == StateIdle {
		sections = append(sections, RenderInputBox(m.textArea.View(), m.width))
		sections = append(sections, RenderMuted("💡 Ctrl+R: Enregistrer | Enter: Envoyer | Ctrl+L: Effacer | Ctrl+M: Mute | Ctrl+S: Changer persona | Ctrl+C: Quitter"))
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
	welcomeMessage := "Bonjour ! Je suis prêt à discuter avec vous. 🎤 Tapez votre message ou utilisez Ctrl+R pour enregistrer un message vocal."
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
		m.errorMsg = fmt.Sprintf("❌ History save error: %v", err)
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

		transcript, err := m.ai.Transcribe(dataToTranscribe)
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
		response, err := m.ai.Chat(aiMessages)
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
			m.errorMsg = fmt.Sprintf("❌ History save error: %v", err)
			return chatFinishedMsg{err: err}
		}

		return chatFinishedMsg{response: response, err: nil}
	}
}

func (m *ChatModel) generateAudio(text string) tea.Cmd {
	return func() tea.Msg {
		data, err := m.ai.GenerateAudio(text, m.persona.Voice.Instructions)
		if err != nil {
			return audioFinishedMsg{err: err}
		}

		audio, err := io.ReadAll(data)
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

		// Create temporary file
		tempFile, err := os.CreateTemp("", "persona-*.mp3")
		if err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Temporary file creation error: %v", err)
			return nil
		}
		defer os.Remove(tempFile.Name())

		// Write audio data
		if err := os.WriteFile(tempFile.Name(), audioData, 0644); err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Audio write error: %v", err)
			return nil
		}

		// Play audio
		err = speak.Play(tempFile.Name())
		if err != nil {
			m.state = StateError
			m.errorMsg = fmt.Sprintf("❌ Audio read error: %v", err)
			return nil
		}

		// Reset to idle state
		m.state = StateIdle
		m.statusMsg = ""

		return nil
	}
}

// Cleanup cleans up resources when the chat is closed
func (m *ChatModel) Cleanup() {
	if m.personaWatcher != nil {
		m.personaWatcher.Stop()
	}

	if m.heartbeatStop != nil {
		close(m.heartbeatStop)
	}

	if m.instanceManager != nil {
		err := m.instanceManager.UnregisterInstance()
		if err != nil {
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

	// Ensure minimum dimensions
	if width < MIN_TERMINAL_WIDTH {
		width = MIN_TERMINAL_WIDTH
	}
	if height < MIN_TERMINAL_HEIGHT {
		height = MIN_TERMINAL_HEIGHT
	}

	// Get available personas
	personas, err := manager.ListPersonas()
	if err != nil {
		personas = []string{}
	}

	// Create persona items for the list
	items := make([]list.Item, 0, len(personas))
	for _, p := range personas {
		// Try to load persona to get description from prompt
		personaData, err := manager.GetPersona(p)
		description := "AI Persona"
		if err == nil && len(personaData.Prompt) > 50 {
			description = personaData.Prompt[:50] + "..."
		} else if err == nil {
			description = personaData.Prompt
		}

		items = append(items, PersonaItem{
			name:        p,
			description: description,
		})
	}

	// Initialize persona list
	l := list.New(items, list.NewDefaultDelegate(), width-4, height-4)
	l.Title = "Sélectionnez un persona"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = l.Styles.Title.
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
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
		personaWatcher.Start()
	} else {
		return fmt.Errorf("unable to initialize persona watcher: %w", err)
	}

	// Initialize instance manager
	m.instanceManager = watcher.NewInstanceManager(m.manager)
	if err := m.instanceManager.RegisterInstance(); err == nil {
		m.heartbeatStop = m.instanceManager.StartHeartbeat()
	} else {
		return fmt.Errorf("unable to initialize instance manager: %w", err)
	}

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
