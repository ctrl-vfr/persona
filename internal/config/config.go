// Config package for persona configuration
// It contains the configuration structure and methods to load and save it
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Models struct {
		Transcription string `yaml:"transcription"`
		Speech        string `yaml:"speech"`
		Chat          string `yaml:"chat"`
	} `yaml:"models"`
	Audio struct {
		InputDevice      string `yaml:"input_device"`
		OutputDevice     string `yaml:"output_device"`
		SilenceThreshold int    `yaml:"silence_threshold"`
		SilenceDuration  int    `yaml:"silence_duration"`
	} `yaml:"audio"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Support both YAML and JSON for backward compatibility
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, c)
	} else {
		// Try YAML first, then JSON for backward compatibility
		err = yaml.Unmarshal(data, c)
		if err != nil {
			// Fallback to JSON if YAML parsing fails
			err = json.Unmarshal(data, c)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Save(path string) error {
	// Always save as YAML for better readability
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
