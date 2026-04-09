package context

import (
	"testing"

	"github.com/donnellyk/lazyruin/pkg/gui/types"
)

func TestInFilterMode(t *testing.T) {
	sc := NewSearchContext()
	if sc.InFilterMode() {
		t.Fatal("expected InFilterMode false on new context")
	}
	sc.OnFilterSubmit = func(string) error { return nil }
	if !sc.InFilterMode() {
		t.Fatal("expected InFilterMode true when OnFilterSubmit is set")
	}
}

func TestClearFilterMode(t *testing.T) {
	sc := NewSearchContext()
	sc.FilterTitle = "Filter Test"
	sc.FilterSeed = "#bug"
	sc.FilterSeedDone = true
	sc.FilterTriggers = func() []types.CompletionTrigger { return nil }
	sc.OnFilterSubmit = func(string) error { return nil }

	sc.ClearFilterMode()

	if sc.FilterTitle != "" {
		t.Fatalf("expected FilterTitle empty, got %q", sc.FilterTitle)
	}
	if sc.FilterSeed != "" {
		t.Fatalf("expected FilterSeed empty, got %q", sc.FilterSeed)
	}
	if sc.FilterSeedDone {
		t.Fatal("expected FilterSeedDone false")
	}
	if sc.FilterTriggers != nil {
		t.Fatal("expected FilterTriggers nil")
	}
	if sc.OnFilterSubmit != nil {
		t.Fatal("expected OnFilterSubmit nil")
	}
}
