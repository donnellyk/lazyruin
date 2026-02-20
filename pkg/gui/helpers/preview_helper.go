package helpers

import (
	"strings"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/gui/context"
	"kvnd/lazyruin/pkg/gui/types"
	"kvnd/lazyruin/pkg/models"

	"github.com/jesseduffield/gocui"
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

func (self *PreviewHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// CurrentPreviewCard returns the currently selected card, or nil if none.
// Returns nil when in pickResults or compose mode since card mutations
// don't apply to those modes.
func (self *PreviewHelper) CurrentPreviewCard() *models.Note {
	contexts := self.c.GuiCommon().Contexts()
	if contexts.ActivePreviewKey == "pickResults" || contexts.ActivePreviewKey == "compose" {
		return nil
	}
	cl := contexts.CardList
	idx := cl.SelectedCardIdx
	if idx >= len(cl.Cards) {
		return nil
	}
	return &cl.Cards[idx]
}

// UpdatePreviewForNotes updates the preview pane to show the selected note.
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
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.ShowCardList(" "+note.Title+" ", []models.Note{note})
}

// UpdatePreviewCardList loads a card list into the preview.
func (self *PreviewHelper) UpdatePreviewCardList(title string, loadFn func() ([]models.Note, error)) {
	notes, err := loadFn()
	if err != nil {
		return
	}
	self.c.Helpers().PreviewNav().PushNavHistory()
	self.ShowCardList(title, notes)
}

// ShowCardList sets the preview to card-list mode with the given cards and title,
// then renders. Does NOT push nav history or change context focus.
func (self *PreviewHelper) ShowCardList(title string, cards []models.Note) {
	contexts := self.c.GuiCommon().Contexts()
	cl := contexts.CardList
	cl.Cards = cards
	cl.SelectedCardIdx = 0
	ns := cl.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "cardList"
	if v := self.view(); v != nil {
		v.Title = title
	}
	self.c.GuiCommon().RenderPreview()
}

// ShowPickResults sets the preview to pick-results mode with the given results
// and title, then renders. Does NOT push nav history or change context focus.
func (self *PreviewHelper) ShowPickResults(title string, results []models.PickResult) {
	contexts := self.c.GuiCommon().Contexts()
	pr := contexts.PickResults
	pr.Results = results
	pr.SelectedCardIdx = 0
	ns := pr.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "pickResults"
	if v := self.view(); v != nil {
		v.Title = title
	}
	self.c.GuiCommon().RenderPreview()
}

// ShowCompose sets the preview to compose mode with the given note and title,
// then renders. Does NOT push nav history or change context focus.
func (self *PreviewHelper) ShowCompose(title string, note models.Note, sourceMap []models.SourceMapEntry, parentUUID, parentTitle string) {
	contexts := self.c.GuiCommon().Contexts()
	comp := contexts.Compose
	comp.Note = note
	comp.SourceMap = sourceMap
	comp.ParentUUID = parentUUID
	comp.ParentTitle = parentTitle
	comp.SelectedCardIdx = 0
	ns := comp.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	contexts.ActivePreviewKey = "compose"
	if v := self.view(); v != nil {
		v.Title = title
	}
	self.c.GuiCommon().RenderPreview()
}

// --- content reload ---

// ReloadActivePreview dispatches to the appropriate reload method based on
// the current ActivePreviewKey.
func (self *PreviewHelper) ReloadActivePreview() {
	switch self.c.GuiCommon().Contexts().ActivePreviewKey {
	case "pickResults":
		self.ReloadPickResults()
	case "compose":
		gui := self.c.GuiCommon()
		comp := gui.Contexts().Compose
		if comp.ParentUUID != "" {
			composed, sm, err := self.c.RuinCmd().Parent.ComposeFlat(comp.ParentUUID, comp.ParentTitle)
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
}

// ReloadPickResults re-runs the pick query and refreshes the results,
// preserving the selected card index.
func (self *PreviewHelper) ReloadPickResults() {
	gui := self.c.GuiCommon()
	pr := gui.Contexts().PickResults
	pickCtx := gui.Contexts().Pick

	tags, filter := ParsePickQuery(pickCtx.Query)
	results, err := self.c.RuinCmd().Pick.Pick(tags, pickCtx.AnyMode, filter)
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
				composed, _, err := self.c.RuinCmd().Parent.ComposeFlat(parent.UUID, parent.Title)
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
	self.ReloadContent()
	return nil
}

// ToggleGlobalTags toggles global tags display.
func (self *PreviewHelper) ToggleGlobalTags() error {
	ds := self.activeCtx().DisplayState()
	ds.ShowGlobalTags = !ds.ShowGlobalTags
	self.ReloadContent()
	return nil
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

	self.c.GuiCommon().ShowMenuDialog("View Options", []types.MenuItem{
		{Label: fmLabel, Key: "f", OnRun: func() error { return self.ToggleFrontmatter() }},
		{Label: titleLabel, Key: "t", OnRun: func() error { return self.ToggleTitle() }},
		{Label: tagsLabel, Key: "T", OnRun: func() error { return self.ToggleGlobalTags() }},
		{Label: mdLabel, Key: "M", OnRun: func() error { return self.ToggleMarkdown() }},
	})
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
