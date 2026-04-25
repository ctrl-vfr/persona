package persona

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// HistoryFileMode and PersonaFileMode mirror storage.UserFileMode.
// Duplicated here to avoid an import cycle (storage imports persona).
const (
	HistoryFileMode os.FileMode = 0o600
	PersonaFileMode os.FileMode = 0o600
)

// writeFileAtomic writes data atomically (tmp + rename in same dir).
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

type Persona struct {
	Name    string    `yaml:"name" json:"name"`
	Voice   Voice     `yaml:"voice" json:"voice"`
	Prompt  string    `yaml:"prompt" json:"prompt"`
	History []Message `yaml:"history,omitempty" json:"history,omitempty"`
}

type Voice struct {
	Name         string `yaml:"name" json:"name"`
	Instructions string `yaml:"instructions" json:"instructions"`
}

type Message struct {
	Role    string `yaml:"role" json:"role"`
	Content string `yaml:"content" json:"content"`
}

func New(name string, voice Voice, prompt string) *Persona {
	return &Persona{
		Name:    name,
		Voice:   voice,
		Prompt:  prompt,
		History: []Message{},
	}
}

func (p *Persona) AddMessage(message Message, limit int) {
	if len(p.History) >= limit {
		p.History = p.History[1:]
	}
	p.History = append(p.History, message)
}

func (p *Persona) ClearHistory() {
	p.History = []Message{}
}

func (p *Persona) SaveHistory(path string) error {
	data, err := yaml.Marshal(p.History)
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	return writeFileAtomic(path, data, HistoryFileMode)
}

func (p *Persona) LoadHistory(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Support both YAML and JSON for backward compatibility
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, &p.History)
	} else {
		// Try YAML first, then JSON for backward compatibility
		err = yaml.Unmarshal(data, &p.History)
		if err != nil {
			// Fallback to JSON if YAML parsing fails
			err = json.Unmarshal(data, &p.History)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (p *Persona) SavePersona(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal persona: %w", err)
	}
	return writeFileAtomic(path, data, PersonaFileMode)
}

func (p *Persona) LoadPersona(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Support both YAML and JSON for backward compatibility
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, p)
	} else {
		// Try YAML first, then JSON for backward compatibility
		err = yaml.Unmarshal(data, p)
		if err != nil {
			// Fallback to JSON if YAML parsing fails
			err = json.Unmarshal(data, p)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (p *Persona) GetMessages() []Message {
	prompt := Message{
		Role:    "system",
		Content: p.Prompt,
	}

	history := []Message{prompt}
	history = append(history, p.History...)
	return history
}
