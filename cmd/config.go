package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Commands to display and manage application configuration",
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		appConfig, err := storageManager.GetConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		if outputJSON {
			data, err := json.MarshalIndent(appConfig, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			configData, err := json.MarshalIndent(appConfig, "", "  ")
			if err != nil {
				fmt.Printf("Configuration formatting error: %v\n", err)
				return
			}
			fmt.Println(string(configData))
			return
		}

		// Default: full formatted output
		fmt.Println(ui.RenderInfo("Configuration:"))
		fmt.Println()
		configData, err := json.MarshalIndent(appConfig, "", "  ")
		if err != nil {
			fmt.Printf("Configuration formatting error: %v\n", err)
			return
		}
		// Indent each line with 2 spaces
		lines := fmt.Sprintf("%s", configData)
		for _, line := range strings.Split(lines, "\n") {
			if line != "" {
				fmt.Printf("  %s\n", line)
			}
		}
	},
}

var pathConfigCmd = &cobra.Command{
	Use:   "path",
	Short: "Display configuration paths",
	Run: func(cmd *cobra.Command, args []string) {
		if outputJSON {
			pathData := map[string]interface{}{
				"config_dir":   storageManager.BasePath,
				"config_file":  storageManager.BasePath + "/config.yaml",
				"personas_dir": storageManager.BasePath + "/personas/",
			}
			data, err := json.MarshalIndent(pathData, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			fmt.Printf("Config directory: %s\n", storageManager.BasePath)
			fmt.Printf("Config file: %s/config.yaml\n", storageManager.BasePath)
			fmt.Printf("Personas directory: %s/personas/\n", storageManager.BasePath)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.RenderInfo("Configuration paths:"))
		fmt.Println()
		fmt.Printf("  - Config directory: %s\n", storageManager.BasePath)
		fmt.Printf("  - Config file: %s/config.yaml\n", storageManager.BasePath)
		fmt.Printf("  - Personas directory: %s/personas/\n", storageManager.BasePath)
	},
}

var setInputDeviceCmd = &cobra.Command{
	Use:   "set-input-device [device]",
	Short: "Set audio input device",
	Long:  "Set the audio input device for voice recording",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceName := args[0]

		appConfig, err := storageManager.GetConfig()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}

		appConfig.Audio.InputDevice = deviceName

		err = storageManager.SaveConfig(appConfig)
		if err != nil {
			fmt.Printf("Configuration save error: %v\n", err)
			return
		}

		if outputJSON {
			result := map[string]interface{}{
				"device": deviceName,
				"status": "configured",
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			fmt.Printf("Input device configured: %s\n", deviceName)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.RenderSuccess("Device configured:"))
		fmt.Println()
		fmt.Printf("  - %s\n", deviceName)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showConfigCmd)
	configCmd.AddCommand(pathConfigCmd)
	configCmd.AddCommand(setInputDeviceCmd)

	// Add output format flags
	showConfigCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	showConfigCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	pathConfigCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	pathConfigCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	setInputDeviceCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	setInputDeviceCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")
}
