package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/config"
	"github.com/donnellyk/lazyruin/pkg/gui"
)

// App is the main application struct that bootstraps and runs lazyruin.
type App struct {
	Config        *config.Config
	RuinCmd       *commands.RuinCommand
	Gui           *gui.Gui
	QuickCapture  bool   // when true, open directly into new note and exit on save
	QuickLink     bool   // when true, open directly into new link and exit on save
	QuickLinkURL  string // when set with QuickLink, skip input popup and resolve directly
	DebugBindings bool   // when true, print all registered bindings and exit
	OpenRef       string // note path/title or parent bookmark to open on launch
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
	// Check ruin CLI version first — an outdated ruin binary may cause
	// CheckVault to fail (e.g., if `ruin today` returns a different error
	// format), so users deserve to see the upgrade hint rather than an
	// opaque vault error. Warning only, never blocks startup.
	versionWarning := ""
	if ok, got, err := a.RuinCmd.CheckVersion(); !ok {
		if err != nil {
			versionWarning = "could not determine ruin version — run `ruin --version` to check"
		} else {
			versionWarning = fmt.Sprintf("ruin %s < %s — run `brew upgrade ruin-cli`", got, commands.MinRuinVersion)
		}
	}

	if err := a.RuinCmd.CheckVault(); err != nil {
		if versionWarning != "" {
			return fmt.Errorf("%s\n%w", versionWarning, err)
		}
		return err
	}

	// Initialize GUI
	a.Gui = gui.NewGui(a.Config, a.RuinCmd)
	a.Gui.QuickCapture = a.QuickCapture
	a.Gui.QuickLink = a.QuickLink
	a.Gui.QuickLinkURL = a.QuickLinkURL
	a.Gui.OpenRef = a.OpenRef
	if versionWarning != "" {
		a.Gui.SetStartupWarning(versionWarning)
	}

	// Debug mode: print all registered bindings and exit without running the TUI.
	if a.DebugBindings {
		for _, b := range a.Gui.DumpBindings() {
			fmt.Println(b)
		}
		return nil
	}

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
