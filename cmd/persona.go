package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ctrl-vfr/persona/internal/ui"

	"github.com/spf13/cobra"
)

// Global output format variables are defined in version.go

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available personas",
	Run: func(cmd *cobra.Command, args []string) {
		personas, err := storageManager.ListPersonas()
		if err != nil {
			fmt.Printf("Error listing personas: %v\n", err)
			return
		}

		if outputJSON {
			data, err := json.MarshalIndent(personas, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			if len(personas) == 0 {
				fmt.Println("No personas found.")
				return
			}
			for _, persona := range personas {
				fmt.Println(persona)
			}
			return
		}

		// Default: full formatted output
		if len(personas) == 0 {
			fmt.Println(ui.TitleStyle.Render("No personas found."))
			fmt.Println(ui.ContentStyle.Render("You can create a new persona with the `persona create` command."))
			return
		}

		fmt.Println(ui.TitleStyle.Render("Available Personas:"))
		for _, persona := range personas {
			fmt.Println(ui.ContentStyle.Render(fmt.Sprint(persona)))
		}
		fmt.Println()
	},
}

var createCmd = &cobra.Command{
	Use:   "create [nom]",
	Short: "Create a new persona",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]

		if storageManager.PersonaExists(personaName) {
			fmt.Printf("Persona '%s' already exists.\n", personaName)
			return
		}

		if err := storageManager.CreatePersona(personaName); err != nil {
			fmt.Printf("Error creating persona: %v\n", err)
			return
		}

		if outputJSON {
			result := map[string]any{
				"persona": personaName,
				"status":  "created",
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
			fmt.Printf("Persona '%s' created successfully\n", personaName)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.TitleStyle.Render("Persona created:"))
		fmt.Println(ui.ContentStyle.Render(personaName))
		fmt.Println()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [nom]",
	Short: "Delete a persona",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]

		if !storageManager.PersonaExists(personaName) {
			fmt.Printf("Persona '%s' does not exist.\n", personaName)
			return
		}

		if err := storageManager.DeletePersona(personaName); err != nil {
			fmt.Printf("Error deleting persona: %v\n", err)
			return
		}

		if outputJSON {
			result := map[string]any{
				"persona": personaName,
				"status":  "deleted",
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
			fmt.Printf("Persona '%s' deleted successfully\n", personaName)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.TitleStyle.Render("Persona deleted:"))
		fmt.Println(ui.ContentStyle.Render(personaName))
		fmt.Println()
	},
}

var showCmd = &cobra.Command{
	Use:   "show [nom]",
	Short: "Show persona details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		personaName := args[0]

		currentPersona, err := storageManager.GetPersona(personaName)
		if err != nil {
			fmt.Printf("Error loading persona: %v\n", err)
			return
		}

		if outputJSON {
			personaData := map[string]any{
				"name":               currentPersona.Name,
				"voice_name":         currentPersona.Voice.Name,
				"voice_instructions": currentPersona.Voice.Instructions,
				"prompt":             currentPersona.Prompt,
				"history_count":      len(currentPersona.History),
			}
			data, err := json.MarshalIndent(personaData, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			fmt.Printf("Name: %s\n", currentPersona.Name)
			fmt.Printf("Voice: %s\n", currentPersona.Voice.Name)
			fmt.Printf("History: %d messages\n", len(currentPersona.History))
			fmt.Println()
			fmt.Println("Instructions:")
			fmt.Println(currentPersona.Voice.Instructions)
			fmt.Println("Prompt:")
			fmt.Println(currentPersona.Prompt)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("Persona: %s", currentPersona.Name)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Name: %s", currentPersona.Name)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Voice: %s", currentPersona.Voice.Name)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("History: %d messages", len(currentPersona.History))))
		fmt.Println()
		fmt.Println(ui.TitleStyle.Render("Instructions:"))
		fmt.Println(ui.ContentStyle.Render(currentPersona.Voice.Instructions))
		fmt.Println(ui.TitleStyle.Render("Prompt:"))
		fmt.Println(ui.ContentStyle.Render(currentPersona.Prompt))
		fmt.Println()
	},
}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Show default persona",
	Run: func(cmd *cobra.Command, args []string) {
		defaultPersona, err := storageManager.GetDefaultPersona()
		if err != nil {
			fmt.Printf("Error loading default persona: %v\n", err)
			return
		}

		if outputJSON {
			personaData := map[string]any{
				"name":               defaultPersona.Name,
				"voice_name":         defaultPersona.Voice.Name,
				"voice_instructions": defaultPersona.Voice.Instructions,
				"prompt":             defaultPersona.Prompt,
				"history_count":      len(defaultPersona.History),
				"is_default":         true,
			}
			data, err := json.MarshalIndent(personaData, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			fmt.Printf("Default persona: %s\n", defaultPersona.Name)
			return
		}

		fmt.Println(ui.TitleStyle.Render("Default persona:"))
		fmt.Println(ui.ContentStyle.Render(defaultPersona.Name))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(defaultCmd)

	// Add output format flags
	listCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	listCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	createCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	createCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	deleteCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	deleteCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	showCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	showCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")

	defaultCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	defaultCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple plain text output")
}
