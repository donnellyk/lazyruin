package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ViewOptions holds preview display preferences that persist across runs.
// Each field has an explicit YAML key so zero-values map cleanly to the
// "unset → default" case when a config file is older than a field.
type ViewOptions struct {
	HideDone bool `yaml:"hide_done"`
}

// Config holds the application configuration.
type Config struct {
	VaultPath   string      `yaml:"vault_path"`
	Editor      string      `yaml:"editor"`
	ChromaTheme string      `yaml:"chroma_theme"`
	ViewOptions ViewOptions `yaml:"view_options,omitempty"`

	// DisableBareURLAsLink opts out of the "submitting a note whose entire
	// body is a URL routes through the link-resolution flow" convenience.
	// Default false (feature on); set true to always take the plain
	// `ruin log` path regardless of content shape.
	DisableBareURLAsLink bool `yaml:"disable_bare_url_as_link,omitempty"`

	// OnboardingOffered is flipped to true after the empty-vault onboarding
	// prompt has been shown once (either accepted or declined), so we do not
	// re-prompt on subsequent launches against empty vaults.
	OnboardingOffered bool `yaml:"onboarding_offered,omitempty"`
}

// Save writes the configuration to the default config file.
func (cfg *Config) Save() error {
	configPath := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
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
