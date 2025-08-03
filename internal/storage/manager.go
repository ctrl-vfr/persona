package storage

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctrl-vfr/persona/internal/config"
	"github.com/ctrl-vfr/persona/internal/persona"

	"gopkg.in/yaml.v3"
)

// Built-in persona templates
//

//go:embed assets/config.yaml
var defaultConfigYAML []byte

//go:embed assets/personas/marceline.yaml
var marcelineYAML []byte

//go:embed assets/personas/freud.yaml
var freudYAML []byte

//go:embed assets/personas/coach.yaml
var coachYAML []byte

//go:embed assets/personas/kevin.yaml
var kevinYAML []byte

//go:embed assets/personas/merlin.yaml
var merlinYAML []byte

//go:embed assets/personas/racoon.yaml
var racoonYAML []byte

type Manager struct {
	BasePath string
}

// BuiltinPersona represents a built-in persona template
type BuiltinPersona struct {
	Name string
	Data []byte
}

// GetBuiltinPersonas returns all built-in persona templates
func GetBuiltinPersonas() []BuiltinPersona {
	return []BuiltinPersona{
		{"marceline", marcelineYAML},
		{"freud", freudYAML},
		{"coach", coachYAML},
		{"kevin", kevinYAML},
		{"merlin", merlinYAML},
		{"racoon", racoonYAML},
	}
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".persona")
	return &Manager{BasePath: basePath}, nil
}

// InitializeStructure creates the default directory structure and files if they don't exist
func (m *Manager) InitializeStructure() error {
	// Create base directory
	if err := os.MkdirAll(m.BasePath, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create personas directory
	personasDir := filepath.Join(m.BasePath, "personas")
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		return fmt.Errorf("failed to create personas directory: %w", err)
	}

	// Create default config.yaml if it doesn't exist
	configPath := m.GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// defaultConfigYAML is already YAML data, write directly
		if err := os.WriteFile(configPath, defaultConfigYAML, 0644); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Create default persona "persona" if it doesn't exist (using marceline as template)
	if err := m.CreatePersonaFromYAMLTemplate("persona", marcelineYAML); err != nil {
		return fmt.Errorf("failed to create persona persona: %w", err)
	}

	// Install built-in personas if they don't exist
	if err := m.InstallBuiltinPersonas(); err != nil {
		return fmt.Errorf("failed to install built-in personas: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func (m *Manager) GetConfigPath() string {
	return filepath.Join(m.BasePath, "config.yaml")
}

// GetPersonaPath returns the path to a persona's files
func (m *Manager) GetPersonaPath(name string) (personaPath, historyPath string) {
	personaDir := filepath.Join(m.BasePath, "personas", name)
	return filepath.Join(personaDir, "persona.yaml"), filepath.Join(personaDir, "history.yaml")
}

// GetConfig loads the configuration using the existing config module
func (m *Manager) GetConfig() (*config.Config, error) {
	cfg := config.NewConfig()
	if err := cfg.Load(m.GetConfigPath()); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

// SaveConfig saves the configuration using the existing config module
func (m *Manager) SaveConfig(cfg *config.Config) error {
	return cfg.Save(m.GetConfigPath())
}

// ListPersonas returns a list of all available persona names
func (m *Manager) ListPersonas() ([]string, error) {
	personasDir := filepath.Join(m.BasePath, "personas")
	entries, err := os.ReadDir(personasDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read personas directory: %w", err)
	}

	var personas []string
	for _, entry := range entries {
		if entry.IsDir() {
			personas = append(personas, entry.Name())
		}
	}

	return personas, nil
}

// CreatePersona creates a new persona with default template
func (m *Manager) CreatePersona(name string) error {
	return m.CreatePersonaFromYAMLTemplate(name, marcelineYAML)
}

// GetPersona loads a persona using the existing persona module
func (m *Manager) GetPersona(name string) (*persona.Persona, error) {
	personaPath, historyPath := m.GetPersonaPath(name)

	p := &persona.Persona{}
	if err := p.LoadPersona(personaPath); err != nil {
		return nil, fmt.Errorf("failed to load persona %s: %w", name, err)
	}

	// Load history if it exists
	if err := p.LoadHistory(historyPath); err != nil {
		// If history file doesn't exist, that's okay - just start with empty history
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load history for persona %s: %w", name, err)
		}
	}

	return p, nil
}

// SavePersona saves a persona using the existing persona module
func (m *Manager) SavePersona(name string, p *persona.Persona) error {
	personaDir := filepath.Join(m.BasePath, "personas", name)

	// Ensure directory exists
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		return fmt.Errorf("failed to create persona directory: %w", err)
	}

	personaPath, historyPath := m.GetPersonaPath(name)

	// Save persona using existing module
	if err := p.SavePersona(personaPath); err != nil {
		return fmt.Errorf("failed to save persona: %w", err)
	}

	// Save history using existing module
	if err := p.SaveHistory(historyPath); err != nil {
		return fmt.Errorf("failed to save history: %w", err)
	}

	return nil
}

// DeletePersona removes a persona and all its files
func (m *Manager) DeletePersona(name string) error {
	// Prevent deletion of the default persona persona
	if name == "persona" {
		return fmt.Errorf("cannot delete the default persona persona")
	}

	personaDir := filepath.Join(m.BasePath, "personas", name)
	return os.RemoveAll(personaDir)
}

// PersonaExists checks if a persona exists
func (m *Manager) PersonaExists(name string) bool {
	personaPath, _ := m.GetPersonaPath(name)
	_, err := os.Stat(personaPath)
	return !os.IsNotExist(err)
}

// GetDefaultPersona returns the default persona persona
func (m *Manager) GetDefaultPersona() (*persona.Persona, error) {
	return m.GetPersona("persona")
}

// InstallBuiltinPersonas installs all built-in personas if they don't exist
func (m *Manager) InstallBuiltinPersonas() error {
	for _, builtinPersona := range GetBuiltinPersonas() {
		if !m.PersonaExists(builtinPersona.Name) {
			if err := m.CreatePersonaFromYAMLTemplate(builtinPersona.Name, builtinPersona.Data); err != nil {
				return fmt.Errorf("failed to install built-in persona '%s': %w", builtinPersona.Name, err)
			}
		}
	}
	return nil
}

// CreatePersonaFromYAMLTemplate creates a persona from a YAML template if it doesn't exist
func (m *Manager) CreatePersonaFromYAMLTemplate(name string, template []byte) error {
	if m.PersonaExists(name) {
		return nil // Already exists
	}

	personaDir := filepath.Join(m.BasePath, "personas", name)

	// Create persona directory
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		return fmt.Errorf("failed to create persona directory: %w", err)
	}

	// Create persona.yaml from template
	personaPath, historyPath := m.GetPersonaPath(name)

	if err := os.WriteFile(personaPath, template, 0644); err != nil {
		return fmt.Errorf("failed to create persona file: %w", err)
	}

	// Create empty history.yaml (empty array)
	emptyHistory := []interface{}{}
	yamlHistoryData, err := yaml.Marshal(emptyHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal empty history: %w", err)
	}
	if err := os.WriteFile(historyPath, yamlHistoryData, 0644); err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}

	return nil
}
