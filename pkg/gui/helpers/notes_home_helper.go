package helpers

import (
	"fmt"
	"time"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/config"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
)

// NotesHomeHelper builds the Home tab's section list and dispatches item
// activations to the appropriate ruin commands.
type NotesHomeHelper struct {
	c             *HelperCommon
	customSection func() []config.NotesPaneSection
}

// NewNotesHomeHelper creates a NotesHomeHelper. The customSections callback
// is invoked on each rebuild so config edits picked up via reload propagate
// without restarting.
func NewNotesHomeHelper(c *HelperCommon, customSections func() []config.NotesPaneSection) *NotesHomeHelper {
	if customSections == nil {
		customSections = func() []config.NotesPaneSection { return nil }
	}
	return &NotesHomeHelper{c: c, customSection: customSections}
}

// inboxLimit caps the Inbox view at a fixed size for v1; configurable
// limits across panels are out of scope (see plan §12).
const inboxLimit = 50

// Refresh rebuilds the section list and assigns it to the NotesHome
// context, preserving selection by ItemID where possible.
func (self *NotesHomeHelper) Refresh() {
	gui := self.c.GuiCommon()
	ctx := gui.Contexts().NotesHome
	if ctx == nil {
		return
	}
	prevID := ctx.SelectedItemID()

	rows := self.BuildRows()
	ctx.SetRowsPreservingSelection(rows, prevID)

	gui.RenderNotes()
}

// BuildRows assembles the full section/item list in display order.
func (self *NotesHomeHelper) BuildRows() []context.NotesHomeRow {
	var rows []context.NotesHomeRow

	// Group 1: Inbox (untitled section, single item).
	rows = append(rows, context.NotesHomeRow{
		Title:  "Inbox",
		ItemID: "hardcoded:inbox",
		Action: context.NotesHomeAction{Kind: context.NotesHomeActionInbox},
	})

	rows = append(rows, context.NotesHomeRow{Blank: true})

	// Group 2: Today + Next 7 Days (untitled section).
	rows = append(rows,
		context.NotesHomeRow{
			Title:  "Today",
			ItemID: "hardcoded:today",
			Action: context.NotesHomeAction{Kind: context.NotesHomeActionToday},
		},
		context.NotesHomeRow{
			Title:  "Next 7 Days",
			ItemID: "hardcoded:next7",
			Action: context.NotesHomeAction{Kind: context.NotesHomeActionNext7},
		},
	)

	rows = append(rows, context.NotesHomeRow{Blank: true})

	// Group 3: Pinned — saved parents + saved queries (each sub-group
	// omitted silently when empty).
	parents, _ := self.c.RuinCmd().Parent.List()
	queries, _ := self.c.RuinCmd().Queries.List()

	if len(parents) > 0 || len(queries) > 0 {
		rows = append(rows, context.NotesHomeRow{IsHeader: true, Title: "Pinned"})
		for i := range parents {
			p := parents[i]
			rows = append(rows, context.NotesHomeRow{
				Title:  parentDisplayTitle(p),
				ItemID: "parent:" + parentBookmarkKey(p),
				Action: context.NotesHomeAction{
					Kind:   context.NotesHomeActionParent,
					Detail: parentBookmarkKey(p),
					Parent: &p,
				},
			})
		}
		if len(parents) > 0 && len(queries) > 0 {
			rows = append(rows, context.NotesHomeRow{Blank: true})
		}
		for _, q := range queries {
			rows = append(rows, context.NotesHomeRow{
				Title:  q.Name,
				ItemID: "query:" + q.Name,
				Action: context.NotesHomeAction{Kind: context.NotesHomeActionQuery, Detail: q.Name},
			})
		}
	}

	// Custom sections from config.
	for sIdx, section := range self.customSection() {
		valid := validateCustomItems(section.Items)
		if len(valid) == 0 {
			continue
		}
		rows = append(rows, context.NotesHomeRow{Blank: true})
		if section.Title != "" {
			rows = append(rows, context.NotesHomeRow{IsHeader: true, Title: section.Title})
		}
		for iIdx, item := range valid {
			rows = append(rows, context.NotesHomeRow{
				Title:  item.Title,
				ItemID: fmt.Sprintf("custom:%d:%d", sIdx, iIdx),
				Action: context.NotesHomeAction{Kind: context.NotesHomeActionEmbed, Detail: item.Embed},
			})
		}
	}

	return trimTrailingNonItems(rows)
}

// trimTrailingNonItems strips trailing blank-spacer and header rows so the
// rendered list never ends on whitespace or a header without items beneath
// it. This keeps the pane visually tight when later groups are omitted.
func trimTrailingNonItems(rows []context.NotesHomeRow) []context.NotesHomeRow {
	for len(rows) > 0 {
		last := rows[len(rows)-1]
		if last.Blank || last.IsHeader {
			rows = rows[:len(rows)-1]
			continue
		}
		break
	}
	return rows
}

// validateCustomItems drops items missing a title or embed string.
func validateCustomItems(items []config.NotesPaneSectionItem) []config.NotesPaneSectionItem {
	var out []config.NotesPaneSectionItem
	for _, it := range items {
		if it.Title == "" || it.Embed == "" {
			continue
		}
		out = append(out, it)
	}
	return out
}

// parentBookmarkKey returns the key used to dispatch a parent bookmark.
// File-based bookmarks aren't supported in Home items (they don't map to a
// search:parent:UUID query); fall back to the bookmark name as a label
// identifier in that case.
func parentBookmarkKey(p models.ParentBookmark) string {
	if p.UUID != "" {
		return p.UUID
	}
	return p.Name
}

func parentDisplayTitle(p models.ParentBookmark) string {
	if p.Title != "" {
		return p.Title
	}
	return p.Name
}

// Activate runs the action attached to a row and commits the result to
// Preview as a navigation entry (Enter handler). Today and Next 7 Days
// dispatch to the rich Date view (tag picks, todos, created/updated
// notes); other rows show as a flat card list.
func (self *NotesHomeHelper) Activate(row context.NotesHomeRow) error {
	if row.IsHeader || row.Blank {
		return nil
	}

	switch row.Action.Kind {
	case context.NotesHomeActionToday:
		return self.c.Helpers().DatePreview().LoadDatePreview(time.Now().Format("2006-01-02"))
	case context.NotesHomeActionNext7:
		start, end := next7DaysRange()
		return self.c.Helpers().DatePreview().LoadDateRangePreview("Next 7 Days", start, end)
	case context.NotesHomeActionParent:
		if row.Action.Parent == nil {
			return nil
		}
		return self.commitParent(*row.Action.Parent)
	}

	title, loadFn := self.dispatch(row)
	if loadFn == nil {
		return nil
	}

	return self.c.Helpers().Navigator().NavigateTo("cardList", title, func() error {
		notes, err := loadFn()
		if err != nil {
			return err
		}
		self.c.Helpers().Preview().ShowCardList(title, notes)
		return nil
	})
}

// Hover runs the row's action as a hover preview — no nav-history entry,
// so j/k browsing through Home items doesn't pollute back/forward state.
// Pressing Enter on the same row promotes the hover to a committed entry
// via Activate.
func (self *NotesHomeHelper) Hover(row context.NotesHomeRow) {
	if row.IsHeader || row.Blank {
		return
	}

	switch row.Action.Kind {
	case context.NotesHomeActionToday:
		_ = self.c.Helpers().DatePreview().HoverDatePreview(time.Now().Format("2006-01-02"))
		return
	case context.NotesHomeActionNext7:
		start, end := next7DaysRange()
		_ = self.c.Helpers().DatePreview().HoverDateRangePreview("Next 7 Days", start, end)
		return
	case context.NotesHomeActionParent:
		if row.Action.Parent == nil {
			return
		}
		_ = self.hoverParent(*row.Action.Parent)
		return
	}

	title, loadFn := self.dispatch(row)
	if loadFn == nil {
		return
	}
	self.c.Helpers().Preview().UpdatePreviewCardList(title, loadFn)
}

// commitParent runs `ruin compose` against a saved parent bookmark and
// commits the result to the Compose preview, mirroring the Queries
// pane's parent-tab activation.
func (self *NotesHomeHelper) commitParent(parent models.ParentBookmark) error {
	gui := self.c.GuiCommon()
	title := "Parent: " + parent.Name
	return self.c.Helpers().Navigator().NavigateTo("compose", title, func() error {
		composed, sourceMap, err := self.c.RuinCmd().Parent.Compose(parent)
		if err != nil {
			gui.ShowError(err)
			return err
		}
		self.c.Helpers().Preview().ShowCompose(title, composed, sourceMap, parent)
		gui.Contexts().Compose.Requery = self.parentRequery(parent)
		return nil
	})
}

// hoverParent is the hover counterpart to commitParent.
func (self *NotesHomeHelper) hoverParent(parent models.ParentBookmark) error {
	gui := self.c.GuiCommon()
	title := "Parent: " + parent.Name
	return self.c.Helpers().Navigator().ShowHover("compose", title, func() error {
		composed, sourceMap, err := self.c.RuinCmd().Parent.Compose(parent)
		if err != nil {
			return err
		}
		self.c.Helpers().Preview().ShowCompose(title, composed, sourceMap, parent)
		gui.Contexts().Compose.Requery = self.parentRequery(parent)
		return nil
	})
}

// parentRequery returns the closure ComposeContext stores so a history
// restore can re-run compose against the same bookmark.
func (self *NotesHomeHelper) parentRequery(parent models.ParentBookmark) context.ComposeRequery {
	return func() (models.Note, []models.SourceMapEntry, error) {
		return self.c.RuinCmd().Parent.Compose(parent)
	}
}

// next7DaysRange returns ISO start/end dates for the [today, today+6]
// window. Inclusive on both ends — same shape as ruin's `between:`.
func next7DaysRange() (string, string) {
	now := time.Now()
	start := now.Format("2006-01-02")
	end := now.AddDate(0, 0, 6).Format("2006-01-02")
	return start, end
}

// dispatch returns the title and a flat-card-list loader for actions
// that don't have a dedicated preview path. Today, Next 7 Days, and
// Parent rows are handled directly in Activate/Hover (Date view and
// Compose view respectively), so they're not represented here.
func (self *NotesHomeHelper) dispatch(row context.NotesHomeRow) (string, func() ([]models.Note, error)) {
	cmd := self.c.RuinCmd()
	opts := self.c.Helpers().Preview().BuildSearchOptions()
	opts.IncludeContent = true
	opts.StripTitle = true
	opts.StripGlobalTags = true

	switch row.Action.Kind {
	case context.NotesHomeActionInbox:
		return "Inbox", func() ([]models.Note, error) {
			o := opts
			o.Sort = "created:desc"
			o.Limit = inboxLimit
			return cmd.Search.Search("tags:none", o)
		}
	case context.NotesHomeActionQuery:
		return row.Title, func() ([]models.Note, error) {
			return cmd.Queries.Run(row.Action.Detail, opts)
		}
	case context.NotesHomeActionEmbed:
		return row.Title, func() ([]models.Note, error) {
			res, err := cmd.Embed.Eval(row.Action.Detail)
			if err != nil {
				return nil, err
			}
			return notesFromEmbedResult(res)
		}
	}
	return "", nil
}

// notesFromEmbedResult flattens an embed result into a slice of preview
// notes. Pick and compose results are deduplicated by UUID since both can
// reference the same note multiple times.
func notesFromEmbedResult(res *commands.EmbedResult) ([]models.Note, error) {
	if res == nil {
		return nil, nil
	}
	switch res.Type {
	case commands.EmbedTypeSearch, commands.EmbedTypeQuery:
		return res.Notes, nil
	case commands.EmbedTypePick:
		return notesFromPickResults(res.Picks), nil
	case commands.EmbedTypeCompose:
		if res.Compose == nil {
			return nil, nil
		}
		seen := map[string]bool{}
		var notes []models.Note
		for _, sm := range res.Compose.SourceMap {
			if sm.UUID == "" || seen[sm.UUID] {
				continue
			}
			seen[sm.UUID] = true
			notes = append(notes, models.Note{UUID: sm.UUID, Title: sm.Title, Path: sm.Path})
		}
		return notes, nil
	}
	return nil, nil
}

// notesFromPickResults dedupes pick groupings to one preview card per note.
// Pick returns groups keyed on UUID with N matched lines each; Preview
// already knows how to expand a card into its content lines.
func notesFromPickResults(picks []models.PickResult) []models.Note {
	seen := map[string]bool{}
	notes := make([]models.Note, 0, len(picks))
	for _, p := range picks {
		if p.UUID == "" || seen[p.UUID] {
			continue
		}
		seen[p.UUID] = true
		notes = append(notes, models.Note{UUID: p.UUID, Title: p.Title, Path: p.File})
	}
	return notes
}
