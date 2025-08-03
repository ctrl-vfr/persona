package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/output"
	"github.com/ctrl-vfr/persona/internal/speak"
	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

var (
	readOutputFormat string
)

var readCmd = &cobra.Command{
	Use:   "read [persona] [file]",
	Short: "Have a persona read a text file",
	Long:  "Ask a persona to read aloud the content of a text file.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]
		filePath := args[1]
		formatter := output.New(output.ParseFormat(readOutputFormat))

		if readOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("📖 Reading by %s", personaName), terminalWidth))
		}

		// Load persona
		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading persona: %v", err))
			return
		}

		// Load configuration
		appConfig, err := storageManager.GetConfig()
		if err != nil {
			formatter.Error(fmt.Sprintf("Error loading configuration: %v", err))
			return
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			formatter.Error(fmt.Sprintf("Error reading file: %v", err))
			return
		}

		textContent := string(content)
		if readOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderUserMessage(textContent, terminalWidth, 0, true))
		}

		// Initialize OpenAI client
		aiClient := openai.New(os.Getenv("OPENAI_API_KEY"), appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔊 Generating audio..."))
		}
		audioResponseData, err := aiClient.GenerateAudio(textContent, currentPersona.Voice.Instructions)
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio generation error: %v", err))
			return
		}

		audioBytes, err := io.ReadAll(audioResponseData)
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio data read error: %v", err))
			return
		}

		tempAudioResponseFile, err := os.CreateTemp("", "persona-read-*.mp3")
		if err != nil {
			formatter.Error(fmt.Sprintf("Temporary audio file creation error: %v", err))
			return
		}

		err = os.WriteFile(tempAudioResponseFile.Name(), audioBytes, 0644)
		if err != nil {
			formatter.Error(fmt.Sprintf("Audio file write error: %v", err))
			return
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Reading text..."))
		}
		err = speak.Play(tempAudioResponseFile.Name())
		if err != nil {
			formatter.Error(fmt.Sprintf("Text reading error: %v", err))
			return
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderSuccess("Reading completed!"))
		}
		err = os.Remove(tempAudioResponseFile.Name())
		if err != nil {
			formatter.Error(fmt.Sprintf("Temporary audio file removal error: %v", err))
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&readOutputFormat, "output", "o", "default", "Output format (default, json, plain)")
}
