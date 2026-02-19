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

func (self *PreviewHelper) ctx() *context.PreviewContext {
	return self.c.GuiCommon().Contexts().Preview
}

// BuildSearchOptions returns SearchOptions based on current preview toggle state.
func (self *PreviewHelper) BuildSearchOptions() commands.SearchOptions {
	pc := self.ctx()
	return commands.SearchOptions{
		IncludeContent:  true,
		StripGlobalTags: !pc.ShowGlobalTags,
		StripTitle:      !pc.ShowTitle,
	}
}

func (self *PreviewHelper) view() *gocui.View {
	return self.c.GuiCommon().GetView("preview")
}

// CurrentPreviewCard returns the currently selected card, or nil if none.
func (self *PreviewHelper) CurrentPreviewCard() *models.Note {
	pc := self.ctx()
	idx := pc.SelectedCardIndex
	if idx >= len(pc.Cards) {
		return nil
	}
	return &pc.Cards[idx]
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
	pc := self.ctx()
	pc.Mode = context.PreviewModeCardList
	pc.Cards = cards
	pc.SelectedCardIndex = 0
	pc.CursorLine = 1
	pc.ScrollOffset = 0
	if v := self.view(); v != nil {
		v.Title = title
	}
	self.c.GuiCommon().RenderPreview()
}

// ShowPickResults sets the preview to pick-results mode with the given results
// and title, then renders. Does NOT push nav history or change context focus.
func (self *PreviewHelper) ShowPickResults(title string, results []models.PickResult) {
	pc := self.ctx()
	pc.Mode = context.PreviewModePickResults
	pc.PickResults = results
	pc.SelectedCardIndex = 0
	pc.CursorLine = 1
	pc.ScrollOffset = 0
	if v := self.view(); v != nil {
		v.Title = title
	}
	self.c.GuiCommon().RenderPreview()
}

// --- content reload ---

// ReloadContent reloads notes and preview cards with current toggle settings.
func (self *PreviewHelper) ReloadContent() {
	gui := self.c.GuiCommon()
	self.c.Helpers().Notes().FetchNotesForCurrentTab(true)

	pc := self.ctx()
	if len(pc.Cards) > 0 {
		savedCardIdx := pc.SelectedCardIndex
		self.reloadPreviewCards()
		if savedCardIdx < len(pc.Cards) {
			pc.SelectedCardIndex = savedCardIdx
		}
	}
	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCards() {
	gui := self.c.GuiCommon()
	pc := self.ctx()
	pc.TemporarilyMoved = nil
	opts := self.BuildSearchOptions()

	searchQuery := gui.Contexts().Search.Query
	if searchQuery != "" {
		notes, err := self.c.RuinCmd().Search.Search(searchQuery, opts)
		if err == nil {
			pc.Cards = notes
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
				pc.Cards = notes
			}
		}
	case "queries":
		queriesCtx := gui.Contexts().Queries
		if queriesCtx.CurrentTab == "parents" {
			if len(queriesCtx.Parents) > 0 {
				parent := queriesCtx.Parents[queriesCtx.ParentsTrait().GetSelectedLineIdx()]
				composed, err := self.c.RuinCmd().Parent.ComposeFlat(parent.UUID, parent.Title)
				if err == nil {
					pc.Cards = []models.Note{composed}
				}
			}
		} else if len(queriesCtx.Queries) > 0 {
			query := queriesCtx.Queries[queriesCtx.QueriesTrait().GetSelectedLineIdx()]
			notes, err := self.c.RuinCmd().Queries.Run(query.Name, opts)
			if err == nil {
				pc.Cards = notes
			}
		}
	default:
		self.reloadPreviewCardsFromNotes()
	}

	gui.RenderPreview()
}

func (self *PreviewHelper) reloadPreviewCardsFromNotes() {
	pc := self.ctx()
	opts := self.BuildSearchOptions()
	updated := make([]models.Note, 0, len(pc.Cards))
	for _, card := range pc.Cards {
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
	pc.Cards = updated
}

// --- display toggles ---

// ToggleMarkdown toggles markdown rendering.
func (self *PreviewHelper) ToggleMarkdown() error {
	self.ctx().RenderMarkdown = !self.ctx().RenderMarkdown
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleFrontmatter toggles frontmatter display.
func (self *PreviewHelper) ToggleFrontmatter() error {
	self.ctx().ShowFrontmatter = !self.ctx().ShowFrontmatter
	self.c.GuiCommon().RenderPreview()
	return nil
}

// ToggleTitle toggles title display.
func (self *PreviewHelper) ToggleTitle() error {
	self.ctx().ShowTitle = !self.ctx().ShowTitle
	self.ReloadContent()
	return nil
}

// ToggleGlobalTags toggles global tags display.
func (self *PreviewHelper) ToggleGlobalTags() error {
	self.ctx().ShowGlobalTags = !self.ctx().ShowGlobalTags
	self.ReloadContent()
	return nil
}

// ViewOptionsDialog shows the view options menu.
func (self *PreviewHelper) ViewOptionsDialog() error {
	pc := self.ctx()
	fmLabel := "Show frontmatter"
	if pc.ShowFrontmatter {
		fmLabel = "Hide frontmatter"
	}
	titleLabel := "Show title"
	if pc.ShowTitle {
		titleLabel = "Hide title"
	}
	tagsLabel := "Show global tags"
	if pc.ShowGlobalTags {
		tagsLabel = "Hide global tags"
	}
	mdLabel := "Render markdown"
	if pc.RenderMarkdown {
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
