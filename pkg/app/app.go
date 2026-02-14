package app

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/gui"
)

// App is the main application struct that bootstraps and runs lazyruin.
type App struct {
	Config       *config.Config
	RuinCmd      *commands.RuinCommand
	Gui          *gui.Gui
	QuickCapture bool // when true, open directly into new note and exit on save
}

// NewApp creates a new application instance.
// vaultOverride and ruinBin can be empty to use default resolution.
func NewApp(vaultOverride, ruinBin string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if ruinBin == "" {
		ruinBin = "ruin"
	}

	vaultPath, err := resolveVaultPath(cfg, vaultOverride, ruinBin)
	if err != nil {
		return nil, err
	}

	ruinCmd := commands.NewRuinCommand(vaultPath, ruinBin)

	return &App{
		Config:  cfg,
		RuinCmd: ruinCmd,
	}, nil
}

// Run starts the application.
func (a *App) Run() error {
	if err := a.RuinCmd.CheckVault(); err != nil {
		return err
	}

	// Initialize and run GUI
	a.Gui = gui.NewGui(a.Config, a.RuinCmd)
	a.Gui.QuickCapture = a.QuickCapture
	return a.Gui.Run()
}

// expandPath expands ~ to the user's home directory and resolves to absolute path.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// resolveVaultPath determines the vault path from CLI flag, config, env, or ruin CLI.
func resolveVaultPath(cfg *config.Config, cliOverride, ruinBin string) (string, error) {
	// 1. Check CLI flag (highest priority)
	if cliOverride != "" {
		return expandPath(cliOverride), nil
	}

	// 2. Check config
	if cfg.VaultPath != "" {
		return expandPath(cfg.VaultPath), nil
	}

	// 3. Check environment
	if envVault := os.Getenv("LAZYRUIN_VAULT"); envVault != "" {
		return expandPath(envVault), nil
	}

	// 4. Ask ruin CLI for its configured vault path
	cmd := exec.Command(ruinBin, "config", "vault_path")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("could not determine vault path - set LAZYRUIN_VAULT or configure vault_path")
	}

	return strings.TrimSpace(string(output)), nil
}
