package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	VaultPath     string                       `yaml:"vault_path"`
	Editor        string                       `yaml:"editor"`
	ChromaTheme   string                       `yaml:"chroma_theme"`
	Abbreviations map[string]map[string]string `yaml:"abbreviations,omitempty"`
}

// VaultAbbreviations returns the snippet map for the given vault.
// Falls back to the legacy "" key if the vault has no snippets.
// Never returns nil.
func (cfg *Config) VaultAbbreviations(vaultPath string) map[string]string {
	if cfg.Abbreviations != nil {
		if m, ok := cfg.Abbreviations[vaultPath]; ok && len(m) > 0 {
			return m
		}
		if m, ok := cfg.Abbreviations[""]; ok && len(m) > 0 {
			return m
		}
	}
	return map[string]string{}
}

// SetVaultAbbreviation sets a snippet for the given vault.
func (cfg *Config) SetVaultAbbreviation(vaultPath, name, expansion string) {
	if cfg.Abbreviations == nil {
		cfg.Abbreviations = make(map[string]map[string]string)
	}
	if cfg.Abbreviations[vaultPath] == nil {
		cfg.Abbreviations[vaultPath] = make(map[string]string)
	}
	cfg.Abbreviations[vaultPath][name] = expansion
}

// DeleteVaultAbbreviation deletes a snippet for the given vault.
func (cfg *Config) DeleteVaultAbbreviation(vaultPath, name string) {
	if cfg.Abbreviations == nil {
		return
	}
	if m, ok := cfg.Abbreviations[vaultPath]; ok {
		delete(m, name)
		if len(m) == 0 {
			delete(cfg.Abbreviations, vaultPath)
		}
	}
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

	// Detect old flat abbreviations format before full unmarshal.
	var raw struct {
		Abbreviations interface{} `yaml:"abbreviations"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	needsMigration := false
	if raw.Abbreviations != nil {
		if flat, ok := raw.Abbreviations.(map[string]interface{}); ok {
			// Check if values are strings (old format) vs maps (new format).
			for _, v := range flat {
				if _, isStr := v.(string); isStr {
					needsMigration = true
				}
				break
			}
		}
	}

	if needsMigration {
		// Unmarshal with old format, then migrate.
		var oldCfg struct {
			VaultPath     string            `yaml:"vault_path"`
			Editor        string            `yaml:"editor"`
			ChromaTheme   string            `yaml:"chroma_theme"`
			Abbreviations map[string]string `yaml:"abbreviations,omitempty"`
		}
		if err := yaml.Unmarshal(data, &oldCfg); err != nil {
			return nil, err
		}
		cfg.VaultPath = oldCfg.VaultPath
		cfg.Editor = oldCfg.Editor
		cfg.ChromaTheme = oldCfg.ChromaTheme
		if len(oldCfg.Abbreviations) > 0 {
			cfg.Abbreviations = map[string]map[string]string{
				"": oldCfg.Abbreviations,
			}
			// Re-save in new format.
			_ = cfg.Save()
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
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
