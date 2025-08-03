package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ctrl-vfr/persona/internal/ffmpeg"
	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

var ffmpegCmd = &cobra.Command{
	Use:   "ffmpeg",
	Short: "FFmpeg management",
	Long:  "Commands to manage FFmpeg and audio devices",
}

var ffmpegListCmd = &cobra.Command{
	Use:   "list",
	Short: "List devices",
	Long:  "List available audio devices",
}

// Global output format variables are defined in version.go

var ffmpegListInputCmd = &cobra.Command{
	Use:   "input",
	Short: "List audio input devices",
	Long:  "List all available audio input devices for recording",
	Run: func(cmd *cobra.Command, args []string) {
		audioDevicesList, err := ffmpeg.ListAudioDevices()
		if err != nil {
			fmt.Printf("Error listing devices: %v\n", err)
			return
		}

		if outputJSON {
			data, err := json.MarshalIndent(audioDevicesList, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			if len(audioDevicesList) == 0 {
				fmt.Println("No audio devices found.")
				return
			}
			for _, deviceName := range audioDevicesList {
				fmt.Println(deviceName)
			}
			return
		}

		// Default: full formatted output
		if len(audioDevicesList) == 0 {
			fmt.Println(ui.TitleStyle.Render("No audio devices found."))
			fmt.Println(ui.ContentStyle.Render("You can list available audio devices with the `ffmpeg list devices` command."))
			fmt.Println()
			return
		}

		fmt.Println(ui.TitleStyle.Render("Audio devices:"))
		for _, deviceName := range audioDevicesList {
			fmt.Println(ui.ContentStyle.Render(deviceName))
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(ffmpegCmd)
	ffmpegCmd.AddCommand(ffmpegListCmd)
	ffmpegListCmd.AddCommand(ffmpegListInputCmd)

	// Add output format flags
	ffmpegListInputCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	ffmpegListInputCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")
}
