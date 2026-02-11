package gui

import (
	"testing"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/config"
	"kvnd/lazyruin/pkg/testutil"

	"github.com/jesseduffield/gocui"
)

// testGui provides a headless gocui instance for integration testing.
// It creates real views via layout() and wires up a MockExecutor,
// allowing handler functions to be called directly and state to be asserted.
type testGui struct {
	gui *Gui
	g   *gocui.Gui
	t   *testing.T
}

// newTestGui creates a headless GUI with mock data.
// Views are created and data is loaded via the mock executor.
func newTestGui(t *testing.T, mock *testutil.MockExecutor) *testGui {
	t.Helper()

	ruin := commands.NewRuinCommandWithExecutor(mock, mock.VaultPath())
	cfg := &config.Config{}
	gui := NewGui(cfg, ruin)

	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputNormal,
		Headless:   true,
		Width:      120,
		Height:     40,
	})
	if err != nil {
		t.Fatalf("failed to create headless gocui: %v", err)
	}

	gui.g = g
	gui.views = &Views{}
	gui.stopBg = make(chan struct{})
	g.Mouse = true

	// Register the layout manager so views can be created
	g.SetManager(gocui.ManagerFunc(gui.layout))

	// Setup keybindings
	if err := gui.setupKeybindings(); err != nil {
		g.Close()
		t.Fatalf("failed to setup keybindings: %v", err)
	}

	// Trigger layout to create views and load initial data.
	if err := g.ForceLayoutAndRedraw(); err != nil {
		g.Close()
		t.Fatalf("layout failed: %v", err)
	}

	return &testGui{gui: gui, g: g, t: t}
}

// Close cleans up the headless GUI resources.
func (tg *testGui) Close() {
	close(tg.gui.stopBg)
	tg.g.Close()
}
