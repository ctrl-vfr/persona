package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ctrl-vfr/persona/internal/ffmpeg"
	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/output"
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
	Short: "Simple discussion with a persona (legacy mode)",
	Long:  "Simple discussion mode, one question-answer at a time. Use 'persona chat' for interactive interface.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]
		formatter := output.New(output.ParseFormat(askOutputFormat))

		if askOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("🎙️ Discussion with %s", personaName), terminalWidth))
		}

		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading persona: %v", err))
			return
		}

		appConfig, err := storageManager.GetConfig()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading configuration: %v", err))
			return
		}

		if appConfig.Audio.InputDevice == "" {
			formatter.Error("Audio input device not configured. Use 'persona config set-input-device <device>'.")
			return
		}

		aiClient := openai.New(os.Getenv("OPENAI_API_KEY"), appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		// Start recording
		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🎤 Recording started... Speak now!"))
		}
		recorder := ffmpeg.New(appConfig.Audio.InputDevice, appConfig.Audio.SilenceThreshold, appConfig.Audio.SilenceDuration)
		tempAudioFile, err := recorder.Record()
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio recording error: %v", err))
			return
		}

		audioDataToTranscribe, err := os.Open(tempAudioFile)
		if err != nil {
			formatter.Error(fmt.Sprintf("Error opening temporary file: %v", err))
			return
		}
		defer os.Remove(tempAudioFile)
		defer audioDataToTranscribe.Close()

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("📝 Transcribing..."))
		}
		transcription, err := aiClient.Transcribe(audioDataToTranscribe)
		if err != nil {
			formatter.Error(fmt.Sprintf("Transcription error: %v", err))
			return
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
			formatter.Error(fmt.Sprintf("AI chat error: %v", err))
			return
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
			formatter.Error(fmt.Sprintf("Audio generation error: %v", err))
			return
		}

		audioBytes, err := io.ReadAll(audioResponseData)
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio data read error: %v", err))
			return
		}

		tempAudioResponseFile, err := os.CreateTemp("", "persona-response-*.mp3")
		if err != nil {
			formatter.Error(fmt.Sprintf("Temporary audio file creation error: %v", err))
			return
		}
		defer os.Remove(tempAudioResponseFile.Name())

		err = os.WriteFile(tempAudioResponseFile.Name(), audioBytes, 0o644)
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio file write error: %v", err))
			return
		}

		if askOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Playing response..."))
		}
		err = speak.Play(tempAudioResponseFile.Name())
		if err != nil {
			formatter.Error(fmt.Sprintf("Response playback error: %v", err))
			return
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
