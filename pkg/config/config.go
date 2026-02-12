package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	VaultPath   string `yaml:"vault_path"`
	ChromaTheme string `yaml:"chroma_theme"`
}

// Load reads the configuration from the default config file.
// Returns default config if file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// getConfigPath returns the path to the config file.
func getConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazyruin", "config.yml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazyruin", "config.yml")
}
