package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ctrl-vfr/persona/internal/openai"
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

		if readOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("📖 Reading by %s", personaName), terminalWidth))
		}

		// Load persona
		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			log.Fatal("Error loading persona:", err)
		}

		// Load configuration
		appConfig, err := storageManager.GetConfig()
		if err != nil {
			log.Fatal("Error loading configuration:", err)
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal("Error reading file:", err)
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
			log.Fatal("Audio generation error:", err)
		}

		audioBytes, err := io.ReadAll(audioResponseData)
		if err != nil {
			log.Fatal("Audio data read error:", err)
		}

		tempAudioResponseFile, err := os.CreateTemp("", "persona-read-*.mp3")
		if err != nil {
			log.Fatal("Temporary audio file creation error:", err)
		}

		err = os.WriteFile(tempAudioResponseFile.Name(), audioBytes, 0644)
		if err != nil {
			log.Fatal("Audio file write error:", err)
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Reading text..."))
		}
		err = speak.Play(tempAudioResponseFile.Name())
		if err != nil {
			log.Fatal("Text reading error:", err)
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderSuccess("Reading completed!"))
		}
		err = os.Remove(tempAudioResponseFile.Name())
		if err != nil {
			log.Fatal("Temporary audio file removal error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&readOutputFormat, "output", "o", "default", "Output format (default, json, plain)")
}
