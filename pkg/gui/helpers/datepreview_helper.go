package helpers

import (
	"strings"
	"time"

	"kvnd/lazyruin/pkg/commands"
	"kvnd/lazyruin/pkg/models"
)

type DatePreviewHelper struct {
	c *HelperCommon
}

func NewDatePreviewHelper(c *HelperCommon) *DatePreviewHelper {
	return &DatePreviewHelper{c: c}
}

func (self *DatePreviewHelper) LoadDatePreview(date string) error {
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
	dp.SelectedCardIdx = 0
	ns := dp.NavState()
	ns.CursorLine = 1
	ns.ScrollOffset = 0
	gui.Contexts().ActivePreviewKey = "datePreview"

	t, _ := time.Parse("2006-01-02", date)
	dp.SetTitle(t.Format("Monday, January 2 2006"))

	gui.RenderPreview()
	gui.PushContextByKey("datePreview")
	return nil
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
