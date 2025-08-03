package cmd

import (
	"fmt"
	"io"
	"log"
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
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]

		if askOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("🎙️ Discussion with %s", personaName), terminalWidth))
		}

		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			log.Fatal("Error loading persona:", err)
		}

		appConfig, err := storageManager.GetConfig()
		if err != nil {
			log.Fatal("Error loading configuration:", err)
		}

		if appConfig.Audio.InputDevice == "" {
			log.Fatal("Audio input device not configured. Use 'persona config set-input-device <device>'.")
		}

		aiClient := openai.New(os.Getenv("OPENAI_API_KEY"), appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		// Start recording
		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🎤 Recording started... Speak now!"))
		}
		recorder := ffmpeg.New(appConfig.Audio.InputDevice, appConfig.Audio.SilenceThreshold, appConfig.Audio.SilenceDuration)
		tempAudioFile, err := recorder.Record()
		if err != nil {
			log.Fatal("Audio recording error:", err)
		}

		audioDataToTranscribe, err := os.Open(tempAudioFile)
		if err != nil {
			log.Fatal("Error opening temporary file:", err)
		}
		defer os.Remove(tempAudioFile)
		defer audioDataToTranscribe.Close()

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("📝 Transcribing..."))
		}
		transcription, err := aiClient.Transcribe(audioDataToTranscribe)
		if err != nil {
			log.Fatal("Transcription error:", err)
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
		aiMessages := []openai.Message{}
		for _, message := range conversationMessages {
			aiMessages = append(aiMessages, openai.Message{
				Role:    message.Role,
				Content: message.Content,
			})
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("💭 Thinking..."))
		}
		aiResponse, err := aiClient.Chat(aiMessages)
		if err != nil {
			log.Fatal("AI chat error:", err)
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
		err = currentPersona.SaveHistory(historyPath)
		if err != nil {
			log.Fatal(err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔊 Generating audio..."))
		}
		audioResponseData, err := aiClient.GenerateAudio(aiResponse, currentPersona.Voice.Instructions)
		if err != nil {
			log.Fatal("Audio generation error:", err)
		}

		audioBytes, err := io.ReadAll(audioResponseData)
		if err != nil {
			log.Fatal("Audio data read error:", err)
		}

		tempAudioResponseFile, err := os.CreateTemp("", "persona-response-*.mp3")
		if err != nil {
			log.Fatal("Temporary audio file creation error:", err)
		}
		defer os.Remove(tempAudioResponseFile.Name())

		err = os.WriteFile(tempAudioResponseFile.Name(), audioBytes, 0o644)
		if err != nil {
			log.Fatal("Audio file write error:", err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Playing response..."))
		}
		err = speak.Play(tempAudioResponseFile.Name())
		if err != nil {
			log.Fatal("Response playback error:", err)
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderSuccess("Conversation completed!"))
		}
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().StringVarP(&askOutputFormat, "output", "o", "default", "Output format (default, json, plain)")
}
