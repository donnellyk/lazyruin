package helpers

// CardListFilterHelper manages the filter dialog and filter state for both
// CardList and PickResults preview modes.
type CardListFilterHelper struct {
	c *HelperCommon
}

func NewCardListFilterHelper(c *HelperCommon) *CardListFilterHelper {
	return &CardListFilterHelper{c: c}
}

func (self *CardListFilterHelper) OpenFilterDialog() error {
	gui := self.c.GuiCommon()
	switch gui.Contexts().ActivePreviewKey {
	case "pickResults":
		return self.openPickResultsFilter()
	default:
		return self.openCardListFilter()
	}
}

func (self *CardListFilterHelper) ApplyFilter(filterText string) error {
	gui := self.c.GuiCommon()
	switch gui.Contexts().ActivePreviewKey {
	case "pickResults":
		return self.applyPickResultsFilter(filterText)
	default:
		return self.applyCardListFilter(filterText)
	}
}

func (self *CardListFilterHelper) ClearFilter() error {
	gui := self.c.GuiCommon()
	switch gui.Contexts().ActivePreviewKey {
	case "pickResults":
		return self.clearPickResultsFilter()
	default:
		return self.clearCardListFilter()
	}
}

func (self *CardListFilterHelper) FilterActive() bool {
	gui := self.c.GuiCommon()
	switch gui.Contexts().ActivePreviewKey {
	case "pickResults":
		return gui.Contexts().PickResults.FilterActive()
	default:
		return gui.Contexts().CardList.FilterActive()
	}
}

// --- CardList ---

func (self *CardListFilterHelper) openCardListFilter() error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList
	if len(cl.Cards) == 0 && !cl.FilterActive() {
		return nil
	}

	title := "Filter " + cl.Title()
	seed := cl.FilterText
	self.c.Helpers().Search().OpenSearchAsFilter(title, seed, cl.Source.Triggers, func(text string) error {
		return self.applyCardListFilter(text)
	})
	return nil
}

func (self *CardListFilterHelper) applyCardListFilter(filterText string) error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList

	if filterText == "" {
		return self.clearCardListFilter()
	}

	if cl.Source.Requery == nil {
		return nil
	}

	notes, err := cl.Source.Requery(filterText)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	if !cl.FilterActive() {
		cl.UnfilteredCount = len(cl.Cards)
	}
	cl.FilterText = filterText
	cl.Cards = notes
	cl.SelectedCardIdx = 0
	gui.RenderPreview()
	return nil
}

func (self *CardListFilterHelper) clearCardListFilter() error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList

	if !cl.FilterActive() {
		return nil
	}

	if cl.Source.Requery == nil {
		return nil
	}

	notes, err := cl.Source.Requery("")
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	cl.ClearFilter()
	cl.Cards = notes
	cl.SelectedCardIdx = 0
	gui.RenderPreview()
	return nil
}

// --- PickResults ---

func (self *CardListFilterHelper) openPickResultsFilter() error {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults
	if len(pr.Results) == 0 && !pr.FilterActive() {
		return nil
	}

	title := "Filter " + pr.Title()
	seed := pr.FilterText
	self.c.Helpers().Search().OpenSearchAsFilter(title, seed, pr.Source.Triggers, func(text string) error {
		return self.applyPickResultsFilter(text)
	})
	return nil
}

func (self *CardListFilterHelper) applyPickResultsFilter(filterText string) error {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults

	if filterText == "" {
		return self.clearPickResultsFilter()
	}

	if pr.Source.Requery == nil {
		return nil
	}

	results, err := pr.Source.Requery(filterText)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	if !pr.FilterActive() {
		pr.UnfilteredCount = len(pr.Results)
	}
	pr.FilterText = filterText
	pr.Results = results
	pr.SelectedCardIdx = 0
	gui.RenderPreview()
	return nil
}

func (self *CardListFilterHelper) clearPickResultsFilter() error {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults

	if !pr.FilterActive() {
		return nil
	}

	if pr.Source.Requery == nil {
		return nil
	}

	results, err := pr.Source.Requery("")
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	pr.ClearFilter()
	pr.Results = results
	pr.SelectedCardIdx = 0
	gui.RenderPreview()
	return nil
}
