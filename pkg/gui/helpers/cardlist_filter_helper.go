package helpers

// CardListFilterHelper manages the CardList filter dialog and filter state.
type CardListFilterHelper struct {
	c *HelperCommon
}

func NewCardListFilterHelper(c *HelperCommon) *CardListFilterHelper {
	return &CardListFilterHelper{c: c}
}

func (self *CardListFilterHelper) OpenFilterDialog() error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList
	if len(cl.Cards) == 0 && !cl.FilterActive() {
		return nil
	}

	title := "Filter " + cl.Title()
	seed := cl.FilterText
	self.c.Helpers().Search().OpenSearchAsFilter(title, seed, cl.Source.Triggers, func(text string) error {
		return self.ApplyFilter(text)
	})
	return nil
}

func (self *CardListFilterHelper) ApplyFilter(filterText string) error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList

	if filterText == "" {
		return self.ClearFilter()
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

func (self *CardListFilterHelper) ClearFilter() error {
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
