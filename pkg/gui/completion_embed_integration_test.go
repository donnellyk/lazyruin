package gui

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

func TestEmbedCandidates(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	t.Run("empty filter returns type prefixes then titles", func(t *testing.T) {
		items := tg.gui.embedCandidates("")
		if len(items) < 4 {
			t.Fatalf("expected at least 4 items (4 type prefixes), got %d", len(items))
		}
		// First four must be type prefixes in the declared order
		expected := []string{"search:", "pick:", "query:", "compose:"}
		for i, want := range expected {
			if items[i].Label != want {
				t.Errorf("items[%d].Label = %q, want %q", i, items[i].Label, want)
			}
			if !items[i].ContinueCompleting {
				t.Errorf("items[%d] (%q) should have ContinueCompleting=true", i, items[i].Label)
			}
		}
	})

	t.Run("filter 'search' narrows to search prefix", func(t *testing.T) {
		items := tg.gui.embedCandidates("search")
		foundType := false
		for _, item := range items {
			if item.Label == "search:" {
				foundType = true
			}
		}
		if !foundType {
			t.Errorf("expected search: prefix in filtered results, got %v", labels(items))
		}
	})
}

func TestSavedQueryCandidates(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	items := tg.gui.savedQueryCandidates("")
	if len(items) == 0 {
		t.Fatalf("expected saved queries from default mock, got empty")
	}
	// defaultMock has daily-notes and work-items
	want := map[string]bool{"daily-notes": true, "work-items": true}
	for _, item := range items {
		delete(want, item.Label)
	}
	if len(want) > 0 {
		t.Errorf("missing saved queries: %v", want)
	}
}

func TestComposeRefCandidates(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	items := tg.gui.composeRefCandidates("")
	if len(items) == 0 {
		t.Fatalf("expected bookmarks + titles from default mock, got empty")
	}
	// defaultMock has a "journal" parent bookmark
	foundBookmark := false
	for _, item := range items {
		if item.Label == "journal" && item.Detail != "" {
			foundBookmark = true
		}
	}
	if !foundBookmark {
		t.Errorf("expected 'journal' bookmark in compose ref candidates, got %v", labels(items))
	}
}

func TestTagCompletion_AfterNegation(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	// Simulate typing `!#f` outside any embed. The `#` trigger's scan-backward
	// should win, and the filter should be "f".
	content := "!#f"
	triggers := tg.gui.captureTriggers()
	trigger, filter, _ := detectTrigger(content, len(content), triggers)
	if trigger == nil {
		t.Fatal("expected a trigger match for `!#f`")
	}
	if trigger.Prefix != "#" {
		t.Errorf("prefix = %q, want `#`", trigger.Prefix)
	}
	if filter != "f" {
		t.Errorf("filter = %q, want `f`", filter)
	}

	// Inside an embed, `!#d` should still fire the `#` trigger.
	embedContent := "![[search: !#d"
	trigger2, filter2, _ := detectTrigger(embedContent, len(embedContent), triggers)
	if trigger2 == nil {
		t.Fatal("expected a trigger match inside embed for `!#d`")
	}
	if trigger2.Prefix != "#" {
		t.Errorf("inside embed: prefix = %q, want `#`", trigger2.Prefix)
	}
	if filter2 != "d" {
		t.Errorf("inside embed: filter = %q, want `d`", filter2)
	}
}

// labels extracts Label fields from completion items for readable error output.
func labels(items []types.CompletionItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

func TestEmbedOptionFreeFormValue_FallsThroughToTagTrigger(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	if v == nil {
		t.Fatal("Capture view should exist")
	}
	state := tg.gui.contexts.Capture.Completion

	// Simulate the reported scenario: user is inside a pick embed's filter=
	// option and types a `#` followed by a partial tag.
	v.TextArea.TypeString("![[pick: #followup | filter=#d")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)

	if !state.Active {
		t.Fatalf("expected tag completion active after filter=#d, items: %v", labels(state.Items))
	}
	foundDaily := false
	for _, item := range state.Items {
		if item.Label == "#daily" {
			foundDaily = true
			break
		}
	}
	if !foundDaily {
		t.Errorf("expected `#daily` in tag candidates, got %v", labels(state.Items))
	}
}

func TestEmbedOptionFreeFormValue_FallsThroughToDateTrigger(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	state := tg.gui.contexts.Capture.Completion

	// `@` inside filter= should fire the date trigger (via the date-prefix
	// fallback scan, which also uses isTriggerBoundary).
	v.TextArea.TypeString("![[pick: #followup | filter=@to")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)

	if !state.Active {
		t.Fatalf("expected date completion active after filter=@to, items: %v", labels(state.Items))
	}
	foundToday := false
	for _, item := range state.Items {
		if item.Label == "@today" {
			foundToday = true
			break
		}
	}
	if !foundToday {
		t.Errorf("expected `@today` in date candidates, got %v", labels(state.Items))
	}
}

func TestEmbedOption_MultipleOptions_LastSegmentCompletes(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	state := tg.gui.contexts.Capture.Completion

	// Multi-option case: earlier option (sort=created:desc) is complete,
	// cursor is in the free-form filter= value with a tag partial.
	v.TextArea.TypeString("![[pick: #followup | sort=created:desc, filter=#d")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)

	if !state.Active {
		t.Fatalf("expected tag completion active in last option, items: %v", labels(state.Items))
	}
	foundDaily := false
	for _, item := range state.Items {
		if item.Label == "#daily" {
			foundDaily = true
			break
		}
	}
	if !foundDaily {
		t.Errorf("expected `#daily` candidate after multi-option filter=#d, got %v", labels(state.Items))
	}
}

func TestEmbedOption_KnownKeyZeroMatches_Dismisses(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	state := tg.gui.contexts.Capture.Completion

	// `format=` is a known key, but `zzz` matches no valid format value.
	// This must NOT fall through to the `![[` / note-title scan — it must
	// dismiss cleanly so the user isn't shown unrelated candidates.
	v.TextArea.TypeString("![[search: #daily | format=zzz")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)

	if state.Active {
		t.Errorf("expected dismissal for known-key zero-match, got active with items: %v", labels(state.Items))
	}
}

func TestEmbedTypePrefix_PreservesBrackets(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	if v == nil {
		t.Fatal("Capture view should exist")
	}
	state := tg.gui.contexts.Capture.Completion

	v.TextArea.TypeString("![[")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)
	if !state.Active || len(state.Items) == 0 {
		t.Fatal("expected completion active with items after typing `![[`")
	}

	// Find the `search:` prefix and select it
	for i, item := range state.Items {
		if item.Label == "search:" {
			state.SelectedIndex = i
			break
		}
	}

	tg.gui.acceptCompletion(v, state, tg.gui.captureTriggers())
	got := v.TextArea.GetUnwrappedContent()
	want := "![[search: "
	if got != want {
		t.Errorf("after accepting `search:` prefix, buffer = %q, want %q", got, want)
	}
}

func TestCaptureTypeaheadEmbedFlow(t *testing.T) {
	tg := newTestGui(t, defaultMock())
	defer tg.Close()

	tg.gui.helpers.Capture().OpenCapture()
	if err := tg.g.ForceLayoutAndRedraw(); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	v := tg.gui.views.Capture
	if v == nil {
		t.Fatal("Capture view should exist after OpenCapture + layout")
	}
	state := tg.gui.contexts.Capture.Completion

	// Phase 1: type `![[`, expect type prefix candidates
	v.TextArea.TypeString("![[")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)
	if !state.Active {
		t.Fatal("expected completion active after typing `![[`")
	}
	// Type prefixes should appear first
	if len(state.Items) < 4 {
		t.Fatalf("expected at least 4 items (type prefixes), got %d", len(state.Items))
	}
	if state.Items[0].Label != "search:" {
		t.Errorf("items[0].Label = %q, want `search:`", state.Items[0].Label)
	}

	// Phase 2: after accepting `search: `, typing `#` fires the tag trigger.
	// Simulate the accepted state by typing `search: ` directly.
	v.TextArea.TypeString("search: #")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)
	if !state.Active {
		t.Fatalf("expected completion active after typing `![[search: #`, items: %v", labels(state.Items))
	}
	// The candidates should come from the `#` tag trigger — expect at
	// least one tag from the default mock (daily, work, project).
	foundTag := false
	for _, item := range state.Items {
		if item.Label == "#daily" || item.Label == "#work" || item.Label == "#project" {
			foundTag = true
			break
		}
	}
	if !foundTag {
		t.Errorf("expected tag candidates from `#` trigger, got %v", labels(state.Items))
	}

	// Phase 3: type options `|`, then assert we get option-key completion
	v.TextArea.TypeString("daily | ")
	tg.gui.updateCompletion(v, tg.gui.captureTriggers(), state)
	if !state.Active {
		t.Fatalf("expected completion active in options phase")
	}
	foundFormat := false
	for _, item := range state.Items {
		if item.Label == "format=" {
			foundFormat = true
			break
		}
	}
	if !foundFormat {
		t.Errorf("expected option key candidates including `format=`, got %v", labels(state.Items))
	}
}
