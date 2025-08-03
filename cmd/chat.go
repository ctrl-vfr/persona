package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat [nom-persona]",
	Short: "Interactive chat interface with a persona",
	Long: `Launch an interactive real-time chat interface with a persona.
If no persona is specified, a selection interface will be displayed.

Features:
• Interactive persona selection
• Persona switching during conversation
• Visible and synchronized conversation history
• Text and voice messages with responsive layout
• Real-time status indicators with emojis
• Modern colorful interface that adapts to terminal
• Multi-instance support with file watching
• Automatic resizing`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var personaName string
		if len(args) > 0 {
			personaName = args[0]
		}

		// If no persona specified, start with selection interface
		if personaName == "" {
			// Load configuration first to check requirements
			appConfig, err := storageManager.GetConfig()
			if err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Error loading configuration: %v", err)))
				return
			}

			// Check if input device is configured
			if appConfig.Audio.InputDevice == "" {
				fmt.Println(ui.RenderError("Audio input device not configured."))
				fmt.Println(ui.RenderMuted("Use 'persona config set-input-device <device>' to configure it."))
				fmt.Println(ui.RenderMuted("List devices with 'persona ffmpeg list input'."))
				return
			}

			// Initialize OpenAI client
			openaiAPIKey := os.Getenv("OPENAI_API_KEY")
			if openaiAPIKey == "" {
				fmt.Println(ui.RenderError("OPENAI_API_KEY is not set in environment variables."))
				return
			}

			// Create chat model with persona selector
			chatModel := ui.NewChatModelWithSelector(
				storageManager,
				appConfig,
				openaiAPIKey,
			)

			// Set up cleanup on interrupt
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				fmt.Println(ui.RenderInfo("🔄 Cleaning up..."))
				chatModel.Cleanup()
				os.Exit(0)
			}()

			// Ensure cleanup on normal exit
			defer chatModel.Cleanup()

			// Start Bubble Tea program
			program := tea.NewProgram(
				chatModel,
				tea.WithAltScreen(),       // Use alternate screen
				tea.WithMouseCellMotion(), // Enable mouse support
			)

			// Run the program
			if _, err := program.Run(); err != nil {
				fmt.Printf("❌ Chat execution error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(ui.RenderSuccess("Chat completed! Goodbye! 👋"))
			return
		}

		// Verify persona exists
		if !storageManager.PersonaExists(personaName) {
			fmt.Println(ui.RenderError(fmt.Sprintf("Persona '%s' does not exist.", personaName)))
			fmt.Println(ui.RenderMuted("Use 'persona list' to see available personas."))
			return
		}

		// Load persona
		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Error loading persona: %v", err)))
			return
		}

		// Load configuration
		appConfig, err := storageManager.GetConfig()
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Error loading configuration: %v", err)))
			return
		}

		// Check if input device is configured
		if appConfig.Audio.InputDevice == "" {
			fmt.Println(ui.RenderError("Audio input device not configured."))
			fmt.Println(ui.RenderMuted("Use 'persona config set-input-device <device>' to configure it."))
			fmt.Println(ui.RenderMuted("List devices with 'persona ffmpeg list input'."))
			return
		}

		// Initialize OpenAI client
		openaiAPIKey := os.Getenv("OPENAI_API_KEY")
		if openaiAPIKey == "" {
			fmt.Println(ui.RenderError("OPENAI_API_KEY is not set in environment variables."))
			return
		}

		aiClient := openai.New(
			openaiAPIKey,
			appConfig.Models.Transcription,
			appConfig.Models.Speech,
			appConfig.Models.Chat,
			currentPersona.Voice.Name,
		)

		// Create chat model
		chatModel := ui.NewChatModel(
			currentPersona,
			aiClient,
			storageManager,
			appConfig.Audio.InputDevice,
			appConfig.Audio.SilenceThreshold,
			appConfig.Audio.SilenceDuration,
		)

		// Set up cleanup on interrupt
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println(ui.RenderInfo("🔄 Cleaning up..."))
			chatModel.Cleanup()
			os.Exit(0)
		}()

		// Ensure cleanup on normal exit
		defer chatModel.Cleanup()

		// Start Bubble Tea program
		program := tea.NewProgram(
			chatModel,
			tea.WithAltScreen(),       // Use alternate screen
			tea.WithMouseCellMotion(), // Enable mouse support
		)

		// Run the program
		if _, err := program.Run(); err != nil {
			fmt.Printf("❌ Chat execution error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(ui.RenderSuccess("Chat completed! Goodbye! 👋"))
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
