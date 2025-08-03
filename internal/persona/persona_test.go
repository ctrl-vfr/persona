package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	voice := Voice{
		Name:         "nova",
		Instructions: "Speak with a friendly tone",
	}

	p := New("test-persona", voice, "You are a helpful assistant")

	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.Name != "test-persona" {
		t.Errorf("Expected name 'test-persona', got '%s'", p.Name)
	}
	if p.Voice.Name != "nova" {
		t.Errorf("Expected voice 'nova', got '%s'", p.Voice.Name)
	}
	if p.Prompt != "You are a helpful assistant" {
		t.Errorf("Expected specific prompt, got '%s'", p.Prompt)
	}
	if len(p.History) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(p.History))
	}
}

func TestPersona_AddMessage(t *testing.T) {
	p := New("test", Voice{Name: "nova"}, "test prompt")

	// Test adding messages within limit
	message1 := Message{Role: "user", Content: "Hello"}
	message2 := Message{Role: "assistant", Content: "Hi there!"}

	p.AddMessage(message1, 10)
	if len(p.History) != 1 {
		t.Errorf("Expected 1 message, got %d", len(p.History))
	}

	p.AddMessage(message2, 10)
	if len(p.History) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(p.History))
	}

	// Verify message content
	if p.History[0].Role != "user" || p.History[0].Content != "Hello" {
		t.Errorf("First message incorrect: %+v", p.History[0])
	}
	if p.History[1].Role != "assistant" || p.History[1].Content != "Hi there!" {
		t.Errorf("Second message incorrect: %+v", p.History[1])
	}
}

func TestPersona_AddMessage_ExceedsLimit(t *testing.T) {
	p := New("test", Voice{Name: "nova"}, "test prompt")

	// Add messages up to limit
	for i := 0; i < 5; i++ {
		msg := Message{Role: "user", Content: fmt.Sprintf("Message %d", i)}
		p.AddMessage(msg, 3) // Limit of 3
	}

	// Should only have 3 messages (the last 3)
	if len(p.History) != 3 {
		t.Errorf("Expected 3 messages after limit, got %d", len(p.History))
	}

	// Verify we have the last 3 messages
	expectedContents := []string{"Message 2", "Message 3", "Message 4"}
	for i, expected := range expectedContents {
		if p.History[i].Content != expected {
			t.Errorf("Message %d: expected '%s', got '%s'", i, expected, p.History[i].Content)
		}
	}
}

func TestPersona_ClearHistory(t *testing.T) {
	p := New("test", Voice{Name: "nova"}, "test prompt")

	// Add some messages
	p.AddMessage(Message{Role: "user", Content: "Hello"}, 10)
	p.AddMessage(Message{Role: "assistant", Content: "Hi"}, 10)

	if len(p.History) == 0 {
		t.Fatal("History should not be empty before clearing")
	}

	p.ClearHistory()

	if len(p.History) != 0 {
		t.Errorf("Expected empty history after clear, got %d messages", len(p.History))
	}
}

func TestPersona_GetMessages(t *testing.T) {
	p := New("test", Voice{Name: "nova"}, "You are helpful")

	// Add some messages
	p.AddMessage(Message{Role: "user", Content: "Question 1"}, 10)
	p.AddMessage(Message{Role: "assistant", Content: "Answer 1"}, 10)

	messages := p.GetMessages()

	// Should include system message + user messages
	expectedLen := 3 // system + user + assistant
	if len(messages) != expectedLen {
		t.Errorf("Expected %d messages, got %d", expectedLen, len(messages))
	}

	// First message should be system message
	if messages[0].Role != "system" {
		t.Errorf("First message should be system, got '%s'", messages[0].Role)
	}
	if messages[0].Content != "You are helpful" {
		t.Errorf("System message content incorrect: '%s'", messages[0].Content)
	}

	// Verify user and assistant messages
	if messages[1].Role != "user" || messages[1].Content != "Question 1" {
		t.Errorf("User message incorrect: %+v", messages[1])
	}
	if messages[2].Role != "assistant" || messages[2].Content != "Answer 1" {
		t.Errorf("Assistant message incorrect: %+v", messages[2])
	}
}

func TestPersona_SaveAndLoadHistory(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "test_history.yaml")

	// Create persona with history
	p := New("test", Voice{Name: "nova"}, "test prompt")
	p.AddMessage(Message{Role: "user", Content: "Test question"}, 10)
	p.AddMessage(Message{Role: "assistant", Content: "Test answer"}, 10)

	// Save history
	err := p.SaveHistory(historyPath)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Load history into new persona
	newPersona := New("test2", Voice{Name: "alloy"}, "different prompt")
	err = newPersona.LoadHistory(historyPath)
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	// Verify loaded history
	if len(newPersona.History) != 2 {
		t.Errorf("Expected 2 messages in loaded history, got %d", len(newPersona.History))
	}

	if newPersona.History[0].Content != "Test question" {
		t.Errorf("First message content mismatch: '%s'", newPersona.History[0].Content)
	}
	if newPersona.History[1].Content != "Test answer" {
		t.Errorf("Second message content mismatch: '%s'", newPersona.History[1].Content)
	}
}

func TestPersona_LoadInvalidHistory(t *testing.T) {
	p := New("test", Voice{Name: "nova"}, "test prompt")

	// Try to load nonexistent file
	err := p.LoadHistory("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error loading nonexistent history file, got nil")
	}
}

func TestPersona_SaveAndLoadPersona(t *testing.T) {
	tempDir := t.TempDir()
	personaPath := filepath.Join(tempDir, "test_persona.yaml")

	// Create persona
	voice := Voice{
		Name:         "echo",
		Instructions: "Speak dramatically",
	}
	originalPersona := New("dramatic-assistant", voice, "You are a dramatic assistant")
	originalPersona.AddMessage(Message{Role: "user", Content: "Hello"}, 10)

	// Save persona
	err := originalPersona.SavePersona(personaPath)
	if err != nil {
		t.Fatalf("Failed to save persona: %v", err)
	}

	// Load persona
	loadedPersona := &Persona{}
	err = loadedPersona.LoadPersona(personaPath)
	if err != nil {
		t.Fatalf("Failed to load persona: %v", err)
	}

	// Verify loaded persona
	if loadedPersona.Name != originalPersona.Name {
		t.Errorf("Name mismatch: expected '%s', got '%s'", originalPersona.Name, loadedPersona.Name)
	}
	if loadedPersona.Voice.Name != originalPersona.Voice.Name {
		t.Errorf("Voice name mismatch: expected '%s', got '%s'", originalPersona.Voice.Name, loadedPersona.Voice.Name)
	}
	if loadedPersona.Prompt != originalPersona.Prompt {
		t.Errorf("Prompt mismatch: expected '%s', got '%s'", originalPersona.Prompt, loadedPersona.Prompt)
	}
	if len(loadedPersona.History) != len(originalPersona.History) {
		t.Errorf("History length mismatch: expected %d, got %d", len(originalPersona.History), len(loadedPersona.History))
	}
}
