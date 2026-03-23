package helpers

import "kvnd/lazyruin/pkg/gui/context"

// CardListFilterHelper manages the filter dialog and filter state for both
// CardList and PickResults preview modes via the Filterable interface.
type CardListFilterHelper struct {
	c *HelperCommon
}

func NewCardListFilterHelper(c *HelperCommon) *CardListFilterHelper {
	return &CardListFilterHelper{c: c}
}

func (self *CardListFilterHelper) filterable() context.Filterable {
	return self.c.GuiCommon().Contexts().ActiveFilterable()
}

func (self *CardListFilterHelper) OpenFilterDialog() error {
	return self.openFilter(self.filterable())
}

func (self *CardListFilterHelper) ApplyFilter(filterText string) error {
	return self.applyFilter(self.filterable(), filterText)
}

func (self *CardListFilterHelper) ClearFilter() error {
	return self.clearFilter(self.filterable())
}

func (self *CardListFilterHelper) FilterActive() bool {
	return self.filterable().FilterActive()
}

func (self *CardListFilterHelper) openFilter(f context.Filterable) error {
	if f.ItemCount() == 0 && !f.FilterActive() {
		return nil
	}

	title := "Filter " + f.Title()
	seed := f.GetFilterText()
	self.c.Helpers().Search().OpenSearchAsFilter(title, seed, f.FilterTriggers(), func(text string) error {
		return self.applyFilter(f, text)
	})
	return nil
}

func (self *CardListFilterHelper) applyFilter(f context.Filterable, filterText string) error {
	gui := self.c.GuiCommon()

	if filterText == "" {
		return self.clearFilter(f)
	}

	if !f.HasRequery() {
		return nil
	}

	// Capture the pre-filter count before requery replaces the items.
	prevCount := f.ItemCount()
	wasActive := f.FilterActive()

	if err := f.RequeryAndApply(filterText); err != nil {
		gui.ShowError(err)
		return nil
	}

	if !wasActive {
		f.SetUnfilteredCount(prevCount)
	}
	f.SetFilterText(filterText)
	f.ResetSelectedCard()
	gui.RenderPreview()
	return nil
}

func (self *CardListFilterHelper) clearFilter(f context.Filterable) error {
	gui := self.c.GuiCommon()

	if !f.FilterActive() {
		return nil
	}

	if !f.HasRequery() {
		return nil
	}

	if err := f.RequeryAndApply(""); err != nil {
		gui.ShowError(err)
		return nil
	}

	f.ClearFilter()
	f.ResetSelectedCard()
	gui.RenderPreview()
	return nil
}
