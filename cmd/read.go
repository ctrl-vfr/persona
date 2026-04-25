package cmd

import (
	"cmp"
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		personaName := args[0]
		filePath := args[1]
		ctx := cmd.Context()

		if readOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderChatBoxTitle(fmt.Sprintf("📖 Reading by %s", personaName), terminalWidth))
		}

		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			return fmt.Errorf("load persona: %w", err)
		}

		appConfig, err := storageManager.GetConfig()
		if err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		textContent := string(content)
		if readOutputFormat == "default" {
			terminalWidth := ui.GetTerminalWidth()
			fmt.Println(ui.RenderUserMessage(textContent, terminalWidth, 0, true))
		}

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is not set")
		}
		aiClient := openai.New(apiKey, appConfig.Models.Transcription, appConfig.Models.Speech, appConfig.Models.Chat, currentPersona.Voice.Name)

		// Split text into manageable chunks for parallel TTS generation.
		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("📄 Splitting text into chunks..."))
		}
		chunks := textsplit.SplitText(textContent)

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo(fmt.Sprintf("🔊 Generating audio for %d chunks in parallel...", len(chunks))))
		}

		ttsGenerator := audio.NewTTSGenerator(aiClient, currentPersona.Voice.Instructions, maxWorkers)

		audioChunks, err := ttsGenerator.GenerateParallel(ctx, chunks)
		if err != nil {
			return fmt.Errorf("parallel audio generation: %w", err)
		}

		successfulChunks, failedErrors := audio.FilterSuccessfulChunks(audioChunks)
		if len(failedErrors) > 0 {
			audio.CleanupAudioChunks(successfulChunks)
			return fmt.Errorf("some chunks failed to generate: %v", failedErrors)
		}
		defer audio.CleanupAudioChunks(successfulChunks)

		// Sort by chunk order so concatenation produces the original text.
		slices.SortFunc(successfulChunks, func(a, b audio.AudioChunk) int {
			return cmp.Compare(a.Order, b.Order)
		})
		audioFiles := audio.ExtractFilePaths(successfulChunks)

		tempAudioResponseFile, err := os.CreateTemp("", "persona-read-final-*.mp3")
		if err != nil {
			return fmt.Errorf("create temp audio file: %w", err)
		}
		tempPath := tempAudioResponseFile.Name()
		if err := tempAudioResponseFile.Close(); err != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("close temp audio file: %w", err)
		}
		defer func() { _ = os.Remove(tempPath) }()

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔗 Concatenating audio files..."))
		}
		if err := ffmpeg.ConcatenateAudioFiles(audioFiles, tempPath); err != nil {
			return fmt.Errorf("audio concatenation: %w", err)
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderInfo("🔈 Reading text..."))
		}
		if err := speak.Play(tempPath); err != nil {
			return fmt.Errorf("playback: %w", err)
		}

		if readOutputFormat == "default" {
			fmt.Println(ui.RenderSuccess("Reading completed!"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&readOutputFormat, "output", "o", "default", "Output format (default, json, plain)")
	readCmd.Flags().IntVarP(&maxWorkers, "workers", "w", 3, "Maximum number of parallel workers for TTS generation (1-5)")
}
