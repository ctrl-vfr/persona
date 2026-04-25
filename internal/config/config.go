// Package config holds the persona application configuration and
// load/save helpers. Configuration is YAML-first; JSON is accepted on
// load only for backward compatibility with early dev builds.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileMode os.FileMode = 0o600

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

// Validate enforces basic invariants on a loaded config so we fail
// early with an explicit error instead of letting bad values reach
// ffmpeg or the OpenAI client.
func (c *Config) Validate() error {
	if c.Models.Transcription == "" {
		return fmt.Errorf("models.transcription is empty")
	}
	if c.Models.Speech == "" {
		return fmt.Errorf("models.speech is empty")
	}
	if c.Models.Chat == "" {
		return fmt.Errorf("models.chat is empty")
	}
	// silencedetect threshold is in dB; positive values are nonsensical
	// and very negative values disable detection in practice.
	if c.Audio.SilenceThreshold < -90 || c.Audio.SilenceThreshold > 0 {
		return fmt.Errorf("audio.silence_threshold %d outside [-90, 0] dB", c.Audio.SilenceThreshold)
	}
	if c.Audio.SilenceDuration < 1 || c.Audio.SilenceDuration > 30 {
		return fmt.Errorf("audio.silence_duration %d outside [1, 30] seconds", c.Audio.SilenceDuration)
	}
	// Reject control characters / NUL in input device name to keep
	// ffmpeg arg construction safe even though we don't use a shell.
	for _, r := range c.Audio.InputDevice {
		if r == 0 || (r < 0x20 && r != '\t') {
			return fmt.Errorf("audio.input_device contains control character")
		}
	}
	return nil
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
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return writeFileAtomic(path, data, configFileMode)
}
