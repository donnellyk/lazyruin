package app

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/gui"
)

// App is the main application struct that bootstraps and runs lazyruin.
type App struct {
	Config  *config.Config
	RuinCmd *commands.RuinCommand
	Gui     *gui.Gui
}

// NewApp creates a new application instance.
// vaultOverride can be empty to use default resolution.
func NewApp(vaultOverride string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	vaultPath, err := resolveVaultPath(cfg, vaultOverride)
	if err != nil {
		return nil, err
	}

	ruinCmd := commands.NewRuinCommand(vaultPath)

	return &App{
		Config:  cfg,
		RuinCmd: ruinCmd,
	}, nil
}

// Run starts the application.
func (a *App) Run() error {
	// Verify vault exists
	if !a.RuinCmd.VaultExists() {
		return errors.New("vault not found - run 'ruin init' or set vault path in config")
	}

	// Initialize and run GUI
	a.Gui = gui.NewGui(a.RuinCmd)
	return a.Gui.Run()
}

// resolveVaultPath determines the vault path from CLI flag, config, env, or ruin CLI.
func resolveVaultPath(cfg *config.Config, cliOverride string) (string, error) {
	// 1. Check CLI flag (highest priority)
	if cliOverride != "" {
		return cliOverride, nil
	}

	// 2. Check config
	if cfg.VaultPath != "" {
		return cfg.VaultPath, nil
	}

	// 3. Check environment
	if envVault := os.Getenv("LAZYRUIN_VAULT"); envVault != "" {
		return envVault, nil
	}

	// 4. Ask ruin CLI for its configured vault path
	cmd := exec.Command("ruin", "config", "vault_path")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("could not determine vault path - set LAZYRUIN_VAULT or configure vault_path")
	}

	return strings.TrimSpace(string(output)), nil
}
