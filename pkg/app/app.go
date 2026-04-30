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
	helperspkg "github.com/donnellyk/lazyruin/pkg/gui/helpers"
	"github.com/donnellyk/lazyruin/pkg/gui/onboarding"
	"github.com/donnellyk/lazyruin/pkg/migrations"
)

// App is the main application struct that bootstraps and runs lazyruin.
type App struct {
	Config          *config.Config
	RuinCmd         *commands.RuinCommand
	Gui             *gui.Gui
	VaultSource     string // human-readable source of the resolved vault path
	LazyruinVersion string // build-time version, used to detect upgrades
	QuickCapture    bool   // when true, open directly into new note and exit on save
	QuickLink       bool   // when true, open directly into new link and exit on save
	QuickLinkURL    string // when set with QuickLink, skip input popup and resolve directly
	DebugBindings   bool   // when true, print all registered bindings and exit
	OpenRef         string // note path/title or parent bookmark to open on launch
}

// NewApp creates a new application instance.
// vaultOverride and ruinBin can be empty to use default resolution.
// lazyruinVersion is the build-time version constant (or "dev"); it
// drives the upgrade-migration detector.
func NewApp(vaultOverride, ruinBin, lazyruinVersion string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if ruinBin == "" {
		ruinBin = "ruin"
	}

	vaultPath, source, err := resolveVaultPath(cfg, vaultOverride, ruinBin)
	if err != nil {
		return nil, err
	}

	ruinCmd := commands.NewRuinCommand(vaultPath, ruinBin)

	return &App{
		Config:          cfg,
		RuinCmd:         ruinCmd,
		VaultSource:     source,
		LazyruinVersion: lazyruinVersion,
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

	// When the vault path exists but hasn't been initialized as a ruin
	// vault, skip CheckVault (which would fail) and let the TUI prompt
	// the user to run `ruin init` via the init dialog.
	needsInit := !a.RuinCmd.IsInitialized()
	if !needsInit {
		if err := a.RuinCmd.CheckVault(); err != nil {
			if versionWarning != "" {
				return fmt.Errorf("%s\n%w", versionWarning, err)
			}
			return err
		}
	}

	// Initialize GUI
	a.Gui = gui.NewGui(a.Config, a.RuinCmd)
	a.Gui.QuickCapture = a.QuickCapture
	a.Gui.QuickLink = a.QuickLink
	a.Gui.QuickLinkURL = a.QuickLinkURL
	a.Gui.OpenRef = a.OpenRef
	a.Gui.VaultSource = a.VaultSource
	if needsInit {
		a.Gui.SetNeedsInit()
	}
	if versionWarning != "" {
		a.Gui.SetStartupWarning(versionWarning)
	}

	// Detect upgrade migrations. needsInit short-circuits this — a
	// vault that hasn't been initialized has nothing to re-index, and
	// the init prompt will fire instead. DebugBindings also short-
	// circuits: --debug-bindings is a read-only inspection mode and
	// must not write state.json as a side effect.
	if !needsInit && !a.DebugBindings {
		a.attachMigrationsHelper()
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

// attachMigrationsHelper computes the pending migration list for the
// current launch and either attaches a helper for the GUI to drive on
// first layout, or records the current versions silently when nothing
// is pending. Failures (state.json malformed, ruin --version errored)
// fall back to "no migration" so the user can always boot lazyruin.
func (a *App) attachMigrationsHelper() {
	store := migrations.NewStore()
	if err := store.Load(); err != nil {
		// Malformed state.json — treat as "no migrations" rather than
		// blocking startup. The user can fix or rm the file.
		a.Gui.SetStartupWarning(fmt.Sprintf("ignoring %s — %s", "lazyruin/state.json", err))
		return
	}

	ruinVer, _ := a.RuinCmd.Version() // empty on failure; Detect handles it
	curr := migrations.VersionPair{Lazyruin: a.LazyruinVersion, Ruin: ruinVer}
	prevEntry, hadEntry := store.VaultEntry(a.RuinCmd.VaultPath())
	prev := migrations.VersionPair{Lazyruin: prevEntry.LastLazyruinVersion, Ruin: prevEntry.LastRuinVersion}

	// First launch against this vault on a migration-aware lazyruin: a
	// fresh vault stays "first install" (Detect skips), but a vault
	// that already has notes is an existing user upgrading into the
	// migration system — bootstrap prev with AncientVersion so any
	// currently-applicable registry entry will fire and re-index.
	if !hadEntry {
		empty, err := onboarding.IsVaultEmpty(a.RuinCmd)
		isEmpty := err == nil && empty
		prev = migrations.BootstrapPrev(isEmpty)
	}

	pending := migrations.Detect(curr, prev, prevEntry.AppliedMigrations)

	helper := helperspkg.NewMigrationsHelper(a.RuinCmd, a.Gui, store, a.RuinCmd.VaultPath(), prev, curr, pending)
	if len(pending) == 0 {
		// Nothing to do — record current versions so a future
		// migration's Applies predicate sees an accurate "previous".
		helper.RecordNoPending()
		return
	}
	a.Gui.SetMigrationsHelper(helper)
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
// Returns a short human-readable label describing where the path came from
// (shown in the about dialog). When no configured source yields a path,
// falls back to the current working directory so the TUI's init dialog
// has a concrete proposal — pressing Yes there invokes `ruin init <cwd>`
// to properly set up the directory as a vault.
func resolveVaultPath(cfg *config.Config, cliOverride, ruinBin string) (string, string, error) {
	// 1. Check CLI flag (highest priority)
	if cliOverride != "" {
		return expandPath(cliOverride), "--vault flag", nil
	}

	// 2. Check config
	if cfg.VaultPath != "" {
		return expandPath(cfg.VaultPath), "lazyruin config", nil
	}

	// 3. Check environment
	if envVault := os.Getenv("LAZYRUIN_VAULT"); envVault != "" {
		return expandPath(envVault), "LAZYRUIN_VAULT env", nil
	}

	// 4. Ask ruin CLI for its configured vault path. The CLI exits 0 with
	// empty output when no vault is configured globally, so treat that the
	// same as a non-zero exit.
	cmd := exec.Command(ruinBin, "config", "vault_path")
	output, err := cmd.Output()
	if err == nil {
		if p := strings.TrimSpace(string(output)); p != "" {
			return p, "ruin CLI config", nil
		}
	}

	// 5. No source resolved — fall back to cwd. The init dialog will
	// propose this path and, on Yes, run `ruin init` to set it up.
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		return "", "", errors.New("could not determine vault path - set LAZYRUIN_VAULT or configure vault_path")
	}
	return cwd, "current directory", nil
}
