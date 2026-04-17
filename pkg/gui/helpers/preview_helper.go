package helpers

import (
	"strings"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/gui/types"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// PreviewHelper handles core preview operations: accessors, content display,
// content reload, and display toggles.
type PreviewHelper struct {
	c *HelperCommon
}

// NewPreviewHelper creates a new PreviewHelper.
func NewPreviewHelper(c *HelperCommon) *PreviewHelper {
	return &PreviewHelper{c: c}
}

func (self *PreviewHelper) activeCtx() context.IPreviewContext {
	return self.c.GuiCommon().Contexts().ActivePreview()
}

func (self *PreviewHelper) cardList() *context.CardListContext {
	return self.c.GuiCommon().Contexts().CardList
}

// BuildSearchOptions returns SearchOptions based on current preview toggle state.
func (self *PreviewHelper) BuildSearchOptions() commands.SearchOptions {
	ds := self.activeCtx().DisplayState()
	return commands.SearchOptions{
		IncludeContent:  true,
		StripGlobalTags: !ds.ShowGlobalTags,
		StripTitle:      !ds.ShowTitle,
	}
}

// NewSearchSource builds a CardListSource that re-queries via ruin search.
// The baseQuery is prepended to any filter text. Sort is applied if non-empty.
func (self *PreviewHelper) NewSearchSource(baseQuery, sort string) context.CardListSource {
	return context.CardListSource{
		Query: baseQuery,
		Requery: func(filterText string) ([]models.Note, error) {
			combined := strings.TrimSpace(baseQuery + " " + filterText)
			o := self.BuildSearchOptions()
			o.Sort = sort
			return self.c.RuinCmd().Search.Search(combined, o)
		},
	}
}

// NewSearchSourceWithExtractSort builds a CardListSource that re-queries via
// ruin search, extracting a sort: token from rawQuery on each re-query. This
// is used for user-typed search strings that may contain "sort:value".
func (self *PreviewHelper) NewSearchSourceWithExtractSort(rawQuery string) context.CardListSource {
	return context.CardListSource{
		Query: rawQuery,
		Requery: func(filterText string) ([]models.Note, error) {
			q, s := ExtractSort(rawQuery)
			combined := strings.TrimSpace(q + " " + filterText)
			o := self.BuildSearchOptions()
			o.Sort = s
			return self.c.RuinCmd().Search.Search(combined, o)
		},
	}
}

// CurrentPreviewCard returns the currently selected card, or nil if none.
// Returns nil in pickResults mode since those are transient results.
func (self *PreviewHelper) CurrentPreviewCard() *models.Note {
	contexts := self.c.GuiCommon().Contexts()
	switch contexts.ActivePreviewKey {
	case "compose":
		return &contexts.Compose.Note
	case "pickResults":
		return nil
	case "datePreview":
		dp := contexts.DatePreview
		idx := dp.SelectedCardIdx
		section := dp.SectionForCard(idx)
		localIdx := dp.LocalCardIdx(idx)
		switch section {
		case context.SectionTagPicks:
			if localIdx < len(dp.TagPicks) {
				r := dp.TagPicks[localIdx]
				return &models.Note{UUID: r.UUID, Path: r.File, Title: r.Title}
			}
		case context.SectionTodoPicks:
			if localIdx < len(dp.TodoPicks) {
				r := dp.TodoPicks[localIdx]
				return &models.Note{UUID: r.UUID, Path: r.File, Title: r.Title}
			}
		case context.SectionNotes:
			if localIdx < len(dp.Notes) {
				return &dp.Notes[localIdx]
			}
		}
		return nil
	default:
		cl := contexts.CardList
		idx := cl.SelectedCardIdx
		if idx >= len(cl.Cards) {
			return nil
		}
		return &cl.Cards[idx]
	}
}

// UpdatePreviewForNotes updates the preview pane to show the selected note
// as a hover preview — the view is not committed to history and the title
// is italicized.
func (self *PreviewHelper) UpdatePreviewForNotes() {
	notes := self.c.GuiCommon().Contexts().Notes
	if len(notes.Items) == 0 {
		return
	}
	idx := notes.GetSelectedLineIdx()
	if idx >= len(notes.Items) {
		return
	}
	note := notes.Items[idx]
	_ = self.c.Helpers().Navigator().ShowHover("cardList", note.Title, func() error {
		self.ShowCardList(note.Title, []models.Note{note}, self.NewSingleNoteSource(note.UUID))
		return nil
	})
}

// UpdatePreviewCardList loads a card list into the preview as a hover
// preview. Does not record a history entry.
func (self *PreviewHelper) UpdatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	_ = self.c.Helpers().Navigator().ShowHover("cardList", title, func() error {
		notes, err := loadFn()
		if err != nil {
			return err
		}
		self.ShowCardList(title, notes)
		return nil
	})
}

// NewSingleNoteSource builds a CardListSource that re-queries a single note
// by UUID — used when a single note is shown in card-list view (hover,
// wiki-link follow, etc.) so re-query on history restore stays in sync
// with edits made to the underlying file.
func (self *PreviewHelper) NewSingleNoteSource(uuid string) context.CardListSource {
	return context.CardListSource{
		Query: uuid,
		Requery: func(_ string) ([]models.Note, error) {
			o := self.BuildSearchOptions()
			note, err := self.c.RuinCmd().Search.Get(uuid, o)
			if err != nil || note == nil {
				return nil, err
			}
			return []models.Note{*note}, nil
		},
	}
}

// ShowCardList sets the preview to card-list mode with the given cards and title,
// then renders. Does NOT push nav history or change context focus.
// The optional source enables filtering; pass a zero-value source to disable.
func (self *PreviewHelper) ShowCardList(title string, cards []models.Note, source ...context.CardListSource) {
	contexts := self.c.GuiCommon().Contexts()
	cl := contexts.CardList
	cl.Cards = cards
	cl.SelectedCardIdx = 0
	cl.SetTitle(title)
	cl.ClearFilter()
	if len(source) > 0 {
		cl.Source = source[0]
	} else {
		cl.Source = context.CardListSource{}
	}
	cl.DisplayState().ShowCompose = false
	cl.ComposedNote = nil
	cl.ComposedSourceMap = nil
	ns := cl.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "cardList"
	self.c.Helpers().TitleCache().PutNotes(cards)
	self.c.Helpers().TitleCache().ResolveUnknownParents(cards)
	self.c.GuiCommon().RenderPreview()
}

// ShowPickResults sets the preview to pick-results mode with the given results
// and title, then renders. Does NOT push nav history or change context focus.
// The optional source enables filtering; pass a zero-value source to disable.
func (self *PreviewHelper) ShowPickResults(title string, results []models.PickResult, source ...context.PickResultsSource) {
	contexts := self.c.GuiCommon().Contexts()
	pr := contexts.PickResults
	pr.Results = results
	pr.SelectedCardIdx = 0
	pr.SetTitle(title)
	pr.ClearFilter()
	if len(source) > 0 {
		pr.Source = source[0]
	} else {
		pr.Source = context.PickResultsSource{}
	}
	ns := pr.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "pickResults"
	self.c.GuiCommon().RenderPreview()
}

// ShowCompose sets the preview to compose mode with the given note and title,
// then renders. Does NOT push nav history or change context focus.
func (self *PreviewHelper) ShowCompose(title string, note models.Note, sourceMap []models.SourceMapEntry, parent models.ParentBookmark) {
	contexts := self.c.GuiCommon().Contexts()
	comp := contexts.Compose
	comp.Note = note
	comp.SourceMap = sourceMap
	comp.Parent = parent
	comp.SetTitle(title)
	comp.SelectedCardIdx = 0
	ns := comp.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "compose"
	self.c.GuiCommon().RenderPreview()
}

// --- content reload ---

// cursorIdentity captures the source-line identity at the current cursor
// position so it can be restored after a reload changes the line array.
type cursorIdentity struct {
	uuid    string
	lineNum int
	valid   bool
}

func (self *PreviewHelper) saveCursorIdentity() cursorIdentity {
	ns := self.activeCtx().NavState()
	if ns.CursorLine >= 0 && ns.CursorLine < len(ns.Lines) {
		sl := ns.Lines[ns.CursorLine]
		if sl.UUID != "" && sl.LineNum > 0 {
			return cursorIdentity{uuid: sl.UUID, lineNum: sl.LineNum, valid: true}
		}
	}
	return cursorIdentity{}
}

func (self *PreviewHelper) restoreCursorIdentity(id cursorIdentity) {
	if !id.valid {
		return
	}
	ns := self.activeCtx().NavState()
	for i, sl := range ns.Lines {
		if sl.UUID == id.uuid && sl.LineNum == id.lineNum {
			ns.CursorLine = i
			self.c.GuiCommon().RenderPreview()
			return
		}
	}
	// Identity not found (line was deleted); leave cursor where reload placed it.
}

// ReloadActivePreview dispatches to the appropriate reload method based on
// the current ActivePreviewKey, preserving the cursor's source-line identity.
func (self *PreviewHelper) ReloadActivePreview() {
	saved := self.saveCursorIdentity()

	switch self.c.GuiCommon().Contexts().ActivePreviewKey {
	case "pickResults":
		self.ReloadPickResults()
	case "datePreview":
		self.c.Helpers().DatePreview().ReloadDatePreview()
	case "compose":
		gui := self.c.GuiCommon()
		comp := gui.Contexts().Compose
		if comp.Parent.Name != "" {
			composed, sm, err := self.c.RuinCmd().Parent.Compose(comp.Parent)
			if err == nil {
				comp.Note = composed
				comp.SourceMap = sm
			}
		}
		self.c.Helpers().Notes().FetchNotesForCurrentTab(true)
		gui.RenderPreview()
	default:
		self.ReloadContent()
	}

	self.restoreCursorIdentity(saved)
}

// ReloadPickResults re-runs the pick query and refreshes the results,
// preserving the selected card index.
func (self *PreviewHelper) ReloadPickResults() {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults
	pickCtx := gui.Contexts().Pick

	tags, date, filter, flags := ParsePickQuery(pickCtx.Query)
	results, err := self.c.RuinCmd().Pick.Pick(tags, commands.PickOpts{
		Any:  pickCtx.AnyMode || flags.Any,
		Todo: pickCtx.TodoMode || flags.Todo,
		Date: date, Filter: filter,
	})
	if err != nil {
		results = nil
	}

	savedIdx := pr.SelectedCardIdx
	pr.Results = results
	if savedIdx >= len(pr.Results) {
		if len(pr.Results) > 0 {
			pr.SelectedCardIdx = len(pr.Results) - 1
		} else {
			pr.SelectedCardIdx = 0
		}
	}

	// Re-render first to rebuild CardLineRanges, then clamp cursor to the
	// selected card's content range so it doesn't point past the new layout.
	self.c.Helpers().Notes().FetchNotesForCurrentTab(true)
	gui.RenderPreview()

	ns := pr.NavState()
	idx := pr.SelectedCardIdx
	if idx < len(ns.CardLineRanges) {
		r := ns.CardLineRanges[idx]
		if ns.CursorLine < r[0]+1 || ns.CursorLine >= r[1]-1 {
			ns.CursorLine = r[0] + 1
			gui.RenderPreview()
		}
	}
}

// ReloadContent reloads notes and preview cards with current toggle settings.
func (self *PreviewHelper) ReloadContent() {
	gui := self.c.GuiCommon()
	self.c.Helpers().Notes().FetchNotesForCurrentTab(true)

	cl := self.cardList()
	if len(cl.Cards) > 0 {
		savedCardIdx := cl.SelectedCardIdx
		self.reloadPreviewCards()
		if savedCardIdx < len(cl.Cards) {
			cl.SelectedCardIdx = savedCardIdx
		}
	}
	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCards() {
	gui := self.c.GuiCommon()
	cl := self.cardList()
	cl.TemporarilyMoved = nil
	opts := self.BuildSearchOptions()

	searchQuery := gui.Contexts().Search.Query
	if searchQuery != "" {
		notes, err := self.c.RuinCmd().Search.Search(searchQuery, opts)
		if err == nil {
			cl.Cards = notes
		}
		gui.RenderPreview()
		return
	}

	switch gui.PreviousContextKey() {
	case "notes":
		self.reloadPreviewCardsFromNotes()
	case "tags":
		tagsCtx := gui.Contexts().Tags
		if len(tagsCtx.Items) > 0 {
			tag := tagsCtx.Items[tagsCtx.GetSelectedLineIdx()]
			notes, err := self.c.RuinCmd().Search.Search(tag.Name, opts)
			if err == nil {
				cl.Cards = notes
			}
		}
	case "queries":
		queriesCtx := gui.Contexts().Queries
		if queriesCtx.CurrentTab == "parents" {
			if len(queriesCtx.Parents) > 0 {
				parent := queriesCtx.Parents[queriesCtx.ParentsTrait().GetSelectedLineIdx()]
				composed, _, err := self.c.RuinCmd().Parent.Compose(parent)
				if err == nil {
					cl.Cards = []models.Note{composed}
				}
			}
		} else if len(queriesCtx.Queries) > 0 {
			query := queriesCtx.Queries[queriesCtx.QueriesTrait().GetSelectedLineIdx()]
			notes, err := self.c.RuinCmd().Queries.Run(query.Name, opts)
			if err == nil {
				cl.Cards = notes
			}
		}
	default:
		self.reloadPreviewCardsFromNotes()
	}

	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCardsFromNotes() {
	cl := self.cardList()
	opts := self.BuildSearchOptions()
	updated := make([]models.Note, 0, len(cl.Cards))
	for _, card := range cl.Cards {
		fresh, err := self.c.RuinCmd().Search.Get(card.UUID, opts)
		if err == nil && fresh != nil {
			if len(fresh.InlineTags) == 0 && len(card.InlineTags) > 0 {
				fresh.InlineTags = card.InlineTags
			}
			updated = append(updated, *fresh)
		} else {
			card.Content = ""
			updated = append(updated, card)
		}
	}
	cl.Cards = updated
}

// --- display toggles ---

// ToggleMarkdown toggles markdown rendering.
func (self *PreviewHelper) ToggleMarkdown() error {
	ds := self.activeCtx().DisplayState()
	ds.RenderMarkdown = !ds.RenderMarkdown
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleFrontmatter toggles frontmatter display.
func (self *PreviewHelper) ToggleFrontmatter() error {
	ds := self.activeCtx().DisplayState()
	ds.ShowFrontmatter = !ds.ShowFrontmatter
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleTitle toggles title display.
func (self *PreviewHelper) ToggleTitle() error {
	ds := self.activeCtx().DisplayState()
	ds.ShowTitle = !ds.ShowTitle
	self.reloadComposeIfActive()
	self.ReloadContent()
	return nil
}

// ToggleGlobalTags toggles global tags display.
func (self *PreviewHelper) ToggleGlobalTags() error {
	ds := self.activeCtx().DisplayState()
	ds.ShowGlobalTags = !ds.ShowGlobalTags
	self.reloadComposeIfActive()
	self.ReloadContent()
	return nil
}

// ToggleDimDone toggles dimming of lines with #done.
func (self *PreviewHelper) ToggleDimDone() error {
	ds := self.activeCtx().DisplayState()
	ds.DimDone = !ds.DimDone
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleHideDone toggles hiding of #done lines and fully-done sections.
// Persists the new value in config so it survives restarts, and applies
// it to every preview context so switching panels mid-session doesn't
// show stale state.
func (self *PreviewHelper) ToggleHideDone() error {
	next := !self.activeCtx().DisplayState().HideDone
	contexts := self.c.GuiCommon().Contexts()
	for _, ctx := range []context.IPreviewContext{
		contexts.CardList,
		contexts.PickResults,
		contexts.Compose,
		contexts.DatePreview,
	} {
		if ctx == nil {
			continue
		}
		ctx.DisplayState().HideDone = next
	}
	cfg := self.c.Config()
	if cfg != nil {
		cfg.ViewOptions.HideDone = next
		if err := cfg.Save(); err != nil {
			self.c.GuiCommon().ShowError(err)
		}
	}
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleCompose toggles the compose view for the currently selected card.
// When turned on, fetches the composed/embed-expanded version of the note.
// Cleared automatically when navigating to a new note (see ShowCardList).
func (self *PreviewHelper) ToggleCompose() error {
	gui := self.c.GuiCommon()
	cl := gui.Contexts().CardList
	ds := cl.DisplayState()

	if ds.ShowCompose {
		ds.ShowCompose = false
		cl.ComposedNote = nil
		cl.ComposedSourceMap = nil
		gui.RenderPreview()
		return nil
	}

	note := self.CurrentPreviewCard()
	if note == nil || note.UUID == "" {
		return nil
	}

	composed, sourceMap, err := self.c.RuinCmd().Parent.ComposeNote(note.UUID, !ds.ShowTitle, !ds.ShowGlobalTags)
	if err != nil {
		gui.ShowError(err)
		return nil
	}

	cl.ComposedNote = &composed
	cl.ComposedSourceMap = sourceMap
	ds.ShowCompose = true
	gui.RenderPreview()
	return nil
}

// reloadComposeIfActive re-fetches the composed note with current display settings.
// Called when ShowTitle or ShowGlobalTags change while ShowCompose is active.
func (self *PreviewHelper) reloadComposeIfActive() {
	cl := self.c.GuiCommon().Contexts().CardList
	ds := cl.DisplayState()
	if !ds.ShowCompose || cl.ComposedNote == nil {
		return
	}
	composed, sourceMap, err := self.c.RuinCmd().Parent.ComposeNote(cl.ComposedNote.UUID, !ds.ShowTitle, !ds.ShowGlobalTags)
	if err != nil {
		return
	}
	cl.ComposedNote = &composed
	cl.ComposedSourceMap = sourceMap
}

// ViewOptionsDialog shows the view options menu.
func (self *PreviewHelper) ViewOptionsDialog() error {
	ds := self.activeCtx().DisplayState()
	fmLabel := "Show frontmatter"
	if ds.ShowFrontmatter {
		fmLabel = "Hide frontmatter"
	}
	titleLabel := "Show title"
	if ds.ShowTitle {
		titleLabel = "Hide title"
	}
	tagsLabel := "Show global tags"
	if ds.ShowGlobalTags {
		tagsLabel = "Hide global tags"
	}
	mdLabel := "Render markdown"
	if ds.RenderMarkdown {
		mdLabel = "Raw markdown"
	}
	doneLabel := "Dim #done lines"
	if ds.DimDone {
		doneLabel = "Undim #done lines"
	}
	hideLabel := "Hide #done lines"
	if ds.HideDone {
		hideLabel = "Show #done lines"
	}

	items := []types.MenuItem{
		{Label: fmLabel, Key: "f", OnRun: func() error { return self.ToggleFrontmatter() }},
		{Label: titleLabel, Key: "t", OnRun: func() error { return self.ToggleTitle() }},
		{Label: tagsLabel, Key: "T", OnRun: func() error { return self.ToggleGlobalTags() }},
		{Label: mdLabel, Key: "M", OnRun: func() error { return self.ToggleMarkdown() }},
		{Label: doneLabel, Key: "d", OnRun: func() error { return self.ToggleDimDone() }},
		{Label: hideLabel, Key: "h", OnRun: func() error { return self.ToggleHideDone() }},
	}

	if self.c.GuiCommon().Contexts().ActivePreviewKey == "cardList" {
		composeLabel := "View compose"
		if ds.ShowCompose {
			composeLabel = "View raw"
		}
		items = append(items, types.MenuItem{Label: composeLabel, Key: "c", OnRun: func() error { return self.ToggleCompose() }})
	}

	self.c.GuiCommon().ShowMenuDialog("View Options", items)
	return nil
}

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var sb strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
