package cmd

import (
	"fmt"
	"os"

	"github.com/ctrl-vfr/persona/internal/ffmpeg"
	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/persona"
	"github.com/ctrl-vfr/persona/internal/speak"
	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

var (
	askOutputFormat string
)

var askCmd = &cobra.Command{
	Use:   "ask [nom]",
	Short: "Simple discussion with a persona (non-interactive)",
	Long:  "Simple discussion mode, one question-answer at a time. Use 'persona chat' for interactive interface.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		personaName := args[0]
		ctx := cmd.Context()

		if askOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("🎙️ Discussion with %s", personaName), terminalWidth))
		}

		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			return fmt.Errorf("load persona: %w", err)
		}

		appConfig, err := storageManager.GetConfig()
		if err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}

		if appConfig.Audio.InputDevice == "" {
			return fmt.Errorf("audio input device not configured (use 'persona config set-input-device <device>')")
		}

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is not set")
		}
		aiClient := openai.New(apiKey, appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🎤 Recording started... Speak now!"))
		}
		recorder := ffmpeg.New(appConfig.Audio.InputDevice, appConfig.Audio.SilenceThreshold, appConfig.Audio.SilenceDuration)
		tempAudioFile, err := recorder.Record()
		if err != nil {
			return fmt.Errorf("audio recording: %w", err)
		}
		defer func() { _ = os.Remove(tempAudioFile) }()

		audioDataToTranscribe, err := os.Open(tempAudioFile)
		if err != nil {
			return fmt.Errorf("open temporary audio file: %w", err)
		}
		defer func() { _ = audioDataToTranscribe.Close() }()

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("📝 Transcribing..."))
		}
		transcription, err := aiClient.Transcribe(ctx, audioDataToTranscribe)
		if err != nil {
			return fmt.Errorf("transcription: %w", err)
		}

		if askOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderUserMessage(transcription, terminalWidth, 0, true))
		}

		currentPersona.History = append(currentPersona.History, persona.Message{
			Role:    "user",
			Content: transcription,
		})

		conversationMessages := currentPersona.GetMessages()
		aiMessages := make([]openai.Message, 0, len(conversationMessages))
		for _, message := range conversationMessages {
			aiMessages = append(aiMessages, openai.Message{
				Role:    message.Role,
				Content: message.Content,
			})
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("💭 Thinking..."))
		}
		aiResponse, err := aiClient.Chat(ctx, aiMessages)
		if err != nil {
			return fmt.Errorf("ai chat: %w", err)
		}

		if askOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderAssistantMessage(currentPersona.Name, aiResponse, terminalWidth, 0, true))
		}

		currentPersona.History = append(currentPersona.History, persona.Message{
			Role:    "assistant",
			Content: aiResponse,
		})
		_, historyPath := storageManager.GetPersonaPath(personaName)
		if err := currentPersona.SaveHistory(historyPath); err != nil {
			return fmt.Errorf("save history: %w", err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔊 Generating audio..."))
		}
		audioBytes, err := aiClient.GenerateAudio(ctx, aiResponse, currentPersona.Voice.Instructions)
		if err != nil {
			return fmt.Errorf("audio generation: %w", err)
		}

		tempAudioResponseFile, err := os.CreateTemp("", "persona-response-*.mp3")
		if err != nil {
			return fmt.Errorf("create temp audio file: %w", err)
		}
		tempPath := tempAudioResponseFile.Name()
		if err := tempAudioResponseFile.Close(); err != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("close temp audio file: %w", err)
		}
		defer func() { _ = os.Remove(tempPath) }()

		if err := os.WriteFile(tempPath, audioBytes, 0o600); err != nil {
			return fmt.Errorf("write audio file: %w", err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Playing response..."))
		}
		if err := speak.Play(tempPath); err != nil {
			return fmt.Errorf("playback: %w", err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderSuccess("Conversation completed!"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().StringVarP(&askOutputFormat, "output", "o", "default", "Output format (default, json, plain)")
}
