package helpers

import (
	"strings"
	"time"

	"github.com/donnellyk/lazyruin/pkg/commands"
	"github.com/donnellyk/lazyruin/pkg/gui/context"
	"github.com/donnellyk/lazyruin/pkg/models"
)

type DatePreviewHelper struct {
	c *HelperCommon
}

func NewDatePreviewHelper(c *HelperCommon) *DatePreviewHelper {
	return &DatePreviewHelper{c: c}
}

// LoadDatePreview loads the date preview for the given date as a committed
// navigation: capture-on-departure, fetch data, record a new history entry.
func (self *DatePreviewHelper) LoadDatePreview(date string) error {
	t, _ := time.Parse("2006-01-02", date)
	title := t.Format("Monday, January 2 2006")

	return self.c.Helpers().Navigator().NavigateTo("datePreview", title, func() error {
		self.loadDatePreviewState(date)
		return nil
	})
}

// HoverDatePreview shows the same date-preview content as
// LoadDatePreview but as a hover (no nav-history entry). Used by
// callers that want preview-on-cursor-move without committing.
func (self *DatePreviewHelper) HoverDatePreview(date string) error {
	t, _ := time.Parse("2006-01-02", date)
	title := t.Format("Monday, January 2 2006")

	return self.c.Helpers().Navigator().ShowHover("datePreview", title, func() error {
		self.loadDatePreviewState(date)
		return nil
	})
}

// LoadDateRangePreview loads a date-preview window over the half-open
// range [start, end] (both inclusive, ISO yyyy-mm-dd). Tag and todo
// picks come from `pick @between:start,end`; the notes section pulls
// notes whose `created:` falls in the range. Title is the caller-
// supplied label (e.g., "Next 7 Days") so it reads naturally rather
// than as a date-formatted line.
func (self *DatePreviewHelper) LoadDateRangePreview(title, start, end string) error {
	return self.c.Helpers().Navigator().NavigateTo("datePreview", title, func() error {
		self.loadDateRangeState(title, start, end)
		return nil
	})
}

// HoverDateRangePreview is the hover counterpart to LoadDateRangePreview.
func (self *DatePreviewHelper) HoverDateRangePreview(title, start, end string) error {
	return self.c.Helpers().Navigator().ShowHover("datePreview", title, func() error {
		self.loadDateRangeState(title, start, end)
		return nil
	})
}

// loadDatePreviewState populates DatePreview context state for the given
// date without touching history or context focus. Used as the load closure
// for Navigator.NavigateTo.
func (self *DatePreviewHelper) loadDatePreviewState(date string) {
	tagPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{Date: "@" + date, All: true})
	tagPicks = filterOutTodoLines(tagPicks)
	tagPicks = sortDonePicksLast(tagPicks)

	todoPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{
		Date: "@" + date,
		Todo: true,
		All:  true,
	})

	opts := commands.SearchOptions{
		Sort: "created", Limit: 100, IncludeContent: true, StripTitle: true,
	}
	created, _ := self.c.RuinCmd().Search.Search("created:"+date, opts)
	updated, _ := self.c.RuinCmd().Search.Search("updated:"+date, opts)
	notes := DeduplicateNotes(created, updated)

	gui := self.c.GuiCommon()
	dp := gui.Contexts().DatePreview
	dp.TargetDate = date
	dp.TagPicks = tagPicks
	dp.TodoPicks = todoPicks
	dp.Notes = notes
	self.c.Helpers().TitleCache().PutNotes(notes)
	self.c.Helpers().TitleCache().ResolveUnknownParents(notes)
	dp.SelectedCardIdx = 0
	ns := dp.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	gui.Contexts().ActivePreviewKey = "datePreview"
	dp.Requery = self.dateRequery(date)

	t, _ := time.Parse("2006-01-02", date)
	dp.SetTitle(t.Format("Monday, January 2 2006"))

	gui.RenderPreview()
}

// loadDateRangeState fills DatePreview state from a [start, end] range
// (rather than a single TargetDate). TargetDate is set to "start..end"
// as a sentinel — it's only used for snapshot restore, and the
// Requery closure handles the actual re-fetch.
func (self *DatePreviewHelper) loadDateRangeState(title, start, end string) {
	tagPicks, todoPicks, notes := self.fetchDateRange(start, end)

	gui := self.c.GuiCommon()
	dp := gui.Contexts().DatePreview
	dp.TargetDate = start + ".." + end
	dp.TagPicks = tagPicks
	dp.TodoPicks = todoPicks
	dp.Notes = notes
	self.c.Helpers().TitleCache().PutNotes(notes)
	self.c.Helpers().TitleCache().ResolveUnknownParents(notes)
	dp.SelectedCardIdx = 0
	ns := dp.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	gui.Contexts().ActivePreviewKey = "datePreview"
	dp.Requery = self.dateRangeRequery(start, end)
	dp.SetTitle(title)

	gui.RenderPreview()
}

func (self *DatePreviewHelper) fetchDateRange(start, end string) ([]models.PickResult, []models.PickResult, []models.Note) {
	between := "@between:" + start + "," + end
	tagPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{Date: between, All: true})
	tagPicks = sortDonePicksLast(filterOutTodoLines(tagPicks))

	todoPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{
		Date: between,
		Todo: true,
		All:  true,
	})

	opts := commands.SearchOptions{
		Sort: "created", Limit: 100, IncludeContent: true, StripTitle: true,
	}
	notes, _ := self.c.RuinCmd().Search.Search("between:"+start+","+end, opts)
	return tagPicks, todoPicks, notes
}

func (self *DatePreviewHelper) dateRangeRequery(start, end string) context.DatePreviewRequery {
	return func() ([]models.PickResult, []models.PickResult, []models.Note, error) {
		tag, todo, notes := self.fetchDateRange(start, end)
		return tag, todo, notes, nil
	}
}

// dateRequery returns a closure that re-fetches the three sections for the
// given date, used as DatePreviewContext.Requery on history restore.
func (self *DatePreviewHelper) dateRequery(date string) context.DatePreviewRequery {
	return func() ([]models.PickResult, []models.PickResult, []models.Note, error) {
		tagPicks, err := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{Date: "@" + date, All: true})
		if err != nil {
			return nil, nil, nil, err
		}
		tagPicks = sortDonePicksLast(filterOutTodoLines(tagPicks))

		todoPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{
			Date: "@" + date, Todo: true, All: true,
		})
		opts := commands.SearchOptions{
			Sort: "created", Limit: 100, IncludeContent: true, StripTitle: true,
		}
		created, _ := self.c.RuinCmd().Search.Search("created:"+date, opts)
		updated, _ := self.c.RuinCmd().Search.Search("updated:"+date, opts)
		notes := DeduplicateNotes(created, updated)
		return tagPicks, todoPicks, notes, nil
	}
}

func (self *DatePreviewHelper) ReloadDatePreview() {
	gui := self.c.GuiCommon()
	dp := gui.Contexts().DatePreview
	savedIdx := dp.SelectedCardIdx

	tagPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{Date: "@" + dp.TargetDate, All: true})
	dp.TagPicks = sortDonePicksLast(filterOutTodoLines(tagPicks))
	todoPicks, _ := self.c.RuinCmd().Pick.Pick(nil, commands.PickOpts{
		Date: "@" + dp.TargetDate, Todo: true, All: true,
	})
	dp.TodoPicks = todoPicks
	opts := commands.SearchOptions{
		Sort: "created", Limit: 100, IncludeContent: true, StripTitle: true,
	}
	created, _ := self.c.RuinCmd().Search.Search("created:"+dp.TargetDate, opts)
	updated, _ := self.c.RuinCmd().Search.Search("updated:"+dp.TargetDate, opts)
	dp.Notes = DeduplicateNotes(created, updated)
	self.c.Helpers().TitleCache().PutNotes(dp.Notes)
	self.c.Helpers().TitleCache().ResolveUnknownParents(dp.Notes)

	total := len(dp.TagPicks) + len(dp.TodoPicks) + len(dp.Notes)
	if savedIdx >= total {
		savedIdx = max(total-1, 0)
	}
	dp.SelectedCardIdx = savedIdx

	self.c.Helpers().Notes().FetchNotesForCurrentTab(true)
	gui.RenderPreview()
}

func filterOutTodoLines(results []models.PickResult) []models.PickResult {
	var filtered []models.PickResult
	for _, r := range results {
		var matches []models.PickMatch
		for _, m := range r.Matches {
			trimmed := strings.TrimSpace(m.Content)
			if !strings.HasPrefix(trimmed, "- [ ]") && !strings.HasPrefix(trimmed, "- [x]") {
				matches = append(matches, m)
			}
		}
		if len(matches) > 0 {
			r.Matches = matches
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func sortDonePicksLast(results []models.PickResult) []models.PickResult {
	var active, done []models.PickResult
	for _, r := range results {
		var activeMatches, doneMatches []models.PickMatch
		for _, m := range r.Matches {
			if m.Done {
				doneMatches = append(doneMatches, m)
			} else {
				activeMatches = append(activeMatches, m)
			}
		}
		if len(activeMatches) > 0 {
			rc := r
			rc.Matches = activeMatches
			active = append(active, rc)
		}
		if len(doneMatches) > 0 {
			rc := r
			rc.Matches = doneMatches
			done = append(done, rc)
		}
	}
	return append(active, done...)
}

func DeduplicateNotes(created, updated []models.Note) []models.Note {
	seen := make(map[string]bool)
	var result []models.Note
	for _, n := range created {
		seen[n.UUID] = true
		result = append(result, n)
	}
	for _, n := range updated {
		if !seen[n.UUID] {
			result = append(result, n)
		}
	}
	return result
}

func CurrentWeekday(target time.Weekday) string {
	now := time.Now()
	todayISO := ISOWeekday(now.Weekday())
	targetISO := ISOWeekday(target)
	diff := targetISO - todayISO
	return now.AddDate(0, 0, diff).Format("2006-01-02")
}

func ISOWeekday(d time.Weekday) int {
	if d == time.Sunday {
		return 7
	}
	return int(d)
}
