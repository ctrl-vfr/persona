package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	// Always save as YAML for better readability
	data, err := yaml.Marshal(p.History)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
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
	// Always save as YAML for better readability
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
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
