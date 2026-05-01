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

// NotesPaneSectionItem describes one selectable item inside a custom section
// of the Home tab. The embed is sent verbatim to `ruin embed eval`.
type NotesPaneSectionItem struct {
	Title string `yaml:"title"`
	Embed string `yaml:"embed"`
}

// NotesPaneSection describes one custom section in the Home tab. Title is
// optional; an empty title renders the items without a header (a blank line
// still separates it from neighbouring sections).
type NotesPaneSection struct {
	Title string                 `yaml:"title,omitempty"`
	Items []NotesPaneSectionItem `yaml:"items"`
}

// NotesPaneConfig configures the Notes pane Home tab. SectionsMode toggles
// the Home/Notes outer-tab UX off (default) or on. CustomSections adds
// user-defined sections below the hardcoded ones (Inbox / Today / Next 7
// Days / Pinned).
type NotesPaneConfig struct {
	SectionsMode   bool               `yaml:"sections_mode"`
	CustomSections []NotesPaneSection `yaml:"custom_sections,omitempty"`
}

// Config holds the application configuration.
type Config struct {
	VaultPath   string          `yaml:"vault_path"`
	Editor      string          `yaml:"editor"`
	ChromaTheme string          `yaml:"chroma_theme"`
	ViewOptions ViewOptions     `yaml:"view_options,omitempty"`
	NotesPane   NotesPaneConfig `yaml:"notes_pane,omitempty"`

	// SidebarWidth overrides the side panel width in columns. When 0 (or
	// unset), the layout uses min(maxX/3, 40). Clamped at runtime so the
	// preview pane keeps a usable minimum width.
	SidebarWidth int `yaml:"sidebar_width,omitempty"`

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
