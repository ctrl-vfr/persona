// Main package for persona command-line tool
package cmd

import (
	"log"

	"github.com/ctrl-vfr/persona/internal/storage"

	"github.com/spf13/cobra"
)

var storageManager *storage.Manager

func init() {
	// Initialize storage manager
	manager, err := storage.NewManager()
	if err != nil {
		log.Fatal("Unable to initialize storage manager:", err)
	}

	storageManager = manager

	// Initialize directory structure and default files
	if err := storageManager.InitializeStructure(); err != nil {
		log.Fatal("Unable to initialize structure:", err)
	}
}

func init() {
	// Only check for OPENAI_API_KEY when actually running commands that need it
	// This allows tests and other commands to run without the API key
}

var rootCmd = &cobra.Command{
	Use:   "persona",
	Short: "Voice assistant persona with interactive interface",
	Long: `Persona is an intelligent voice assistant that can:
• Record and transcribe your voice messages
• Chat with different AI personas
• Manage conversation history
• Provide a colorful and interactive interface`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
