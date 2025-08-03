package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctrl-vfr/persona/internal/ui"
)

// Format represents the output format type
type Format string

const (
	// FormatDefault uses the default UI formatting
	FormatDefault Format = "default"
	// FormatJSON outputs data in JSON format
	FormatJSON Format = "json"
	// FormatPlain outputs plain text without formatting
	FormatPlain Format = "plain"
)

// Formatter handles different output formats
type Formatter struct {
	format Format
}

// New creates a new formatter with the specified format
func New(format Format) *Formatter {
	return &Formatter{format: format}
}

// Output formats and prints the given data according to the formatter's format
func (f *Formatter) Output(data any) error {
	switch f.format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatPlain:
		return f.outputPlain(data)
	default:
		return f.outputDefault(data)
	}
}

// outputJSON outputs data in JSON format
func (f *Formatter) outputJSON(data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputPlain outputs data in plain text format
func (f *Formatter) outputPlain(data any) error {
	switch v := data.(type) {
	case string:
		fmt.Println(v)
	case []string:
		for _, item := range v {
			fmt.Println(item)
		}
	case map[string]any:
		for key, value := range v {
			fmt.Printf("%s: %v\n", key, value)
		}
	default:
		fmt.Printf("%v\n", v)
	}
	return nil
}

// outputDefault outputs data using the default UI formatting
func (f *Formatter) outputDefault(data any) error {
	switch v := data.(type) {
	case string:
		fmt.Println(v)
	case []string:
		terminalWidth := ui.GetTerminalWidth()
		fmt.Println(ui.RenderPersonaListBox(v, terminalWidth))
	case map[string]any:
		var content strings.Builder
		for key, value := range v {
			content.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		fmt.Println(ui.ContentStyle.Render(content.String()))
	default:
		fmt.Printf("%v\n", v)
	}
	return nil
}

// Error outputs an error message according to the format
func (f *Formatter) Error(message string) {
	switch f.format {
	case FormatJSON:
		errorData := map[string]string{"error": message}
		jsonData, _ := json.MarshalIndent(errorData, "", "  ")
		fmt.Println(string(jsonData))
	case FormatPlain:
		fmt.Printf("Error: %s\n", message)
	default:
		fmt.Println(ui.RenderError(message))
	}
}

// Success outputs a success message according to the format
func (f *Formatter) Success(message string) {
	switch f.format {
	case FormatJSON:
		successData := map[string]string{"status": "success", "message": message}
		jsonData, _ := json.MarshalIndent(successData, "", "  ")
		fmt.Println(string(jsonData))
	case FormatPlain:
		fmt.Println(message)
	default:
		fmt.Println(ui.RenderSuccess(message))
	}
}

// Info outputs an info message according to the format
func (f *Formatter) Info(message string) {
	switch f.format {
	case FormatJSON:
		infoData := map[string]string{"info": message}
		jsonData, _ := json.MarshalIndent(infoData, "", "  ")
		fmt.Println(string(jsonData))
	case FormatPlain:
		fmt.Println(message)
	default:
		fmt.Println(ui.RenderInfo(message))
	}
}

// ParseFormat parses a string into a Format type
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "plain":
		return FormatPlain
	default:
		return FormatDefault
	}
}
