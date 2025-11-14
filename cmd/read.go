package cmd

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/ctrl-vfr/persona/internal/audio"
	"github.com/ctrl-vfr/persona/internal/ffmpeg"
	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/speak"
	"github.com/ctrl-vfr/persona/internal/textsplit"
	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

var (
	readOutputFormat string
	maxWorkers       int
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

		// NOTE: Initialize OpenAI client
		aiClient := openai.New(os.Getenv("OPENAI_API_KEY"), appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		// REVIEW: Split text into manageable chunks
		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("📄 Splitting text into chunks..."))
		}
		chunks := textsplit.SplitText(textContent)

		if readOutputFormat == "default" {
			fmt.Printf(ui.RenderInfo("🔊 Generating audio for %d chunks in parallel..."), len(chunks))
		}

		// TODO: Add configuration for TTS generator settings
		// Create TTS generator with proper separation of concerns
		ttsGenerator := audio.NewTTSGenerator(aiClient, currentPersona.Voice.Instructions, maxWorkers)

		// Generate audio for all chunks in parallel
		audioChunks, err := ttsGenerator.GenerateParallel(chunks)
		if err != nil {
			log.Fatal("Parallel audio generation error:", err)
		}

		// OPTIMIZE: Filter successful chunks and handle errors
		successfulChunks, failedErrors := audio.FilterSuccessfulChunks(audioChunks)

		if len(failedErrors) > 0 {
			log.Fatal("Some chunks failed to generate:\n" + fmt.Sprintf("%v", failedErrors))
		}

		// NOTE: Sort chunks by order for proper concatenation
		slices.SortFunc(successfulChunks, func(a, b audio.AudioChunk) int {
			return cmp.Compare(a.Order, b.Order)
		})

		// Extract file paths for FFmpeg concatenation
		audioFiles := audio.ExtractFilePaths(successfulChunks)

		// Create final output file
		tempAudioResponseFile, err := os.CreateTemp("", "persona-read-final-*.mp3")
		if err != nil {
			audio.CleanupAudioChunks(successfulChunks)
			log.Fatal("Temporary audio file creation error:", err)
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔗 Concatenating audio files..."))
		}

		// WARNING: Use FFmpeg package for concatenation - proper separation of concerns
		err = ffmpeg.ConcatenateAudioFiles(audioFiles, tempAudioResponseFile.Name())
		if err != nil {
			audio.CleanupAudioChunks(successfulChunks)
			log.Fatal("Audio concatenation error:", err)
		}

		defer audio.CleanupAudioChunks(successfulChunks)

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
	readCmd.Flags().IntVarP(&maxWorkers, "workers", "w", 3, "Maximum number of parallel workers for TTS generation (1-5)")
}
