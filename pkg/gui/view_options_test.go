package gui

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/config"
)

func TestToggleHideDone(t *testing.T) {
	// Redirect the config file to a tempdir so cfg.Save() from the toggle
	// doesn't clobber the developer's real ~/.config/lazyruin/config.yml.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	ds := tg.gui.contexts.ActivePreview().DisplayState()
	if ds.HideDone {
		t.Fatal("HideDone should default to false")
	}

	tg.gui.helpers.Preview().ToggleHideDone()
	if !ds.HideDone {
		t.Error("HideDone should be true after toggle")
	}
	if !tg.gui.config.ViewOptions.HideDone {
		t.Error("config.ViewOptions.HideDone should be true after toggle (for persistence)")
	}

	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load after toggle: %v", err)
	}
	if !reloaded.ViewOptions.HideDone {
		t.Error("reloaded config.ViewOptions.HideDone should be true after toggle")
	}

	tg.gui.helpers.Preview().ToggleHideDone()
	if ds.HideDone {
		t.Error("HideDone should be false after second toggle")
	}
	if tg.gui.config.ViewOptions.HideDone {
		t.Error("config.ViewOptions.HideDone should be false after second toggle")
	}
}

// TestToggleHideDone_SyncsAcrossContexts guards against the drift where a
// toggle on the active preview left the other preview contexts stale until
// the next app launch. Because HideDone is persisted and seeded at boot
// across every preview context, runtime toggles must keep the four in sync.
func TestToggleHideDone_SyncsAcrossContexts(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Preview().ToggleHideDone()

	for name, ds := range map[string]*struct{ HideDone bool }{
		"CardList":    {HideDone: tg.gui.contexts.CardList.DisplayState().HideDone},
		"PickResults": {HideDone: tg.gui.contexts.PickResults.DisplayState().HideDone},
		"Compose":     {HideDone: tg.gui.contexts.Compose.DisplayState().HideDone},
		"DatePreview": {HideDone: tg.gui.contexts.DatePreview.DisplayState().HideDone},
	} {
		if !ds.HideDone {
			t.Errorf("%s context should have HideDone=true after toggle", name)
		}
	}

	tg.gui.helpers.Preview().ToggleHideDone()

	for name, ds := range map[string]*struct{ HideDone bool }{
		"CardList":    {HideDone: tg.gui.contexts.CardList.DisplayState().HideDone},
		"PickResults": {HideDone: tg.gui.contexts.PickResults.DisplayState().HideDone},
		"Compose":     {HideDone: tg.gui.contexts.Compose.DisplayState().HideDone},
		"DatePreview": {HideDone: tg.gui.contexts.DatePreview.DisplayState().HideDone},
	} {
		if ds.HideDone {
			t.Errorf("%s context should have HideDone=false after second toggle", name)
		}
	}
}
