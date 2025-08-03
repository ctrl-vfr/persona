package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseFormat(t *testing.T) {
	testCases := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"plain", FormatPlain},
		{"PLAIN", FormatPlain},
		{"default", FormatDefault},
		{"DEFAULT", FormatDefault},
		{"invalid", FormatDefault},
		{"", FormatDefault},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseFormat(tc.input)
			if result != tc.expected {
				t.Errorf("ParseFormat(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatterOutput(t *testing.T) {
	testCases := []struct {
		name     string
		format   Format
		data     any
		contains []string
	}{
		{
			name:     "JSON string",
			format:   FormatJSON,
			data:     "test string",
			contains: []string{`"test string"`},
		},
		{
			name:   "JSON map",
			format: FormatJSON,
			data:   map[string]any{"key": "value"},

			contains: []string{`"key"`, `"value"`},
		},
		{
			name:     "Plain string",
			format:   FormatPlain,
			data:     "test string",
			contains: []string{"test string"},
		},
		{
			name:     "Plain slice",
			format:   FormatPlain,
			data:     []string{"item1", "item2"},
			contains: []string{"item1", "item2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := New(tc.format)
			err := formatter.Output(tc.data)
			if err != nil {
				t.Fatalf("Output() error = %v", err)
			}

			// Restore stdout
			err = w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("Error copying output: %v", err)
			}
			output := buf.String()

			// Check if output contains expected strings
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestFormatterError(t *testing.T) {
	testCases := []struct {
		name     string
		format   Format
		message  string
		contains []string
	}{
		{
			name:     "JSON error",
			format:   FormatJSON,
			message:  "test error",
			contains: []string{`"error"`, `"test error"`},
		},
		{
			name:     "Plain error",
			format:   FormatPlain,
			message:  "test error",
			contains: []string{"Error:", "test error"},
		},
		{
			name:     "Default error",
			format:   FormatDefault,
			message:  "test error",
			contains: []string{"test error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := New(tc.format)
			formatter.Error(tc.message)

			// Restore stdout
			err := w.Close()
			if err != nil {
				t.Fatalf("Error closing pipe: %v", err)
			}
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("Error copying output: %v", err)
			}
			output := buf.String()

			// Check if output contains expected strings
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Error output should contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestFormatterSuccess(t *testing.T) {
	formatter := New(FormatJSON)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.Success("test success")

	// Restore stdout
	err := w.Close()
	if err != nil {
		t.Fatalf("Error closing pipe: %v", err)
	}
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Error copying output: %v", err)
	}
	output := buf.String()

	// Parse JSON to verify structure
	var result map[string]string
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got %q", result["status"])
	}
	if result["message"] != "test success" {
		t.Errorf("Expected message 'test success', got %q", result["message"])
	}
}

func TestNew(t *testing.T) {
	formatter := New(FormatJSON)
	if formatter == nil {
		t.Fatal("New() returned nil")
	}
	if formatter.format != FormatJSON {
		t.Errorf("Expected format %v, got %v", FormatJSON, formatter.format)
	}
}
