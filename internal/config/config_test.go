package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}
}

func TestConfig_LoadYAML(t *testing.T) {
	// Create temporary YAML config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `models:
  transcription: "whisper-1"
  speech: "tts-1"
  chat: "gpt-4"
audio:
  input_device: "microphone"
  output_device: "speakers"
  silence_threshold: -40
  silence_duration: 3`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	config := NewConfig()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	// Verify loaded values
	if config.Models.Transcription != "whisper-1" {
		t.Errorf("Expected transcription model 'whisper-1', got '%s'", config.Models.Transcription)
	}
	if config.Audio.SilenceThreshold != -40 {
		t.Errorf("Expected silence threshold -40, got %d", config.Audio.SilenceThreshold)
	}
	if config.Audio.SilenceDuration != 3 {
		t.Errorf("Expected silence duration 3, got %d", config.Audio.SilenceDuration)
	}
}

func TestConfig_LoadInvalidFile(t *testing.T) {
	config := NewConfig()
	err := config.Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error loading nonexistent file, got nil")
	}
}

func TestConfig_LoadInvalidFormat(t *testing.T) {
	// Create temporary file with invalid content
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	invalidContent := `invalid: yaml: content: [unclosed bracket`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config := NewConfig()
	err = config.Load(configPath)
	if err == nil {
		t.Error("Expected error loading invalid YAML, got nil")
	}
}

func TestConfig_SaveYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	config := NewConfig()
	config.Models.Transcription = "whisper-1"
	config.Models.Speech = "tts-1-hd"
	config.Models.Chat = "gpt-4"
	config.Audio.InputDevice = "test-mic"
	config.Audio.SilenceThreshold = -35

	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify content
	newConfig := NewConfig()
	err = newConfig.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if newConfig.Models.Transcription != config.Models.Transcription {
		t.Errorf("Transcription model mismatch: expected '%s', got '%s'",
			config.Models.Transcription, newConfig.Models.Transcription)
	}
	if newConfig.Audio.SilenceThreshold != config.Audio.SilenceThreshold {
		t.Errorf("Silence threshold mismatch: expected %d, got %d",
			config.Audio.SilenceThreshold, newConfig.Audio.SilenceThreshold)
	}
}

func TestConfig_SaveToInvalidPath(t *testing.T) {
	config := NewConfig()
	err := config.Save("/invalid/path/config.yaml")
	if err == nil {
		t.Error("Expected error saving to invalid path, got nil")
	}
}
