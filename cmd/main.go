// Package cmd wires the cobra command tree for the persona CLI.
// All commands share access to a single *storage.Manager via the
// package-level storageManager variable, which must be initialised
// by main.go via Setup before Execute is called. This avoids the
// previous init()-time side effect of creating ~/.persona on import,
// which broke isolation in tests and prevented dependency injection.
package cmd

import (
	"github.com/ctrl-vfr/persona/internal/storage"

	"github.com/spf13/cobra"
)

// storageManager is the active storage manager shared by all commands.
// It is set by Setup; reading it before Setup is called is a programmer
// error and panics with a clear message.
var storageManager *storage.Manager

// Setup wires the cmd package with a ready-to-use storage manager.
// Call this exactly once from main before invoking the cobra command
// tree. Tests can pass a manager pointing at a temporary directory to
// run commands in isolation.
func Setup(mgr *storage.Manager) {
	storageManager = mgr
}

var rootCmd = &cobra.Command{
	Use:   "persona",
	Short: "Voice assistant persona with interactive interface",
	Long: `Persona is an intelligent voice assistant that can:
• Record and transcribe your voice messages
• Chat with different AI personas
• Manage conversation history
• Provide a colorful and interactive interface`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
