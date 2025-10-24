package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the game configuration
type Config struct {
	MaxRounds int      `yaml:"max_rounds"`
	WordList  []string `yaml:"word_list"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Validate configuration
	if config.MaxRounds <= 0 {
		return nil, errors.New("max_rounds must be positive")
	}

	if len(config.WordList) == 0 {
		return nil, errors.New("word list cannot be empty")
	}

	return &config, nil
}

// LoadWordsFromFile loads words from a text file (one word per line)
func LoadWordsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := []string{}
	current := ""
	for _, b := range data {
		if b == '\n' || b == '\r' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(b)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxRounds: 6,
		WordList: []string{
			"CRANE", "SLATE", "ABOUT", "APPLE", "HOUSE",
			"WORLD", "THINK", "GREAT", "PLACE", "BRAIN",
			"PHONE", "SMILE", "LIGHT", "PEACE", "DREAM",
			"OCEAN", "PIANO", "BREAD", "MUSIC", "TABLE",
		},
	}
}
